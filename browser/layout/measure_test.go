package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeasureText(t *testing.T) {
	// Save original TextMeasurer and restore after tests
	originalMeasurer := TextMeasurer
	defer func() { TextMeasurer = originalMeasurer }()

	t.Run("with default estimation (no custom measurer)", func(t *testing.T) {
		TextMeasurer = nil

		tests := []struct {
			name     string
			text     string
			fontSize float64
			expected float64
		}{
			{"empty string returns 0", "", 16, 0},
			{"single character", "a", 16, 8},    // 1 * 16 * 0.5 = 8
			{"single character larger font", "a", 32, 16}, // 1 * 32 * 0.5 = 16
			{"multiple characters", "hello", 16, 40}, // 5 * 16 * 0.5 = 40
			{"space counts as character", " ", 16, 8},
			{"text with spaces", "a b", 16, 24}, // 3 * 16 * 0.5 = 24
			{"longer text", "Hello World", 16, 88}, // 11 * 16 * 0.5 = 88
			{"zero font size", "hello", 0, 0},
			{"small font size", "ab", 10, 10}, // 2 * 10 * 0.5 = 10
			{"large font size", "ab", 100, 100}, // 2 * 100 * 0.5 = 100
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := MeasureText(tt.text, tt.fontSize)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("with custom TextMeasurer", func(t *testing.T) {
		// Track calls to the custom measurer
		var calledWith struct {
			text     string
			fontSize float64
			bold     bool
			italic   bool
		}

		TextMeasurer = func(text string, fontSize float64, bold, italic bool) float64 {
			calledWith.text = text
			calledWith.fontSize = fontSize
			calledWith.bold = bold
			calledWith.italic = italic
			return 999.0 // Return a distinctive value
		}

		result := MeasureText("test", 20)

		assert.Equal(t, 999.0, result, "should return custom measurer result")
		assert.Equal(t, "test", calledWith.text, "should pass text to measurer")
		assert.Equal(t, 20.0, calledWith.fontSize, "should pass fontSize to measurer")
		assert.False(t, calledWith.bold, "should pass bold=false")
		assert.False(t, calledWith.italic, "should pass italic=false")
	})

	t.Run("custom measurer receives empty string", func(t *testing.T) {
		var receivedText string
		TextMeasurer = func(text string, fontSize float64, bold, italic bool) float64 {
			receivedText = text
			return 0
		}

		MeasureText("", 16)
		assert.Equal(t, "", receivedText)
	})

	t.Run("custom measurer overrides default behavior", func(t *testing.T) {
		TextMeasurer = func(text string, fontSize float64, bold, italic bool) float64 {
			// Custom measurer that returns different value than default
			return float64(len(text)) * 10.0 // Different multiplier
		}

		result := MeasureText("hello", 16)
		// Default would be 5 * 16 * 0.5 = 40
		// Custom returns 5 * 10 = 50
		assert.Equal(t, 50.0, result)
	})
}

func TestMeasureTextFormula(t *testing.T) {
	// Ensure TextMeasurer is nil to test default formula
	originalMeasurer := TextMeasurer
	TextMeasurer = nil
	defer func() { TextMeasurer = originalMeasurer }()

	// Test the formula: len(text) * fontSize * 0.5
	tests := []struct {
		text     string
		fontSize float64
	}{
		{"a", 12},
		{"ab", 14},
		{"abc", 16},
		{"test", 18},
		{"hello world", 20},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			expected := float64(len(tt.text)) * tt.fontSize * 0.5
			result := MeasureText(tt.text, tt.fontSize)
			assert.Equal(t, expected, result)
		})
	}
}

func TestWrapText(t *testing.T) {
	// Use default estimation: len(text) * fontSize * 0.5
	originalMeasurer := TextMeasurer
	TextMeasurer = nil
	defer func() { TextMeasurer = originalMeasurer }()

	tests := []struct {
		name          string
		text          string
		fontSize      float64
		maxWidth      float64
		expectedLines int
		expectedFirst string
	}{
		{"empty text", "", 16, 100, 0, ""},
		{"fits in one line", "short", 16, 100, 1, "short"},
		{"single word too long still one line", "superlongword", 16, 50, 1, "superlongword"},
		{"wraps to two lines", "hello world", 16, 50, 2, "hello"},
		{"wraps to three lines", "one two three", 16, 40, 3, "one"},
		{"zero maxWidth no wrap", "hello world", 16, 0, 1, "hello world"},
		{"negative maxWidth no wrap", "hello world", 16, -10, 1, "hello world"},
		{"exact fit one line", "ab", 16, 16, 1, "ab"},
		{"just over wraps", "abc de", 16, 24, 2, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.fontSize, tt.maxWidth)
			assert.Equal(t, tt.expectedLines, len(result), "wrong number of lines")
			if tt.expectedLines > 0 && tt.expectedFirst != "" {
				assert.Equal(t, tt.expectedFirst, result[0], "first line mismatch")
			}
		})
	}
}
