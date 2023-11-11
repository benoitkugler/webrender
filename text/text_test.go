package text

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

var (
	sansFonts = pr.Strings{"DejaVu Sans", "sans"}
	monoFonts = pr.Strings{"DejaVu Sans Mono", "monospace"}
)

const fontmapCache = "testdata/cache.fc"

var fontmap *fcfonts.FontMap

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fontconfig.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	panic(err)
	// }

	fs, err := fontconfig.LoadFontsetFile(fontmapCache)
	if err != nil {
		panic(err)
	}
	fontmap = fcfonts.NewFontMap(fontconfig.Standard, fs)
}

func assert(t *testing.T, b bool, msg string) {
	if !b {
		t.Fatal(msg)
	}
}

type textContext struct {
	fontmap *fcfonts.FontMap
	dict    map[HyphenDictKey]hyphen.Hyphener
}

func (tc textContext) Fonts() FontConfiguration {
	return &FontConfigurationPango{fontmap: tc.fontmap}
}
func (tc textContext) HyphenCache() map[HyphenDictKey]hyphen.Hyphener { return tc.dict }
func (tc textContext) StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float {
	return make(map[StrutLayoutKey][2]pr.Float)
}

// Wrapper for SplitFirstLine() creating a style dict.
func makeText(text string, width pr.MaybeFloat, style pr.Properties) Splitted {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(style)
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}
	return SplitFirstLine(text, newStyle, ct, width, false, true)
}

func TestLineContent(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	for _, v := range []struct {
		remaining string
		width     pr.Float
	}{
		{"text for test", 100},
		{"is a text for test", 45},
	} {
		text := "This is a text for test"
		sp := makeText(text, v.width, pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
		textRunes := []rune(text)
		assert(t, string(textRunes[sp.ResumeAt:]) == v.remaining, "unexpected remaining")
		assert(t, sp.Length+1 == sp.ResumeAt, fmt.Sprintf("%v: expected %d, got %d", v.width, sp.ResumeAt, sp.Length+1)) // +1 for the removed trailing space
	}
}

func TestLineWithAnyWidth(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	sp1 := makeText("some text", nil, nil)
	sp2 := makeText("some text some text", nil, nil)
	assert(t, sp1.Width < sp2.Width, "unexpected width")
}

func TestLineBreaking(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "Thïs is a text for test"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "is a text for test", "")

	sp = makeText(str, pr.Float(100), pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "text for test", "")
}

func TestLineBreakingRTL(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "لوريم ايبسوم دولا"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "ايبسوم دولا", "")
}

func TestTextDimension(t *testing.T) {
	cl := tu.CaptureLogs()
	defer cl.AssertNoLogs(t)

	str := "This is a text for test. This is a test for text.py"
	sp1 := makeText(str, pr.Float(200), pr.Properties{pr.PFontSize: pr.FToV(12)})
	sp2 := makeText(str, pr.Float(200), pr.Properties{pr.PFontSize: pr.FToV(20)})
	assert(t, sp1.Width*sp1.Height < sp2.Width*sp2.Height, "")
}

func BenchmarkSplitFirstLine(b *testing.B) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}

	text := "This is a text for test. This is a test for text.py"
	for i := 0; i < b.N; i++ {
		SplitFirstLine(text, newStyle, ct, pr.Float(200), false, true)
	}
}

func TestGetLastWordEnd(t *testing.T) {
	fc := &FontConfigurationPango{fontmap: fontmap}
	if i := GetLastWordEnd(fc, []rune{99, 99, 32, 99}); i != 2 {
		t.Fatalf("expected %d, got %d", 2, i)
	}
}

func TestHeightAndBaseline(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	families := pr.Strings{
		"Helvetica",
		"Apple Color Emoji",
	}
	newStyle.SetFontFamily(families)

	newStyle.SetFontSize(pr.FToV(36))
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}

	fc := NewFontConfigurationPango(fontmap)
	for _, desc := range []validation.FontFaceDescriptors{
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/googlesans/v36/4UaGrENHsxJlGDuGo1OIlL3Owps.ttf"}}, FontFamily: "Google Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/googlesans/v36/4UabrENHsxJlGDuGo1OIlLU94YtzCwM.ttf"}}, FontFamily: "Google Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/materialicons/v117/flUhRq6tzZclQEJ-Vdg-IuiaDsNZ.ttf"}}, FontFamily: "Material Icons", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/opensans/v27/memSYaGs126MiZpBA-UvWbX2vVnXBbObj2OVZyOOSr4dVJWUgsjZ0B4gaVc.ttf"}}, FontFamily: "Open Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}, FontStretch: "normal"},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/roboto/v29/KFOmCnqEu92Fr1Mu4mxP.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/roboto/v29/KFOlCnqEu92Fr1MmEU9fBBc9.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/roboto/v29/KFOlCnqEu92Fr1MmWUlfBBc9.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 700}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K0nXBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K3vXBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.NamedString{{Name: "external", String: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K5fQBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 600}},
	} {
		fc.AddFontFace(desc, utils.DefaultUrlFetcher)
	}

	spi := SplitFirstLine("Go 1.17 Release Notes", newStyle, ct, pr.Float(595), false, true)
	height, baseline := spi.Height, spi.Baseline

	if int((height-43)/10) != 0 {
		t.Fatalf("unexpected height %f", height)
	}
	if int((baseline-33)/10) != 0 {
		t.Fatalf("unexpected baseline %f", baseline)
	}
}

func newContextWithWeasyFont(t *testing.T) textContext {
	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}
	fc := NewFontConfigurationPango(fontmap)
	url, err := utils.PathToURL("../resources_test/weasyprint.otf")
	if err != nil {
		t.Fatal(err)
	}
	fc.AddFontFace(validation.FontFaceDescriptors{
		Src:        []pr.NamedString{{Name: "external", String: url}},
		FontFamily: "weasyprint",
	}, utils.DefaultUrlFetcher)
	return ct
}

func TestLayoutFirstLine(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(pr.Strings{"weasyprint"})
	newStyle.SetFontSize(pr.FToV(16))
	newStyle.SetWhiteSpace("normal")

	ct := newContextWithWeasyFont(t)

	layout := createLayout("a a ", NewTextStyle(newStyle, false), ct.Fonts(), pr.Float(63))
	_, index := layout.GetFirstLine()
	if index != -1 {
		t.Fatalf("unexpected first line index: %d", index)
	}
}

// func TestChWidth(t *testing.T) {
// 	newStyle := pr.InitialValues.Copy()
// 	newStyle.SetFontFamily(pr.Strings{"arial"})
// 	newStyle.SetFontSize(pr.FToV(16))
// 	//  pr.FToV(-0.04444)
// 	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}
// 	if w := CharacterRatio(dummyStyle{newStyle}, pr.NewTextRatioCache(), true, ct); utils.RoundPrec(pr.Fl(w), 3) != 8.854 {
// 		t.Fatalf("unexpected ch width %v", w)
// 	}
// }

func TestSplitFirstLine(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(pr.Strings{"arial"})
	newStyle.SetFontSize(pr.FToV(16))

	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}

	out := SplitFirstLine(" of the element's ", newStyle, ct, pr.Float(120.18628), false, true)

	if out.ResumeAt != -1 {
		t.Fatalf("unexpected resume index %d", out.ResumeAt)
	}
}

func TestCanBreakText(t *testing.T) {
	tests := []struct {
		s    string
		want pr.MaybeBool
	}{
		{" s", pr.True},
		{"\u00a0L", pr.False},
		{"\u00a0d", pr.False},
		{"r\u00a0", pr.False},
		{" “", pr.True},
		{"” ", pr.False},
		{"t\u00a0", pr.False},
		{"\u00a0L", pr.False},
		{"\u00a0d", pr.False},
		{"r\u00a0", pr.False},
		{" “", pr.True},
		{"” ", pr.False},
		{"t\u00a0", pr.False},
		{"a⺀", pr.True},
		{"⺀b", pr.True},
		{"bc", pr.False},
		{"a⺀", pr.True},
		{"⺀b", pr.True},
		{"bc", pr.False},
		{"", nil},
		{"c ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"a ", pr.False},
		{"a", nil},
		{"a ", pr.False},
		{"a", nil},
		{"⺀ ", pr.False},
		{"a", nil},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"a", nil},
		{"a", nil},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"c\u00a0", pr.False},
		{"i", nil},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"ii", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0a", pr.False},
		{" a", pr.True},
		{"\u00a0 ", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"c\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"c\u00a0", pr.False},
		{"\u200f\u00a0i", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u200f\u00a0i", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"\u00a0\u200f", pr.False},
		{"\u200f\u00a0ii", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u200f\u00a0ii", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"bb", pr.False},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"c", nil},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"a", nil},
		{"abc", pr.False},
		{"abcde", pr.False},
		{"abcde", pr.False},
		{"[initial]", pr.False},
		{"[]", pr.False},
		{"o", nil},
		{"abcde", pr.False},
		{"ab", pr.False},
		{"cd", pr.False},
		{"bc", pr.False},
		{"b", nil},
		{"a", nil},
		{"e", nil},
		{"de", pr.False},
		{"a", nil},
		{"b", nil},
		{"cd", pr.False},
		{"abcde", pr.False},
		{"ace", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" 4", pr.True},
		{"4 ", pr.False},
		{"  ", pr.False},
		{" h", pr.True},
		{" i", pr.True},
		{"z ", pr.False},
		{" a", pr.True},
		{"a ", pr.False},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"t ", pr.False},
		{" A", pr.True},
		{"t ", pr.False},
		{"test", pr.False},
	}
	fc := &FontConfigurationPango{fontmap: fontmap}
	for _, tt := range tests {
		if got := CanBreakText(fc, []rune(tt.s)); got != tt.want {
			t.Errorf("CanBreakText() = %v, want %v", got, tt.want)
		}
	}
}
