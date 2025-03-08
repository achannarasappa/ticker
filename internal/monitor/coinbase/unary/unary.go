package unary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
)

const (
	productTypeFuture = "FUTURE"
)

// Response represents the container object from the API response
type Response struct {
	Products []ResponseQuote `json:"products"`
}

// ResponseQuoteFcmTradingSessionDetails represents the trading session details for a product
type ResponseQuoteFcmTradingSessionDetails struct {
	IsSessionOpen bool `json:"is_session_open"`
}

// ResponseQuoteFutureProductDetails represents the details specific to futures contracts
type ResponseQuoteFutureProductDetails struct {
	ContractDisplayName string `json:"contract_display_name"`
	GroupDescription    string `json:"group_description"`
	ContractRootUnit    string `json:"contract_root_unit"`
	ExpirationDate      string `json:"contract_expiry"`
	ExpirationTimezone  string `json:"expiration_timezone"`
	NonCrypto           bool   `json:"non_crypto"`
}

// ResponseQuote represents a quote of a single product from the Coinbase API
type ResponseQuote struct {
	Symbol                   string                                `json:"base_display_symbol"`
	ProductID                string                                `json:"product_id"`
	ShortName                string                                `json:"base_name"`
	Price                    string                                `json:"price"`
	PriceChange24H           string                                `json:"price_percentage_change_24h"`
	Volume24H                string                                `json:"volume_24h"`
	DisplayName              string                                `json:"display_name"`
	MarketState              string                                `json:"status"`
	Currency                 string                                `json:"quote_currency_id"`
	ExchangeName             string                                `json:"product_venue"`
	FcmTradingSessionDetails ResponseQuoteFcmTradingSessionDetails `json:"fcm_trading_session_details"`
	FutureProductDetails     ResponseQuoteFutureProductDetails     `json:"future_product_details"`
	ProductType              string                                `json:"product_type"`
}

type AssetQuotesIndexed struct {
	AssetQuotes            []c.AssetQuote
	AssetQuotesByProductId map[string]*c.AssetQuote
}

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

func formatExpiry(expirationDate time.Time) string {
	now := time.Now()
	diff := expirationDate.Sub(now)
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	if days == 0 {
		return fmt.Sprintf("%dh %dmin", hours, minutes)
	}

	return fmt.Sprintf("%dd %dh", days, hours)
}

func transformResponseQuote(responseQuote ResponseQuote) c.AssetQuote {
	price, _ := strconv.ParseFloat(responseQuote.Price, 64)
	volume, _ := strconv.ParseFloat(responseQuote.Volume24H, 64)
	changePercent, _ := strconv.ParseFloat(responseQuote.PriceChange24H, 64)

	// Calculate absolute price change from percentage change
	change := price * (changePercent / 100)

	name := responseQuote.ShortName
	symbol := responseQuote.Symbol
	isActive := responseQuote.MarketState == "online"
	class := c.AssetClassCryptocurrency
	quoteFutures := c.QuoteFutures{}

	if responseQuote.ProductType == productTypeFuture {

		name = responseQuote.FutureProductDetails.GroupDescription
		symbol = responseQuote.ProductID
		isActive = responseQuote.FcmTradingSessionDetails.IsSessionOpen
		class = c.AssetClassFuturesContract
		expirationTimezone, _ := time.LoadLocation(responseQuote.FutureProductDetails.ExpirationTimezone)
		expirationDate, _ := time.ParseInLocation(time.RFC3339, responseQuote.FutureProductDetails.ExpirationDate, expirationTimezone)

		quoteFutures = c.QuoteFutures{
			SymbolUnderlying: responseQuote.FutureProductDetails.ContractRootUnit + "-USD",
			Expiry:           formatExpiry(expirationDate),
		}
	}

	return c.AssetQuote{
		Name:   name,
		Symbol: symbol,
		Class:  class,
		Currency: c.Currency{
			FromCurrencyCode: strings.ToUpper(responseQuote.Currency),
		},
		QuotePrice: c.QuotePrice{
			Price:         price,
			Change:        change,
			ChangePercent: changePercent,
		},
		QuoteExtended: c.QuoteExtended{
			Volume: volume,
		},
		QuoteFutures: quoteFutures,
		QuoteSource:  c.QuoteSourceCoinbase,
		Exchange: c.Exchange{
			Name:                    responseQuote.ExchangeName,
			State:                   c.ExchangeStateOpen,
			IsActive:                isActive,
			IsRegularTradingSession: true, // Crypto markets are always in regular session
		},
		Meta: c.Meta{
			IsVariablePrecision: true,
			SymbolInSourceAPI:   responseQuote.ProductID,
		},
	}
}

func transformResponseQuotes(responseQuotes []ResponseQuote) ([]c.AssetQuote, map[string]*c.AssetQuote) {
	quotes := make([]c.AssetQuote, 0)
	quotesByProductId := make(map[string]*c.AssetQuote)

	// Transform quotes
	for _, responseQuote := range responseQuotes {

		quote := transformResponseQuote(responseQuote)
		quotes = append(quotes, quote)
		quotesByProductId[quote.Meta.SymbolInSourceAPI] = &quote
	}

	return quotes, quotesByProductId
}

func (u *UnaryAPI) GetAssetQuotes(symbols []string) ([]c.AssetQuote, map[string]*c.AssetQuote, error) {
	if len(symbols) == 0 {
		return nil, nil, nil
	}

	// Build URL with query parameters
	reqURL, _ := url.Parse(u.baseURL + "/api/v3/brokerage/market/products")

	q := reqURL.Query()
	q["product_ids"] = symbols
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

	quotes, quotesByProductId := transformResponseQuotes(result.Products)

	return quotes, quotesByProductId, nil
}
