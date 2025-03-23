package utils

import (
	"strings"
)

// WordWrap wraps text to the specified line width
func WordWrap(text string, lineWidth int) string {
	var result string
	words := strings.Fields(text)

	if len(words) == 0 {
		return ""
	}

	line := words[0]

	for _, word := range words[1:] {
		if len(line)+len(word)+1 > lineWidth {
			result += line + "\n"
			line = word
		} else {
			line += " " + word
		}
	}
	result += line
	return result
}
