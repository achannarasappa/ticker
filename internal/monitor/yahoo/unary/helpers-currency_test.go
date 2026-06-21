package unary_test

import (
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinorUnitForCurrencyCode", func() {

	When("the major currency has a minor form", func() {
		It("should return the minor currency code and unit", func() {
			ok, minorCode, minorUnit := unary.MinorUnitForCurrencyCode("GBP")

			Expect(ok).To(BeTrue())
			Expect(minorCode).To(Equal("GBp"))
			Expect(minorUnit).To(Equal(2.0))
		})
	})

	When("the major currency does not have a minor form", func() {
		It("should return false and zero values", func() {
			ok, minorCode, minorUnit := unary.MinorUnitForCurrencyCode("JPY")

			Expect(ok).To(BeFalse())
			Expect(minorCode).To(Equal(""))
			Expect(minorUnit).To(Equal(0.0))
		})
	})

})
