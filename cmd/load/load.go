package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lynbklk/tradebot/pkg/download"
	"github.com/lynbklk/tradebot/pkg/exchange"
)

var (
	Pair      string
	Timeframe string
	Days      int
	Output    string
)

func main() {
	flag.Parse()
	ctx := context.Background()
	binance, err := exchange.NewBinance(ctx, exchange.WithBinanceCredentials("", ""))
	if err != nil {
		fmt.Errorf("creating exchange failed. error: %v", err)
		return
	}

	loader := download.NewDownloader(binance)
	if loader.Download(ctx, Pair, Timeframe, Output, download.WithDays(Days)) != nil {
		fmt.Errorf("download failed. error: %v", err)
		return
	}
	fmt.Println("download succeed.")
	return
}

func init() {
	flag.StringVar(&Pair, "pair", "", "coin pair")
	flag.StringVar(&Timeframe, "timeframe", "1m", "timeframe")
	flag.IntVar(&Days, "days", 1, "from this num of days ago")
	flag.StringVar(&Output, "output", "", "output file")
}
