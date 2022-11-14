package strategy

import "github.com/lynbklk/tradebot/pkg/indicator"

type Strategy interface {
	AddIndicator(pair string, timeframe string)
	Update(indicator *indicator.Indicator)
	Run()
}
