package asset_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/internal/asset"
	c "github.com/achannarasappa/ticker/internal/common"
)

var _ = Describe("Asset", func() {

	Describe("GetAssets", func() {
		It("should return assets", func() {
			inputContext := c.Context{}
			expectedAssets := fixtureAssets
			expectedHoldingSummary := HoldingSummary{}
			outputAssets, outputHoldingSummary := GetAssets(inputContext, fixtureAssetQuotes)

			Expect(outputAssets).To(Equal(expectedAssets))
			Expect(outputHoldingSummary).To(Equal(expectedHoldingSummary))
		})
	})
})
