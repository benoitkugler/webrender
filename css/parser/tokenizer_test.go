package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func loadJson(t *testing.T, filename string) ([]string, []string) {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("css-parsing-tests", filename))
	tu.AssertNoErr(t, err)

	var l []interface{}
	err = json.Unmarshal(b, &l)
	tu.AssertNoErr(t, err)

	tu.AssertEqual(t, len(l)%2, 0)
	inputs, resJsons := make([]string, len(l)/2), make([]string, len(l)/2)
	for i := 0; i < len(l); i += 2 {
		inputs[i/2] = l[i].(string)
		res, err := json.Marshal(l[i+1])
		tu.AssertNoErr(t, err)
		resJsons[i/2] = string(res)
	}
	return inputs, resJsons
}

type TC interface{ dump() interface{} }

func fromT(l []Token) []TC {
	if l == nil {
		return nil
	}
	out := make([]TC, len(l))
	for i, v := range l {
		out[i] = v.(TC)
	}
	return out
}

func fromC(l []Compound) []TC {
	out := make([]TC, len(l))
	for i, v := range l {
		out[i] = v.(TC)
	}
	return out
}

func dumpList(l []TC) []interface{} {
	// if l == nil {
	// 	return nil
	// }
	out := make([]interface{}, len(l))
	for i, v := range l {
		out[i] = v.dump()
	}
	return out
}

func marshalJSON(l []TC) (string, error) {
	b, err := json.Marshal(dumpList(l))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func runTest(t *testing.T, css, resJson []string, fn func(input string) []TC) {
	for i, input := range css {
		resToTest := fn(input)
		res, err := marshalJSON(resToTest)
		tu.AssertNoErr(t, err)

		if res != resJson[i] {
			t.Fatalf(fmt.Sprintf("input %d : \n %s \n failed : expected \n %s \n got  \n %s \n", i, input, resJson[i], res))
		}
	}
}

func runTestOne(t *testing.T, css, resJson []string, fn func(input string) TC) {
	t.Helper()

	for i, input := range css {
		resToTest := fn(input)
		b, err := json.Marshal(resToTest.dump())
		if err != nil {
			t.Fatal(err)
		}
		res := string(b)
		if res != resJson[i] {
			t.Fatalf(fmt.Sprintf("input %d : \n %s \n failed : expected \n %s \n got  \n %s \n", i, input, resJson[i], res))
		}
	}
}

func tokenizeString(css string, skipComments bool) []Token {
	return Tokenize([]byte(css), skipComments)
}

func TestComponentValueList(t *testing.T) {
	inputs, resJson := loadJson(t, "component_value_list.json")
	runTest(t, inputs, resJson, func(s string) []TC {
		return fromT(tokenizeString(s, true))
	})
}

func TestOneComponentValue(t *testing.T) {
	inputs, resJson := loadJson(t, "one_component_value.json")
	runTestOne(t, inputs, resJson, func(input string) TC {
		return parseOneComponentValue([]byte(input), true).(TC)
	})
}

func TestNoSkipComments(t *testing.T) {
	source := `
    /* foo */
    @media print {
        #foo {
            width: /* bar*/4px;
            color: green;
        }
    }
    `
	tokens := tokenizeString(source, false)
	tu.AssertEqual(t, Serialize(tokens), source)
}

func TestParseDeclarationValueColor(t *testing.T) {
	source := "color:#369"
	declaration := parseOneDeclarationString(source, false)
	decl, ok := declaration.(Declaration)
	tu.AssertEqual(t, ok, true)
	tu.AssertEqual(t, ParseColor(decl.Value[0]).RGBA, RGBA{R: 0.2, G: 0.4, B: 0.6, A: 1})
}

func TestDataurl(t *testing.T) {
	input := `@import "data:text/css;charset=utf-16le;base64,\
				bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=";`
	s := Serialize(tokenizeString(input, true))
	tu.AssertEqual(t, s, `@import "data:text/css;charset=utf-16le;base64,				bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=";`)
}
