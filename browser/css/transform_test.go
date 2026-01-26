package css

import "testing"

func TestApplyTextTransform(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		transform string
		want      string
	}{
		// uppercase
		{"uppercase basic", "hello world", "uppercase", "HELLO WORLD"},
		{"uppercase mixed", "Hello World", "uppercase", "HELLO WORLD"},
		{"uppercase already upper", "HELLO", "uppercase", "HELLO"},
		{"uppercase empty", "", "uppercase", ""},

		// lowercase
		{"lowercase basic", "HELLO WORLD", "lowercase", "hello world"},
		{"lowercase mixed", "Hello World", "lowercase", "hello world"},
		{"lowercase already lower", "hello", "lowercase", "hello"},
		{"lowercase empty", "", "lowercase", ""},

		// capitalize
		{"capitalize basic", "hello world", "capitalize", "Hello World"},
		{"capitalize from upper", "HELLO WORLD", "capitalize", "Hello World"},
		{"capitalize mixed", "hELLO wORLD", "capitalize", "Hello World"},
		{"capitalize single word", "hello", "capitalize", "Hello"},
		{"capitalize empty", "", "capitalize", ""},

		// none/default
		{"none preserves text", "Hello World", "none", "Hello World"},
		{"empty transform preserves text", "Hello World", "", "Hello World"},
		{"unknown transform preserves text", "Hello World", "invalid", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTextTransform(tt.text, tt.transform)
			if got != tt.want {
				t.Errorf("ApplyTextTransform(%q, %q) = %q, want %q", tt.text, tt.transform, got, tt.want)
			}
		})
	}
}

func TestCapitalizeWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basic", "hello world", "Hello World"},
		{"from uppercase", "HELLO WORLD", "Hello World"},
		{"mixed case", "hELLO wORLD", "Hello World"},
		{"single word", "hello", "Hello"},
		{"single char words", "a b c", "A B C"},
		{"empty string", "", ""},
		{"multiple spaces", "hello   world", "Hello World"},
		{"leading spaces", "  hello world", "Hello World"},
		{"trailing spaces", "hello world  ", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CapitalizeWords(tt.input)
			if got != tt.want {
				t.Errorf("CapitalizeWords(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
