package util

var (
	TimeframePeriods = map[string][]int{
		"1m": []int{5, 10, 20, 60},
		"1h": []int{2, 4, 12, 24},
		"1d": []int{3, 5, 10, 30},
	}
)
