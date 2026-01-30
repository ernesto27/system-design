package layout

import (
	"browser/css"
	"browser/dom"
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultMargin      = 8.0
	DefaultImageWidth  = 200.0
	DefaultImageHeight = 150.0
)

// getDefaultLineHeight returns default line heights for different elements
func getDefaultLineHeight(tagName string) float64 {
	switch tagName {
	case dom.TagH1:
		return 40.0
	case dom.TagH2:
		return 32.0
	case dom.TagH3:
		return 26.0
	case dom.TagH4:
		return 24.0
	case dom.TagH5:
		return 22.0
	case dom.TagH6:
		return 20.0
	case dom.TagSmall:
		return 18.0
	default:
		return 24.0
	}
}

// Font sizes for text measurement (should match render/paint.go)
func getFontSize(tagName string) float64 {
	switch tagName {
	case dom.TagH1:
		return 32.0
	case dom.TagH2:
		return 24.0
	case dom.TagH3:
		return 18.0
	case dom.TagH4:
		return 16.0
	case dom.TagH5:
		return 14.0
	case dom.TagH6:
		return 12.0
	case dom.TagSmall:
		return 12.0
	default:
		return 16.0
	}
}

func ComputeLayout(root *LayoutBox, containerWidth float64) {
	computeBlockLayout(root, containerWidth, 0, 0, "")
}

func computeBlockLayout(box *LayoutBox, containerWidth float64, startX, startY float64, parentTag string) {
	// Separate positioned children from normal flow
	var positionedChildren []*LayoutBox
	var floatedChildren []*LayoutBox
	var normalChildren []*LayoutBox

	for _, child := range box.Children {
		if child.Position == "absolute" {
			positionedChildren = append(positionedChildren, child)
		} else if child.Float == "left" || child.Float == "right" {
			floatedChildren = append(floatedChildren, child)
		} else {
			normalChildren = append(normalChildren, child)
		}
	}
	box.Children = normalChildren

	box.Rect.X = startX
	box.Rect.Y = startY
	box.Rect.Width = containerWidth

	if box.Style.Width > 0 {
		box.Rect.Width = box.Style.Width
	}

	if box.Style.MinWidth > 0 && box.Rect.Width < box.Style.MinWidth {
		box.Rect.Width = box.Style.MinWidth
	}

	if box.Style.MaxWidth > 0 && box.Rect.Width > box.Style.MaxWidth {
		box.Rect.Width = box.Style.MaxWidth
	}

	innerX := startX
	innerWidth := box.Rect.Width

	// Get current tag name
	currentTag := ""
	if box.Node != nil {
		currentTag = box.Node.TagName
	}

	// Body gets margin
	if currentTag == dom.TagBody {
		box.Margin = EdgeSizes{
			Top: DefaultMargin, Right: DefaultMargin,
			Bottom: DefaultMargin, Left: DefaultMargin,
		}
		innerX = startX + DefaultMargin
		innerWidth = containerWidth - (DefaultMargin * 2)
	}

	// Lists get indentation
	if currentTag == dom.TagUL || currentTag == dom.TagOL {
		innerX = startX + 20
		innerWidth = containerWidth - 20
	}

	if currentTag == dom.TagBlockquote {
		innerX = startX + 30
		innerWidth = containerWidth - 30
	}

	if currentTag == dom.TagDD {
		innerX = startX + 40
		innerWidth = containerWidth - 40
	}

	// Fieldset default styling
	if box.Type == FieldsetBox {
		box.Padding = EdgeSizes{Top: 10, Right: 10, Bottom: 10, Left: 10}
		box.Style.BorderTopWidth = 1
		box.Style.BorderRightWidth = 1
		box.Style.BorderBottomWidth = 1
		box.Style.BorderLeftWidth = 1
	}

	// Default margins for block elements
	switch currentTag {
	case dom.TagP:
		box.Margin.Top = 12
		box.Margin.Bottom = 12
	case dom.TagH1:
		box.Margin.Top = 6
		box.Margin.Bottom = 6
	case dom.TagH2:
		box.Margin.Top = 5
		box.Margin.Bottom = 5
	case dom.TagH3:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	case dom.TagH4, dom.TagH5, dom.TagH6:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	case dom.TagUL, dom.TagOL:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	}

	// Apply CSS margins from inline style (override defaults)
	if box.Style.MarginTop > 0 {
		box.Margin.Top = box.Style.MarginTop
	}
	if box.Style.MarginBottom > 0 {
		box.Margin.Bottom = box.Style.MarginBottom
	}

	// Handle auto margins for horizontal centering
	if box.Style.MarginLeftAuto && box.Style.MarginRightAuto && box.Style.Width > 0 {
		// Both auto = center horizontally
		leftover := containerWidth - box.Rect.Width
		if leftover > 0 {
			autoMargin := leftover / 2
			box.Rect.X = startX + autoMargin
			innerX = box.Rect.X
			box.Margin.Left = autoMargin
			box.Margin.Right = autoMargin
		}
	} else {
		if box.Style.MarginLeft > 0 {
			innerX += box.Style.MarginLeft
			innerWidth -= box.Style.MarginLeft
		}
		if box.Style.MarginRight > 0 {
			innerWidth -= box.Style.MarginRight
		}
	}

	// Apply CSS padding from inline style
	if box.Style.PaddingTop > 0 {
		box.Padding.Top = box.Style.PaddingTop
	}
	if box.Style.PaddingBottom > 0 {
		box.Padding.Bottom = box.Style.PaddingBottom
	}
	if box.Style.PaddingLeft > 0 {
		box.Padding.Left = box.Style.PaddingLeft
		innerX += box.Style.PaddingLeft
		innerWidth -= box.Style.PaddingLeft
	}
	if box.Style.PaddingRight > 0 {
		box.Padding.Right = box.Style.PaddingRight
		innerWidth -= box.Style.PaddingRight
	}

	// Apply border widths to inner content area
	if box.Style.BorderLeftWidth > 0 {
		innerX += box.Style.BorderLeftWidth
		innerWidth -= box.Style.BorderLeftWidth
	}
	if box.Style.BorderRightWidth > 0 {
		innerWidth -= box.Style.BorderRightWidth
	}

	yOffset := startY + box.Margin.Top + box.Padding.Top + box.Style.BorderTopWidth

	// Line state for inline flow
	currentX := innerX
	lineStartY := yOffset
	lineHeight := 0.0
	var lineBoxes []*LayoutBox

	// Handle legend for fieldset
	var legendBox *LayoutBox
	if box.Type == FieldsetBox {
		for i, child := range box.Children {
			if child.Type == LegendBox {
				legendBox = child
				// Remove legend from normal children flow
				box.Children = append(box.Children[:i], box.Children[i+1:]...)
				break
			}
		}

		// Position legend on the border
		if legendBox != nil {
			legendText := GetLegendText(legendBox)
			legendWidth := MeasureText(legendText, 16) + 16 // 8px padding each side
			legendHeight := 20.0

			legendBox.Rect.X = innerX + 12                              // 12px from left edge
			legendBox.Rect.Y = startY + box.Margin.Top - legendHeight/2 // Centered on border
			legendBox.Rect.Width = legendWidth
			legendBox.Rect.Height = legendHeight

			// Layout legend's children (the text)
			textX := legendBox.Rect.X + 4
			for _, child := range legendBox.Children {
				if child.Type == TextBox {
					child.Rect.X = textX
					child.Rect.Y = legendBox.Rect.Y
					child.Rect.Width = MeasureText(child.Text, 16)
					child.Rect.Height = legendHeight
				}
			}

			// Add legend back to children so paint.go can find it
			box.Children = append([]*LayoutBox{legendBox}, box.Children...)
		}
	}

	for _, child := range box.Children {
		// Skip LegendBox - already positioned above
		if child.Type == LegendBox {
			continue
		}

		var childWidth, childHeight float64

		switch child.Type {
		case TextBox:
			fontSize := getFontSize(parentTag)
			// Check if inside a <pre> element
			if isInsidePre(child) {
				// Handle multi-line preformatted text
				lines := strings.Split(child.Text, "\n")
				lineHeight := fontSize * 1.5 // Match render/paint.go line height

				// Find the widest line
				maxWidth := 0.0
				for _, line := range lines {
					w := MeasureText(line, fontSize)
					if w > maxWidth {
						maxWidth = w
					}
				}

				childWidth = maxWidth
				childHeight = float64(len(lines)) * lineHeight
			} else {
				// Wrap text to fit container width
				child.WrappedLines = WrapText(child.Text, fontSize, innerWidth)

				lineHeight := getLineHeightFromStyle(box.Style, parentTag)
				numLines := len(child.WrappedLines)
				if numLines == 0 {
					numLines = 1
				}

				// Width is the widest wrapped line
				maxLineWidth := 0.0
				for _, line := range child.WrappedLines {
					w := MeasureText(line, fontSize)
					if w > maxLineWidth {
						maxLineWidth = w
					}
				}
				childWidth = maxLineWidth
				childHeight = float64(numLines) * lineHeight
			}

		case InlineBox:
			// Compute inline box size from its content
			childWidth, childHeight = computeInlineSize(child, parentTag)

		case ImageBox:
			childWidth, childHeight = getImageSize(child.Node)
		case InputBox:
			childWidth = 200.0
			childHeight = 28.0
		case RadioBox:
			childWidth = 20.0
			childHeight = 20.0
		case CheckboxBox:
			childWidth = 20.0
			childHeight = 20.0
		case ButtonBox:
			buttonText := getButtonText(child)
			fontSize := getFontSize(parentTag)
			childWidth = MeasureText(buttonText, fontSize) + 24.0
			childHeight = 32.0
		case TextareaBox:
			childWidth = 300.0
			childHeight = 80.0
		case SelectBox:
			childWidth = 200.0
			childHeight = 28.0
		case FileInputBox:
			childWidth = 250.0
			childHeight = 32.0

		case HRBox:
			// Block element - flush line first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
			}
			child.Rect.X = innerX
			child.Rect.Y = yOffset + 8
			child.Rect.Width = innerWidth
			child.Rect.Height = 2
			yOffset += 18
			// Reset line state
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		case BRBox:
			// Line break - flush current line
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
			} else {
				yOffset += getLineHeightFromStyle(box.Style, parentTag)
			}
			child.Rect.X = currentX
			child.Rect.Y = yOffset
			child.Rect.Width = 0
			child.Rect.Height = 0
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		case TableBox:
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			computeTableLayout(child, innerWidth, innerX, yOffset)
			yOffset += child.Rect.Height
			// Reset line state
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		default:
			// Block element - flush line first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
				lineStartY = yOffset
				lineHeight = 0
			}
			currentX = innerX

			childTag := ""
			if child.Node != nil {
				childTag = child.Node.TagName
			}
			computeBlockLayout(child, innerWidth, innerX, yOffset, childTag)
			yOffset += child.Rect.Height
			lineStartY = yOffset
			continue
		}

		// Inline element - check if we need to wrap
		if currentX+childWidth > innerX+innerWidth && currentX > innerX {
			// Wrap to new line - apply alignment first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			yOffset = lineStartY + lineHeight
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
		}

		// Position inline element
		child.Rect.X = currentX
		child.Rect.Y = lineStartY
		child.Rect.Width = childWidth
		child.Rect.Height = childHeight

		// For InlineBox, position its children within it
		if child.Type == InlineBox {
			layoutInlineChildren(child, parentTag)
		}

		// Track this element for alignment
		lineBoxes = append(lineBoxes, child)

		// Advance horizontal position
		currentX += childWidth
		if childHeight > lineHeight {
			lineHeight = childHeight
		}
	}

	// Final line
	applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
	if lineHeight > 0 {
		yOffset = lineStartY + lineHeight
	}

	if box.Style.Height > 0 {
		box.Rect.Height = box.Style.Height
	} else {
		box.Rect.Height = yOffset - startY + box.Margin.Bottom + box.Padding.Bottom + box.Style.BorderBottomWidth
	}

	if box.Style.MinHeight > 0 && box.Rect.Height < box.Style.MinHeight {
		box.Rect.Height = box.Style.MinHeight
	}

	if box.Style.MaxHeight > 0 && box.Rect.Height > box.Style.MaxHeight {
		box.Rect.Height = box.Style.MaxHeight
	}

	// Position absolute children
	for _, child := range positionedChildren {
		childWidth := child.Style.Width
		if childWidth <= 0 {
			childWidth = containerWidth
		}

		// First, compute layout to determine child dimensions
		computeBlockLayout(child, childWidth, 0, 0, "")

		// Calculate X position
		childX := startX
		if child.Left > 0 {
			childX = startX + child.Left
		} else if child.Right > 0 {
			childX = startX + box.Rect.Width - child.Right - child.Rect.Width
		}

		// Calculate Y position
		childY := startY
		if child.Top > 0 {
			childY = startY + child.Top
		} else if child.Bottom > 0 {
			childY = startY + box.Rect.Height - child.Bottom - child.Rect.Height
		}

		// Apply final position by offsetting the entire subtree
		offsetBox(child, childX, childY)

		box.Children = append(box.Children, child)
	}

	// Position floated children (inside padding area)
	leftFloatX := innerX
	rightFloatX := innerX + innerWidth
	floatY := startY + box.Padding.Top + box.Style.BorderTopWidth

	for _, child := range floatedChildren {
		childWidth := child.Style.Width
		if childWidth <= 0 {
			childWidth = 100 // Default width for floats without explicit width
		}

		// Compute layout to determine dimensions
		computeBlockLayout(child, childWidth, 0, 0, "")

		switch child.Float {
		case "left":
			offsetBox(child, leftFloatX, floatY)
			leftFloatX += child.Rect.Width
		case "right":
			offsetBox(child, rightFloatX-child.Rect.Width, floatY)
			rightFloatX -= child.Rect.Width
		}

		box.Children = append(box.Children, child)
	}

}

// offsetBox moves a box and all its children by (dx, dy)
func offsetBox(box *LayoutBox, dx, dy float64) {
	box.Rect.X += dx
	box.Rect.Y += dy
	for _, child := range box.Children {
		offsetBox(child, dx, dy)
	}
}

// applyLineAlignment repositions inline elements based on text-align
func applyLineAlignment(lineBoxes []*LayoutBox, innerX, innerWidth float64, textAlign string) {
	if len(lineBoxes) == 0 || textAlign == "" || textAlign == "left" {
		return
	}

	// Calculate actual line width used
	lineWidth := 0.0
	for _, b := range lineBoxes {
		lineWidth += b.Rect.Width
	}

	// Calculate offset based on textAlign
	var offset float64
	switch textAlign {
	case "center":
		offset = (innerWidth - lineWidth) / 2
	case "right":
		offset = innerWidth - lineWidth
	}

	// Apply offset to all boxes
	for _, b := range lineBoxes {
		b.Rect.X += offset
	}
}

// computeInlineSize calculates the total size of an inline box from its children
func computeInlineSize(box *LayoutBox, parentTag string) (float64, float64) {
	var totalWidth float64
	var maxHeight float64

	// Use the inline element's tag if it affects font size (e.g., <small>)
	tagForSize := parentTag
	if box.Node != nil && box.Node.TagName == dom.TagSmall {
		tagForSize = dom.TagSmall
	}

	for _, child := range box.Children {
		var w, h float64
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(tagForSize)
			text := css.ApplyTextTransform(child.Text, box.Style.TextTransform)
			w = MeasureText(text, fontSize)
			h = getLineHeightFromStyle(box.Style, tagForSize)
		case InlineBox:
			w, h = computeInlineSize(child, parentTag)
		case ImageBox:
			w, h = getImageSize(child.Node)
		case CheckboxBox, RadioBox:
			w = 20.0
			h = 20.0
		}
		totalWidth += w
		if h > maxHeight {
			maxHeight = h
		}
	}

	return totalWidth, maxHeight
}

// layoutInlineChildren positions children within an inline box
func layoutInlineChildren(box *LayoutBox, parentTag string) {
	// Use the inline element's tag if it affects font size (e.g., <small>)
	tagForSize := parentTag
	if box.Node != nil && box.Node.TagName == dom.TagSmall {
		tagForSize = dom.TagSmall
	}

	// Calculate vertical offset for baseline alignment
	parentLineHeight := getDefaultLineHeight(parentTag)
	childLineHeight := getLineHeightFromStyle(box.Style, tagForSize)
	baselineOffset := (parentLineHeight - childLineHeight) / 2

	offsetX := 0.0
	for _, child := range box.Children {
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(tagForSize)
			text := css.ApplyTextTransform(child.Text, box.Style.TextTransform)
			w := MeasureText(text, fontSize)
			h := getLineHeightFromStyle(box.Style, tagForSize)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y + baselineOffset
			child.Rect.Width = w
			child.Rect.Height = h
			offsetX += w
		case InlineBox:
			w, h := computeInlineSize(child, parentTag)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y + baselineOffset
			child.Rect.Width = w
			child.Rect.Height = h
			layoutInlineChildren(child, parentTag)
			offsetX += w
		case ImageBox:
			w, h := getImageSize(child.Node)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
			child.Rect.Width = w
			child.Rect.Height = h
			offsetX += w
		case CheckboxBox, RadioBox:
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
			child.Rect.Width = 20.0
			child.Rect.Height = 20.0
			offsetX += 20.0
		}
	}
}

// computeTableLayout handles table, row, and cell positioning
func computeTableLayout(table *LayoutBox, containerWidth float64, startX, startY float64) {
	table.Rect.X = startX
	table.Rect.Y = startY
	table.Rect.Width = containerWidth

	// Collect all rows (may be direct children or inside tbody/thead/tfoot)
	var rows []*LayoutBox
	for _, child := range table.Children {
		switch child.Type {
		case TableRowBox:
			rows = append(rows, child)
		case TableBox:
			// This is tbody/thead/tfoot - get rows from inside
			for _, grandchild := range child.Children {
				if grandchild.Type == TableRowBox {
					rows = append(rows, grandchild)
				}
			}
		}
	}

	// Count max columns in any row
	numCols := 0
	for _, row := range rows {
		cellCount := 0
		for _, cell := range row.Children {
			if cell.Type == TableCellBox {
				cellCount++
			}
		}
		if cellCount > numCols {
			numCols = cellCount
		}
	}

	if numCols == 0 {
		table.Rect.Height = 0
		return
	}

	// Simple approach: equal column widths
	cellPadding := 8.0
	colWidth := containerWidth / float64(numCols)

	yOffset := startY

	// Handle caption first (renders above the table rows, centered)
	for _, child := range table.Children {
		if child.Type == TableCaptionBox {
			child.Rect.X = startX
			child.Rect.Y = yOffset
			child.Rect.Width = containerWidth

			// Center the caption text
			captionHeight := 24.0
			for _, textChild := range child.Children {
				if textChild.Type == TextBox {
					fontSize := 16.0
					textWidth := MeasureText(textChild.Text, fontSize)
					textChild.Rect.X = startX + (containerWidth-textWidth)/2 // centered
					textChild.Rect.Y = yOffset
					textChild.Rect.Width = textWidth
					textChild.Rect.Height = 24.0
				}
			}
			child.Rect.Height = captionHeight
			yOffset += captionHeight + 4
		}
	}

	// Layout each row
	for _, row := range rows {
		row.Rect.X = startX
		row.Rect.Y = yOffset
		row.Rect.Width = containerWidth

		// Layout cells in this row
		rowHeight := 24.0 // minimum row height
		xOffset := startX

		for _, cell := range row.Children {
			if cell.Type != TableCellBox {
				continue
			}

			cell.Rect.X = xOffset
			cell.Rect.Y = yOffset
			cell.Rect.Width = colWidth

			// Compute cell content height
			cellHeight := computeCellContent(cell, colWidth-cellPadding*2, xOffset+cellPadding, yOffset+cellPadding)
			cell.Rect.Height = cellHeight + cellPadding*2

			if cell.Rect.Height > rowHeight {
				rowHeight = cell.Rect.Height
			}

			xOffset += colWidth
		}

		// Set all cells to same height (tallest cell)
		for _, cell := range row.Children {
			if cell.Type == TableCellBox {
				cell.Rect.Height = rowHeight
			}
		}

		row.Rect.Height = rowHeight
		yOffset += rowHeight
	}

	table.Rect.Height = yOffset - startY
}

// computeCellContent layouts the content inside a table cell
func computeCellContent(cell *LayoutBox, width float64, startX, startY float64) float64 {
	yOffset := startY

	for _, child := range cell.Children {
		if child.Type == TextBox {
			fontSize := 16.0
			textWidth := MeasureText(child.Text, fontSize)
			child.Rect.X = startX
			child.Rect.Y = yOffset
			child.Rect.Width = textWidth
			child.Rect.Height = 24.0
			yOffset += 24.0
		}
	}

	return yOffset - startY
}

// getImageSize reads width/height attributes or returns defaults
func getImageSize(node *dom.Node) (float64, float64) {
	if node == nil {
		return DefaultImageWidth, DefaultImageHeight
	}

	width := DefaultImageWidth
	height := DefaultImageHeight

	if w, ok := node.Attributes["width"]; ok {
		if parsed, err := strconv.ParseFloat(strings.TrimSuffix(w, "px"), 64); err == nil {
			width = parsed
		}
	}

	if h, ok := node.Attributes["height"]; ok {
		if parsed, err := strconv.ParseFloat(strings.TrimSuffix(h, "px"), 64); err == nil {
			height = parsed
		}
	}

	return width, height
}

func (box *LayoutBox) Print(indent int) {
	prefix := strings.Repeat("  ", indent)

	typeName := "Block"
	switch box.Type {
	case InlineBox:
		typeName = "Inline"
	case TextBox:
		typeName = "Text"
	case ImageBox:
		typeName = "Image"
	}

	if box.Type == TextBox {
		fmt.Printf("%s[%s] \"%s\" (%.0f,%.0f) %.0fx%.0f\n",
			prefix, typeName, truncate(box.Text, 20),
			box.Rect.X, box.Rect.Y,
			box.Rect.Width, box.Rect.Height)
	} else {
		tag := ""
		if box.Node != nil && box.Node.TagName != "" {
			tag = "<" + box.Node.TagName + "> "
		}
		fmt.Printf("%s[%s] %s(%.0f,%.0f) %.0fx%.0f\n",
			prefix, typeName, tag,
			box.Rect.X, box.Rect.Y,
			box.Rect.Width, box.Rect.Height)
	}

	for _, child := range box.Children {
		child.Print(indent + 1)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func isInsidePre(box *LayoutBox) bool {
	for p := box.Parent; p != nil; p = p.Parent {
		if p.Node != nil && p.Node.TagName == dom.TagPre {
			return true
		}
	}
	return false
}

// getButtonText extracts text content from a button element
func getButtonText(box *LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == TextBox {
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

// GetLegendText extracts text content from a legend element
func GetLegendText(box *LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == TextBox {
			return child.Text
		}
	}
	return ""
}

func getLineHeightFromStyle(style css.Style, tagName string) float64 {
	if style.LineHeight > 0 {
		return style.LineHeight
	}
	return getDefaultLineHeight(tagName)
}
