package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "ticker/internal/ui/util"
)

var _ = Describe("Util", func() {
	Describe("ConvertFloatToString", func() {
		It("should convert a float to a string with a precision of two", func() {
			output := ConvertFloatToString(12.5634)
			Expect(output).To(Equal("12.56"))
		})
	})
	Describe("NewStyle", func() {
		It("should generate text with a background and foreground color", func() {
			inputStyleFn := NewStyle("#ffffff", "#000000", false)
			output := inputStyleFn("test")
			Expect(output).To(Equal("\x1b[38;5;231;48;5;16mtest\x1b[0m"))
		})
		It("should generate text with bold styling", func() {
			inputStyleFn := NewStyle("#ffffff", "#000000", true)
			output := inputStyleFn("test")
			Expect(output).To(Equal("\x1b[38;5;231;48;5;16;1mtest\x1b[0m"))
		})
	})
})
