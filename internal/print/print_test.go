package print_test

import (
	"io/ioutil"
	"os"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/print"
	. "github.com/achannarasappa/ticker/test/http"
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

	BeforeEach(func() {
		MockResponseYahooQuotes()
	})

	Describe("Run", func() {
		var (
			inputOptions = print.Options{}
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
			inputDependencies = c.Dependencies{
				HttpClients: c.DependenciesHttpClients{
					Default: client,
					Yahoo:   client,
				},
			}
		)

		It("should print holdings in JSON format", func() {
			output := getStdout(func() {
				print.Run(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal("[{\"name\":\"Alphabet Inc.\",\"symbol\":\"GOOG\",\"price\":\"2838.420000\",\"value\":\"28384.200000\",\"cost\":\"10000.000000\",\"quantity\":\"10.000000\",\"weight\":\"96.996890\"},{\"name\":\"Roblox Corporation\",\"symbol\":\"RBLX\",\"price\":\"87.880000\",\"value\":\"878.800000\",\"cost\":\"500.000000\",\"quantity\":\"10.000000\",\"weight\":\"3.003110\"}]\n"))
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
})
