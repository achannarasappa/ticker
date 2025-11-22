package asset

import (
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"fmt"
	"github.com/achannarasappa/ticker/v4/internal/currency"
	// "github.com/achannarasappa/ticker/internal/alert"
)

// AggregatedLot represents a cost basis lot of an asset grouped by symbol
type AggregatedLot struct {
	Symbol     string
	Cost       float64
	Quantity   float64
	OrderIndex int
	TargetPrice *float64
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

	// Evaluate assets and trigger alerts if necessary
    processAssets(assets)

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

func processAssets(assets []c.Asset) {
	for _, asset := range assets {
		if asset.Holding.TargetPrice == nil {
			continue // No target price, so no alert is required
		}

		currentPrice := asset.QuotePrice.Price
		targetPrice := *asset.Holding.TargetPrice

		// Trigger an alert if the current price exceeds the target price
		if currentPrice >= targetPrice {
			// fmt.Printf("📢 Alert: %s has reached the target price! Current: $%.2f, Target: $%.2f\n", asset.Symbol, currentPrice, targetPrice)
		}
	}
}

func getHoldingFromAssetQuote(assetQuote c.AssetQuote, lotsBySymbol map[string]AggregatedLot, currencyRateByUse currency.CurrencyRateByUse) c.Holding {
	if aggregatedLot, ok := lotsBySymbol[assetQuote.Symbol]; ok {
		value := aggregatedLot.Quantity * assetQuote.QuotePrice.Price * currencyRateByUse.QuotePrice
		cost := aggregatedLot.Cost * currencyRateByUse.PositionCost
		totalChangeAmount := value - cost

		totalChangePercent := calculateChangePercent(totalChangeAmount, cost)

		return c.Holding{
			Value:      value,
			Cost:       cost,
			Quantity:   aggregatedLot.Quantity,
			UnitValue:  value / aggregatedLot.Quantity,
			UnitCost:   cost / aggregatedLot.Quantity,
			DayChange: c.HoldingChange{
				Amount:  assetQuote.QuotePrice.Change * aggregatedLot.Quantity * currencyRateByUse.QuotePrice,
				Percent: assetQuote.QuotePrice.ChangePercent,
			},
			TotalChange: c.HoldingChange{
				Amount:  totalChangeAmount,
				Percent: totalChangePercent,
			},
			Weight:      0,
			TargetPrice: aggregatedLot.TargetPrice, // Assign TargetPrice from AggregatedLot
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
				TargetPrice: lot.TargetPrice, // Store the target price from the lot
			}
		} else {
			aggregatedLot.Quantity += lot.Quantity
			aggregatedLot.Cost += lot.Quantity * lot.UnitCost

			// If the TargetPrice is not set, set it from the lot
			if aggregatedLot.TargetPrice == nil && lot.TargetPrice != nil {
				aggregatedLot.TargetPrice = lot.TargetPrice
			}

			aggregatedLots[lot.Symbol] = aggregatedLot
		}
	}

	return aggregatedLots
}

/*
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
*/