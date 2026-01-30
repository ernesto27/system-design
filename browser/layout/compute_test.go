package layout

import (
	"browser/css"
	"browser/dom"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLineHeight(t *testing.T) {
	tests := []struct {
		tag      string
		expected float64
	}{
		{"h1", 40.0},
		{"h2", 32.0},
		{"h3", 26.0},
		{"h4", 24.0},
		{"h5", 22.0},
		{"h6", 20.0},
		{"small", 18.0},
		{"p", 24.0},
		{"div", 24.0},
		{"span", 24.0},
		{"", 24.0},
		{"unknown", 24.0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := getDefaultLineHeight(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLineHeightFromStyle(t *testing.T) {
	tests := []struct {
		name        string
		lineHeight  float64
		tagName     string
		expected    float64
	}{
		{"style has line-height", 32.0, "p", 32.0},
		{"style has line-height overrides tag default", 50.0, "h1", 50.0},
		{"no line-height falls back to h1 default", 0, "h1", 40.0},
		{"no line-height falls back to h2 default", 0, "h2", 32.0},
		{"no line-height falls back to p default", 0, "p", 24.0},
		{"small line-height value", 12.0, "p", 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := css.Style{LineHeight: tt.lineHeight}
			result := getLineHeightFromStyle(style, tt.tagName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFontSize(t *testing.T) {
	tests := []struct {
		tag      string
		expected float64
	}{
		{"h1", 32.0},
		{"h2", 24.0},
		{"h3", 18.0},
		{"h4", 16.0},
		{"h5", 14.0},
		{"h6", 12.0},
		{"small", 12.0},
		{"p", 16.0},
		{"div", 16.0},
		{"span", 16.0},
		{"", 16.0},
		{"unknown", 16.0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := getFontSize(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	tests := []struct {
		name           string
		attrs          map[string]string
		expectedWidth  float64
		expectedHeight float64
	}{
		{"nil node", nil, 200.0, 150.0},
		{"no attributes", map[string]string{}, 200.0, 150.0},
		{"width only", map[string]string{"width": "300"}, 300.0, 150.0},
		{"height only", map[string]string{"height": "200"}, 200.0, 200.0},
		{"both attributes", map[string]string{"width": "400", "height": "300"}, 400.0, 300.0},
		{"with px suffix", map[string]string{"width": "250px", "height": "180px"}, 250.0, 180.0},
		{"invalid width", map[string]string{"width": "abc"}, 200.0, 150.0},
		{"invalid height", map[string]string{"height": "xyz"}, 200.0, 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node *dom.Node
			if tt.attrs != nil {
				node = dom.NewElement("img", tt.attrs)
			}
			w, h := getImageSize(node)
			assert.Equal(t, tt.expectedWidth, w)
			assert.Equal(t, tt.expectedHeight, h)
		})
	}
}

func TestIsInsidePre(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *LayoutBox
		expected bool
	}{
		{
			name: "direct child of pre",
			setup: func() *LayoutBox {
				pre := createElementBox(BlockBox, "pre")
				text := createTextBox("code")
				addChild(pre, text)
				return text
			},
			expected: true,
		},
		{
			name: "grandchild of pre",
			setup: func() *LayoutBox {
				pre := createElementBox(BlockBox, "pre")
				span := createElementBox(InlineBox, "span")
				text := createTextBox("code")
				addChild(pre, span)
				addChild(span, text)
				return text
			},
			expected: true,
		},
		{
			name: "not inside pre",
			setup: func() *LayoutBox {
				div := createElementBox(BlockBox, "div")
				text := createTextBox("text")
				addChild(div, text)
				return text
			},
			expected: false,
		},
		{
			name: "no parent",
			setup: func() *LayoutBox {
				return createTextBox("orphan")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := tt.setup()
			result := isInsidePre(box)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetButtonText(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *LayoutBox
		expected string
	}{
		{
			name: "text child",
			setup: func() *LayoutBox {
				btn := createElementBox(ButtonBox, "button")
				text := createTextBox("Click Me")
				addChild(btn, text)
				return btn
			},
			expected: "Click Me",
		},
		{
			name: "value attribute",
			setup: func() *LayoutBox {
				attrs := map[string]string{"value": "Submit"}
				return &LayoutBox{
					Type: ButtonBox,
					Node: dom.NewElement("button", attrs),
				}
			},
			expected: "Submit",
		},
		{
			name: "default when no text or value",
			setup: func() *LayoutBox {
				return createElementBox(ButtonBox, "button")
			},
			expected: "Button",
		},
		{
			name: "text child takes priority over value",
			setup: func() *LayoutBox {
				attrs := map[string]string{"value": "Ignored"}
				btn := &LayoutBox{
					Type: ButtonBox,
					Node: dom.NewElement("button", attrs),
				}
				text := createTextBox("Visible")
				addChild(btn, text)
				return btn
			},
			expected: "Visible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := tt.setup()
			result := getButtonText(box)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyLineAlignment(t *testing.T) {
	tests := []struct {
		name       string
		boxes      []*LayoutBox
		innerX     float64
		innerWidth float64
		textAlign  string
		expectedX  []float64
	}{
		{
			name:       "empty boxes",
			boxes:      []*LayoutBox{},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{},
		},
		{
			name: "left alignment no change",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "left",
			expectedX:  []float64{0},
		},
		{
			name: "center alignment",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{25}, // (100-50)/2 = 25
		},
		{
			name: "right alignment",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "right",
			expectedX:  []float64{50}, // 100-50 = 50
		},
		{
			name: "multiple boxes centered",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 30}},
				{Rect: Rect{X: 30, Width: 20}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{25, 55}, // offset = (100-50)/2 = 25
		},
		{
			name: "empty text align",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "",
			expectedX:  []float64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyLineAlignment(tt.boxes, tt.innerX, tt.innerWidth, tt.textAlign)
			for i, box := range tt.boxes {
				assert.Equal(t, tt.expectedX[i], box.Rect.X)
			}
		})
	}
}

func TestComputeLayout(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		containerWidth float64
		verify         func(t *testing.T, tree *LayoutBox)
	}{
		{
			name:           "div inside body offset by margin",
			html:           "<div></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				// div is inside body which has 8px margin
				assert.Equal(t, 8.0, div.Rect.X)
				assert.Equal(t, 8.0, div.Rect.Y)
			},
		},
		{
			name:           "div width reduced by body margin",
			html:           "<div></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				// 800 - 8*2 = 784
				assert.Equal(t, 784.0, div.Rect.Width)
			},
		},
		{
			name:           "body has 8px margin",
			html:           "<body><p>Text</p></body>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				body := findBoxByTag(tree, "body")
				assert.Equal(t, 8.0, body.Margin.Top)
				assert.Equal(t, 8.0, body.Margin.Right)
				assert.Equal(t, 8.0, body.Margin.Bottom)
				assert.Equal(t, 8.0, body.Margin.Left)
			},
		},
		{
			name:           "p has vertical margin",
			html:           "<p>Text</p>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				p := findBoxByTag(tree, "p")
				// Browser default: 1em margin (16px)
				assert.Equal(t, 16.0, p.Margin.Top)
				assert.Equal(t, 16.0, p.Margin.Bottom)
			},
		},
		{
			name:           "h1 has vertical margin",
			html:           "<h1>Title</h1>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				h1 := findBoxByTag(tree, "h1")
				// Browser default: 0.67em margin (16px * 0.67 = 10.72)
				assert.Equal(t, 10.72, h1.Margin.Top)
				assert.Equal(t, 10.72, h1.Margin.Bottom)
			},
		},
		{
			name:           "hr has fixed height",
			html:           "<div><hr></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				hr := findBoxByTag(tree, "hr")
				assert.Equal(t, 2.0, hr.Rect.Height)
			},
		},
		{
			name:           "explicit CSS width respected",
			html:           `<div style="width: 400px"></div>`,
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 400.0, div.Rect.Width)
			},
		},
		{
			name:           "min-width respected",
			html:           `<div style="min-width: 500px"></div>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 500.0, div.Rect.Width)
			},
		},
		{
			name:           "explicit CSS height respected",
			html:           `<div style="height: 100px"></div>`,
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 100.0, div.Rect.Height)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, tt.containerWidth)
			tt.verify(t, tree)
		})
	}
}
