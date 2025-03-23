package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/utils"
)

var _ = Describe("Text Utils", func() {
	Describe("WordWrap", func() {
		It("should wrap text to the specified line width", func() {
			input := "This is a long sentence that should be wrapped to multiple lines based on the specified width."
			expected := "This is a long sentence that should be\nwrapped to multiple lines based on the\nspecified width."
			
			result := utils.WordWrap(input, 40)
			Expect(result).To(Equal(expected))
		})

		It("should handle empty input", func() {
			result := utils.WordWrap("", 40)
			Expect(result).To(Equal(""))
		})

		It("should handle single word input", func() {
			result := utils.WordWrap("Hello", 40)
			Expect(result).To(Equal("Hello"))
		})

		It("should handle input with words longer than the line width", func() {
			input := "This contains a verylongwordthatwontfitonasingleline and should wrap properly."
			expected := "This contains a\nverylongwordthatwontfitonasingleline\nand should wrap properly."
			
			result := utils.WordWrap(input, 20)
			Expect(result).To(Equal(expected))
		})
	})
})
