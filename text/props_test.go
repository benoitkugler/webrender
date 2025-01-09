package text

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestDefaultValues(t *testing.T) {
	ts := NewTextStyle(pr.InitialValues, false)
	tu.AssertEqual(t, ts.FontDescription.Family, []string{"serif"})
	tu.AssertEqual(t, ts.FontDescription.Style, FSyNormal)
	tu.AssertEqual(t, ts.FontDescription.Stretch, FSeNormal)
	tu.AssertEqual(t, ts.FontDescription.Weight, uint16(400))
	tu.AssertEqual(t, ts.FontDescription.Size, pr.Fl(16))
	tu.AssertEqual(t, ts.FontDescription.VariationSettings, []Variation(nil))

	tu.AssertEqual(t, ts.FontLanguageOverride, fontLanguageOverride{})
	tu.AssertEqual(t, ts.Lang, "")
	tu.AssertEqual(t, ts.TextDecorationLine, pr.Decorations{})
	tu.AssertEqual(t, ts.WhiteSpace, WNormal)
	tu.AssertEqual(t, ts.LetterSpacing, pr.Fl(0))
	tu.AssertEqual(t, ts.FontFeatures, []Feature(nil))
}

func TestCollectStyles(t *testing.T) {
	t.Skip()

	f, err := os.Open("../html/document/styles.json")
	if err != nil {
		t.Fatal(err)
	}
	var styles []TextStyle
	err = json.NewDecoder(f).Decode(&styles)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	fmt.Println(len(styles))
	m := map[string]FontDescription{}
	for _, sty := range styles {
		m[string(sty.FontDescription.hash(true))] = sty.FontDescription
	}
	var desc []FontDescription
	for _, fd := range m {
		desc = append(desc, fd)
	}
	sort.Slice(desc, func(i, j int) bool { return string(desc[i].hash(true)) < string(desc[j].hash(true)) })
	f, err = os.Create("testdata/font_descriptions.json")
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent(" ", " ")
	err = enc.Encode(desc)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}
