package exchange

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/model"
	"time"
)

type Feeder interface {
	GetAssetsInfo(pair string) model.AssetInfo
	GetLastQuote(ctx context.Context, pair string) (float64, error)
	GetCandlesByPeriod(ctx context.Context, pair, period string, start, end time.Time) ([]model.Candle, error)
	GetCandlesByLimit(ctx context.Context, pair, period string, limit int) ([]model.Candle, error)
	SubscribeCandle(ctx context.Context, pair, timeframe string) (chan *model.Candle, chan error)
}

type Trader interface {
	Account() (model.Account, error)
	Position(pair string) (asset, quote float64, err error)
	Order(pair string, id int64) (model.Order, error)
	CreateOrderOCO(side model.SideType, pair string, size, price, stop, stopLimit float64) ([]model.Order, error)
	CreateOrderLimit(side model.SideType, pair string, size float64, limit float64) (model.Order, error)
	CreateOrderMarket(side model.SideType, pair string, size float64) (model.Order, error)
	CreateOrderMarketQuote(side model.SideType, pair string, quote float64) (model.Order, error)
	CreateOrderStop(pair string, quantity float64, limit float64) (model.Order, error)
	Cancel(model.Order) error
}

type Exchange interface {
	Feeder
	Trader
}
