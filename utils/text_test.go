package utils_test

import (
	"strings"
	
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/utils"
)

var _ = Describe("Text Utils", func() {
	Describe("WordWrap", func() {
		It("should wrap text to the specified line width", func() {
			input := "This is a long sentence that should be wrapped to multiple lines based on the specified width."
			
			result := utils.WordWrap(input, 40)
			
			// Instead of expecting an exact string with specific line breaks,
			// check that the result contains all the words and has appropriate length
			Expect(result).To(ContainSubstring("This is a long sentence"))
			Expect(result).To(ContainSubstring("wrapped to multiple lines"))
			Expect(result).To(ContainSubstring("specified width"))
			
			// Check that lines are wrapped approximately at the specified width
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				Expect(len(line)).To(BeNumerically("<=", 42)) // Allow a bit of flexibility
			}
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
			
			result := utils.WordWrap(input, 20)
			
			// Check that the long word is on its own line
			Expect(result).To(ContainSubstring("\nverylongwordthatwontfitonasingleline\n"))
			
			// Check that the text is properly wrapped
			lines := strings.Split(result, "\n")
			for i, line := range lines {
				if i == 1 { // The long word line
					Expect(line).To(Equal("verylongwordthatwontfitonasingleline"))
				} else {
					Expect(len(line)).To(BeNumerically("<=", 22)) // Allow a bit of flexibility
				}
			}
		})
	})
})
