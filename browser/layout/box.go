package layout

import (
	"browser/css"
	"browser/dom"
)

type Rect struct {
	X, Y, Width, Height float64
}

type EdgeSizes struct {
	Top, Right, Bottom, Left float64
}

type BoxType int

const (
	BlockBox BoxType = iota
	InlineBox
	TextBox
	ImageBox
	HRBox
	BRBox
	TableBox
	TableRowBox
	TableCellBox
	InputBox
	ButtonBox
	TextareaBox
	SelectBox
	RadioBox
	CheckboxBox
)

type LayoutBox struct {
	Type     BoxType
	Rect     Rect
	Margin   EdgeSizes
	Padding  EdgeSizes
	Children []*LayoutBox
	Node     *dom.Node
	Text     string
	Parent   *LayoutBox
	Style    css.Style
}

// IsInline returns true if the box should flow horizontally (inline)
func (box *LayoutBox) IsInline() bool {
	switch box.Type {
	case TextBox, InlineBox, ImageBox:
		return true
	default:
		return false
	}
}
