package market

import (
	"context"
	"github.com/StudioSol/set"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/util"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
	"time"
)

type ExchangeWatcher struct {
	ctx       context.Context
	Exchange  exchange.Exchange
	Feeds     map[string]*ExchangeFeed
	Notifiers map[string][]Notifier
	Keys      *set.LinkedHashSetString
}

type ExchangeFeed struct {
	Pair      string
	Timeframe string
	Data      chan *model.Candle
	Err       chan error
}

func NewExchangeWatcher(ctx context.Context, e exchange.Exchange) Watcher {
	return &ExchangeWatcher{
		ctx:       ctx,
		Exchange:  e,
		Feeds:     make(map[string]*ExchangeFeed),
		Notifiers: make(map[string][]Notifier),
		Keys:      set.NewLinkedHashSetString(),
	}
}

func (w *ExchangeWatcher) RegistNotifier(notifier Notifier) {
	dataInfo := notifier.GetDataInfo()
	key := util.PairTimeframeToKey(dataInfo.Pair, dataInfo.Timeframe)
	w.Keys.Add(key)
	w.Notifiers[key] = append(w.Notifiers[key], notifier)
}

func (w *ExchangeWatcher) Watch() {
	w.connect()
	wg := new(sync.WaitGroup)
	for key, feed := range w.Feeds {
		wg.Add(1)
		go func(key string, feed *ExchangeFeed) {
			for {
				select {
				case candle, ok := <-feed.Data:
					if !ok {
						wg.Done()
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
	wg.Wait()
}

func (w *ExchangeWatcher) connect() {
	log.Info().Msg("Connecting to the exchange.")
	for key := range w.Keys.Iter() {
		pair, timeframe := util.PairTimeframeFromKey(key)
		// preload
		days := -1
		if strings.HasSuffix(timeframe, "d") {
			periods := util.TimeframePeriods[timeframe]
			days = 0 - periods[len(periods)-1] - 2
		}
		candles, _ := w.Exchange.GetCandlesByPeriod(w.ctx, pair, timeframe, time.Now().AddDate(0, 0, days), time.Now())
		log.Info().Msgf("preload candles, pair: %s, timeframe: %s, len: %d", pair, timeframe, len(candles))
		for _, candle := range candles {
			for _, notifier := range w.Notifiers[key] {
				notifier.Notify(&candle, true)
			}
		}
		// subscribe
		data, err := w.Exchange.SubscribeCandle(w.ctx, pair, timeframe)
		w.Feeds[key] = &ExchangeFeed{
			Pair:      pair,
			Timeframe: timeframe,
			Data:      data,
			Err:       err,
		}
	}
}
