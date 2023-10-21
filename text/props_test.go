package text

import (
	"testing"

	pr "github.com/benoitkugler/webrender/css/properties"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

func TestDefaultValues(t *testing.T) {
	ts := NewTextStyle(pr.InitialValues)
	tu.AssertEqual(t, ts.FontFamily, []string{"serif"}, "")
	tu.AssertEqual(t, ts.FontStyle, FSyNormal, "")
	tu.AssertEqual(t, ts.FontStretch, FSeNormal, "")
	tu.AssertEqual(t, ts.FontWeight, 400, "")
	tu.AssertEqual(t, ts.FontSize, pr.Fl(16), "")
	tu.AssertEqual(t, ts.FontVariationSettings, []Variation(nil), "")

	tu.AssertEqual(t, ts.FontLanguageOverride, FontLanguageOverride{}, "")
	tu.AssertEqual(t, ts.Lang, "", "")
	tu.AssertEqual(t, ts.TextDecorationLine, pr.Decorations{}, "")
	tu.AssertEqual(t, ts.WhiteSpace, WNormal, "")
	tu.AssertEqual(t, ts.LetterSpacing, pr.Fl(0), "")
	tu.AssertEqual(t, ts.FontFeatures, []Feature(nil), "")
}
