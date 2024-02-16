package symbol_test

import (
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/achannarasappa/ticker/v4/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	h "github.com/achannarasappa/ticker/v4/test/http"
)

var _ = Describe("Symbol", func() {

	BeforeEach(func() {

		h.MockTickerSymbols()

	})

	Describe("GetTickerSymbols", func() {

		It("should get ticker symbols", func() {

			expectedSymbols := symbol.TickerSymbolToSourceSymbol{
				"ADA.X": symbol.SymbolSourceMap{
					TickerSymbol: "ADA.X",
					SourceSymbol: "cardano",
					Source:       c.QuoteSourceCoingecko,
				},
				"ALGO.X": symbol.SymbolSourceMap{
					TickerSymbol: "ALGO.X",
					SourceSymbol: "algorand",
					Source:       c.QuoteSourceCoingecko,
				},
				"BTC.X": symbol.SymbolSourceMap{
					TickerSymbol: "BTC.X",
					SourceSymbol: "bitcoin",
					Source:       c.QuoteSourceCoingecko,
				},
				"ETH.X": symbol.SymbolSourceMap{
					TickerSymbol: "ETH.X",
					SourceSymbol: "ethereum",
					Source:       c.QuoteSourceCoingecko,
				},
				"DOGE.X": symbol.SymbolSourceMap{
					TickerSymbol: "DOGE.X",
					SourceSymbol: "dogecoin",
					Source:       c.QuoteSourceCoingecko,
				},
				"DOT.X": symbol.SymbolSourceMap{
					TickerSymbol: "DOT.X",
					SourceSymbol: "polkadot",
					Source:       c.QuoteSourceCoingecko,
				},
				"SOL.X": symbol.SymbolSourceMap{
					TickerSymbol: "SOL.X",
					SourceSymbol: "solana",
					Source:       c.QuoteSourceCoingecko,
				},
				"USDC.X": symbol.SymbolSourceMap{
					TickerSymbol: "USDC.X",
					SourceSymbol: "usd-coin",
					Source:       c.QuoteSourceCoingecko,
				},
				"XRP.X": symbol.SymbolSourceMap{
					TickerSymbol: "XRP.X",
					SourceSymbol: "ripple",
					Source:       c.QuoteSourceCoingecko,
				},
			}

			outputSymbols, outputErr := symbol.GetTickerSymbols(*client)

			Expect(outputSymbols).To(Equal(expectedSymbols))
			Expect(outputErr).To(BeNil())
		})

		When("a ticker symbol has an unknown source", func() {

			It("should get ticker symbols", func() {

				responseFixture := `"SOMESYMBOL.X","some-symbol","uk"
`
				responseUrl := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
					return resp, nil
				})

				expectedSymbols := symbol.TickerSymbolToSourceSymbol{
					"SOMESYMBOL.X": symbol.SymbolSourceMap{
						TickerSymbol: "SOMESYMBOL.X",
						SourceSymbol: "some-symbol",
						Source:       c.QuoteSourceUnknown,
					},
				}

				outputSymbols, outputErr := symbol.GetTickerSymbols(*client)

				Expect(outputSymbols).To(Equal(expectedSymbols))
				Expect(outputErr).To(BeNil())
			})

		})

		When("a malformed CSV is returned", func() {

			It("should get ticker symbols", func() {

				responseFixture := `"SOMESYMBOL.X","some-symbol","uk"
"SOMESYMBOL.X","some-symbol","uk", "abc"
"test"
`
				responseUrl := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
					return resp, nil
				})

				_, outputErr := symbol.GetTickerSymbols(*client)

				Expect(outputErr).ToNot(BeNil())
			})

		})

		When("there is a server error", func() {

			It("returns the error", func() {

				h.MockTickerSymbolsError()

				_, outputErr := symbol.GetTickerSymbols(*client)

				Expect(outputErr).ToNot(BeNil())

			})

		})

	})

})
