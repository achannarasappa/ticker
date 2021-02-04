package quote_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xeipuuv/gojsonschema"

	. "github.com/onsi/ginkgo"
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
							"symbol": {
								"type": "string"
							}
						}
					}
				},
				"required": ["quoteResponse"]
			  }`

			resp, err := http.Get("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			bodyString := string(body)
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
