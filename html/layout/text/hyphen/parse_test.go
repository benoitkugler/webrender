package hyphen

import (
	"bytes"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/encoding/unicode"
)

var parseHex = regexp.MustCompile(`\^{2}([0-9a-f]{2})`)

func assertNoHexEscape(datas fs.FS, filename string) error {
	b, err := fs.ReadFile(datas, filename)
	if err != nil {
		return err
	}

	lines := bytes.Split(b, []byte{'\n'})
	if len(lines) == 0 {
		return nil
	}

	header, patterns := lines[0], lines[1:]
	cs := strings.ToLower(strings.TrimSpace(string(header)))
	enco := encodings[cs]
	if enco == nil {
		enco = unicode.UTF8
	}
	dec := enco.NewDecoder()

	for _, line := range patterns {
		utf8Pattern, err := dec.Bytes(line)
		if err != nil {
			return fmt.Errorf("invalid pattern: %s (%s)", line, err)
		}
		pat := string(bytes.TrimSpace(utf8Pattern))

		if parseHex.MatchString(pat) {
			return fmt.Errorf("unsupported escape sequence in: %s", pat)
		}
	}
	return nil
}

func TestAssertNoHexEscape(t *testing.T) {
	for _, v := range languages {
		err := assertNoHexEscape(dictionaries, v)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParse(t *testing.T) {
	for _, v := range languages {
		_, err := parseHyphDic(dictionaries, v)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range languages {
			parseHyphDic(dictionaries, v)
		}
	}
}

func TestLanguages(t *testing.T) {
	l, err := getLanguages(dictionaries)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Languages available :", len(l))
}

var patternSamples = []string{
	"'anti1a2",
	".anti1e2",
	"'anti1e2",
	".anti1é2",
	"'anti1é2",
	".anti2en1ne",
	"'anti2en1ne",
	"anti1fe",
	"antifer1me",
	"antifer3ment.",
	".anti1s2",
	"'anti1s2",
	"a1po",
	".a1po",
	"'a1po",
	".apo2s3ta",
	"'apo2s3ta",
	"apo2s3t2r",
	"ap1pa",
	"appa1re",
	"appa3rent.",
	"ar1c2h",
	"arc1hi",
	"archié1pi",
	"archi1é2pis",
	".ar1de",
	".ar3dent.",
	".ar1ge",
	"'ar1ge",
	".ar3gent.",
	"'ar3gent.",
	"ar1me",
	"ar2ment.",
	"ar1mi",
	"armil5l",
	".ar1pe",
	"'ar1pe",
	".ar3pent.",
	"'ar3pent.",
	"as1me",
	"as2ment.",
	".as2ta",
	"'as2ta",
	"as1t2r",
	"a2s3t1ro",
	"au1me",
	"au2ment.",
	"a1vi",
	"avil4l",
	"1ba",
	".1ba",
	"1bâ",
	".bai1se",
	".baise1ma",
	".bai2se3main",
	"1be",
	"1bé",
	".ас1п",
	".ау2",
	".аш1х",
	".аэ2",
	".бе2з1а2",
	".бе2з1у2",
	".бе2з3о2",
	".бе2с1т",
	".без1на",
	".без1р",
	".би2б1л",
	".бу1г",
	".взъ2",
	".во1в2",
	".во2п1л",
	".во2с3тор",
	".во2ск",
	".во3п2ло",
	".воз1на",
	".вс6п",
	".въ2",
	".вып2ле",
	".выс2п",
	".гос1к",
	".дво2е",
	".де2зи",
	".ди2а",
	".ди2сто",
	".до1см",
	".за3в2ра",
	".за3п2н",
}

func TestParsePattern(t *testing.T) {
	toRune := func(s string) rune {
		if s == "" {
			return 0
		}
		return []rune(s)[0]
	}

	for _, data := range patternSamples {
		expecteds := parse.FindAllStringSubmatch(data, -1)
		gots := parsePattern(data, nil)
		if len(expecteds) != len(gots) {
			t.Fatalf("parsing failed for: %s: %d %d", data, len(expecteds), len(gots))
		}

		for i, exp := range expecteds {
			exp := [2]rune{toRune(exp[1]), toRune(exp[2])}
			got := gots[i]
			if exp != got {
				t.Fatalf("%v != %v", exp, got)
			}
		}
	}
}

var parse = regexp.MustCompile(`(\d?)(\D?)`)

func BenchmarkParsePattern(b *testing.B) {
	b.Run("regexp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, sample := range patternSamples {
				parse.FindAllStringSubmatch(sample, -1)
			}
		}
	})

	b.Run("manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, sample := range patternSamples {
				parsePattern(sample, nil)
			}
		}
	})
}
