package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInline(t *testing.T) {
	tests := []struct {
		name     string
		boxType  BoxType
		expected bool
	}{
		// Inline types (return true)
		{"TextBox is inline", TextBox, true},
		{"InlineBox is inline", InlineBox, true},
		{"ImageBox is inline", ImageBox, true},

		// Block types (return false)
		{"BlockBox is not inline", BlockBox, false},
		{"TableBox is not inline", TableBox, false},
		{"TableRowBox is not inline", TableRowBox, false},
		{"TableCellBox is not inline", TableCellBox, false},
		{"HRBox is not inline", HRBox, false},
		{"BRBox is not inline", BRBox, false},

		// Form element types (return false)
		{"InputBox is not inline", InputBox, false},
		{"ButtonBox is not inline", ButtonBox, false},
		{"TextareaBox is not inline", TextareaBox, false},
		{"SelectBox is not inline", SelectBox, false},
		{"RadioBox is not inline", RadioBox, false},
		{"CheckboxBox is not inline", CheckboxBox, false},
		{"FileInputBox is not inline", FileInputBox, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := &LayoutBox{Type: tt.boxType}
			result := box.IsInline()
			assert.Equal(t, tt.expected, result, "LayoutBox{Type: %v}.IsInline()", tt.boxType)
		})
	}
}
