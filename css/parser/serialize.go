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

func Serialize(l []Token) string {
	var w strings.Builder
	serializeTo(l, &w)
	return w.String()
}

// Serialize any string as a CSS identifier
// Returns an Unicode string
// that would parse as an `Ident`
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

func (t Literal) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
}

func (t ParseError) serializeTo(writer io.StringWriter) {
	switch t.kind {
	case errBadString:
		writer.WriteString("\"[bad string]\n")
	case errBadURL:
		writer.WriteString("url([bad url])")
	case errP, errB, errC:
		writer.WriteString(string(t.kind))
	case errEofInString, errEofInUrl:
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

func (t Whitespace) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
}

func (t Ident) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(t.Value))
}

func (t AtKeyword) serializeTo(writer io.StringWriter) {
	writer.WriteString("@")
	writer.WriteString(serializeIdentifier(t.Value))
}

func (t Hash) serializeTo(writer io.StringWriter) {
	writer.WriteString("#")
	if t.isIdentifier() {
		writer.WriteString(serializeIdentifier(t.Value))
	} else {
		writer.WriteString(serializeName(t.Value))
	}
}

func (t String) serializeTo(writer io.StringWriter) {
	writer.WriteString(`"`)
	writer.WriteString(serializeStringValue(t.Value))
	if !t.isError() {
		writer.WriteString(`"`)
	}
}

func (t URL) serializeTo(writer io.StringWriter) {
	tmp := `url(` + serializeURL(t.Value) + ")"
	if t.flag&isErrorInString != 0 {
		tmp = tmp[:len(tmp)-2]
	} else if t.flag&isErrorInURL != 0 {
		tmp = tmp[:len(tmp)-1]
	}
	writer.WriteString(tmp)
}

func (t UnicodeRange) serializeTo(writer io.StringWriter) {
	if t.End == t.Start {
		writer.WriteString(fmt.Sprintf("U+%X", t.Start))
	} else {
		writer.WriteString(fmt.Sprintf("U+%X-%X", t.Start, t.End))
	}
}

func (t Number) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
}

func (t Percentage) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
	writer.WriteString("%")
}

func (t Dimension) serializeTo(writer io.StringWriter) {
	writer.WriteString(t.Value)
	// Disambiguate with scientific notation
	if t.Unit == "e" || t.Unit == "E" || strings.HasPrefix(t.Unit, "e-") || strings.HasPrefix(t.Unit, "E-") {
		writer.WriteString("\\65 ")
		writer.WriteString(serializeName(t.Unit[1:]))
	} else {
		writer.WriteString(serializeIdentifier(t.Unit))
	}
}

func (t ParenthesesBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("(")
	serializeTo(t.Arguments, writer)
	writer.WriteString(")")
}

func (t SquareBracketsBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("[")
	serializeTo(t.Arguments, writer)
	writer.WriteString("]")
}

func (t CurlyBracketsBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString("{")
	serializeTo(t.Arguments, writer)
	writer.WriteString("}")
}

func (t FunctionBlock) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(string(t.Name)))
	writer.WriteString("(")
	serializeTo(t.Arguments, writer)

	// recursively check for a parsing error
	var argVal Token = t
	for fn, ok := argVal.(FunctionBlock); ok; fn, ok = argVal.(FunctionBlock) {
		if len(fn.Arguments) == 0 {
			break
		}
		lastArg := (fn.Arguments)[len(fn.Arguments)-1]
		if asParse, ok := lastArg.(ParseError); ok && asParse.kind == errEofInString {
			return
		}
		argVal = lastArg
	}
	writer.WriteString(")")
}

// http://drafts.csswg.org/csswg/css-syntax/#serialization-tables
// Serialize an iterable of nodes to CSS syntax,
// writing chunks as Unicode string
// by calling the provided `write` callback.
func serializeTo(nodes []Token, writer io.StringWriter) {
	var previousType string
	for _, node := range nodes {
		serializationType := node.Kind().String()
		if literal, ok := node.(Literal); ok {
			serializationType = literal.Value
		}
		if badPairs[[2]string{previousType, serializationType}] {
			writer.WriteString("/**/")
		} else if previousType == "\\" {
			whitespace, ok := node.(Whitespace)
			ok = ok && strings.HasPrefix(whitespace.Value, "\n")
			if !ok {
				writer.WriteString("\n")
			}
		}
		node.serializeTo(writer)
		previousType = serializationType
	}
}

func (t QualifiedRule) serializeTo(writer io.StringWriter) {
	serializeTo(t.Prelude, writer)
	writer.WriteString("{")
	serializeTo(t.Content, writer)
	writer.WriteString("}")
}

func (t AtRule) serializeTo(writer io.StringWriter) {
	writer.WriteString("@")
	writer.WriteString(serializeIdentifier(t.AtKeyword))
	serializeTo(t.Prelude, writer)
	if t.Content == nil {
		writer.WriteString(";")
	} else {
		writer.WriteString("{")
		serializeTo(t.Content, writer)
		writer.WriteString("}")
	}
}

func (t Declaration) serializeTo(writer io.StringWriter) {
	writer.WriteString(serializeIdentifier(t.Name))
	writer.WriteString(":")
	serializeTo(t.Value, writer)
	if t.Important {
		writer.WriteString("!important")
	}
}
