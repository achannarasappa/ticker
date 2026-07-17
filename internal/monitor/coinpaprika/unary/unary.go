package unary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

const (
	pathTickers = "/v1/tickers/"
	pathSearch  = "/v1/search/"
	// searchResultLimit is the maximum number of search results to request when
	// resolving a ticker symbol to a CoinPaprika coin id; the search endpoint
	// matches by prefix so popular symbols can have many near matches
	searchResultLimit = 250
)

// ResponseQuote represents a quote of a single coin from the CoinPaprika API
type ResponseQuote struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Symbol string         `json:"symbol"`
	Quotes ResponseQuotes `json:"quotes"`
}

// ResponseQuotes represents the price quotes by currency from the CoinPaprika API
type ResponseQuotes struct {
	USD ResponseQuoteUSD `json:"USD"`
}

// ResponseQuoteUSD represents the USD price quote properties from the CoinPaprika API
type ResponseQuoteUSD struct {
	Price            float64 `json:"price"`
	Volume24H        float64 `json:"volume_24h"`
	MarketCap        float64 `json:"market_cap"`
	PercentChange24H float64 `json:"percent_change_24h"`
}

// ResponseSearch represents the container object from the CoinPaprika search API response
type ResponseSearch struct {
	Currencies []ResponseSearchCurrency `json:"currencies"`
}

// ResponseSearchCurrency represents a single currency from the CoinPaprika search API response
type ResponseSearchCurrency struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Rank     int    `json:"rank"`
	IsActive bool   `json:"is_active"`
}

// UnaryAPI is a client for the CoinPaprika REST API
type UnaryAPI struct {
	client  *http.Client
	baseURL string
}

func NewUnaryAPI(baseURL string) *UnaryAPI {
	return &UnaryAPI{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func transformResponseQuote(responseQuote ResponseQuote) c.AssetQuote {
	price := responseQuote.Quotes.USD.Price
	changePercent := responseQuote.Quotes.USD.PercentChange24H

	// Calculate absolute price change from percentage change (same convention as the Coinbase monitor)
	change := price * (changePercent / 100)

	return c.AssetQuote{
		Name:   responseQuote.Name,
		Symbol: strings.ToUpper(responseQuote.Symbol) + ".CP",
		Class:  c.AssetClassCryptocurrency,
		Currency: c.Currency{
			FromCurrencyCode: "USD",
		},
		QuotePrice: c.QuotePrice{
			Price:         price,
			Change:        change,
			ChangePercent: changePercent,
		},
		QuoteExtended: c.QuoteExtended{
			MarketCap: responseQuote.Quotes.USD.MarketCap,
			Volume:    responseQuote.Quotes.USD.Volume24H,
		},
		QuoteSource: c.QuoteSourceCoinPaprika,
		Exchange: c.Exchange{
			Name:                    "CoinPaprika",
			State:                   c.ExchangeStateOpen,
			IsActive:                true, // Aggregated crypto prices are always active
			IsRegularTradingSession: true, // Crypto markets are always in regular session
		},
		Meta: c.Meta{
			IsVariablePrecision: true,
			SymbolInSourceAPI:   responseQuote.ID,
		},
	}
}

// GetAssetQuotes retrieves price quotes for the given CoinPaprika coin ids;
// ids that are not known to the API are skipped
func (u *UnaryAPI) GetAssetQuotes(ids []string) ([]c.AssetQuote, map[string]*c.AssetQuote, error) {
	quotes := make([]c.AssetQuote, 0, len(ids))
	quotesByID := make(map[string]*c.AssetQuote, len(ids))

	if len(ids) == 0 {
		return quotes, quotesByID, nil
	}

	// The CoinPaprika API does not support querying a subset of tickers in a
	// single request so each coin id is requested individually
	for _, id := range ids {
		responseQuote, ok, err := u.getTicker(id)

		if err != nil {
			return nil, nil, err
		}

		// Skip coin ids which are not known to the API rather than failing all quotes
		if !ok {
			continue
		}

		quote := transformResponseQuote(responseQuote)
		quotes = append(quotes, quote)
		quotesByID[quote.Meta.SymbolInSourceAPI] = &quote
	}

	return quotes, quotesByID, nil
}

// GetCoinIDs resolves ticker symbols (e.g. BTC) to CoinPaprika coin ids
// (e.g. btc-bitcoin) by selecting the highest ranked active coin with an
// exactly matching symbol; symbols that cannot be resolved are omitted
func (u *UnaryAPI) GetCoinIDs(symbols []string) (map[string]string, error) {
	symbolToCoinID := make(map[string]string, len(symbols))

	for _, symbol := range symbols {
		reqURL, _ := url.Parse(u.baseURL + pathSearch)
		q := reqURL.Query()
		q.Set("q", symbol)
		q.Set("c", "currencies")
		q.Set("modifier", "symbol_search")
		q.Set("limit", strconv.Itoa(searchResultLimit))
		reqURL.RawQuery = q.Encode()

		resp, err := u.client.Get(reqURL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to make search request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()

			return nil, fmt.Errorf("search request failed with status %d", resp.StatusCode)
		}

		var result ResponseSearch
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to decode search response: %w", err)
		}

		// Select the highest ranked (lowest non-zero rank) active coin with an
		// exactly matching symbol since multiple coins can share a symbol
		bestRank := 0

		for _, currency := range result.Currencies {
			if !currency.IsActive || currency.Rank <= 0 || !strings.EqualFold(currency.Symbol, symbol) {
				continue
			}

			if bestRank == 0 || currency.Rank < bestRank {
				bestRank = currency.Rank
				symbolToCoinID[symbol] = currency.ID
			}
		}
	}

	return symbolToCoinID, nil
}

func (u *UnaryAPI) getTicker(id string) (ResponseQuote, bool, error) {
	// Coin ids are lowercase in the CoinPaprika API and lookups are case-sensitive
	reqURL, _ := url.Parse(u.baseURL + pathTickers + url.PathEscape(strings.ToLower(id)))

	resp, err := u.client.Get(reqURL.String())
	if err != nil {
		return ResponseQuote{}, false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ResponseQuote{}, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return ResponseQuote{}, false, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var responseQuote ResponseQuote
	if err := json.NewDecoder(resp.Body).Decode(&responseQuote); err != nil {
		return ResponseQuote{}, false, fmt.Errorf("failed to decode response: %w", err)
	}

	return responseQuote, true, nil
}
