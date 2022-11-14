package order

import (
	"fmt"
	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/olekukonko/tablewriter"
	"math"
	"strconv"
	"strings"
)

type summary struct {
	Pair   string
	Win    []float64
	Lose   []float64
	Volume float64
}

func (s summary) Profit() float64 {
	profit := 0.0
	for _, value := range append(s.Win, s.Lose...) {
		profit += value
	}
	return profit
}

func (s summary) SQN() float64 {
	total := float64(len(s.Win) + len(s.Lose))
	avgProfit := s.Profit() / total
	stdDev := 0.0
	for _, profit := range append(s.Win, s.Lose...) {
		stdDev += math.Pow(profit-avgProfit, 2)
	}
	stdDev = math.Sqrt(stdDev / total)
	return math.Sqrt(total) * (s.Profit() / total) / stdDev
}

func (s summary) Payoff() float64 {
	avgWin := 0.0
	avgLose := 0.0

	for _, value := range s.Win {
		avgWin += value
	}

	for _, value := range s.Lose {
		avgLose += value
	}

	if len(s.Win) == 0 || len(s.Lose) == 0 || avgLose == 0 {
		return 0
	}

	return (avgWin / float64(len(s.Win))) / math.Abs(avgLose/float64(len(s.Lose)))
}

func (s summary) WinPercentage() float64 {
	if len(s.Win)+len(s.Lose) == 0 {
		return 0
	}

	return float64(len(s.Win)) / float64(len(s.Win)+len(s.Lose)) * 100
}

func (s summary) String() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	_, quote := exchange.SplitAssetQuote(s.Pair)
	data := [][]string{
		{"Coin", s.Pair},
		{"Trades", strconv.Itoa(len(s.Lose) + len(s.Win))},
		{"Win", strconv.Itoa(len(s.Win))},
		{"Loss", strconv.Itoa(len(s.Lose))},
		{"% Win", fmt.Sprintf("%.1f", s.WinPercentage())},
		{"Payoff", fmt.Sprintf("%.1f", s.Payoff()*100)},
		{"Profit", fmt.Sprintf("%.4f %s", s.Profit(), quote)},
		{"Volume", fmt.Sprintf("%.4f %s", s.Volume, quote)},
	}
	table.AppendBulk(data)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})
	table.Render()
	return tableString.String()
}
