package validation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/css/parser"
	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func processFontFace(css string, t *testing.T) FontFaceDescriptors {
	t.Helper()

	stylesheet := parser.ParseStylesheetBytes([]byte(css), false, false)
	atRule, ok := stylesheet[0].(parser.AtRule)
	if !ok || atRule.AtKeyword != "font-face" {
		t.Fatalf("expected @font-face got %v", stylesheet[0])
	}
	tokens := parser.ParseDeclarationList(*atRule.Content, false, false)
	return PreprocessFontFaceDescriptors("https://weasyprint.org/foo/", tokens)
}

func checkNameDescriptor(ref, got interface{}, t *testing.T) {
	t.Helper()

	if !reflect.DeepEqual(ref, got) {
		t.Fatalf("expected %v got %v", ref, got)
	}
}

// Test the “font-face“ rule.
func TestFontFace(t *testing.T) {
	l := processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: url(https://example.com/fonts/Gentium.woff);
	  }`, t)
	checkNameDescriptor(pr.String("Gentium Hard"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "external", String: "https://example.com/fonts/Gentium.woff"}}, l.Src, t)

	l = processFontFace(`@font-face {
          font-family: "Fonty Smiley";
         src: url(Fonty-Smiley.woff);
          font-style: italic;
         font-weight: 200;
         font-stretch: condensed;
        }`, t)
	checkNameDescriptor(pr.String("Fonty Smiley"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "external", String: "https://weasyprint.org/foo/Fonty-Smiley.woff"}}, l.Src, t)
	checkNameDescriptor(pr.String("italic"), l.FontStyle, t)
	checkNameDescriptor(pr.IntString{Int: 200}, l.FontWeight, t)
	checkNameDescriptor(pr.String("condensed"), l.FontStretch, t)

	l = processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: local();
        }`, t)
	checkNameDescriptor(pr.String("Gentium Hard"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "local", String: ""}}, l.Src, t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/487
	l = processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: local(Gentium Hard);
        }`, t)
	checkNameDescriptor(pr.String("Gentium Hard"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "local", String: "Gentium Hard"}}, l.Src, t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1653
	capt := tu.CaptureLogs()
	l = processFontFace(`@font-face {
          font-family: Gentium Hard;
          src: local(Gentium Hard);
          src: local(Gentium Soft),
        }`, t)
	logs := capt.Logs()
	tu.AssertEqual(t, len(logs), 1)
	tu.AssertEqual(t, strings.Contains(logs[0], "invalid or unsupported values"), true)

	checkNameDescriptor(pr.String("Gentium Hard"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "local", String: "Gentium Hard"}}, l.Src, t)
}

// Test bad “font-face“ rules.
func TestBadFontFace(t *testing.T) {
	logs := tu.CaptureLogs()
	l := processFontFace(`@font-face {`+
		`  font-family: "Bad Font";`+
		`  src: url(BadFont.woff);`+
		`  font-stretch: expanded;`+
		`  font-style: wrong;`+
		`  font-weight: bolder;`+
		`  font-stretch: wrong;`+
		`}`, t)

	checkNameDescriptor(pr.String("Bad Font"), l.FontFamily, t)
	checkNameDescriptor([]pr.NamedString{{Name: "external", String: "https://weasyprint.org/foo/BadFont.woff"}}, l.Src, t)
	checkNameDescriptor(pr.String("expanded"), l.FontStretch, t)

	logs.CheckEqual([]string{
		"Ignored `font-style: wrong` at 1:91, unsupported font-style descriptor: wrong.",
		"Ignored `font-weight: bolder` at 1:111, invalid or unsupported values for a known CSS property.",
		"Ignored `font-stretch: wrong` at 1:133, unsupported font-stretch descriptor: wrong.",
	}, t)
}

// see style/style_test.go for other font face tests
