package util

import (
	"fmt"
	"strings"
)

func PairTimeframeToKey(pair string, timeframe string) string {
	return fmt.Sprintf("%s--%s", pair, timeframe)
}

func PairTimeframeFromKey(key string) (pair, timeframe string) {
	parts := strings.Split(key, "--")
	return parts[0], parts[1]
}
