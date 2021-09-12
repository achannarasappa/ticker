package print_test

import (
	"io/ioutil"
	"os"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/print"
	. "github.com/onsi/ginkgo"
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
		mockResponse()
	})

	Describe("Run", func() {
		var (
			inputOptions = print.Options{}
			inputContext = c.Context{Config: c.Config{Lots: []c.Lot{
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
			}}}
			inputDependencies = c.Dependencies{
				HttpClient: client,
			}
		)

		It("should print holdings in JSON format", func() {
			output := getStdout(func() {
				print.Run(&inputDependencies, &inputContext, &inputOptions)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal("[{\"name\":\"Alphabet Inc.\",\"symbol\":\"GOOG\",\"price\":2838.42,\"value\":28384.2,\"cost\":10000,\"quantity\":10,\"weight\":96.99689027099068},{\"name\":\"Roblox Corporation\",\"symbol\":\"RBLX\",\"price\":87.88,\"value\":878.8,\"cost\":500,\"quantity\":10,\"weight\":3.0031097290093287}]\n"))
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
