package properties

import (
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

	KnownProperties = NewSetK()

	// Do not list shorthand properties here as we handle them before inheritance.
	//
	// Values inherited but not applicable to print are not included.
	//
	// text_decoration is not a really inherited, see
	// http://www.w3.org/TR/CSS2/text.html#propdef-text-decoration
	//
	// link: click events normally bubble up to link ancestors
	// See http://lists.w3.org/Archives/Public/www-style/2012Jun/0315.html
	Inherited = NewSetK(
		PBorderCollapse,
		PBorderSpacing,
		PCaptionSide,
		PColor,
		PDirection,
		PEmptyCells,
		PFontFamily,
		PFontFeatureSettings,
		PFontKerning,
		PFontLanguageOverride,
		PFontSize,
		PFontStyle,
		PFontStretch,
		PFontVariant,
		PFontVariantAlternates,
		PFontVariantCaps,
		PFontVariantEastAsian,
		PFontVariantLigatures,
		PFontVariantNumeric,
		PFontVariantPosition,
		PFontVariationSettings,
		PFontWeight,
		PHyphens,
		PHyphenateCharacter,
		PHyphenateLimitChars,
		PHyphenateLimitZone,
		PImageRendering,
		PImageResolution,
		PLang,
		PLetterSpacing,
		PLineHeight,
		PLink,
		PListStyleImage,
		PListStylePosition,
		PListStyleType,
		POrphans,
		POverflowWrap,
		PQuotes,
		PTabSize,
		PTextAlignAll,
		PTextAlignLast,
		PTextIndent,
		PTextTransform,
		PVisibility,
		PWhiteSpace,
		PWidows,
		PWordSpacing,
		PWordBreak,
	)

	// http://www.w3.org/TR/CSS21/tables.html#model
	// See also http://lists.w3.org/Archives/Public/www-style/2012Jun/0066.html
	// Only non-inherited properties need to be included here.
	TableWrapperBoxProperties = NewSetK(
		PBottom,
		PBreakAfter,
		PBreakBefore,
		PBreakInside,
		PClear,
		PCounterIncrement,
		PCounterReset,
		PCounterSet,
		PFloat,
		PLeft,
		PMarginTop,
		PMarginBottom,
		PMarginLeft,
		PMarginRight,
		POpacity,
		POverflow,
		PPosition,
		PRight,
		PTop,
		PTransform,
		PTransformOrigin,
		PVerticalAlign,
		PZIndex,
	)

	InitialNotComputed = NewSetK(
		PDisplay,
		PColumnGap,
		PBleedTop,
		PBleedLeft,
		PBleedBottom,
		PBleedRight,
		POutlineWidth,
		POutlineColor,
		PColumnRuleWidth,
		PColumnRuleColor,
		PBorderTopWidth,
		PBorderLeftWidth,
		PBorderBottomWidth,
		PBorderRightWidth,
		PBorderTopColor,
		PBorderLeftColor,
		PBorderBottomColor,
		PBorderRightColor,
	)
)

func init() {
	for name := range InitialValues {
		KnownProperties.Add(name)
	}
}
