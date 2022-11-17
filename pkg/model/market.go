package model

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

type DataInfo struct {
	Pair      string
	Timeframe string
}

type AssetInfo struct {
	BaseAsset             string
	QuoteAsset            string
	MinPrice              float64
	MaxPrice              float64
	MinQuantity           float64
	MaxQuantity           float64
	StepSize              float64
	TickSize              float64
	QtyDecimalPrecision   int64
	PriceDecimalPrecision int64
}

type Dataframe struct {
	Pair       string
	Close      Series
	Open       Series
	High       Series
	Low        Series
	Volume     Series
	Time       []time.Time
	LastUpdate time.Time
	Metadata   map[string]Series
}

type Candle struct {
	Pair      string
	Timeframe string
	Time      time.Time
	UpdatedAt time.Time
	Open      float64
	Close     float64
	Low       float64
	High      float64
	Volume    float64
	Trades    int64
	Complete  bool
}

type HeikinAshi struct {
	PreviousHACandle Candle
}

func NewHeikinAshi() *HeikinAshi {
	return &HeikinAshi{}
}

func (c Candle) ToSlice(precision int) []string {
	return []string{
		fmt.Sprintf("%d", c.Time.Unix()),
		strconv.FormatFloat(c.Open, 'f', precision, 64),
		strconv.FormatFloat(c.Close, 'f', precision, 64),
		strconv.FormatFloat(c.Low, 'f', precision, 64),
		strconv.FormatFloat(c.High, 'f', precision, 64),
		fmt.Sprintf("%.1f", c.Volume),
		fmt.Sprintf("%d", c.Trades),
	}
}

func (c Candle) ToHeikinAshi(ha *HeikinAshi) Candle {
	haCandle := ha.CalculateHeikinAshi(c)
	return Candle{
		Pair:      c.Pair,
		Open:      haCandle.Open,
		High:      haCandle.High,
		Low:       haCandle.Low,
		Close:     haCandle.Close,
		Volume:    c.Volume,
		Complete:  c.Complete,
		Time:      c.Time,
		UpdatedAt: c.UpdatedAt,
		Trades:    c.Trades,
	}
}

func (c Candle) Less(j Item) bool {
	diff := j.(Candle).Time.Sub(c.Time)
	if diff < 0 {
		return false
	}
	if diff > 0 {
		return true
	}

	diff = j.(Candle).UpdatedAt.Sub(c.UpdatedAt)
	if diff < 0 {
		return false
	}
	if diff > 0 {
		return true
	}

	return c.Pair < j.(Candle).Pair
}

func (ha *HeikinAshi) CalculateHeikinAshi(c Candle) Candle {
	var hkCandle Candle

	highValues := []float64{c.High, c.Open, c.Close}
	sort.Float64s(highValues)

	lowValues := []float64{c.Low, c.Open, c.Close}
	sort.Float64s(lowValues)

	openValue := ha.PreviousHACandle.Open
	closeValue := ha.PreviousHACandle.Close

	// First HA candle is calculated using current candle
	if (ha.PreviousHACandle == Candle{}) {
		openValue = c.Open
		closeValue = c.Close
	}

	hkCandle.Open = (openValue + closeValue) / 2
	hkCandle.High = highValues[2]
	hkCandle.Close = (c.Open + c.High + c.Low + c.Close) / 4
	hkCandle.Low = lowValues[0]
	ha.PreviousHACandle = hkCandle

	return hkCandle
}
