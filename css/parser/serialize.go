package parser

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var badPairs = map[[2]string]bool{}

func init() {
	for _, a := range []string{"ident", "at-keyword", "hash", "dimension", "#", "-", "number"} {
		for _, b := range []string{"ident", "function", "url", "number", "percentage", "dimension", "unicode-range"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"ident", "at-keyword", "hash", "dimension"} {
		for _, b := range []string{"-", "-->"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"#", "-", "number", "@"} {
		for _, b := range []string{"ident", "function", "url"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"unicode-range", ".", "+"} {
		for _, b := range []string{"number", "percentage", "dimension"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, b := range []string{"ident", "function", "url", "unicode-range", "-"} {
		badPairs[[2]string{"@", b}] = true
	}
	for _, b := range []string{"ident", "function", "?"} {
		badPairs[[2]string{"unicode-range", b}] = true
	}
	for _, a := range []string{"$", "*", "^", "~", "|"} {
		badPairs[[2]string{a, "="}] = true
	}
	badPairs[[2]string{"ident", "() block"}] = true
	badPairs[[2]string{"|", "|"}] = true
	badPairs[[2]string{"/", "*"}] = true
}

// Serialize nodes to CSS syntax.
// This should be used for `ComponentValue` as it takes care of corner cases such as ``;`` between declarations,
// and consecutive identifiers that would otherwise parse back as the same token.
func Serialize(nodes []Token) string {
	var tmp strings.Builder
	serializeTo(nodes, &tmp)
	return tmp.String()
}

// Serialize this node to CSS syntax
func SerializeOne(node Token) string {
	var tmp strings.Builder
	node.serializeTo(&tmp)
	return tmp.String()
}

// Serialize any string as a CSS identifier
// Returns an Unicode string
// that would parse as an `IdentToken`
// whose value attribute equals the passed `value` argument.
func serializeIdentifier(value string) string {
	if value == "-" {
		return `\-`
	}

	if len(value) >= 2 && value[:2] == "--" {
		return "--" + serializeName(value[2:])
	}
	var result string
	if value[0] == '-' {
		result = "-"
		value = value[1:]
	} else {
		result = ""
	}
	c, w := utf8.DecodeRuneInString(value)
	var suffix string
	switch c {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '_', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		suffix = string(c)
	case '\n':
		suffix = `\A `
	case '\r':
		suffix = `\D `
	case '\f':
		suffix = `\C `
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		suffix = fmt.Sprintf("\\%X", c)
	default:
		if c > 0x7F {
			suffix = string(c)
		} else {
			suffix = "\\" + string(c)
		}

	}
	result += suffix + serializeName(value[w:])
	return result
}

func serializeName(value string) string {
	var chuncks strings.Builder
	for _, c := range value {
		var mapped string
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '-', '_', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			mapped = string(c)
		case '\n':
			mapped = `\A `
		case '\r':
			mapped = `\D `
		case '\f':
			mapped = `\C `
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			mapped = string(c)
		default:
			if c > 0x7F {
				mapped = string(c)
			} else {
				mapped = "\\" + string(c)
			}
		}
		chuncks.WriteString(mapped)
	}
	return chuncks.String()
}

func serializeStringValue(value string) string {
	var chuncks strings.Builder
	for _, c := range value {
		var mapped string
		switch c {
		case '"':
			mapped = `\"`
		case '\\':
			mapped = `\\`
		case '\n':
			mapped = `\A `
		case '\r':
			mapped = `\D `
		case '\f':
			mapped = `\C `
		default:
			mapped = string(c)
		}
		chuncks.WriteString(mapped)
	}
	return chuncks.String()
}

func serializeURL(value string) string {
	var chuncks strings.Builder
	for _, c := range value {
		var mapped string
		switch c {
		case '\'':
			mapped = `\'`
		case '"':
			mapped = `\"`
		case '\\':
			mapped = `\\`
		case ' ':
			mapped = `\ `
		case '\t':
			mapped = `\9 `
		case '\n':
			mapped = `\A `
		case '\r':
			mapped = `\D `
		case '\f':
			mapped = `\C `
		case '(':
			mapped = `\(`
		case ')':
			mapped = `\)`
		default:
			mapped = string(c)
		}
		chuncks.WriteString(mapped)
	}
	return chuncks.String()
}

// http://dev.w3.org/csswg/css-syntax/#serialization-tables
// Serialize an iterable of nodes to CSS syntax,
// writing chunks as Unicode string
// by calling the provided `write` callback.
func serializeTo(nodes []Token, writer io.StringWriter) {
	var previousType string
	for _, node := range nodes {
		serializationType := string(node.Type())
		if literal, ok := node.(LiteralToken); ok {
			serializationType = literal.Value
		}
		if badPairs[[2]string{previousType, serializationType}] {
			writer.WriteString("/**/")
		} else if previousType == "\\" {
			whitespace, ok := node.(WhitespaceToken)
			ok = ok && strings.HasPrefix(whitespace.Value, "\n")
			if !ok {
				writer.WriteString("\n")
			}
		}
		node.serializeTo(writer)
		if serializationType == string(DeclarationT) {
			writer.WriteString(";")
		}
		previousType = serializationType
	}
}

func (t QualifiedRule) serializeTo(writer io.StringWriter) {
	serializeTo(*t.Prelude, writer)
	writer.WriteString("{")
	serializeTo(*t.Content, writer)
	writer.WriteString("}")
}

func (t AtRule) serializeTo(writer io.StringWriter) {
	writer.WriteString("@")
	writer.WriteString(serializeIdentifier(string(t.AtKeyword)))
	serializeTo(*t.Prelude, writer)
	if t.Content == nil {
		writer.WriteString(";")
	} else {
		writer.WriteString("{")
		serializeTo(*t.Content, writer)
		writer.WriteString("}")
	}
}

func (t Declaration) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(string(t.Name)))
	writer.WriteString(":")
	serializeTo(t.Value, writer)
	if t.Important {
		writer.WriteString("!important")
	}
}

func (t ParseError) serializeTo(writer io.StringWriter) {
	switch t.Kind {
	case "bad-string":
		writer.WriteString("\"[bad string]\n")
	case "bad-url":
		writer.WriteString("url([bad url])")
	case ")", "]", "}":
		writer.WriteString(t.Kind)
	case "eof-in-string", "eof-in-url":
		// pass
	default: // pragma: no cover
		panic(fmt.Sprint("Can not serialize token", t))
	}
}

func (t Comment) serializeTo(writer io.StringWriter) {
	writer.WriteString("/*")
	writer.WriteString(t.Value)
	writer.WriteString("*/")
}

func (t WhitespaceToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
}

func (t LiteralToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
}

func (t IdentToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(string(t.Value)))
}

func (t AtKeywordToken) serializeTo(writer io.StringWriter) {
	writer.WriteString("@")
	writer.WriteString(serializeIdentifier(string(t.Value)))
}

func (t HashToken) serializeTo(writer io.StringWriter) {
	writer.WriteString("#")
	if t.IsIdentifier {
		writer.WriteString(serializeIdentifier(t.Value))
	} else {
		writer.WriteString(serializeName(t.Value))
	}
}

func (t StringToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(`"`)
	writer.WriteString(serializeStringValue(t.Value))
	if !t.isError {
		writer.WriteString(`"`)
	}
}

func (t URLToken) serializeTo(writer io.StringWriter) {
	tmp := `url(` + serializeURL(t.Value) + ")"
	if t.isError == errorInString {
		tmp = tmp[:len(tmp)-2]
	} else if t.isError == errorInURL {
		tmp = tmp[:len(tmp)-1]
	}
	writer.WriteString(tmp)
}

func (t UnicodeRangeToken) serializeTo(writer io.StringWriter) {
	if t.End == t.Start {
		writer.WriteString(fmt.Sprintf("U+%X", t.Start))
	} else {
		writer.WriteString(fmt.Sprintf("U+%X-%X", t.Start, t.End))
	}
}

func (t NumberToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Representation)
}

func (t PercentageToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Representation)
	writer.WriteString("%")
}

func (t DimensionToken) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Representation)
	// Disambiguate with scientific notation
	unit := string(t.Unit)
	if unit == "e" || unit == "E" || strings.HasPrefix(unit, "e-") || strings.HasPrefix(unit, "E-") {
		writer.WriteString("\\65 ")
		writer.WriteString(serializeName(unit[1:]))
	} else {
		writer.WriteString(serializeIdentifier(unit))
	}
}

func (t ParenthesesBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("(")
	serializeTo(*t.Content, writer)
	writer.WriteString(")")
}

func (t SquareBracketsBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("[")
	serializeTo(*t.Content, writer)
	writer.WriteString("]")
}

func (t CurlyBracketsBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("{")
	serializeTo(*t.Content, writer)
	writer.WriteString("}")
}

func (t FunctionBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(string(t.Name)))
	writer.WriteString("(")
	serializeTo(*t.Arguments, writer)
	if t.Arguments != nil {
		var fn Token = t
		for asFn, ok := fn.(FunctionBlock); ok; asFn, ok = fn.(FunctionBlock) {
			if len(*asFn.Arguments) == 0 {
				break
			}
			lastArg := (*asFn.Arguments)[len(*asFn.Arguments)-1]
			if asParse, ok := lastArg.(ParseError); ok && asParse.Kind == "eof-in-string" {
				return
			}
			fn = lastArg
		}
	}
	writer.WriteString(")")
}
