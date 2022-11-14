package order

import "github.com/lynbklk/tradebot/pkg/model"

type Watcher interface {
	Watch(order model.Order)
	GetPair() string
}
