package bot

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/config"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/indicator"
	"github.com/lynbklk/tradebot/pkg/market"
	"github.com/lynbklk/tradebot/pkg/notifier"
	"github.com/lynbklk/tradebot/pkg/order"
	"github.com/rs/zerolog/log"
)

type TelegramBot struct {
	exchange      exchange.Exchange
	agent         indicator.Agent
	notifier      notifier.Notifier
	marketWatcher market.Watcher
	orderMonitor  order.Monitor
}

type Option func(*TelegramBot)

func NewTelegramBot() *TelegramBot {
	ctx := context.Background()
	// new exchange
	binance, err := exchange.NewBinance(
		ctx,
		exchange.WithBinanceCredentials(config.C.Key, config.C.Secret))
	if err != nil {
		log.Fatal().Err(err).Msg("init binance failed.")
		return nil
	}
	// new market monitor
	marketWatchr := market.NewExchangeWatcher(ctx, binance)
	// TODO: marketMonitor.RegistWatcher()
	// TODO: marketMonitor.Start()

	telegramBot := &TelegramBot{
		exchange:      binance,
		marketWatcher: marketWatchr,
	}
	return telegramBot
}
