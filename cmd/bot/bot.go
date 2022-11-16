package main

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/indicator"
	"github.com/lynbklk/tradebot/pkg/strategy"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

func main() {
	writers := make([]io.Writer, 0)
	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	writers = append(writers, newRollingFile("./log"))
	mw := io.MultiWriter(writers...)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(mw).With().Caller().Logger()

	log.Info().Msg("bot started.")

	ctx := context.Background()
	binance, err := exchange.NewBinance(ctx, exchange.WithBinanceCredentials("", ""))
	if err != nil {
		log.Fatal().Msg("init binance failed.")
	}

	agent := indicator.NewAgent(ctx, indicator.WithExchange(binance))

	strategyEma45 := strategy.NewEma45Strategy(agent)
	strategyEma45.AddIndicators("BTCUSDT", "1m")
	strategyEma45.AddIndicators("BTCUSDT", "1h")
	strategyEma45.AddIndicators("BTCUSDT", "1d")
	strategyEma45.Run()
}

func newRollingFile(file string) io.Writer {
	return &lumberjack.Logger{
		Filename:   file,
		MaxBackups: 3,
		MaxSize:    500,
		MaxAge:     30,
	}
}
