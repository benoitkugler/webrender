package properties

import (
	"strings"

	"github.com/benoitkugler/webrender/utils"

	"github.com/benoitkugler/webrender/css/parser"
)

const ( // zero field corresponds to null content
	Scalar Unit = iota + 1 // means no unit, but a valid value
	Perc                   // percentage (%)
	Ex
	Em
	Ch
	Rem
	Px
	Pt
	Pc
	In
	Cm
	Mm
	Q

	Rad
	Turn
	Deg
	Grad
)

var (
	ZeroPixels      = Dimension{Unit: Px}
	zeroPixelsValue = ZeroPixels.ToValue()

	CurrentColor = Color{Type: parser.ColorCurrentColor}
	// How many CSS pixels is one <unit>?
	// http://www.w3.org/TR/CSS21/syndata.html#length-units
	LengthsToPixels = map[Unit]Float{
		Px: 1,
		Pt: 1. / 0.75,
		Pc: 16.,             // LengthsToPixels["pt"] * 12
		In: 96.,             // LengthsToPixels["pt"] * 72
		Cm: 96. / 2.54,      // LengthsToPixels["in"] / 2.54
		Mm: 96. / 25.4,      // LengthsToPixels["in"] / 25.4
		Q:  96. / 25.4 / 4., // LengthsToPixels[Mm] / 4
	}

	// Value in pixels of font-size for <absolute-size> keywords: 12pt (16px) for
	// medium, and scaling factors given in CSS3 for others:
	// http://www.w3.org/TR/css3-fonts/#font-size-prop
	FontSizeKeywords = map[string]Float{ // medium is 16px, others are a ratio of medium
		"xx-small": InitialValues.GetFontSize().Value * 3 / 5,
		"x-small":  InitialValues.GetFontSize().Value * 3 / 4,
		"small":    InitialValues.GetFontSize().Value * 8 / 9,
		"medium":   InitialValues.GetFontSize().Value * 1 / 1,
		"large":    InitialValues.GetFontSize().Value * 6 / 5,
		"x-large":  InitialValues.GetFontSize().Value * 3 / 2,
		"xx-large": InitialValues.GetFontSize().Value * 2 / 1,
	}
	FontSizeKeywordsOrder = []string{"xx-small", "x-small", "small", "medium", "large", "x-large", "xx-large"}

	// http://www.w3.org/TR/css3-page/#size
	PageSizes = map[string]Point{
		"a10":     {Dimension{Value: 26, Unit: Mm}, Dimension{Value: 37, Unit: Mm}},
		"a9":      {Dimension{Value: 37, Unit: Mm}, Dimension{Value: 52, Unit: Mm}},
		"a8":      {Dimension{Value: 52, Unit: Mm}, Dimension{Value: 74, Unit: Mm}},
		"a7":      {Dimension{Value: 74, Unit: Mm}, Dimension{Value: 105, Unit: Mm}},
		"a6":      {Dimension{Value: 105, Unit: Mm}, Dimension{Value: 148, Unit: Mm}},
		"a5":      {Dimension{Value: 148, Unit: Mm}, Dimension{Value: 210, Unit: Mm}},
		"a4":      A4,
		"a3":      {Dimension{Value: 297, Unit: Mm}, Dimension{Value: 420, Unit: Mm}},
		"a2":      {Dimension{Value: 420, Unit: Mm}, Dimension{Value: 594, Unit: Mm}},
		"a1":      {Dimension{Value: 594, Unit: Mm}, Dimension{Value: 841, Unit: Mm}},
		"a0":      {Dimension{Value: 841, Unit: Mm}, Dimension{Value: 1189, Unit: Mm}},
		"b10":     {Dimension{Value: 31, Unit: Mm}, Dimension{Value: 44, Unit: Mm}},
		"b9":      {Dimension{Value: 44, Unit: Mm}, Dimension{Value: 62, Unit: Mm}},
		"b8":      {Dimension{Value: 62, Unit: Mm}, Dimension{Value: 88, Unit: Mm}},
		"b7":      {Dimension{Value: 88, Unit: Mm}, Dimension{Value: 125, Unit: Mm}},
		"b6":      {Dimension{Value: 125, Unit: Mm}, Dimension{Value: 176, Unit: Mm}},
		"b5":      {Dimension{Value: 176, Unit: Mm}, Dimension{Value: 250, Unit: Mm}},
		"b4":      {Dimension{Value: 250, Unit: Mm}, Dimension{Value: 353, Unit: Mm}},
		"b3":      {Dimension{Value: 353, Unit: Mm}, Dimension{Value: 500, Unit: Mm}},
		"b2":      {Dimension{Value: 500, Unit: Mm}, Dimension{Value: 707, Unit: Mm}},
		"b1":      {Dimension{Value: 707, Unit: Mm}, Dimension{Value: 1000, Unit: Mm}},
		"b0":      {Dimension{Value: 1000, Unit: Mm}, Dimension{Value: 1414, Unit: Mm}},
		"c10":     {Dimension{Value: 28, Unit: Mm}, Dimension{Value: 40, Unit: Mm}},
		"c9":      {Dimension{Value: 40, Unit: Mm}, Dimension{Value: 57, Unit: Mm}},
		"c8":      {Dimension{Value: 57, Unit: Mm}, Dimension{Value: 81, Unit: Mm}},
		"c7":      {Dimension{Value: 81, Unit: Mm}, Dimension{Value: 114, Unit: Mm}},
		"c6":      {Dimension{Value: 114, Unit: Mm}, Dimension{Value: 162, Unit: Mm}},
		"c5":      {Dimension{Value: 162, Unit: Mm}, Dimension{Value: 229, Unit: Mm}},
		"c4":      {Dimension{Value: 229, Unit: Mm}, Dimension{Value: 324, Unit: Mm}},
		"c3":      {Dimension{Value: 324, Unit: Mm}, Dimension{Value: 458, Unit: Mm}},
		"c2":      {Dimension{Value: 458, Unit: Mm}, Dimension{Value: 648, Unit: Mm}},
		"c1":      {Dimension{Value: 648, Unit: Mm}, Dimension{Value: 917, Unit: Mm}},
		"c0":      {Dimension{Value: 917, Unit: Mm}, Dimension{Value: 1297, Unit: Mm}},
		"jis-b10": {Dimension{Value: 32, Unit: Mm}, Dimension{Value: 45, Unit: Mm}},
		"jis-b9":  {Dimension{Value: 45, Unit: Mm}, Dimension{Value: 64, Unit: Mm}},
		"jis-b8":  {Dimension{Value: 64, Unit: Mm}, Dimension{Value: 91, Unit: Mm}},
		"jis-b7":  {Dimension{Value: 91, Unit: Mm}, Dimension{Value: 128, Unit: Mm}},
		"jis-b6":  {Dimension{Value: 128, Unit: Mm}, Dimension{Value: 182, Unit: Mm}},
		"jis-b5":  {Dimension{Value: 182, Unit: Mm}, Dimension{Value: 257, Unit: Mm}},
		"jis-b4":  {Dimension{Value: 257, Unit: Mm}, Dimension{Value: 364, Unit: Mm}},
		"jis-b3":  {Dimension{Value: 364, Unit: Mm}, Dimension{Value: 515, Unit: Mm}},
		"jis-b2":  {Dimension{Value: 515, Unit: Mm}, Dimension{Value: 728, Unit: Mm}},
		"jis-b1":  {Dimension{Value: 728, Unit: Mm}, Dimension{Value: 1030, Unit: Mm}},
		"jis-b0":  {Dimension{Value: 1030, Unit: Mm}, Dimension{Value: 1456, Unit: Mm}},
		"letter":  {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 11, Unit: In}},
		"legal":   {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 14, Unit: In}},
		"ledger":  {Dimension{Value: 11, Unit: In}, Dimension{Value: 17, Unit: In}},
	}

	A4 = Point{Dimension{Value: 210, Unit: Mm}, Dimension{Value: 297, Unit: Mm}}

	KnownProperties = utils.Set{}

	// Do not list shorthand properties here as we handle them before inheritance.
	//
	// Values inherited but not applicable to print are not included.
	//
	// text_decoration is not a really inherited, see
	// http://www.w3.org/TR/CSS2/text.html#propdef-text-decoration
	//
	// link: click events normally bubble up to link ancestors
	// See http://lists.w3.org/Archives/Public/www-style/2012Jun/0315.html
	Inherited = utils.NewSet(
		"border_collapse",
		"border_spacing",
		"caption_side",
		"color",
		"direction",
		"empty_cells",
		"font_family",
		"font_feature_settings",
		"font_kerning",
		"font_language_override",
		"font_size",
		"font_style",
		"font_stretch",
		"font_variant",
		"font_variant_alternates",
		"font_variant_caps",
		"font_variant_east_asian",
		"font_variant_ligatures",
		"font_variant_numeric",
		"font_variant_position",
		"font_variation_settings",
		"font_weight",
		"hyphens",
		"hyphenate_character",
		"hyphenate_limit_chars",
		"hyphenate_limit_zone",
		"image_rendering",
		"image_resolution",
		"lang",
		"letter_spacing",
		"line_height",
		"link",
		"list_style_image",
		"list_style_position",
		"list_style_type",
		"orphans",
		"overflow_wrap",
		"quotes",
		"tab_size",
		"text_align_all",
		"text_align_last",
		"text_indent",
		"text_transform",
		"visibility",
		"white_space",
		"widows",
		"word_spacing",
		"word_break",
	)

	// Not applicable to the print media
	NotPrintMedia = utils.NewSet(
		// Aural media
		"azimuth",
		"cue",
		"cue-after",
		"cue-before",
		"elevation",
		"pause",
		"pause-after",
		"pause-before",
		"pitch-range",
		"pitch",
		"play-during",
		"richness",
		"speak-header",
		"speak-numeral",
		"speak-punctuation",
		"speak",
		"speech-rate",
		"stress",
		"voice-family",
		"volume",
		// Interactive
		"cursor",
		// Animations and transitions
		"animation",
		"animation-delay",
		"animation-direction",
		"animation-duration",
		"animation-fill-mode",
		"animation-iteration-count",
		"animation-name",
		"animation-play-state",
		"animation-timing-function",
		"transition",
		"transition-delay",
		"transition-duration",
		"transition-property",
		"transition-timing-function",
	)

	// http://www.w3.org/TR/CSS21/tables.html#model
	// See also http://lists.w3.org/Archives/Public/www-style/2012Jun/0066.html
	// Only non-inherited properties need to be included here.
	TableWrapperBoxProperties = utils.NewSet(
		"bottom",
		"break_after",
		"break_before",
		"break_inside",
		"clear",
		"counter_increment",
		"counter_reset",
		"counter_set",
		"float",
		"left",
		"margin_top",
		"margin_bottom",
		"margin_left",
		"margin_right",
		"opacity",
		"overflow",
		"position",
		"right",
		"top",
		"transform",
		"transform_origin",
		"vertical_align",
		"z_index",
	)

	InitialNotComputed = utils.NewSet(
		"display",
		"column_gap",
		"bleed_top",
		"bleed_left",
		"bleed_bottom",
		"bleed_right",
		"outline_width",
		"outline_color",
		"column_rule_width",
		"column_rule_color",
		"border_top_width",
		"border_left_width",
		"border_bottom_width",
		"border_right_width",
		"border_top_color",
		"border_left_color",
		"border_bottom_color",
		"border_right_color",
	)
)

func init() {
	for name := range InitialValues {
		KnownProperties.Add(strings.ReplaceAll(name, "_", "-"))
	}
}
