package yahoo_test

import (
	"fmt"

	yahooClient "github.com/achannarasappa/ticker/v4/internal/quote/yahoo/client"
	"github.com/go-resty/resty/v2"

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
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketChangePercent": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketPrice": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketTime": {
								"$ref": "#/definitions/fieldInteger"
							},
							"regularMarketPreviousClose": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketOpen": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketDayRange": {
								"$ref": "#/definitions/fieldString"
							},
							"regularMarketDayHigh": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketDayLow": {
								"$ref": "#/definitions/fieldNumber"
							},
							"regularMarketVolume": {
								"$ref": "#/definitions/fieldNumber"
							},
							"postMarketChange": {
								"$ref": "#/definitions/fieldNumber"
							},
							"postMarketChangePercent": {
								"$ref": "#/definitions/fieldNumber"
							},
							"postMarketPrice": {
								"$ref": "#/definitions/fieldNumber"
							},
							"preMarketChange": {
								"$ref": "#/definitions/fieldNumber"
							},
							"preMarketChangePercent": {
								"$ref": "#/definitions/fieldNumber"
							},
							"preMarketPrice": {
								"$ref": "#/definitions/fieldNumber"
							},
							"fiftyTwoWeekHigh": {
								"$ref": "#/definitions/fieldNumber"
							},
							"fiftyTwoWeekLow": {
								"$ref": "#/definitions/fieldNumber"
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
								"$ref": "#/definitions/fieldNumber"
							},
							"quoteType": {
								"type": "string"
							}
						}
					},
					"fieldNumber": {
						"properties": {
							"raw": {
								"type": "number"
							},
							"fmt": {
								"type": "string"
							}
						}
					},
					"fieldInteger": {
						"properties": {
							"raw": {
								"type": "integer"
							},
							"fmt": {
								"type": "string"
							}
						}
					},
					"fieldString": {
						"properties": {
							"raw": {
								"type": "string"
							},
							"fmt": {
								"type": "string"
							}
						}
					}
				},
				"required": [
					"quoteResponse"
				]
			}`

			client := yahooClient.New(resty.New(), resty.New())

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
			Expect(resp.StatusCode()).To(Equal(200))
		})
	})
})
