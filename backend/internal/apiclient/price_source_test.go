package apiclient_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Naitik2411/stockit/internal/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPriceSource struct {
	quotes map[string]struct{ price, change string }
	err    error
}

func (m *mockPriceSource) GetQuote(_ context.Context, ticker string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	q, ok := m.quotes[ticker]
	if !ok {
		return "", "", errors.New("not found")
	}
	return q.price, q.change, nil
}

func TestMockPriceSource_GetQuote(t *testing.T) {
	var _ apiclient.PriceSource = (*mockPriceSource)(nil)

	m := &mockPriceSource{
		quotes: map[string]struct{ price, change string }{
			"AAPL": {"178.50", "1.25"},
		},
	}

	price, change, err := m.GetQuote(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "178.50", price)
	assert.Equal(t, "1.25", change)

	_, _, err = m.GetQuote(context.Background(), "MSFT")
	require.Error(t, err)

	m.err = errors.New("api down")
	_, _, err = m.GetQuote(context.Background(), "AAPL")
	require.Error(t, err)
}
