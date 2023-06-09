package boxes

// Code generated by macros/boxes.py DO NOT EDIT

import "github.com/benoitkugler/webrender/html/tree"

// An atomic box in an inline formatting context.
// This inline-level box cannot be split for line breaks.
type AtomicInlineLevelBoxITF interface {
	InlineLevelBoxITF
	isAtomicInlineLevelBox()
}

// A block-level box that is also a block container.
// A non-replaced element with a ``display`` value of ``block``, ``list-item``
// generates a block box.
type BlockBoxITF interface {
	BlockContainerBoxITF
	BlockLevelBoxITF
	isBlockBox()
}

func (BlockBox) Type() BoxType        { return BlockT }
func (b *BlockBox) Box() *BoxFields   { return &b.BoxFields }
func (b BlockBox) Copy() Box          { return &b }
func (BlockBox) IsClassicalBox() bool { return true }
func (BlockBox) isBlockBox()          {}
func (BlockBox) isBlockContainerBox() {}
func (BlockBox) isBlockLevelBox()     {}
func (BlockBox) isParentBox()         {}

func BlockBoxAnonymousFrom(parent Box, children []Box) *BlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewBlockBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only block-level boxes or only line boxes.
// A box that either contains only block-level boxes or establishes an inline
// formatting context and thus contains only line boxes.
// A non-replaced element with a ``display`` value of ``block``,
// ``list-item``, ``inline-block`` or 'table-cell' generates a block container
// box.
type BlockContainerBoxITF interface {
	ParentBoxITF
	isBlockContainerBox()
}

// A box that participates in an block formatting context.
// An element with a ``display`` value of ``block``, ``list-item`` or
// ``table`` generates a block-level box.
type BlockLevelBoxITF interface {
	BoxITF
	isBlockLevelBox()
	methodsBlockLevelBox
}

// A box that is both replaced and block-level.
// A replaced element with a ``display`` value of ``block``, ``liste-item`` or
// ``table`` generates a block-level replaced box.
type BlockReplacedBoxITF interface {
	BlockLevelBoxITF
	ReplacedBoxITF
	isBlockReplacedBox()
}

func (BlockReplacedBox) Type() BoxType        { return BlockReplacedT }
func (b *BlockReplacedBox) Box() *BoxFields   { return &b.BoxFields }
func (b BlockReplacedBox) Copy() Box          { return &b }
func (BlockReplacedBox) IsClassicalBox() bool { return true }
func (BlockReplacedBox) isBlockReplacedBox()  {}
func (BlockReplacedBox) isBlockLevelBox()     {}
func (BlockReplacedBox) isReplacedBox()       {}

// A box that is both block-level and a flex container.
// It behaves as block on the outside and as a flex container on the inside.
type FlexBoxITF interface {
	BlockLevelBoxITF
	FlexContainerBoxITF
	isFlexBox()
}

func (FlexBox) Type() BoxType        { return FlexT }
func (b *FlexBox) Box() *BoxFields   { return &b.BoxFields }
func (b FlexBox) Copy() Box          { return &b }
func (FlexBox) IsClassicalBox() bool { return true }
func (FlexBox) isFlexBox()           {}
func (FlexBox) isBlockLevelBox()     {}
func (FlexBox) isFlexContainerBox()  {}
func (FlexBox) isParentBox()         {}

func FlexBoxAnonymousFrom(parent Box, children []Box) *FlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewFlexBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only flex-items.
type FlexContainerBoxITF interface {
	ParentBoxITF
	isFlexContainerBox()
}

// Box displaying footnotes, as defined in GCPM.
type FootnoteAreaBoxITF interface {
	BlockBoxITF
	isFootnoteAreaBox()
}

func (FootnoteAreaBox) Type() BoxType        { return FootnoteAreaT }
func (b *FootnoteAreaBox) Box() *BoxFields   { return &b.BoxFields }
func (b FootnoteAreaBox) Copy() Box          { return &b }
func (FootnoteAreaBox) IsClassicalBox() bool { return true }
func (FootnoteAreaBox) isFootnoteAreaBox()   {}
func (FootnoteAreaBox) isBlockBox()          {}
func (FootnoteAreaBox) isBlockContainerBox() {}
func (FootnoteAreaBox) isBlockLevelBox()     {}
func (FootnoteAreaBox) isParentBox()         {}

// A box that is both inline-level and a block container.
// It behaves as inline on the outside and as a block on the inside.
// A non-replaced element with a 'display' value of 'inline-block' generates
// an inline-block box.
type InlineBlockBoxITF interface {
	AtomicInlineLevelBoxITF
	BlockContainerBoxITF
	isInlineBlockBox()
}

func (InlineBlockBox) Type() BoxType           { return InlineBlockT }
func (b *InlineBlockBox) Box() *BoxFields      { return &b.BoxFields }
func (b InlineBlockBox) Copy() Box             { return &b }
func (InlineBlockBox) IsClassicalBox() bool    { return true }
func (InlineBlockBox) isInlineBlockBox()       {}
func (InlineBlockBox) isAtomicInlineLevelBox() {}
func (InlineBlockBox) isBlockContainerBox()    {}
func (InlineBlockBox) isInlineLevelBox()       {}
func (InlineBlockBox) isParentBox()            {}

func InlineBlockBoxAnonymousFrom(parent Box, children []Box) *InlineBlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewInlineBlockBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// An inline box with inline children.
// A box that participates in an inline formatting context and whose content
// also participates in that inline formatting context.
// A non-replaced element with a ``display`` value of ``inline`` generates an
// inline box.
type InlineBoxITF interface {
	InlineLevelBoxITF
	ParentBoxITF
	isInlineBox()
}

func (InlineBox) Type() BoxType        { return InlineT }
func (b *InlineBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineBox) Copy() Box          { return &b }
func (InlineBox) IsClassicalBox() bool { return true }
func (InlineBox) isInlineBox()         {}
func (InlineBox) isInlineLevelBox()    {}
func (InlineBox) isParentBox()         {}

func InlineBoxAnonymousFrom(parent Box, children []Box) *InlineBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewInlineBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that is both inline-level and a flex container.
// It behaves as inline on the outside and as a flex container on the inside.
type InlineFlexBoxITF interface {
	FlexContainerBoxITF
	InlineLevelBoxITF
	isInlineFlexBox()
}

func (InlineFlexBox) Type() BoxType        { return InlineFlexT }
func (b *InlineFlexBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineFlexBox) Copy() Box          { return &b }
func (InlineFlexBox) IsClassicalBox() bool { return true }
func (InlineFlexBox) isInlineFlexBox()     {}
func (InlineFlexBox) isFlexContainerBox()  {}
func (InlineFlexBox) isInlineLevelBox()    {}
func (InlineFlexBox) isParentBox()         {}

func InlineFlexBoxAnonymousFrom(parent Box, children []Box) *InlineFlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewInlineFlexBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that participates in an inline formatting context.
// An inline-level box that is not an inline box is said to be "atomic". Such
// boxes are inline blocks, replaced elements and inline tables.
// An element with a ``display`` value of ``inline``, ``inline-table``, or
// ``inline-block`` generates an inline-level box.
type InlineLevelBoxITF interface {
	BoxITF
	isInlineLevelBox()
}

// A box that is both replaced and inline-level.
// A replaced element with a ``display`` value of ``inline``,
// ``inline-table``, or ``inline-block`` generates an inline-level replaced
// box.
type InlineReplacedBoxITF interface {
	AtomicInlineLevelBoxITF
	ReplacedBoxITF
	isInlineReplacedBox()
}

func (InlineReplacedBox) Type() BoxType           { return InlineReplacedT }
func (b *InlineReplacedBox) Box() *BoxFields      { return &b.BoxFields }
func (b InlineReplacedBox) Copy() Box             { return &b }
func (InlineReplacedBox) IsClassicalBox() bool    { return true }
func (InlineReplacedBox) isInlineReplacedBox()    {}
func (InlineReplacedBox) isAtomicInlineLevelBox() {}
func (InlineReplacedBox) isInlineLevelBox()       {}
func (InlineReplacedBox) isReplacedBox()          {}

// Box for elements with ``display: inline-table``
type InlineTableBoxITF interface {
	TableBoxITF
	isInlineTableBox()
}

func (InlineTableBox) Type() BoxType        { return InlineTableT }
func (b *InlineTableBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineTableBox) Copy() Box          { return &b }
func (InlineTableBox) IsClassicalBox() bool { return true }
func (InlineTableBox) isInlineTableBox()    {}
func (InlineTableBox) isBlockLevelBox()     {}
func (InlineTableBox) isParentBox()         {}
func (InlineTableBox) isTableBox()          {}

func InlineTableBoxAnonymousFrom(parent Box, children []Box) *InlineTableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewInlineTableBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that represents a line in an inline formatting context.
// Can only contain inline-level boxes.
// In early stages of building the box tree a single line box contains many
// consecutive inline boxes. Later, during layout phase, each line boxes will
// be split into multiple line boxes, one for each actual line.
type LineBoxITF interface {
	ParentBoxITF
	isLineBox()
}

func (LineBox) Type() BoxType        { return LineT }
func (b *LineBox) Box() *BoxFields   { return &b.BoxFields }
func (b LineBox) Copy() Box          { return &b }
func (LineBox) IsClassicalBox() bool { return true }
func (LineBox) isLineBox()           {}
func (LineBox) isParentBox()         {}

// Box in page margins, as defined in CSS3 Paged Media
type MarginBoxITF interface {
	BlockContainerBoxITF
	isMarginBox()
}

func (MarginBox) Type() BoxType        { return MarginT }
func (b *MarginBox) Box() *BoxFields   { return &b.BoxFields }
func (b MarginBox) Copy() Box          { return &b }
func (MarginBox) IsClassicalBox() bool { return true }
func (MarginBox) isMarginBox()         {}
func (MarginBox) isBlockContainerBox() {}
func (MarginBox) isParentBox()         {}

// Box for a page.
// Initially the whole document will be in the box for the root element.
// During layout a new page box is created after every page break.
type PageBoxITF interface {
	ParentBoxITF
	isPageBox()
}

func (PageBox) Type() BoxType        { return PageT }
func (b *PageBox) Box() *BoxFields   { return &b.BoxFields }
func (b PageBox) Copy() Box          { return &b }
func (PageBox) IsClassicalBox() bool { return true }
func (PageBox) isPageBox()           {}
func (PageBox) isParentBox()         {}

// A box that has children.
type ParentBoxITF interface {
	BoxITF
	isParentBox()
}

// A box whose content is replaced.
// For example, ``<img>`` are replaced: their content is rendered externally
// and is opaque from CSS’s point of view.
type ReplacedBoxITF interface {
	BoxITF
	isReplacedBox()
	methodsReplacedBox
}

func (ReplacedBox) Type() BoxType        { return ReplacedT }
func (b *ReplacedBox) Box() *BoxFields   { return &b.BoxFields }
func (b ReplacedBox) Copy() Box          { return &b }
func (ReplacedBox) IsClassicalBox() bool { return true }
func (ReplacedBox) isReplacedBox()       {}

// Box for elements with ``display: table``
type TableBoxITF interface {
	BlockLevelBoxITF
	ParentBoxITF
	isTableBox()
	methodsTableBox
}

func (TableBox) Type() BoxType        { return TableT }
func (b *TableBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableBox) Copy() Box          { return &b }
func (TableBox) IsClassicalBox() bool { return true }
func (TableBox) isTableBox()          {}
func (TableBox) isBlockLevelBox()     {}
func (TableBox) isParentBox()         {}

func TableBoxAnonymousFrom(parent Box, children []Box) *TableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-caption``
type TableCaptionBoxITF interface {
	BlockBoxITF
	isTableCaptionBox()
}

func (TableCaptionBox) Type() BoxType        { return TableCaptionT }
func (b *TableCaptionBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableCaptionBox) Copy() Box          { return &b }
func (TableCaptionBox) IsClassicalBox() bool { return true }
func (TableCaptionBox) isTableCaptionBox()   {}
func (TableCaptionBox) isBlockBox()          {}
func (TableCaptionBox) isBlockContainerBox() {}
func (TableCaptionBox) isBlockLevelBox()     {}
func (TableCaptionBox) isParentBox()         {}

func TableCaptionBoxAnonymousFrom(parent Box, children []Box) *TableCaptionBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableCaptionBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-cell``
type TableCellBoxITF interface {
	BlockContainerBoxITF
	isTableCellBox()
}

func (TableCellBox) Type() BoxType        { return TableCellT }
func (b *TableCellBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableCellBox) Copy() Box          { return &b }
func (TableCellBox) IsClassicalBox() bool { return true }
func (TableCellBox) isTableCellBox()      {}
func (TableCellBox) isBlockContainerBox() {}
func (TableCellBox) isParentBox()         {}

func TableCellBoxAnonymousFrom(parent Box, children []Box) *TableCellBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableCellBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-column``
type TableColumnBoxITF interface {
	ParentBoxITF
	isTableColumnBox()
}

func (TableColumnBox) Type() BoxType        { return TableColumnT }
func (b *TableColumnBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableColumnBox) Copy() Box          { return &b }
func (TableColumnBox) IsClassicalBox() bool { return true }
func (TableColumnBox) isTableColumnBox()    {}
func (TableColumnBox) isParentBox()         {}

func TableColumnBoxAnonymousFrom(parent Box, children []Box) *TableColumnBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableColumnBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-column-group``
type TableColumnGroupBoxITF interface {
	ParentBoxITF
	isTableColumnGroupBox()
}

func (TableColumnGroupBox) Type() BoxType          { return TableColumnGroupT }
func (b *TableColumnGroupBox) Box() *BoxFields     { return &b.BoxFields }
func (b TableColumnGroupBox) Copy() Box            { return &b }
func (TableColumnGroupBox) IsClassicalBox() bool   { return true }
func (TableColumnGroupBox) isTableColumnGroupBox() {}
func (TableColumnGroupBox) isParentBox()           {}

func TableColumnGroupBoxAnonymousFrom(parent Box, children []Box) *TableColumnGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableColumnGroupBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-row``
type TableRowBoxITF interface {
	ParentBoxITF
	isTableRowBox()
}

func (TableRowBox) Type() BoxType        { return TableRowT }
func (b *TableRowBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableRowBox) Copy() Box          { return &b }
func (TableRowBox) IsClassicalBox() bool { return true }
func (TableRowBox) isTableRowBox()       {}
func (TableRowBox) isParentBox()         {}

func TableRowBoxAnonymousFrom(parent Box, children []Box) *TableRowBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableRowBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-row-group``
type TableRowGroupBoxITF interface {
	ParentBoxITF
	isTableRowGroupBox()
}

func (TableRowGroupBox) Type() BoxType        { return TableRowGroupT }
func (b *TableRowGroupBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableRowGroupBox) Copy() Box          { return &b }
func (TableRowGroupBox) IsClassicalBox() bool { return true }
func (TableRowGroupBox) isTableRowGroupBox()  {}
func (TableRowGroupBox) isParentBox()         {}

func TableRowGroupBoxAnonymousFrom(parent Box, children []Box) *TableRowGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
	out := NewTableRowGroupBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only text and has no box children.
// Any text in the document ends up in a text box. What CSS calls "anonymous
// inline boxes" are also text boxes.
type TextBoxITF interface {
	InlineLevelBoxITF
	isTextBox()
}

func (TextBox) Type() BoxType        { return TextT }
func (b *TextBox) Box() *BoxFields   { return &b.BoxFields }
func (b TextBox) Copy() Box          { return &b }
func (TextBox) IsClassicalBox() bool { return true }
func (TextBox) isTextBox()           {}
func (TextBox) isInlineLevelBox()    {}

// BoxType represents a box type.
type BoxType uint8

const (
	invalidType BoxType = iota
	AtomicInlineLevelT
	BlockT
	BlockContainerT
	BlockLevelT
	BlockReplacedT
	T
	FlexT
	FlexContainerT
	FootnoteAreaT
	InlineBlockT
	InlineT
	InlineFlexT
	InlineLevelT
	InlineReplacedT
	InlineTableT
	LineT
	MarginT
	PageT
	ParentT
	ReplacedT
	TableT
	TableCaptionT
	TableCellT
	TableColumnT
	TableColumnGroupT
	TableRowT
	TableRowGroupT
	TextT
)

// Returns true is the box is an instance of t.
func (t BoxType) IsInstance(box BoxITF) bool {
	var isInstance bool
	switch t {
	case AtomicInlineLevelT:
		_, isInstance = box.(AtomicInlineLevelBoxITF)
	case BlockT:
		_, isInstance = box.(BlockBoxITF)
	case BlockContainerT:
		_, isInstance = box.(BlockContainerBoxITF)
	case BlockLevelT:
		_, isInstance = box.(BlockLevelBoxITF)
	case BlockReplacedT:
		_, isInstance = box.(BlockReplacedBoxITF)
	case T:
		_, isInstance = box.(BoxITF)
	case FlexT:
		_, isInstance = box.(FlexBoxITF)
	case FlexContainerT:
		_, isInstance = box.(FlexContainerBoxITF)
	case FootnoteAreaT:
		_, isInstance = box.(FootnoteAreaBoxITF)
	case InlineBlockT:
		_, isInstance = box.(InlineBlockBoxITF)
	case InlineT:
		_, isInstance = box.(InlineBoxITF)
	case InlineFlexT:
		_, isInstance = box.(InlineFlexBoxITF)
	case InlineLevelT:
		_, isInstance = box.(InlineLevelBoxITF)
	case InlineReplacedT:
		_, isInstance = box.(InlineReplacedBoxITF)
	case InlineTableT:
		_, isInstance = box.(InlineTableBoxITF)
	case LineT:
		_, isInstance = box.(LineBoxITF)
	case MarginT:
		_, isInstance = box.(MarginBoxITF)
	case PageT:
		_, isInstance = box.(PageBoxITF)
	case ParentT:
		_, isInstance = box.(ParentBoxITF)
	case ReplacedT:
		_, isInstance = box.(ReplacedBoxITF)
	case TableT:
		_, isInstance = box.(TableBoxITF)
	case TableCaptionT:
		_, isInstance = box.(TableCaptionBoxITF)
	case TableCellT:
		_, isInstance = box.(TableCellBoxITF)
	case TableColumnT:
		_, isInstance = box.(TableColumnBoxITF)
	case TableColumnGroupT:
		_, isInstance = box.(TableColumnGroupBoxITF)
	case TableRowT:
		_, isInstance = box.(TableRowBoxITF)
	case TableRowGroupT:
		_, isInstance = box.(TableRowGroupBoxITF)
	case TextT:
		_, isInstance = box.(TextBoxITF)
	}
	return isInstance
}

func (t BoxType) String() string {
	switch t {
	case AtomicInlineLevelT:
		return "AtomicInlineLevelBox"
	case BlockT:
		return "BlockBox"
	case BlockContainerT:
		return "BlockContainerBox"
	case BlockLevelT:
		return "BlockLevelBox"
	case BlockReplacedT:
		return "BlockReplacedBox"
	case T:
		return "Box"
	case FlexT:
		return "FlexBox"
	case FlexContainerT:
		return "FlexContainerBox"
	case FootnoteAreaT:
		return "FootnoteAreaBox"
	case InlineBlockT:
		return "InlineBlockBox"
	case InlineT:
		return "InlineBox"
	case InlineFlexT:
		return "InlineFlexBox"
	case InlineLevelT:
		return "InlineLevelBox"
	case InlineReplacedT:
		return "InlineReplacedBox"
	case InlineTableT:
		return "InlineTableBox"
	case LineT:
		return "LineBox"
	case MarginT:
		return "MarginBox"
	case PageT:
		return "PageBox"
	case ParentT:
		return "ParentBox"
	case ReplacedT:
		return "ReplacedBox"
	case TableT:
		return "TableBox"
	case TableCaptionT:
		return "TableCaptionBox"
	case TableCellT:
		return "TableCellBox"
	case TableColumnT:
		return "TableColumnBox"
	case TableColumnGroupT:
		return "TableColumnGroupBox"
	case TableRowT:
		return "TableRowBox"
	case TableRowGroupT:
		return "TableRowGroupBox"
	case TextT:
		return "TextBox"
	}
	return "<invalid box type>"
}

var (
	_ BlockBoxITF            = (*BlockBox)(nil)
	_ BlockReplacedBoxITF    = (*BlockReplacedBox)(nil)
	_ FlexBoxITF             = (*FlexBox)(nil)
	_ FootnoteAreaBoxITF     = (*FootnoteAreaBox)(nil)
	_ InlineBlockBoxITF      = (*InlineBlockBox)(nil)
	_ InlineBoxITF           = (*InlineBox)(nil)
	_ InlineFlexBoxITF       = (*InlineFlexBox)(nil)
	_ InlineReplacedBoxITF   = (*InlineReplacedBox)(nil)
	_ InlineTableBoxITF      = (*InlineTableBox)(nil)
	_ LineBoxITF             = (*LineBox)(nil)
	_ MarginBoxITF           = (*MarginBox)(nil)
	_ PageBoxITF             = (*PageBox)(nil)
	_ ReplacedBoxITF         = (*ReplacedBox)(nil)
	_ TableBoxITF            = (*TableBox)(nil)
	_ TableCaptionBoxITF     = (*TableCaptionBox)(nil)
	_ TableCellBoxITF        = (*TableCellBox)(nil)
	_ TableColumnBoxITF      = (*TableColumnBox)(nil)
	_ TableColumnGroupBoxITF = (*TableColumnGroupBox)(nil)
	_ TableRowBoxITF         = (*TableRowBox)(nil)
	_ TableRowGroupBoxITF    = (*TableRowGroupBox)(nil)
	_ TextBoxITF             = (*TextBox)(nil)
)

func (t BoxType) AnonymousFrom(parent Box, children []Box) Box {
	switch t {
	case BlockT:
		return BlockBoxAnonymousFrom(parent, children)
	case FlexT:
		return FlexBoxAnonymousFrom(parent, children)
	case InlineBlockT:
		return InlineBlockBoxAnonymousFrom(parent, children)
	case InlineT:
		return InlineBoxAnonymousFrom(parent, children)
	case InlineFlexT:
		return InlineFlexBoxAnonymousFrom(parent, children)
	case InlineTableT:
		return InlineTableBoxAnonymousFrom(parent, children)
	case TableT:
		return TableBoxAnonymousFrom(parent, children)
	case TableCaptionT:
		return TableCaptionBoxAnonymousFrom(parent, children)
	case TableCellT:
		return TableCellBoxAnonymousFrom(parent, children)
	case TableColumnT:
		return TableColumnBoxAnonymousFrom(parent, children)
	case TableColumnGroupT:
		return TableColumnGroupBoxAnonymousFrom(parent, children)
	case TableRowT:
		return TableRowBoxAnonymousFrom(parent, children)
	case TableRowGroupT:
		return TableRowGroupBoxAnonymousFrom(parent, children)
	}
	return nil
}
