package sorter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSorter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sorter Suite")
}
