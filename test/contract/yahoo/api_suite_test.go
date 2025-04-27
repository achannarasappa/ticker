package api_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestYahooAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Yahoo API Contract Test Suite")
}
