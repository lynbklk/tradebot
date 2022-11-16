package strategy

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/lynbklk/tradebot/pkg/indicator"
	"github.com/lynbklk/tradebot/pkg/util"
)

type Ema45Strategy struct {
	set   mapset.Set[string]
	agent *indicator.Agent
}

func NewEma45Strategy(agent *indicator.Agent) *Ema45Strategy {
	return &Ema45Strategy{
		set:   mapset.NewSet[string](),
		agent: agent,
	}
}

func (s *Ema45Strategy) AddIndicators(pair string, timeframe string) {
	key := util.PairTimeframeToKey(pair, timeframe)
	if !s.set.Contains(key) {
		s.set.Add(key)
	}
}

func (s *Ema45Strategy) Update(indicator *indicator.Indicator) {
	ema5 := indicator.GetDataframe().Metadata["ema5"].Last(0)
	close := indicator.GetDataframe().Close.Last(0)
	fmt.Printf("ema8: %f, close: %f\n", ema5, close)
	return
}

func (s *Ema45Strategy) Run() {
	keys := s.set.ToSlice()
	for _, key := range keys {
		pair, timeframe := util.PairTimeframeFromKey(key)
		s.agent.Regist(pair, timeframe, s.Update)
	}
	s.agent.Run()
}
