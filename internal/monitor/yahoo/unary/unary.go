package unary

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
)

//nolint:gochecknoglobals
var (
	postMarketStatuses = map[string]bool{"POST": true, "POSTPOST": true}
)

// UnaryAPI is a client for the API
type UnaryAPI struct {
	client  *http.Client
	baseURL string
}

// NewUnaryAPI creates a new client
func NewUnaryAPI(baseURL string) *UnaryAPI {
	return &UnaryAPI{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

// GetAssetQuotes issues a HTTP request to retrieve quotes from the API and process the response
func (u *UnaryAPI) GetAssetQuotes(symbols []string) ([]c.AssetQuote, map[string]*c.AssetQuote, error) {
	if len(symbols) == 0 {
		return []c.AssetQuote{}, make(map[string]*c.AssetQuote), errors.New("no symbols provided")
	}

	// Build URL with query parameters
	reqURL, _ := url.Parse(u.baseURL + "/v7/finance/quote")
	q := reqURL.Query()
	q.Set("fields", "shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap")
	q.Set("symbols", strings.Join(symbols, ","))
	reqURL.RawQuery = q.Encode()

	// Make request
	resp, err := u.client.Get(reqURL.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Decode response
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	quotes, quotesBySymbol := transformResponseQuotes(result.QuoteResponse.Quotes)

	return quotes, quotesBySymbol, nil
}
