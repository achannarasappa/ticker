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

// UnaryAPI is a client for the API
type UnaryAPI struct {
	client            *http.Client
	baseURL           string
	sessionRootURL    string
	sessionCrumbURL   string
	sessionConsentURL string
	cookies           []*http.Cookie
	crumb             string
}

// Config contains configuration options for the UnaryAPI client
type Config struct {
	BaseURL           string
	SessionRootURL    string
	SessionCrumbURL   string
	SessionConsentURL string
}

// NewUnaryAPI creates a new client
func NewUnaryAPI(config Config) *UnaryAPI {
	// Create client with limited redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 1 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return &UnaryAPI{
		client:            client,
		baseURL:           config.BaseURL,
		sessionRootURL:    config.SessionRootURL,
		sessionCrumbURL:   config.SessionCrumbURL,
		sessionConsentURL: config.SessionConsentURL,
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

	// Add common Yahoo Finance query parameters
	q.Set("formatted", "true")
	q.Set("lang", "en-US")
	q.Set("region", "US")
	q.Set("corsDomain", "finance.yahoo.com")

	// Add crumb if available
	if u.crumb != "" {
		q.Set("crumb", u.crumb)
	}

	reqURL.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("authority", "query1.finance.yahoo.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", defaultAcceptLang)
	req.Header.Set("origin", u.sessionRootURL)
	req.Header.Set("user-agent", defaultUserAgent)

	// Add cookies if available
	if len(u.cookies) > 0 {
		for _, cookie := range u.cookies {
			req.AddCookie(cookie)
		}
	}

	// Make request
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		// Try to refresh session and retry once
		if err := u.refreshSession(); err != nil {
			return nil, nil, fmt.Errorf("session refresh failed: %w", err)
		}

		// Retry request with refreshed session
		return u.GetAssetQuotes(symbols)
	}

	// Decode response
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	quotes, quotesBySymbol := transformResponseQuotes(result.QuoteResponse.Quotes)

	return quotes, quotesBySymbol, nil
}
