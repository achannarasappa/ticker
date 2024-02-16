package sorter

import (
	"sort"

	c "github.com/achannarasappa/ticker/v4/internal/common"
)

// Sorter represents a function that sorts quotes
type Sorter func([]c.Asset) []c.Asset

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

func sortByUser(assetsIn []c.Asset) []c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]c.Asset, assetCount)
	copy(assets, assetsIn)

	sort.SliceStable(assets, func(i, j int) bool {

		prevIndex := assetCount
		nextIndex := assetCount

		if assets[i].Holding != (c.Holding{}) {
			prevIndex = assets[i].Meta.OrderIndex
		}

		if assets[j].Holding != (c.Holding{}) {
			nextIndex = assets[j].Meta.OrderIndex
		}

		return nextIndex > prevIndex
	})

	return assets

}

func sortByAlpha(assetsIn []c.Asset) []c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]c.Asset, assetCount)
	copy(assets, assetsIn)

	sort.SliceStable(assets, func(i, j int) bool {
		return assets[j].Symbol > assets[i].Symbol
	})

	return assets
}

func sortByValue(assetsIn []c.Asset) []c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	sort.SliceStable(inactiveAssets, func(i, j int) bool {
		return inactiveAssets[j].Holding.Value < inactiveAssets[i].Holding.Value
	})

	sort.SliceStable(activeAssets, func(i, j int) bool {
		return activeAssets[j].Holding.Value < activeAssets[i].Holding.Value
	})

	return append(activeAssets, inactiveAssets...)
}

func sortByChange(assetsIn []c.Asset) []c.Asset {

	assetCount := len(assetsIn)

	if assetCount <= 0 {
		return assetsIn
	}

	assets := make([]c.Asset, assetCount)
	copy(assets, assetsIn)

	activeAssets, inactiveAssets := splitActiveAssets(assets)

	sort.SliceStable(activeAssets, func(i, j int) bool {
		return activeAssets[j].QuotePrice.ChangePercent < activeAssets[i].QuotePrice.ChangePercent
	})

	sort.SliceStable(inactiveAssets, func(i, j int) bool {
		return inactiveAssets[j].QuotePrice.ChangePercent < inactiveAssets[i].QuotePrice.ChangePercent
	})

	return append(activeAssets, inactiveAssets...)

}

func splitActiveAssets(assets []c.Asset) ([]c.Asset, []c.Asset) {

	activeAssets := make([]c.Asset, 0)
	inactiveAssets := make([]c.Asset, 0)

	for _, asset := range assets {
		if asset.Exchange.IsActive {
			activeAssets = append(activeAssets, asset)
		} else {
			inactiveAssets = append(inactiveAssets, asset)
		}
	}

	return activeAssets, inactiveAssets
}
