package market

import (
	"context"
	"github.com/StudioSol/set"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/util"
	log "github.com/sirupsen/logrus"
	"sync"
)

type ExchangeWatcher struct {
	ctx       context.Context
	Exchange  exchange.Exchange
	Feeds     map[string]*Feed
	Notifiers map[string][]Notifier
	Keys      *set.LinkedHashSetString
}

type Feed struct {
	Pair      string
	Timeframe string
	Data      chan model.Candle
	Err       chan error
}

func NewExchangeWatcher(ctx context.Context, e exchange.Exchange) Watcher {
	return &ExchangeWatcher{
		ctx:       ctx,
		Exchange:  e,
		Feeds:     make(map[string]*Feed),
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

func (w *ExchangeWatcher) Preload(pair string, timeframe string, candles []model.Candle) {
	log.Infof("[SETUP] preloading %d candles for %s-%s", len(candles), pair, timeframe)
	key := util.PairTimeframeToKey(pair, timeframe)
	for _, candle := range candles {
		if !candle.Complete {
			continue
		}

		for _, notifier := range w.Notifiers[key] {
			notifier.Notify(candle)
		}
	}
}

func (w *ExchangeWatcher) Watch() {
	w.connect()
	wg := new(sync.WaitGroup)
	for key, feed := range w.Feeds {
		wg.Add(1)
		go func(key string, feed *Feed) {
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
						notifier.Notify(candle)
					}
				case err := <-feed.Err:
					if err != nil {
						log.Error("dataFeedSubscription/start: ", err)
					}
				}
			}
		}(key, feed)
	}

	log.Infof("Data feed connected.")
	wg.Wait()
}

func (w *ExchangeWatcher) connect() {
	log.Infof("Connecting to the exchange.")
	for key := range w.Keys.Iter() {
		pair, timeframe := util.PairTimeframeFromKey(key)
		// preload
		candles, _ := w.Exchange.GetCandlesByLimit(w.ctx, pair, timeframe, 30)
		for _, candle := range candles {
			for _, notifier := range w.Notifiers[key] {
				notifier.Notify(candle)
			}
		}
		// subscribe
		data, err := w.Exchange.SubscribeCandle(w.ctx, pair, timeframe)
		w.Feeds[key] = &Feed{
			Pair:      pair,
			Timeframe: timeframe,
			Data:      data,
			Err:       err,
		}
	}
}
