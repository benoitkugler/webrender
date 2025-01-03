package text

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/webrender/css/properties"
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
	fc := NewFontConfigurationPango(fontmap)
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
	tu.AssertEqual(t, s, "abc0=4.000000,abcd=8.000000")
}
