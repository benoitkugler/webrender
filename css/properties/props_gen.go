package properties

// Code generated from properties/properties.go DO NOT EDIT

func (s Properties) GetAlignContent() String  { return s[PAlignContent].(String) }
func (s Properties) SetAlignContent(v String) { s[PAlignContent] = v }

func (s Properties) GetAlignItems() String  { return s[PAlignItems].(String) }
func (s Properties) SetAlignItems(v String) { s[PAlignItems] = v }

func (s Properties) GetAlignSelf() String  { return s[PAlignSelf].(String) }
func (s Properties) SetAlignSelf(v String) { s[PAlignSelf] = v }

func (s Properties) GetAnchor() String  { return s[PAnchor].(String) }
func (s Properties) SetAnchor(v String) { s[PAnchor] = v }

func (s Properties) GetAppearance() String  { return s[PAppearance].(String) }
func (s Properties) SetAppearance(v String) { s[PAppearance] = v }

func (s Properties) GetBackgroundAttachment() Strings  { return s[PBackgroundAttachment].(Strings) }
func (s Properties) SetBackgroundAttachment(v Strings) { s[PBackgroundAttachment] = v }

func (s Properties) GetBackgroundClip() Strings  { return s[PBackgroundClip].(Strings) }
func (s Properties) SetBackgroundClip(v Strings) { s[PBackgroundClip] = v }

func (s Properties) GetBackgroundColor() Color  { return s[PBackgroundColor].(Color) }
func (s Properties) SetBackgroundColor(v Color) { s[PBackgroundColor] = v }

func (s Properties) GetBackgroundImage() Images  { return s[PBackgroundImage].(Images) }
func (s Properties) SetBackgroundImage(v Images) { s[PBackgroundImage] = v }

func (s Properties) GetBackgroundOrigin() Strings  { return s[PBackgroundOrigin].(Strings) }
func (s Properties) SetBackgroundOrigin(v Strings) { s[PBackgroundOrigin] = v }

func (s Properties) GetBackgroundPosition() Centers  { return s[PBackgroundPosition].(Centers) }
func (s Properties) SetBackgroundPosition(v Centers) { s[PBackgroundPosition] = v }

func (s Properties) GetBackgroundRepeat() Repeats  { return s[PBackgroundRepeat].(Repeats) }
func (s Properties) SetBackgroundRepeat(v Repeats) { s[PBackgroundRepeat] = v }

func (s Properties) GetBackgroundSize() Sizes  { return s[PBackgroundSize].(Sizes) }
func (s Properties) SetBackgroundSize(v Sizes) { s[PBackgroundSize] = v }

func (s Properties) GetBleedBottom() DimOrS  { return s[PBleedBottom].(DimOrS) }
func (s Properties) SetBleedBottom(v DimOrS) { s[PBleedBottom] = v }

func (s Properties) GetBleedLeft() DimOrS  { return s[PBleedLeft].(DimOrS) }
func (s Properties) SetBleedLeft(v DimOrS) { s[PBleedLeft] = v }

func (s Properties) GetBleedRight() DimOrS  { return s[PBleedRight].(DimOrS) }
func (s Properties) SetBleedRight(v DimOrS) { s[PBleedRight] = v }

func (s Properties) GetBleedTop() DimOrS  { return s[PBleedTop].(DimOrS) }
func (s Properties) SetBleedTop(v DimOrS) { s[PBleedTop] = v }

func (s Properties) GetBlockEllipsis() TaggedString  { return s[PBlockEllipsis].(TaggedString) }
func (s Properties) SetBlockEllipsis(v TaggedString) { s[PBlockEllipsis] = v }

func (s Properties) GetBookmarkLabel() ContentProperties {
	return s[PBookmarkLabel].(ContentProperties)
}
func (s Properties) SetBookmarkLabel(v ContentProperties) { s[PBookmarkLabel] = v }

func (s Properties) GetBookmarkLevel() IntString  { return s[PBookmarkLevel].(IntString) }
func (s Properties) SetBookmarkLevel(v IntString) { s[PBookmarkLevel] = v }

func (s Properties) GetBookmarkState() String  { return s[PBookmarkState].(String) }
func (s Properties) SetBookmarkState(v String) { s[PBookmarkState] = v }

func (s Properties) GetBorderBottomColor() Color  { return s[PBorderBottomColor].(Color) }
func (s Properties) SetBorderBottomColor(v Color) { s[PBorderBottomColor] = v }

func (s Properties) GetBorderBottomLeftRadius() Point  { return s[PBorderBottomLeftRadius].(Point) }
func (s Properties) SetBorderBottomLeftRadius(v Point) { s[PBorderBottomLeftRadius] = v }

func (s Properties) GetBorderBottomRightRadius() Point  { return s[PBorderBottomRightRadius].(Point) }
func (s Properties) SetBorderBottomRightRadius(v Point) { s[PBorderBottomRightRadius] = v }

func (s Properties) GetBorderBottomStyle() String  { return s[PBorderBottomStyle].(String) }
func (s Properties) SetBorderBottomStyle(v String) { s[PBorderBottomStyle] = v }

func (s Properties) GetBorderBottomWidth() DimOrS  { return s[PBorderBottomWidth].(DimOrS) }
func (s Properties) SetBorderBottomWidth(v DimOrS) { s[PBorderBottomWidth] = v }

func (s Properties) GetBorderCollapse() String  { return s[PBorderCollapse].(String) }
func (s Properties) SetBorderCollapse(v String) { s[PBorderCollapse] = v }

func (s Properties) GetBorderImageOutset() Values  { return s[PBorderImageOutset].(Values) }
func (s Properties) SetBorderImageOutset(v Values) { s[PBorderImageOutset] = v }

func (s Properties) GetBorderImageRepeat() Strings  { return s[PBorderImageRepeat].(Strings) }
func (s Properties) SetBorderImageRepeat(v Strings) { s[PBorderImageRepeat] = v }

func (s Properties) GetBorderImageSlice() Values  { return s[PBorderImageSlice].(Values) }
func (s Properties) SetBorderImageSlice(v Values) { s[PBorderImageSlice] = v }

func (s Properties) GetBorderImageSource() Image  { return s[PBorderImageSource].(Image) }
func (s Properties) SetBorderImageSource(v Image) { s[PBorderImageSource] = v }

func (s Properties) GetBorderImageWidth() Values  { return s[PBorderImageWidth].(Values) }
func (s Properties) SetBorderImageWidth(v Values) { s[PBorderImageWidth] = v }

func (s Properties) GetBorderLeftColor() Color  { return s[PBorderLeftColor].(Color) }
func (s Properties) SetBorderLeftColor(v Color) { s[PBorderLeftColor] = v }

func (s Properties) GetBorderLeftStyle() String  { return s[PBorderLeftStyle].(String) }
func (s Properties) SetBorderLeftStyle(v String) { s[PBorderLeftStyle] = v }

func (s Properties) GetBorderLeftWidth() DimOrS  { return s[PBorderLeftWidth].(DimOrS) }
func (s Properties) SetBorderLeftWidth(v DimOrS) { s[PBorderLeftWidth] = v }

func (s Properties) GetBorderRightColor() Color  { return s[PBorderRightColor].(Color) }
func (s Properties) SetBorderRightColor(v Color) { s[PBorderRightColor] = v }

func (s Properties) GetBorderRightStyle() String  { return s[PBorderRightStyle].(String) }
func (s Properties) SetBorderRightStyle(v String) { s[PBorderRightStyle] = v }

func (s Properties) GetBorderRightWidth() DimOrS  { return s[PBorderRightWidth].(DimOrS) }
func (s Properties) SetBorderRightWidth(v DimOrS) { s[PBorderRightWidth] = v }

func (s Properties) GetBorderSpacing() Point  { return s[PBorderSpacing].(Point) }
func (s Properties) SetBorderSpacing(v Point) { s[PBorderSpacing] = v }

func (s Properties) GetBorderTopColor() Color  { return s[PBorderTopColor].(Color) }
func (s Properties) SetBorderTopColor(v Color) { s[PBorderTopColor] = v }

func (s Properties) GetBorderTopLeftRadius() Point  { return s[PBorderTopLeftRadius].(Point) }
func (s Properties) SetBorderTopLeftRadius(v Point) { s[PBorderTopLeftRadius] = v }

func (s Properties) GetBorderTopRightRadius() Point  { return s[PBorderTopRightRadius].(Point) }
func (s Properties) SetBorderTopRightRadius(v Point) { s[PBorderTopRightRadius] = v }

func (s Properties) GetBorderTopStyle() String  { return s[PBorderTopStyle].(String) }
func (s Properties) SetBorderTopStyle(v String) { s[PBorderTopStyle] = v }

func (s Properties) GetBorderTopWidth() DimOrS  { return s[PBorderTopWidth].(DimOrS) }
func (s Properties) SetBorderTopWidth(v DimOrS) { s[PBorderTopWidth] = v }

func (s Properties) GetBottom() DimOrS  { return s[PBottom].(DimOrS) }
func (s Properties) SetBottom(v DimOrS) { s[PBottom] = v }

func (s Properties) GetBoxDecorationBreak() String  { return s[PBoxDecorationBreak].(String) }
func (s Properties) SetBoxDecorationBreak(v String) { s[PBoxDecorationBreak] = v }

func (s Properties) GetBoxSizing() String  { return s[PBoxSizing].(String) }
func (s Properties) SetBoxSizing(v String) { s[PBoxSizing] = v }

func (s Properties) GetBreakAfter() String  { return s[PBreakAfter].(String) }
func (s Properties) SetBreakAfter(v String) { s[PBreakAfter] = v }

func (s Properties) GetBreakBefore() String  { return s[PBreakBefore].(String) }
func (s Properties) SetBreakBefore(v String) { s[PBreakBefore] = v }

func (s Properties) GetBreakInside() String  { return s[PBreakInside].(String) }
func (s Properties) SetBreakInside(v String) { s[PBreakInside] = v }

func (s Properties) GetCaptionSide() String  { return s[PCaptionSide].(String) }
func (s Properties) SetCaptionSide(v String) { s[PCaptionSide] = v }

func (s Properties) GetClear() String  { return s[PClear].(String) }
func (s Properties) SetClear(v String) { s[PClear] = v }

func (s Properties) GetClip() Values  { return s[PClip].(Values) }
func (s Properties) SetClip(v Values) { s[PClip] = v }

func (s Properties) GetColor() Color  { return s[PColor].(Color) }
func (s Properties) SetColor(v Color) { s[PColor] = v }

func (s Properties) GetColumnCount() IntString  { return s[PColumnCount].(IntString) }
func (s Properties) SetColumnCount(v IntString) { s[PColumnCount] = v }

func (s Properties) GetColumnFill() String  { return s[PColumnFill].(String) }
func (s Properties) SetColumnFill(v String) { s[PColumnFill] = v }

func (s Properties) GetColumnGap() DimOrS  { return s[PColumnGap].(DimOrS) }
func (s Properties) SetColumnGap(v DimOrS) { s[PColumnGap] = v }

func (s Properties) GetColumnRuleColor() Color  { return s[PColumnRuleColor].(Color) }
func (s Properties) SetColumnRuleColor(v Color) { s[PColumnRuleColor] = v }

func (s Properties) GetColumnRuleStyle() String  { return s[PColumnRuleStyle].(String) }
func (s Properties) SetColumnRuleStyle(v String) { s[PColumnRuleStyle] = v }

func (s Properties) GetColumnRuleWidth() DimOrS  { return s[PColumnRuleWidth].(DimOrS) }
func (s Properties) SetColumnRuleWidth(v DimOrS) { s[PColumnRuleWidth] = v }

func (s Properties) GetColumnSpan() String  { return s[PColumnSpan].(String) }
func (s Properties) SetColumnSpan(v String) { s[PColumnSpan] = v }

func (s Properties) GetColumnWidth() DimOrS  { return s[PColumnWidth].(DimOrS) }
func (s Properties) SetColumnWidth(v DimOrS) { s[PColumnWidth] = v }

func (s Properties) GetContent() SContent  { return s[PContent].(SContent) }
func (s Properties) SetContent(v SContent) { s[PContent] = v }

func (s Properties) GetContinue() String  { return s[PContinue].(String) }
func (s Properties) SetContinue(v String) { s[PContinue] = v }

func (s Properties) GetCounterIncrement() SIntStrings  { return s[PCounterIncrement].(SIntStrings) }
func (s Properties) SetCounterIncrement(v SIntStrings) { s[PCounterIncrement] = v }

func (s Properties) GetCounterReset() SIntStrings  { return s[PCounterReset].(SIntStrings) }
func (s Properties) SetCounterReset(v SIntStrings) { s[PCounterReset] = v }

func (s Properties) GetCounterSet() SIntStrings  { return s[PCounterSet].(SIntStrings) }
func (s Properties) SetCounterSet(v SIntStrings) { s[PCounterSet] = v }

func (s Properties) GetDirection() String  { return s[PDirection].(String) }
func (s Properties) SetDirection(v String) { s[PDirection] = v }

func (s Properties) GetDisplay() Display  { return s[PDisplay].(Display) }
func (s Properties) SetDisplay(v Display) { s[PDisplay] = v }

func (s Properties) GetEmptyCells() String  { return s[PEmptyCells].(String) }
func (s Properties) SetEmptyCells(v String) { s[PEmptyCells] = v }

func (s Properties) GetFlexBasis() DimOrS  { return s[PFlexBasis].(DimOrS) }
func (s Properties) SetFlexBasis(v DimOrS) { s[PFlexBasis] = v }

func (s Properties) GetFlexDirection() String  { return s[PFlexDirection].(String) }
func (s Properties) SetFlexDirection(v String) { s[PFlexDirection] = v }

func (s Properties) GetFlexGrow() Float  { return s[PFlexGrow].(Float) }
func (s Properties) SetFlexGrow(v Float) { s[PFlexGrow] = v }

func (s Properties) GetFlexShrink() Float  { return s[PFlexShrink].(Float) }
func (s Properties) SetFlexShrink(v Float) { s[PFlexShrink] = v }

func (s Properties) GetFlexWrap() String  { return s[PFlexWrap].(String) }
func (s Properties) SetFlexWrap(v String) { s[PFlexWrap] = v }

func (s Properties) GetFloat() String  { return s[PFloat].(String) }
func (s Properties) SetFloat(v String) { s[PFloat] = v }

func (s Properties) GetFontFamily() Strings  { return s[PFontFamily].(Strings) }
func (s Properties) SetFontFamily(v Strings) { s[PFontFamily] = v }

func (s Properties) GetFontFeatureSettings() SIntStrings {
	return s[PFontFeatureSettings].(SIntStrings)
}
func (s Properties) SetFontFeatureSettings(v SIntStrings) { s[PFontFeatureSettings] = v }

func (s Properties) GetFontKerning() String  { return s[PFontKerning].(String) }
func (s Properties) SetFontKerning(v String) { s[PFontKerning] = v }

func (s Properties) GetFontLanguageOverride() String  { return s[PFontLanguageOverride].(String) }
func (s Properties) SetFontLanguageOverride(v String) { s[PFontLanguageOverride] = v }

func (s Properties) GetFontSize() DimOrS  { return s[PFontSize].(DimOrS) }
func (s Properties) SetFontSize(v DimOrS) { s[PFontSize] = v }

func (s Properties) GetFontStretch() String  { return s[PFontStretch].(String) }
func (s Properties) SetFontStretch(v String) { s[PFontStretch] = v }

func (s Properties) GetFontStyle() String  { return s[PFontStyle].(String) }
func (s Properties) SetFontStyle(v String) { s[PFontStyle] = v }

func (s Properties) GetFontVariant() String  { return s[PFontVariant].(String) }
func (s Properties) SetFontVariant(v String) { s[PFontVariant] = v }

func (s Properties) GetFontVariantAlternates() String  { return s[PFontVariantAlternates].(String) }
func (s Properties) SetFontVariantAlternates(v String) { s[PFontVariantAlternates] = v }

func (s Properties) GetFontVariantCaps() String  { return s[PFontVariantCaps].(String) }
func (s Properties) SetFontVariantCaps(v String) { s[PFontVariantCaps] = v }

func (s Properties) GetFontVariantEastAsian() SStrings  { return s[PFontVariantEastAsian].(SStrings) }
func (s Properties) SetFontVariantEastAsian(v SStrings) { s[PFontVariantEastAsian] = v }

func (s Properties) GetFontVariantLigatures() SStrings  { return s[PFontVariantLigatures].(SStrings) }
func (s Properties) SetFontVariantLigatures(v SStrings) { s[PFontVariantLigatures] = v }

func (s Properties) GetFontVariantNumeric() SStrings  { return s[PFontVariantNumeric].(SStrings) }
func (s Properties) SetFontVariantNumeric(v SStrings) { s[PFontVariantNumeric] = v }

func (s Properties) GetFontVariantPosition() String  { return s[PFontVariantPosition].(String) }
func (s Properties) SetFontVariantPosition(v String) { s[PFontVariantPosition] = v }

func (s Properties) GetFontVariationSettings() SFloatStrings {
	return s[PFontVariationSettings].(SFloatStrings)
}
func (s Properties) SetFontVariationSettings(v SFloatStrings) { s[PFontVariationSettings] = v }

func (s Properties) GetFontWeight() IntString  { return s[PFontWeight].(IntString) }
func (s Properties) SetFontWeight(v IntString) { s[PFontWeight] = v }

func (s Properties) GetFootnoteDisplay() String  { return s[PFootnoteDisplay].(String) }
func (s Properties) SetFootnoteDisplay(v String) { s[PFootnoteDisplay] = v }

func (s Properties) GetFootnotePolicy() String  { return s[PFootnotePolicy].(String) }
func (s Properties) SetFootnotePolicy(v String) { s[PFootnotePolicy] = v }

func (s Properties) GetGridAutoColumns() GridAuto  { return s[PGridAutoColumns].(GridAuto) }
func (s Properties) SetGridAutoColumns(v GridAuto) { s[PGridAutoColumns] = v }

func (s Properties) GetGridAutoFlow() Strings  { return s[PGridAutoFlow].(Strings) }
func (s Properties) SetGridAutoFlow(v Strings) { s[PGridAutoFlow] = v }

func (s Properties) GetGridAutoRows() GridAuto  { return s[PGridAutoRows].(GridAuto) }
func (s Properties) SetGridAutoRows(v GridAuto) { s[PGridAutoRows] = v }

func (s Properties) GetGridColumnEnd() GridLine  { return s[PGridColumnEnd].(GridLine) }
func (s Properties) SetGridColumnEnd(v GridLine) { s[PGridColumnEnd] = v }

func (s Properties) GetGridColumnStart() GridLine  { return s[PGridColumnStart].(GridLine) }
func (s Properties) SetGridColumnStart(v GridLine) { s[PGridColumnStart] = v }

func (s Properties) GetGridRowEnd() GridLine  { return s[PGridRowEnd].(GridLine) }
func (s Properties) SetGridRowEnd(v GridLine) { s[PGridRowEnd] = v }

func (s Properties) GetGridRowStart() GridLine  { return s[PGridRowStart].(GridLine) }
func (s Properties) SetGridRowStart(v GridLine) { s[PGridRowStart] = v }

func (s Properties) GetGridTemplateAreas() GridTemplateAreas {
	return s[PGridTemplateAreas].(GridTemplateAreas)
}
func (s Properties) SetGridTemplateAreas(v GridTemplateAreas) { s[PGridTemplateAreas] = v }

func (s Properties) GetGridTemplateColumns() GridTemplate {
	return s[PGridTemplateColumns].(GridTemplate)
}
func (s Properties) SetGridTemplateColumns(v GridTemplate) { s[PGridTemplateColumns] = v }

func (s Properties) GetGridTemplateRows() GridTemplate  { return s[PGridTemplateRows].(GridTemplate) }
func (s Properties) SetGridTemplateRows(v GridTemplate) { s[PGridTemplateRows] = v }

func (s Properties) GetHeight() DimOrS  { return s[PHeight].(DimOrS) }
func (s Properties) SetHeight(v DimOrS) { s[PHeight] = v }

func (s Properties) GetHyphenateCharacter() String  { return s[PHyphenateCharacter].(String) }
func (s Properties) SetHyphenateCharacter(v String) { s[PHyphenateCharacter] = v }

func (s Properties) GetHyphenateLimitChars() Ints3  { return s[PHyphenateLimitChars].(Ints3) }
func (s Properties) SetHyphenateLimitChars(v Ints3) { s[PHyphenateLimitChars] = v }

func (s Properties) GetHyphenateLimitZone() DimOrS  { return s[PHyphenateLimitZone].(DimOrS) }
func (s Properties) SetHyphenateLimitZone(v DimOrS) { s[PHyphenateLimitZone] = v }

func (s Properties) GetHyphens() String  { return s[PHyphens].(String) }
func (s Properties) SetHyphens(v String) { s[PHyphens] = v }

func (s Properties) GetImageOrientation() SBoolFloat  { return s[PImageOrientation].(SBoolFloat) }
func (s Properties) SetImageOrientation(v SBoolFloat) { s[PImageOrientation] = v }

func (s Properties) GetImageRendering() String  { return s[PImageRendering].(String) }
func (s Properties) SetImageRendering(v String) { s[PImageRendering] = v }

func (s Properties) GetImageResolution() DimOrS  { return s[PImageResolution].(DimOrS) }
func (s Properties) SetImageResolution(v DimOrS) { s[PImageResolution] = v }

func (s Properties) GetJustifyContent() String  { return s[PJustifyContent].(String) }
func (s Properties) SetJustifyContent(v String) { s[PJustifyContent] = v }

func (s Properties) GetJustifyItems() String  { return s[PJustifyItems].(String) }
func (s Properties) SetJustifyItems(v String) { s[PJustifyItems] = v }

func (s Properties) GetJustifySelf() String  { return s[PJustifySelf].(String) }
func (s Properties) SetJustifySelf(v String) { s[PJustifySelf] = v }

func (s Properties) GetLang() NamedString  { return s[PLang].(NamedString) }
func (s Properties) SetLang(v NamedString) { s[PLang] = v }

func (s Properties) GetLeft() DimOrS  { return s[PLeft].(DimOrS) }
func (s Properties) SetLeft(v DimOrS) { s[PLeft] = v }

func (s Properties) GetLetterSpacing() DimOrS  { return s[PLetterSpacing].(DimOrS) }
func (s Properties) SetLetterSpacing(v DimOrS) { s[PLetterSpacing] = v }

func (s Properties) GetLineHeight() DimOrS  { return s[PLineHeight].(DimOrS) }
func (s Properties) SetLineHeight(v DimOrS) { s[PLineHeight] = v }

func (s Properties) GetLink() NamedString  { return s[PLink].(NamedString) }
func (s Properties) SetLink(v NamedString) { s[PLink] = v }

func (s Properties) GetListStyleImage() Image  { return s[PListStyleImage].(Image) }
func (s Properties) SetListStyleImage(v Image) { s[PListStyleImage] = v }

func (s Properties) GetListStylePosition() String  { return s[PListStylePosition].(String) }
func (s Properties) SetListStylePosition(v String) { s[PListStylePosition] = v }

func (s Properties) GetListStyleType() CounterStyleID  { return s[PListStyleType].(CounterStyleID) }
func (s Properties) SetListStyleType(v CounterStyleID) { s[PListStyleType] = v }

func (s Properties) GetMarginBottom() DimOrS  { return s[PMarginBottom].(DimOrS) }
func (s Properties) SetMarginBottom(v DimOrS) { s[PMarginBottom] = v }

func (s Properties) GetMarginBreak() String  { return s[PMarginBreak].(String) }
func (s Properties) SetMarginBreak(v String) { s[PMarginBreak] = v }

func (s Properties) GetMarginLeft() DimOrS  { return s[PMarginLeft].(DimOrS) }
func (s Properties) SetMarginLeft(v DimOrS) { s[PMarginLeft] = v }

func (s Properties) GetMarginRight() DimOrS  { return s[PMarginRight].(DimOrS) }
func (s Properties) SetMarginRight(v DimOrS) { s[PMarginRight] = v }

func (s Properties) GetMarginTop() DimOrS  { return s[PMarginTop].(DimOrS) }
func (s Properties) SetMarginTop(v DimOrS) { s[PMarginTop] = v }

func (s Properties) GetMarks() Marks  { return s[PMarks].(Marks) }
func (s Properties) SetMarks(v Marks) { s[PMarks] = v }

func (s Properties) GetMaxHeight() DimOrS  { return s[PMaxHeight].(DimOrS) }
func (s Properties) SetMaxHeight(v DimOrS) { s[PMaxHeight] = v }

func (s Properties) GetMaxLines() TaggedInt  { return s[PMaxLines].(TaggedInt) }
func (s Properties) SetMaxLines(v TaggedInt) { s[PMaxLines] = v }

func (s Properties) GetMaxWidth() DimOrS  { return s[PMaxWidth].(DimOrS) }
func (s Properties) SetMaxWidth(v DimOrS) { s[PMaxWidth] = v }

func (s Properties) GetMinHeight() DimOrS  { return s[PMinHeight].(DimOrS) }
func (s Properties) SetMinHeight(v DimOrS) { s[PMinHeight] = v }

func (s Properties) GetMinWidth() DimOrS  { return s[PMinWidth].(DimOrS) }
func (s Properties) SetMinWidth(v DimOrS) { s[PMinWidth] = v }

func (s Properties) GetObjectFit() String  { return s[PObjectFit].(String) }
func (s Properties) SetObjectFit(v String) { s[PObjectFit] = v }

func (s Properties) GetObjectPosition() Center  { return s[PObjectPosition].(Center) }
func (s Properties) SetObjectPosition(v Center) { s[PObjectPosition] = v }

func (s Properties) GetOpacity() Float  { return s[POpacity].(Float) }
func (s Properties) SetOpacity(v Float) { s[POpacity] = v }

func (s Properties) GetOrder() Int  { return s[POrder].(Int) }
func (s Properties) SetOrder(v Int) { s[POrder] = v }

func (s Properties) GetOrphans() Int  { return s[POrphans].(Int) }
func (s Properties) SetOrphans(v Int) { s[POrphans] = v }

func (s Properties) GetOutlineColor() Color  { return s[POutlineColor].(Color) }
func (s Properties) SetOutlineColor(v Color) { s[POutlineColor] = v }

func (s Properties) GetOutlineStyle() String  { return s[POutlineStyle].(String) }
func (s Properties) SetOutlineStyle(v String) { s[POutlineStyle] = v }

func (s Properties) GetOutlineWidth() DimOrS  { return s[POutlineWidth].(DimOrS) }
func (s Properties) SetOutlineWidth(v DimOrS) { s[POutlineWidth] = v }

func (s Properties) GetOverflow() String  { return s[POverflow].(String) }
func (s Properties) SetOverflow(v String) { s[POverflow] = v }

func (s Properties) GetOverflowWrap() String  { return s[POverflowWrap].(String) }
func (s Properties) SetOverflowWrap(v String) { s[POverflowWrap] = v }

func (s Properties) GetPaddingBottom() DimOrS  { return s[PPaddingBottom].(DimOrS) }
func (s Properties) SetPaddingBottom(v DimOrS) { s[PPaddingBottom] = v }

func (s Properties) GetPaddingLeft() DimOrS  { return s[PPaddingLeft].(DimOrS) }
func (s Properties) SetPaddingLeft(v DimOrS) { s[PPaddingLeft] = v }

func (s Properties) GetPaddingRight() DimOrS  { return s[PPaddingRight].(DimOrS) }
func (s Properties) SetPaddingRight(v DimOrS) { s[PPaddingRight] = v }

func (s Properties) GetPaddingTop() DimOrS  { return s[PPaddingTop].(DimOrS) }
func (s Properties) SetPaddingTop(v DimOrS) { s[PPaddingTop] = v }

func (s Properties) GetPage() Page  { return s[PPage].(Page) }
func (s Properties) SetPage(v Page) { s[PPage] = v }

func (s Properties) GetPosition() BoolString  { return s[PPosition].(BoolString) }
func (s Properties) SetPosition(v BoolString) { s[PPosition] = v }

func (s Properties) GetQuotes() Quotes  { return s[PQuotes].(Quotes) }
func (s Properties) SetQuotes(v Quotes) { s[PQuotes] = v }

func (s Properties) GetRight() DimOrS  { return s[PRight].(DimOrS) }
func (s Properties) SetRight(v DimOrS) { s[PRight] = v }

func (s Properties) GetRowGap() DimOrS  { return s[PRowGap].(DimOrS) }
func (s Properties) SetRowGap(v DimOrS) { s[PRowGap] = v }

func (s Properties) GetSize() Point  { return s[PSize].(Point) }
func (s Properties) SetSize(v Point) { s[PSize] = v }

func (s Properties) GetStringSet() StringSet  { return s[PStringSet].(StringSet) }
func (s Properties) SetStringSet(v StringSet) { s[PStringSet] = v }

func (s Properties) GetTabSize() DimOrS  { return s[PTabSize].(DimOrS) }
func (s Properties) SetTabSize(v DimOrS) { s[PTabSize] = v }

func (s Properties) GetTableLayout() String  { return s[PTableLayout].(String) }
func (s Properties) SetTableLayout(v String) { s[PTableLayout] = v }

func (s Properties) GetTextAlignAll() String  { return s[PTextAlignAll].(String) }
func (s Properties) SetTextAlignAll(v String) { s[PTextAlignAll] = v }

func (s Properties) GetTextAlignLast() String  { return s[PTextAlignLast].(String) }
func (s Properties) SetTextAlignLast(v String) { s[PTextAlignLast] = v }

func (s Properties) GetTextDecorationColor() Color  { return s[PTextDecorationColor].(Color) }
func (s Properties) SetTextDecorationColor(v Color) { s[PTextDecorationColor] = v }

func (s Properties) GetTextDecorationLine() Decorations  { return s[PTextDecorationLine].(Decorations) }
func (s Properties) SetTextDecorationLine(v Decorations) { s[PTextDecorationLine] = v }

func (s Properties) GetTextDecorationStyle() String  { return s[PTextDecorationStyle].(String) }
func (s Properties) SetTextDecorationStyle(v String) { s[PTextDecorationStyle] = v }

func (s Properties) GetTextIndent() DimOrS  { return s[PTextIndent].(DimOrS) }
func (s Properties) SetTextIndent(v DimOrS) { s[PTextIndent] = v }

func (s Properties) GetTextOverflow() String  { return s[PTextOverflow].(String) }
func (s Properties) SetTextOverflow(v String) { s[PTextOverflow] = v }

func (s Properties) GetTextTransform() String  { return s[PTextTransform].(String) }
func (s Properties) SetTextTransform(v String) { s[PTextTransform] = v }

func (s Properties) GetTop() DimOrS  { return s[PTop].(DimOrS) }
func (s Properties) SetTop(v DimOrS) { s[PTop] = v }

func (s Properties) GetTransform() Transforms  { return s[PTransform].(Transforms) }
func (s Properties) SetTransform(v Transforms) { s[PTransform] = v }

func (s Properties) GetTransformOrigin() Point  { return s[PTransformOrigin].(Point) }
func (s Properties) SetTransformOrigin(v Point) { s[PTransformOrigin] = v }

func (s Properties) GetUnicodeBidi() String  { return s[PUnicodeBidi].(String) }
func (s Properties) SetUnicodeBidi(v String) { s[PUnicodeBidi] = v }

func (s Properties) GetVerticalAlign() DimOrS  { return s[PVerticalAlign].(DimOrS) }
func (s Properties) SetVerticalAlign(v DimOrS) { s[PVerticalAlign] = v }

func (s Properties) GetVisibility() String  { return s[PVisibility].(String) }
func (s Properties) SetVisibility(v String) { s[PVisibility] = v }

func (s Properties) GetWhiteSpace() String  { return s[PWhiteSpace].(String) }
func (s Properties) SetWhiteSpace(v String) { s[PWhiteSpace] = v }

func (s Properties) GetWidows() Int  { return s[PWidows].(Int) }
func (s Properties) SetWidows(v Int) { s[PWidows] = v }

func (s Properties) GetWidth() DimOrS  { return s[PWidth].(DimOrS) }
func (s Properties) SetWidth(v DimOrS) { s[PWidth] = v }

func (s Properties) GetWordBreak() String  { return s[PWordBreak].(String) }
func (s Properties) SetWordBreak(v String) { s[PWordBreak] = v }

func (s Properties) GetWordSpacing() DimOrS  { return s[PWordSpacing].(DimOrS) }
func (s Properties) SetWordSpacing(v DimOrS) { s[PWordSpacing] = v }

func (s Properties) GetZIndex() IntString  { return s[PZIndex].(IntString) }
func (s Properties) SetZIndex(v IntString) { s[PZIndex] = v }

type StyleAccessor interface {
	GetAlignContent() String
	SetAlignContent(v String)

	GetAlignItems() String
	SetAlignItems(v String)

	GetAlignSelf() String
	SetAlignSelf(v String)

	GetAnchor() String
	SetAnchor(v String)

	GetAppearance() String
	SetAppearance(v String)

	GetBackgroundAttachment() Strings
	SetBackgroundAttachment(v Strings)

	GetBackgroundClip() Strings
	SetBackgroundClip(v Strings)

	GetBackgroundColor() Color
	SetBackgroundColor(v Color)

	GetBackgroundImage() Images
	SetBackgroundImage(v Images)

	GetBackgroundOrigin() Strings
	SetBackgroundOrigin(v Strings)

	GetBackgroundPosition() Centers
	SetBackgroundPosition(v Centers)

	GetBackgroundRepeat() Repeats
	SetBackgroundRepeat(v Repeats)

	GetBackgroundSize() Sizes
	SetBackgroundSize(v Sizes)

	GetBleedBottom() DimOrS
	SetBleedBottom(v DimOrS)

	GetBleedLeft() DimOrS
	SetBleedLeft(v DimOrS)

	GetBleedRight() DimOrS
	SetBleedRight(v DimOrS)

	GetBleedTop() DimOrS
	SetBleedTop(v DimOrS)

	GetBlockEllipsis() TaggedString
	SetBlockEllipsis(v TaggedString)

	GetBookmarkLabel() ContentProperties
	SetBookmarkLabel(v ContentProperties)

	GetBookmarkLevel() IntString
	SetBookmarkLevel(v IntString)

	GetBookmarkState() String
	SetBookmarkState(v String)

	GetBorderBottomColor() Color
	SetBorderBottomColor(v Color)

	GetBorderBottomLeftRadius() Point
	SetBorderBottomLeftRadius(v Point)

	GetBorderBottomRightRadius() Point
	SetBorderBottomRightRadius(v Point)

	GetBorderBottomStyle() String
	SetBorderBottomStyle(v String)

	GetBorderBottomWidth() DimOrS
	SetBorderBottomWidth(v DimOrS)

	GetBorderCollapse() String
	SetBorderCollapse(v String)

	GetBorderImageOutset() Values
	SetBorderImageOutset(v Values)

	GetBorderImageRepeat() Strings
	SetBorderImageRepeat(v Strings)

	GetBorderImageSlice() Values
	SetBorderImageSlice(v Values)

	GetBorderImageSource() Image
	SetBorderImageSource(v Image)

	GetBorderImageWidth() Values
	SetBorderImageWidth(v Values)

	GetBorderLeftColor() Color
	SetBorderLeftColor(v Color)

	GetBorderLeftStyle() String
	SetBorderLeftStyle(v String)

	GetBorderLeftWidth() DimOrS
	SetBorderLeftWidth(v DimOrS)

	GetBorderRightColor() Color
	SetBorderRightColor(v Color)

	GetBorderRightStyle() String
	SetBorderRightStyle(v String)

	GetBorderRightWidth() DimOrS
	SetBorderRightWidth(v DimOrS)

	GetBorderSpacing() Point
	SetBorderSpacing(v Point)

	GetBorderTopColor() Color
	SetBorderTopColor(v Color)

	GetBorderTopLeftRadius() Point
	SetBorderTopLeftRadius(v Point)

	GetBorderTopRightRadius() Point
	SetBorderTopRightRadius(v Point)

	GetBorderTopStyle() String
	SetBorderTopStyle(v String)

	GetBorderTopWidth() DimOrS
	SetBorderTopWidth(v DimOrS)

	GetBottom() DimOrS
	SetBottom(v DimOrS)

	GetBoxDecorationBreak() String
	SetBoxDecorationBreak(v String)

	GetBoxSizing() String
	SetBoxSizing(v String)

	GetBreakAfter() String
	SetBreakAfter(v String)

	GetBreakBefore() String
	SetBreakBefore(v String)

	GetBreakInside() String
	SetBreakInside(v String)

	GetCaptionSide() String
	SetCaptionSide(v String)

	GetClear() String
	SetClear(v String)

	GetClip() Values
	SetClip(v Values)

	GetColor() Color
	SetColor(v Color)

	GetColumnCount() IntString
	SetColumnCount(v IntString)

	GetColumnFill() String
	SetColumnFill(v String)

	GetColumnGap() DimOrS
	SetColumnGap(v DimOrS)

	GetColumnRuleColor() Color
	SetColumnRuleColor(v Color)

	GetColumnRuleStyle() String
	SetColumnRuleStyle(v String)

	GetColumnRuleWidth() DimOrS
	SetColumnRuleWidth(v DimOrS)

	GetColumnSpan() String
	SetColumnSpan(v String)

	GetColumnWidth() DimOrS
	SetColumnWidth(v DimOrS)

	GetContent() SContent
	SetContent(v SContent)

	GetContinue() String
	SetContinue(v String)

	GetCounterIncrement() SIntStrings
	SetCounterIncrement(v SIntStrings)

	GetCounterReset() SIntStrings
	SetCounterReset(v SIntStrings)

	GetCounterSet() SIntStrings
	SetCounterSet(v SIntStrings)

	GetDirection() String
	SetDirection(v String)

	GetDisplay() Display
	SetDisplay(v Display)

	GetEmptyCells() String
	SetEmptyCells(v String)

	GetFlexBasis() DimOrS
	SetFlexBasis(v DimOrS)

	GetFlexDirection() String
	SetFlexDirection(v String)

	GetFlexGrow() Float
	SetFlexGrow(v Float)

	GetFlexShrink() Float
	SetFlexShrink(v Float)

	GetFlexWrap() String
	SetFlexWrap(v String)

	GetFloat() String
	SetFloat(v String)

	GetFontFamily() Strings
	SetFontFamily(v Strings)

	GetFontFeatureSettings() SIntStrings
	SetFontFeatureSettings(v SIntStrings)

	GetFontKerning() String
	SetFontKerning(v String)

	GetFontLanguageOverride() String
	SetFontLanguageOverride(v String)

	GetFontSize() DimOrS
	SetFontSize(v DimOrS)

	GetFontStretch() String
	SetFontStretch(v String)

	GetFontStyle() String
	SetFontStyle(v String)

	GetFontVariant() String
	SetFontVariant(v String)

	GetFontVariantAlternates() String
	SetFontVariantAlternates(v String)

	GetFontVariantCaps() String
	SetFontVariantCaps(v String)

	GetFontVariantEastAsian() SStrings
	SetFontVariantEastAsian(v SStrings)

	GetFontVariantLigatures() SStrings
	SetFontVariantLigatures(v SStrings)

	GetFontVariantNumeric() SStrings
	SetFontVariantNumeric(v SStrings)

	GetFontVariantPosition() String
	SetFontVariantPosition(v String)

	GetFontVariationSettings() SFloatStrings
	SetFontVariationSettings(v SFloatStrings)

	GetFontWeight() IntString
	SetFontWeight(v IntString)

	GetFootnoteDisplay() String
	SetFootnoteDisplay(v String)

	GetFootnotePolicy() String
	SetFootnotePolicy(v String)

	GetGridAutoColumns() GridAuto
	SetGridAutoColumns(v GridAuto)

	GetGridAutoFlow() Strings
	SetGridAutoFlow(v Strings)

	GetGridAutoRows() GridAuto
	SetGridAutoRows(v GridAuto)

	GetGridColumnEnd() GridLine
	SetGridColumnEnd(v GridLine)

	GetGridColumnStart() GridLine
	SetGridColumnStart(v GridLine)

	GetGridRowEnd() GridLine
	SetGridRowEnd(v GridLine)

	GetGridRowStart() GridLine
	SetGridRowStart(v GridLine)

	GetGridTemplateAreas() GridTemplateAreas
	SetGridTemplateAreas(v GridTemplateAreas)

	GetGridTemplateColumns() GridTemplate
	SetGridTemplateColumns(v GridTemplate)

	GetGridTemplateRows() GridTemplate
	SetGridTemplateRows(v GridTemplate)

	GetHeight() DimOrS
	SetHeight(v DimOrS)

	GetHyphenateCharacter() String
	SetHyphenateCharacter(v String)

	GetHyphenateLimitChars() Ints3
	SetHyphenateLimitChars(v Ints3)

	GetHyphenateLimitZone() DimOrS
	SetHyphenateLimitZone(v DimOrS)

	GetHyphens() String
	SetHyphens(v String)

	GetImageOrientation() SBoolFloat
	SetImageOrientation(v SBoolFloat)

	GetImageRendering() String
	SetImageRendering(v String)

	GetImageResolution() DimOrS
	SetImageResolution(v DimOrS)

	GetJustifyContent() String
	SetJustifyContent(v String)

	GetJustifyItems() String
	SetJustifyItems(v String)

	GetJustifySelf() String
	SetJustifySelf(v String)

	GetLang() NamedString
	SetLang(v NamedString)

	GetLeft() DimOrS
	SetLeft(v DimOrS)

	GetLetterSpacing() DimOrS
	SetLetterSpacing(v DimOrS)

	GetLineHeight() DimOrS
	SetLineHeight(v DimOrS)

	GetLink() NamedString
	SetLink(v NamedString)

	GetListStyleImage() Image
	SetListStyleImage(v Image)

	GetListStylePosition() String
	SetListStylePosition(v String)

	GetListStyleType() CounterStyleID
	SetListStyleType(v CounterStyleID)

	GetMarginBottom() DimOrS
	SetMarginBottom(v DimOrS)

	GetMarginBreak() String
	SetMarginBreak(v String)

	GetMarginLeft() DimOrS
	SetMarginLeft(v DimOrS)

	GetMarginRight() DimOrS
	SetMarginRight(v DimOrS)

	GetMarginTop() DimOrS
	SetMarginTop(v DimOrS)

	GetMarks() Marks
	SetMarks(v Marks)

	GetMaxHeight() DimOrS
	SetMaxHeight(v DimOrS)

	GetMaxLines() TaggedInt
	SetMaxLines(v TaggedInt)

	GetMaxWidth() DimOrS
	SetMaxWidth(v DimOrS)

	GetMinHeight() DimOrS
	SetMinHeight(v DimOrS)

	GetMinWidth() DimOrS
	SetMinWidth(v DimOrS)

	GetObjectFit() String
	SetObjectFit(v String)

	GetObjectPosition() Center
	SetObjectPosition(v Center)

	GetOpacity() Float
	SetOpacity(v Float)

	GetOrder() Int
	SetOrder(v Int)

	GetOrphans() Int
	SetOrphans(v Int)

	GetOutlineColor() Color
	SetOutlineColor(v Color)

	GetOutlineStyle() String
	SetOutlineStyle(v String)

	GetOutlineWidth() DimOrS
	SetOutlineWidth(v DimOrS)

	GetOverflow() String
	SetOverflow(v String)

	GetOverflowWrap() String
	SetOverflowWrap(v String)

	GetPaddingBottom() DimOrS
	SetPaddingBottom(v DimOrS)

	GetPaddingLeft() DimOrS
	SetPaddingLeft(v DimOrS)

	GetPaddingRight() DimOrS
	SetPaddingRight(v DimOrS)

	GetPaddingTop() DimOrS
	SetPaddingTop(v DimOrS)

	GetPage() Page
	SetPage(v Page)

	GetPosition() BoolString
	SetPosition(v BoolString)

	GetQuotes() Quotes
	SetQuotes(v Quotes)

	GetRight() DimOrS
	SetRight(v DimOrS)

	GetRowGap() DimOrS
	SetRowGap(v DimOrS)

	GetSize() Point
	SetSize(v Point)

	GetStringSet() StringSet
	SetStringSet(v StringSet)

	GetTabSize() DimOrS
	SetTabSize(v DimOrS)

	GetTableLayout() String
	SetTableLayout(v String)

	GetTextAlignAll() String
	SetTextAlignAll(v String)

	GetTextAlignLast() String
	SetTextAlignLast(v String)

	GetTextDecorationColor() Color
	SetTextDecorationColor(v Color)

	GetTextDecorationLine() Decorations
	SetTextDecorationLine(v Decorations)

	GetTextDecorationStyle() String
	SetTextDecorationStyle(v String)

	GetTextIndent() DimOrS
	SetTextIndent(v DimOrS)

	GetTextOverflow() String
	SetTextOverflow(v String)

	GetTextTransform() String
	SetTextTransform(v String)

	GetTop() DimOrS
	SetTop(v DimOrS)

	GetTransform() Transforms
	SetTransform(v Transforms)

	GetTransformOrigin() Point
	SetTransformOrigin(v Point)

	GetUnicodeBidi() String
	SetUnicodeBidi(v String)

	GetVerticalAlign() DimOrS
	SetVerticalAlign(v DimOrS)

	GetVisibility() String
	SetVisibility(v String)

	GetWhiteSpace() String
	SetWhiteSpace(v String)

	GetWidows() Int
	SetWidows(v Int)

	GetWidth() DimOrS
	SetWidth(v DimOrS)

	GetWordBreak() String
	SetWordBreak(v String)

	GetWordSpacing() DimOrS
	SetWordSpacing(v DimOrS)

	GetZIndex() IntString
	SetZIndex(v IntString)
}

var propsNames = [...]string{
	PAlignContent:            "align-content",
	PAlignItems:              "align-items",
	PAlignSelf:               "align-self",
	PAnchor:                  "anchor",
	PAppearance:              "appearance",
	PBackgroundAttachment:    "background-attachment",
	PBackgroundClip:          "background-clip",
	PBackgroundColor:         "background-color",
	PBackgroundImage:         "background-image",
	PBackgroundOrigin:        "background-origin",
	PBackgroundPosition:      "background-position",
	PBackgroundRepeat:        "background-repeat",
	PBackgroundSize:          "background-size",
	PBleedBottom:             "bleed-bottom",
	PBleedLeft:               "bleed-left",
	PBleedRight:              "bleed-right",
	PBleedTop:                "bleed-top",
	PBlockEllipsis:           "block-ellipsis",
	PBookmarkLabel:           "bookmark-label",
	PBookmarkLevel:           "bookmark-level",
	PBookmarkState:           "bookmark-state",
	PBorderBottomColor:       "border-bottom-color",
	PBorderBottomLeftRadius:  "border-bottom-left-radius",
	PBorderBottomRightRadius: "border-bottom-right-radius",
	PBorderBottomStyle:       "border-bottom-style",
	PBorderBottomWidth:       "border-bottom-width",
	PBorderCollapse:          "border-collapse",
	PBorderImageOutset:       "border-image-outset",
	PBorderImageRepeat:       "border-image-repeat",
	PBorderImageSlice:        "border-image-slice",
	PBorderImageSource:       "border-image-source",
	PBorderImageWidth:        "border-image-width",
	PBorderLeftColor:         "border-left-color",
	PBorderLeftStyle:         "border-left-style",
	PBorderLeftWidth:         "border-left-width",
	PBorderRightColor:        "border-right-color",
	PBorderRightStyle:        "border-right-style",
	PBorderRightWidth:        "border-right-width",
	PBorderSpacing:           "border-spacing",
	PBorderTopColor:          "border-top-color",
	PBorderTopLeftRadius:     "border-top-left-radius",
	PBorderTopRightRadius:    "border-top-right-radius",
	PBorderTopStyle:          "border-top-style",
	PBorderTopWidth:          "border-top-width",
	PBottom:                  "bottom",
	PBoxDecorationBreak:      "box-decoration-break",
	PBoxSizing:               "box-sizing",
	PBreakAfter:              "break-after",
	PBreakBefore:             "break-before",
	PBreakInside:             "break-inside",
	PCaptionSide:             "caption-side",
	PClear:                   "clear",
	PClip:                    "clip",
	PColor:                   "color",
	PColumnCount:             "column-count",
	PColumnFill:              "column-fill",
	PColumnGap:               "column-gap",
	PColumnRuleColor:         "column-rule-color",
	PColumnRuleStyle:         "column-rule-style",
	PColumnRuleWidth:         "column-rule-width",
	PColumnSpan:              "column-span",
	PColumnWidth:             "column-width",
	PContent:                 "content",
	PContinue:                "continue",
	PCounterIncrement:        "counter-increment",
	PCounterReset:            "counter-reset",
	PCounterSet:              "counter-set",
	PDirection:               "direction",
	PDisplay:                 "display",
	PEmptyCells:              "empty-cells",
	PFlexBasis:               "flex-basis",
	PFlexDirection:           "flex-direction",
	PFlexGrow:                "flex-grow",
	PFlexShrink:              "flex-shrink",
	PFlexWrap:                "flex-wrap",
	PFloat:                   "float",
	PFontFamily:              "font-family",
	PFontFeatureSettings:     "font-feature-settings",
	PFontKerning:             "font-kerning",
	PFontLanguageOverride:    "font-language-override",
	PFontSize:                "font-size",
	PFontStretch:             "font-stretch",
	PFontStyle:               "font-style",
	PFontVariant:             "font-variant",
	PFontVariantAlternates:   "font-variant-alternates",
	PFontVariantCaps:         "font-variant-caps",
	PFontVariantEastAsian:    "font-variant-east-asian",
	PFontVariantLigatures:    "font-variant-ligatures",
	PFontVariantNumeric:      "font-variant-numeric",
	PFontVariantPosition:     "font-variant-position",
	PFontVariationSettings:   "font-variation-settings",
	PFontWeight:              "font-weight",
	PFootnoteDisplay:         "footnote-display",
	PFootnotePolicy:          "footnote-policy",
	PGridAutoColumns:         "grid-auto-columns",
	PGridAutoFlow:            "grid-auto-flow",
	PGridAutoRows:            "grid-auto-rows",
	PGridColumnEnd:           "grid-column-end",
	PGridColumnStart:         "grid-column-start",
	PGridRowEnd:              "grid-row-end",
	PGridRowStart:            "grid-row-start",
	PGridTemplateAreas:       "grid-template-areas",
	PGridTemplateColumns:     "grid-template-columns",
	PGridTemplateRows:        "grid-template-rows",
	PHeight:                  "height",
	PHyphenateCharacter:      "hyphenate-character",
	PHyphenateLimitChars:     "hyphenate-limit-chars",
	PHyphenateLimitZone:      "hyphenate-limit-zone",
	PHyphens:                 "hyphens",
	PImageOrientation:        "image-orientation",
	PImageRendering:          "image-rendering",
	PImageResolution:         "image-resolution",
	PJustifyContent:          "justify-content",
	PJustifyItems:            "justify-items",
	PJustifySelf:             "justify-self",
	PLang:                    "lang",
	PLeft:                    "left",
	PLetterSpacing:           "letter-spacing",
	PLineHeight:              "line-height",
	PLink:                    "link",
	PListStyleImage:          "list-style-image",
	PListStylePosition:       "list-style-position",
	PListStyleType:           "list-style-type",
	PMarginBottom:            "margin-bottom",
	PMarginBreak:             "margin-break",
	PMarginLeft:              "margin-left",
	PMarginRight:             "margin-right",
	PMarginTop:               "margin-top",
	PMarks:                   "marks",
	PMaxHeight:               "max-height",
	PMaxLines:                "max-lines",
	PMaxWidth:                "max-width",
	PMinHeight:               "min-height",
	PMinWidth:                "min-width",
	PObjectFit:               "object-fit",
	PObjectPosition:          "object-position",
	POpacity:                 "opacity",
	POrder:                   "order",
	POrphans:                 "orphans",
	POutlineColor:            "outline-color",
	POutlineStyle:            "outline-style",
	POutlineWidth:            "outline-width",
	POverflow:                "overflow",
	POverflowWrap:            "overflow-wrap",
	PPaddingBottom:           "padding-bottom",
	PPaddingLeft:             "padding-left",
	PPaddingRight:            "padding-right",
	PPaddingTop:              "padding-top",
	PPage:                    "page",
	PPosition:                "position",
	PQuotes:                  "quotes",
	PRight:                   "right",
	PRowGap:                  "row-gap",
	PSize:                    "size",
	PStringSet:               "string-set",
	PTabSize:                 "tab-size",
	PTableLayout:             "table-layout",
	PTextAlignAll:            "text-align-all",
	PTextAlignLast:           "text-align-last",
	PTextDecorationColor:     "text-decoration-color",
	PTextDecorationLine:      "text-decoration-line",
	PTextDecorationStyle:     "text-decoration-style",
	PTextIndent:              "text-indent",
	PTextOverflow:            "text-overflow",
	PTextTransform:           "text-transform",
	PTop:                     "top",
	PTransform:               "transform",
	PTransformOrigin:         "transform-origin",
	PUnicodeBidi:             "unicode-bidi",
	PVerticalAlign:           "vertical-align",
	PVisibility:              "visibility",
	PWhiteSpace:              "white-space",
	PWidows:                  "widows",
	PWidth:                   "width",
	PWordBreak:               "word-break",
	PWordSpacing:             "word-spacing",
	PZIndex:                  "z-index",
}

// PropsFromNames maps CSS property names to internal enum tags.
var PropsFromNames = map[string]KnownProp{
	"align-content":              PAlignContent,
	"align-items":                PAlignItems,
	"align-self":                 PAlignSelf,
	"anchor":                     PAnchor,
	"appearance":                 PAppearance,
	"background-attachment":      PBackgroundAttachment,
	"background-clip":            PBackgroundClip,
	"background-color":           PBackgroundColor,
	"background-image":           PBackgroundImage,
	"background-origin":          PBackgroundOrigin,
	"background-position":        PBackgroundPosition,
	"background-repeat":          PBackgroundRepeat,
	"background-size":            PBackgroundSize,
	"bleed-bottom":               PBleedBottom,
	"bleed-left":                 PBleedLeft,
	"bleed-right":                PBleedRight,
	"bleed-top":                  PBleedTop,
	"block-ellipsis":             PBlockEllipsis,
	"bookmark-label":             PBookmarkLabel,
	"bookmark-level":             PBookmarkLevel,
	"bookmark-state":             PBookmarkState,
	"border-bottom-color":        PBorderBottomColor,
	"border-bottom-left-radius":  PBorderBottomLeftRadius,
	"border-bottom-right-radius": PBorderBottomRightRadius,
	"border-bottom-style":        PBorderBottomStyle,
	"border-bottom-width":        PBorderBottomWidth,
	"border-collapse":            PBorderCollapse,
	"border-image-outset":        PBorderImageOutset,
	"border-image-repeat":        PBorderImageRepeat,
	"border-image-slice":         PBorderImageSlice,
	"border-image-source":        PBorderImageSource,
	"border-image-width":         PBorderImageWidth,
	"border-left-color":          PBorderLeftColor,
	"border-left-style":          PBorderLeftStyle,
	"border-left-width":          PBorderLeftWidth,
	"border-right-color":         PBorderRightColor,
	"border-right-style":         PBorderRightStyle,
	"border-right-width":         PBorderRightWidth,
	"border-spacing":             PBorderSpacing,
	"border-top-color":           PBorderTopColor,
	"border-top-left-radius":     PBorderTopLeftRadius,
	"border-top-right-radius":    PBorderTopRightRadius,
	"border-top-style":           PBorderTopStyle,
	"border-top-width":           PBorderTopWidth,
	"bottom":                     PBottom,
	"box-decoration-break":       PBoxDecorationBreak,
	"box-sizing":                 PBoxSizing,
	"break-after":                PBreakAfter,
	"break-before":               PBreakBefore,
	"break-inside":               PBreakInside,
	"caption-side":               PCaptionSide,
	"clear":                      PClear,
	"clip":                       PClip,
	"color":                      PColor,
	"column-count":               PColumnCount,
	"column-fill":                PColumnFill,
	"column-gap":                 PColumnGap,
	"column-rule-color":          PColumnRuleColor,
	"column-rule-style":          PColumnRuleStyle,
	"column-rule-width":          PColumnRuleWidth,
	"column-span":                PColumnSpan,
	"column-width":               PColumnWidth,
	"content":                    PContent,
	"continue":                   PContinue,
	"counter-increment":          PCounterIncrement,
	"counter-reset":              PCounterReset,
	"counter-set":                PCounterSet,
	"direction":                  PDirection,
	"display":                    PDisplay,
	"empty-cells":                PEmptyCells,
	"flex-basis":                 PFlexBasis,
	"flex-direction":             PFlexDirection,
	"flex-grow":                  PFlexGrow,
	"flex-shrink":                PFlexShrink,
	"flex-wrap":                  PFlexWrap,
	"float":                      PFloat,
	"font-family":                PFontFamily,
	"font-feature-settings":      PFontFeatureSettings,
	"font-kerning":               PFontKerning,
	"font-language-override":     PFontLanguageOverride,
	"font-size":                  PFontSize,
	"font-stretch":               PFontStretch,
	"font-style":                 PFontStyle,
	"font-variant":               PFontVariant,
	"font-variant-alternates":    PFontVariantAlternates,
	"font-variant-caps":          PFontVariantCaps,
	"font-variant-east-asian":    PFontVariantEastAsian,
	"font-variant-ligatures":     PFontVariantLigatures,
	"font-variant-numeric":       PFontVariantNumeric,
	"font-variant-position":      PFontVariantPosition,
	"font-variation-settings":    PFontVariationSettings,
	"font-weight":                PFontWeight,
	"footnote-display":           PFootnoteDisplay,
	"footnote-policy":            PFootnotePolicy,
	"grid-auto-columns":          PGridAutoColumns,
	"grid-auto-flow":             PGridAutoFlow,
	"grid-auto-rows":             PGridAutoRows,
	"grid-column-end":            PGridColumnEnd,
	"grid-column-start":          PGridColumnStart,
	"grid-row-end":               PGridRowEnd,
	"grid-row-start":             PGridRowStart,
	"grid-template-areas":        PGridTemplateAreas,
	"grid-template-columns":      PGridTemplateColumns,
	"grid-template-rows":         PGridTemplateRows,
	"height":                     PHeight,
	"hyphenate-character":        PHyphenateCharacter,
	"hyphenate-limit-chars":      PHyphenateLimitChars,
	"hyphenate-limit-zone":       PHyphenateLimitZone,
	"hyphens":                    PHyphens,
	"image-orientation":          PImageOrientation,
	"image-rendering":            PImageRendering,
	"image-resolution":           PImageResolution,
	"justify-content":            PJustifyContent,
	"justify-items":              PJustifyItems,
	"justify-self":               PJustifySelf,
	"lang":                       PLang,
	"left":                       PLeft,
	"letter-spacing":             PLetterSpacing,
	"line-height":                PLineHeight,
	"link":                       PLink,
	"list-style-image":           PListStyleImage,
	"list-style-position":        PListStylePosition,
	"list-style-type":            PListStyleType,
	"margin-bottom":              PMarginBottom,
	"margin-break":               PMarginBreak,
	"margin-left":                PMarginLeft,
	"margin-right":               PMarginRight,
	"margin-top":                 PMarginTop,
	"marks":                      PMarks,
	"max-height":                 PMaxHeight,
	"max-lines":                  PMaxLines,
	"max-width":                  PMaxWidth,
	"min-height":                 PMinHeight,
	"min-width":                  PMinWidth,
	"object-fit":                 PObjectFit,
	"object-position":            PObjectPosition,
	"opacity":                    POpacity,
	"order":                      POrder,
	"orphans":                    POrphans,
	"outline-color":              POutlineColor,
	"outline-style":              POutlineStyle,
	"outline-width":              POutlineWidth,
	"overflow":                   POverflow,
	"overflow-wrap":              POverflowWrap,
	"padding-bottom":             PPaddingBottom,
	"padding-left":               PPaddingLeft,
	"padding-right":              PPaddingRight,
	"padding-top":                PPaddingTop,
	"page":                       PPage,
	"position":                   PPosition,
	"quotes":                     PQuotes,
	"right":                      PRight,
	"row-gap":                    PRowGap,
	"size":                       PSize,
	"string-set":                 PStringSet,
	"tab-size":                   PTabSize,
	"table-layout":               PTableLayout,
	"text-align-all":             PTextAlignAll,
	"text-align-last":            PTextAlignLast,
	"text-decoration-color":      PTextDecorationColor,
	"text-decoration-line":       PTextDecorationLine,
	"text-decoration-style":      PTextDecorationStyle,
	"text-indent":                PTextIndent,
	"text-overflow":              PTextOverflow,
	"text-transform":             PTextTransform,
	"top":                        PTop,
	"transform":                  PTransform,
	"transform-origin":           PTransformOrigin,
	"unicode-bidi":               PUnicodeBidi,
	"vertical-align":             PVerticalAlign,
	"visibility":                 PVisibility,
	"white-space":                PWhiteSpace,
	"widows":                     PWidows,
	"width":                      PWidth,
	"word-break":                 PWordBreak,
	"word-spacing":               PWordSpacing,
	"z-index":                    PZIndex,
}
