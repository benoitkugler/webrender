package hyphen

import (
	"embed"
	"fmt"

	"github.com/benoitkugler/textlayout/language"
)

//go:embed dictionaries
var dictionaries embed.FS

var languages map[language.Language]string

func init() {
	var err error
	languages, err = getLanguages(dictionaries)
	if err != nil {
		panic(fmt.Errorf("hyphen: invalid embedded dict: %s", err))
	}
}
