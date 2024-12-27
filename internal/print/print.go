package print //nolint:predeclared

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/achannarasappa/ticker/v4/internal/asset"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	quote "github.com/achannarasappa/ticker/v4/internal/quote"
	"github.com/achannarasappa/ticker/v4/internal/ui/util"

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

type jsonSummary struct {
	TotalValue         string `json:"total_value"`
	TotalCost          string `json:"total_cost"`
	DayChangeAmount    string `json:"day_change_amount"`
	DayChangePercent   string `json:"day_change_percent"`
	TotalChangeAmount  string `json:"total_change_amount"`
	TotalChangePercent string `json:"total_change_percent"`
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

	if len(rows) == 0 {
		return "[]"
	}

	out, err := json.Marshal(rows)

	if err != nil {
		return err.Error()
	}

	return string(out)

}

func convertSummaryToJSON(summary asset.HoldingSummary) string {
	row := jsonSummary{
		TotalValue:         fmt.Sprintf("%f", summary.Value),
		TotalCost:          fmt.Sprintf("%f", summary.Cost),
		DayChangeAmount:    fmt.Sprintf("%f", summary.DayChange.Amount),
		DayChangePercent:   fmt.Sprintf("%f", summary.DayChange.Percent),
		TotalChangeAmount:  fmt.Sprintf("%f", summary.TotalChange.Amount),
		TotalChangePercent: fmt.Sprintf("%f", summary.TotalChange.Percent),
	}

	out, err := json.Marshal(row)

	if err != nil {
		return err.Error()
	}

	return string(out)
}

func convertSummaryToCSV(summary asset.HoldingSummary) string {
	rows := [][]string{
		{"total_value", "total_cost", "day_change_amount", "day_change_percent", "total_change_amount", "total_change_percent"},
		{
			fmt.Sprintf("%f", summary.Value),
			fmt.Sprintf("%f", summary.Cost),
			fmt.Sprintf("%f", summary.DayChange.Amount),
			fmt.Sprintf("%f", summary.DayChange.Percent),
			fmt.Sprintf("%f", summary.TotalChange.Amount),
			fmt.Sprintf("%f", summary.TotalChange.Percent),
		},
	}

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	//nolint:errcheck
	w.WriteAll(rows)

	return b.String()
}

// Run prints holdings to the terminal
func Run(dep *c.Dependencies, ctx *c.Context, options *Options) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {

		assetGroupQuote := quote.GetAssetGroupQuote(*dep, ctx.Reference)(ctx.Groups[0])
		assets, _ := asset.GetAssets(*ctx, assetGroupQuote)

		if options.Format == "csv" {
			fmt.Println(convertAssetsToCSV(assets))

			return
		}

		fmt.Println(convertAssetsToJSON(assets))
	}
}

// RunSummary handles the print summary command
func RunSummary(dep *c.Dependencies, ctx *c.Context, options *Options) func(cmd *cobra.Command, args []string) {
	return func(_ *cobra.Command, _ []string) {
		assetGroupQuote := quote.GetAssetGroupQuote(*dep, ctx.Reference)(ctx.Groups[0])
		_, holdingSummary := asset.GetAssets(*ctx, assetGroupQuote)

		if options.Format == "csv" {
			fmt.Println(convertSummaryToCSV(holdingSummary))

			return
		}

		fmt.Println(convertSummaryToJSON(holdingSummary))
	}
}
