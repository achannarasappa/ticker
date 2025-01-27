package coinbase_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/v4/internal/quote/coinbase"
)

var _ = Describe("Coinbase Quote", func() {

	Describe("GetUnderlyingAssetSymbols", func() {
		It("should return the underlying asset symbols for a futures contract", func() {
			responseFixture := `{
				"products": [
					{
						"product_id": "BIT-31JAN25-CDE",
						"product_type": "FUTURE",
						"future_product_details": {
							"contract_root_unit": "BTC"
						}
					}
				]
			}`
			responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BIT\-31JAN25\-CDE.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output, err := GetUnderlyingAssetSymbols(*client, []string{"BIT-31JAN25-CDE"})
			Expect(output).To(Equal([]string{"BTC-USD"}))
			Expect(err).To(BeNil())
		})

		When("the asset is not a futures contract", func() {
			It("should ignore the asset", func() {
				responseFixture := `{
					"products": [
						{
							"product_id": "BTC-USD",
							"product_type": "SPOT"
						}
					]
				}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output, err := GetUnderlyingAssetSymbols(*client, []string{"BTC-USD"})
				Expect(output).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})

	})
})
