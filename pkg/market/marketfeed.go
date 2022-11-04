package market

import (
	"github.com/lynbklk/tradebot/model"
)

type MarketFeed interface {
}

type MarketFeedManager struct {
	exchange       exchange.Exchange
	MarketFeeds    map[string]*MarketFeed
	MarketWatchers map[string][]MarketWatcher
}
