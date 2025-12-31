package css

import (
	"strings"
	"unicode"
)

type Parser struct {
	input string
	pos   int
}

func Parse(input string) Stylesheet {
	p := &Parser{input: input, pos: 0}
	return p.parseStylesheet()
}

func (p *Parser) parseStylesheet() Stylesheet {
	var rules []Rule
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		rule := p.parseRule()
		rules = append(rules, rule)
	}
	return Stylesheet{Rules: rules}
}

func (p *Parser) parseRule() Rule {
	selectors := p.parseSelectors()
	declarations := p.parseDeclarations()
	return Rule{Selectors: selectors, Declarations: declarations}
}

func (p *Parser) parseSelectors() []Selector {
	var selectors []Selector
	for {
		sel := p.parseSelector()
		if sel.TagName != "" || sel.ID != "" || len(sel.Classes) > 0 {
			selectors = append(selectors, sel)
		}
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '{' {
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++ // skip comma
		} else {
			// Skip unknown characters (like :link, :hover, etc.)
			p.pos++
		}
	}
	return selectors
}

func (p *Parser) parseSelector() Selector {
	p.skipWhitespace()
	sel := Selector{}

	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '#' {
			p.pos++
			sel.ID = p.parseIdentifier()
		} else if c == '.' {
			p.pos++
			sel.Classes = append(sel.Classes, p.parseIdentifier())
		} else if isIdentChar(rune(c)) {
			sel.TagName = p.parseIdentifier()
		} else {
			break
		}
	}
	return sel
}

func (p *Parser) parseDeclarations() []Declaration {
	var decls []Declaration
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		p.pos++ // skip {
	}

	for p.pos < len(p.input) && p.input[p.pos] != '}' {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '}' {
			break
		}
		decl := p.parseDeclaration()
		if decl.Property != "" {
			decls = append(decls, decl)
		}
	}

	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++ // skip }
	}
	return decls
}

func (p *Parser) parseDeclaration() Declaration {
	p.skipWhitespace()
	property := p.parseIdentifier()
	p.skipWhitespace()

	if p.pos < len(p.input) && p.input[p.pos] == ':' {
		p.pos++ // skip :
	}

	p.skipWhitespace()
	value := p.parseValue()

	if p.pos < len(p.input) && p.input[p.pos] == ';' {
		p.pos++ // skip ;
	}

	return Declaration{Property: property, Value: value}
}

func (p *Parser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) && isIdentChar(rune(p.input[p.pos])) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *Parser) parseValue() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ';' || c == '}' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func isIdentChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '-' || c == '_'
}
