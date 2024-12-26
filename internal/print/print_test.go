package print_test

import (
	"io/ioutil"
	"os"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/print"
	. "github.com/achannarasappa/ticker/v4/test/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		inputOptions      = print.Options{}
		inputContext      = c.Context{}
		inputDependencies = c.Dependencies{
			HttpClients: c.DependenciesHttpClients{
				Default: client,
				Yahoo:   client,
			},
		}
	)

	BeforeEach(func() {
		MockResponseYahooQuotes()
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
				Expect(output).To(Equal("name,symbol,price,value,cost,quantity,weight\nAlphabet Inc.,GOOG,2838.42,28384,10000,10.000,96.997\nRoblox Corporation,RBLX,87.880,878.80,500.00,10.000,3.0031\n\n"))
			})
		})
	})

	Describe("RunSummary", func() {

		It("should print the holdings summary in JSON format", func() {
			output := getStdout(func() {
				print.RunSummary(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal("{\"total_value\":\"29263.000000\",\"total_cost\":\"10500.000000\",\"day_change_amount\":\"-583.200992\",\"day_change_percent\":\"-1.992964\",\"total_change_amount\":\"18763.000000\",\"total_change_percent\":\"178.695238\"}\n"))
		})

		When("the format option is set to csv", func() {
			It("should print the holdings summary in CSV format", func() {
				inputOptions := print.Options{
					Format: "csv",
				}
				output := getStdout(func() {
					print.RunSummary(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
				})
				Expect(output).To(Equal("total_value,total_cost,day_change_amount,day_change_percent,total_change_amount,total_change_percent\n29263.000000,10500.000000,-583.200992,-1.992964,18763.000000,178.695238\n\n"))
			})
		})

	})

})
