package document

import (
	"fmt"
	"io"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils/tracer"
	"github.com/go-text/typesetting/fontscan"
)

const fontmapCache = "../../text/testdata/cache.fc"

var fc *text.FontConfigurationPango

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fc.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	panic(err)
	// }

	fs, err := fontconfig.LoadFontsetFile(fontmapCache)
	if err != nil {
		panic(err)
	}
	fc = text.NewFontConfigurationPango(fcfonts.NewFontMap(fontconfig.Standard.Copy(), fs))
}

func TestStacking(t *testing.T) {
	var s StackingContext
	if s.Type().IsClassical() {
		t.Fatal("should not be a classical box")
	}
}

func TestSVG(t *testing.T) {
	tmp := headerSVG + crop + cross
	tp := template.Must(template.New("svg").Parse(tmp))
	if err := tp.Execute(io.Discard, svgArgs{}); err != nil {
		t.Fatal(err)
	}
}

func TestWriteSimpleDocument(t *testing.T) {
	htmlContent := `      
	<style>
		@page { @top-left  { content: "[" string(left) "]" } }
		p { page-break-before: always }
		.initial { string-set: left "initial" }
		.empty   { string-set: left ""        }
		.space   { string-set: left " "       }
	</style>

	<p class="initial">Initial</p>
	<p class="empty">Empty</p>
	<p class="space">Space</p>
	`

	doc, err := tree.NewHTML(utils.InputString(htmlContent), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerNoOp(), 1, nil)
}

func TestWriteDocument(t *testing.T) {
	doc, err := tree.NewHTML(utils.InputFilename("../../resources_test/acid2-test.html"), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerNoOp(), 1, nil)
}

func TestCrash(t *testing.T) {
	doc, err := tree.NewHTML(utils.InputFilename("../../resources_test/preserveAspectRatio.html"), "https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/preserveAspectRatio", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerNoOp(), 1, nil)
}

func renderUrl(t testing.TB, url string) {
	doc, err := tree.NewHTML(utils.InputUrl(url), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerNoOp(), 1, nil)
}

func TestRealPage(t *testing.T) {
	t.Skip()

	logger.WarningLogger.SetOutput(io.Discard)
	defer logger.WarningLogger.SetOutput(os.Stdout)

	renderUrl(t, "http://www.google.com")
	renderUrl(t, "https://weasyprint.org/")
	renderUrl(t, "https://en.wikipedia.org/wiki/Go_(programming_language)") // rather big document
	renderUrl(t, "https://golang.org/doc/go1.17")                           // slow because of text layout
	renderUrl(t, "https://github.com/Kozea/WeasyPrint")
	// renderUrl(t, "https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/preserveAspectRatio") // large page, very slow !
}

func BenchmarkRender(b *testing.B) {
	logger.ProgressLogger.SetOutput(io.Discard)
	logger.WarningLogger.SetOutput(io.Discard)
	defer func() {
		logger.WarningLogger.SetOutput(os.Stdout)
		logger.ProgressLogger.SetOutput(os.Stdout)
	}()

	doc, err := tree.NewHTML(utils.InputFilename("../../resources_test/acid2-test.html"), "", nil, "")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		out := Render(doc, nil, true, fc)
		out.Write(tracer.NewDrawerNoOp(), 1, nil)
	}
}

func BenchmarkRenderText(b *testing.B) {
	logger.ProgressLogger.SetOutput(io.Discard)
	logger.WarningLogger.SetOutput(io.Discard)
	defer func() {
		logger.WarningLogger.SetOutput(os.Stdout)
		logger.ProgressLogger.SetOutput(os.Stdout)
	}()

	doc, err := tree.NewHTML(utils.InputFilename("testdata/go1.17.html"), "", nil, "")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		out := Render(doc, nil, true, fc)
		out.Write(tracer.NewDrawerNoOp(), 1, nil)
	}
}

func TestLeaderCrash(t *testing.T) {
	input := `
	<style>
		@font-face {src: url(../../resources_test/weasyprint.otf); font-family: weasyprint}
		@page {
		background: white;
		size: 16px 6px;
		}
		body {
		color: red;
		counter-reset: count;
		direction: rtl;
		font-family: weasyprint;
		font-size: 2px;
		line-height: 1;
		}
		div::after {
		color: blue;
		/* RTL Mark used in second space */
		content: ' ' leader(dotted) '\u200f ' counter(count, lower-roman);
		counter-increment: count;
		}
  	</style>
	<div>a</div>
	<div>bb</div>
	<div>c</div>`
	doc, err := tree.NewHTML(utils.InputString(input), ".", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc.UAStyleSheet = tree.TestUAStylesheet
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerNoOp(), 4./30, nil)
}

func TestDebug(t *testing.T) {
	input := `
	 <style>
        @page {
          size: 20px 6px;
          margin: 1px;
        }
        @font-face {
          src: url(../resources_test/weasyprint.otb_fixed);
          font-family: weasyprint-otb;
        }
        body {
          color: red;
          font-family: weasyprint-otb;
          font-size: 4px;
          line-height: 0.8;
        }
      </style>
      AaA`

	doc, err := tree.NewHTML(utils.InputString(input), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc.UAStyleSheet = tree.TestUAStylesheet
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.Write(tracer.NewDrawerFile("/tmp/drawer_go.txt"), 4./30, nil)
}

func BenchmarkRenderAttestation(b *testing.B) {
	logger.ProgressLogger.SetOutput(io.Discard)
	logger.WarningLogger.SetOutput(io.Discard)
	defer func() {
		logger.WarningLogger.SetOutput(os.Stdout)
		logger.ProgressLogger.SetOutput(os.Stdout)
	}()

	doc, err := tree.NewHTML(utils.InputFilename("../../resources_test/modele.html"), "", nil, "")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		out := Render(doc, nil, true, fc)
		out.Write(tracer.NewDrawerNoOp(), 1, nil)
	}
}

func TestLayoutTime(t *testing.T) {
	t.Skip() // TODO
	logger.ProgressLogger.SetOutput(io.Discard)
	logger.WarningLogger.SetOutput(io.Discard)
	defer func() {
		logger.WarningLogger.SetOutput(os.Stdout)
		logger.ProgressLogger.SetOutput(os.Stdout)
	}()

	fm := fontscan.NewFontMap(nil)
	fm.UseSystemFonts(os.TempDir())
	fcGotext := text.NewFontConfigurationGotext(fm)

	doc, err := tree.NewHTML(utils.InputFilename("testdata/fiche_sanitaire.html"), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now()
	_ = Render(doc, nil, true, fc)
	fmt.Println(time.Since(ti))

	ti = time.Now()
	_ = Render(doc, nil, true, fcGotext)
	fmt.Println(time.Since(ti))
}

func Benchmark(b *testing.B) {
	logger.ProgressLogger.SetOutput(io.Discard)
	logger.WarningLogger.SetOutput(io.Discard)
	defer func() {
		logger.WarningLogger.SetOutput(os.Stdout)
		logger.ProgressLogger.SetOutput(os.Stdout)
	}()

	fm := fontscan.NewFontMap(nil)
	fm.UseSystemFonts(os.TempDir())
	fcGotext := text.NewFontConfigurationGotext(fm)

	doc, err := tree.NewHTML(utils.InputFilename("testdata/fiche_sanitaire_1.html"), baseUrl, nil, "")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Render(doc, nil, true, fcGotext)
	}
}
