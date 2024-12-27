package coinbase

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

const (
	productTypeFuture = "FUTURE"
)

// ResponseQuote represents a quote of a single product from the Coinbase API
type ResponseQuote struct {
	Symbol                   string `json:"base_display_symbol"`
	ProductID                string `json:"product_id"`
	ShortName                string `json:"base_name"`
	Price                    string `json:"price"`
	PriceChange24H           string `json:"price_percentage_change_24h"`
	Volume24H                string `json:"volume_24h"`
	DisplayName              string `json:"display_name"`
	MarketState              string `json:"status"`
	Currency                 string `json:"quote_currency_id"`
	ExchangeName             string `json:"product_venue"`
	FcmTradingSessionDetails struct {
		IsSessionOpen bool `json:"is_session_open"`
	} `json:"fcm_trading_session_details"`
	FutureProductDetails struct {
		ContractDisplayName string `json:"contract_display_name"`
		GroupDescription    string `json:"group_description"`
		ContractRootUnit    string `json:"contract_root_unit"`
		ExpirationDate      string `json:"contract_expiry"`
		ExpirationTimezone  string `json:"expiration_timezone"`
	} `json:"future_product_details"`
	ProductType string `json:"product_type"`
}

// Response represents the container object from the API response
type Response struct {
	Products []ResponseQuote `json:"products"`
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

func transformResponseQuote(responseQuote ResponseQuote, responseQuoteUnderlying ResponseQuote) c.AssetQuote {
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
			SymbolUnderlying: responseQuote.FutureProductDetails.ContractRootUnit,
			Expiry:           formatExpiry(expirationDate),
		}

		// If there is a quote for the underlying asset, calculate the index price and basis
		if responseQuoteUnderlying != (ResponseQuote{}) {
			priceUnderlying, _ := strconv.ParseFloat(responseQuoteUnderlying.Price, 64)
			quoteFutures.IndexPrice = priceUnderlying
			quoteFutures.Basis = (priceUnderlying - price) / price
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
		},
	}
}

func transformResponseQuotes(symbols []string, responseQuotes []ResponseQuote) []c.AssetQuote {
	quotes := make([]c.AssetQuote, 0)
	responseQuotesBySymbol := make(map[string]ResponseQuote)
	symbolsMap := make(map[string]bool)

	// Create map of explicitly requested symbols
	for _, symbol := range symbols {
		symbolsMap[symbol] = true
	}

	// Index quotes by symbol
	for _, responseQuote := range responseQuotes {
		responseQuotesBySymbol[responseQuote.ProductID] = responseQuote
	}

	// Transform quotes
	for _, responseQuote := range responseQuotes {

		// Skip quotes only used for lookup purposes (futures contracts underlying asset quote)
		if _, exists := symbolsMap[responseQuote.ProductID]; !exists {
			continue
		}

		responseQuoteUnderlying := ResponseQuote{}
		symbolUnderlying := responseQuote.FutureProductDetails.ContractRootUnit + "-USD"

		// Lookup underlying asset quote for futures contracts
		if responseQuote.ProductType == productTypeFuture {

			// Skip futures contracts without a price quote (underlying assets of some CDE contracts are not listed on CBE)
			if _, exists := responseQuotesBySymbol[symbolUnderlying]; exists {
				responseQuoteUnderlying = responseQuotesBySymbol[symbolUnderlying]
			}
		}

		quotes = append(quotes, transformResponseQuote(responseQuote, responseQuoteUnderlying))
	}

	return quotes
}

// GetAssetQuotes issues a HTTP request to retrieve quotes from the Coinbase API
func GetAssetQuotes(client resty.Client, symbols []string, symbolsUnderlying []string) []c.AssetQuote {

	symbolsMerged := make([]string, 0)
	symbolsMerged = append(symbolsMerged, symbolsUnderlying...)
	symbolsMerged = append(symbolsMerged, symbols...)
	slices.Sort(symbolsMerged)
	symbolsMerged = slices.Compact(symbolsMerged)

	res, _ := client.R().
		SetResult(Response{}).
		SetQueryParamsFromValues(url.Values{"product_ids": symbolsMerged}).
		Get("https://api.coinbase.com/api/v3/brokerage/market/products")

	return transformResponseQuotes(symbols, res.Result().(*Response).Products) //nolint:forcetypeassert
}
