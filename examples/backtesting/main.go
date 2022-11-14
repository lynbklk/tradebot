package main

/*
func main() {
	ctx := context.Background()

	settings := tradebot.Settings{
		Pairs: []string{
			"BTCUSDT",
			"ETHUSDT",
		},
	}

	strategy := new(strategies.CrossEMA)

	csvFeed, err := exchange.NewCSVFeed(
		strategy.Timeframe(),
		exchange.PairFeed{
			Pair:      "BTCUSDT",
			File:      "testdata/btc-1h.csv",
			Timeframe: "1h",
		},
		exchange.PairFeed{
			Pair:      "ETHUSDT",
			File:      "testdata/eth-1h.csv",
			Timeframe: "1h",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.FromMemory()
	if err != nil {
		log.Fatal(err)
	}

	wallet := exchange.NewPaperWallet(
		ctx,
		"USDT",
		exchange.WithPaperAsset("USDT", 10000),
		exchange.WithDataFeed(csvFeed),
	)

	chart, err := plot.NewChart(plot.WithIndicators(
		indicator.EMA(8, "red"),
		indicator.SMA(21, "#000"),
		indicator.RSI(14, "purple"),
	), plot.WithPaperWallet(wallet))
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tradebot.NewBot(
		ctx,
		settings,
		wallet,
		strategy,
		tradebot.WithBacktest(wallet),
		tradebot.WithStorage(storage),
		tradebot.WithCandleSubscription(chart),
		tradebot.WithOrderSubscription(chart),
		tradebot.WithLogLevel(log.WarnLevel),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = bot.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Print bot results
	bot.Summary()

	// Display candlesticks chart in browser
	err = chart.Start()
	if err != nil {
		log.Fatal(err)
	}
}
*/
