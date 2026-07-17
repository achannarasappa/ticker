package unary_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUnary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unary Suite")
}
