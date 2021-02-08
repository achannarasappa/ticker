package util_test

import (
	"github.com/lucasb-eyer/go-colorful"

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
			Expect(output).To(ContainSubstring("test\x1b[0m"))
		})
		It("should generate text with bold styling", func() {
			inputStyleFn := NewStyle("#ffffff", "#000000", true)
			output := inputStyleFn("test")
			Expect(output).To(ContainSubstring("test\x1b[0m"))
		})
	})
	Describe("NewStyleFromGradient", func() {
		c1, _ := colorful.Hex("#ffffff")
		c2, _ := colorful.Hex("#000000")
		inputGradientFn := NewStyleFromGradient("#ffffff", "#000000")
		When("the percent given is 100%", func() {
			It("should generate text with the gradient of two colors relative to the percentage given", func() {
				inputStyleFn := inputGradientFn(100)
				output := inputStyleFn("test")
				expectedOutput := NewStyle(c1.BlendHsv(c2, 1.0).Hex(), "", false)("test")
				Expect(output).To(Equal(expectedOutput))
			})
		})
		When("the percent given is 1%", func() {
			It("should generate text with the gradient of two colors relative to the percentage given", func() {
				inputStyleFn := inputGradientFn(1)
				output := inputStyleFn("test")
				expectedOutput := NewStyle(c1.BlendHsv(c2, 0.01).Hex(), "", false)("test")
				Expect(output).To(Equal(expectedOutput))
			})
		})
	})
})
