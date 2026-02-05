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

// TestImportantOverride tests that !important declarations win in the cascade
func TestImportantOverride(t *testing.T) {
	tests := []struct {
		name          string
		css           string
		expectedColor color.Color
	}{
		{
			name: "important wins over later rule",
			css: `
				p { color: blue !important; }
				p { color: red; }
			`,
			expectedColor: color.RGBA{0, 0, 255, 255}, // blue
		},
		{
			name: "later important wins over earlier important",
			css: `
				p { color: blue !important; }
				p { color: red !important; }
			`,
			expectedColor: color.RGBA{255, 0, 0, 255}, // red
		},
		{
			name: "non-important cannot override important",
			css: `
				p { color: green !important; }
				p { color: yellow; }
				p { color: orange; }
			`,
			expectedColor: color.RGBA{0, 128, 0, 255}, // green
		},
		{
			name: "important on different property does not affect others",
			css: `
				p { color: blue !important; font-size: 20px; }
				p { color: red; font-size: 30px; }
			`,
			expectedColor: color.RGBA{0, 0, 255, 255}, // blue (font-size would be 30px)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sheet := Parse(tt.css)
			style := ApplyStylesheet(sheet, "p", "", nil)
			assert.True(t, colorsEqual(style.Color, tt.expectedColor),
				"expected %v, got %v", tt.expectedColor, style.Color)
		})
	}
}

// TestImportantWithContext tests !important with ApplyStylesheetWithContext
func TestImportantWithContext(t *testing.T) {
	tests := []struct {
		name             string
		css              string
		expectedFontSize float64
	}{
		{
			name: "important font-size wins",
			css: `
				p { font-size: 20px !important; }
				p { font-size: 30px; }
			`,
			expectedFontSize: 20,
		},
		{
			name: "later important font-size wins",
			css: `
				p { font-size: 20px !important; }
				p { font-size: 30px !important; }
			`,
			expectedFontSize: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sheet := Parse(tt.css)
			style := ApplyStylesheetWithContext(sheet, "p", "", nil, 16, DefaultViewportWidth, DefaultViewportHeight)
			assert.Equal(t, tt.expectedFontSize, style.FontSize)
		})
	}
}

// TestInlineStyleImportant tests that !important is handled in inline styles
func TestInlineStyleImportant(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedColor color.Color
	}{
		{
			name:          "inline style with !important",
			input:         "color: red !important",
			expectedColor: color.RGBA{255, 0, 0, 255},
		},
		{
			name:          "inline style with !important no space",
			input:         "color: blue!important",
			expectedColor: color.RGBA{0, 0, 255, 255},
		},
		{
			name:          "inline style with !IMPORTANT uppercase",
			input:         "color: green !IMPORTANT",
			expectedColor: color.RGBA{0, 128, 0, 255},
		},
		{
			name:          "important beats later non-important in same inline",
			input:         "color: green !important; color: blue",
			expectedColor: color.RGBA{0, 128, 0, 255}, // green wins
		},
		{
			name:          "later important beats earlier important in same inline",
			input:         "color: red !important; color: blue !important",
			expectedColor: color.RGBA{0, 0, 255, 255}, // blue wins
		},
		{
			name:          "multiple non-important after important",
			input:         "color: green !important; color: red; color: blue",
			expectedColor: color.RGBA{0, 128, 0, 255}, // green stays
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := ParseInlineStyle(tt.input)
			assert.True(t, colorsEqual(style.Color, tt.expectedColor),
				"expected %v, got %v", tt.expectedColor, style.Color)
		})
	}
}

func TestParseLineHeight(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fontSize float64
		expected float64
	}{
		{"unitless 1", "1", 16.0, 16.0},
		{"unitless 1.5", "1.5", 16.0, 24.0},
		{"unitless 2", "2", 16.0, 32.0},
		{"unitless with larger font", "1.5", 20.0, 30.0},
		{"pixel value", "24px", 16.0, 24.0},
		{"pixel value ignores font-size", "40px", 20.0, 40.0},
		{"normal keyword", "normal", 16.0, 19.2},
		{"normal keyword larger font", "normal", 20.0, 24.0},
		{"zero", "0", 16.0, 0.0},
		{"invalid", "invalid", 16.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLineHeight(tt.value, tt.fontSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitBackgroundValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single color", "red", []string{"red"}},
		{"single url", "url(img.png)", []string{"url(img.png)"}},
		{"color and url", "red url(img.png)", []string{"red", "url(img.png)"}},
		{"url with spaces", "url(my image.png)", []string{"url(my image.png)"}},
		{"color and url with spaces", "blue url(path/to image.png)", []string{"blue", "url(path/to image.png)"}},
		{"hex color", "#ff0000 url(x.png)", []string{"#ff0000", "url(x.png)"}},
		{"url then color", "url(test.png) green", []string{"url(test.png)", "green"}},
		{"multiple spaces between", "red   url(x.png)", []string{"red", "url(x.png)"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitBackgroundValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseBackgroundShorthand(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedColor color.Color
		expectedImage string
	}{
		{"color only", "red", color.RGBA{255, 0, 0, 255}, ""},
		{"hex color only", "#00ff00", color.RGBA{0, 255, 0, 255}, ""},
		{"url only", "url(cat.png)", nil, "cat.png"},
		{"url with single quotes", "url('cat.png')", nil, "cat.png"},
		{"url with double quotes", `url("cat.png")`, nil, "cat.png"},
		{"color and url", "blue url(dog.png)", color.RGBA{0, 0, 255, 255}, "dog.png"},
		{"url and color reversed", "url(dog.png) blue", color.RGBA{0, 0, 255, 255}, "dog.png"},
		{"none keyword", "none", nil, ""},
		{"named color lightblue", "lightblue", color.RGBA{173, 216, 230, 255}, ""},
		{"url with path", "url(images/bg.png)", nil, "images/bg.png"},
		{"color and url with path", "yellow url(path/to/img.jpg)", color.RGBA{255, 255, 0, 255}, "path/to/img.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotColor, gotImage := parseBackgroundShorthand(tt.input)
			assert.True(t, colorsEqual(gotColor, tt.expectedColor),
				"color mismatch for %q: expected %v, got %v", tt.input, tt.expectedColor, gotColor)
			assert.Equal(t, tt.expectedImage, gotImage,
				"image mismatch for %q", tt.input)
		})
	}
}

func TestBackgroundShorthandInlineStyle(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedColor color.Color
		expectedImage string
	}{
		{
			name:          "background color only",
			input:         "background: red",
			expectedColor: color.RGBA{255, 0, 0, 255},
			expectedImage: "",
		},
		{
			name:          "background hex color",
			input:         "background: #0000ff",
			expectedColor: color.RGBA{0, 0, 255, 255},
			expectedImage: "",
		},
		{
			name:          "background url only",
			input:         "background: url(test.png)",
			expectedColor: nil,
			expectedImage: "test.png",
		},
		{
			name:          "background color and url",
			input:         "background: green url(bg.jpg)",
			expectedColor: color.RGBA{0, 128, 0, 255},
			expectedImage: "bg.jpg",
		},
		{
			name:          "background url and color reversed",
			input:         "background: url(bg.jpg) purple",
			expectedColor: color.RGBA{128, 0, 128, 255},
			expectedImage: "bg.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := ParseInlineStyle(tt.input)
			assert.True(t, colorsEqual(style.BackgroundColor, tt.expectedColor),
				"BackgroundColor mismatch: expected %v, got %v", tt.expectedColor, style.BackgroundColor)
			assert.Equal(t, tt.expectedImage, style.BackgroundImage,
				"BackgroundImage mismatch")
		})
	}
}
