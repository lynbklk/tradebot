package bot

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/config"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/indicator"
	"github.com/lynbklk/tradebot/pkg/market"
	"github.com/lynbklk/tradebot/pkg/notifier"
	"github.com/lynbklk/tradebot/pkg/order"
	log "github.com/sirupsen/logrus"
)

type TelegramBot struct {
	exchange      exchange.Exchange
	agent         indicator.Agent
	notifier      notifier.Notifier
	marketMonitor market.Monitor
	orderMonitor  order.Monitor
}

type Option func(*TelegramBot)

func NewTelegramBot() *TelegramBot {
	// new exchange
	binance, err := exchange.NewBinance(
		context.Background(),
		exchange.WithBinanceCredentials(config.C.Key, config.C.Secret))
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	// new market monitor
	marketMonitor := market.NewMonitor(binance)
	// TODO: marketMonitor.RegistWatcher()
	// TODO: marketMonitor.Start()

	telegramBot := &TelegramBot{
		exchange:      binance,
		marketMonitor: marketMonitor,
	}
	return telegramBot
}
