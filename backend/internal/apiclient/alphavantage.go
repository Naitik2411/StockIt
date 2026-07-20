package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AlphaVantageClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAlphaVantageClient(apiKey string) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *AlphaVantageClient) GetQuote(ctx context.Context, ticker string) (string, string, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", ticker, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("alpha vantage request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("alpha vantage status: %d", resp.StatusCode)
	}
	var result struct {
		GlobalQuote map[string]string `json:"Global Quote"`
		Note        string            `json:"Note"`
		Information string            `json:"Information"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode alpha vantage : %w", err)
	}
	if result.Note != "" || result.Information != "" {
		return "", "", fmt.Errorf("alpha vantage rate limited: %s", result.Note+result.Information)
	}
	price := result.GlobalQuote["05. price"]
	changePct := strings.TrimSuffix(result.GlobalQuote["10. change percent"], "%")

	if price == "" {
		return "", "", fmt.Errorf("no price returned for %s", ticker)
	}
	return price, changePct, nil
}
