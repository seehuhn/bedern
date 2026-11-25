// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"seehuhn.de/go/bedern/internal/font/italic/assembly"
	"seehuhn.de/go/bedern/internal/tex"
	"seehuhn.de/go/bedern/internal/tfm"
)

// italicFont returns the italic as it was rendered, and skips the test in the
// routine tier, which leaves out the italic (TestMain).
func italicFont(t *testing.T) *rendered {
	t.Helper()
	if testing.Short() {
		t.Skip("the full tier builds the italic")
	}
	r := ran(t)
	for i, f := range fonts {
		if f.italic {
			return r.fonts[i]
		}
	}
	t.Fatal("the italic is not among the fonts")
	return nil
}

// italicTFM is the path of the metric file of the italic in the encoding e.
func italicTFM(e *tex.Encoding) string {
	return tex.DirTFM + "/" + (&tex.Font{Style: assembly.Style()}).TFM(e) + ".tfm"
}

// TestItalicMetrics reads the metric files of the italic in both encodings
// back: they are free of findings, the characters exist for the positions
// which have a glyph, the width of a character is the advance of its glyph,
// the slant is that of the outlines, the italic corrections are those of the
// outlines up to the reduction of the table, and the face byte is that of an
// italic.
func TestItalicMetrics(t *testing.T) {
	t.Parallel()
	it := italicFont(t)
	tree := ran(t).tree
	fix := func(v float64) tfm.Fix { return tfm.FixFromFloat(v / 1000) }
	for _, e := range tex.Encodings {
		name := italicTFM(e)
		file, err := tfm.Read(bytes.NewReader(tree[name]))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, f := range file.Check() {
			t.Errorf("%s: %s", name, f.Text)
		}
		index, _ := e.Resolve(it.tex)
		for c, i := range index {
			if file.Exists(c) != (i >= 0) {
				t.Errorf("%s: the character %#02x exists: %t, its glyph: %t", name, c, file.Exists(c), i >= 0)
			}
			if i < 0 {
				continue
			}
			g := it.tex[i]
			if got, want := file.CharWidth(c), fix(g.Advance); got != want {
				t.Errorf("%s: width of %s %v, want %v", name, g.Name, got, want)
			}
			want := assembly.TeXItalicCorrection(g)
			if got := file.CharItalic(c).Float() * 1000; math.Abs(got-want) > 8 || (want == 0) != (got == 0) {
				t.Errorf("%s: italic correction of %s %.1f, outlines %.0f", name, g.Name, got, want)
			}
		}
		if got, want := file.Param[0].Float(), assembly.Default().Slant(); math.Abs(got-want) > 1e-4 {
			t.Errorf("%s: slant %v, want %v", name, got, want)
		}
		if got := file.Param[1].Float() * 1000; math.Abs(got-assembly.WordSpace) > 1 {
			t.Errorf("%s: space %v", name, got)
		}
		if face, _ := file.Face(); face&1 == 0 {
			t.Errorf("%s: the face byte %d is not italic", name, face)
		}
	}
	file, err := tfm.Read(bytes.NewReader(tree[italicTFM(tex.T1)]))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range "fTVY" {
		if file.CharItalic(int(c)) == 0 {
			t.Errorf("%c has no italic correction", c)
		}
	}
}

// TestItalicFilesForTeX checks the files of the italic in the tree: the map
// file has one line for each metric file and does not slant the outlines, the
// italic is the shape it of the series m in both font definition files and no
// slanted instance of it exists, and the outline file exists.
func TestItalicFilesForTeX(t *testing.T) {
	t.Parallel()
	italicFont(t)
	tree := ran(t).tree
	mapFile := string(tree[tex.DirMap+"/"+tex.MapFile])
	for _, e := range tex.Encodings {
		base := strings.TrimSuffix(filepath.Base(italicTFM(e)), ".tfm")
		line := base + " " + assembly.Style().PostScript + " \"" + e.VectorName() + " ReEncodeFont\" <" +
			e.EncFile() + " <" + assembly.Style().PostScript + ".pfb\n"
		if !strings.Contains(mapFile, "\n"+line) {
			t.Errorf("the map file lacks the line %s", line)
		}
		if strings.Contains(mapFile, base+"-sl") || strings.Contains(mapFile, "italic-sl") {
			t.Errorf("the map file has a slanted instance of the italic")
		}
		fd := string(tree[tex.DirLaTeX+"/"+tex.FDFile(e)])
		want := fmt.Sprintf("\\DeclareFontShape{%s}{bedern}{m}{it}{<-> %s}{}\n", e.Name, base)
		if !strings.Contains(fd, want) {
			t.Errorf("%s lacks %s", tex.FDFile(e), want)
		}
		if strings.Contains(fd, "bedern/m/sl") {
			t.Errorf("%s takes a shape from the slanted regular", tex.FDFile(e))
		}
	}
	if len(tree[tex.DirPFB+"/"+assembly.Style().PostScript+".pfb"]) == 0 {
		t.Error("no outline file of the italic")
	}
	if !bytes.Contains(tree[tex.DirPlain+"/"+tex.PlainFile], []byte("\\def\\it{\\fam\\itfam\\bedernit}\n")) {
		t.Error("the macro file does not set \\it in the italic")
	}
}

// TestItalicKerns checks that the kern of every pair of the lowercase
// letters in the metric file of T1 is the kern of the kerning of the italic,
// and that the ligatures with f give the glyphs of the italic.
func TestItalicKerns(t *testing.T) {
	t.Parallel()
	it := italicFont(t)
	file, err := tfm.Read(bytes.NewReader(ran(t).tree[italicTFM(tex.T1)]))
	if err != nil {
		t.Fatal(err)
	}
	index, _ := tex.T1.Resolve(it.tex)
	for _, l := range []struct{ a, b, res string }{
		{"f", "f", "ff"}, {"f", "i", "fi"}, {"f", "l", "fl"}, {"ff", "i", "ffi"}, {"ff", "l", "ffl"},
	} {
		slot := func(name string) int {
			for c, i := range index {
				if i >= 0 && it.tex[i].Name == name {
					return c
				}
			}
			return -1
		}
		ca, cb, cr := slot(l.a), slot(l.b), slot(l.res)
		if ca < 0 || cb < 0 || cr < 0 {
			t.Errorf("T1 has no slot for %s, %s or %s", l.a, l.b, l.res)
			continue
		}
		s, ok := file.Step(ca, cb)
		if !ok || s.Op != tfm.OpLig || int(s.Result) != cr {
			t.Errorf("%s followed by %s does not form %s: %+v", l.a, l.b, l.res, s)
		}
	}
	kerned := 0
	for _, x := range "abcdeghmnopqrstuvwxyz" {
		for _, y := range "abcdeghmnopqrstuvwxyz" {
			want := float64(it.kern.Kern(index[x], index[y]))
			got := file.KernBetween(int(x), int(y)).Float() * 1000
			if math.Abs(got-want) > 1 {
				t.Errorf("kern of %c%c is %.1f, kerning %.0f", x, y, got, want)
			}
			if want != 0 {
				kerned++
			}
		}
	}
	if kerned < 20 {
		t.Errorf("only %d of the pairs of lowercase letters are kerned", kerned)
	}
}

// TestItalicConverters runs TFtoPL and PLtoTF on the metric files of the
// italic: the PL text which TFtoPL writes is read by ReadPL, and PLtoTF gives
// back the bytes of the file.
func TestItalicConverters(t *testing.T) {
	t.Parallel()
	italicFont(t)
	tree := ran(t).tree
	tftopl, pltotf := needTool(t, "tftopl"), needTool(t, "pltotf")
	dir := t.TempDir()
	for _, e := range tex.Encodings {
		name := italicTFM(e)
		path := filepath.Join(dir, filepath.Base(name))
		if err := os.WriteFile(path, tree[name], 0o666); err != nil {
			t.Fatal(err)
		}
		var pl, msg bytes.Buffer
		cmd := exec.Command(tftopl, path)
		cmd.Stdout, cmd.Stderr = &pl, &msg
		if err := cmd.Run(); err != nil || msg.Len() > 0 {
			t.Errorf("%s: TFtoPL: %v %s", name, err, msg.String())
			continue
		}
		if _, err := tfm.ReadPL(bytes.NewReader(pl.Bytes())); err != nil {
			t.Errorf("%s: the PL text does not read: %v", name, err)
		}
		plPath, out := filepath.Join(dir, "in.pl"), filepath.Join(dir, "out.tfm")
		if err := os.WriteFile(plPath, pl.Bytes(), 0o666); err != nil {
			t.Fatal(err)
		}
		os.Remove(out)
		msg.Reset()
		cmd = exec.Command(pltotf, plPath, out)
		cmd.Stderr = &msg
		if err := cmd.Run(); err != nil || msg.Len() > 0 {
			t.Errorf("%s: PLtoTF: %v %s", name, err, msg.String())
			continue
		}
		back, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(back, tree[name]) {
			t.Errorf("%s: PLtoTF gives other bytes", name)
		}
	}
}

// ts1Symbols are the symbols of TS1 which the italic sets, by their commands
// and the characters which the text of the PDF file holds (the commands of
// textcomp).
var ts1Symbols = []struct {
	cmd string
	r   rune
}{
	{`\textcent`, '¢'}, {`\textsterling`, '£'}, {`\textcurrency`, '¤'}, {`\textyen`, '¥'},
	{`\textflorin`, 'ƒ'}, {`\textnaira`, '₦'}, {`\textwon`, '₩'}, {`\textpeso`, '₱'},
	{`\textlira`, '₤'}, {`\textdong`, '₫'}, {`\textbaht`, '฿'}, {`\textguarani`, '₲'},
	{`\textcolonmonetary`, '₡'}, {`\texttrademark`, '™'}, {`\textservicemark`, '℠'},
	{`\textcircledP`, '℗'}, {`\textnumero`, '№'}, {`\textrecipe`, '℞'}, {`\textcelsius`, '℃'},
	{`\textohm`, 'Ω'}, {`\textmho`, '℧'}, {`\textestimated`, '℮'},
	{`\textinterrobang`, '‽'}, {`\textinterrobangdown`, '⸘'}, {`\textreferencemark`, '※'},
	{`\textdblhyphen`, '⹀'}, {`\textlangle`, '〈'}, {`\textrangle`, '〉'},
	{`\textlbrackdbl`, '⟦'}, {`\textrbrackdbl`, '⟧'}, {`\textlquill`, '⁅'}, {`\textrquill`, '⁆'},
	{`\textbardbl`, '‖'}, {`\textopenbullet`, '◦'}, {`\textbigcircle`, '◯'},
	{`\textleftarrow`, '←'}, {`\textrightarrow`, '→'}, {`\textuparrow`, '↑'}, {`\textdownarrow`, '↓'},
	{`\textdiscount`, '⁒'}, {`\textmusicalnote`, '♪'},
	{`\textdollaroldstyle`, '$'}, {`\textcentoldstyle`, '¢'},
}

// ts1Pictures are the pictorial signs and the marks of TS1, whose code
// points the text of the PDF file does not tell from others: the log is
// looked at for lost characters.
const ts1Pictures = `\textborn\ \textdied\ \textmarried\ \textdivorced\ \textleaf\ \textOrogate\ \textorogate\` +
	`\textcompwordmark\ \textcapitalcompwordmark\ \textascendercompwordmark\ ` +
	`\textquotestraightbase\ \textquotestraightdblbase\ \texttwelveudash\ \textthreequartersemdash\ ` +
	`\texttildelow\ \textacutedbl\ \textgravedbl\ \textblank\ \textasteriskcentered\ \textsurd`

// italicDocument returns the test document: roman, italic and emphasised text
// with ligatures, kerns, italic corrections, every position of T1 and TS1
// which the italic has, and accents set with \accent.
func italicDocument(it *rendered) string {
	var s strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&s, format+"\n", args...) }
	w(`\documentclass{article}`)
	w(`\usepackage[T1]{fontenc}`)
	w(`\usepackage{textcomp}`)
	w(`\usepackage{bedern}`)
	w(`\tracinglostchars=3`)
	w(`\setlength{\parindent}{0pt}`)
	w(`\begin{document}`)
	w(`The office staff finds five affluent waffles.  \textit{The office staff finds five affluent waffles, with fj and fl and ffl.}\par`)
	w(`\emph{Emphasis in roman text, with \emph{nested emphasis} inside.}  A final \emph{of}) and \emph{If}: \emph{Tea} \emph{Wave} \emph{Yo}.\par`)
	w(`\textit{italic}\/ roman, \textit{f}\/ roman, \textit{Y}\/ roman, \itshape ffi\/ffl\/ \upshape upright, \textsl{slanted}, \textsl{slanted again}.\par`)
	w(`{\itshape caf\'e, na\"ive, \accent"01 a \accent"02 e \accent"03 o \accent"05 u \accent"09 i \accent"0A z \accent"07 n \accent"08 g, stra\ss e, \AE sop, \oe uvre, \L{}\'od\'z, \c c, \k a, \v s, \u g, \H o, \r a\par}`)
	w("{\\itshape ``Quoted'' --- text -- with dashes, <<guillemets>>, ,,low quotes'', 1234567890 (a) [b] \\$ \\& \\%% \\# \\_ ?` !`\\par}")
	w(`\textit{Symbols of TS1: \texttrademark\ \textregistered\ \texteuro\ \textdegree\ \textbullet\ \textonehalf}\par`)
	w(`{\itshape\fontencoding{TS1}\selectfont\typeout{TS1FONT=\fontname\font}}`)
	var syms strings.Builder
	for _, s := range ts1Symbols {
		syms.WriteString(s.cmd + `\ `)
	}
	w(`\textit{Italic symbols: %s}\par`, syms.String())
	w(`\textit{Italic figures: \oldstylenums{0123456789}}\par`)
	w(`\textit{Pictures and marks: %s}\par`, ts1Pictures)
	w(`\textit{Capital accents: \capitalgrave{A} \capitalacute{E} \capitalcircumflex{I} \capitaltilde{O} \capitaldieresis{U} \capitalhungarumlaut{O} \capitalring{A} \capitalcaron{C} \capitalbreve{G} \capitalmacron{E} \capitaldotaccent{Z} \capitaltie{AB} \newtie{ab} \capitalnewtie{AB} \t{ab}}\par`)
	w(`\fontencoding{T1}\fontfamily{bedern}\fontshape{it}\selectfont`)
	var slots strings.Builder
	index, _ := tex.T1.Resolve(it.tex)
	for c, i := range index {
		if i >= 0 {
			fmt.Fprintf(&slots, "\\char%d\\allowbreak{}", c)
		}
	}
	w(`{\raggedright %s\par}`, slots.String())
	slots.Reset()
	index, _ = tex.TS1.Resolve(it.tex)
	w(`\font\tsit=%s\tsit`, strings.TrimSuffix(filepath.Base(italicTFM(tex.TS1)), ".tfm"))
	for c, i := range index {
		if i >= 0 {
			fmt.Fprintf(&slots, "\\char%d\\allowbreak{}", c)
		}
	}
	w(`{\raggedright %s\par}`, slots.String())
	w(`\fontencoding{T1}\selectfont`)
	slots.Reset()
	index, _ = tex.T1.Resolve(it.tex)
	for c, i := range index {
		if i >= 0 && unicode.IsLetter(tex.T1.Slots[c].Rune) {
			fmt.Fprintf(&slots, "\\char%d\\allowbreak{}", c)
		}
	}
	w(`{\raggedright LETTERS %s\par}`, slots.String())
	w(`\end{document}`)
	return s.String()
}

// TestItalicLaTeX compiles a LaTeX document of roman, italic and emphasised
// text against the tree: the log has no lost character, no font shape which
// is not declared and no warning of pdfTeX, the shape it of TS1 is the metric
// file of the italic and no slanted shape of TS1 is used, the PDF file holds
// the italic embedded as a subset, and the text of the italic copies out as
// typed.
func TestItalicLaTeX(t *testing.T) {
	t.Parallel()
	it := italicFont(t)
	pdftotext := needTool(t, "pdftotext")
	pdflatex := needTool(t, "pdflatex")
	dir, tree := t.TempDir(), sharedTree(t)
	_, log, ok := compile(t, dir, tree, pdflatex, "italic", italicDocument(it))
	if !ok {
		return
	}
	flatLog := bytes.ReplaceAll(log, []byte("\n"), nil)
	// the shape it of TS1 is the metric file of the italic, and not the
	// slanted shape of the regular
	if !bytes.Contains(flatLog, []byte("TS1FONT="+strings.TrimSuffix(filepath.Base(italicTFM(tex.TS1)), ".tfm"))) {
		t.Errorf("the shape it of TS1 does not use %s", filepath.Base(italicTFM(tex.TS1)))
	}
	if bytes.Contains(flatLog, []byte("TS1/bedern/m/sl")) {
		t.Errorf("the log names the slanted shape of TS1")
	}
	for _, bad := range []string{"Missing character", "Font shape", "Font Warning", "undefined", "pdfTeX warning", "not loadable", "Overfull"} {
		if i := bytes.Index(flatLog, []byte(bad)); i >= 0 {
			t.Errorf("the log says %q: %s", bad, flatLog[i:min(len(flatLog), i+200)])
		}
	}
	pdf := filepath.Join(dir, "italic.pdf")
	out, err := exec.Command(needTool(t, "pdffonts"), pdf).Output()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) < 5 || !strings.HasSuffix(f[0], "+"+assembly.Style().PostScript) {
			continue
		}
		found = true
		if emb, sub := f[len(f)-5], f[len(f)-4]; emb != "yes" || sub != "yes" {
			t.Errorf("the font is not embedded as a subset: %s", line)
		}
	}
	if !found {
		t.Errorf("pdffonts does not list %s:\n%s", assembly.Style().PostScript, out)
	}
	text, err := exec.Command(pdftotext, "-enc", "UTF-8", pdf, "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	flat := strings.Join(strings.Fields(norm.NFKC.String(string(text))), "")
	var letters strings.Builder
	index, _ := tex.T1.Resolve(it.tex)
	for c, i := range index {
		if i >= 0 && unicode.IsLetter(tex.T1.Slots[c].Rune) {
			letters.WriteRune(tex.T1.Slots[c].Rune)
		}
	}
	var symbols strings.Builder
	for _, s := range ts1Symbols {
		symbols.WriteRune(s.r)
	}
	for _, want := range []string{
		"Italic symbols: " + norm.NFKC.String(symbols.String()),
		"Italic figures: 0123456789",
		"LETTERS " + norm.NFKC.String(letters.String()),
		"The office staff finds five affluent waffles, with fj and fl and ffl.",
		"Emphasis in roman text, with nested emphasis inside.",
		"A final of) and If: Tea Wave Yo.",
		"italic roman, f roman, Y roman, ffiffl",
		"upright, slanted, slanted again.",
		"café, naïve,",
		"straße, Æsop, œuvre, Łódź,",
	} {
		if !strings.Contains(flat, strings.Join(strings.Fields(want), "")) {
			t.Errorf("the text does not hold %q", want)
		}
	}
}
