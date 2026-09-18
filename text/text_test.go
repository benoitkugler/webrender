package text

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	pr "github.com/benoitkugler/webrender/css/properties"
	"github.com/benoitkugler/webrender/css/validation"
	"github.com/benoitkugler/webrender/text/hyphen"
	"github.com/benoitkugler/webrender/utils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/fontscan"
)

var (
	sansFonts = pr.Strings{"DejaVu Sans", "sans"}
	monoFonts = pr.Strings{"DejaVu Sans Mono", "monospace"}
)

const fontmapCache = "testdata/cache.fc"

var (
	fontmapPango  *fcfonts.FontMap
	fontmapGotext *fontscan.FontMap
)

var textSamples = [...]string{
	"نص حكيم له سر قاطع وذو شأن عظيم مكتوب على ثوب أخضر ومغلف بجلد أزرق.",
	"আমি কাঁচ খেতে পারি, তাতে আমার কোনো ক্ষতি হয় না।",

	"Ek kan glas eet, maar dit doen my nie skade nie.",
	"Gvxam mincetu apocikvyeh: ñizol ce mamvj ka raq kuse bafkeh mew.",
	"I koh Glos esa, und es duard ma ned wei.",
	"Под южно дърво, цъфтящо в синьо, бягаше малко пухкаво зайче.",
	"Mi save kakae glas, hemi no save katem mi.",
	"ཤེལ་སྒོ་ཟ་ནས་ང་ན་གི་མ་རེད།",
	"Fin džip, gluh jež i čvrst konjić dođoše bez moljca.",
	"Jove xef, porti whisky amb quinze glaçons d'hidrogen, coi!",
	"Siña yo' chumocho krestat, ti ha na'lalamen yo'.",
	"Příliš žluťoučký kůň úpěl ďábelské ódy.",
	"Dw i'n gallu bwyta gwydr, 'dyw e ddim yn gwneud dolur i mi.",
	"Quizdeltagerne spiste jordbær med fløde, mens cirkusklovnen Walther spillede på xylofon.",
	"Zwölf Boxkämpfer jagen Viktor quer über den großen Sylter Deich.",
	"މާއްދާ 1 – ހުރިހާ އިންސާނުން ވެސް އުފަންވަނީ، ދަރަޖަ އާއި ޙައްޤު ތަކުގައި މިނިވަންކަމާއި ހަމަހަމަކަން ލިބިގެންވާ ބައެއްގެ ގޮތުގައެވެ.",
	"Θέλει αρετή και τόλμη η ελευθερία. (Ανδρέας Κάλβος)",
	"The quick brown fox jumps over the lazy dog.",
	"Ich canne glas eten and hit hirtiþ me nouȝt.",
	"Eĥoŝanĝo ĉiuĵaŭde.",
	"Jovencillo emponzoñado de whisky: ¡qué figurota exhibe!",
	"See väike mölder jõuab rongile hüpata.",
	"Kristala jan dezaket, ez dit minik ematen.",
	"«الا یا اَیُّها السّاقی! اَدِرْ کَأساً وَ ناوِلْها!» که عشق آسان نمود اوّل، ولی افتاد مشکل‌ها!",
	"Viekas kettu punaturkki laiskan koiran takaa kurkki.",
	"Voix ambiguë d'un cœur qui, au zéphyr, préfère les jattes de kiwis.",
	"Je puis mangier del voirre. Ne me nuit.",
	"Chuaigh bé mhórshách le dlúthspád fíorfhinn trí hata mo dhea-phorcáin bhig.",
	"S urrainn dhomh gloinne ithe; cha ghoirtich i mi.",
	"Eu podo xantar cristais e non cortarme.",
	"𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸.",
	"Foddym gee glonney agh cha jean eh gortaghey mee.",
	"Hiki iaʻu ke ʻai i ke aniani; ʻaʻole nō lā au e ʻeha.",
	"דג סקרן שט לו בים זך אך לפתע פגש חבורה נחמדה שצצה כך.",
	"नहीं नजर किसी की बुरी नहीं किसी का मुँह काला जो करे सो उपर वाला",
	"Deblji krojač: zgužvah smeđ filc u tanjušni džepić.",
	"Egy hűtlen vejét fülöncsípő, dühös mexikói úr Wesselényinél mázol Quitóban.",
	"Կրնամ ապակի ուտել և ինծի անհանգիստ չըներ։",
	"Kæmi ný öxi hér ykist þjófum nú bæði víl og ádrepa",
	"Ma la volpe, col suo balzo, ha raggiunto il quieto Fido.",
	"Chruu, a kwik di kwik brong fox a jomp huova di liezi daag de, yu no siit?",
	"Aku isa mangan beling tanpa lara.",
	"მინას ვჭამ და არა მტკივა.",
	"ನಾನು ಗಾಜನ್ನು ತಿನ್ನಬಲ್ಲೆ ಮತ್ತು ಅದರಿಂದ ನನಗೆ ನೋವಾಗುವುದಿಲ್ಲ.",
	"다람쥐 헌 쳇바퀴에 타고파",
	"Mý a yl dybry gwéder hag éf ny wra ow ankenya.",
	"Sic surgens, dux, zelotypos quam karus haberis",
	"Įlinkdama fechtuotojo špaga sublykčiojusi pragręžė apvalų arbūzą.",
	"Sarkanās jūrascūciņas peld pa jūru.",
	"E koʻana e kai i te karahi, mea ʻā, ʻaʻe hauhau.",
	"Можам да јадам стакло, а не ме штета.",
	"വേദനയില്ലാതെ കുപ്പിചില്ലു് എനിയ്ക്കു് കഴിയ്ക്കാം.",
	"ᠪᠢ ᠰᠢᠯᠢ ᠢᠳᠡᠶᠦ ᠴᠢᠳᠠᠨᠠ ᠂ ᠨᠠᠳᠤᠷ ᠬᠣᠤᠷᠠᠳᠠᠢ ᠪᠢᠰᠢ",
	"मी काच खाऊ शकतो, मला ते दुखत नाही.",
	"Saya boleh makan kaca dan ia tidak mencederakan saya.",
	"M' pozz magna' o'vetr, e nun m' fa mal.",
	"Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi.",
	"Eg kan eta glas utan å skada meg.",
	"Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi.",
	"Tsésǫʼ yishą́ągo bííníshghah dóó doo shił neezgai da.",
	"Pòdi manjar de veire, me nafrariá pas.",
	"ମୁଁ କାଚ ଖାଇପାରେ ଏବଂ ତାହା ମୋର କ୍ଷତି କରିନଥାଏ।.",
	"ਮੈਂ ਗਲਾਸ ਖਾ ਸਕਦਾ ਹਾਂ ਅਤੇ ਇਸ ਨਾਲ ਮੈਨੂੰ ਕੋਈ ਤਕਲੀਫ ਨਹੀਂ.",
	"Pchnąć w tę łódź jeża lub ośm skrzyń fig.",
	"Vejam a bruxa da raposa Salta-Pocinhas e o cão feliz que dorme regalado.",
	"À noite, vovô Kowalsky vê o ímã cair no pé do pingüim queixoso e vovó põe açúcar no chá de tâmaras do jabuti feliz.",
	"Fumegând hipnotic sașiul azvârle mreje în bălți.",
	"В чащах юга жил бы цитрус? Да, но фальшивый экземпляр!",
	"काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥",
	"Puotsu mangiari u vitru, nun mi fa mali.",
	"Starý kôň na hŕbe kníh žuje tíško povädnuté ruže, na stĺpe sa ďateľ učí kvákať novú ódu o živote.",
	"Šerif bo za vajo spet kuhal domače žgance.",
	"Unë mund të ha qelq dhe nuk më gjen gjë.",
	"Чешће цeђење мрeжастим џаком побољшава фертилизацију генских хибрида.",
	"Flygande bäckasiner söka strax hwila på mjuka tuvor.",
	"I kå Glas frässa, ond des macht mr nix!",
	"நான் கண்ணாடி சாப்பிடுவேன், அதனால் எனக்கு ஒரு கேடும் வராது.",
	"నేను గాజు తినగలను అయినా నాకు యేమీ కాదు.",
	"Kaya kong kumain nang bubog at hindi ako masaktan.",
	"Pijamalı hasta yağız şoföre çabucak güvendi.",
	"Metumi awe tumpan, ɜnyɜ me hwee.",
	"Чуєш їх, доцю, га? Кумедна ж ти, прощайся без ґольфів!",
	"میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔",
	"Mi posso magnare el vetro, no'l me fa mae.",
	"Con sói nâu nhảy qua con chó lười.",
	"Dji pou magnî do vêre, çoula m' freut nén må.",
	"איך קען עסן גלאָז און עס טוט מיר נישט װײ.",
	"Mo lè je̩ dígí, kò ní pa mí lára.",
	"Saya boleh makan kaca dan ia tidak mencederakan saya.",

	// harfbuzz version issue
	// "නොපු",
	// "මනොපුබ්‌බඞ්‌ගමා ධම්‌මා, මනොසෙට්‌ඨා මනොමයා; මනසා චෙ පදුට්‌ඨෙන, භාසති වා කරොති වා; තතො නං දුක්‌ඛමන්‌වෙති, චක්‌කංව වහතො පදං.",
	// not the same resvoled font
	//	"હું કાચ ખાઇ શકુ છુ અને તેનાથી મને દર્દ નથી થતુ."
	// "ဘာသာပြန်နှင့် စာပေပြုစုရေး ကော်မရှင်"

	// the following do not use unicode word boundaries

	"ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ",
	"いろはにほへと ちりぬるを 色は匂へど 散りぬるを",
	".o'i mu xagji sofybakni cu zvati le purdi",
	"ຂອ້ຍກິນແກ້ວໄດ້ໂດຍທີ່ມັນບໍ່ໄດ້ເຮັດໃຫ້ຂອ້ຍເຈັບ",
	"เป็นมนุษย์สุดประเสริฐเลิศคุณค่า - กว่าบรรดาฝูงสัตว์เดรัจฉาน - จงฝ่าฟันพัฒนาวิชาการ อย่าล้างผลาญฤๅเข่นฆ่าบีฑาใคร - ไม่ถือโทษโกรธแช่งซัดฮึดฮัดด่า - หัดอภัยเหมือนกีฬาอัชฌาสัย - ปฏิบัติประพฤติกฎกำหนดใจ - พูดจาให้จ๊ะ ๆ จ๋า ๆ น่าฟังเอยฯ",
	"Pa's wijze lynx bezag vroom het fikse aquaduct.",
	"Ch'peux mingi du verre, cha m'foé mie n'ma.",
	"我能吞下玻璃而不伤身体。",
	"我能吞下玻璃而不傷身體。",
	"我能吞下玻璃而不伤身体。",
	"我能吞下玻璃而不傷身體。",
}

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fontconfig.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	panic(err)
	// }

	fs, err := fontconfig.LoadFontsetFile(fontmapCache)
	if err != nil {
		panic(err)
	}
	fontmapPango = fcfonts.NewFontMap(fontconfig.Standard, fs)

	fontmapGotext = fontscan.NewFontMap(log.New(io.Discard, "", 0))
	err = fontmapGotext.UseSystemFonts("testdata")
	if err != nil {
		panic(err)
	}
}

func assert(t *testing.T, b bool, msg string) {
	if !b {
		t.Fatal(msg)
	}
}

type textContext struct {
	fc   FontConfiguration
	dict map[HyphenDictKey]hyphen.Hyphener
}

func textContextPango() textContext {
	return textContext{&FontConfigurationPango{fontmap: fontmapPango}, make(map[HyphenDictKey]hyphen.Hyphener)}
}

func textContextGotext() textContext {
	return textContext{NewFontConfigurationGotext(fontmapGotext), make(map[HyphenDictKey]hyphen.Hyphener)}
}

func (tc textContext) Fonts() FontConfiguration                       { return tc.fc }
func (tc textContext) HyphenCache() map[HyphenDictKey]hyphen.Hyphener { return tc.dict }
func (tc textContext) StrutLayoutsCache() map[StrutLayoutKey][2]pr.Float {
	return make(map[StrutLayoutKey][2]pr.Float)
}

// Wrapper for SplitFirstLine() creating a style dict.
func makeText(text string, width pr.MaybeFloat, style pr.Properties) FirstLine {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(style)
	ct := textContextPango()
	return SplitFirstLine([]rune(text), newStyle, ct, width, false, true)
}

func TestLineContent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	for _, v := range []struct {
		remaining string
		width     pr.Float
	}{
		{"text for test", 100},
		{"is a text for test", 45},
	} {
		text := "This is a text for test"
		sp := makeText(text, v.width, pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
		textRunes := []rune(text)
		assert(t, string(textRunes[sp.ResumeAt:]) == v.remaining, "unexpected remaining")
		assert(t, sp.Length+1 == sp.ResumeAt, fmt.Sprintf("%v: expected %d, got %d", v.width, sp.ResumeAt, sp.Length+1)) // +1 for the removed trailing space
	}
}

func TestLineWithAnyWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	sp1 := makeText("some text", nil, nil)
	sp2 := makeText("some text some text", nil, nil)
	assert(t, sp1.Width < sp2.Width, "unexpected width")
}

func TestLineBreaking(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	str := "Thïs is a text for test"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "is a text for test", "")

	sp = makeText(str, pr.Float(100), pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "text for test", "")
}

func TestLineBreakingRTL(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	str := "لوريم ايبسوم دولا"
	// These two tests do not really rely on installed fonts
	sp := makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(1)})
	assert(t, sp.ResumeAt == -1, "")

	sp = makeText(str, pr.Float(90), pr.Properties{pr.PFontSize: pr.FToV(100)})
	assert(t, string([]rune(str)[sp.ResumeAt:]) == "ايبسوم دولا", "")
}

func TestTextDimension(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)

	str := "This is a text for test. This is a test for text.py"
	sp1 := makeText(str, pr.Float(200), pr.Properties{pr.PFontSize: pr.FToV(12)})
	sp2 := makeText(str, pr.Float(200), pr.Properties{pr.PFontSize: pr.FToV(20)})
	assert(t, sp1.Width*sp1.Height < sp2.Width*sp2.Height, "")
}

func TestGetLastWordEnd(t *testing.T) {
	pango := &FontConfigurationPango{}
	gotext := &FontConfigurationGotext{}
	for _, test := range []struct {
		text     string
		expected int
	}{
		{"aa a", 2},
		{"aaa cc", 3},
		{"a", -1},
	} {
		tu.AssertEqual(t, GetLastWordEnd(pango, []rune(test.text)), test.expected)
		tu.AssertEqual(t, gotext.GetLastWordEnd([]rune(test.text)), test.expected)
	}
}

func TestHeightAndBaseline(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	families := pr.Strings{
		"Helvetica",
		"Apple Color Emoji",
	}
	newStyle.SetFontFamily(families)

	newStyle.SetFontSize(pr.FToV(36))
	// ct := textContextPango()
	ct := textContextGotext()

	// fc := NewFontConfigurationPango(fontmapPango)
	for _, desc := range []validation.FontFaceDescriptors{
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/googlesans/v36/4UaGrENHsxJlGDuGo1OIlL3Owps.ttf"}}, FontFamily: "Google Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/googlesans/v36/4UabrENHsxJlGDuGo1OIlLU94YtzCwM.ttf"}}, FontFamily: "Google Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/materialicons/v117/flUhRq6tzZclQEJ-Vdg-IuiaDsNZ.ttf"}}, FontFamily: "Material Icons", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/opensans/v27/memSYaGs126MiZpBA-UvWbX2vVnXBbObj2OVZyOOSr4dVJWUgsjZ0B4gaVc.ttf"}}, FontFamily: "Open Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}, FontStretch: "normal"},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/roboto/v29/KFOmCnqEu92Fr1Mu4mxP.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/roboto/v29/KFOlCnqEu92Fr1MmEU9fBBc9.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/roboto/v29/KFOlCnqEu92Fr1MmWUlfBBc9.ttf"}}, FontFamily: "Roboto", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 700}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K0nXBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 400}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K3vXBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 500}},
		{Src: []pr.TaggedString{{Tag: pr.External, S: "https://fonts.gstatic.com/s/worksans/v13/QGY_z_wNahGAdqQ43RhVcIgYT2Xz5u32K5fQBi8Jow.ttf"}}, FontFamily: "Work Sans", FontStyle: "normal", FontWeight: pr.IntString{String: "", Int: 600}},
	} {
		ct.fc.AddFontFace(desc, utils.DefaultUrlFetcher)
	}

	spi := SplitFirstLine([]rune("Go 1.17 Release Notes"), newStyle, ct, pr.Float(595), false, true)
	height, baseline := spi.Height, spi.Baseline

	if int((height-43)/10) != 0 {
		t.Fatalf("unexpected height %f", height)
	}
	if int((baseline-33)/10) != 0 {
		t.Fatalf("unexpected baseline %f", baseline)
	}
}

func addWeayprintFont(t *testing.T, fc FontConfiguration) {
	url, err := utils.PathToURL("../resources_test/weasyprint.otf")
	tu.AssertNoErr(t, err)
	fc.AddFontFace(validation.FontFaceDescriptors{
		Src:        []pr.TaggedString{{Tag: pr.External, S: url}},
		FontFamily: "weasyprint",
	}, utils.DefaultUrlFetcher)
}

func newContextWithWeasyFont(t *testing.T) textContext {
	ct := textContextPango()
	fc := NewFontConfigurationPango(fontmapPango)
	addWeayprintFont(t, fc)
	return ct
}

func TestLayoutFirstLine(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(pr.Strings{"weasyprint"})
	newStyle.SetFontSize(pr.FToV(16))
	newStyle.SetWhiteSpace(pr.Normal)
	ts := NewTextStyle(newStyle, false)

	ct := newContextWithWeasyFont(t)

	layout := createLayout("a a ", ts, ct.Fonts(), pr.Float(63))
	_, index := layout.GetFirstLine()
	if index != -1 {
		t.Fatalf("unexpected first line index: %d", index)
	}
}

// func TestChWidth(t *testing.T) {
// 	newStyle := pr.InitialValues.Copy()
// 	newStyle.SetFontFamily(pr.Strings{"arial"})
// 	newStyle.SetFontSize(pr.FToV(16))
// 	//  pr.FToV(-0.04444)
// 	ct := textContext{fontmap: fontmap, dict: make(map[HyphenDictKey]hyphen.Hyphener)}
// 	if w := CharacterRatio(dummyStyle{newStyle}, pr.NewTextRatioCache(), true, ct); utils.RoundPrec(pr.Fl(w), 3) != 8.854 {
// 		t.Fatalf("unexpected ch width %v", w)
// 	}
// }

func TestSplitFirstLine(t *testing.T) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(pr.Strings{"arial"})
	newStyle.SetFontSize(pr.FToV(16))

	ct := textContextPango()

	out := SplitFirstLine([]rune(" of the element's "), newStyle, ct, pr.Float(120.18628), false, true)

	if out.ResumeAt != -1 {
		t.Fatalf("unexpected resume index %d", out.ResumeAt)
	}
}

func TestCanBreakText(t *testing.T) {
	tests := []struct {
		s    string
		want pr.MaybeBool
	}{
		{" s", pr.True},
		{"\u00a0L", pr.False},
		{"\u00a0d", pr.False},
		{"r\u00a0", pr.False},
		{" “", pr.True},
		{"” ", pr.False},
		{"t\u00a0", pr.False},
		{"\u00a0L", pr.False},
		{"\u00a0d", pr.False},
		{"r\u00a0", pr.False},
		{" “", pr.True},
		{"” ", pr.False},
		{"t\u00a0", pr.False},
		{"a⺀", pr.True},
		{"⺀b", pr.True},
		{"bc", pr.False},
		{"a⺀", pr.True},
		{"⺀b", pr.True},
		{"bc", pr.False},
		{"", nil},
		{"c ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"a ", pr.False},
		{"a", nil},
		{"a ", pr.False},
		{"a", nil},
		{"⺀ ", pr.False},
		{"a", nil},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"a", nil},
		{"a", nil},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"c\u00a0", pr.False},
		{"i", nil},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"ii", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"\u00a0\u00a0", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0a", pr.False},
		{" a", pr.True},
		{"\u00a0 ", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"b\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"c\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"c\u00a0", pr.False},
		{"\u200f\u00a0i", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u200f\u00a0i", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"\u00a0\u200f", pr.False},
		{"\u200f\u00a0ii", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u200f\u00a0ii", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"\u00a0\u200f", pr.False},
		{"a\u00a0", pr.False},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"bb", pr.False},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"c", nil},
		{"a", nil},
		{"a", nil},
		{"\u00a0a", pr.False},
		{"a", nil},
		{"abc", pr.False},
		{"abcde", pr.False},
		{"abcde", pr.False},
		{"[initial]", pr.False},
		{"[]", pr.False},
		{"o", nil},
		{"abcde", pr.False},
		{"ab", pr.False},
		{"cd", pr.False},
		{"bc", pr.False},
		{"b", nil},
		{"a", nil},
		{"e", nil},
		{"de", pr.False},
		{"a", nil},
		{"b", nil},
		{"cd", pr.False},
		{"abcde", pr.False},
		{"ace", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{" ⺀", pr.True},
		{"⺀ ", pr.False},
		{" 4", pr.True},
		{"4 ", pr.False},
		{"  ", pr.False},
		{" h", pr.True},
		{" i", pr.True},
		{"z ", pr.False},
		{" a", pr.True},
		{"a ", pr.False},
		{"⺀ ", pr.False},
		{"⺀ ", pr.False},
		{"t ", pr.False},
		{" A", pr.True},
		{"t ", pr.False},
		{"test", pr.False},
	}
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	fcGotext := &FontConfigurationGotext{}
	for _, tt := range tests {
		if got := fcPango.CanBreakText([]rune(tt.s)); got != tt.want {
			t.Errorf("pango.CanBreakText(%s) = %v, want %v", tt.s, got, tt.want)
		}
		if got := fcGotext.CanBreakText([]rune(tt.s)); got != tt.want {
			t.Errorf("gotext.CanBreakText(%s) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func wrapPango(fc *FontConfigurationPango, text string, style *TextStyle, maxWidth pr.MaybeFloat) FirstLine {
	layout := createLayout(text, style, fc, maxWidth)
	firstLine, resumeIndex := layout.GetFirstLine()
	// fmt.Println("Pango ->")
	// for run := firstLine.Runs; run != nil; run = run.Next {
	// 	fmt.Println(run.Data.Item.Analysis.Font.FaceID())
	// 	for _, glyph := range run.Data.Glyphs.Glyphs {
	// 		fmt.Println(glyph.Glyph.GID(), glyph.Geometry.Width)
	// 	}
	// }
	return firstLineMetrics(firstLine, []rune(text), layout, resumeIndex, style.spaceCollapse(), style, false, "")
}

func assertApprox(t *testing.T, got, exp pr.Float, context string) {
	t.Helper()
	if !(pr.Abs(got-exp) < exp/200) { // 0.5% of tolerance
		t.Fatalf("%s: expected %v, got %v", context, exp, got)
	}
}

func TestWrap(t *testing.T) {
	fcG := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}

	for _, family := range []string{"Nimbus Sans", "Nimbus Roman", "DejaVu Sans", "Liberation Mono", "Arimo"} {
		for _, w := range []uint16{400} { // weights
			for _, s := range []pr.Fl{12, 13, 16, 18, 32, 33} { // sizes
				style := &TextStyle{FontDescription: FontDescription{
					Family:  []string{family},
					Weight:  w,
					Size:    s,
					Stretch: FStr_Normal,
				}}

				for _, text := range textSamples {
					if text == "ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ" || text == "ຂອ້ຍກິນແກ້ວໄດ້ໂດຍທີ່ມັນບໍ່ໄດ້ເຮັດໃຫ້ຂອ້ຍເຈັບ" {
						// pango does not select the same fonts
						continue
					}

					// no max width
					refPango := wrapPango(fcPango, text, style, nil)
					lineGotext := fcG.wrap([]rune(text), style, pr.Inf)

					tu.AssertEqual(t, lineGotext.Length, len([]rune(text)))
					tu.AssertEqual(t, lineGotext.ResumeAt, -1)

					assertApprox(t, lineGotext.Width, refPango.Width, "Width")
					assertApprox(t, lineGotext.Height, refPango.Height, "Height")
					assertApprox(t, lineGotext.Baseline, refPango.Baseline, "Baseline")

					for _, maxWidth := range []pr.Float{10, 50, 101, 201, 1000} {
						refPango := wrapPango(fcPango, text, style, maxWidth)
						lineGotext := fcG.wrap([]rune(text), style, maxWidth)

						tu.AssertEqual(t, lineGotext.Length, refPango.Length)
						tu.AssertEqual(t, lineGotext.ResumeAt, refPango.ResumeAt)

						assertApprox(t, lineGotext.Width, refPango.Width, fmt.Sprintf("Width for %v", maxWidth))
						assertApprox(t, lineGotext.Height, refPango.Height, fmt.Sprintf("Height for %v", maxWidth))
						assertApprox(t, lineGotext.Baseline, refPango.Baseline, fmt.Sprintf("Baseline for %v", maxWidth))
					}
				}
			}
		}
	}
}

func BenchmarkWrap(b *testing.B) {
	fcG := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	const text = "Une superbe phrase en français ! And also some english and שלום أهلا שלום أه"
	text_ := []rune(text)
	b.ResetTimer()

	b.Run("pango", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, family := range []string{"Nimbus Sans", "Nimbus Roman", "DejaVu Sans", "Liberation Mono", "Arimo"} {
				for _, w := range []uint16{400, 700} { // weights
					for _, s := range []pr.Fl{12, 13, 16, 18, 32, 33} { // sizes
						style := &TextStyle{FontDescription: FontDescription{
							Family:  []string{family},
							Weight:  w,
							Stretch: FStr_Normal,
							Size:    s * 100,
						}}
						_ = wrapPango(fcPango, text, style, nil)
					}
				}
			}
		}
	})

	b.Run("Gotext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, family := range []string{"Nimbus Sans", "Nimbus Roman", "DejaVu Sans", "Liberation Mono", "Arimo"} {
				for _, w := range []uint16{400, 700} { // weights
					for _, s := range []pr.Fl{12, 13, 16, 18, 32, 33} { // sizes
						style := &TextStyle{FontDescription: FontDescription{
							Family:  []string{family},
							Weight:  w,
							Stretch: FStr_Normal,
							Size:    s * 100,
						}}
						_ = fcG.wrap(text_, style, pr.Inf)
					}
				}
			}
		}
	})
}

func TestSplit(t *testing.T) {
	gotext := textContextGotext()
	pango := textContextPango()
	style := pr.InitialValues.Copy()
	style.SetLang(pr.TaggedString{S: "fr"})
	style.SetHyphens("auto")
	style.SetWordBreak("break-word")
	style.SetOverflowWrap("break-word")
	style.SetFontFamily(pr.Strings{"NotoSans"})

	for maxWidth := pr.Float(60); maxWidth < 100; maxWidth += 10 {
		lineP := SplitFirstLine([]rune("Une jolie phrase - hahaha"), style, pango, maxWidth, false, true)
		lineG := SplitFirstLine([]rune("Une jolie phrase - hahaha"), style, gotext, maxWidth, false, true)
		tu.AssertEqual(t, lineG.ResumeAt, lineP.ResumeAt)
		tu.AssertEqual(t, lineG.FirstLineRTL, lineP.FirstLineRTL)
		tu.AssertEqual(t, lineG.Length, lineP.Length)
	}
}

func BenchmarkSplitFirstLine(b *testing.B) {
	newStyle := pr.InitialValues.Copy()
	newStyle.SetFontFamily(monoFonts)
	newStyle.UpdateWith(pr.Properties{pr.PFontFamily: sansFonts, pr.PFontSize: pr.FToV(19)})
	cPango := textContextPango()
	cGotext := textContextGotext()

	text := []rune("Une superbe phrase en français ! And also some english and שלום أهلا שלום أه")

	b.Run("pango", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for maxWidth := pr.Float(60); maxWidth < 100; maxWidth += 10 {
				_ = SplitFirstLine(text, newStyle, cPango, maxWidth, false, true)
			}
		}
	})

	b.Run("go-text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for maxWidth := pr.Float(60); maxWidth < 100; maxWidth += 10 {
				_ = SplitFirstLine(text, newStyle, cGotext, maxWidth, false, true)
			}
		}
	})
}

func TestLetterAndWordSpacing(t *testing.T) {
	t.Skip("TODO: investigate")

	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	style := &TextStyle{
		FontDescription: FontDescription{
			Family:  []string{"Nimbus Sans"},
			Weight:  400,
			Stretch: FStr_Normal,
			Size:    12,
		},
	}

	for _, text := range textSamples {
		for _, ls := range [...]pr.Fl{0, 1, 2, 10} {
			for _, ws := range [...]pr.Fl{0, 1, 2, 10} {
				style.LetterSpacing = ls
				style.WordSpacing = ws
				lineP := wrapPango(fcPango, text, style, nil)
				lineG := fcGotext.wrap([]rune(text), style, pr.Inf)
				assertApprox(t, lineP.Width, lineG.Width, fmt.Sprintf("text: %s, letter-spacing: %v, word-spacing: %v", text, ls, ws))
			}
		}
	}
}

func TestWordBoundaries(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}

	for _, text := range textSamples[:len(textSamples)-11] {
		tu.AssertEqual(t, fcGotext.wordBoundaries([]rune(text)), fcPango.wordBoundaries([]rune(text)))
	}
}

func TestResolveFace(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"Nimbus Sans"},
		Style:   FSty_Normal,
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    12,
	}}
	t.Run("style normal", func(t *testing.T) {
		for _, text := range textSamples {
			lineP := resolveFacePango(fcPango, text, style)
			lineG := resolveFaceGotext(fcGotext, text, style)
			tu.AssertEqualG(t, lineG, lineP)
		}
	})

	t.Run("style bold", func(t *testing.T) {
		style.Weight = 700 // Bold
		for _, text := range textSamples {
			if text == "ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ" || text == "いろはにほへと ちりぬるを 色は匂へど 散りぬるを" { // ignore this test since Pango does not produce sounds results
				continue
			}
			lineP := resolveFacePango(fcPango, text, style)
			lineG := resolveFaceGotext(fcGotext, text, style)
			tu.AssertEqual(t, lineG, lineP)
		}
	})
}

type faceRun struct {
	offset, length int
	face           string // resolved family
}

func resolveFacePango(fc *FontConfigurationPango, text string, style *TextStyle) (out []faceRun) {
	fixExp := func(s string) string {
		switch style.Weight {
		case 700:
			// pango does not find some bold fonts
			switch [2]string{text, s} {
			case [...]string{"আমি কাঁচ খেতে পারি, তাতে আমার কোনো ক্ষতি হয় না।", "Lohit Bengali"}:
				return "Mukti"
			case [...]string{"नहीं नजर किसी की बुरी नहीं किसी का मुँह काला जो करे सो उपर वाला", "Lohit Devanagari"}:
				return "Annapurna SIL"
			case [...]string{"मी काच खाऊ शकतो, मला ते दुखत नाही.", "Lohit Devanagari"}:
				return "Annapurna SIL"
			case [...]string{"काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥", "Lohit Devanagari"}:
				return "Annapurna SIL"
			case [...]string{"ನಾನು ಗಾಜನ್ನು ತಿನ್ನಬಲ್ಲೆ ಮತ್ತು ಅದರಿಂದ ನನಗೆ ನೋವಾಗುವುದಿಲ್ಲ.", "Lohit Kannada"}:
				return "FreeSerif"
			case [...]string{"ਮੈਂ ਗਲਾਸ ਖਾ ਸਕਦਾ ਹਾਂ ਅਤੇ ਇਸ ਨਾਲ ਮੈਨੂੰ ਕੋਈ ਤਕਲੀਫ ਨਹੀਂ.", "Lohit Gurmukhi"}:
				return "FreeSerif"
			case [...]string{"நான் கண்ணாடி சாப்பிடுவேன், அதனால் எனக்கு ஒரு கேடும் வராது.", "Lohit Tamil"}:
				return "FreeSerif"
			case [...]string{"വേദനയില്ലാതെ കുപ്പിചില്ലു് എനിയ്ക്കു് കഴിയ്ക്കാം.", "Meera"}:
				return "Manjari"
			case [...]string{"ମୁଁ କାଚ ଖାଇପାରେ ଏବଂ ତାହା ମୋର କ୍ଷତି କରିନଥାଏ।.", "Lohit Odia"}:
				return "Noto Sans Oriya"
			case [...]string{"నేను గాజు తినగలను అయినా నాకు యేమీ కాదు.", "Lohit Telugu"}:
				return "Noto Sans Telugu"
			default:
				return s
			}
		default:
			switch [2]string{text, s} {
			case [...]string{"ဘာသာပြန်နှင့် စာပေပြုစုရေး ကော်မရှင်", "Padauk"}:
				return "Noto Sans Myanmar"
			case [...]string{"હું કાચ ખાઇ શકુ છુ અને તેનાથી મને દર્દ નથી થતુ.", "padmaa"}: // font weight issue
				return "Lohit Gujarati"
			default:
				return s
			}
		}
	}

	lineP := wrapPango(fc, text, style, nil)
	line, _ := lineP.Layout.(*TextLayoutPango).GetFirstLine()
	for run := line.Runs; run != nil; run = run.Next {
		out = append(out, faceRun{
			run.Data.Item.Offset, run.Data.Item.Length,
			fixExp(run.Data.Item.Analysis.Font.Describe(true).FamilyName),
		})
	}
	return out
}

func resolveFaceGotext(fc *FontConfigurationGotext, text string, style *TextStyle) (out []faceRun) {
	lineG := fc.wrap([]rune(text), style, pr.Inf)
	line := lineG.Layout.(TextLayoutGotext).Line
	for _, run := range line {
		out = append(out, faceRun{
			run.Runes.Offset, run.Runes.Count,
			run.Face.Font.Describe().Family,
		})
	}
	return out
}

func TestBaseline(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := NewFontConfigurationPango(fontmapPango)

	addWeayprintFont(t, fcGotext)
	addWeayprintFont(t, fcPango)

	style := &TextStyle{FontDescription: FontDescription{Family: []string{"weasyprint"}, Size: 1}}
	text := "abc def ghi jkl "
	lineP := wrapPango(fcPango, text, style, nil)
	lineG := fcGotext.wrap([]rune(text), style, pr.Inf)
	tu.AssertEqual(t, lineP.Baseline, lineG.Baseline)
}

func TestSpaceHeight(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := NewFontConfigurationPango(fontmapPango)

	addWeayprintFont(t, fcGotext)
	addWeayprintFont(t, fcPango)

	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"weasyprint"},
		Style:   FSty_Normal,
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    1,
	}}

	_, exp := fcPango.spaceHeight(style)
	_, got := fcGotext.spaceHeight(style)
	fmt.Println(exp, got)
}

func TestForcedLineBreak(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := NewFontConfigurationPango(fontmapPango)
	addWeayprintFont(t, fcGotext)
	addWeayprintFont(t, fcPango)

	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"weasyprint"},
		Style:   FSty_Normal,
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    10,
	}}
	lineGotext1 := fcGotext.wrap([]rune("test\n256"), style, pr.Inf)
	linePango1 := wrapPango(fcPango, "test\n256", style, nil)
	tu.AssertEqual(t, lineGotext1.Length, linePango1.Length)
	tu.AssertEqual(t, lineGotext1.ResumeAt, linePango1.ResumeAt)
	tu.AssertEqual(t, lineGotext1.Width, linePango1.Width)

	lineGotext2 := fcGotext.wrap([]rune("\n"), style, pr.Inf)
	linePango2 := wrapPango(fcPango, "\n", style, nil)
	tu.AssertEqual(t, lineGotext2.Length, linePango2.Length)
	tu.AssertEqual(t, lineGotext2.ResumeAt, linePango2.ResumeAt)
	tu.AssertEqual(t, lineGotext2.Width, linePango2.Width)
}

func TestSplitRTL(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := NewFontConfigurationPango(fontmapPango)

	style := &TextStyle{FontDescription: FontDescription{
		Style:   FSty_Normal,
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    10,
	}, Direction: pr.Rtl}

	gotext := fcGotext.wrap([]rune("abc "), style, pr.Inf)
	pango := wrapPango(fcPango, "abc ", style, nil)

	tu.AssertEqual(t, pango.FirstLineRTL, true)
	runsPango := pango.Layout.(*TextLayoutPango).Layout.GetLine(0).Runs
	runP0, runP1 := runsPango.Data, runsPango.Next.Data
	tu.Assert(t, runP0.Item.Length == 1 && runP1.Item.Length == 3)
	tu.Assert(t, runP0.Item.Analysis.Level%2 == 1 && runP1.Item.Analysis.Level%2 == 0) // RTL, LTR

	tu.AssertEqual(t, gotext.FirstLineRTL, true)
	runsGotext := gotext.Layout.(TextLayoutGotext).Line
	runG0, runG1 := runsGotext[0], runsGotext[1]
	tu.AssertEqual(t, len(runsGotext), 2)
	tu.Assert(t, runG0.Runes.Count == 1 && runG1.Runes.Count == 3)
	tu.Assert(t, runG0.Direction == di.DirectionRTL && runG1.Direction == di.DirectionLTR)
}

func TestSegmentRTL(t *testing.T) {
	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	addWeayprintFont(t, fcGotext)

	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"weasyprint"},
		Style:   FSty_Normal,
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    10,
	}}

	gotext := fcGotext.wrap([]rune("\u200fabc"), style, pr.Inf)
	runs := gotext.Layout.(TextLayoutGotext).Line
	tu.Assert(t, len(runs) == 2 && runs[0].Face == runs[1].Face) // dont change for "\u200f"
	tu.AssertEqualG(t, gotext.Height, 10)
}

func TestDebug(t *testing.T) {
	t.Skip("dev test only")

	fcGotext := NewFontConfigurationGotext(fontmapGotext)
	fcPango := &FontConfigurationPango{fontmap: fontmapPango}
	style := &TextStyle{FontDescription: FontDescription{
		Family:  []string{"Liberation Mono"},
		Weight:  400,
		Stretch: FStr_Normal,
		Size:    12,
	}}
	const text = "ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ"

	lineP := wrapPango(fcPango, text, style, nil)
	lineG := fcGotext.wrap([]rune(text), style, pr.Inf)
	fmt.Printf("%s :\n%v\n%v\n\n", text, lineP.Width, lineG.Width)

	// style.LetterSpacing = 10
	// style.WordSpacing = 2
	// lineP = wrapPango(fcPango, text, style, nil)
	// lineG = fcGotext.wrap([]rune(text), style, pr.Inf)
	// fmt.Printf("%s :\n%v\n%v\n\n", text, lineP.Width, lineG.Width)

	// style.LetterSpacing = 10
	// style.WordSpacing = 4
	// lineP = wrapPango(fcPango, text, style, nil)
	// lineG = fcGotext.wrap([]rune(text), style, pr.Inf)
	// fmt.Printf("%s :\n%v\n%v\n\n", text, lineP.Width, lineG.Width)
}
