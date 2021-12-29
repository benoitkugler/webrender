package hyphen

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/language"
)

// def test_inserted():
//     """Test the ``inserted`` method."""
//     dic = pyphen.Pyphen(lang='nl_NL')
//     assert dic.inserted('lettergrepen') == 'let-ter-gre-pen'

// def test_wrap():
//     """Test the ``wrap`` method."""
//     dic = pyphen.Pyphen(lang='nl_NL')
//     assert dic.wrap('autobandventieldopje', 11) == (
//         'autoband-', 'ventieldopje')

func TestIterate(t *testing.T) {
	dic := NewHyphener(language.NewLanguage("nl_NL"), 2, 2)
	exp := []string{"Amster", "Am"}
	got := dic.Iterate("Amsterdam")
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestAlternative(t *testing.T) {
	dic := NewHyphener("hu", 1, 1)
	exp := []string{"kulisz", "ku"}
	got := dic.Iterate("kulissza")
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func TestFallbackDict(t *testing.T) {
	dic := NewHyphener(language.NewLanguage("nl_NL-variant"), 2, 2)
	exp := []string{"Amster", "Am"}
	got := dic.Iterate("Amsterdam")
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

// def test_personal_dict():
//     """Test a personal dict."""
//     dic = pyphen.Pyphen(lang='fr')
//     assert dic.inserted('autobandventieldopje') != 'au-to-band-ven-tiel-dop-je'
//     pyphen.LANGUAGES['fr'] = pyphen.LANGUAGES['nl_NL']
//     dic = pyphen.Pyphen(lang='fr')
//     assert dic.inserted('autobandventieldopje') == 'au-to-band-ven-tiel-dop-je'

// def test_left_right():
//     """Test the ``left`` and ``right`` parameters."""
//     dic = pyphen.Pyphen(lang='nl_NL')
//     assert dic.inserted('lettergrepen') == 'let-ter-gre-pen'
//     dic = pyphen.Pyphen(lang='nl_NL', left=4)
//     assert dic.inserted('lettergrepen') == 'letter-gre-pen'
//     dic = pyphen.Pyphen(lang='nl_NL', right=4)
//     assert dic.inserted('lettergrepen') == 'let-ter-grepen'
//     dic = pyphen.Pyphen(lang='nl_NL', left=4, right=4)
//     assert dic.inserted('lettergrepen') == 'letter-grepen'

// def test_filename():
//     """Test the ``filename`` parameter."""
//     dic = pyphen.Pyphen(filename=pyphen.LANGUAGES['nl_NL'])
//     assert dic.inserted('lettergrepen') == 'let-ter-gre-pen'

// def test_upper():
//     """Test uppercase."""
//     dic = pyphen.Pyphen(lang='nl_NL')
//     assert dic.inserted('LETTERGREPEN') == 'LET-TER-GRE-PEN'

// def test_upper_alternative():
//     """Test uppercase with alternative parser."""
//     dic = pyphen.Pyphen(lang='hu', left=1, right=1)
//     assert tuple(dic.iterate('KULISSZA')) == (
//         ('KULISZ', 'SZA'), ('KU', 'LISSZA'))
//     assert dic.inserted('KULISSZA') == 'KU-LISZ-SZA'

// def test_all_dictionaries():
//     """Test that all included dictionaries can be parsed."""
//     for lang in pyphen.LANGUAGES:
//         pyphen.Pyphen(lang=lang)

func TestFallback(t *testing.T) {
	if LanguageFallback(language.NewLanguage("en")) != language.NewLanguage("en") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("en_US")) != language.NewLanguage("en_US") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("en_FR")) != language.NewLanguage("en") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("sr-Latn")) != language.NewLanguage("sr_Latn") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("sr-Cyrl")) != language.NewLanguage("sr") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("fr-Latn-FR")) != language.NewLanguage("fr") {
		t.Fatal("unexpected fallback")
	}
	if LanguageFallback(language.NewLanguage("en-US_variant1-x")) != language.NewLanguage("en_US") {
		t.Fatal("unexpected fallback")
	}
}

func TestCache(t *testing.T) {
	dic := NewHyphener("fr", 2, 2)
	dic.Iterate("éléphant")
	if len(dic.hd.cache) == 0 {
		t.Fatal("missing cache")
	}
}

func TestUnicode(t *testing.T) {
	dictionary := NewHyphener("fr", 2, 2)
	res := dictionary.Iterate("hyphénation")
	if !reflect.DeepEqual(res, []string{"hyphéna", "hyphé", "hy"}) {
		t.Fatal()
	}
}
