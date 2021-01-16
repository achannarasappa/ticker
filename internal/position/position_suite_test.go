package position_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPosition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Position Suite")
}
