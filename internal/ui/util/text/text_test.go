package text_test

import (
	. "github.com/achannarasappa/ticker/internal/ui/util/text"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Text", func() {

	Describe("TextAlign", func() {
		When("left align is selected", func() {
			It("returns the text for left align text", func() {
				input := LeftAlign.String()
				output := "LeftAlign"
				Expect(input).To(Equal(output))
			})
		})
		When("right align is selected", func() {
			It("returns the text for right align text", func() {
				input := RightAlign.String()
				output := "RightAlign"
				Expect(input).To(Equal(output))
			})
		})
	})

	Describe("Line", func() {
		It("should output text with spaces filled to the width", func() {
			input := Line(10, Cell{
				Text: "test",
			})
			output := "test      "
			Expect(input).To(Equal(output))
		})
		Context("when there are multiple cells", func() {
			It("should output text equally spaced apart", func() {
				input := Line(20,
					Cell{
						Text: "test1",
					},
					Cell{
						Text: "test2",
					},
				)
				output := "test1     test2     "
				Expect(input).To(Equal(output))
			})
		})
		Context("when cell width is set", func() {
			It("should set the cell's width", func() {
				input := Line(20,
					Cell{
						Text:  "test1",
						Width: 7,
					},
					Cell{
						Text: "test2",
					},
				)
				output := "test1  test2        "
				Expect(input).To(Equal(output))
			})

			It("should evenly distribute the remaining space across the other cells", func() {
				input := Line(23,
					Cell{
						Text:  "test1",
						Width: 10,
					},
					Cell{
						Text: "test2",
					},
					Cell{
						Text: "test3",
					},
				)
				output := "test1     test2  test3 "
				Expect(input).To(Equal(output))
			})
		})
		Context("when align is set to LeftAlign", func() {
			It("should add spaces to the end of the text up to the width", func() {
				input := Line(10, Cell{
					Text:  "test",
					Align: LeftAlign,
				})
				output := "test      "
				Expect(input).To(Equal(output))
			})
		})
		Context("when align is set to RightAlign", func() {
			It("should add spaces to the beggining of the text up to the width", func() {
				input := Line(10, Cell{
					Text:  "test",
					Align: RightAlign,
				})
				output := "      test"
				Expect(input).To(Equal(output))
			})
		})
		Context("when the text length exceeds the line width", func() {
			It("should truncate the text", func() {
				input := Line(2, Cell{
					Text: "test",
				})
				output := "te"
				Expect(input).To(Equal(output))
			})
		})
	})

	Describe("JoinLines", func() {
		It("should join text with a newline between", func() {
			Expect(JoinLines("a", "b", "c")).To(Equal("a\nb\nc"))
		})
	})
})
