package monitorCoinbase_test

// import (
// 	"time"

// 	"github.com/go-resty/resty/v2"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"

// 	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
// )

// var _ = Describe("Monitor Coinbase", func() {

// 	Describe("NewMonitorCoinbase", func() {
// 		PIt("should return a new MonitorCoinbase", func() {
// 			monitor := monitorCoinbase.NewMonitorCoinbase(time.Second, *resty.New(), []string{"BTC-USD"}, func() {})

// 			Expect(monitor).NotTo(BeNil())
// 		})
// 	})

// 	Describe("GetAssetQuotes", func() {
// 		PIt("should return the asset quotes", func() {
// 			monitor := monitorCoinbase.NewMonitorCoinbase(time.Second, *resty.New(), []string{"BTC-USD"}, func() {})

// 			quotes := monitor.GetAssetQuotes()
// 			Expect(quotes).NotTo(BeEmpty())
// 		})
// 	})
// })
