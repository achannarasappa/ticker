package monitorCoinbase_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCoinbase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coinbase Suite")
}
