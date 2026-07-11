package asset

import (
	"strconv"
	"strings"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// AggregatedLot represents a cost basis lot of an asset grouped by symbol
type AggregatedLot struct {
	Symbol     string
	Cost       float64
	Quantity   float64
	OrderIndex int
}

// PositionSummary represents a summary of all asset positions at a point in time
type PositionSummary struct {
	Value       float64
	Cost        float64
	TotalChange c.PositionChange
	DayChange   c.PositionChange
}

// GetAssets returns assets from an asset group quote
func GetAssets(ctx c.Context, assetGroupQuote c.AssetGroupQuote) ([]c.Asset, PositionSummary) {

	lots := assetGroupQuote.AssetGroup.ConfigAssetGroup.Lots

	var positionSummary PositionSummary
	assets := make([]c.Asset, 0)
	summaryValues := make([]float64, 0) // per-asset value expressed in the summary/display currency, used for weights
	lotsBySymbol := getLots(lots)
	optionsBySymbol := getOptions(assetGroupQuote.AssetGroup.ConfigAssetGroup.Options)
	orderIndex := make(map[string]int)

	for i, lot := range lots {
		if _, exists := orderIndex[lot.Symbol]; !exists {
			orderIndex[strings.ToLower(lot.Symbol)] = i
		}
	}

	optionOffset := len(lots)
	for i, option := range assetGroupQuote.AssetGroup.ConfigAssetGroup.Options {
		if _, exists := orderIndex[option.Symbol]; !exists {
			orderIndex[strings.ToLower(option.Symbol)] = optionOffset + i
		}
	}

	watchlistOffset := optionOffset + len(assetGroupQuote.AssetGroup.ConfigAssetGroup.Options)
	for i, symbol := range assetGroupQuote.AssetGroup.ConfigAssetGroup.Watchlist {
		if _, exists := orderIndex[symbol]; !exists {
			orderIndex[strings.ToLower(symbol)] = watchlistOffset + i
		}
	}

	watchlistSet := getWatchlistSet(assetGroupQuote.AssetGroup.ConfigAssetGroup.Watchlist)

	for _, assetQuote := range assetGroupQuote.AssetQuotes {

		currencyRateByUse := getCurrencyRateByUse(ctx, assetQuote.Class, assetQuote.Currency.FromCurrencyCode, assetQuote.Currency.ToCurrencyCode, assetQuote.Currency.Rate)

		position := getPositionFromAssetQuote(assetQuote, lotsBySymbol, currencyRateByUse)
		positionSummary = addPositionToPositionSummary(positionSummary, position, currencyRateByUse)

		currency := c.Currency{
			FromCurrencyCode: assetQuote.Currency.FromCurrencyCode,
			ToCurrencyCode:   currencyRateByUse.ToCurrencyCode,
		}
		quotePrice := convertAssetQuotePriceCurrency(currencyRateByUse, assetQuote.QuotePrice)
		quoteExtended := convertAssetQuoteExtendedCurrency(currencyRateByUse, assetQuote.QuoteExtended)
		meta := c.Meta{
			IsVariablePrecision: assetQuote.Meta.IsVariablePrecision,
			OrderIndex:          orderIndex[strings.ToLower(assetQuote.Symbol)],
		}

		options := optionsBySymbol[assetQuote.Symbol]
		_, isPosition := lotsBySymbol[assetQuote.Symbol]
		isWatchlist := watchlistSet[strings.ToUpper(assetQuote.Symbol)]

		// Emit a non-option row for plain quotes, positions, and watchlist entries.
		// A symbol present solely as an option underlying gets only its option rows.
		if len(options) == 0 || isPosition || isWatchlist {
			assets = append(assets, c.Asset{
				Name:          assetQuote.Name,
				Symbol:        assetQuote.Symbol,
				Class:         assetQuote.Class,
				Currency:      currency,
				Position:      position,
				QuotePrice:    quotePrice,
				QuoteExtended: quoteExtended,
				QuoteFutures:  assetQuote.QuoteFutures,
				QuoteSource:   assetQuote.QuoteSource,
				Exchange:      assetQuote.Exchange,
				Meta:          meta,
			})
			summaryValues = append(summaryValues, position.Value*currencyRateByUse.SummaryValue)
		}

		// Emit one row per option contract on this underlying so multiple options
		// (and a position held on the same underlying) each get their own row.
		for _, option := range options {
			assets = append(assets, c.Asset{
				Name:          optionLabel(option),
				Symbol:        assetQuote.Symbol,
				Class:         c.AssetClassOption,
				Currency:      currency,
				QuotePrice:    quotePrice,
				QuoteExtended: quoteExtended,
				QuoteOption:   computeQuoteOption(option, assetQuote.QuotePrice.Price),
				QuoteSource:   assetQuote.QuoteSource,
				Exchange:      assetQuote.Exchange,
				Meta:          meta,
			})
			summaryValues = append(summaryValues, 0) // options don't contribute a position value/weight
		}

	}

	assets = updatePositionWeights(assets, summaryValues, positionSummary)

	return assets, positionSummary

}

func getWatchlistSet(watchlist []string) map[string]bool {

	watchlistSet := make(map[string]bool, len(watchlist))

	for _, symbol := range watchlist {
		watchlistSet[strings.ToUpper(symbol)] = true
	}

	return watchlistSet
}

// optionLabel builds a compact display name that distinguishes options on the
// same underlying (e.g. "CALL 417.5 07/10"). The row's Symbol stays the underlying.
func optionLabel(option c.Option) string {

	label := strings.ToUpper(option.Type) + " " + strconv.FormatFloat(option.StrikePrice, 'f', -1, 64)

	if expiration := shortExpiration(option.Expiration); expiration != "" {
		label += " " + expiration
	}

	return label
}

func shortExpiration(expiration string) string {

	if date, err := time.Parse("2006-01-02", expiration); err == nil {
		return date.Format("01/02")
	}

	return expiration
}

// calculateChangePercent calculates the percentage change, returning 0 if base is 0 to avoid division by zero
func calculateChangePercent(changeAmount float64, base float64) float64 {
	if base == 0 {
		return 0
	}

	return (changeAmount / base) * 100
}

func addPositionToPositionSummary(positionSummary PositionSummary, position c.Position, currencyRateByUse currencyRateByUse) PositionSummary {

	if position.Value == 0 {
		return positionSummary
	}

	value := positionSummary.Value + (position.Value * currencyRateByUse.SummaryValue)
	cost := positionSummary.Cost + (position.Cost * currencyRateByUse.SummaryCost)
	dayChange := positionSummary.DayChange.Amount + (position.DayChange.Amount * currencyRateByUse.SummaryValue)
	totalChange := value - cost

	totalChangePercent := calculateChangePercent(totalChange, cost)
	dayChangePercent := (dayChange / value) * 100

	return PositionSummary{
		Value: value,
		Cost:  cost,
		TotalChange: c.PositionChange{
			Amount:  totalChange,
			Percent: totalChangePercent,
		},
		DayChange: c.PositionChange{
			Amount:  dayChange,
			Percent: dayChangePercent,
		},
	}
}

func updatePositionWeights(assets []c.Asset, summaryValues []float64, positionSummary PositionSummary) []c.Asset {

	if positionSummary.Value == 0 {
		return assets
	}

	// Weight uses each position's value expressed in the summary/display currency so the numerator and
	// denominator share a currency (e.g. a GBp-quoted position whose value is otherwise left in pence).
	for i := range assets {
		assets[i].Position.Weight = (summaryValues[i] / positionSummary.Value) * 100
	}

	return assets

}

func getPositionFromAssetQuote(assetQuote c.AssetQuote, lotsBySymbol map[string]AggregatedLot, currencyRateByUse currencyRateByUse) c.Position {

	if aggregatedLot, ok := lotsBySymbol[assetQuote.Symbol]; ok {
		// For futures contracts, multiply price by contract size for PnL calculations
		// The displayed price remains unchanged (uses QuotePrice.Price directly)
		priceForPosition := assetQuote.QuotePrice.Price
		changeForPosition := assetQuote.QuotePrice.Change

		if assetQuote.Class == c.AssetClassFuturesContract {
			contractSize := assetQuote.QuoteFutures.ContractSize
			priceForPosition = assetQuote.QuotePrice.Price * contractSize
			changeForPosition = assetQuote.QuotePrice.Change * contractSize
		}

		value := aggregatedLot.Quantity * priceForPosition * currencyRateByUse.QuotePrice
		cost := aggregatedLot.Cost * currencyRateByUse.PositionCost
		totalChangeAmount := value - cost

		totalChangePercent := calculateChangePercent(totalChangeAmount, cost)

		var unitValue, unitCost float64
		if aggregatedLot.Quantity != 0 {
			unitValue = value / aggregatedLot.Quantity
			unitCost = cost / aggregatedLot.Quantity
		}

		return c.Position{
			Value:     value,
			Cost:      cost,
			Quantity:  aggregatedLot.Quantity,
			UnitValue: unitValue,
			UnitCost:  unitCost,
			DayChange: c.PositionChange{
				Amount:  changeForPosition * aggregatedLot.Quantity * currencyRateByUse.QuotePrice,
				Percent: assetQuote.QuotePrice.ChangePercent,
			},
			TotalChange: c.PositionChange{
				Amount:  totalChangeAmount,
				Percent: totalChangePercent,
			},
			Weight: 0,
		}
	}

	return c.Position{}

}

func getLots(lots []c.Lot) map[string]AggregatedLot {

	if lots == nil {
		return map[string]AggregatedLot{}
	}

	aggregatedLots := map[string]AggregatedLot{}

	for i, lot := range lots {

		aggregatedLot, ok := aggregatedLots[lot.Symbol]

		if !ok {

			aggregatedLots[lot.Symbol] = AggregatedLot{
				Symbol:     lot.Symbol,
				Cost:       (lot.UnitCost * lot.Quantity) + lot.FixedCost,
				Quantity:   lot.Quantity,
				OrderIndex: i,
			}

		} else {

			aggregatedLot.Quantity += lot.Quantity
			aggregatedLot.Cost += lot.Quantity * lot.UnitCost

			aggregatedLots[lot.Symbol] = aggregatedLot

		}

	}

	return aggregatedLots
}

func getOptions(options []c.Option) map[string][]c.Option {

	optionsBySymbol := map[string][]c.Option{}

	for _, option := range options {
		optionsBySymbol[option.Symbol] = append(optionsBySymbol[option.Symbol], option)
	}

	return optionsBySymbol
}

// computeQuoteOption derives the display values for a single option contract
// against the current underlying price.
func computeQuoteOption(option c.Option, underlyingPrice float64) c.QuoteOption {

	// Calculate breakeven price
	// For a call: breakeven = strike + premium
	// For a put: breakeven = strike - premium
	breakevenPrice := option.StrikePrice
	switch strings.ToLower(option.Type) {
	case "call":
		breakevenPrice = option.StrikePrice + option.Premium
	case "put":
		breakevenPrice = option.StrikePrice - option.Premium
	}

	return c.QuoteOption{
		StrikePrice:    option.StrikePrice,
		BreakevenPrice: breakevenPrice,
		Type:           option.Type,
		Premium:        option.Premium,
		CurrentPremium: option.CurrentPremium,
		Contracts:      option.Contracts,
		DiffToStrike:   underlyingPrice - option.StrikePrice,
		Expiration:     option.Expiration,
	}
}
