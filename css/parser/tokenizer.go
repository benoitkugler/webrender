package parser

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/benoitkugler/webrender/utils"
)

var (
	numberRe    = regexp.MustCompile(`^[-+]?([0-9]*\.)?[0-9]+([eE][+-]?[0-9]+)?`)
	hexEscapeRe = regexp.MustCompile(`^([0-9A-Fa-f]{1,6})[ \n\t]?`)
)

type nestedBlock struct {
	tokens  *[]Token
	endChar byte
}

func tokenizeString(css string, skipComments bool) []Token {
	return Tokenize([]byte(css), skipComments)
}

// Tokenize parses a list of component values.
// If `skipComments` is true, ignore CSS comments :
// the return values (and recursively its blocks and functions)
// will not contain any `Comment` object.
func Tokenize(css []byte, skipComments bool) []Token {
	// This turns out to be faster than a regexp:
	css = bytes.ReplaceAll(css, []byte("\u0000"), []byte("\uFFFD"))
	css = bytes.ReplaceAll(css, []byte("\r\n"), []byte("\n"))
	css = bytes.ReplaceAll(css, []byte("\r"), []byte("\n"))
	css = bytes.ReplaceAll(css, []byte("\f"), []byte("\n"))

	length := len(css)
	tokenStartPos, pos := 0, 0
	line, lastNewline := 1, -1
	var out []Token  // possibly nested tokens
	ts := &out       // current stack of tokens
	var endChar byte // Pop the stack when encountering this character.
	var stack []nestedBlock
	var err error

mainLoop:
	for pos < length {
		newline := bytes.LastIndexByte(css[tokenStartPos:pos], '\n')
		if newline != -1 {
			newline += tokenStartPos
			line += 1 + bytes.Count(css[tokenStartPos:newline], []byte{'\n'})
			lastNewline = newline
		}
		// First character in a line is in column 1.
		column := pos - lastNewline
		tokenPos := newPosition(line, column)

		tokenStartPos = pos
		c := css[pos]

		switch c {
		case ' ', '\n', '\t':
			pos += 1
			for ; pos < length; pos += 1 {
				u := css[pos]
				if !(u == ' ' || u == '\n' || u == '\t') {
					break
				}
			}
			value := css[tokenStartPos:pos]
			*ts = append(*ts, WhitespaceToken{Pos: tokenPos, Value: string(value)})
			continue
		case 'U', 'u':
			if pos+2 < length && css[pos+1] == '+' && strings.ContainsRune("0123456789abcdefABCDEF?", rune(css[pos+2])) {
				var start, end int64
				start, end, pos, err = consumeUnicodeRange(css, pos+2)
				if err != nil {
					*ts = append(*ts, ParseError{Pos: tokenPos, Kind: "invalid number", Message: err.Error()})
				} else {
					*ts = append(*ts, UnicodeRangeToken{Pos: tokenPos, Start: uint32(start), End: uint32(end)})
				}
				continue
			}
		}
		if bytes.HasPrefix(css[pos:], []byte("-->")) { // Check before identifiers
			*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: "-->"})
			pos += 3
			continue
		} else if isIdentStart(css, pos) {
			var value string
			value, pos = consumeIdent(css, pos)
			if !(pos < length && css[pos] == '(') { // Not a function
				*ts = append(*ts, IdentToken{Pos: tokenPos, Value: LowerableString(value)})
				continue
			}
			pos += 1 // Skip the "("
			if utils.AsciiLower(value) == "url" {
				urlPos := pos
				for urlPos < length && (css[urlPos] == ' ' || css[urlPos] == '\n' || css[urlPos] == '\t') {
					urlPos += 1
				}
				if urlPos >= length || (css[urlPos] != '"' && css[urlPos] != '\'') {
					var addValue bool
					value, pos, addValue, err = consumeUrl(css, pos)
					if addValue {
						var isError uint8
						if err != nil {
							switch err.Error() {
							case "eof-in-string":
								isError = errorInString
							case "eof-in-url":
								isError = errorInURL
							}
						}
						*ts = append(*ts, URLToken{Pos: tokenPos, Value: value, isError: isError})
					}
					if err != nil {
						*ts = append(*ts, ParseError{Pos: tokenPos, Kind: err.Error(), Message: err.Error()})
					}
					continue
				}
			}
			funcBlock := FunctionBlock{
				Pos:       tokenPos,
				Name:      LowerableString(value),
				Arguments: new([]Token),
			}
			*ts = append(*ts, funcBlock)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ')'
			ts = funcBlock.Arguments
			continue
		}

		value := css[pos:]
		match := numberRe.FindIndex(value)
		if match != nil {
			repr := string(css[pos+match[0] : pos+match[1]])
			pos += match[1]
			value, _ := strconv.ParseFloat(repr, 32)
			if value == 0 {
				value = 0. // workaround -0
			}
			_, err = strconv.ParseInt(repr, 10, 0)
			isInt := err == nil
			n := NumericToken{
				Pos:            tokenPos,
				Representation: repr,
				IsInteger:      isInt,
				Value:          utils.Fl(value),
			}
			if pos < length && isIdentStart(css, pos) {
				var unit string
				unit, pos = consumeIdent(css, pos)
				*ts = append(*ts, DimensionToken{NumericToken: n, Unit: LowerableString(unit)})
			} else if pos < length && css[pos] == '%' {
				pos += 1
				*ts = append(*ts, PercentageToken(n))
			} else {
				*ts = append(*ts, NumberToken(n))
			}
			continue
		}
		switch c {
		case '@':
			pos += 1
			if pos < length && isIdentStart(css, pos) {
				var ident string
				ident, pos = consumeIdent(css, pos)
				*ts = append(*ts, AtKeywordToken{Pos: tokenPos, Value: LowerableString(ident)})
			} else {
				*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: "@"})
			}
		case '#':
			pos += 1
			if pos < length {
				r, _ := utf8.DecodeRune(css[pos:])
				if ('0' <= r && r <= '9' || 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '-' || r == '_') ||
					r > 0x7F || // Non-ASCII
					(r == '\\' && !bytes.HasPrefix(css[pos:], []byte("\\\n"))) { // Valid escape
					isIdentifier := isIdentStart(css, pos)
					var ident string
					ident, pos = consumeIdent(css, pos)
					*ts = append(*ts, HashToken{Pos: tokenPos, Value: ident, IsIdentifier: isIdentifier})
					continue
				}
			}
			*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: "#"})
		case '{':
			brack := CurlyBracketsBlock{Pos: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = '}'
			ts = brack.Content
			pos += 1
		case '[':
			brack := SquareBracketsBlock{Pos: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ']'
			ts = brack.Content
			pos += 1
		case '(':
			brack := ParenthesesBlock{Pos: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ')'
			ts = brack.Content
			pos += 1
		case 0: // remove this case to avoid false comparaison with endChar
		case endChar: // Matching }, ] or ), or 0
			// The top-level endChar is 0, so we never get here if the stack is empty.
			var block nestedBlock
			block, stack = stack[len(stack)-1], stack[:len(stack)-1]
			ts, endChar = block.tokens, block.endChar
			pos += 1
		case '}', ']', ')':
			*ts = append(*ts, ParseError{Pos: tokenPos, Kind: string(rune(c)), Message: "Unmatched " + string(rune(c))})
			pos += 1
		case '\'', '"':
			var (
				quotedString string
				addValue     bool
			)
			quotedString, pos, addValue, err = consumeQuotedString(css, pos)
			if addValue {
				*ts = append(*ts, StringToken{Pos: tokenPos, Value: quotedString, isError: err != nil})
			}
			if err != nil {
				*ts = append(*ts, ParseError{Pos: tokenPos, Kind: err.Error(), Message: "bad string token"})
			}
		default:
			switch {
			case bytes.HasPrefix(css[pos:], []byte("/*")): // Comment
				index := bytes.Index(css[pos+2:], []byte("*/"))
				pos += 2 + index
				if index == -1 {
					if !skipComments {
						*ts = append(*ts, Comment{Pos: tokenPos, Value: string(css[tokenStartPos+2:])})
					}
					break mainLoop
				}
				if !skipComments {
					*ts = append(*ts, Comment{Pos: tokenPos, Value: string(css[tokenStartPos+2 : pos])})
				}
				pos += 2
			case bytes.HasPrefix(css[pos:], []byte("<!--")):
				*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: "<!--"})
				pos += 4
			case bytes.HasPrefix(css[pos:], []byte("||")):
				*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: "||"})
				pos += 2
			case c == '~' || c == '|' || c == '^' || c == '$' || c == '*':
				pos += 1
				if bytes.HasPrefix(css[pos:], []byte{'='}) {
					pos += 1
					*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: string(rune(c)) + "="})
				} else {
					*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: string(rune(c))})
				}
			default:
				r, w := utf8.DecodeRune(css[pos:])
				pos += w
				*ts = append(*ts, LiteralToken{Pos: tokenPos, Value: string(r)})
			}
		}
	}
	return out
}

const (
	charUnicodeRange = "0123456789abcdefABCDEF"
	nonPrintable     = "\"'(\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f"
)

// Return true if the given character is a name-start code point.
func isNameStart(css []byte, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#name-start-code-point
	c, _ := utf8.DecodeRune(css[pos:])
	return c > 0x7F || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_'
}

// Return true if the given position is the start of a CSS identifier.
func isIdentStart(css []byte, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#would-start-an-identifier
	if isNameStart(css, pos) {
		return true
	} else if css[pos] == '-' {
		pos += 1
		// Name-start code point
		nameStart := pos < len(css) && (isNameStart(css, pos) || css[pos] == '-')
		// Valid escape
		validEscape := css[pos] == '\\' && !bytes.HasPrefix(css[pos:], []byte("\\\n"))
		return nameStart || validEscape
	} else if css[pos] == '\\' {
		return !bytes.HasPrefix(css[pos:], []byte("\\\n"))
	}
	return false
}

func consumeIdent(value []byte, pos int) (string, int) {
	// http://dev.w3.org/csswg/css-syntax/#consume-a-name
	var chunks strings.Builder
	L := len(value)
	startPos := pos
	for pos < L {
		c, w := utf8.DecodeRune(value[pos:])
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz-_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", c) || c > 0x7F {
			pos += w
		} else if c == '\\' && !bytes.HasPrefix(value[pos:], []byte("\\\n")) {
			// Valid escape
			chunks.Write(value[startPos:pos])
			var car string
			car, pos = consumeEscape(value, pos+w)
			chunks.WriteString(car)
			startPos = pos
		} else {
			break
		}
	}
	chunks.Write(value[startPos:pos])
	return chunks.String(), pos
}

// Return the range
// http://dev.w3.org/csswg/css-syntax/#consume-a-unicode-range-token
func consumeUnicodeRange(css []byte, pos int) (start, end int64, newPos int, err error) {
	length := len(css)
	startPos := pos
	maxPos := utils.MinInt(pos+6, length)
	var _start, _end string
	for pos < maxPos {
		r, w := utf8.DecodeRune(css[pos:])
		if !strings.ContainsRune(charUnicodeRange, r) {
			break
		}
		pos += w
	}
	_start = string(css[startPos:pos])
	questionMarks := 0
	// Same maxPos as before: total of hex digits && question marks <= 6
	for pos < maxPos {
		r, w := utf8.DecodeRune(css[pos:])
		if r != '?' {
			break
		}
		pos += w
		questionMarks += 1
	}

	if questionMarks != 0 {
		_end = _start + strings.Repeat("F", questionMarks)
		_start = _start + strings.Repeat("0", questionMarks)
	} else if pos+1 < length && css[pos] == '-' && strings.ContainsRune(charUnicodeRange, rune(css[pos+1])) {
		pos += utf8.RuneLen(rune(css[pos+1]))
		startPos = pos
		maxPos = utils.MinInt(pos+6, length)
		for pos < maxPos {
			r, w := utf8.DecodeRune(css[pos:])
			if !strings.ContainsRune(charUnicodeRange, r) {
				break
			}
			pos += w
		}
		_end = string(css[startPos:pos])
	} else {
		_end = _start
	}
	start, err = strconv.ParseInt(_start, 16, 0)
	if err != nil {
		newPos = pos
		return
	}
	end, err = strconv.ParseInt(_end, 16, 0)
	return start, end, pos, err
}

// http://dev.w3.org/csswg/css-syntax/#consume-a-url-token
func consumeUrl(css []byte, pos int) (value string, newPos int, addValue bool, err error) {
	length := len(css)
	// Skip whitespace
	for pos < length && strings.ContainsRune(" \n\t", rune(css[pos])) {
		pos += 1
	}
	if pos >= length { // EOF
		return "", pos, true, errors.New("eof-in-url")
	}
	c := rune(css[pos])
	if c == '"' || c == '\'' {
		value, pos, addValue, err = consumeQuotedString(css, pos)
	} else if c == ')' {
		return "", pos + 1, true, nil
	} else {
		var chunks strings.Builder
		startPos := pos
	mainLoop:
		for {
			if pos >= length { // EOF
				chunks.Write(css[startPos:pos])
				return chunks.String(), pos, true, errors.New("eof-in-url")
			}
			c, w := utf8.DecodeRune(css[pos:])
			switch {
			case c == ')':
				chunks.Write(css[startPos:pos])
				pos += w
				return chunks.String(), pos, true, nil
			case c == ' ' || c == '\n' || c == '\t':
				chunks.Write(css[startPos:pos])
				value = chunks.String()
				pos += w
				break mainLoop
			case c == '\\' && !bytes.HasPrefix(css[pos:], []byte("\\\n")):
				// Valid escape
				chunks.Write(css[startPos:pos])
				var cs string
				cs, pos = consumeEscape(css, pos+w)
				chunks.WriteString(cs)
				startPos = pos
			default:
				pos += w
				// http://dev.w3.org/csswg/css-syntax/#non-printable-character
				if strings.ContainsRune(nonPrintable, c) {
					err = errors.New("non printable char")
					break mainLoop
				}
			}
		}
	}

	if err == nil {
		for pos < length {
			r, w := utf8.DecodeRune(css[pos:])
			if strings.ContainsRune(" \n\t", r) {
				pos += w
			} else {
				break
			}
		}
		if pos < length {
			if css[pos] == ')' {
				return value, pos + 1, true, err
			}
		} else {
			if err == nil {
				err = errors.New("eof-in-url")
			}
			return value, pos, true, err
		}
	}

	// http://dev.w3.org/csswg/css-syntax/#consume-the-remnants-of-a-bad-url0
	for pos < length {
		if bytes.HasPrefix(css[pos:], []byte("\\)")) {
			pos += 2
		} else if css[pos] == ')' {
			pos += 1
			break
		} else {
			_, w := utf8.DecodeRune(css[pos:])
			pos += w
		}
	}
	return "", pos, false, errors.New("bad-url") // bad-url
}

// Returns unescapedValue
// http://dev.w3.org/csswg/css-syntax/#consume-a-string-token
// css[pos] is assumed to be a quote
func consumeQuotedString(css []byte, pos int) (string, int, bool, error) {
	quote := rune(css[pos])
	pos += 1
	var chunks strings.Builder
	length := len(css)
	startPos := pos
	hasBroken := false
mainLoop:
	for pos < length {
		c, w := utf8.DecodeRune(css[pos:])
		switch c {
		case quote:
			chunks.Write(css[startPos:pos])
			pos += w
			hasBroken = true
			break mainLoop
		case '\\':
			chunks.Write(css[startPos:pos])
			pos += w
			if pos < length {
				if css[pos] == '\n' { // Ignore escaped newlines
					pos += 1
				} else {
					var cs string
					cs, pos = consumeEscape(css, pos)
					chunks.WriteString(cs)
				}
			} // else: Escaped EOF, do nothing
			startPos = pos
		case '\n': // Unescaped newline
			return "", pos, false, errors.New("bad-string") // bad-string
		default:
			pos += w
		}
	}
	var err error
	if !hasBroken {
		chunks.Write(css[startPos:pos])
		err = errors.New("eof-in-string")
	}
	return chunks.String(), pos, true, err
}

// Return (unescapedChar, newPos).
// Assumes a valid escape: pos is just after '\' and not followed by '\n'.
func consumeEscape(css []byte, pos int) (string, int) {
	// http://dev.w3.org/csswg/css-syntax/#consume-an-escaped-character
	hexMatch := hexEscapeRe.FindSubmatch(css[pos:])
	if len(hexMatch) >= 2 {
		codepoint, err := strconv.ParseInt(string(hexMatch[1]), 16, 0)
		if err != nil {
			// the regexp ensure its a valid hex number
			panic(fmt.Sprintf("codepoint should be valid hexadecimal, got %s", hexMatch[0]))
		}
		char := "\uFFFD"
		if 0 < codepoint && codepoint <= unicode.MaxRune {
			char = string(rune(codepoint))
		}
		return char, pos + len(hexMatch[0])
	} else if pos < len(css) {
		r, w := utf8.DecodeRune(css[pos:])
		return string(r), pos + w
	} else {
		return "\uFFFD", pos
	}
}
