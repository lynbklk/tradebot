package main

import (
	"context"
	"os"
	"strconv"

	"github.com/lynbklk/tradebot/plot"
	"github.com/lynbklk/tradebot/plot/indicator"

	"github.com/lynbklk/tradebot/examples/strategies"
	"github.com/lynbklk/tradebot/exchange"
	"github.com/lynbklk/tradebot/storage"
	"github.com/lynbklk/tradebot/tradebot"

	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		ctx             = context.Background()
		telegramToken   = os.Getenv("TELEGRAM_TOKEN")
		telegramUser, _ = strconv.Atoi(os.Getenv("TELEGRAM_USER"))
	)

	settings := tradebot.Settings{
		Pairs: []string{
			"BTCUSDT",
			"ETHUSDT",
			"BNBUSDT",
			"LTCUSDT",
		},
		Telegram: tradebot.TelegramSettings{
			Enabled: true,
			Token:   telegramToken,
			Users:   []int{telegramUser},
		},
	}

	// Use binance for realtime data feed
	binance, err := exchange.NewBinance(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// creating a storage to save trades
	storage, err := storage.FromMemory()
	if err != nil {
		log.Fatal(err)
	}

	// creating a paper wallet to simulate an exchange waller for fake operataions
	paperWallet := exchange.NewPaperWallet(
		ctx,
		"USDT",
		exchange.WithPaperFee(0.001, 0.001),
		exchange.WithPaperAsset("USDT", 10000),
		exchange.WithDataFeed(binance),
	)

	// initializing my strategy
	strategy := new(strategies.CrossEMA)

	chart, err := plot.NewChart(
		plot.WithIndicators(
			indicator.EMA(8, "red"),
			indicator.SMA(21, "blue"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// initializer tradebot
	bot, err := tradebot.NewBot(
		ctx,
		settings,
		paperWallet,
		strategy,
		tradebot.WithStorage(storage),
		tradebot.WithPaperWallet(paperWallet),
		tradebot.WithCandleSubscription(chart),
		tradebot.WithOrderSubscription(chart),
	)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		err := chart.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = bot.Run(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
