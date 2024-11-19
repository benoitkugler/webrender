package tree

import (
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// Return the boolean evaluation of `queryList` for the given
// `deviceMediaType`.
func evaluateMediaQuery(queryList []string, deviceMediaType string) bool {
	// TODO: actual support for media queries, not just media types
	for _, query := range queryList {
		if query == "all" || query == deviceMediaType {
			return true
		}
	}
	return false
}

func parseMediaQuery(tokens []Token) []string {
	tokens = parser.RemoveWhitespace(tokens)
	if len(tokens) == 0 {
		return []string{"all"}
	} else {
		var media []string
		for _, part := range parser.SplitOnComma(tokens) {
			if len(part) == 1 {
				if ident, ok := part[0].(parser.Ident); ok {
					media = append(media, utils.AsciiLower(ident.Value))
					continue
				}
			}

			logger.WarningLogger.Printf("Expected a media type, got %s", parser.Serialize(part))
			return nil
		}
		return media
	}
}
