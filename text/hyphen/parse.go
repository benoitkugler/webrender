package hyphen

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/benoitkugler/textlayout/language"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

// Copyright 2008 - Wilbert Berendsen <info@wilbertberendsen.nl>
// Copyright 2012-2013 - Guillaume Ayoub <guillaume.ayoub@kozea.fr>

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

func parseHyphDic(datas fs.FS, filename string) (out hyphDicReference, err error) {
	b, err := fs.ReadFile(datas, filename)
	if err != nil {
		return out, err
	}

	lines := bytes.Split(b, []byte{'\n'})
	if len(lines) == 0 {
		return out, nil
	}

	header, patterns := lines[0], lines[1:]
	cs := strings.ToLower(strings.TrimSpace(string(header)))
	enco := encodings[cs]
	if enco == nil {
		enco = unicode.UTF8
	}
	dec := enco.NewDecoder()

	out.Patterns = make(map[string]pattern, len(patterns)/2)
	var (
		tags    []string
		values  []dataOrInt
		matches [][2]rune
	)
	for _, line := range patterns {
		utf8Pattern, err := dec.Bytes(line)
		if err != nil {
			return out, fmt.Errorf("invalid pattern: %s (%s)", line, err)
		}
		pat := string(bytes.TrimSpace(utf8Pattern))
		if pat == "" || strings.HasPrefix(pat, "%") || strings.HasPrefix(pat, "#") || strings.HasPrefix(pat, "LEFTHYPHENMIN") ||
			strings.HasPrefix(pat, "RIGHTHYPHENMIN") || strings.HasPrefix(pat, "COMPOUNDLEFTHYPHENMIN") || strings.HasPrefix(pat, "COMPOUNDRIGHTHYPHENMIN") {
			continue
		}

		// read nonstandard hyphen alternatives
		var factory parser
		if strings.ContainsRune(pat, '/') && strings.ContainsRune(pat, '=') {
			ls := strings.SplitN(pat, "/", 2)
			pat = ls[0]
			factory, err = newAlternativeParser(pat, ls[1])
			if err != nil {
				return out, err
			}
		} else {
			factory = defaultParser{}
		}

		matches = matches[:0]
		tags = tags[:0]
		values = values[:0]

		matches = parsePattern(pat, matches)
		tags = append(tags, make([]string, len(matches))...)
		values = append(values, make([]dataOrInt, len(matches))...)

		for j, match := range matches {
			i, str := match[0], match[1]
			if i == 0 {
				i = '0'
			}
			v := factory.parse(i)
			values[j] = v
			if str == 0 {
				tags[j] = ""
			} else {
				tags[j] = string(str)
			}
		}

		// if only zeros, skip this pattern
		if max(values) == 0 {
			continue
		}

		// chop zeros from beginning and end, and store start offset
		start, end := 0, len(values)
		for values[start].V == 0 {
			start += 1
		}
		for values[end-1].V == 0 {
			end -= 1
		}

		key := strings.Join(tags, "")

		valuesOut := make([]dataOrInt, end-start)
		copy(valuesOut, values[start:end])

		out.Patterns[key] = pattern{Start: start, Values: valuesOut}

		if len([]rune(key)) > out.MaxLength {
			out.MaxLength = len(tags)
		}
	}

	return out, nil
}

func max(values []dataOrInt) int {
	ma := math.MinInt32
	for _, val := range values {
		if val.V > ma {
			ma = val.V
		}
	}
	return ma
}

type parser interface {
	parse(c rune) dataOrInt
}

type defaultParser struct{}

func (defaultParser) parse(c rune) dataOrInt {
	v := int(c - '0')
	return dataOrInt{V: v}
}

type alternativeParser struct {
	complexHyphenation
}

func newAlternativeParser(pattern, alternative string) (*alternativeParser, error) {
	alternatives := strings.Split(alternative, ",")
	change := alternatives[0]
	changes := strings.Split(change, "=")
	if len(changes) != 2 {
		return nil, fmt.Errorf("invalid change instruction: %s", change)
	}
	index, err := strconv.Atoi(alternatives[1])
	if err != nil {
		return nil, err
	}
	cut, err := strconv.Atoi(alternatives[2])
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(pattern, ".") {
		index += 1
	}
	return &alternativeParser{complexHyphenation: complexHyphenation{
		Changes: [2]string{changes[0], changes[1]},
		Index:   index, Cut: cut,
	}}, nil
}

func (p *alternativeParser) parse(c rune) dataOrInt {
	p.Index = -1
	v := int(c - '0')
	var data *complexHyphenation
	if v&1 != 0 {
		tmp := p.complexHyphenation
		data = &tmp // copy
	}
	return dataOrInt{V: v, Data: data}
}

func getLanguages(dir embed.FS) (map[language.Language]string, error) {
	l, err := fs.ReadDir(dir, "dictionaries")
	if err != nil {
		return nil, err
	}

	out := map[language.Language]string{}

	for _, file := range l {
		filename := file.Name()
		if !strings.HasSuffix(filename, ".dic") {
			continue
		}

		name := language.NewLanguage(filename[5 : len(filename)-4])
		fullPath := filepath.Join("dictionaries", filename)

		out[name] = fullPath
		shortName := language.NewLanguage(strings.Split(string(name), "-")[0])
		if _, ok := out[shortName]; !ok {
			out[shortName] = fullPath
		}
	}

	return out, nil
}

func parsePattern(pat string, out [][2]rune) [][2]rune {
	var lastNumber rune
	for _, c := range pat {
		if '0' <= c && c <= '9' {
			if lastNumber != 0 { // two consective digits
				out = append(out, [2]rune{lastNumber, 0})
			}
			lastNumber = c
		} else {
			out = append(out, [2]rune{lastNumber, c})
			lastNumber = 0
		}
	}
	// check potential last alone number
	if lastNumber != 0 {
		out = append(out, [2]rune{lastNumber, 0})
	}
	return out
}
