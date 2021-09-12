package print

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/achannarasappa/ticker/internal/asset"
	c "github.com/achannarasappa/ticker/internal/common"
	quote "github.com/achannarasappa/ticker/internal/quote/yahoo"
	"github.com/achannarasappa/ticker/internal/ui/util"

	"github.com/spf13/cobra"
)

// Options to configure print behavior
type Options struct {
	Format string
}

type jsonRow struct {
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Value    float64 `json:"value"`
	Cost     float64 `json:"cost"`
	Quantity float64 `json:"quantity"`
	Weight   float64 `json:"weight"`
}

func convertAssetsToCSV(assets []c.Asset) string {
	rows := [][]string{
		{"name", "symbol", "price", "value", "cost", "quantity", "weight"},
	}

	for _, asset := range assets {
		if asset.Holding.Quantity > 0 {
			rows = append(rows, []string{
				asset.Name,
				asset.Symbol,
				util.ConvertFloatToString(asset.QuotePrice.Price, true),
				util.ConvertFloatToString(asset.Holding.Value, true),
				util.ConvertFloatToString(asset.Holding.Cost, true),
				util.ConvertFloatToString(asset.Holding.Quantity, true),
				util.ConvertFloatToString(asset.Holding.Weight, true),
			})
		}
	}

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	//nolint:errcheck
	w.WriteAll(rows)

	return b.String()

}

func convertAssetsToJSON(assets []c.Asset) string {
	var rows []jsonRow

	for _, asset := range assets {
		if asset.Holding.Quantity > 0 {
			rows = append(rows, jsonRow{
				Name:     asset.Name,
				Symbol:   asset.Symbol,
				Price:    asset.QuotePrice.Price,
				Value:    asset.Holding.Value,
				Cost:     asset.Holding.Cost,
				Quantity: asset.Holding.Quantity,
				Weight:   asset.Holding.Weight,
			})
		}
	}

	out, _ := json.Marshal(rows)

	return string(out)

}

// RunHolding prints holdings to the terminal
func Run(dep *c.Dependencies, ctx *c.Context, options *Options) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx.Config.ShowHoldings = true
		symbols := asset.GetSymbols(ctx.Config)
		assetQuotes := quote.GetAssetQuotes(*dep.HttpClient, symbols)()
		assets, _ := asset.GetAssets(*ctx, assetQuotes)

		if options.Format == "csv" {
			fmt.Println(convertAssetsToCSV(assets))
			return
		}

		fmt.Println(convertAssetsToJSON(assets))
	}
}
