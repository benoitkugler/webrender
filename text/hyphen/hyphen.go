// Package hyphen hyphenates text using existing Hunspell hyphenation dictionaries.
//
// This is a port of https://github.com/Kozea/Pyphen
package hyphen

import (
	"strings"
	"sync"
	"unicode"

	"github.com/benoitkugler/textlayout/language"
)

var (
	dictionariesCache     = map[string]hyphDicReference{}
	dictionariesCacheLock sync.Mutex
)

type Hyphener struct {
	hd          hyphDic
	left, right int
}

func NewHyphener(lang language.Language, left, right int) Hyphener {
	filename := languages[LanguageFallback(lang)]
	var out Hyphener
	out.left, out.right = left, right

	dictionariesCacheLock.Lock()
	defer dictionariesCacheLock.Unlock()

	if dic, ok := dictionariesCache[filename]; ok {
		out.hd.data = dic
	} else {
		dic, _ := parseHyphDic(dictionaries, filename) // Test assert thaht it wont fail
		dictionariesCache[filename] = dic
		out.hd.data = dic
	}

	out.hd.cache = make(map[string][]dataOrInt)
	return out
}

// Get a list of positions where the word can be hyphenated.
// See also `HyphDict.positions`. The points that are too far to the
// left or right are removed.
func (h Hyphener) positions(word []rune) []dataOrInt {
	right := len(word) - h.right
	var out []dataOrInt
	for _, index := range h.hd.positions(word) {
		if h.left <= index.V && index.V <= right {
			out = append(out, index)
		}
	}
	return out
}

// Iterates over all hyphenation possibilities, the longest first,
// for `word`.
// The returned slice contains the starts of each possibility.
func (h Hyphener) Iterate(word string) []string {
	word_ := []rune(word)
	pos := h.positions(word_)
	L := len(pos)
	out := make([]string, L)
	wordIsUpper := strings.IndexFunc(word, func(r rune) bool { return !unicode.IsUpper(r) }) == -1

	for i := L - 1; i >= 0; i-- { // reverse
		index := pos[i]
		var subs string
		if index.Data != nil { // get the nonstandard hyphenation data
			data := *index.Data
			data.Index += index.V
			c1, _ := data.Changes[0], data.Changes[1]
			if wordIsUpper {
				c1 = strings.ToUpper(c1)
			}
			subs = string(word_[:data.Index]) + c1
		} else {
			subs = string(word_[:index.V])
		}
		out[L-1-i] = subs
	}
	return out
}

// Iterates over all hyphenation possibilities, the longest first,
// for `word`.
// The returned slice contains the starts of each possibility.
func (h Hyphener) IterateRunes(word []rune) []string {
	pos := h.positions(word)
	L := len(pos)
	out := make([]string, L)
	wordIsUpper := true
	for _, r := range word {
		if !unicode.IsUpper(r) {
			wordIsUpper = false
		}
	}

	for i := L - 1; i >= 0; i-- { // reverse
		index := pos[i]
		var subs string
		if index.Data != nil { // get the nonstandard hyphenation data
			data := *index.Data
			data.Index += index.V
			c1, _ := data.Changes[0], data.Changes[1]
			if wordIsUpper {
				c1 = strings.ToUpper(c1)
			}
			subs = string(word[:data.Index]) + c1
		} else {
			subs = string(word[:index.V])
		}
		out[L-1-i] = subs
	}
	return out
}

type hyphDicReference struct {
	Patterns  map[string]pattern
	MaxLength int // in runes
}

type hyphDic struct {
	cache map[string][]dataOrInt
	data  hyphDicReference
}

// Get a list of positions where the word can be hyphenated.
//
// E.g. for the dutch word 'lettergrepen' this method returns `[3, 6, 9]`.
//
// Each position is a [dataOrInt] : if the data attribute is not nil,
// it contains information about nonstandard hyphenation at that point
func (dic hyphDic) positions(word_ []rune) []dataOrInt {
	word := strings.ToLower(string(word_))
	if points, ok := dic.cache[word]; ok {
		return points
	}
	pointedWord := []rune("." + word + ".")
	references := make([]dataOrInt, len(pointedWord)+1)

	for i := 0; i < len(pointedWord)-1; i++ {
		for j := i + 1; j <= i+dic.data.MaxLength && j <= len(pointedWord); j++ {
			pat, ok := dic.data.Patterns[string(pointedWord[i:j])]
			if ok {
				offset, values := pat.Start, pat.Values
				slice := references[i+offset : i+offset+len(values)]
				for k := range slice {
					max := slice[k]
					if values[k].V > slice[k].V {
						max = values[k]
					}
					slice[k] = max
				}
			}
		}
	}
	var points []dataOrInt
	for i, reference := range references {
		if reference.V%2 != 0 {
			points = append(points, dataOrInt{V: i - 1, Data: reference.Data})
		}
	}

	dic.cache[word] = points
	return points
}

type pattern struct {
	Values []dataOrInt
	Start  int
}

type dataOrInt struct {
	Data *complexHyphenation // optional
	V    int
}

// complexHyphenation stores information about nonstandard hyphenation at a point.
type complexHyphenation struct {
	//  a string like `'ff=f'`, that describes how hyphenation should
	//  take place.
	Changes [2]string
	//  where to substitute the change, counting from the current point
	Index int
	//  how many characters to remove while substituting the nonstandard
	//  hyphenation
	Cut int
}

// Get a fallback language available in our dictionaries.
//
// http://www.unicode.org/reports/tr35/#Locale_Inheritance
//
// We use the normal truncation inheritance. This function needs aliases
// including scripts for languages with multiple regions available.
func LanguageFallback(lang language.Language) language.Language {
	for _, lg := range lang.SimpleInheritance() {
		if _, ok := languages[lg]; ok {
			return lg
		}
	}
	return ""
}
