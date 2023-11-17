package text

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/webrender/css/properties"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
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

func TestAddFace(t *testing.T) {
	fc := NewFontConfigurationPango(fontmapPango)
	url, err := utils.PathToURL("../resources_test/weasyprint.otf")
	if err != nil {
		t.Fatal(err)
	}
	filename := fc.AddFontFace(validation.FontFaceDescriptors{
		Src:        []properties.NamedString{{Name: "external", String: url}},
		FontFamily: "weasyprint",
	}, utils.DefaultUrlFetcher)

	_, err = fc.LoadFace(fonts.FaceID{File: filename}, fontconfig.TrueType)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := os.ReadFile("../resources_test/weasyprint.otf")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, fc.FontContent(FontOrigin{File: filename})) {
		t.Fatal()
	}
}

func TestVariations(t *testing.T) {
	s := pangoFontVariations([]Variation{
		{[4]byte{'a', 'b', 'c', '0'}, 4},
		{[4]byte{'a', 'b', 'c', 'd'}, 8},
	})
	tu.AssertEqual(t, s, "abc0=4.000000,abcd=8.000000", "")
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
		utils.RoundPrec(hx, 2),
		utils.RoundPrec(w0, 2),
		utils.RoundPrec(pr.Fl(height), 2),
		utils.RoundPrec(pr.Fl(baseline), 2),
	}
}

// func TestGenerateMetrics(t *testing.T) {
// 	var descriptions []FontDescription
// 	loadJson(t, "font_descriptions.json", &descriptions)

// 	fc := &FontConfigurationPango{fontmap: fontmap}
// 	mets := make([]metrics, len(descriptions))
// 	for i, desc := range descriptions {
// 		mets[i] = newMetrics(fc, desc)
// 	}

// 	f, err := os.Create("testdata/metrics_linux.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer f.Close()
// 	enc := json.NewEncoder(f)
// 	enc.SetIndent(" ", "")
// 	err = enc.Encode(mets)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestMetrics(t *testing.T) {
	// test are generated with pango as reference
	// on a linux system
	var descriptions []FontDescription
	loadJson(t, "font_descriptions.json", &descriptions)
	var metrics []metrics
	loadJson(t, "metrics_linux.json", &metrics)

	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	for i, desc := range descriptions {
		got := newMetrics(fcPango, desc)
		tu.AssertEqual(t, got, metrics[i], fmt.Sprint(desc))

		fcGotext.heightx(&TextStyle{FontDescription: desc})
	}
}

func TestMetricsLinuxFonts(t *testing.T) {
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	fcGotext := NewFontConfigurationGotext(fontmapGotext)

	desc := FontDescription{
		Style:   FSyNormal,
		Stretch: FSeNormal,
	}

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
				got := newMetrics(fcGotext, desc)
				tu.AssertEqual(t, exp, got, "")
			}
		}
	}
}

func BenchmarkMetrics(b *testing.B) {
	var descriptions []FontDescription
	loadJson(b, "font_descriptions.json", &descriptions)

	fc := &FontConfigurationPango{fontmap: fontmapPango}
	fcGotext := NewFontConfigurationGotext(fontmapGotext)

	b.Run("Pango", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, desc := range descriptions {
				_ = newMetrics(fc, desc)
			}
		}
	})

	b.Run("go-text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, desc := range descriptions {
				_ = newMetrics(fcGotext, desc)
			}
		}
	})
}
