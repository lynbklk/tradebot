package indicator

import (
	"fmt"
	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/util"
	"github.com/markcheno/go-talib"
	"github.com/rs/zerolog/log"
	"strings"
)

type Indicator struct {
	Pair          string
	Timeframe     string
	OnCandleClose bool
	Dataframe     *model.Dataframe
	Agent         *Agent
}

type Option func(*Indicator)

func WithAgent(agent *Agent) Option {
	return func(indicator *Indicator) {
		indicator.Agent = agent
	}
}

func WithPairTimeframe(pair string, timeframe string) Option {
	return func(indicator *Indicator) {
		indicator.Pair = pair
		indicator.Timeframe = timeframe
	}
}

func WithCandleClose(close bool) Option {
	return func(indicator *Indicator) {
		indicator.OnCandleClose = close
	}
}

func NewIndicator(options ...Option) *Indicator {
	indicator := &Indicator{
		Dataframe: &model.Dataframe{},
	}
	for _, option := range options {
		option(indicator)
	}
	return indicator
}

func (i *Indicator) GetDataInfo() model.DataInfo {
	return model.DataInfo{
		Pair:      i.Pair,
		Timeframe: i.Timeframe,
	}
}

func (i *Indicator) IsOnCandleClose() bool {
	return i.OnCandleClose
}

func (i *Indicator) Notify(candle model.Candle, preload bool) {
	log.Info().Msgf("indicator notify. indicator: %v, candle: %v", i.GetDataInfo(), candle)
	i.updateDataframe(candle)
	if !preload && len(i.Dataframe.Close) >= 30 {
		i.updateMetaData()
		i.Agent.Notify(util.PairTimeframeToKey(i.Pair, i.Timeframe))
	}
}

func (i *Indicator) GetValue(t string, period int) model.Series {
	key := fmt.Sprintf("%s%d", strings.ToLower(t), period)
	return i.Dataframe.Metadata[key]
}

func (i *Indicator) GetDataframe() *model.Dataframe {
	return i.Dataframe
}

func (i *Indicator) updateDataframe(candle model.Candle) {
	if len(i.Dataframe.Time) > 0 && candle.Time.Equal(i.Dataframe.Time[len(i.Dataframe.Time)-1]) {
		last := len(i.Dataframe.Time) - 1
		i.Dataframe.Close[last] = candle.Close
		i.Dataframe.Open[last] = candle.Open
		i.Dataframe.High[last] = candle.High
		i.Dataframe.Low[last] = candle.Low
		i.Dataframe.Volume[last] = candle.Volume
		i.Dataframe.Time[last] = candle.Time
	} else {
		i.Dataframe.Close = append(i.Dataframe.Close, candle.Close)
		i.Dataframe.Open = append(i.Dataframe.Open, candle.Open)
		i.Dataframe.High = append(i.Dataframe.High, candle.High)
		i.Dataframe.Low = append(i.Dataframe.Low, candle.Low)
		i.Dataframe.Volume = append(i.Dataframe.Volume, candle.Volume)
		i.Dataframe.Time = append(i.Dataframe.Time, candle.Time)
		i.Dataframe.LastUpdate = candle.Time
	}
}

func (i *Indicator) updateMetaData() {
	i.Dataframe.Metadata["ema8"] = talib.Ema(i.Dataframe.Close, 8)
	i.Dataframe.Metadata["sma21"] = talib.Sma(i.Dataframe.Close, 21)
}
