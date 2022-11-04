package strategies

import (
	"github.com/lynbklk/tradebot/service"
	"github.com/lynbklk/tradebot/tradebot"

	"github.com/markcheno/go-talib"
	log "github.com/sirupsen/logrus"
)

// https://www.investopedia.com/articles/trading/08/turtle-trading.asp
type Turtle struct{}

func (e Turtle) Timeframe() string {
	return "4h"
}

func (e Turtle) WarmupPeriod() int {
	return 40
}

func (e Turtle) Indicators(df *tradebot.Dataframe) {
	df.Metadata["turtleHighest"] = talib.Max(df.Close, 40)
	df.Metadata["turtleLowest"] = talib.Min(df.Close, 20)
}

func (e *Turtle) OnCandle(df *tradebot.Dataframe, broker service.Broker) {
	closePrice := df.Close.Last(0)
	highest := df.Metadata["turtleHighest"].Last(0)
	lowest := df.Metadata["turtleLowest"].Last(0)

	assetPosition, quotePosition, err := broker.Position(df.Pair)
	if err != nil {
		log.Error(err)
		return
	}

	// If position already open wait till it will be closed
	if assetPosition == 0 && closePrice >= highest {
		_, err := broker.CreateOrderMarketQuote(tradebot.SideTypeBuy, df.Pair, quotePosition/2)
		if err != nil {
			log.WithFields(map[string]interface{}{
				"pair":  df.Pair,
				"side":  tradebot.SideTypeBuy,
				"close": closePrice,
				"asset": assetPosition,
				"quote": quotePosition,
			}).Error(err)
		}
		return
	}

	if assetPosition > 0 && closePrice <= lowest {
		_, err := broker.CreateOrderMarket(tradebot.SideTypeSell, df.Pair, assetPosition)
		if err != nil {
			log.WithFields(map[string]interface{}{
				"pair":  df.Pair,
				"side":  tradebot.SideTypeSell,
				"close": closePrice,
				"asset": assetPosition,
				"quote": quotePosition,
				"size":  assetPosition,
			}).Error(err)
		}
	}
}
