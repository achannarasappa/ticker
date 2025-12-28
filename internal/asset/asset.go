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
	lotsBySymbol := getLots(lots)
	orderIndex := make(map[string]int)

	for i, lot := range lots {
		if _, exists := orderIndex[lot.Symbol]; !exists {
			orderIndex[strings.ToLower(lot.Symbol)] = i
		}
	}

	for i, symbol := range assetGroupQuote.AssetGroup.ConfigAssetGroup.Watchlist {
		if _, exists := orderIndex[symbol]; !exists {
			orderIndex[strings.ToLower(symbol)] = i + len(lots)
		}
	}

	for _, assetQuote := range assetGroupQuote.AssetQuotes {

		currencyRateByUse := getCurrencyRateByUse(ctx, assetQuote.Currency.FromCurrencyCode, assetQuote.Currency.ToCurrencyCode, assetQuote.Currency.Rate)

		position := getPositionFromAssetQuote(assetQuote, lotsBySymbol, currencyRateByUse)
		positionSummary = addPositionToPositionSummary(positionSummary, position, currencyRateByUse)

		assets = append(assets, c.Asset{
			Name:   assetQuote.Name,
			Symbol: assetQuote.Symbol,
			Class:  assetQuote.Class,
			Currency: c.Currency{
				FromCurrencyCode: assetQuote.Currency.FromCurrencyCode,
				ToCurrencyCode:   currencyRateByUse.ToCurrencyCode,
			},
			Position:      position,
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

	assets = updatePositionWeights(assets, positionSummary)

	return assets, positionSummary

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

func updatePositionWeights(assets []c.Asset, positionSummary PositionSummary) []c.Asset {

	if positionSummary.Value == 0 {
		return assets
	}

	for i, asset := range assets {
		assets[i].Position.Weight = (asset.Position.Value / positionSummary.Value) * 100
	}

	return assets

}

func getPositionFromAssetQuote(assetQuote c.AssetQuote, lotsBySymbol map[string]AggregatedLot, currencyRateByUse currencyRateByUse) c.Position {

	if aggregatedLot, ok := lotsBySymbol[assetQuote.Symbol]; ok {
		value := aggregatedLot.Quantity * assetQuote.QuotePrice.Price * currencyRateByUse.QuotePrice
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
				Amount:  assetQuote.QuotePrice.Change * aggregatedLot.Quantity * currencyRateByUse.QuotePrice,
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
