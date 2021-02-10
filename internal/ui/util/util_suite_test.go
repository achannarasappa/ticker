package util_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	termColor string
)

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util Suite")
}

var _ = BeforeSuite(func() {
	termColor = os.Getenv("TERM")
	os.Setenv("TERM", "xterm-256color")
})

var _ = AfterSuite(func() {
	os.Setenv("TERM", termColor)
})
