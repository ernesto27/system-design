package layout

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
