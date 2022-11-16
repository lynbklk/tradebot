package market

import (
	"encoding/csv"
	"github.com/StudioSol/set"
	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/util"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

type CsvWatcher struct {
	Feeds     map[string]*CsvFeed
	Notifiers map[string][]Notifier
	Files     map[string]string
	Keys      *set.LinkedHashSetString
}

type CsvFeed struct {
	Pair      string
	Timeframe string
	Data      chan model.Candle
	Err       chan error
	File      string
	Buf       [][]string
	eof       bool
}

func NewCsvWatcher(files map[string]string) Watcher {
	return &CsvWatcher{
		Files: files,
	}
}

func (w *CsvWatcher) RegistNotifier(notifier Notifier) {
	dataInfo := notifier.GetDataInfo()
	key := util.PairTimeframeToKey(dataInfo.Pair, dataInfo.Timeframe)
	w.Keys.Add(key)
	w.Notifiers[key] = append(w.Notifiers[key], notifier)
}

func (w *CsvWatcher) readOneCandleFromBuf(key string, index int) (*model.Candle, error) {
	pair, _ := util.PairTimeframeFromKey(key)
	if index >= len(w.Feeds[key].Buf) {
		return nil, nil
	}
	line := w.Feeds[key].Buf[index]
	timestamp, err := strconv.Atoi(line[0])
	if err != nil {
		log.Fatal().Err(err).Msg("read line failed")
		return nil, err
	}

	candle := &model.Candle{
		Time:      time.Unix(int64(timestamp), 0).UTC(),
		UpdatedAt: time.Unix(int64(timestamp), 0).UTC(),
		Pair:      pair,
		Complete:  true,
	}

	candle.Open, err = strconv.ParseFloat(line[1], 64)
	if err != nil {
		return nil, err
	}

	candle.Close, err = strconv.ParseFloat(line[2], 64)
	if err != nil {
		return nil, err
	}

	candle.Low, err = strconv.ParseFloat(line[3], 64)
	if err != nil {
		return nil, err
	}

	candle.High, err = strconv.ParseFloat(line[4], 64)
	if err != nil {
		return nil, err
	}

	candle.Volume, err = strconv.ParseFloat(line[5], 64)
	if err != nil {
		return nil, err
	}
	return candle, nil
}

func (w *CsvWatcher) Watch() {
	keyCandles := make(map[string]model.Candle)

	for key, feed := range w.Feeds {
		candle, err := w.readOneCandleFromBuf(key, 0)
		if err != nil {
			log.Fatal().Err(err).Msgf("read one line failed. key: %s", key)
		}
		if candle == nil {
			keyEof[key] = true
		}

		go func(key string, feed *CsvFeed) {
			for {
				select {
				case candle, ok := <-feed.Data:
					if !ok {
						return
					}
					for _, notifier := range w.Notifiers[key] {
						if notifier.IsOnCandleClose() && !candle.Complete {
							continue
						}
						notifier.Notify(candle, false)
					}
				case err := <-feed.Err:
					if err != nil {
						log.Error().Err(err).Msg("dataFeedSubscription start failed.")
					}
				}
			}
		}(key, feed)
	}

	log.Info().Msg("Data feed connected.")
}

func (w *CsvWatcher) connect() {
	log.Info().Msg("Connecting to the csv files.")
	for key, file := range w.Files {
		pair, timeframe := util.PairTimeframeFromKey(key)
		csvFile, err := os.Open(file)
		if err != nil {
			log.Fatal().Msgf("open csv file failed. file: %s", file)
		}
		csvLines, err := csv.NewReader(csvFile).ReadAll()
		if err != nil {
			log.Fatal().Msgf("read csv file failed. file: %s", file)
		}
		w.Feeds[key] = &CsvFeed{
			Pair:      pair,
			Timeframe: timeframe,
			Buf:       csvLines,
			Data:      make(chan model.Candle),
			File:      file,
			eof:       false,
		}
	}
}
