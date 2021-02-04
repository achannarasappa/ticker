package summary_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestSummary(t *testing.T) {
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	RunSpecs(t, "Summary Suite")
}
