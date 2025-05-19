package symbol_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/achannarasappa/ticker/v4/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Symbol", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetTickerSymbols", func() {

		It("should get ticker symbols", func() {
			// Set up mock response
			responseFixture := `"BTC.X","BTC-USDC","cb"
"XRP.X","XRP-USDC","cb"
"ETH.X","ETH-USD","cb"
"SOL.X","SOL-USD","cb"
"SUI.X","SUI-USD","cb"
`
			server.RouteToHandler("GET", "/symbols.csv",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/symbols.csv"),
					ghttp.RespondWith(http.StatusOK, responseFixture, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
				),
			)

			expectedSymbols := symbol.TickerSymbolToSourceSymbol{
				"BTC.X": symbol.SymbolSourceMap{
					TickerSymbol: "BTC.X",
					SourceSymbol: "BTC-USDC",
					Source:       c.QuoteSourceCoinbase,
				},
				"XRP.X": symbol.SymbolSourceMap{
					TickerSymbol: "XRP.X",
					SourceSymbol: "XRP-USDC",
					Source:       c.QuoteSourceCoinbase,
				},
				"ETH.X": symbol.SymbolSourceMap{
					TickerSymbol: "ETH.X",
					SourceSymbol: "ETH-USD",
					Source:       c.QuoteSourceCoinbase,
				},
				"SOL.X": symbol.SymbolSourceMap{
					TickerSymbol: "SOL.X",
					SourceSymbol: "SOL-USD",
					Source:       c.QuoteSourceCoinbase,
				},
				"SUI.X": symbol.SymbolSourceMap{
					TickerSymbol: "SUI.X",
					SourceSymbol: "SUI-USD",
					Source:       c.QuoteSourceCoinbase,
				},
			}

			outputSymbols, outputErr := symbol.GetTickerSymbols(server.URL() + "/symbols.csv")

			Expect(outputSymbols).To(Equal(expectedSymbols))
			Expect(outputErr).To(BeNil())
		})

		When("a ticker symbol has an unknown source", func() {

			It("should get ticker symbols", func() {
				// Set up mock response
				responseFixture := `"SOMESYMBOL.X","some-symbol","uk"
`
				server.RouteToHandler("GET", "/symbols.csv",
					ghttp.CombineHandlers(
						ghttp.RespondWith(http.StatusOK, responseFixture, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
					),
				)

				expectedSymbols := symbol.TickerSymbolToSourceSymbol{
					"SOMESYMBOL.X": symbol.SymbolSourceMap{
						TickerSymbol: "SOMESYMBOL.X",
						SourceSymbol: "some-symbol",
						Source:       c.QuoteSourceUnknown,
					},
				}

				outputSymbols, outputErr := symbol.GetTickerSymbols(server.URL() + "/symbols.csv")

				Expect(outputSymbols).To(Equal(expectedSymbols))
				Expect(outputErr).To(BeNil())
			})

		})

		When("a malformed CSV is returned", func() {

			It("should get ticker symbols", func() {
				// Set up mock response
				responseFixture := `"SOMESYMBOL.X","some-symbol","uk"
"SOMESYMBOL.X","some-symbol","uk", "abc"
"test"
`
				server.RouteToHandler("GET", "/symbols.csv",
					ghttp.CombineHandlers(
						ghttp.RespondWith(http.StatusOK, responseFixture, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
					),
				)

				_, outputErr := symbol.GetTickerSymbols(server.URL() + "/symbols.csv")

				Expect(outputErr).ToNot(BeNil())
			})

		})

		When("there is a server error", func() {

			It("returns the error", func() {
				// Set up mock response for server error
				server.RouteToHandler("GET", "/symbols.csv",
					ghttp.CombineHandlers(
						ghttp.RespondWith(http.StatusInternalServerError, "", nil),
					),
				)

				_, outputErr := symbol.GetTickerSymbols(server.URL() + "/symbols.csv")

				Expect(outputErr).ToNot(BeNil())
			})

		})

	})

})
