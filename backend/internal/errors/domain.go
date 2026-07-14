package errorss

import "errors"

var (
	ErrTickerNotFound     = errors.New("ticker not found in cache")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInsufficientShares = errors.New("insufficient shares")
	ErrPositionNotFound   = errors.New("position not found")
	ErrPortfolioNotFound  = errors.New("portfolio not found")
	ErrMarketClosed       = errors.New("market closed")
)
