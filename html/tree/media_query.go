package tree

import (
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/logger"
)

// Return the boolean evaluation of `queryList` for the given
// `deviceMediaType`.
func evaluateMediaQuery(queryList []string, deviceMediaType string) bool {
	// TODO: actual support for media queries, not just media types
	for _, query := range queryList {
		if "all" == query || deviceMediaType == query {
			return true
		}
	}
	return false
}

func parseMediaQuery(tokens []Token) []string {
	tokens = validation.RemoveWhitespace(tokens)
	if len(tokens) == 0 {
		return []string{"all"}
	} else {
		var media []string
		for _, part := range validation.SplitOnComma(tokens) {
			if len(part) == 1 {
				if ident, ok := part[0].(parser.IdentToken); ok {
					media = append(media, ident.Value.Lower())
					continue
				}
			}

			logger.WarningLogger.Printf("Expected a media type, got %s", parser.Serialize(part))
			return nil
		}
		return media
	}
}
