// Package parser implements the tokenization of
// a CSS input and the construction of the corresponding Abstract Syntax Tree.
package parser

import (
	"fmt"
)

// ---------------- Parsing ----------------------------------------------------

type TokenIterator struct {
	tokens []Token
	index  int
}

func NewTokenIterator(tokens []Token) *TokenIterator {
	return &TokenIterator{
		tokens: tokens,
	}
}

func (it TokenIterator) HasNext() bool {
	return it.index < len(it.tokens)
}

// Next returns the next token or nil at the end
func (it *TokenIterator) Next() (t Token) {
	if it.HasNext() {
		t = it.tokens[it.index]
		it.index += 1
	}
	return t
}

// Return the next significant (neither whitespace || comment) token.
//     :param tokens: An *iterator* yielding :term:`component values`.
//     :returns: A :term:`component value`, || :obj:`None`.
//
func nextSignificant(tokens *TokenIterator) Token {
	for tokens.HasNext() {
		token := tokens.Next()
		switch token.(type) {
		case WhitespaceToken, Comment:
			continue
		default:
			return token
		}
	}
	return nil
}

// Parse a single `component value`.
// This is used e.g. for an attribute value referred to by ``attr(foo length)``.
func ParseOneComponentValue(input []Token) Token {
	tokens := NewTokenIterator(input)
	first := nextSignificant(tokens)
	second := nextSignificant(tokens)
	if first == nil {
		return ParseError{Pos: newPosition(1, 1), Kind: "empty", Message: "Input is empty"}
	}
	if second != nil {
		return ParseError{Pos: second.Position(), Kind: "extra-input", Message: "Got more than one token"}
	}
	return first
}

// If `skipComments`,  ignore all CSS comments.
//   skipComments = false
func parseOneComponentValueString(css string, skipComments bool) Token {
	l := tokenizeString(css, skipComments)
	return ParseOneComponentValue(l)
}

// Parse a single `declaration`.
// This is used e.g. for a declaration in an `@supports
// <http://dev.w3.org/csswg/css-conditional/#at-supports>`_ test.
// Any whitespace or comment before the ``:`` colon is dropped.
func ParseOneDeclaration(input []Token) Token {
	tokens := NewTokenIterator(input)
	firstToken := nextSignificant(tokens)
	if firstToken == nil {
		return ParseError{Pos: newPosition(1, 1), Kind: "empty", Message: "Input is empty"}
	}
	return parseDeclaration(firstToken, tokens)
}

// If  `skipComments`, ignore all CSS comments.
// skipComments=false
func parseOneDeclarationString(css string, skipComments bool) Token {
	l := tokenizeString(css, skipComments)
	return ParseOneDeclaration(l)
}

// Parse a declaration.
//     Consume :obj:`tokens` until the end of the declaration or the first error.
//     :param firstToken: The first :term:`component value` of the rule.
//     :param tokens: An *iterator* yielding :term:`component values`.
func parseDeclaration(firstToken Token, tokens *TokenIterator) Token {
	name, ok := firstToken.(IdentToken)
	if !ok {
		return ParseError{
			Pos:     name.Pos,
			Kind:    "invalid",
			Message: fmt.Sprintf("Expected <ident> for declaration name, got %s.", firstToken.Type()),
		}
	}
	colon := nextSignificant(tokens)
	if colon == nil {
		return ParseError{
			Pos:     name.Pos,
			Kind:    "invalid",
			Message: "Expected ':' after declaration name, got EOF",
		}
	}

	if lit, ok := colon.(LiteralToken); !ok || lit.Value != ":" {
		return ParseError{
			Pos:     colon.Position(),
			Kind:    "invalid",
			Message: fmt.Sprintf("Expected ':' after declaration name, got %s.", colon.Type()),
		}
	}

	var value []Token
	state := "value"
	bangPosition, i := 0, -1
	for tokens.HasNext() {
		i += 1
		_token := tokens.Next()
		switch token := _token.(type) {
		case LiteralToken:
			if state == "value" && token.Value == "!" {
				state = "bang"
				bangPosition = i
			} else {
				state = "value"
			}
		case IdentToken:
			if state == "bang" && token.Value.Lower() == "important" {
				state = "important"
			}
		default:
			if _token.Type() != "whitespace" && _token.Type() != "comment" {
				state = "value"
			}
		}
		value = append(value, _token)
	}

	if state == "important" {
		value = value[:bangPosition]
	}

	return Declaration{
		Pos:       name.Pos,
		Name:      name.Value,
		Value:     value,
		Important: state == "important",
	}
}

// Like :func:`parseDeclaration`, but stop at the first ``;``.
func consumeDeclarationInList(firstToken Token, tokens *TokenIterator) Token {
	var otherDeclarationTokens []Token
	for tokens.HasNext() {
		token := tokens.Next()
		if lit, ok := token.(LiteralToken); ok && lit.Value == ";" {
			break
		}
		otherDeclarationTokens = append(otherDeclarationTokens, token)
	}
	return parseDeclaration(firstToken, NewTokenIterator(otherDeclarationTokens))
}

// Parse a `declaration list` (which may also contain at-rules).
// This is used e.g. for the `QualifiedRule.content`
// of a style rule or ``@page`` rule, or for the ``style`` attribute of an HTML element.
// In contexts that don’t expect any at-rule, all :class:`AtRule` objects should simply be rejected as invalid.
// If `skipComments`, ignore CSS comments at the top-level of the list. If the input is a string, ignore all comments.
// If `skipWhitespace`, ignore whitespace at the top-level of the list. Whitespace is still preserved in
// the `Declaration.value` of declarations and the `AtRule.prelude` and `AtRule.content` of at-rules.
func ParseDeclarationList(input []Token, skipComments, skipWhitespace bool) []Token {
	tokens := NewTokenIterator(input)
	var result []Token

	for tokens.HasNext() {
		_token := tokens.Next()
		switch token := _token.(type) {
		case WhitespaceToken:
			if !skipWhitespace {
				result = append(result, token)
			}
		case Comment:
			if !skipComments {
				result = append(result, token)
			}
		case AtKeywordToken:
			val := consumeAtRule(token, tokens)
			result = append(result, val)
		case LiteralToken:
			if token.Value != ";" {
				val := consumeDeclarationInList(token, tokens)
				result = append(result, val)
			}
		default:
			val := consumeDeclarationInList(token, tokens)
			result = append(result, val)
		}
	}
	return result
}

// ParseDeclarationListString tokenizes `css` and calls `ParseDeclarationList`.
func ParseDeclarationListString(css string, skipComments, skipWhitespace bool) []Token {
	l := tokenizeString(css, skipComments)
	return ParseDeclarationList(l, skipComments, skipWhitespace)
}

// Parse a single `qualified rule` or `at-rule`.
// Any whitespace or comment before or after the rule is dropped.
func parseOneRule(input []Token) Token {
	tokens := NewTokenIterator(input)
	first := nextSignificant(tokens)
	if first == nil {
		return ParseError{Pos: newPosition(1, 1), Kind: "empty", Message: "Input is empty"}
	}

	rule := consumeRule(first, tokens)
	next := nextSignificant(tokens)
	if next != nil {
		return ParseError{
			Pos: next.Position(), Kind: "extra-input",
			Message: fmt.Sprintf("Expected a single rule, got %s after the first rule.", next.Type()),
		}
	}
	return rule
}

// Parse a non-top-level `rule list`.
// This is used for parsing the `AtRule.content`
// of nested rules like ``@media``.
// This differs from :func:`ParseStylesheet` in that
// top-level ``<!--`` and ``-->`` tokens are not ignored.
// skipComments:
//     Ignore CSS comments at the top-level of the list.
//     If the input is a string, ignore all comments.
// skipWhitespace:
//     Ignore whitespace at the top-level of the list.
//     Whitespace is still preserved
//     in the `QualifiedRule.prelude`
//     and the `QualifiedRule.content` of rules.
func ParseRuleList(input []Token, skipComments, skipWhitespace bool) []Token {
	tokens := NewTokenIterator(input)
	var result []Token
	for tokens.HasNext() {
		token := tokens.Next()
		switch token.Type() {
		case WhitespaceTokenT:
			if !skipWhitespace {
				result = append(result, token)
			}
		case CommentT:
			if !skipComments {
				result = append(result, token)
			}
		default:
			val := consumeRule(token, tokens)
			result = append(result, val)
		}
	}
	return result
}

// ParseRuleListString tokenizes `css` and calls `ParseRuleListString`.
func ParseRuleListString(css string, skipComments, skipWhitespace bool) []Token {
	l := tokenizeString(css, skipComments)
	return ParseRuleList(l, skipComments, skipWhitespace)
}

// Parse a stylesheet from tokens.
// This is used e.g. for a ``<style>`` HTML element.
// This differs from `parseRuleList` in that
// top-level ``<!--`` && ``-->`` tokens are ignored.
// This is a legacy quirk for the ``<style>`` HTML element.
// If `skipComments` is true, ignore CSS comments at the top-level of the stylesheet.
// If the input is a string, ignore all comments.
// If `skipWhitespace` is true, ignore whitespace at the top-level of the stylesheet.
// Whitespace is still preserved  in the `QualifiedRule.Prelude`
// and the `QualifiedRule.Content` of rules.
func ParseStylesheet(input []Token, skipComments, skipWhitespace bool) []Token {
	iter := NewTokenIterator(input)
	var result []Token
	for iter.HasNext() {
		token := iter.Next()
		switch token.Type() {
		case WhitespaceTokenT:
			if !skipWhitespace {
				result = append(result, token)
			}
		case CommentT:
			if !skipComments {
				result = append(result, token)
			}
		case LiteralTokenT:
			if lit, ok := token.(LiteralToken); !ok || (lit.Value != "<!--" && lit.Value != "-->") {
				result = append(result, consumeRule(token, iter))
			}
		default:
			result = append(result, consumeRule(token, iter))
		}
	}
	return result
}

// ParseStylesheetBytes tokenizes `input` and calls `ParseStylesheet`.
func ParseStylesheetBytes(input []byte, skipComments, skipWhitespace bool) []Token {
	l := Tokenize(input, skipComments)
	return ParseStylesheet(l, skipComments, skipWhitespace)
}

// Parse a qualified rule or at-rule.
// Consume just enough of :obj:`tokens` for this rule.
// :param firstToken: The first :term:`component value` of the rule.
// :param tokens: An *iterator* yielding :term:`component values`.
func consumeRule(_firstToken Token, tokens *TokenIterator) Token {
	var (
		prelude []Token
		block   CurlyBracketsBlock
	)
	switch firstToken := _firstToken.(type) {
	case AtKeywordToken:
		return consumeAtRule(firstToken, tokens)
	case CurlyBracketsBlock:
		block = firstToken
	default:
		prelude = []Token{firstToken}
		hasBroken := false
		for tokens.HasNext() {
			token := tokens.Next()
			if curly, ok := token.(CurlyBracketsBlock); ok {
				block = curly
				hasBroken = true
				break
			}
			prelude = append(prelude, token)
		}
		if !hasBroken {
			return ParseError{
				Pos:     prelude[len(prelude)-1].Position(),
				Kind:    "invalid",
				Message: "EOF reached before {} block for a qualified rule.",
			}
		}
	}
	return QualifiedRule{
		Pos:     _firstToken.Position(),
		Content: block.Content,
		Prelude: &prelude,
	}
}

// Parse an at-rule.
// Consume just enough of :obj:`tokens` for this rule.
// :param atKeyword: The :class:`AtKeywordToken` object starting this rule.
// :param tokens: An *iterator* yielding :term:`component values`.
func consumeAtRule(atKeyword AtKeywordToken, tokens *TokenIterator) AtRule {
	var (
		prelude []Token
		content *[]Token
	)
	for tokens.HasNext() {
		token := tokens.Next()
		if curly, ok := token.(CurlyBracketsBlock); ok {
			content = curly.Content
			break
		}
		lit, ok := token.(LiteralToken)
		if ok && lit.Value == ";" {
			break
		}
		prelude = append(prelude, token)
	}
	return AtRule{
		AtKeyword: atKeyword.Value,
		QualifiedRule: QualifiedRule{
			Pos:     atKeyword.Pos,
			Prelude: &prelude,
			Content: content,
		},
	}
}
