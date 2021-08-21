package asset

import (
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
)

func getAssetFixedClass(configAssetClass string) c.AssetClass {

	input := strings.TrimSpace(strings.ToLower(configAssetClass))

	if input == "cash" {
		return c.AssetClassCash
	}

	if input == "private-security" {
		return c.AssetClassPrivateSecurity
	}

	return c.AssetClassUnknown

}

func getAssetsFixed(ctx c.Context) []c.Asset {
	if !ctx.Config.ShowHoldings {
		return []c.Asset{}
	}

	// TODO: Add currency convert
	assetsFixed := make([]c.Asset, 0)

	for _, configHolding := range ctx.Config.Holdings {
		currencyRateByUse := currency.GetCurrencyRateFromContext(ctx, configHolding.Currency)
		currencyCode := currencyRateByUse.ToCurrencyCode

		cost := (configHolding.UnitCost * configHolding.Quantity) + configHolding.FixedCost
		value := configHolding.UnitValue * configHolding.Quantity

		assetsFixed = append(assetsFixed, c.Asset{
			Name:   configHolding.Description,
			Symbol: configHolding.Symbol,
			Class:  getAssetFixedClass(configHolding.Class),
			Currency: c.Currency{
				FromCurrencyCode: configHolding.Currency,
				ToCurrencyCode:   currencyCode,
			},
			Holding: c.Holding{
				Value:     value,
				Cost:      cost,
				Quantity:  configHolding.Quantity,
				UnitValue: configHolding.UnitValue,
				UnitCost:  configHolding.UnitCost,
				Weight:    0,
			},
			Meta: c.Meta{
				IsVariablePrecision: false,
				OrderIndex:          -1,
			},
		})
	}

	return assetsFixed
}
