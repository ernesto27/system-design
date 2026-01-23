package layout

import "strings"

// MeasureTextFunc is a function that measures text width given text, fontSize, bold, italic
type MeasureTextFunc func(text string, fontSize float64, bold bool, italic bool) float64

// TextMeasurer is the function used to measure text width.
// Set this to use accurate font measurements (e.g., from Fyne).
// If nil, falls back to estimation.
var TextMeasurer MeasureTextFunc

// MeasureText returns the width of text.
// Uses TextMeasurer if set, otherwise estimates.
func MeasureText(text string, fontSize float64) float64 {
	if TextMeasurer != nil {
		return TextMeasurer(text, fontSize, false, false)
	}
	// Fallback: rough estimation
	if len(text) == 0 {
		return 0
	}
	avgCharWidth := fontSize * 0.5
	return float64(len(text)) * avgCharWidth
}

// WrapText breaks text into lines that fit within maxWidth.
// Returns slice of lines. Words are not broken mid-word.
func WrapText(text string, fontSize float64, maxWidth float64) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Try adding word to current line
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth := MeasureText(testLine, fontSize)

		if lineWidth <= maxWidth || currentLine.Len() == 0 {
			// Word fits, or it's the first word (must include even if too long)
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		} else {
			// Word doesn't fit, start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	// Don't forget the last line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
