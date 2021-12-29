package parser

import (
	"regexp"
	"strconv"
	"strings"
)

var nDashDigitRe = regexp.MustCompile("^n(-[0-9]+)$")

func ParseNth2(css string) []int {
	l := Tokenize(css, true)
	return ParseNth(l)
}

// Parse `<An+B> <http://dev.w3.org/csswg/css-syntax-3/#anb>`_,
//     as found in `:nth-child()
//     <http://dev.w3.org/csswg/selectors/#nth-child-pseudo>`
//     and related Selector pseudo-classes.
//     Although tinycss2 does not include a full Selector parser,
//     this bit of syntax is included as it is particularly tricky to define
//     on top of a CSS tokenizer.
//     Returns  ``(a, b)`` slice of integers or nil
func ParseNth(input []Token) []int {
	tokens := NewTokenIterator(input)
	token_ := nextSignificant(tokens)
	if token_ == nil {
		return nil
	}
	switch token := token_.(type) {
	case NumberToken:
		if token.IsInteger {
			return parseEnd(tokens, 0, token.IntValue())
		}
	case DimensionToken:
		if token.IsInteger {
			unit := token.Unit.Lower()
			if unit == "n" {
				return parseB(tokens, token.IntValue())
			} else if unit == "n-" {
				return parseSignlessB(tokens, token.IntValue(), -1)
			} else {
				if match, b := matchInt(unit); match {
					return parseEnd(tokens, token.IntValue(), b)
				}
			}
		}
	case IdentToken:
		ident := token.Value.Lower()
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
	case LiteralToken:
		if token.Value == "+" {
			token_ = tokens.Next() // Whitespace after an initial "+" is invalid.
			if identToken, ok := token_.(IdentToken); ok {
				ident := identToken.Value.Lower()
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

func parseB(tokens *tokenIterator, a int) []int {
	token := nextSignificant(tokens)
	if token == nil {
		return []int{a, 0}
	}
	lit, ok := token.(LiteralToken)
	if ok && lit.Value == "+" {
		return parseSignlessB(tokens, a, 1)
	} else if ok && lit.Value == "-" {
		return parseSignlessB(tokens, a, -1)
	}
	if number, ok := token.(NumberToken); ok && number.IsInteger && strings.Contains("-+", number.Representation[0:1]) {
		return parseEnd(tokens, a, number.IntValue())
	}
	return nil
}

func parseSignlessB(tokens *tokenIterator, a, bSign int) []int {
	token := nextSignificant(tokens)
	if number, ok := token.(NumberToken); ok && number.IsInteger && !strings.Contains("-+", number.Representation[0:1]) {
		return parseEnd(tokens, a, bSign*number.IntValue())
	}
	return nil
}

func parseEnd(tokens *tokenIterator, a, b int) []int {
	if nextSignificant(tokens) == nil {
		return []int{a, b}
	}
	return nil
}
