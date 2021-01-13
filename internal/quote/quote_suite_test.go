package quote_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// var client = resty.New()

var _ = BeforeSuite(func() {
	// httpmock.ActivateNonDefault(client.GetClient())
})

// var _ = BeforeEach(func() {
// 	httpmock.Reset()
// })

var _ = AfterSuite(func() {
	// httpmock.DeactivateAndReset()
})

func TestQuote(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Quote Suite")
}
