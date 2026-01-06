package render

import "image/color"

// Base colors
var (
	ColorWhite = color.RGBA{255, 255, 255, 255}
	ColorBlack = color.RGBA{0, 0, 0, 255}
)

// Link colors
var (
	ColorLink = color.RGBA{0, 0, 238, 255} // Blue
)

// Border colors
var (
	ColorBorder         = color.RGBA{180, 180, 180, 255} // Normal border
	ColorBorderFocused  = color.RGBA{0, 120, 215, 255}   // Focused border (blue)
	ColorBorderDisabled = color.RGBA{200, 200, 200, 255} // Disabled border
)

// Input background colors
var (
	ColorInputBg         = color.RGBA{255, 255, 255, 255} // White
	ColorInputBgDisabled = color.RGBA{240, 240, 240, 255} // Light gray
)

// Button colors
var (
	ColorButtonBg               = color.RGBA{225, 225, 225, 255}
	ColorButtonBgDisabled       = color.RGBA{240, 240, 240, 255}
	ColorButtonHighlight        = color.RGBA{255, 255, 255, 255}
	ColorButtonHighlightDisabled = color.RGBA{250, 250, 250, 255}
	ColorButtonShadow           = color.RGBA{150, 150, 150, 255}
	ColorButtonShadowDisabled   = color.RGBA{200, 200, 200, 255}
)

// Text colors
var (
	ColorText                = color.RGBA{0, 0, 0, 255}       // Black
	ColorTextDisabled        = color.RGBA{160, 160, 160, 255}
	ColorPlaceholder         = color.RGBA{150, 150, 150, 255}
	ColorPlaceholderDisabled = color.RGBA{180, 180, 180, 255}
)

// Checkbox and Radio colors
var (
	ColorCheckboxBorder         = color.RGBA{100, 100, 100, 255}
	ColorCheckboxBorderDisabled = color.RGBA{180, 180, 180, 255}
	ColorAccent                 = color.RGBA{0, 120, 215, 255}   // Blue accent
	ColorAccentDisabled         = color.RGBA{160, 160, 160, 255} // Gray accent
)

// Select dropdown colors
var (
	ColorSelectArrow     = color.RGBA{100, 100, 100, 255}
	ColorSelectHighlight = color.RGBA{0, 120, 215, 40} // Translucent blue
)

// Background colors
var (
	ColorPageBackground = color.RGBA{240, 240, 240, 255} // Light gray
	ColorPreBackground  = color.RGBA{245, 245, 245, 255} // Very light gray for <pre>
)

// Table colors
var (
	ColorTableBorder = color.Gray{Y: 180}
)

// HR color
var (
	ColorHR = color.Gray{Y: 180}
)
