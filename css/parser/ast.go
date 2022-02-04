package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/benoitkugler/webrender/utils"
)

const (
	QualifiedRuleT       TokenType = "qualified-rule"
	AtRuleT              TokenType = "at-rule"
	DeclarationT         TokenType = "declaration"
	ParseErrorT          TokenType = "error"
	CommentT             TokenType = "comment"
	WhitespaceTokenT     TokenType = "whitespace"
	LiteralTokenT        TokenType = "literal"
	IdentTokenT          TokenType = "ident"
	AtKeywordTokenT      TokenType = "at-keyword"
	HashTokenT           TokenType = "hash"
	StringTokenT         TokenType = "string"
	URLTokenT            TokenType = "url"
	UnicodeRangeTokenT   TokenType = "unicode-range"
	NumberTokenT         TokenType = "number"
	PercentageTokenT     TokenType = "percentage"
	DimensionTokenT      TokenType = "dimension"
	ParenthesesBlockT    TokenType = "() block"
	SquareBracketsBlockT TokenType = "[] block"
	CurlyBracketsBlockT  TokenType = "{} block"
	FunctionBlockT       TokenType = "function"
)

type TokenType string

type Token interface {
	jsonisable
	Position() Pos
	Type() TokenType
	serializeTo(io.StringWriter)
}

type jsonisable interface {
	toJson() jsonisable // json representation, for tests
}

// LowerableString is a string which can be
// normalized to ASCII lower case
type LowerableString string

func (s LowerableString) Lower() string {
	return utils.AsciiLower(string(s))
}

type Pos struct {
	Line, Column int
}

func newPosition(line, column int) Pos {
	return Pos{Line: line, Column: column}
}

func (n Pos) Position() Pos { return n }

// shared tokens
type stringToken struct {
	Value string
	Pos
}

type bracketsBlock struct {
	Content *[]Token
	Pos
}

type NumericToken struct {
	Representation string
	Pos
	Value     utils.Fl
	IsInteger bool
}

type QualifiedRule struct {
	Prelude, Content *[]Token
	Pos
}

type AtRule struct {
	AtKeyword LowerableString
	QualifiedRule
}

type Declaration struct {
	Name  LowerableString
	Value []Token
	Pos
	Important bool
}

type ParseError struct {
	Kind    string
	Message string
	Pos
}

type (
	Comment         stringToken
	WhitespaceToken stringToken
	LiteralToken    stringToken
	IdentToken      struct {
		Value LowerableString
		Pos
	}
)

type AtKeywordToken struct {
	Value LowerableString
	Pos
}

type HashToken struct {
	Value string
	Pos
	IsIdentifier bool
}

type StringToken struct {
	Value string
	Pos
	isError bool
}

const (
	errorInString = iota + 1
	errorInURL
)

type URLToken struct {
	Value string
	Pos
	isError uint8
}

type (
	UnicodeRangeToken struct {
		Pos
		Start, End uint32
	}
)

type (
	NumberToken     NumericToken
	PercentageToken NumericToken
	DimensionToken  struct {
		Unit LowerableString
		NumericToken
	}
)

func NewNumberToken(v utils.Fl, pos Pos) NumberToken {
	isInt := v == utils.Fl(math.Trunc((float64(v))))
	return NumberToken{Pos: pos, Value: v, IsInteger: isInt, Representation: fmt.Sprintf("%v", v)}
}

type (
	ParenthesesBlock    bracketsBlock
	SquareBracketsBlock bracketsBlock
	CurlyBracketsBlock  bracketsBlock
	FunctionBlock       struct {
		Arguments *[]Token
		Name      LowerableString
		Pos
	}
)

// ----------- boilerplate code for token type -------------------------------------

func (t QualifiedRule) Type() TokenType       { return QualifiedRuleT }
func (t AtRule) Type() TokenType              { return AtRuleT }
func (t Declaration) Type() TokenType         { return DeclarationT }
func (t ParseError) Type() TokenType          { return ParseErrorT }
func (t Comment) Type() TokenType             { return CommentT }
func (t WhitespaceToken) Type() TokenType     { return WhitespaceTokenT }
func (t LiteralToken) Type() TokenType        { return LiteralTokenT }
func (t IdentToken) Type() TokenType          { return IdentTokenT }
func (t AtKeywordToken) Type() TokenType      { return AtKeywordTokenT }
func (t HashToken) Type() TokenType           { return HashTokenT }
func (t StringToken) Type() TokenType         { return StringTokenT }
func (t URLToken) Type() TokenType            { return URLTokenT }
func (t UnicodeRangeToken) Type() TokenType   { return UnicodeRangeTokenT }
func (t NumberToken) Type() TokenType         { return NumberTokenT }
func (t PercentageToken) Type() TokenType     { return PercentageTokenT }
func (t DimensionToken) Type() TokenType      { return DimensionTokenT }
func (t ParenthesesBlock) Type() TokenType    { return ParenthesesBlockT }
func (t SquareBracketsBlock) Type() TokenType { return SquareBracketsBlockT }
func (t CurlyBracketsBlock) Type() TokenType  { return CurlyBracketsBlockT }
func (t FunctionBlock) Type() TokenType       { return FunctionBlockT }

// ---------------------------------- Methods ----------------------------------

// IntValue returns the rounded value
// Should be used only if  `IsInteger` is true
func (t NumericToken) IntValue() int {
	return int(t.Value)
}

func (t NumberToken) IntValue() int {
	return NumericToken(t).IntValue()
}

func (t PercentageToken) IntValue() int {
	return NumericToken(t).IntValue()
}

// ---------------- JSON -------------------------------------------
type (
	myString string
	myFloat  utils.Fl
	myBool   bool
	myInt    int
)

func (s myString) toJson() jsonisable { return s }
func (s myFloat) toJson() jsonisable  { return s }
func (s myBool) toJson() jsonisable   { return s }
func (s myInt) toJson() jsonisable    { return s }

type jsonList []jsonisable

func (s jsonList) toJson() jsonisable {
	for i, v := range s {
		s[i] = v.toJson()
	}
	return s
}

func (n NumericToken) toJson() jsonList {
	l := jsonList{myString(n.Representation), myFloat(n.Value)}
	if n.IsInteger {
		l = append(l, myString("integer"))
	} else {
		l = append(l, myString("number"))
	}
	return l
}

func toJson(l []Token) jsonList {
	out := make(jsonList, len(l))
	for i, t := range l {
		out[i] = t.toJson()
	}
	return out
}

func marshalJSON(l []Token) (string, error) {
	normalize := toJson(l)
	b, err := json.Marshal(normalize)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t QualifiedRule) toJson() jsonisable {
	prelude := toJson(*t.Prelude)
	content := toJson(*t.Content)
	return jsonList{myString("qualified rule"), prelude, content}
}

func (t AtRule) toJson() jsonisable {
	prelude := toJson(*t.Prelude)
	var content jsonisable
	if t.Content != nil {
		content = toJson(*t.Content)
	}
	return jsonList{myString("at-rule"), myString(t.AtKeyword), prelude, content}
}

func (t Declaration) toJson() jsonisable {
	content := toJson(t.Value)
	return jsonList{myString("declaration"), myString(t.Name), content, myBool(t.Important)}
}

func (t ParseError) toJson() jsonisable {
	return jsonList{myString("error"), myString(t.Kind)}
}

func (t Comment) toJson() jsonisable {
	return myString("/* … */")
}

func (t WhitespaceToken) toJson() jsonisable {
	return myString(" ")
}

func (t LiteralToken) toJson() jsonisable {
	return myString(t.Value)
}

func (t IdentToken) toJson() jsonisable {
	return jsonList{myString("ident"), myString(t.Value)}
}

func (t AtKeywordToken) toJson() jsonisable {
	return jsonList{myString("at-keyword"), myString(t.Value)}
}

func (t HashToken) toJson() jsonisable {
	l := jsonList{myString("hash"), myString(t.Value)}
	if t.IsIdentifier {
		l = append(l, myString("id"))
	} else {
		l = append(l, myString("unrestricted"))
	}
	return l
}

func (t StringToken) toJson() jsonisable {
	return jsonList{myString("string"), myString(t.Value)}
}

func (t URLToken) toJson() jsonisable {
	return jsonList{myString("url"), myString(t.Value)}
}

func (t UnicodeRangeToken) toJson() jsonisable {
	return jsonList{myString("unicode-range"), myInt(t.Start), myInt(t.End)}
}

func (t NumberToken) toJson() jsonisable {
	return append(jsonList{myString("number")}, NumericToken(t).toJson()...)
}

func (t PercentageToken) toJson() jsonisable {
	return append(jsonList{myString("percentage")}, NumericToken(t).toJson()...)
}

func (t DimensionToken) toJson() jsonisable {
	return append(append(jsonList{myString("dimension")}, t.NumericToken.toJson()...), myString(t.Unit))
}

func (t ParenthesesBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("()")}, content...)
}

func (t SquareBracketsBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("[]")}, content...)
}

func (t CurlyBracketsBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("{}")}, content...)
}

func (t FunctionBlock) toJson() jsonisable {
	content := toJson(*t.Arguments)
	return append(jsonList{myString("function"), myString(t.Name)}, content...)
}
