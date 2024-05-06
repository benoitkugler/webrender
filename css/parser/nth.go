package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/benoitkugler/webrender/utils"
)

var nDashDigitRe = regexp.MustCompile("^n(-[0-9]+)$")

// Parse <An+B> (see <http://drafts.csswg.org/csswg/css-syntax-3/#anb>),
// as found in `:nth-child()` and related selector pseudo-classes.
//
// Returns  [a, b] or nil
func ParseNth(input []Token) *[2]int {
	tokens := NewIter(input)
	token_ := tokens.NextSignificant()
	if token_ == nil {
		return nil
	}
	switch token := token_.(type) {
	case Number:
		if token.IsInt() {
			return parseEnd(tokens, 0, token.Int())
		}
	case Dimension:
		if token.IsInt() {
			unit := utils.AsciiLower(token.Unit)
			if unit == "n" {
				return parseB(tokens, token.Int())
			} else if unit == "n-" {
				return parseSignlessB(tokens, token.Int(), -1)
			} else {
				if match, b := matchInt(unit); match {
					return parseEnd(tokens, token.Int(), b)
				}
			}
		}
	case Ident:
		ident := utils.AsciiLower(token.Value)
		if ident == "even" {
			return parseEnd(tokens, 2, 0)
		} else if ident == "odd" {
			return parseEnd(tokens, 2, 1)
		} else if ident == "n" {
			return parseB(tokens, 1)
		} else if ident == "-n" {
			return parseB(tokens, -1)
		} else if ident == "n-" {
			return parseSignlessB(tokens, 1, -1)
		} else if ident == "-n-" {
			return parseSignlessB(tokens, -1, -1)
		} else if ident[0] == '-' {
			if match, b := matchInt(ident[1:]); match {
				return parseEnd(tokens, -1, b)
			}
		} else {
			if match, b := matchInt(ident); match {
				return parseEnd(tokens, 1, b)
			}
		}
	case Literal:
		if token.Value == "+" {
			token_ = tokens.Next() // Whitespace after an initial "+" is invalid.
			if identToken, ok := token_.(Ident); ok {
				ident := utils.AsciiLower(identToken.Value)
				if ident == "n" {
					return parseB(tokens, 1)
				} else if ident == "n-" {
					return parseSignlessB(tokens, 1, -1)
				} else {
					if match, b := matchInt(ident); match {
						return parseEnd(tokens, 1, b)
					}
				}
			}
		}
	}
	return nil
}

func matchInt(s string) (bool, int) {
	match := nDashDigitRe.FindStringSubmatch(s)
	if len(match) > 0 {
		if out, err := strconv.Atoi(match[1]); err == nil {
			return true, out
		}
	}
	return false, 0
}

func parseB(tokens *TokensIter, a int) *[2]int {
	token := tokens.NextSignificant()
	if token == nil {
		return &[2]int{a, 0}
	}
	lit, ok := token.(Literal)
	if ok && lit.Value == "+" {
		return parseSignlessB(tokens, a, 1)
	} else if ok && lit.Value == "-" {
		return parseSignlessB(tokens, a, -1)
	}
	if number, ok := token.(Number); ok && number.IsInt() && strings.Contains("-+", number.Value[0:1]) {
		return parseEnd(tokens, a, number.Int())
	}
	return nil
}

func parseSignlessB(tokens *TokensIter, a, bSign int) *[2]int {
	token := tokens.NextSignificant()
	if number, ok := token.(Number); ok && number.IsInt() && !strings.Contains("-+", number.Value[0:1]) {
		return parseEnd(tokens, a, bSign*number.Int())
	}
	return nil
}

func parseEnd(tokens *TokensIter, a, b int) *[2]int {
	if tokens.NextSignificant() == nil {
		return &[2]int{a, b}
	}
	return nil
}
