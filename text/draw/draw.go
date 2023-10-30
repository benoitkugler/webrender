// Package draw use a backend and a layout object to
// draw glyphs on the ouput.
package draw

import (
	"github.com/benoitkugler/webrender/backend"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

type Context struct {
	Output backend.Canvas         // where to draw the text
	Fonts  text.FontConfiguration // used to find fonts
}

// CreateFirstLine create the text for the first line of [layout], starting at position `(x,y)`.
// It also register the fonts used with [backend.Canvas.AddFont].
func (ctx Context) CreateFirstLine(layout text.EngineLayout, textOverflow string, blockEllipsis pr.TaggedString, x, y, angle pr.Fl,
) backend.TextDrawing {
	if layout, ok := layout.(*text.TextLayoutPango); ok {
		return ctx.createFirstLinePango(layout, textOverflow, blockEllipsis, x, y, angle)
	}
	return backend.TextDrawing{}
}

// DrawEmoji loads and draws `glyph` onto `dst`.
// It may be used by backend implementations to render emojis.
func DrawEmoji(font backend.Font, glyph backend.GID, extents backend.GlyphExtents,
	fontSize, x, y, xAdvance utils.Fl, dst backend.Canvas,
) {
	if pFont, ok := font.(*pangoFont); ok {
		drawEmojiPango(pFont, glyph, extents, fontSize, x, y, xAdvance, dst)
	}
}
