package layout

import (
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

// Heights for different elements
func getLineHeight(tagName string) float64 {
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
	default:
		return 16.0
	}
}

func ComputeLayout(root *LayoutBox, containerWidth float64) {
	computeBlockLayout(root, containerWidth, 0, 0, "")
}

func computeBlockLayout(box *LayoutBox, containerWidth float64, startX, startY float64, parentTag string) {
	box.Rect.X = startX
	box.Rect.Y = startY
	box.Rect.Width = containerWidth

	innerX := startX
	innerWidth := containerWidth

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

	// Default margins for block elements
	switch currentTag {
	case dom.TagP:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
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
	if box.Style.MarginLeft > 0 {
		innerX += box.Style.MarginLeft
		innerWidth -= box.Style.MarginLeft
	}
	if box.Style.MarginRight > 0 {
		innerWidth -= box.Style.MarginRight
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

	yOffset := startY + box.Margin.Top + box.Padding.Top

	// Line state for inline flow
	currentX := innerX
	lineStartY := yOffset
	lineHeight := 0.0
	var lineBoxes []*LayoutBox

	for _, child := range box.Children {
		var childWidth, childHeight float64

		switch child.Type {
		case TextBox:
			fontSize := getFontSize(parentTag)
			childWidth = MeasureText(child.Text, fontSize)
			childHeight = getLineHeight(parentTag)

		case InlineBox:
			// Compute inline box size from its content
			childWidth, childHeight = computeInlineSize(child, parentTag)

		case ImageBox:
			childWidth, childHeight = getImageSize(child.Node)

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
				yOffset += getLineHeight(parentTag)
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

	box.Rect.Height = yOffset - startY + box.Margin.Bottom + box.Padding.Bottom
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

	for _, child := range box.Children {
		var w, h float64
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(parentTag)
			w = MeasureText(child.Text, fontSize)
			h = getLineHeight(parentTag)
		case InlineBox:
			w, h = computeInlineSize(child, parentTag)
		case ImageBox:
			w, h = getImageSize(child.Node)
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
	offsetX := 0.0
	for _, child := range box.Children {
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(parentTag)
			w := MeasureText(child.Text, fontSize)
			h := getLineHeight(parentTag)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
			child.Rect.Width = w
			child.Rect.Height = h
			offsetX += w
		case InlineBox:
			w, h := computeInlineSize(child, parentTag)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
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
