package indicator

import (
	"context"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/market"
	"github.com/lynbklk/tradebot/pkg/util"
	"sync"
)

type Notifier func(*Indicator)

type Agent struct {
	ExchangeWatcher market.Watcher
	Indicators      map[string]*Indicator
	Notifiers       map[string][]Notifier
	mutex           sync.Mutex
	ctx             context.Context
}

type AgentOption func(agent *Agent)

func WithExchange(exchange exchange.Exchange) AgentOption {
	return func(agent *Agent) {
		agent.ExchangeWatcher = market.NewExchangeWatcher(agent.ctx, exchange)
	}
}

func NewAgent(ctx context.Context, options ...AgentOption) *Agent {
	agent := &Agent{
		ctx:        ctx,
		Indicators: make(map[string]*Indicator),
		Notifiers:  make(map[string][]Notifier),
	}
	for _, option := range options {
		option(agent)
	}
	return agent
}

func (a *Agent) Run() {
	for _, indicator := range a.Indicators {
		a.ExchangeWatcher.RegistNotifier(indicator)
	}
	a.ExchangeWatcher.Watch()
}

func (a *Agent) Regist(pair string, timeframe string, notifier Notifier) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	key := util.PairTimeframeToKey(pair, timeframe)

	if _, ok := a.Indicators[key]; !ok {
		a.Indicators[key] = NewIndicator(
			WithPairTimeframe(pair, timeframe),
			WithCandleClose(true),
			WithAgent(a))
	}
	a.Notifiers[key] = append(a.Notifiers[key], notifier)
}

func (a *Agent) Notify(key string) {
	if indicator, ok := a.Indicators[key]; ok {
		if notifiers, ok := a.Notifiers[key]; ok {
			for _, notifier := range notifiers {
				notifier(indicator)
			}
		}
	}
}
