package css

import (
	"image/color"
	"strconv"
	"strings"
)

type Style struct {
	Color           color.Color
	BackgroundColor color.Color
	FontSize        float64
	Bold            bool
	Italic          bool
	MarginTop       float64
	MarginBottom    float64
	MarginLeft      float64
	MarginRight     float64
	PaddingTop      float64
	PaddingBottom   float64
	PaddingLeft     float64
	PaddingRight    float64
	TextAlign       string
	Display         string
	TextDecoration  string
	Opacity         float64
	Visibility      string
	Cursor          string

	// Border properties
	BorderTopWidth    float64
	BorderRightWidth  float64
	BorderBottomWidth float64
	BorderLeftWidth   float64
	BorderTopColor    color.Color
	BorderRightColor  color.Color
	BorderBottomColor color.Color
	BorderLeftColor   color.Color
	BorderTopStyle    string
	BorderRightStyle  string
	BorderBottomStyle string
	BorderLeftStyle   string
}

func DefaultStyle() Style {
	return Style{
		Color:           color.Black,
		BackgroundColor: color.White,
		FontSize:        16,
		Bold:            false,
		Italic:          false,
		Opacity:         1.0,
	}
}

func ParseInlineStyle(styleAttr string) Style {
	style := DefaultStyle()

	parts := strings.Split(styleAttr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		property := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		applyDeclaration(&style, property, value)
	}
	return style
}

// parseBorderShorthand parses "1px solid black" into width, style, color
func parseBorderShorthand(value string) (float64, string, color.Color) {
	parts := strings.Fields(value)
	var width float64
	var borderStyle string
	var borderColor color.Color
	for _, part := range parts {
		if w := ParseSize(part); w > 0 {
			width = w
		} else if part == "solid" || part == "dashed" || part == "dotted" || part == "none" {
			borderStyle = part
		} else if c := ParseColor(part); c != nil {
			borderColor = c
		}
	}
	return width, borderStyle, borderColor
}

// ParseSize converts "10px" to float64
func ParseSize(value string) float64 {
	value = strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(value, "px") {
		num := strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(num, 64); err == nil {
			return size
		}
	}
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}
	return 0
}

// ParseColor converts color names or hex to color.Color
func ParseColor(value string) color.Color {
	value = strings.ToLower(value)

	// Named colors (CSS standard)
	colors := map[string]color.Color{
		// Basic colors
		"black":   color.Black,
		"white":   color.White,
		"red":     color.RGBA{255, 0, 0, 255},
		"green":   color.RGBA{0, 128, 0, 255},
		"blue":    color.RGBA{0, 0, 255, 255},
		"yellow":  color.RGBA{255, 255, 0, 255},
		"purple":  color.RGBA{128, 0, 128, 255},
		"orange":  color.RGBA{255, 165, 0, 255},
		"gray":    color.RGBA{128, 128, 128, 255},
		"grey":    color.RGBA{128, 128, 128, 255},
		"cyan":    color.RGBA{0, 255, 255, 255},
		"magenta": color.RGBA{255, 0, 255, 255},
		"pink":    color.RGBA{255, 192, 203, 255},
		"brown":   color.RGBA{165, 42, 42, 255},

		// Light variants
		"lightgray":   color.RGBA{211, 211, 211, 255},
		"lightgrey":   color.RGBA{211, 211, 211, 255},
		"lightblue":   color.RGBA{173, 216, 230, 255},
		"lightgreen":  color.RGBA{144, 238, 144, 255},
		"lightyellow": color.RGBA{255, 255, 224, 255},
		"lightpink":   color.RGBA{255, 182, 193, 255},
		"lightcyan":   color.RGBA{224, 255, 255, 255},

		// Dark variants
		"darkgray":    color.RGBA{169, 169, 169, 255},
		"darkgrey":    color.RGBA{169, 169, 169, 255},
		"darkblue":    color.RGBA{0, 0, 139, 255},
		"darkgreen":   color.RGBA{0, 100, 0, 255},
		"darkred":     color.RGBA{139, 0, 0, 255},
		"darkcyan":    color.RGBA{0, 139, 139, 255},
		"darkmagenta": color.RGBA{139, 0, 139, 255},
		"darkorange":  color.RGBA{255, 140, 0, 255},

		// Other common colors
		"navy":       color.RGBA{0, 0, 128, 255},
		"teal":       color.RGBA{0, 128, 128, 255},
		"maroon":     color.RGBA{128, 0, 0, 255},
		"olive":      color.RGBA{128, 128, 0, 255},
		"silver":     color.RGBA{192, 192, 192, 255},
		"aqua":       color.RGBA{0, 255, 255, 255},
		"lime":       color.RGBA{0, 255, 0, 255},
		"fuchsia":    color.RGBA{255, 0, 255, 255},
		"gold":       color.RGBA{255, 215, 0, 255},
		"coral":      color.RGBA{255, 127, 80, 255},
		"salmon":     color.RGBA{250, 128, 114, 255},
		"tomato":     color.RGBA{255, 99, 71, 255},
		"crimson":    color.RGBA{220, 20, 60, 255},
		"indigo":     color.RGBA{75, 0, 130, 255},
		"violet":     color.RGBA{238, 130, 238, 255},
		"plum":       color.RGBA{221, 160, 221, 255},
		"khaki":      color.RGBA{240, 230, 140, 255},
		"beige":      color.RGBA{245, 245, 220, 255},
		"ivory":      color.RGBA{255, 255, 240, 255},
		"wheat":      color.RGBA{245, 222, 179, 255},
		"tan":        color.RGBA{210, 180, 140, 255},
		"chocolate":  color.RGBA{210, 105, 30, 255},
		"firebrick":  color.RGBA{178, 34, 34, 255},
		"skyblue":    color.RGBA{135, 206, 235, 255},
		"steelblue":  color.RGBA{70, 130, 180, 255},
		"slategray":  color.RGBA{112, 128, 144, 255},
		"slategrey":  color.RGBA{112, 128, 144, 255},
		"dimgray":    color.RGBA{105, 105, 105, 255},
		"dimgrey":    color.RGBA{105, 105, 105, 255},
		"whitesmoke": color.RGBA{245, 245, 245, 255},
		"snow":       color.RGBA{255, 250, 250, 255},
		"honeydew":   color.RGBA{240, 255, 240, 255},
		"mintcream":  color.RGBA{245, 255, 250, 255},
		"azure":      color.RGBA{240, 255, 255, 255},
		"aliceblue":  color.RGBA{240, 248, 255, 255},
		"lavender":   color.RGBA{230, 230, 250, 255},
		"linen":      color.RGBA{250, 240, 230, 255},
		"seashell":   color.RGBA{255, 245, 238, 255},
	}

	if c, ok := colors[value]; ok {
		return c
	}

	// Hex color: #RGB or #RRGGBB
	if strings.HasPrefix(value, "#") {
		hex := value[1:]
		if len(hex) == 3 {
			// #RGB -> #RRGGBB
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) == 6 {
			r, _ := strconv.ParseUint(hex[0:2], 16, 8)
			g, _ := strconv.ParseUint(hex[2:4], 16, 8)
			b, _ := strconv.ParseUint(hex[4:6], 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
		}
	}

	return nil
}

type Selector struct {
	TagName string
	ID      string
	Classes []string
}

type Declaration struct {
	Property string
	Value    string
}

type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
}

type Stylesheet struct {
	Rules []Rule
}

// MatchSelector checks if a selector matches a DOM node
func MatchSelector(sel Selector, tagName string, id string, classes []string) bool {
	// Check tag name
	if sel.TagName != "" && sel.TagName != tagName {
		return false
	}

	// Check ID
	if sel.ID != "" && sel.ID != id {
		return false
	}

	// Check classes (all selector classes must be present)
	for _, selClass := range sel.Classes {
		found := false
		for _, nodeClass := range classes {
			if selClass == nodeClass {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// ApplyStylesheet applies matching rules from stylesheet to a base style
func ApplyStylesheet(sheet Stylesheet, tagName string, id string, classes []string) Style {
	style := DefaultStyle()

	// Check each rule
	for _, rule := range sheet.Rules {
		// Check if any selector matches
		matches := false
		for _, sel := range rule.Selectors {
			if MatchSelector(sel, tagName, id, classes) {
				matches = true
				break
			}
		}

		if !matches {
			continue
		}

		// Apply declarations
		for _, decl := range rule.Declarations {
			applyDeclaration(&style, decl.Property, decl.Value)
		}
	}

	return style
}

// applyDeclaration applies a single CSS property to a style
func applyDeclaration(style *Style, property, value string) {
	switch property {
	case "color":
		if c := ParseColor(value); c != nil {
			style.Color = c
		}
	case "background-color":
		if c := ParseColor(value); c != nil {
			style.BackgroundColor = c
		}
	case "font-size":
		if size := ParseSize(value); size > 0 {
			style.FontSize = size
		}
	case "font-weight":
		style.Bold = (value == "bold")
	case "font-style":
		style.Italic = (value == "italic")
	case "margin":
		m := ParseSize(value)
		style.MarginTop = m
		style.MarginBottom = m
		style.MarginLeft = m
		style.MarginRight = m
	case "margin-top":
		style.MarginTop = ParseSize(value)
	case "margin-bottom":
		style.MarginBottom = ParseSize(value)
	case "margin-left":
		style.MarginLeft = ParseSize(value)
	case "margin-right":
		style.MarginRight = ParseSize(value)
	case "padding":
		p := ParseSize(value)
		style.PaddingTop = p
		style.PaddingBottom = p
		style.PaddingLeft = p
		style.PaddingRight = p
	case "padding-top":
		style.PaddingTop = ParseSize(value)
	case "padding-bottom":
		style.PaddingBottom = ParseSize(value)
	case "padding-left":
		style.PaddingLeft = ParseSize(value)
	case "padding-right":
		style.PaddingRight = ParseSize(value)
	case "text-align":
		style.TextAlign = value
	case "display":
		style.Display = value
	case "text-decoration":
		style.TextDecoration = value
	case "opacity":
		if op, err := strconv.ParseFloat(value, 64); err == nil {
			if op < 0 {
				op = 0
			} else if op > 1 {
				op = 1
			}
			style.Opacity = op
		}
	case "visibility":
		style.Visibility = value
	case "cursor":
		style.Cursor = value
	case "border":
		w, s, c := parseBorderShorthand(value)
		style.BorderTopWidth = w
		style.BorderRightWidth = w
		style.BorderBottomWidth = w
		style.BorderLeftWidth = w
		style.BorderTopStyle = s
		style.BorderRightStyle = s
		style.BorderBottomStyle = s
		style.BorderLeftStyle = s
		style.BorderTopColor = c
		style.BorderRightColor = c
		style.BorderBottomColor = c
		style.BorderLeftColor = c
	case "border-width":
		w := ParseSize(value)
		style.BorderTopWidth = w
		style.BorderRightWidth = w
		style.BorderBottomWidth = w
		style.BorderLeftWidth = w
	case "border-color":
		if c := ParseColor(value); c != nil {
			style.BorderTopColor = c
			style.BorderRightColor = c
			style.BorderBottomColor = c
			style.BorderLeftColor = c
		}
	case "border-style":
		style.BorderTopStyle = value
		style.BorderRightStyle = value
		style.BorderBottomStyle = value
		style.BorderLeftStyle = value
	case "border-top":
		w, s, c := parseBorderShorthand(value)
		style.BorderTopWidth = w
		style.BorderTopStyle = s
		style.BorderTopColor = c
	case "border-right":
		w, s, c := parseBorderShorthand(value)
		style.BorderRightWidth = w
		style.BorderRightStyle = s
		style.BorderRightColor = c
	case "border-bottom":
		w, s, c := parseBorderShorthand(value)
		style.BorderBottomWidth = w
		style.BorderBottomStyle = s
		style.BorderBottomColor = c
	case "border-left":
		w, s, c := parseBorderShorthand(value)
		style.BorderLeftWidth = w
		style.BorderLeftStyle = s
		style.BorderLeftColor = c
	}
}
