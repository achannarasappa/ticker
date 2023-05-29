package yahoo_test

import (
	"fmt"

	yahooClient "github.com/achannarasappa/ticker/internal/quote/yahoo/client"

	"github.com/xeipuuv/gojsonschema"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quote", func() {
	Describe("GetQuotes Response", func() {
		It("should have expected fields in the response", func() {
			const responseSchema = `{
				"properties": {
					"quoteResponse": {
						"type": "object",
						"properties": {
							"result": {
								"$ref": "#/definitions/result"
							},
							"error": {
								"type": "null"
							}
						}
					}
				},
				"definitions": {
					"result": {
						"type": "array",
						"items": {
							"$ref": "#/definitions/quote"
						}
					},
					"quote": {
						"properties": {
							"marketState": {
								"type": "string"
							},
							"shortName": {
								"type": "string"
							},
							"regularMarketChange": {
								"type": "number"
							},
							"regularMarketChangePercent": {
								"type": "number"
							},
							"regularMarketPrice": {
								"type": "number"
							},
							"regularMarketTime": {
								"type": "integer"
							},
							"regularMarketPreviousClose": {
								"type": "number"
							},
							"regularMarketOpen": {
								"type": "number"
							},
							"regularMarketDayRange": {
								"type": "string"
							},
							"regularMarketDayHigh": {
								"type": "number"
							},
							"regularMarketDayLow": {
								"type": "number"
							},
							"regularMarketVolume": {
								"type": "number"
							},
							"postMarketChange": {
								"type": "number"
							},
							"postMarketChangePercent": {
								"type": "number"
							},
							"postMarketPrice": {
								"type": "number"
							},
							"preMarketChange": {
								"type": "number"
							},
							"preMarketChangePercent": {
								"type": "number"
							},
							"preMarketPrice": {
								"type": "number"
							},
							"fiftyTwoWeekHigh": {
								"type": "number"
							},
							"fiftyTwoWeekLow": {
								"type": "number"
							},
							"symbol": {
								"type": "string"
							},
							"fullExchangeName": {
								"type": "string"
							},
							"exchangeDataDelayedBy": {
								"type": "number"
							},
							"marketCap": {
								"type": "number"
							},
							"quoteType": {
								"type": "string"
							}
						}
					}
				},
				"required": ["quoteResponse"]
			  }`

			client := yahooClient.New()

			resp, err := client.R().
				SetQueryParam("fields", "shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap").
				SetQueryParam("symbols", "NET").
				Get("/v7/finance/quote")

			if err != nil {
				panic(err)
			}

			bodyString := resp.String()
			expectedSchema := gojsonschema.NewStringLoader(responseSchema)
			actualResponse := gojsonschema.NewStringLoader(bodyString)
			result, err := gojsonschema.Validate(expectedSchema, actualResponse)

			if err != nil {
				panic(err.Error())
			}

			if !result.Valid() {
				fmt.Printf("Expected fields are not present in the response. see errors :\n")
				for _, desc := range result.Errors() {
					fmt.Printf("- %s\n", desc)
				}
			}

			Expect(result.Valid()).To(Equal(true))
			Expect(resp.Status).To(Equal("200 OK"))
		})
	})
})
