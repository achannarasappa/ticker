package asset_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAsset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Asset Suite")
}
