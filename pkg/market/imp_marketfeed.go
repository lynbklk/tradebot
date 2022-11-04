package market

import (
	"github.com/lynbklk/tradebot/model"
)

type MarketFeed struct {
	Feed chan model.Candle
	Err  chan error
}

type MarketFeedManager struct {
	exchange       exchange.Exchange
	MarketFeeds    map[string]*MarketFeed
	MarketWatchers map[string][]MarketWatcher
}
