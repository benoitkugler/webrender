package text

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textprocessing/fontconfig"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/fontscan"
)

func TestAddConfig(t *testing.T) {
	fontFilename := "dummy"
	fontFamily := "arial"
	fontconfigStyle := "roman"
	fontconfigWeight := "regular"
	fontconfigStretch := "normal"
	featuresSttring := ""
	xml := fmt.Sprintf(`<?xml version="1.0"?>
			<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
			<fontconfig>
			  <match target="scan">
				<test name="file" compare="eq">
				  <string>%s</string>
				</test>
				<edit name="family" mode="assign_replace">
				  <string>%s</string>
				</edit>
				<edit name="slant" mode="assign_replace">
				  <const>%s</const>
				</edit>
				<edit name="weight" mode="assign_replace">
				  <const>%s</const>
				</edit>
				<edit name="width" mode="assign_replace">
				  <const>%s</const>
				</edit>
			  </match>
			  <match target="font">
				<test name="file" compare="eq">
				  <string>%s</string>
				</test>
				<edit name="fontfeatures" mode="assign_replace">%s</edit>
			  </match>
			</fontconfig>`, fontFilename, fontFamily, fontconfigStyle,
		fontconfigWeight, fontconfigStretch, fontFilename, featuresSttring)

	config := fontconfig.Standard.Copy()
	err := config.LoadFromMemory(bytes.NewReader([]byte(xml)))
	if err != nil {
		t.Fatalf("Failed to load fontconfig config: %s", err)
	}
}

func TestAddFontFace(t *testing.T) {
	fcP := NewFontConfigurationPango(fontmapPango)
	fcG := NewFontConfigurationGotext(fontmapGotext)

	url, err := utils.PathToURL("../resources_test/weasyprint.otf")
	tu.AssertNoErr(t, err)
	desc := validation.FontFaceDescriptors{
		Src:        []pr.TaggedString{{Tag: pr.External, S: url}},
		FontFamily: "weasyprint",
	}

	expected, err := os.ReadFile("../resources_test/weasyprint.otf")
	tu.AssertNoErr(t, err)

	// Pango
	filename := fcP.AddFontFace(desc, utils.DefaultUrlFetcher)
	_, err = fcP.LoadFace(fonts.FaceID{File: filename}, fontconfig.TrueType)
	tu.AssertNoErr(t, err)
	if !bytes.Equal(expected, fcP.FontContent(FontOrigin{File: filename})) {
		t.Fatal()
	}

	// Gotext
	filename2 := fcG.AddFontFace(desc, utils.DefaultUrlFetcher)
	if !bytes.Equal(expected, fcG.FontContent(FontOrigin{File: filename2})) {
		t.Fatal()
	}
	face := fcG.resolveFace('a', FontDescription{Family: []string{"weasyprint"}})
	tu.AssertEqual(t, face != nil, true)
	tu.AssertEqual(t, len(fcG.fontsFeatures[face.Font]), 1)
	tu.AssertEqual(t, fcG.fontsFeatures[face.Font][0].String(), "'kern'=1")
}

func TestAddFontFaceAspect(t *testing.T) {
	fcG := NewFontConfigurationGotext(fontmapGotext)

	url, err := utils.PathToURL("../resources_test/weasyprint.otf")
	if err != nil {
		t.Fatal(err)
	}

	desc := validation.FontFaceDescriptors{
		Src:        []pr.TaggedString{{Tag: pr.External, S: url}},
		FontFamily: "weasyprint",
		// provide user metadata
		FontStyle:   "italic",
		FontWeight:  pr.IntString{String: "bold"},
		FontStretch: "condensed",
	}

	_ = fcG.AddFontFace(desc, utils.DefaultUrlFetcher)
	face := fcG.resolveFace('a', FontDescription{Family: []string{"weasyprint"}, Style: FSty_Italic, Weight: 700, Stretch: FStr_Condensed})
	family, aspect := fcG.fm.FontMetadata(face.Font)
	tu.AssertEqual(t, family, "weasyprint")
	tu.AssertEqual(t, aspect, font.Aspect{
		Style:   font.StyleItalic,
		Weight:  font.WeightBold,
		Stretch: font.StretchCondensed,
	})
}

func TestVariations(t *testing.T) {
	s := pangoFontVariations([]Variation{
		{[4]byte{'a', 'b', 'c', '0'}, 4},
		{[4]byte{'a', 'b', 'c', 'd'}, 8},
	})
	tu.AssertEqual(t, s, "abc0=4.000000,abcd=8.000000")
}

func loadJson(t testing.TB, file string, out interface{}) {
	f, err := os.Open(filepath.Join("testdata", file))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(out)
	if err != nil {
		t.Fatal(err)
	}
}

// we round to 2 digits and multiply by 100
type metrics struct {
	Heightx          pr.Fl
	Width0           pr.Fl
	Height, Baseline pr.Fl
}

func newMetrics(fc FontConfiguration, desc FontDescription) metrics {
	style := &TextStyle{FontDescription: desc}
	hx := fc.heightx(style)
	w0 := fc.width0(style)
	height, baseline := fc.spaceHeight(style)
	return metrics{
		utils.RoundPrec(hx, 1),
		utils.RoundPrec(w0, 1),
		utils.RoundPrec(pr.Fl(height), 1),
		utils.RoundPrec(pr.Fl(baseline), 1),
	}
}

type descriptionAndMetrics struct {
	Description FontDescription
	Metrics     metrics
}

func TestGenerateGoldenMetrics(t *testing.T) {
	t.Skip("dev only")

	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	desc := FontDescription{
		Style:   FSty_Normal,
		Stretch: FStr_Normal,
	}

	var out []descriptionAndMetrics
	// we assume we have the following fonts
	//	- urw-base35/NimbusSans-Regular.otf
	//	- urw-base35/NimbusRoman-Regular.otf
	// 	- dejavu/DejaVuSans.ttf
	// 	- liberation2/LiberationMono-Regular.ttf
	//  - croscore/Arimo-Regular.ttf
	for _, family := range []string{"Nimbus Sans", "Nimbus Roman", "DejaVu Sans", "Liberation Mono", "Arimo"} {
		for _, w := range []uint16{400, 700} { // weights
			for _, s := range []pr.Fl{12, 13, 16, 18, 32, 33} { // sizes
				desc.Family = []string{family}
				desc.Weight = w
				desc.Size = s * 10 // remove some pesky rounding errors
				exp := newMetrics(fcPango, desc)
				out = append(out, descriptionAndMetrics{desc, exp})
			}
		}
	}
	f, err := os.Create("testdata/descriptions_metrics_linux.json")
	tu.AssertNoErr(t, err)
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent(" ", "")
	err = enc.Encode(out)
	tu.AssertNoErr(t, err)
}

func TestMetricsLinuxFonts(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux is required")
	}

	fcGotext := NewFontConfigurationGotext(fontmapGotext)

	var descriptions []descriptionAndMetrics
	loadJson(t, "descriptions_metrics_linux.json", &descriptions)

	for _, test := range descriptions {
		got := newMetrics(fcGotext, test.Description)
		tu.AssertEqual(t, got, test.Metrics)
	}
}

func TestResolveFont(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip()
	}

	for _, test := range []struct {
		query    []string
		resolved string
	}{
		{[]string{"Helvetica"}, "Nimbus Sans"},
		{[]string{"BlinkMacSystemFont", "Helvetica"}, "Nimbus Sans"},
		{[]string{"Times"}, "Nimbus Roman"},
		{[]string{"Mononoki"}, "Noto Sans"},
		{[]string{"serif"}, "Noto Serif"},
	} {
		fm := NewFontConfigurationGotext(fontmapGotext).fm
		fm.SetQuery(fontscan.Query{Families: test.query})
		face := fm.ResolveFace('a')
		tu.AssertEqual(t, face.Font.Describe().Family, test.resolved)
	}
}

func Test_heightx(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)

	addWeayprintFont(t, fcGotext)

	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"weasyprint"},
		Style:   FSty_Normal,
		Stretch: FStr_Normal,
		Weight:  400,
		Size:    100,
	}}
	h := fcGotext.heightx(style)
	assertApprox(t, pr.Float(h), 79.98, "")

	style.Size = 1000
	h = fcGotext.heightx(style)
	assertApprox(t, pr.Float(h), 799.8, "")
}

func BenchmarkMetrics(b *testing.B) {
	var descriptions []descriptionAndMetrics
	loadJson(b, "descriptions_metrics_linux.json", &descriptions)

	fc := &FontConfigurationPango{fontmap: fontmapPango}
	fcGotext := NewFontConfigurationGotext(fontmapGotext)

	b.Run("Pango", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, desc := range descriptions {
				_ = newMetrics(fc, desc.Description)
			}
		}
	})

	b.Run("go-text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, desc := range descriptions {
				_ = newMetrics(fcGotext, desc.Description)
			}
		}
	})
}
