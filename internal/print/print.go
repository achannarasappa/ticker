package print //nolint:predeclared

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/achannarasappa/ticker/internal/asset"
	c "github.com/achannarasappa/ticker/internal/common"
	quote "github.com/achannarasappa/ticker/internal/quote"
	"github.com/achannarasappa/ticker/internal/ui/util"

	"github.com/spf13/cobra"
)

// Options to configure print behavior
type Options struct {
	Format string
}

type jsonRow struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Price    string `json:"price"`
	Value    string `json:"value"`
	Cost     string `json:"cost"`
	Quantity string `json:"quantity"`
	Weight   string `json:"weight"`
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
				Price:    fmt.Sprintf("%f", asset.QuotePrice.Price),
				Value:    fmt.Sprintf("%f", asset.Holding.Value),
				Cost:     fmt.Sprintf("%f", asset.Holding.Cost),
				Quantity: fmt.Sprintf("%f", asset.Holding.Quantity),
				Weight:   fmt.Sprintf("%f", asset.Holding.Weight),
			})
		}
	}

	out, err := json.Marshal(rows)

	if err != nil {
		return err.Error()
	}

	return string(out)

}

// Run prints holdings to the terminal
func Run(dep *c.Dependencies, ctx *c.Context, options *Options) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {

		assetGroupQuote := quote.GetAssetGroupQuote(*dep)(ctx.Groups[0])
		assets, _ := asset.GetAssets(*ctx, assetGroupQuote)

		if options.Format == "csv" {
			fmt.Println(convertAssetsToCSV(assets))

			return
		}

		fmt.Println(convertAssetsToJSON(assets))
	}
}
