package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-text/typesetting/language"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

var encodings = map[string]encoding.Encoding{
	"utf-8":            unicode.UTF8,
	"cp1251":           charmap.Windows1251,
	"microsoft-cp1251": charmap.Windows1251,
	"iso8859-1":        charmap.ISO8859_1,
	"iso8859-2":        charmap.ISO8859_2,
	"iso8859-5":        charmap.ISO8859_5,
	"iso8859-7":        charmap.ISO8859_7,
	"iso8859-13":       charmap.ISO8859_13,
	"iso8859-15":       charmap.ISO8859_15,
}

func main() {
	contents, map_, err := getLanguages()
	if err != nil {
		panic(err)
	}

	fmt.Println("Languages available :", len(contents))

	for _, content := range contents {
		err = os.WriteFile("../"+content.path(), []byte(content.content), os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	outFile, err := os.Create("../datas_gen.go")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, `package hyphen
	
	import (
		"embed"

		"github.com/go-text/typesetting/language"
	)

	//go:embed dictionaries
	var dictionaries embed.FS

	`)

	var keys []language.Language
	for k := range map_ {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	fmt.Fprintln(outFile, "var languages = map[language.Language]string{")
	for _, l := range keys {
		index := map_[l]
		fmt.Fprintf(outFile, "%q: %q,\n", l, contents[index].path())
	}
	fmt.Fprintln(outFile, "}")
}

//go:embed dictionaries
var dictionaries embed.FS

type alphabet struct {
	name    string
	content string
}

func (al alphabet) path() string {
	return filepath.Join("dictionaries", al.name+".dic")
}

func getLanguages() ([]alphabet, map[language.Language]int, error) {
	l, err := fs.ReadDir(dictionaries, "dictionaries")
	if err != nil {
		return nil, nil, err
	}

	var contents []alphabet
	languageToContent := map[language.Language]int{}

	for _, file := range l {
		filename := file.Name()
		if !strings.HasSuffix(filename, ".dic") {
			continue
		}
		index := len(contents)

		languageName := language.NewLanguage(filename[5 : len(filename)-4])

		languageToContent[languageName] = index
		shortName := language.NewLanguage(strings.Split(string(languageName), "-")[0])
		if _, ok := languageToContent[shortName]; !ok {
			languageToContent[shortName] = index
		}

		fullPath := filepath.Join("dictionaries", filename)
		b, err := fs.ReadFile(dictionaries, fullPath)
		if err != nil {
			return nil, nil, err
		}
		firstLine, remaining, _ := bytes.Cut(b, []byte{'\n'})
		dec := getEncoding(firstLine)

		utf8, err := dec.Bytes(remaining)
		if err != nil {
			return nil, nil, err
		}
		contents = append(contents, alphabet{name: strings.ReplaceAll(string(languageName), "-", "_"), content: string(utf8)})
	}

	return contents, languageToContent, nil
}

func getEncoding(firstLine []byte) *encoding.Decoder {
	cs := strings.ToLower(strings.TrimSpace(string(firstLine)))
	enco := encodings[cs]
	if enco == nil {
		enco = unicode.UTF8
	}
	return enco.NewDecoder()
}
