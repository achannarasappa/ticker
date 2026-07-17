package monitorPriceCoinPaprika_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMonitorPriceCoinPaprika(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Monitor Price CoinPaprika Suite")
}
