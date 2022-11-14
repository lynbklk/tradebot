package main

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/indicator"
	"github.com/lynbklk/tradebot/pkg/strategy"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	binance, err := exchange.NewBinance(ctx, exchange.WithBinanceCredentials("", ""))
	if err != nil {
		log.Fatal("init binance failed.")
	}

	agent := indicator.NewAgent(ctx, indicator.WithExchange(binance))

	strategyEma45 := strategy.NewEma45Strategy(agent)
	strategyEma45.AddIndicators("btcusdt", "1m")
	strategyEma45.Run()
}
