package order

import (
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/model"
)

type Monitor interface {
	Start()
	RegistWatcher(watcher Watcher)
	Publish(order model.Order)
}

type Feed struct {
	Pair string
	Data chan model.Order
	Err  chan error
}

type MonitorV1 struct {
	Exchange exchange.Exchange
	Feeds    map[string]*Feed
	Watchers map[string][]Watcher
}

func NewMonitor(e exchange.Exchange) Monitor {
	return &MonitorV1{
		Exchange: e,
		Feeds:    make(map[string]*Feed),
		Watchers: make(map[string][]Watcher),
	}
}

func (m *MonitorV1) Start() {
	for pair := range m.Feeds {
		go func(pair string, feed *Feed) {
			for order := range feed.Data {
				for _, watcher := range m.Watchers[pair] {
					watcher.Watch(order)
				}
			}
		}(pair, m.Feeds[pair])
	}
}

func (m *MonitorV1) RegistWatcher(watcher Watcher) {
	pair := watcher.GetPair()
	if _, ok := m.Feeds[pair]; !ok {
		m.Feeds[pair] = &Feed{
			Pair: pair,
			Data: make(chan model.Order),
			Err:  make(chan error),
		}
	}
	m.Watchers[pair] = append(m.Watchers[pair], watcher)
}

func (m *MonitorV1) Publish(order model.Order) {
	if _, ok := m.Feeds[order.Pair]; ok {
		m.Feeds[order.Pair].Data <- order
	}
}
