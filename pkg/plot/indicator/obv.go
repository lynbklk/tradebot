package indicator

import (
	"time"

	"github.com/lynbklk/tradebot/pkg/model"
	"github.com/lynbklk/tradebot/pkg/plot"

	"github.com/markcheno/go-talib"
)

func OBV(color string) plot.Indicator {
	return &obv{
		Color: color,
	}
}

type obv struct {
	Color  string
	Values model.Series
	Time   []time.Time
}

func (e obv) Name() string {
	return "OBV"
}

func (e obv) Overlay() bool {
	return false
}

func (e *obv) Load(df *model.Dataframe) {
	e.Values = talib.Obv(df.Close, df.Volume)
	e.Time = df.Time
}

func (e obv) Metrics() []plot.IndicatorMetric {
	return []plot.IndicatorMetric{
		{
			Color:  e.Color,
			Style:  "line",
			Values: e.Values,
			Time:   e.Time,
		},
	}
}
