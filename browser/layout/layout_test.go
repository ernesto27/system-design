package layout

import (
	"browser/css"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLayoutTreeBoxTypes(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		findTag     string
		expectedBox BoxType
		shouldExist bool
	}{
		// Block elements
		{"div is BlockBox", "<div></div>", "div", BlockBox, true},
		{"p is BlockBox", "<p></p>", "p", BlockBox, true},
		{"h1 is BlockBox", "<h1></h1>", "h1", BlockBox, true},
		{"h2 is BlockBox", "<h2></h2>", "h2", BlockBox, true},
		{"h3 is BlockBox", "<h3></h3>", "h3", BlockBox, true},
		{"h4 is BlockBox", "<h4></h4>", "h4", BlockBox, true},
		{"h5 is BlockBox", "<h5></h5>", "h5", BlockBox, true},
		{"h6 is BlockBox", "<h6></h6>", "h6", BlockBox, true},
		{"ul is BlockBox", "<ul></ul>", "ul", BlockBox, true},
		{"ol is BlockBox", "<ol></ol>", "ol", BlockBox, true},
		{"li is BlockBox", "<ul><li></li></ul>", "li", BlockBox, true},
		{"header is BlockBox", "<header></header>", "header", BlockBox, true},
		{"footer is BlockBox", "<footer></footer>", "footer", BlockBox, true},
		{"blockquote is BlockBox", "<blockquote></blockquote>", "blockquote", BlockBox, true},
		{"pre is BlockBox", "<pre></pre>", "pre", BlockBox, true},
		{"form is BlockBox", "<form></form>", "form", BlockBox, true},

		// Inline elements
		{"span is InlineBox", "<div><span></span></div>", "span", InlineBox, true},
		{"a is InlineBox", "<div><a></a></div>", "a", InlineBox, true},
		{"strong is InlineBox", "<div><strong></strong></div>", "strong", InlineBox, true},
		{"em is InlineBox", "<div><em></em></div>", "em", InlineBox, true},
		{"b is InlineBox", "<div><b></b></div>", "b", InlineBox, true},
		{"i is InlineBox", "<div><i></i></div>", "i", InlineBox, true},
		{"small is InlineBox", "<div><small></small></div>", "small", InlineBox, true},
		{"u is InlineBox", "<div><u></u></div>", "u", InlineBox, true},

		// Special elements
		{"img is ImageBox", "<div><img src=\"test.png\"></div>", "img", ImageBox, true},
		{"hr is HRBox", "<div><hr></div>", "hr", HRBox, true},
		{"br is BRBox", "<p>a<br>b</p>", "br", BRBox, true},

		// Table elements
		{"table is TableBox", "<table></table>", "table", TableBox, true},
		{"tbody is TableBox", "<table><tbody></tbody></table>", "tbody", TableBox, true},
		{"thead is TableBox", "<table><thead></thead></table>", "thead", TableBox, true},
		{"tfoot is TableBox", "<table><tfoot></tfoot></table>", "tfoot", TableBox, true},
		{"tr is TableRowBox", "<table><tr></tr></table>", "tr", TableRowBox, true},
		{"td is TableCellBox", "<table><tr><td></td></tr></table>", "td", TableCellBox, true},
		{"th is TableCellBox", "<table><tr><th></th></tr></table>", "th", TableCellBox, true},

		// Form elements
		{"button is ButtonBox", "<form><button></button></form>", "button", ButtonBox, true},
		{"textarea is TextareaBox", "<form><textarea></textarea></form>", "textarea", TextareaBox, true},
		{"select is SelectBox", "<form><select></select></form>", "select", SelectBox, true},

		// Input types
		{"input text is InputBox", "<form><input type=\"text\"></form>", "input", InputBox, true},
		{"input password is InputBox", "<form><input type=\"password\"></form>", "input", InputBox, true},
		{"input email is InputBox", "<form><input type=\"email\"></form>", "input", InputBox, true},
		{"input default is InputBox", "<form><input></form>", "input", InputBox, true},
		{"input radio is RadioBox", "<form><input type=\"radio\"></form>", "input", RadioBox, true},
		{"input checkbox is CheckboxBox", "<form><input type=\"checkbox\"></form>", "input", CheckboxBox, true},
		{"input file is FileInputBox", "<form><input type=\"file\"></form>", "input", FileInputBox, true},
		{"input hidden returns nil", "<form><input type=\"hidden\"></form>", "input", InputBox, false},

		// Skip elements
		{"script is skipped", "<div><script></script></div>", "script", BlockBox, false},
		{"style is skipped", "<div><style></style></div>", "style", BlockBox, false},
		{"head is skipped", "<html><head></head></html>", "head", BlockBox, false},
		{"meta is skipped", "<head><meta></head>", "meta", BlockBox, false},
		{"link is skipped", "<head><link></head>", "link", BlockBox, false},
		{"option is skipped", "<select><option></option></select>", "option", BlockBox, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			box := findBoxByTag(tree, tt.findTag)
			if tt.shouldExist {
				assert.NotNil(t, box)
				assert.Equal(t, tt.expectedBox, box.Type)
			} else {
				assert.Nil(t, box)
			}
		})
	}
}

func TestBuildLayoutTreeDisplayNone(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		css     string
		findTag string
	}{
		{"inline style display none", `<div><p style="display: none">Hidden</p></div>`, "", "p"},
		{"stylesheet display none", `<div><p class="hidden">Hidden</p></div>`, `.hidden { display: none; }`, "p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *LayoutBox
			if tt.css == "" {
				tree = buildTree(tt.html)
			} else {
				tree = buildTreeWithCSS(tt.html, tt.css)
			}
			box := findBoxByTag(tree, tt.findTag)
			assert.Nil(t, box)
		})
	}
}

func TestBuildLayoutTreeStructure(t *testing.T) {
	t.Run("parent and children linked correctly", func(t *testing.T) {
		html := `<div><p>A</p><p>B</p></div>`
		tree := buildTree(html)
		divBox := findBoxByTag(tree, "div")

		assert.NotNil(t, divBox)
		pCount := 0
		for _, child := range divBox.Children {
			if child.Node != nil && child.Node.TagName == "p" {
				pCount++
				assert.Same(t, divBox, child.Parent)
			}
		}
		assert.Equal(t, 2, pCount)
	})

	t.Run("text content preserved", func(t *testing.T) {
		html := `<p>Hello World</p>`
		tree := buildTree(html)
		textBox := findBoxByType(tree, TextBox)
		assert.NotNil(t, textBox)
		assert.Equal(t, "Hello World", textBox.Text)
	})
}

func TestBuildLayoutTreeStyles(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		css      string
		findTag  string
		expected float64
	}{
		{"stylesheet applied", "<p>Text</p>", "p { font-size: 20px; }", "p", 20.0},
		{"inline overrides stylesheet", `<p style="font-size: 30px">Text</p>`, "p { font-size: 20px; }", "p", 30.0},
		{"id selector matched", `<p id="main">Text</p>`, "#main { font-size: 24px; }", "p", 24.0},
		{"class selector matched", `<p class="big">Text</p>`, ".big { font-size: 18px; }", "p", 18.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTreeWithCSS(tt.html, tt.css)
			box := findBoxByTag(tree, tt.findTag)
			assert.NotNil(t, box)
			assert.Equal(t, tt.expected, box.Style.FontSize)
		})
	}
}

func TestMergeStyles(t *testing.T) {
	tests := []struct {
		name   string
		base   css.Style
		inline css.Style
		verify func(t *testing.T, result *css.Style)
	}{
		{
			name:   "color merged",
			base:   css.Style{},
			inline: css.Style{Color: color.RGBA{255, 0, 0, 255}},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, color.RGBA{255, 0, 0, 255}, r.Color) },
		},
		{
			name:   "background color merged",
			base:   css.Style{},
			inline: css.Style{BackgroundColor: color.RGBA{0, 255, 0, 255}},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, color.RGBA{0, 255, 0, 255}, r.BackgroundColor) },
		},
		{
			name:   "font size merged when positive",
			base:   css.Style{FontSize: 16},
			inline: css.Style{FontSize: 24},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, 24.0, r.FontSize) },
		},
		{
			name:   "font size zero does not override",
			base:   css.Style{FontSize: 16},
			inline: css.Style{FontSize: 0},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, 16.0, r.FontSize) },
		},
		{
			name:   "bold merged",
			base:   css.Style{Bold: false},
			inline: css.Style{Bold: true},
			verify: func(t *testing.T, r *css.Style) { assert.True(t, r.Bold) },
		},
		{
			name:   "italic merged",
			base:   css.Style{Italic: false},
			inline: css.Style{Italic: true},
			verify: func(t *testing.T, r *css.Style) { assert.True(t, r.Italic) },
		},
		{
			name:   "margins merged",
			base:   css.Style{},
			inline: css.Style{MarginTop: 10, MarginRight: 20, MarginBottom: 30, MarginLeft: 40},
			verify: func(t *testing.T, r *css.Style) {
				assert.Equal(t, 10.0, r.MarginTop)
				assert.Equal(t, 20.0, r.MarginRight)
				assert.Equal(t, 30.0, r.MarginBottom)
				assert.Equal(t, 40.0, r.MarginLeft)
			},
		},
		{
			name:   "paddings merged",
			base:   css.Style{},
			inline: css.Style{PaddingTop: 5, PaddingRight: 10, PaddingBottom: 15, PaddingLeft: 20},
			verify: func(t *testing.T, r *css.Style) {
				assert.Equal(t, 5.0, r.PaddingTop)
				assert.Equal(t, 10.0, r.PaddingRight)
				assert.Equal(t, 15.0, r.PaddingBottom)
				assert.Equal(t, 20.0, r.PaddingLeft)
			},
		},
		{
			name:   "text align merged",
			base:   css.Style{TextAlign: "left"},
			inline: css.Style{TextAlign: "center"},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, "center", r.TextAlign) },
		},
		{
			name:   "empty text align does not override",
			base:   css.Style{TextAlign: "left"},
			inline: css.Style{TextAlign: ""},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, "left", r.TextAlign) },
		},
		{
			name:   "display merged",
			base:   css.Style{Display: "block"},
			inline: css.Style{Display: "none"},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, "none", r.Display) },
		},
		{
			name:   "opacity merged when not 1.0",
			base:   css.Style{Opacity: 1.0},
			inline: css.Style{Opacity: 0.5},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, 0.5, r.Opacity) },
		},
		{
			name:   "opacity 1.0 does not override",
			base:   css.Style{Opacity: 0.5},
			inline: css.Style{Opacity: 1.0},
			verify: func(t *testing.T, r *css.Style) { assert.Equal(t, 0.5, r.Opacity) },
		},
		{
			name:   "border widths merged",
			base:   css.Style{},
			inline: css.Style{BorderTopWidth: 1, BorderRightWidth: 2, BorderBottomWidth: 3, BorderLeftWidth: 4},
			verify: func(t *testing.T, r *css.Style) {
				assert.Equal(t, 1.0, r.BorderTopWidth)
				assert.Equal(t, 2.0, r.BorderRightWidth)
				assert.Equal(t, 3.0, r.BorderBottomWidth)
				assert.Equal(t, 4.0, r.BorderLeftWidth)
			},
		},
		{
			name:   "sizing properties merged",
			base:   css.Style{},
			inline: css.Style{Width: 100, Height: 200, MinWidth: 50},
			verify: func(t *testing.T, r *css.Style) {
				assert.Equal(t, 100.0, r.Width)
				assert.Equal(t, 200.0, r.Height)
				assert.Equal(t, 50.0, r.MinWidth)
			},
		},
		{
			name:   "zero sizing does not override",
			base:   css.Style{Width: 100, Height: 200, MinWidth: 50},
			inline: css.Style{Width: 0, Height: 0, MinWidth: 0},
			verify: func(t *testing.T, r *css.Style) {
				assert.Equal(t, 100.0, r.Width)
				assert.Equal(t, 200.0, r.Height)
				assert.Equal(t, 50.0, r.MinWidth)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.base
			inline := tt.inline
			mergeStyles(&base, &inline)
			tt.verify(t, &base)
		})
	}
}
