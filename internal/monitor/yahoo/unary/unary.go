package unary

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net/http"
	"net/url"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
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

type SymbolToCurrency struct {
	Symbol       string
	FromCurrency string
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
		return []c.AssetQuote{}, make(map[string]*c.AssetQuote), nil
	}

	result, err := u.getQuotes(symbols, []string{"shortName", "regularMarketChange", "regularMarketChangePercent", "regularMarketPrice", "regularMarketPreviousClose", "regularMarketOpen", "regularMarketDayRange", "regularMarketDayHigh", "regularMarketDayLow", "regularMarketVolume", "postMarketChange", "postMarketChangePercent", "postMarketPrice", "preMarketChange", "preMarketChangePercent", "preMarketPrice", "fiftyTwoWeekHigh", "fiftyTwoWeekLow", "marketCap"})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	quotes, quotesBySymbol := transformResponseQuotes(result.QuoteResponse.Quotes)

	return quotes, quotesBySymbol, nil
}

// GetCurrencyMap retrieves the currency which the price quote will be denominated in for the given symbols
func (u *UnaryAPI) GetCurrencyMap(symbols []string) (map[string]SymbolToCurrency, error) {
	if len(symbols) == 0 {
		return map[string]SymbolToCurrency{}, nil
	}

	result, err := u.getQuotes(symbols, []string{"regularMarketPrice", "currency"})

	if err != nil {
		return map[string]SymbolToCurrency{}, err
	}

	symbolToCurrency := make(map[string]SymbolToCurrency)

	for _, quote := range result.QuoteResponse.Quotes {
		symbolToCurrency[quote.Symbol] = SymbolToCurrency{
			Symbol:       quote.Symbol,
			FromCurrency: quote.Currency,
		}
	}

	return symbolToCurrency, nil
}

// GetCurrencyRates accepts an array of ISO 4217 currency codes and a target ISO 4217 currency code and returns a conversion rate for each of the input currencies to the target currency
func (u *UnaryAPI) GetCurrencyRates(fromCurrencies []string, toCurrency string) (c.CurrencyRates, error) {
	if toCurrency == "" {
		toCurrency = "USD"
	}

	if len(fromCurrencies) == 0 {
		return c.CurrencyRates{}, nil
	}

	// Create currency pair symbols in format "FROMTO=X" (e.g., "EURUSD=X")
	currencyPairSymbols := make([]string, 0)
	currencyPairSymbolsUnique := make(map[string]bool)

	for _, fromCurrency := range fromCurrencies {

		if fromCurrency == "" {
			continue
		}

		if fromCurrency == toCurrency {
			continue
		}

		pair := strings.ToUpper(fromCurrency) + toCurrency + "=X"

		if _, exists := currencyPairSymbolsUnique[pair]; !exists {
			currencyPairSymbolsUnique[pair] = true
			currencyPairSymbols = append(currencyPairSymbols, pair)
		}
	}

	if len(currencyPairSymbols) == 0 {
		return c.CurrencyRates{}, nil
	}

	// Get quotes for currency pairs
	result, err := u.getQuotes(currencyPairSymbols, []string{"currency", "regularMarketPrice"})
	if err != nil {
		return c.CurrencyRates{}, fmt.Errorf("failed to get currency rates: %w", err)
	}

	// Transform result to currency rates
	currencyRates := make(map[string]c.CurrencyRate)

	// The Yahoo API forces uppercase and so, even though the currency symbol GBpGBP=x is submitted, it is interpreted and returned as
	// GBPGBP=X with Rate = 1.0. So, assume a major to minor ISO 4217 digital currency code conversion when fromCurrency=toCurrency
	for _, quote := range result.QuoteResponse.Quotes {
		fromCurrency := strings.TrimSuffix(strings.TrimSuffix(quote.Symbol, "=X"), toCurrency)
		currencyRates[fromCurrency] = c.CurrencyRate{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         quote.RegularMarketPrice.Raw,
		}
	}

	// If any currency in the map has a minor form then add it to the map as well (with a modified rate.)
	minorCurrencyRates := make(map[string]c.CurrencyRate)
	for fromCurrency, currencyRate := range currencyRates {
		if ok, minorCurrencyCode, minorUnit := MinorUnitForCurrencyCode(fromCurrency); ok {
			minorCurrencyRates[minorCurrencyCode] = c.CurrencyRate{
				FromCurrency: minorCurrencyCode,
				ToCurrency:   currencyRate.ToCurrency,
				Rate:         currencyRate.Rate * math.Pow(10, -minorUnit),
			}
		}
	}

	maps.Copy(currencyRates, minorCurrencyRates)

	return currencyRates, nil
}

func (u *UnaryAPI) getQuotes(symbols []string, fields []string) (Response, error) {

	// Build URL with query parameters
	reqURL, err := url.Parse(u.baseURL + "/v7/finance/quote")
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	q := reqURL.Query()
	q.Set("fields", strings.Join(fields, ","))
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
	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)

	// Set common headers
	req.Header.Set("Authority", "query1.finance.yahoo.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", defaultAcceptLang)
	req.Header.Set("Origin", u.baseURL)
	req.Header.Set("User-Agent", defaultUserAgent)

	// Add cookies if available
	if len(u.cookies) > 0 {
		for _, cookie := range u.cookies {
			req.AddCookie(cookie)
		}
	}

	// Make request
	resp, err := u.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle not ok responses
	if resp.StatusCode >= 400 {
		// Try to refresh session and retry once
		if err := u.refreshSession(); err != nil {
			return Response{}, fmt.Errorf("session refresh failed: %w", err)
		}

		// Retry request with refreshed session
		return u.getQuotes(symbols, fields)
	}

	// Handle unexpected responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode < 400 {
		return Response{}, fmt.Errorf("unexpected response: %d", resp.StatusCode)
	}

	// Decode response
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
