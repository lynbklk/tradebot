package exchange

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInsufficientFunds = errors.New("insufficient funds or locked")
	ErrInvalidAsset      = errors.New("invalid asset")
)

// UserInfo user
type UserInfo struct {
	MakerCommission float64
	TakerCommission float64
}

// OrderError Order
type OrderError struct {
	Err      error
	Pair     string
	Quantity float64
}

func (o *OrderError) Error() string {
	return fmt.Sprintf("order error: %v", o.Err)
}

type AssetQuote struct {
	Quote string
	Asset string
}

var (
	once              sync.Once
	pairAssetQuoteMap = make(map[string]AssetQuote)
)

func SplitAssetQuote(pair string) (asset string, quote string) {
	once.Do(func() {
		client := binance.NewClient("", "")
		info, err := client.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			log.Fatalf("failed to get exchange info: %v", err)
		}

		for _, info := range info.Symbols {
			pairAssetQuoteMap[info.Symbol] = AssetQuote{
				Quote: info.QuoteAsset,
				Asset: info.BaseAsset,
			}
		}
	})

	data := pairAssetQuoteMap[pair]
	return data.Asset, data.Quote
}
