package css

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

// colorsEqual compares two color.Color values by their RGBA components
func colorsEqual(c1, c2 color.Color) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected color.Color
	}{
		// Named colors
		{"named black", "black", color.Black},
		{"named white", "white", color.White},
		{"named red", "red", color.RGBA{255, 0, 0, 255}},
		{"named blue", "blue", color.RGBA{0, 0, 255, 255}},
		{"named green", "green", color.RGBA{0, 128, 0, 255}},

		// Hex colors - 6 digit
		{"hex 6-digit red", "#ff0000", color.RGBA{255, 0, 0, 255}},
		{"hex 6-digit green", "#00ff00", color.RGBA{0, 255, 0, 255}},
		{"hex 6-digit blue", "#0000ff", color.RGBA{0, 0, 255, 255}},
		{"hex 6-digit mixed", "#1a2b3c", color.RGBA{26, 43, 60, 255}},

		// Hex colors - 3 digit
		{"hex 3-digit red", "#f00", color.RGBA{255, 0, 0, 255}},
		{"hex 3-digit white", "#fff", color.RGBA{255, 255, 255, 255}},
		{"hex 3-digit gray", "#888", color.RGBA{136, 136, 136, 255}},

		// Case insensitivity
		{"hex uppercase", "#FF0000", color.RGBA{255, 0, 0, 255}},
		{"hex mixed case", "#FfAa00", color.RGBA{255, 170, 0, 255}},
		{"named uppercase", "RED", color.RGBA{255, 0, 0, 255}},
		{"named mixed case", "Blue", color.RGBA{0, 0, 255, 255}},

		// Invalid colors
		{"invalid color name", "notacolor", nil},
		{"empty string", "", nil},
		{"hex missing hash", "ff0000", nil},
		{"hex wrong length 4", "#ff00", nil},
		{"hex wrong length 5", "#ff000", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColor(tt.input)
			assert.True(t, colorsEqual(result, tt.expected), "ParseColor(%q) color mismatch", tt.input)
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		// Valid pixel values
		{"pixel value", "10px", 10.0},
		{"pixel decimal", "10.5px", 10.5},
		{"pixel zero", "0px", 0.0},

		// Plain numbers
		{"plain integer", "10", 10.0},
		{"plain decimal", "10.5", 10.5},
		{"plain zero", "0", 0.0},

		// With whitespace
		{"leading space", "  10px", 10.0},
		{"trailing space", "10px  ", 10.0},
		{"both spaces", "  10px  ", 10.0},

		// Case insensitivity
		{"uppercase PX", "10PX", 10.0},
		{"mixed case Px", "10Px", 10.0},

		// Invalid values
		{"invalid string", "abc", 0.0},
		{"empty string", "", 0.0},
		{"only px", "px", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSize(tt.input)
			assert.Equal(t, tt.expected, result, "ParseSize(%q)", tt.input)
		})
	}
}

func TestMatchSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector Selector
		tagName  string
		id       string
		classes  []string
		expected bool
	}{
		// Tag matching
		{
			name:     "tag match",
			selector: Selector{TagName: "div"},
			tagName:  "div", id: "", classes: nil,
			expected: true,
		},
		{
			name:     "tag no match",
			selector: Selector{TagName: "div"},
			tagName:  "span", id: "", classes: nil,
			expected: false,
		},

		// ID matching
		{
			name:     "id match",
			selector: Selector{ID: "main"},
			tagName:  "div", id: "main", classes: nil,
			expected: true,
		},
		{
			name:     "id no match",
			selector: Selector{ID: "main"},
			tagName:  "div", id: "other", classes: nil,
			expected: false,
		},

		// Class matching
		{
			name:     "class match",
			selector: Selector{Classes: []string{"foo"}},
			tagName:  "div", id: "", classes: []string{"foo"},
			expected: true,
		},
		{
			name:     "class no match",
			selector: Selector{Classes: []string{"foo"}},
			tagName:  "div", id: "", classes: []string{"bar"},
			expected: false,
		},
		{
			name:     "multiple classes all present",
			selector: Selector{Classes: []string{"a", "b"}},
			tagName:  "div", id: "", classes: []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "multiple classes partial match",
			selector: Selector{Classes: []string{"a", "b"}},
			tagName:  "div", id: "", classes: []string{"a"},
			expected: false,
		},

		// Combined selectors
		{
			name:     "tag and class match",
			selector: Selector{TagName: "div", Classes: []string{"foo"}},
			tagName:  "div", id: "", classes: []string{"foo"},
			expected: true,
		},
		{
			name:     "tag and class - tag mismatch",
			selector: Selector{TagName: "div", Classes: []string{"foo"}},
			tagName:  "span", id: "", classes: []string{"foo"},
			expected: false,
		},
		{
			name:     "tag id and class all match",
			selector: Selector{TagName: "div", ID: "main", Classes: []string{"foo"}},
			tagName:  "div", id: "main", classes: []string{"foo", "bar"},
			expected: true,
		},

		// Empty selector matches everything
		{
			name:     "empty selector matches all",
			selector: Selector{},
			tagName:  "div", id: "any", classes: []string{"any"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchSelector(tt.selector, tt.tagName, tt.id, tt.classes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseInlineStyle(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, s Style)
	}{
		{
			name:  "single color property",
			input: "color: red",
			verify: func(t *testing.T, s Style) {
				expected := color.RGBA{255, 0, 0, 255}
				assert.True(t, colorsEqual(s.Color, expected), "Color mismatch")
			},
		},
		{
			name:  "single font-size property",
			input: "font-size: 20px",
			verify: func(t *testing.T, s Style) {
				assert.Equal(t, 20.0, s.FontSize)
			},
		},
		{
			name:  "multiple properties",
			input: "color: blue; font-size: 24px; font-weight: bold",
			verify: func(t *testing.T, s Style) {
				expectedColor := color.RGBA{0, 0, 255, 255}
				assert.True(t, colorsEqual(s.Color, expectedColor), "Color mismatch")
				assert.Equal(t, 24.0, s.FontSize)
				assert.True(t, s.Bold)
			},
		},
		{
			name:  "margin property",
			input: "margin: 10px",
			verify: func(t *testing.T, s Style) {
				assert.Equal(t, 10.0, s.MarginTop)
				assert.Equal(t, 10.0, s.MarginBottom)
				assert.Equal(t, 10.0, s.MarginLeft)
				assert.Equal(t, 10.0, s.MarginRight)
			},
		},
		{
			name:  "padding property",
			input: "padding: 5px",
			verify: func(t *testing.T, s Style) {
				assert.Equal(t, 5.0, s.PaddingTop)
				assert.Equal(t, 5.0, s.PaddingBottom)
				assert.Equal(t, 5.0, s.PaddingLeft)
				assert.Equal(t, 5.0, s.PaddingRight)
			},
		},
		{
			name:  "with trailing semicolon",
			input: "color: red;",
			verify: func(t *testing.T, s Style) {
				expected := color.RGBA{255, 0, 0, 255}
				assert.True(t, colorsEqual(s.Color, expected), "Color mismatch")
			},
		},
		{
			name:  "extra whitespace",
			input: "  color :  red  ;  font-size : 16px  ",
			verify: func(t *testing.T, s Style) {
				expected := color.RGBA{255, 0, 0, 255}
				assert.True(t, colorsEqual(s.Color, expected), "Color mismatch")
				assert.Equal(t, 16.0, s.FontSize)
			},
		},
		{
			name:  "empty string returns default",
			input: "",
			verify: func(t *testing.T, s Style) {
				def := DefaultStyle()
				assert.Equal(t, def.FontSize, s.FontSize)
				assert.True(t, colorsEqual(s.Color, def.Color), "Color should be default")
			},
		},
		{
			name:  "malformed property ignored",
			input: "color",
			verify: func(t *testing.T, s Style) {
				def := DefaultStyle()
				assert.True(t, colorsEqual(s.Color, def.Color), "Malformed property should be ignored")
			},
		},
		{
			name:  "opacity property",
			input: "opacity: 0.5",
			verify: func(t *testing.T, s Style) {
				assert.Equal(t, 0.5, s.Opacity)
			},
		},
		{
			name:  "display property",
			input: "display: none",
			verify: func(t *testing.T, s Style) {
				assert.Equal(t, "none", s.Display)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInlineStyle(tt.input)
			tt.verify(t, result)
		})
	}
}
