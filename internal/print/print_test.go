package print_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
	"github.com/achannarasappa/ticker/v4/internal/print"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/spf13/cobra"
)

func getStdout(fn func()) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out)
}

var _ = Describe("Print", func() {

	var (
		server            *ghttp.Server
		inputOptions      = print.Options{}
		inputContext      = c.Context{}
		inputDependencies c.Dependencies
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

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

		server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
			ghttp.CombineHandlers(
				func(w http.ResponseWriter, r *http.Request) {
					query := r.URL.Query()
					fields := query.Get("fields")

					if fields == "regularMarketPrice,currency" {
						json.NewEncoder(w).Encode(currencyResponseFixture)
					} else {
						json.NewEncoder(w).Encode(quoteCloudflareFixture)
					}
				},
			),
		)

		inputDependencies = c.Dependencies{
			SymbolsURL:                       server.URL() + "/symbols.csv",
			MonitorYahooBaseURL:              server.URL(),
			MonitorYahooSessionRootURL:       server.URL(),
			MonitorYahooSessionCrumbURL:      server.URL(),
			MonitorYahooSessionConsentURL:    server.URL(),
			MonitorPriceCoinbaseBaseURL:      server.URL(),
			MonitorPriceCoinbaseStreamingURL: server.URL(),
		}

		inputContext = c.Context{
			Groups: []c.AssetGroup{
				{
					SymbolsBySource: []c.AssetGroupSymbolsBySource{
						{
							Source: c.QuoteSourceYahoo,
							Symbols: []string{
								"GOOG",
								"RBLX",
							},
						},
					},
					ConfigAssetGroup: c.ConfigAssetGroup{
						Holdings: []c.Lot{
							{
								Symbol:   "GOOG",
								UnitCost: 1000,
								Quantity: 10,
							},
							{
								Symbol:   "RBLX",
								UnitCost: 50,
								Quantity: 10,
							},
						},
					},
				},
			},
		}
	})

	Describe("Run", func() {

		It("should print holdings in JSON format", func() {

			output := getStdout(func() {
				print.Run(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal("[{\"name\":\"Alphabet Inc.\",\"symbol\":\"GOOG\",\"price\":\"2838.420000\",\"value\":\"28384.200000\",\"cost\":\"10000.000000\",\"quantity\":\"10.000000\",\"weight\":\"96.996890\"},{\"name\":\"Roblox Corporation\",\"symbol\":\"RBLX\",\"price\":\"87.880000\",\"value\":\"878.800000\",\"cost\":\"500.000000\",\"quantity\":\"10.000000\",\"weight\":\"3.003110\"}]\n"))
		})

		When("there are no holdings in the default group", func() {
			BeforeEach(func() {
				inputContext.Groups[0].ConfigAssetGroup.Holdings = []c.Lot{}
			})

			It("should print an empty array", func() {

				output := getStdout(func() {
					print.Run(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
				})
				Expect(output).To(Equal("[]\n"))
			})
		})

		When("the format option is set to csv", func() {
			It("should print the holdings in CSV format", func() {
				inputOptions := print.Options{
					Format: "csv",
				}
				output := getStdout(func() {
					print.Run(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
				})
				Expect(output).To(Equal("name,symbol,price,value,cost,quantity,weight\nAlphabet Inc.,GOOG,2838.42,28384.20,10000.00,10.000,96.997\nRoblox Corporation,RBLX,87.880,878.80,500.00,10.000,3.0031\n\n"))
			})
		})
	})

	Describe("RunSummary", func() {

		It("should print the holdings summary in JSON format", func() {
			output := getStdout(func() {
				print.RunSummary(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal("{\"total_value\":\"29263.000000\",\"total_cost\":\"10500.000000\",\"day_change_amount\":\"2750.500000\",\"day_change_percent\":\"9.399241\",\"total_change_amount\":\"18763.000000\",\"total_change_percent\":\"178.695238\"}\n"))
		})

		When("the format option is set to csv", func() {
			It("should print the holdings summary in CSV format", func() {
				inputOptions := print.Options{
					Format: "csv",
				}
				output := getStdout(func() {
					print.RunSummary(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
				})
				Expect(output).To(Equal("total_value,total_cost,day_change_amount,day_change_percent,total_change_amount,total_change_percent\n29263.000000,10500.000000,2750.500000,9.399241,18763.000000,178.695238\n\n"))
			})
		})

	})

})

var currencyResponseFixture = unary.Response{
	QuoteResponse: unary.ResponseQuoteResponse{
		Quotes: []unary.ResponseQuote{
			{
				Currency: "USD",
				Symbol:   "RBLX",
			},
			{
				Currency: "USD",
				Symbol:   "GOOG",
			},
		},
	},
}

var quoteCloudflareFixture = unary.Response{
	QuoteResponse: unary.ResponseQuoteResponse{
		Quotes: []unary.ResponseQuote{
			{
				ShortName:                  "Alphabet Inc.",
				Symbol:                     "GOOG",
				MarketState:                "REGULAR",
				Currency:                   "USD",
				RegularMarketPrice:         unary.ResponseFieldFloat{Raw: 2838.42, Fmt: "2838.42"},
				RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 10.00, Fmt: "10.00"},
				RegularMarketChange:        unary.ResponseFieldFloat{Raw: 283.84, Fmt: "283.84"},
			},
			{
				ShortName:                  "Roblox Corporation",
				Symbol:                     "RBLX",
				MarketState:                "REGULAR",
				Currency:                   "USD",
				RegularMarketPrice:         unary.ResponseFieldFloat{Raw: 87.88, Fmt: "87.88"},
				RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: -10.00, Fmt: "-10.00"},
				RegularMarketChange:        unary.ResponseFieldFloat{Raw: -8.79, Fmt: "-8.79"},
			},
		},
	},
}
