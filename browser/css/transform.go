package css

import "strings"

// ApplyTextTransform transforms text based on CSS text-transform property
func ApplyTextTransform(text, transform string) string {
	switch transform {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		return CapitalizeWords(text)
	default:
		return text
	}
}

// CapitalizeWords capitalizes the first letter of each word
func CapitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
