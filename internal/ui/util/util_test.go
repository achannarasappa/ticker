package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/internal/ui/util"
)

var _ = Describe("Util", func() {
	Describe("ConvertFloatToString", func() {
		It("should convert a float to a string with a precision of two", func() {
			output := ConvertFloatToString(12.5634)
			Expect(output).To(Equal("12.56"))
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
			expectedOutput := NewStyle("#d4d4d4", "", false)("435.32")
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
	Describe("NewStyleFromGradient", func() {
		inputGradientFn := NewStyleFromGradient("#ffffff", "#000000")
		When("the percent given is 100%", func() {
			It("should generate text with the gradient of two colors relative to the percentage given", func() {
				inputStyleFn := inputGradientFn(100)
				output := inputStyleFn("test")
				expectedASCII := "test"
				expectedANSI16Color := "\x1b[30mtest\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;16mtest\x1b[0m"
				expectedTrueColor := "\x1b[38;2;0;0;0mtest\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})
		When("the percent given is 1%", func() {
			It("should generate text with the gradient of two colors relative to the percentage given", func() {
				inputStyleFn := inputGradientFn(1)
				output := inputStyleFn("test")
				expectedASCII := "test"
				expectedANSI16Color := "\x1b[37mtest\x1b[0m"
				expectedANSI256Color := "\x1b[38;5;188mtest\x1b[0m"
				expectedTrueColor := "\x1b[38;2;230;230;230mtest\x1b[0m"
				Expect(output).To(SatisfyAny(Equal(expectedASCII), Equal(expectedANSI16Color), Equal(expectedANSI256Color), Equal(expectedTrueColor)))
			})
		})
	})
})
