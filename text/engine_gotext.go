package text

import (
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/go-text/typesetting/segmenter"
)

type FontConfigurationGotext struct {
	segmenter segmenter.Segmenter
}

func (fc *FontConfigurationGotext) CanBreakText(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	fc.segmenter.Init(t)
	iter := fc.segmenter.LineIterator()
	if iter.Next() {
		line := iter.Line()
		end := line.Offset + len(line.Text)
		if end < len(t) {
			return pr.True
		}
	}
	return pr.False
}
