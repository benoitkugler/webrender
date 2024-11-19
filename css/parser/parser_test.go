package parser

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
)

// Parse a single `qualified rule` or `at-rule`.
// Any whitespace or comment before or after the rule is dropped.
func parseOneRule(input []Token) Compound {
	tokens := NewIter(input)
	first := tokens.NextSignificant()
	if first == nil {
		return ParseError{pos: Pos{1, 1}, kind: errEmpty, Message: "Input is empty"}
	}

	rule := consumeRule(first, tokens)
	next := tokens.NextSignificant()
	if next != nil {
		return ParseError{
			pos: next.Pos(), kind: errExtraInput,
			Message: fmt.Sprintf("Expected a single rule, got %s after the first rule.", next.Kind()),
		}
	}
	return rule
}

// ParseRuleListString tokenizes `css` and calls `ParseRuleListString`.
func ParseRuleListString(css string, skipComments, skipWhitespace bool) []Compound {
	l := tokenizeString(css, skipComments)
	return ParseRuleList(l, skipComments, skipWhitespace)
}

func parseOneDeclarationString(css string, skipComments bool) Compound {
	l := tokenizeString(css, skipComments)
	return ParseOneDeclaration(l)
}

func TestDeclarationList(t *testing.T) {
	inputs, resJson := loadJson(t, "declaration_list.json")
	runTest(t, inputs, resJson, func(s string) []TC {
		return fromC(ParseDeclarationListString(s, true, true))
	})
}

func TestBlocksContents(t *testing.T) {
	inputs, resJson := loadJson(t, "blocks_contents.json")
	runTest(t, inputs, resJson, func(s string) []TC {
		return fromC(ParseBlocksContents(tokenizeString(s, true), true))
	})
}

func TestOneDeclaration(t *testing.T) {
	inputs, resJson := loadJson(t, "one_declaration.json")
	runTestOne(t, inputs, resJson, func(s string) TC {
		return parseOneDeclarationString(s, true).(TC)
	})
}

func TestStylesheet(t *testing.T) {
	inputs, resJson := loadJson(t, "stylesheet.json")
	runTest(t, inputs, resJson, func(s string) []TC {
		return fromC(ParseStylesheetBytes([]byte(s), true, true))
	})
}

func TestRuleList(t *testing.T) {
	inputs, resJson := loadJson(t, "rule_list.json")
	runTest(t, inputs, resJson, func(s string) []TC {
		return fromC(ParseRuleListString(s, true, true))
	})
}

func TestOneRule(t *testing.T) {
	inputs, resJson := loadJson(t, "one_rule.json")
	runTestOne(t, inputs, resJson, func(input string) TC {
		l := tokenizeString(input, true)
		return parseOneRule(l).(TC)
	})
}

func TestColor3(t *testing.T) {
	inputs, resJson := loadJson(t, "color3.json")
	runTestOne(t, inputs, resJson, func(input string) TC {
		return ParseColorString(input)
	})
}

func parseNthString(css string) *[2]int {
	l := tokenizeString(css, true)
	return ParseNth(l)
}

type nth [2]int

func (l *nth) dump() interface{} {
	if l != nil {
		return *l
	}
	return []int(nil)
}

func TestNth(t *testing.T) {
	inputs, resJson := loadJson(t, "An+B.json")
	runTestOne(t, inputs, resJson, func(s string) TC {
		return (*nth)(parseNthString(s))
	})
}

func TestColor3Hsl(t *testing.T) {
	inputs, resJson := loadJson(t, "color3_hsl.json")
	runTestOne(t, inputs, resJson, func(input string) TC {
		return ParseColorString(input)
	})
}

type color3 Color

func (c color3) dump() interface{} {
	if Color(c).IsNone() {
		return []int(nil)
	}
	return []utils.Fl{c.RGBA.R * 255, c.RGBA.G * 255, c.RGBA.B * 255, c.RGBA.A}
}

func TestColor3Keywords(t *testing.T) {
	inputs, resJson := loadJson(t, "color3_keywords.json")

	runTestOne(t, inputs, resJson, func(input string) TC {
		return color3(ParseColorString(input))
	})
}

func TestNilContent(t *testing.T) {
	rule := parseOneRule(tokenizeString("@font-face{}", true)).(AtRule)
	testutils.AssertEqual(t, rule.Content != nil, true)

	rule = parseOneRule(tokenizeString("@font-face", true)).(AtRule)
	testutils.AssertEqual(t, rule.Content == nil, true)
}
