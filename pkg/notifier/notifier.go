package notifier

import "github.com/lynbklk/tradebot/pkg/model"

type Notifier interface {
	Notify(string)
	OnOrder(order model.Order)
	OnError(err error)
	Start()
}
