package sorter

import (
	"cmp"
	"slices"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// Sorter represents a function that sorts quotes
type Sorter func([]*c.Asset) []*c.Asset

// NewSorter creates a sorting function
func NewSorter(sort string) Sorter {
	var sortDict = map[string]Sorter{
		"alpha": sortByAlpha,
		"value": sortByValue,
		"user":  sortByUser,
	}
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	}

	return sortByChange
}

func sortByUser(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(a.Meta.OrderIndex, b.Meta.OrderIndex)
	})

	return assets

}

func sortByAlpha(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	slices.SortStableFunc(assets, func(a, b *c.Asset) int {
		return cmp.Compare(a.Symbol, b.Symbol)
	})

	return assets
}

func sortByValue(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	slices.SortStableFunc(inactiveAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.Position.Value, a.Position.Value)
	})

	slices.SortStableFunc(activeAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.Position.Value, a.Position.Value)
	})

	return append(activeAssets, inactiveAssets...)
}

func sortByChange(assetsIn []*c.Asset) []*c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]*c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	slices.SortStableFunc(activeAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.ChangePercent, a.QuotePrice.ChangePercent)
	})

	slices.SortStableFunc(inactiveAssets, func(a, b *c.Asset) int {
		return cmp.Compare(b.QuotePrice.ChangePercent, a.QuotePrice.ChangePercent)
	})

	return append(activeAssets, inactiveAssets...)

}

func splitActiveAssets(assets []*c.Asset) ([]*c.Asset, []*c.Asset) {

	activeAssets := make([]*c.Asset, 0)
	inactiveAssets := make([]*c.Asset, 0)

	for _, asset := range assets {
		if asset.Exchange.IsActive {
			activeAssets = append(activeAssets, asset)
		} else {
			inactiveAssets = append(inactiveAssets, asset)
		}
	}

	return activeAssets, inactiveAssets
}
