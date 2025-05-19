package monitorPriceYahoo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestYahoo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Yahoo Suite")
}
