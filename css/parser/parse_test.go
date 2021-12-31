package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func loadJson(t *testing.T, filename string) ([]string, []string) {
	t.Helper()

	b, err := ioutil.ReadFile(filepath.Join("css-parsing-tests", filename))
	if err != nil {
		t.Fatal(err)
	}
	var l []interface{}
	if err = json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	if len(l)%2 != 0 {
		t.Fatal("number of tests in list should be even !")
	}
	inputs, resJsons := make([]string, len(l)/2), make([]string, len(l)/2)
	for i := 0; i < len(l); i += 2 {
		inputs[i/2] = l[i].(string)
		res, err := json.Marshal(l[i+1])
		if err != nil {
			t.Fatal(err)
		}
		resJsons[i/2] = string(res)
	}
	return inputs, resJsons
}

func runTest(t *testing.T, css, resJson []string, fn func(input string) []Token) {
	for i, input := range css {
		resToTest := fn(input)
		res, err := marshalJSON(resToTest)
		if err != nil {
			t.Fatal(err)
		}
		if res != resJson[i] {
			t.Fatalf(fmt.Sprintf("input %d : \n %s \n failed : expected \n %s \n got  \n %s \n", i, input, resJson[i], res))
		}
	}
}

func runTestOneToken(t *testing.T, css, resJson []string, fn func(input string) jsonisable) {
	t.Helper()

	for i, input := range css {
		resToTest := fn(input)
		b, err := json.Marshal(resToTest.toJson())
		if err != nil {
			t.Fatal(err)
		}
		res := string(b)
		if res != resJson[i] {
			t.Fatalf(fmt.Sprintf("input %d : \n %s \n failed : expected \n %s \n got  \n %s \n", i, input, resJson[i], res))
		}
	}
}

func TestComponentValueList(t *testing.T) {
	inputs, resJson := loadJson(t, "component_value_list.json")
	runTest(t, inputs, resJson, func(s string) []Token {
		return tokenizeString(s, true)
	})
}

func TestOneComponentValue(t *testing.T) {
	inputs, resJson := loadJson(t, "one_component_value.json")
	runTestOneToken(t, inputs, resJson, func(input string) jsonisable {
		return parseOneComponentValueString(input, true)
	})
}

func TestDeclarationList(t *testing.T) {
	inputs, resJson := loadJson(t, "declaration_list.json")
	runTest(t, inputs, resJson, func(s string) []Token {
		return ParseDeclarationListString(s, true, true)
	})
}

func TestOneDeclaration(t *testing.T) {
	inputs, resJson := loadJson(t, "one_declaration.json")
	runTestOneToken(t, inputs, resJson, func(s string) jsonisable {
		return parseOneDeclarationString(s, true)
	})
}

func TestStylesheet(t *testing.T) {
	inputs, resJson := loadJson(t, "stylesheet.json")
	runTest(t, inputs, resJson, func(s string) []Token {
		return ParseStylesheetBytes([]byte(s), true, true)
	})
}

func TestRuleList(t *testing.T) {
	inputs, resJson := loadJson(t, "rule_list.json")
	runTest(t, inputs, resJson, func(s string) []Token {
		return ParseRuleListString(s, true, true)
	})
}

func TestOneRule(t *testing.T) {
	inputs, resJson := loadJson(t, "one_rule.json")
	runTestOneToken(t, inputs, resJson, func(input string) jsonisable {
		l := tokenizeString(input, true)
		return parseOneRule(l)
	})
}

func TestColor3(t *testing.T) {
	inputs, resJson := loadJson(t, "color3.json")
	runTestOneToken(t, inputs, resJson, func(input string) jsonisable {
		return ParseColorString(input)
	})
}

func parseNthString(css string) *[2]int {
	l := tokenizeString(css, true)
	return ParseNth(l)
}

func TestNth(t *testing.T) {
	inputs, resJson := loadJson(t, "An+B.json")
	runTestOneToken(t, inputs, resJson, func(s string) jsonisable {
		l := parseNthString(s)
		if l != nil {
			return jsonList{myInt(l[0]), myInt(l[1])}
		}
		var out jsonList
		return out
	})
}

func TestColor3Hsl(t *testing.T) {
	inputs, resJson := loadJson(t, "color3_hsl.json")
	runTestOneToken(t, inputs, resJson, func(input string) jsonisable {
		return ParseColorString(input)
	})
}

func TestColor3Keywords(t *testing.T) {
	inputs, resJson := loadJson(t, "color3_keywords.json")

	runTestOneToken(t, inputs, resJson, func(input string) jsonisable {
		var resToTest jsonList
		color := ParseColorString(input)
		if !color.IsNone() {
			resToTest = jsonList{myFloat(color.RGBA.R) * 255, myFloat(color.RGBA.G) * 255, myFloat(color.RGBA.B) * 255, myFloat(color.RGBA.A)}
		}
		return resToTest
	})
}

// func TestStylesheetBytes(t *testing.T) {
//     kwargs["cssBytes"] = kwargs["cssBytes"].encode("latin1")
//     kwargs.pop("comment", None)
//     if kwargs.get("environmentEncoding") {
//         kwargs["environmentEncoding"] = lookup(kwargs["environmentEncoding"])
//     } kwargs.update(SKIP)
//     return parseStylesheetBytes(**kwargs)
// }
