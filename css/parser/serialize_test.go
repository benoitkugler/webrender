package parser

import (
	"fmt"
	"testing"
)

func TestSerialization(t *testing.T) {
	inputs, resJson := loadJson("component_value_list.json")
	runTest(t, inputs, resJson, func(css string) []Token {
		parsed := tokenizeString(css, true)
		return tokenizeString(Serialize(parsed), true)
	})
}

func TestIdentifiers(t *testing.T) {
	source := "\fezeze"
	ref := tokenizeString(source, false)
	resToTest := tokenizeString(Serialize(ref), false)
	res, err := marshalJSON(resToTest)
	if err != nil {
		t.Fatal(err)
	}
	refJson, err := marshalJSON(ref)
	if err != nil {
		t.Fatal(err)
	}
	if res != refJson {
		t.Fatalf(fmt.Sprintf("expected \n %s \n got \n %s \n", ref, res))
	}
}

func TestSkip(t *testing.T) {
	source := `
    /* foo */
    @media print {
        #foo {
            width: /* bar*/4px;
            color: green;
        }
    }
    `
	noWs := ParseStylesheetBytes([]byte(source), false, true)
	noComment := ParseStylesheetBytes([]byte(source), true, false)
	default_ := tokenizeString(source, false)
	if Serialize(noWs) == source {
		t.Fail()
	}
	if Serialize(noComment) == source {
		t.Fail()
	}
	if Serialize(default_) != source {
		t.Fail()
	}
}

func TestCommentEof(t *testing.T) {
	source := "/* foo "
	parsed := tokenizeString(source, false)
	if Serialize(parsed) != "/* foo */" {
		t.Fail()
	}
}

func TestParseDeclarationValueColor(t *testing.T) {
	source := "color:#369"
	declaration := parseOneDeclarationString(source, false)
	decl, ok := declaration.(Declaration)
	if !ok || (ParseColor(decl.Value[0]).RGBA != RGBA{R: 0.2, G: 0.4, B: 0.6, A: 1}) {
		t.Fail()
	}
	if SerializeOne(declaration) != source {
		t.Fail()
	}
}

func TestSerializeRules(t *testing.T) {
	source := `@import "a.css"; foo#bar.baz { color: red } /**/ @media print{}`
	rules := ParseRuleListString(source, false, false)
	if Serialize(rules) != source {
		t.Fail()
	}
}

func TestSerializeDeclarations(t *testing.T) {
	source := "color: #123; /**/ @top-left {} width:7px !important;"
	rules := ParseDeclarationListString(source, false, false)
	if Serialize(rules) != source {
		t.Fail()
	}
}

func TestBackslashDelim(t *testing.T) {
	source := "\\\nfoo"
	tokens := tokenizeString(source, false)
	if len(tokens) != 3 {
		t.Fatalf("bad token length : expected 3 got %d", len(tokens))
	}
	if lit, ok := tokens[0].(LiteralToken); !ok || lit.Value != "\\" {
		t.Errorf("expected litteral \\ got %s", tokens[0])
	}
	if tokens[1].Type() != WhitespaceTokenT || tokens[2].Type() != IdentTokenT {
		t.Errorf("expected whitespace and ident got : %s and %s", tokens[1].Type(), tokens[2].Type())
	}
	tokens = []Token{tokens[0], tokens[2]}
	ser := Serialize(tokens)
	if ser != source {
		t.Errorf("expected %s got %s", source, ser)
	}
}

func TestDataurl(t *testing.T) {
	input := `@import "data:text/css;charset=utf-16le;base64,\
				bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=";`
	fmt.Println(Serialize(tokenizeString(input, true)))
}

func TestDebug(t *testing.T) {
	ls := tokenizeString(`.foo\:bar`, false)
	fmt.Println(Serialize(ls))
}
