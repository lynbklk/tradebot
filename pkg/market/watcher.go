package market

import (
	"github.com/lynbklk/tradebot/pkg/model"
)

type Watcher interface {
	Watch()
	RegistNotifier(notifier Notifier)
}

type Notifier interface {
	GetDataInfo() model.DataInfo
	Notify(candle model.Candle, preload bool)
	IsOnCandleClose() bool
}
