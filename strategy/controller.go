package strategy

import (
	"github.com/lynbklk/tradebot/pkg/exchange"
	log "github.com/sirupsen/logrus"

	"github.com/lynbklk/tradebot/pkg/model"
)

type Controller struct {
	strategy  Strategy
	dataframe *model.Dataframe
	trader    exchange.Trader
	started   bool
}

func NewStrategyController(pair string, strategy Strategy, trader exchange.Trader) *Controller {
	dataframe := &model.Dataframe{
		Pair:     pair,
		Metadata: make(map[string]model.Series),
	}

	return &Controller{
		dataframe: dataframe,
		strategy:  strategy,
		trader:    trader,
	}
}

func (s *Controller) Start() {
	s.started = true
}

func (s *Controller) OnPartialCandle(candle model.Candle) {
	if !candle.Complete && len(s.dataframe.Close) >= s.strategy.WarmupPeriod() {
		if str, ok := s.strategy.(HighFrequencyStrategy); ok {
			s.updateDataFrame(candle)
			str.Indicators(s.dataframe)
			str.OnPartialCandle(s.dataframe, s.trader)
		}
	}
}

func (s *Controller) updateDataFrame(candle model.Candle) {
	if len(s.dataframe.Time) > 0 && candle.Time.Equal(s.dataframe.Time[len(s.dataframe.Time)-1]) {
		last := len(s.dataframe.Time) - 1
		s.dataframe.Close[last] = candle.Close
		s.dataframe.Open[last] = candle.Open
		s.dataframe.High[last] = candle.High
		s.dataframe.Low[last] = candle.Low
		s.dataframe.Volume[last] = candle.Volume
		s.dataframe.Time[last] = candle.Time
	} else {
		s.dataframe.Close = append(s.dataframe.Close, candle.Close)
		s.dataframe.Open = append(s.dataframe.Open, candle.Open)
		s.dataframe.High = append(s.dataframe.High, candle.High)
		s.dataframe.Low = append(s.dataframe.Low, candle.Low)
		s.dataframe.Volume = append(s.dataframe.Volume, candle.Volume)
		s.dataframe.Time = append(s.dataframe.Time, candle.Time)
		s.dataframe.LastUpdate = candle.Time
	}
}

func (s *Controller) OnCandle(candle model.Candle) {
	if len(s.dataframe.Time) > 0 && candle.Time.Before(s.dataframe.Time[len(s.dataframe.Time)-1]) {
		log.Errorf("late candle received: %#v", candle)
		return
	}

	s.updateDataFrame(candle)

	if len(s.dataframe.Close) >= s.strategy.WarmupPeriod() {
		s.strategy.Indicators(s.dataframe)
		if s.started {
			s.strategy.OnCandle(s.dataframe, s.trader)
		}
	}
}
