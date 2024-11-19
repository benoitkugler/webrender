package parser

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils"
)

func (c Color) dump() interface{} {
	switch c.Type {
	case ColorInvalid:
		return nil
	case ColorCurrentColor:
		return "currentColor"
	default:
		return []utils.Fl{utils.Round6(c.RGBA.R), utils.Round6(c.RGBA.G), utils.Round6(c.RGBA.B), utils.Round6(c.RGBA.A)}
	}
}

func (t ParseError) dump() interface{} {
	var kind string
	switch t.kind {
	case errBadString:
		kind = "bad-string"
	case errBadURL:
		kind = "bad-url"
	case errP, errB, errC:
		kind = string(t.kind)
	case errEofInString:
		kind = "eof-in-string"
	case errEofInUrl:
		kind = "eof-in-url"
	case errEmpty:
		kind = "empty"
	case errExtraInput:
		kind = "extra-input"
	case errInvalid:
		kind = "invalid"
	}
	return []string{"error", kind}
}

func (t Comment) dump() interface{} { return "/* … */" }

func (t Whitespace) dump() interface{} { return " " }

func (t Literal) dump() interface{} {
	return t.Value
}

func (t Ident) dump() interface{} {
	return []string{"ident", t.Value}
}

func (t AtKeyword) dump() interface{} {
	return []string{"at-keyword", t.Value}
}

func (t Hash) dump() interface{} {
	out := []string{"hash", t.Value}
	if t.isIdentifier() {
		out = append(out, "id")
	} else {
		out = append(out, "unrestricted")
	}
	return out
}

func (t String) dump() interface{} {
	return []string{"string", t.Value}
}

func (t URL) dump() interface{} {
	return []string{"url", t.Value}
}

func (t UnicodeRange) dump() interface{} {
	return []interface{}{"unicode-range", t.Start, t.End}
}

func (t Number) dump() interface{} {
	return append([]interface{}{"number"}, t.dumpNumeric()...)
}

func (t Percentage) dump() interface{} {
	return append([]interface{}{"percentage"}, t.dumpNumeric()...)
}

func (t Dimension) dump() interface{} {
	return append(append([]interface{}{"dimension"}, t.dumpNumeric()...), t.Unit)
}

func (t ParenthesesBlock) dump() interface{} {
	content := dumpList(fromT(t.Arguments))
	return append([]interface{}{"()"}, content...)
}

func (t SquareBracketsBlock) dump() interface{} {
	content := dumpList(fromT(t.Arguments))
	return append([]interface{}{"[]"}, content...)
}

func (t CurlyBracketsBlock) dump() interface{} {
	content := dumpList(fromT(t.Arguments))
	return append([]interface{}{"{}"}, content...)
}

func (t FunctionBlock) dump() interface{} {
	content := dumpList(fromT(t.Arguments))
	return append([]interface{}{"function", t.Name}, content...)
}

func (t numberVal) dumpNumeric() []interface{} {
	l := []interface{}{t.Value, t.ValueF}
	if t.IsInt() {
		l = append(l, "integer")
	} else {
		l = append(l, "number")
	}
	return l
}

func (t QualifiedRule) dump() interface{} {
	prelude := dumpList(fromT(t.Prelude))
	content := dumpList(fromT(t.Content))
	return []interface{}{"qualified rule", prelude, content}
}

func (t AtRule) dump() interface{} {
	prelude := dumpList(fromT(t.Prelude))
	var content []interface{} // preserve nil
	if t.Content != nil {
		content = dumpList(fromT(t.Content))
	}
	return []interface{}{"at-rule", t.AtKeyword, prelude, content}
}

func (t Declaration) dump() interface{} {
	content := dumpList(fromT(t.Value))
	return []interface{}{"declaration", t.Name, content, t.Important}
}

func TestIdentifiers(t *testing.T) {
	source := "\fezeze"
	ref := tokenizeString(source, false)
	resToTest := tokenizeString(Serialize(ref), false)
	res, err := marshalJSON(fromT(resToTest))
	if err != nil {
		t.Fatal(err)
	}
	refJson, err := marshalJSON(fromT(ref))
	if err != nil {
		t.Fatal(err)
	}
	if res != refJson {
		t.Fatalf(fmt.Sprintf("expected \n %s \n got \n %s \n", ref, res))
	}
}

func TestCommentEof(t *testing.T) {
	source := "/* foo "
	parsed := tokenizeString(source, false)
	if Serialize(parsed) != "/* foo */" {
		t.Fail()
	}
}

func TestBackslashDelim(t *testing.T) {
	source := "\\\nfoo"
	tokens := tokenizeString(source, false)
	if len(tokens) != 3 {
		t.Fatalf("bad token length : expected 3 got %d", len(tokens))
	}
	if lit, ok := tokens[0].(Literal); !ok || lit.Value != "\\" {
		t.Errorf("expected litteral \\ got %s", tokens[0])
	}
	if k1, k2 := tokens[1].Kind(), tokens[2].Kind(); k1 != KWhitespace || k2 != KIdent {
		t.Errorf("expected whitespace and ident got : %s and %s", k1, k2)
	}
	tokens = []Token{tokens[0], tokens[2]}
	ser := Serialize(tokens)
	if ser != source {
		t.Errorf("expected %s got %s", source, ser)
	}
}

func TestSerialization(t *testing.T) {
	inputs, resJson := loadJson(t, "component_value_list.json")
	runTest(t, inputs, resJson, func(css string) []TC {
		parsed := Tokenize([]byte(css), true)
		return fromT(tokenizeString(Serialize(parsed), true))
	})
}
