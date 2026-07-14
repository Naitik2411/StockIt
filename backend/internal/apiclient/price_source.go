package apiclient

import "context"

type PriceSource interface {
	GetQuote(ctx context.Context, ticker string) (price string, changePct string, err error)
}
