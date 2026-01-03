package dom

import "testing"

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty input
		{"empty string", "", ""},

		// All whitespace with newlines (HTML formatting between blocks)
		{"newlines only", "\n\n", ""},
		{"newlines with spaces", "\n    \n", ""},
		{"tabs and newlines", "\t\n\t", ""},

		// All whitespace without newlines (inline separator)
		{"single space", " ", " "},
		{"multiple spaces", "   ", " "},
		{"tabs only", "\t\t", " "},

		// Text with leading space (inline element after text)
		{"leading space", " hello", " hello"},
		{"leading tab", "\thello", " hello"},

		// Text with trailing space (text before inline element)
		{"trailing space", "hello ", "hello "},
		{"trailing tab", "hello\t", "hello "},

		// Text with both leading and trailing space
		{"both spaces", " hello ", " hello "},

		// Text with internal whitespace (should collapse)
		{"internal spaces", "hello    world", "hello world"},
		{"internal newline", "hello\nworld", "hello world"},
		{"internal mixed", "hello  \n  world", "hello world"},

		// Text starting/ending with newline (no boundary space)
		{"starts with newline", "\nhello", "hello"},
		{"ends with newline", "hello\n", "hello"},
		{"both newlines", "\nhello\n", "hello"},

		// Real-world HTML scenarios
		{"before inline tag", "This is ", "This is "},
		{"after inline tag", " using strong tag.", " using strong tag."},
		{"between block tags", "\n    \n    ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
