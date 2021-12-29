package hyphen

import (
	"embed"
	"log"

	"github.com/benoitkugler/textlayout/language"
)

//go:embed dictionaries
var dictionaries embed.FS

var languages map[language.Language]string

func init() {
	var err error
	languages, err = getLanguages(dictionaries)
	if err != nil {
		log.Fatal(err)
	}
}
