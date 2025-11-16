package asset

import (
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// AggregatedLot represents a cost basis lot of an asset grouped by symbol
type AggregatedLot struct {
	Symbol     string
	Cost       float64
	Quantity   float64
	OrderIndex int
}

// HoldingSummary represents a summary of all asset holdings at a point in time
type HoldingSummary struct {
	Value       float64
	Cost        float64
	TotalChange c.HoldingChange
	DayChange   c.HoldingChange
}

// GetAssets returns assets from an asset group quote
func GetAssets(ctx c.Context, assetGroupQuote c.AssetGroupQuote) ([]c.Asset, HoldingSummary) {

	var holdingSummary HoldingSummary
	assets := make([]c.Asset, 0)
	holdingsBySymbol := getLots(assetGroupQuote.AssetGroup.ConfigAssetGroup.Holdings)
	orderIndex := make(map[string]int)

	for i, symbol := range assetGroupQuote.AssetGroup.ConfigAssetGroup.Holdings {
		if _, exists := orderIndex[symbol.Symbol]; !exists {
			orderIndex[strings.ToLower(symbol.Symbol)] = i
		}
	}

	for i, symbol := range assetGroupQuote.AssetGroup.ConfigAssetGroup.Watchlist {
		if _, exists := orderIndex[symbol]; !exists {
			orderIndex[strings.ToLower(symbol)] = i + len(assetGroupQuote.AssetGroup.ConfigAssetGroup.Holdings)
		}
	}

	for _, assetQuote := range assetGroupQuote.AssetQuotes {

		currencyRateByUse := getCurrencyRateByUse(ctx, assetQuote.Currency.FromCurrencyCode, assetQuote.Currency.ToCurrencyCode, assetQuote.Currency.Rate)

		holding := getHoldingFromAssetQuote(assetQuote, holdingsBySymbol, currencyRateByUse)
		holdingSummary = addHoldingToHoldingSummary(holdingSummary, holding, currencyRateByUse)

		assets = append(assets, c.Asset{
			Name:   assetQuote.Name,
			Symbol: assetQuote.Symbol,
			Class:  assetQuote.Class,
			Currency: c.Currency{
				FromCurrencyCode: assetQuote.Currency.FromCurrencyCode,
				ToCurrencyCode:   currencyRateByUse.ToCurrencyCode,
			},
			Holding:       holding,
			QuotePrice:    convertAssetQuotePriceCurrency(currencyRateByUse, assetQuote.QuotePrice),
			QuoteExtended: convertAssetQuoteExtendedCurrency(currencyRateByUse, assetQuote.QuoteExtended),
			QuoteFutures:  assetQuote.QuoteFutures,
			QuoteSource:   assetQuote.QuoteSource,
			Exchange:      assetQuote.Exchange,
			Meta: c.Meta{
				IsVariablePrecision: assetQuote.Meta.IsVariablePrecision,
				OrderIndex:          orderIndex[strings.ToLower(assetQuote.Symbol)],
			},
		})

	}

	assets = updateHoldingWeights(assets, holdingSummary)

	return assets, holdingSummary

}

// calculateChangePercent calculates the percentage change, returning 0 if base is 0 to avoid division by zero
func calculateChangePercent(changeAmount float64, base float64) float64 {
	if base == 0 {
		return 0
	}

	return (changeAmount / base) * 100
}

func addHoldingToHoldingSummary(holdingSummary HoldingSummary, holding c.Holding, currencyRateByUse currencyRateByUse) HoldingSummary {

	if holding.Value == 0 {
		return holdingSummary
	}

	value := holdingSummary.Value + (holding.Value * currencyRateByUse.SummaryValue)
	cost := holdingSummary.Cost + (holding.Cost * currencyRateByUse.SummaryCost)
	dayChange := holdingSummary.DayChange.Amount + (holding.DayChange.Amount * currencyRateByUse.SummaryValue)
	totalChange := value - cost

	totalChangePercent := calculateChangePercent(totalChange, cost)
	dayChangePercent := (dayChange / value) * 100

	return HoldingSummary{
		Value: value,
		Cost:  cost,
		TotalChange: c.HoldingChange{
			Amount:  totalChange,
			Percent: totalChangePercent,
		},
		DayChange: c.HoldingChange{
			Amount:  dayChange,
			Percent: dayChangePercent,
		},
	}
}

func updateHoldingWeights(assets []c.Asset, holdingSummary HoldingSummary) []c.Asset {

	if holdingSummary.Value == 0 {
		return assets
	}

	for i, asset := range assets {
		assets[i].Holding.Weight = (asset.Holding.Value / holdingSummary.Value) * 100
	}

	return assets

}

func getHoldingFromAssetQuote(assetQuote c.AssetQuote, lotsBySymbol map[string]AggregatedLot, currencyRateByUse currencyRateByUse) c.Holding {

	if aggregatedLot, ok := lotsBySymbol[assetQuote.Symbol]; ok {
		value := aggregatedLot.Quantity * assetQuote.QuotePrice.Price * currencyRateByUse.QuotePrice
		cost := aggregatedLot.Cost * currencyRateByUse.PositionCost
		totalChangeAmount := value - cost

		totalChangePercent := calculateChangePercent(totalChangeAmount, cost)

		return c.Holding{
			Value:     value,
			Cost:      cost,
			Quantity:  aggregatedLot.Quantity,
			UnitValue: value / aggregatedLot.Quantity,
			UnitCost:  cost / aggregatedLot.Quantity,
			DayChange: c.HoldingChange{
				Amount:  assetQuote.QuotePrice.Change * aggregatedLot.Quantity * currencyRateByUse.QuotePrice,
				Percent: assetQuote.QuotePrice.ChangePercent,
			},
			TotalChange: c.HoldingChange{
				Amount:  totalChangeAmount,
				Percent: totalChangePercent,
			},
			Weight: 0,
		}
	}

	return c.Holding{}

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
