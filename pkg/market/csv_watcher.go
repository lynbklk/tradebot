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

var (
	TimeframeValue = map[string]int{
		"m": 1,
		"h": 2,
		"d": 3,
		"w": 4,
	}
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
	Data      chan *model.Candle
	Err       chan error
	File      string
	Buf       [][]string
	eof       bool
}

type CandleIndex struct {
	candle *model.Candle
	index  int
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
	pair, timeframe := util.PairTimeframeFromKey(key)
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
		Timeframe: timeframe,
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

func (w *CsvWatcher) isCandleEnd(keyCandles map[string]*model.Candle) bool {
	for _, candle := range keyCandles {
		if candle != nil {
			return false
		}
	}
	return true
}

func (w *CsvWatcher) largeTimevalCandle(first *model.Candle, second *model.Candle) bool {
	if first.Time.Equal(second.Time) {
		tf1, tf2 := first.Timeframe, second.Timeframe
		tv1, tv2 := tf1[len(tf1)-1:], tf2[len(tf2)-1:]
		if tv1 != tv2 {
			return TimeframeValue[tv1] > TimeframeValue[tv2]
		}
		tn1, _ := strconv.Atoi(tf1[0 : len(tf1)-1])
		tn2, _ := strconv.Atoi(tf2[0 : len(tf2)-1])
		return tn1 > tn2
	}
	return first.Time.After(second.Time)
}

func (w *CsvWatcher) getLatestCandle(keyCandles map[string]*CandleIndex) (string, *model.Candle, bool) {
	latestKey := ""
	for key, candleIndex := range keyCandles {
		if candleIndex == nil {
			continue
		}
		if len(latestKey) == 0 || w.largeTimevalCandle(keyCandles[latestKey].candle, keyCandles[key].candle) {
			latestKey = key
		}
	}
	if len(latestKey) == 0 {
		return "", nil, true
	}
	return latestKey, keyCandles[latestKey].candle, false
}

func (w *CsvWatcher) Watch() {
	keyCandles := make(map[string]*CandleIndex)
	for key, _ := range w.Feeds {
		candle, err := w.readOneCandleFromBuf(key, 0)
		if err != nil {
			log.Fatal().Err(err).Msgf("read one line failed. key: %s", key)
		}
		keyCandles[key] = &CandleIndex{
			candle: candle,
			index:  0,
		}
	}
	for {
		key, candle, end := w.getLatestCandle(keyCandles)
		if end {
			break
		}
		for _, notifier := range w.Notifiers[key] {
			notifier.Notify(candle, false)
		}
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
			Data:      make(chan *model.Candle),
			File:      file,
			eof:       false,
		}
	}
}
