package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/internal/ui/util"
)

var _ = Describe("Util", func() {
	Describe("ConvertFloatToString", func() {
		It("should convert a float to a precision of two", func() {
			output := ConvertFloatToString(0.563412, false)
			Expect(output).To(Equal("0.56"))
		})

		When("there using variable precision", func() {
			It("should convert a float that smaller than 10 to a string with a precision of four", func() {
				output := ConvertFloatToString(0.563412, true)
				Expect(output).To(Equal("0.5634"))
			})
			It("should convert a float that between 10 and 100 to a string with a precision of three", func() {
				output := ConvertFloatToString(12.5634, true)
				Expect(output).To(Equal("12.563"))
			})
			It("should convert a float that greater than 100 to a string with a precision of two", func() {
				output := ConvertFloatToString(204.4325, true)
				Expect(output).To(Equal("204.43"))
			})
			It("should set a precision of two when the value is zero", func() {
				output := ConvertFloatToString(0.0, true)
				Expect(output).To(Equal("0.00"))
			})
			It("should set a precision of zero when the value is over 10000", func() {
				output := ConvertFloatToString(10000.0, true)
				Expect(output).To(Equal("10000"))
			})
		})

	})
	Describe("ValueText", func() {
		When("value is <= 0.0", func() {
			It("should return an empty string", func() {
				output := ValueText(0.0)
				Expect(output).To(ContainSubstring(""))
			})
		})
		It("should generate text for values", func() {
			output := ValueText(435.32)
			expectedOutput := NewStyle("#d0d0d0", "", false)("435.32")
			Expect(output).To(Equal(expectedOutput))
		})
	})
	Describe("NewStyle", func() {
		It("should generate text with a background and foreground color", func() {
			inputStyleFn := NewStyle("#ffffff", "#000000", false)
			output := inputStyleFn("test")
			expectedASCII := "\x1b[;mtest\x1b[0m"
			expectedANSI16Color := "\x1b[97;40mtest\x1b[0m"
			expectedANSI256Color := "\x1b[38;5;231;48;5;16mtest\x1b[0m"
			expectedTrueColor := "\x1b[38;2;255;255;255;48;2;0;0;0mtest\x1b[0m"
			Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
		})
		It("should generate text with bold styling", func() {
			inputStyleFn := NewStyle("#ffffff", "#000000", true)
			output := inputStyleFn("test")
			expectedASCII := "\x1b[;;1mtest\x1b[0m"
			expectedANSI16Color := "\x1b[97;40;1mtest\x1b[0m"
			expectedANSI256Color := "\x1b[38;5;231;48;5;16;1mtest\x1b[0m"
			expectedTrueColor := "\x1b[38;2;255;255;255;48;2;0;0;0;1mtest\x1b[0m"
			Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
		})
	})

	Describe("StylePrice", func() {
		When("there is no percent change", func() {
			It("should color text grey", func() {
				output := StylePrice(0.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[90m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;241m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;5;241m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change over 10%", func() {
			It("should color text dark green", func() {
				output := StylePrice(11.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[32m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;70m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;119;153;40m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change between 5% and 10%", func() {
			It("should color text medium green", func() {
				output := StylePrice(7.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[92m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;76m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;143;184;48m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change between 0% and 5%", func() {
			It("should color text light green", func() {
				output := StylePrice(3.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[92m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;82m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;174;224;56m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change over -10%", func() {
			It("should color text dark red", func() {
				output := StylePrice(-11.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[31m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;124m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;153;73;38m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change between -5% and -10%", func() {
			It("should color text medium red", func() {
				output := StylePrice(-7.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[91m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;160m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;184;87;46m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

		When("there is a percent change between 0% and -5%", func() {
			It("should color text light red", func() {
				output := StylePrice(-3.0, "$100.00")
				expectedASCII := "$100.00"
				expectedANSI16Color := "\x1b[91m$100.00\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;196m$100.00\x1b[0m"
				expectedTrueColor := "\x1b[38;2;224;107;56m$100.00\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})

	})
})
