package css

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Stylesheet
	}{
		{
			name:  "empty stylesheet",
			input: "",
			expected: Stylesheet{
				Rules: nil,
			},
		},
		{
			name:  "whitespace only",
			input: "   \n\t  ",
			expected: Stylesheet{
				Rules: nil,
			},
		},
		{
			name:  "single rule with tag selector",
			input: "div { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
		{
			name:  "single rule with id selector",
			input: "#main { font-size: 16px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{ID: "main"}},
						Declarations: []Declaration{{Property: "font-size", Value: "16px"}},
					},
				},
			},
		},
		{
			name:  "single rule with class selector",
			input: ".container { margin: 10px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{Classes: []string{"container"}}},
						Declarations: []Declaration{{Property: "margin", Value: "10px"}},
					},
				},
			},
		},
		{
			name:  "combined selector tag and class",
			input: "div.foo { padding: 5px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div", Classes: []string{"foo"}}},
						Declarations: []Declaration{{Property: "padding", Value: "5px"}},
					},
				},
			},
		},
		{
			name:  "combined selector tag class and id",
			input: "div.foo#bar { color: blue; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div", ID: "bar", Classes: []string{"foo"}}},
						Declarations: []Declaration{{Property: "color", Value: "blue"}},
					},
				},
			},
		},
		{
			name:  "multiple selectors",
			input: "div, p { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{
							{TagName: "div"},
							{TagName: "p"},
						},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
		{
			name:  "multiple declarations",
			input: "div { color: red; font-size: 16px; margin: 10px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{{TagName: "div"}},
						Declarations: []Declaration{
							{Property: "color", Value: "red"},
							{Property: "font-size", Value: "16px"},
							{Property: "margin", Value: "10px"},
						},
					},
				},
			},
		},
		{
			name:  "multiple rules",
			input: "div { color: red; } p { color: blue; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
					{
						Selectors:    []Selector{{TagName: "p"}},
						Declarations: []Declaration{{Property: "color", Value: "blue"}},
					},
				},
			},
		},
		{
			name: "multiline stylesheet",
			input: `
				body {
					background-color: white;
					font-size: 14px;
				}
				h1 {
					color: black;
				}
			`,
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{{TagName: "body"}},
						Declarations: []Declaration{
							{Property: "background-color", Value: "white"},
							{Property: "font-size", Value: "14px"},
						},
					},
					{
						Selectors:    []Selector{{TagName: "h1"}},
						Declarations: []Declaration{{Property: "color", Value: "black"}},
					},
				},
			},
		},
		{
			name:  "multiple classes in selector",
			input: ".foo.bar { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{Classes: []string{"foo", "bar"}}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)

			assert.Len(t, result.Rules, len(tt.expected.Rules), "number of rules")

			for i, rule := range result.Rules {
				if i >= len(tt.expected.Rules) {
					break
				}
				expectedRule := tt.expected.Rules[i]
				assert.Equal(t, expectedRule.Selectors, rule.Selectors, "Rule[%d].Selectors", i)
				assert.Equal(t, expectedRule.Declarations, rule.Declarations, "Rule[%d].Declarations", i)
			}
		})
	}
}
