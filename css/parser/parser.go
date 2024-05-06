package parser

import (
	"fmt"

	"github.com/benoitkugler/webrender/utils"
)

// Compound is a compound CSS chunk, like a declaration
// or a qualified rule.
type Compound interface {
	Pos() Pos
	isCompound()
}

type QualifiedRule struct {
	Prelude, Content []Token
	pos              Pos
}

type AtRule struct {
	AtKeyword string
	QualifiedRule
}

type Declaration struct {
	Name      string
	Value     []Token
	pos       Pos
	Important bool
}

func (QualifiedRule) isCompound() {}
func (AtRule) isCompound()        {}
func (Declaration) isCompound()   {}
func (ParseError) isCompound()    {}
func (Whitespace) isCompound()    {}
func (Comment) isCompound()       {}

func (t QualifiedRule) Pos() Pos { return t.pos }
func (t AtRule) Pos() Pos        { return t.pos }
func (t Declaration) Pos() Pos   { return t.pos }

// Already defined:
// 	ParseError
// 	Whitespace
// 	Comment

// Parse a single `declaration`, returning a [ParseError] or a [Declaration]
//
// This is used e.g. for a declaration in an `@supports
// <http://drafts.csswg.org/csswg/css-conditional/#at-supports>.
// Any whitespace or comment before the “:“ colon is dropped.
func ParseOneDeclaration(input []Token) Compound {
	tokens := NewIter(input)
	firstToken := tokens.NextSignificant()
	if firstToken == nil {
		return ParseError{pos: Pos{1, 1}, kind: errEmpty, Message: "Input is empty"}
	}
	return parseDeclaration(firstToken, tokens)
}

// parses a declaration, by consuming `tokens`
// until the end of the declaration or the first error.
// returns either a [ParseError] or a [Declaration]
func parseDeclaration(firstToken Token, tokens *TokensIter) Compound {
	name, ok := firstToken.(Ident)
	if !ok {
		return ParseError{
			pos:     firstToken.Pos(),
			kind:    errInvalid,
			Message: fmt.Sprintf("Expected <ident> for declaration name, got %s.", firstToken.Kind()),
		}
	}
	colon := tokens.NextSignificant()
	if colon == nil {
		return ParseError{
			pos:     firstToken.Pos(),
			kind:    errInvalid,
			Message: "Expected ':' after declaration name, got EOF",
		}
	}

	if lit, ok := colon.(Literal); !ok || lit.Value != ":" {
		return ParseError{
			pos:     colon.Pos(),
			kind:    errInvalid,
			Message: fmt.Sprintf("Expected ':' after declaration name, got %s.", colon.Kind()),
		}
	}

	const (
		_ = iota
		sValue
		sImportant
		sBang
	)
	var (
		value           []Token
		state           = sValue
		bangPosition, i = 0, -1
	)
	for tokens.HasNext() {
		i += 1
		token := tokens.Next()
		switch token := token.(type) {
		case Literal:
			if state == sValue && token.Value == "!" {
				state = sBang
				bangPosition = i
			} else {
				state = sValue
			}
		case Ident:
			if state == sBang && utils.AsciiLower(token.Value) == "important" {
				state = sImportant
			}
		default:
			if token.Kind() != KWhitespace && token.Kind() != KComment {
				state = sValue
			}
		}
		value = append(value, token)
	}

	if state == sImportant {
		value = value[:bangPosition]
	}

	return Declaration{
		pos:       name.pos,
		Name:      name.Value,
		Value:     value,
		Important: state == sImportant,
	}
}

// Like `parseDeclaration`, but stop at the first “;“.
func consumeDeclarationInList(firstToken Token, tokens *TokensIter) Compound {
	var otherDeclarationTokens []Token
	for tokens.HasNext() {
		token := tokens.Next()
		if lit, ok := token.(Literal); ok && lit.Value == ";" {
			break
		}
		otherDeclarationTokens = append(otherDeclarationTokens, token)
	}
	return parseDeclaration(firstToken, &TokensIter{otherDeclarationTokens, 0})
}

// ParseDeclarationListString tokenizes `css` and calls `ParseDeclarationList`.
func ParseDeclarationListString(css string, skipComments, skipWhitespace bool) []Compound {
	l := Tokenize([]byte(css), skipComments)
	return ParseDeclarationList(l, skipComments, skipWhitespace)
}

// Parse a `declaration list` (which may also contain at-rules).
// This is used e.g. for the `QualifiedRule.content`
// of a style rule or “@page“ rule, or for the “style“ attribute of an HTML element.
//
// In contexts that don’t expect any at-rule, all `AtRule` objects should simply be rejected as invalid.
//
// If `skipComments`, ignore CSS comments at the top-level of the list.
// If `skipWhitespace`, ignore whitespace at the top-level of the list. Whitespace is still preserved in
// the `Declaration.value` of declarations and the `AtRule.prelude` and `AtRule.content` of at-rules.
func ParseDeclarationList(input []Token, skipComments, skipWhitespace bool) []Compound {
	tokens := NewIter(input)
	var result []Compound

	for tokens.HasNext() {
		token := tokens.Next()
		switch token := token.(type) {
		case Whitespace:
			if !skipWhitespace {
				result = append(result, token)
			}
		case Comment:
			if !skipComments {
				result = append(result, token)
			}
		case AtKeyword:
			val := consumeAtRule(token, tokens)
			result = append(result, val)
		case Literal:
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

// Parse an at-rule, by consuming just enough of `tokens` for this rule.
// [atKeyword] is the token starting this rule.
func consumeAtRule(atKeyword AtKeyword, tokens *TokensIter) AtRule {
	var (
		prelude []Token
		content []Token
	)
	for tokens.HasNext() {
		token := tokens.Next()
		if curly, ok := token.(CurlyBracketsBlock); ok {
			content = curly.Arguments
			if content == nil {
				content = []Token{}
			}
			break
		}
		lit, ok := token.(Literal)
		if ok && lit.Value == ";" {
			break
		}
		prelude = append(prelude, token)
	}
	return AtRule{
		AtKeyword: atKeyword.Value,
		QualifiedRule: QualifiedRule{
			pos:     atKeyword.pos,
			Prelude: prelude,
			Content: content,
		},
	}
}

// Parse a qualified rule or at-rule, by
// consuming just enough of `tokens` for this rule.
func consumeRule(firstToken Token, tokens *TokensIter) Compound {
	var (
		prelude []Token
		block   CurlyBracketsBlock
	)
	switch firstToken := firstToken.(type) {
	case AtKeyword:
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
				pos:     prelude[len(prelude)-1].Pos(),
				kind:    errInvalid,
				Message: "EOF reached before {} block for a qualified rule.",
			}
		}
	}
	return QualifiedRule{
		pos:     firstToken.Pos(),
		Content: block.Arguments,
		Prelude: prelude,
	}
}

// Parse a non-top-level `rule list`.
//
// This is used for parsing the `AtRule.content` of nested rules like “@media“.
// This differs from `ParseStylesheet` in that top-level “<!--“ and “-->“ tokens are not ignored.
//
// If [skipComments] is true, ignores CSS comments at the top-level of the list.
//
// If [skipWhitespace] is true, ignores whitespace at the top-level of the list.
// Whitespace are still preserved in the `QualifiedRule.Prelude` and the `QualifiedRule.Content` of rules.
func ParseRuleList(input []Token, skipComments, skipWhitespace bool) []Compound {
	tokens := NewIter(input)
	var result []Compound
	for tokens.HasNext() {
		token := tokens.Next()
		switch token := token.(type) {
		case Whitespace:
			if !skipWhitespace {
				result = append(result, token)
			}
		case Comment:
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

// Parse a stylesheet from tokens.
//
// This is used e.g. for a “<style>“ HTML element.
// This differs from `ParseRuleList` in that top-level “<!--“ && “-->“ tokens are ignored.
// This is a legacy quirk for the “<style>“ HTML element.
func ParseStylesheet(input []Token, skipComments, skipWhitespace bool) []Compound {
	iter := NewIter(input)
	var result []Compound
	for iter.HasNext() {
		token := iter.Next()
		switch token := token.(type) {
		case Whitespace:
			if !skipWhitespace {
				result = append(result, token)
			}
		case Comment:
			if !skipComments {
				result = append(result, token)
			}
		case Literal:
			if token.Value != "<!--" && token.Value != "-->" {
				result = append(result, consumeRule(token, iter))
			}
		default:
			result = append(result, consumeRule(token, iter))
		}
	}
	return result
}

// ParseStylesheetBytes tokenizes `input` and calls `ParseStylesheet`.
func ParseStylesheetBytes(input []byte, skipComments, skipWhitespace bool) []Compound {
	l := Tokenize(input, skipComments)
	return ParseStylesheet(l, skipComments, skipWhitespace)
}
