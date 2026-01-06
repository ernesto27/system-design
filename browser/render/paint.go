package render

import (
	"browser/dom"
	"browser/layout"
	"fmt"
	"image/color"
	"strings"
)

// Colors are defined in colors.go

// Font sizes
var (
	SizeH1     float32 = 32
	SizeH2     float32 = 24
	SizeH3     float32 = 18
	SizeH4     float32 = 16
	SizeH5     float32 = 14
	SizeH6     float32 = 12
	SizeNormal float32 = 16
	SizeSmall  float32 = 12
)

// TextStyle holds inherited text styling
type TextStyle struct {
	Color          color.Color
	Size           float32
	Bold           bool
	Italic         bool
	Monospace      bool
	TextDecoration string
	Opacity        float64
	Visibility     string
}

type DrawInput struct {
	layout.Rect
	Placeholder string
	Value       string
	IsFocused   bool
	IsPassword  bool
	IsDisabled  bool
	IsReadonly  bool
}

type DrawButton struct {
	layout.Rect
	Text       string
	IsDisabled bool
}

type DrawTextarea struct {
	layout.Rect
	Placeholder string
	Value       string
	IsFocused   bool
	IsDisabled  bool
	IsReadonly  bool
}

type DrawSelect struct {
	layout.Rect
	Options       []string // List of option texts
	SelectedValue string   // Currently selected value
	IsOpen        bool     // Is dropdown open?
	IsDisabled    bool
	IsReadonly    bool
}

type DrawRadio struct {
	layout.Rect
	IsChecked  bool
	IsDisabled bool
}

type DrawCheckbox struct {
	layout.Rect
	IsChecked  bool
	IsDisabled bool
	IsReadonly bool
}

// InputState holds all interactive form state for rendering
type InputState struct {
	InputValues    map[*dom.Node]string // Text input values
	FocusedNode    *dom.Node            // Currently focused input/textarea
	OpenSelectNode *dom.Node            // Which select dropdown is open
	RadioValues    map[string]*dom.Node // Selected radio per group (key: name attr)
	CheckboxValues map[*dom.Node]bool   // Checked state per check
}

// DefaultStyle returns the default text style
func DefaultStyle() TextStyle {
	return TextStyle{
		Color:   ColorBlack,
		Size:    SizeNormal,
		Bold:    false,
		Italic:  false,
		Opacity: 1.0,
	}
}

// applyOpacity returns a color with opacity applied to alpha channel
func applyOpacity(c color.Color, opacity float64) color.Color {
	if opacity >= 1.0 {
		return c
	}
	r, g, b, a := c.RGBA()
	newAlpha := uint8(float64(a>>8) * opacity)
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), newAlpha}
}

type DisplayCommand any

type DrawRect struct {
	layout.Rect
	Color color.Color
}

type DrawText struct {
	Text          string
	X, Y          float64
	Width         float64
	Color         color.Color
	Size          float32
	Bold          bool
	Italic        bool
	Monospace     bool
	Underline     bool
	Strikethrough bool
}

type DrawImage struct {
	layout.Rect
	URL string
}

type DrawHR struct {
	layout.Rect
}

func BuildDisplayList(root *layout.LayoutBox) []DisplayCommand {
	var commands []DisplayCommand

	// Calculate actual content height from layout tree
	contentHeight := root.Rect.Y + root.Rect.Height
	if contentHeight < 600 {
		contentHeight = 600 // Minimum height
	}

	commands = append(commands, DrawRect{
		Rect:  layout.Rect{X: 0, Y: 0, Width: 3000, Height: contentHeight}, // Wide enough for most screens
		Color: color.White,
	})

	paintLayoutBox(root, &commands, DefaultStyle())

	return commands
}

func BuildDisplayListWithInputs(root *layout.LayoutBox, state InputState) []DisplayCommand {
	var commands []DisplayCommand

	contentHeight := root.Rect.Y + root.Rect.Height
	if contentHeight < 600 {
		contentHeight = 600
	}

	commands = append(commands, DrawRect{
		Rect:  layout.Rect{X: 0, Y: 0, Width: 3000, Height: contentHeight},
		Color: color.White,
	})

	paintLayoutBoxWithInputs(root, &commands, DefaultStyle(), state)

	return commands
}

func paintLayoutBoxWithInputs(box *layout.LayoutBox, commands *[]DisplayCommand, style TextStyle, state InputState) {
	currentStyle := style

	// Apply inline styles from CSS
	if box.Style.Color != nil {
		currentStyle.Color = box.Style.Color
	}
	if box.Style.FontSize > 0 {
		currentStyle.Size = float32(box.Style.FontSize)
	}
	if box.Style.Bold {
		currentStyle.Bold = true
	}
	if box.Style.Italic {
		currentStyle.Italic = true
	}
	if box.Style.TextDecoration != "" {
		currentStyle.TextDecoration = box.Style.TextDecoration
	}
	if box.Style.Opacity > 0 {
		currentStyle.Opacity = box.Style.Opacity
	}
	if box.Style.Visibility != "" {
		currentStyle.Visibility = box.Style.Visibility
	}

	isHidden := currentStyle.Visibility == "hidden"

	// Draw background if set
	if box.Style.BackgroundColor != nil && !isHidden {
		*commands = append(*commands, DrawRect{
			Rect:  box.Rect,
			Color: applyOpacity(box.Style.BackgroundColor, currentStyle.Opacity),
		})
	}

	// Draw borders if set
	if !isHidden {
		if box.Style.BorderTopWidth > 0 && box.Style.BorderTopStyle != "none" && box.Style.BorderTopColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width, Height: box.Style.BorderTopWidth},
				Color: applyOpacity(box.Style.BorderTopColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderBottomWidth > 0 && box.Style.BorderBottomStyle != "none" && box.Style.BorderBottomColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y + box.Rect.Height - box.Style.BorderBottomWidth, Width: box.Rect.Width, Height: box.Style.BorderBottomWidth},
				Color: applyOpacity(box.Style.BorderBottomColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderLeftWidth > 0 && box.Style.BorderLeftStyle != "none" && box.Style.BorderLeftColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Style.BorderLeftWidth, Height: box.Rect.Height},
				Color: applyOpacity(box.Style.BorderLeftColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderRightWidth > 0 && box.Style.BorderRightStyle != "none" && box.Style.BorderRightColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X + box.Rect.Width - box.Style.BorderRightWidth, Y: box.Rect.Y, Width: box.Style.BorderRightWidth, Height: box.Rect.Height},
				Color: applyOpacity(box.Style.BorderRightColor, currentStyle.Opacity),
			})
		}
	}

	// Apply tag-based styles
	if box.Node != nil {
		switch box.Node.TagName {
		case dom.TagH1:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH1
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH2:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH2
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH3:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH3
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH4:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH4
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH5:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH5
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH6:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH6
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagA:
			if box.Style.Color == nil {
				currentStyle.Color = ColorLink
			}
			if box.Style.TextDecoration == "" {
				currentStyle.TextDecoration = "underline"
			}
		case dom.TagStrong, dom.TagB:
			currentStyle.Bold = true
		case dom.TagEm, dom.TagI:
			currentStyle.Italic = true
		case dom.TagSmall:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeSmall
			}
		case dom.TagU:
			currentStyle.TextDecoration = "underline"
		case dom.TagPre:
			currentStyle.Monospace = true
			if box.Style.BackgroundColor == nil && !isHidden {
				*commands = append(*commands, DrawRect{
					Rect:  box.Rect,
					Color: color.RGBA{245, 245, 245, 255},
				})
			}
		case dom.TagTH:
			currentStyle.Bold = true
		}
	}

	// Draw text
	if box.Type == layout.TextBox && box.Text != "" && !isHidden {
		text := box.Text
		if isListItem, isOrdered, index := getListInfo(box); isListItem {
			if isOrdered {
				text = fmt.Sprintf("%d. %s", index, text)
			} else {
				text = "• " + text
			}
		}

		if currentStyle.Monospace && strings.Contains(text, "\n") {
			lines := strings.Split(text, "\n")
			lineHeight := float64(currentStyle.Size) * 1.5
			y := box.Rect.Y
			for _, line := range lines {
				*commands = append(*commands, DrawText{
					Text: line, X: box.Rect.X, Y: y, Width: box.Rect.Width,
					Size: currentStyle.Size, Color: applyOpacity(currentStyle.Color, currentStyle.Opacity),
					Bold: currentStyle.Bold, Italic: currentStyle.Italic, Monospace: currentStyle.Monospace,
					Underline:     currentStyle.TextDecoration == "underline",
					Strikethrough: currentStyle.TextDecoration == "line-through",
				})
				y += lineHeight
			}
		} else {
			*commands = append(*commands, DrawText{
				Text: text, X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width,
				Size: currentStyle.Size, Color: applyOpacity(currentStyle.Color, currentStyle.Opacity),
				Bold: currentStyle.Bold, Italic: currentStyle.Italic, Monospace: currentStyle.Monospace,
				Underline:     currentStyle.TextDecoration == "underline",
				Strikethrough: currentStyle.TextDecoration == "line-through",
			})
		}
	}

	// Draw image
	if box.Type == layout.ImageBox && box.Node != nil && !isHidden {
		if src := box.Node.Attributes["src"]; src != "" {
			*commands = append(*commands, DrawImage{
				Rect: box.Rect,
				URL:  src,
			})
		}
	}

	if box.Type == layout.HRBox && !isHidden {
		*commands = append(*commands, DrawHR{
			Rect: box.Rect,
		})
	}

	// Input with state - use DOM node for lookup (stable across reflow)
	if box.Type == layout.InputBox && box.Node != nil && !isHidden {
		value := state.InputValues[box.Node]
		isFocused := (box.Node == state.FocusedNode)

		placeholder := box.Node.Attributes["placeholder"]
		if placeholder == "" {
			placeholder = box.Node.Attributes["value"]
		}

		inputType := strings.ToLower(box.Node.Attributes["type"])
		isPassword := (inputType == "password")

		_, isDisabled := box.Node.Attributes["disabled"]
		_, isReadonly := box.Node.Attributes["readonly"]

		if isDisabled {
			isFocused = false
		}

		*commands = append(*commands, DrawInput{
			Rect:        box.Rect,
			Placeholder: placeholder,
			Value:       value,
			IsFocused:   isFocused,
			IsPassword:  isPassword,
			IsDisabled:  isDisabled,
			IsReadonly:  isReadonly,
		})
	}

	if box.Type == layout.ButtonBox && !isHidden {
		*commands = append(*commands, DrawButton{
			Rect: box.Rect,
			Text: getButtonTextFromBox(box),
		})
	}

	if box.Type == layout.TextareaBox && box.Node != nil && !isHidden {
		value := state.InputValues[box.Node]
		isFocused := (box.Node == state.FocusedNode)

		_, isDisabled := box.Node.Attributes["disabled"]
		_, isReadonly := box.Node.Attributes["readonly"]

		if isDisabled {
			isFocused = false
		}

		*commands = append(*commands, DrawTextarea{
			Rect:        box.Rect,
			Placeholder: box.Node.Attributes["placeholder"],
			Value:       value,
			IsFocused:   isFocused,
			IsDisabled:  isDisabled,
			IsReadonly:  isReadonly,
		})
	}

	if box.Type == layout.SelectBox && box.Node != nil && !isHidden {
		// Get options from <option> children
		var options []string
		fmt.Printf("Select box found, children: %d\n", len(box.Node.Children))
		for _, child := range box.Node.Children {
			fmt.Printf("  Child: TagName=%s, Type=%d\n", child.TagName, child.Type)
			if child.TagName == "option" {
				for _, textNode := range child.Children {
					fmt.Printf("    TextNode: Type=%d, Text=%q\n", textNode.Type, textNode.Text)
					if textNode.Type == dom.Text {
						options = append(options, textNode.Text)
						break
					}
				}
			}
		}

		selectedValue := state.InputValues[box.Node]
		_, isDisabled := box.Node.Attributes["disabled"]

		// Disabled selects cannot be open
		isOpen := (box.Node == state.OpenSelectNode) && !isDisabled
		fmt.Printf("Select: options=%v, isOpen=%v, openSelectNode=%p, box.Node=%p\n", options, isOpen, state.OpenSelectNode, box.Node)

		*commands = append(*commands, DrawSelect{
			Rect:          box.Rect,
			Options:       options,
			SelectedValue: selectedValue,
			IsOpen:        isOpen,
			IsDisabled:    isDisabled,
		})
	}

	// Radio button
	if box.Type == layout.RadioBox && box.Node != nil && !isHidden {
		name := box.Node.Attributes["name"]
		isChecked := false
		if name != "" && state.RadioValues != nil {
			isChecked = (state.RadioValues[name] == box.Node)
		}
		// Fallback to HTML checked attribute if no runtime state
		if !isChecked {
			_, isChecked = box.Node.Attributes["checked"]
		}

		_, isDisabled := box.Node.Attributes["disabled"]

		*commands = append(*commands, DrawRadio{
			Rect:       box.Rect,
			IsChecked:  isChecked,
			IsDisabled: isDisabled,
		})
	}

	if box.Type == layout.CheckboxBox && box.Node != nil && !isHidden {
		isChecked := false
		if state.CheckboxValues != nil {
			isChecked = state.CheckboxValues[box.Node]
		}

		_, isDisabled := box.Node.Attributes["disabled"]
		*commands = append(*commands, DrawCheckbox{
			Rect:       box.Rect,
			IsChecked:  isChecked,
			IsDisabled: isDisabled,
		})
	}

	// Draw table cell border
	if box.Type == layout.TableCellBox {
		borderColor := color.Gray{Y: 180}
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width, Height: 1}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y + box.Rect.Height - 1, Width: box.Rect.Width, Height: 1}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: 1, Height: box.Rect.Height}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X + box.Rect.Width - 1, Y: box.Rect.Y, Width: 1, Height: box.Rect.Height}, Color: borderColor})
	}

	// Paint children with input state
	for _, child := range box.Children {
		paintLayoutBoxWithInputs(child, commands, currentStyle, state)
	}
}

func paintLayoutBox(box *layout.LayoutBox, commands *[]DisplayCommand, style TextStyle) {
	// Delegate to the stateful version with empty state
	paintLayoutBoxWithInputs(box, commands, style, InputState{})
}

// getListInfo returns (isListItem, isOrdered, itemIndex)
func getListInfo(box *layout.LayoutBox) (bool, bool, int) {
	// Check if parent is <li>
	if box.Parent == nil || box.Parent.Node == nil {
		return false, false, 0
	}
	if box.Parent.Node.TagName != dom.TagLI {
		return false, false, 0
	}

	li := box.Parent

	// Check if grandparent is <ul> or <ol>
	if li.Parent == nil || li.Parent.Node == nil {
		return false, false, 0
	}

	listTag := li.Parent.Node.TagName
	if listTag != dom.TagUL && listTag != dom.TagOL {
		return false, false, 0
	}

	isOrdered := listTag == dom.TagOL

	// Count which <li> index this is
	index := 1
	for _, sibling := range li.Parent.Children {
		if sibling == li {
			break
		}
		if sibling.Node != nil && sibling.Node.TagName == dom.TagLI {
			index++
		}
	}

	return true, isOrdered, index
}

func getButtonTextFromBox(box *layout.LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == layout.TextBox {
			return child.Text
		}
	}
	if box.Node != nil {
		if val, ok := box.Node.Attributes["value"]; ok {
			return val
		}
	}
	return "Button"
}
