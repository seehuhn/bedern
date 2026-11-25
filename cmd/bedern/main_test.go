// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
	"seehuhn.de/go/bedern/internal/font/mathfont"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/tex"
	"seehuhn.de/go/bedern/internal/tfm"
	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
)

// The tests share the builds of the glyphs.
func init() { glyphSet = glyphstest.Set }

// runTime is the time at which the tests render the fonts, and runDate its
// date.
const (
	runTime = 1700000000
	runDate = "2023-11-14"
)

// version is the version of the fonts.
const version = "0.001"

// A runResult holds what run wrote into the directories which it created:
// the names of the font files and the contents of each, the files for TeX by
// their paths in the tree, and the fonts as they were built.
type runResult struct {
	names []string
	files map[string][]byte
	tree  map[string][]byte
	fonts []*rendered
}

var (
	runOnce   sync.Once
	runOut    runResult
	runFailed error
)

// ran returns what the steps of run, called once for all tests at runTime in
// a temporary directory of the first test which asks, have written.
func ran(t *testing.T) runResult {
	t.Helper()
	runOnce.Do(func() {
		base := t.TempDir()
		dir, tree := filepath.Join(base, "fonts"), filepath.Join(base, "tree")
		now := time.Unix(runTime, 0).UTC()
		if runOut.fonts, runFailed = renderAll(now, true, true); runFailed != nil {
			return
		}
		// The routine tier leaves out Bedern Math, which costs about 15
		// seconds of processor time, and the italic (see TestMain).
		var math map[string][]byte
		if !testing.Short() {
			var err error
			if math, err = renderMath(now, true); err != nil {
				runFailed = err
				return
			}
		}
		if runFailed = write(dir, tree, runOut.fonts, math, now); runFailed != nil {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			runFailed = err
			return
		}
		runOut.files = map[string][]byte{}
		for _, e := range entries {
			runOut.names = append(runOut.names, e.Name())
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				runFailed = err
				return
			}
			runOut.files[e.Name()] = data
		}
		runOut.tree = map[string][]byte{}
		runFailed = filepath.WalkDir(tree, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(tree, path)
			if err != nil {
				return err
			}
			runOut.tree[filepath.ToSlash(rel)], err = os.ReadFile(path)
			return err
		})
	})
	if runFailed != nil {
		t.Fatal(runFailed)
	}
	return runOut
}

var (
	texOnce    sync.Once
	texFilesOf map[string][]byte
	texReports []tex.Report
	texErr     error
)

// texOf returns the files for TeX of the fonts of ran and the reports on the
// metric files, made once for all tests.
func texOf(t *testing.T) (map[string][]byte, []tex.Report) {
	t.Helper()
	r := ran(t)
	texOnce.Do(func() {
		texFilesOf, texReports, texErr = texFiles(r.fonts, time.Unix(runTime, 0).UTC())
	})
	if texErr != nil {
		t.Fatal(texErr)
	}
	return texFilesOf, texReports
}

var (
	treeOnce sync.Once
	treeDir  string
	treeErr  error
)

// sharedTree returns the directory in which the tree of the generated files
// is written, once for all tests.  TestMain removes it.
func sharedTree(t *testing.T) string {
	t.Helper()
	r := ran(t)
	treeOnce.Do(func() {
		if treeDir, treeErr = os.MkdirTemp("", "bedern-tree-"); treeErr == nil {
			treeErr = writeFiles(treeDir, r.tree)
		}
	})
	if treeErr != nil {
		t.Fatal(treeErr)
	}
	return treeDir
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// The routine tier leaves out the italic, which costs about 20
		// seconds of processor time; the tree then has the italic shape
		// of the slanted instance.
		fonts = slices.DeleteFunc(slices.Clone(fonts), func(f fontSpec) bool { return f.italic })
	}
	code := m.Run()
	if treeDir != "" {
		os.RemoveAll(treeDir)
	}
	os.Exit(code)
}

// TestReproducible renders the regular font at the time at which run wrote its
// file, which gives the same bytes, and the bold font likewise, and renders the
// regular font at another time, which gives other bytes.  It runs in the full
// tier only: each rendering costs 2.5 seconds of processor time.
func TestReproducible(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("the full tier renders the fonts again")
	}
	type job struct {
		f    fontSpec
		sec  int64
		same bool
	}
	jobs := []job{{fonts[0], runTime, true}, {fonts[1], runTime, true}, {fonts[0], runTime + 3*86400, false}}
	out := make([][]byte, len(jobs))
	errs := make([]error, len(jobs))
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			errs[i] = renderFont(&buf, j.f, time.Unix(j.sec, 0).UTC(), true)
			out[i] = buf.Bytes()
		}()
	}
	files := ran(t).files
	wg.Wait()
	for i, j := range jobs {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if same := bytes.Equal(out[i], files[j.f.file]); same != j.same {
			t.Errorf("%s at %d: identical bytes %t, want %t", j.f.file, j.sec, same, j.same)
		}
	}
}

// TestRun checks that run writes the font files and the licence, and nothing
// else, into a directory which it creates.
func TestRun(t *testing.T) {
	t.Parallel()
	r := ran(t)
	want := []string{bedern.LicenseFile}
	files := fontFiles()
	if !testing.Short() {
		want = append(want, mathfont.File, mathfont.RangesFile)
		files = append(files, mathfont.File)
	}
	want = append(want, fontFiles()...)
	slices.Sort(want)
	if !slices.Equal(r.names, want) {
		t.Errorf("files %v, want %v", r.names, want)
	}
	if string(r.files[bedern.LicenseFile]) != bedern.OFL {
		t.Error("OFL.txt differs from the embedded licence")
	}
	for _, file := range files {
		if !bytes.HasPrefix(r.files[file], []byte("OTTO")) {
			t.Errorf("%s does not start with the tag of a CFF font", file)
		}
	}
	if testing.Short() {
		return
	}
	ranges := string(r.files[mathfont.RangesFile])
	for _, want := range []string{`\def\BedernMathRanges{`, `\setBedernMath{`, mathfont.File,
		"SPDX-License-Identifier: LPPL-1.3c", "% This work consists of the files "} {
		if !strings.Contains(ranges, want) {
			t.Errorf("%s lacks %s", mathfont.RangesFile, want)
		}
	}
	if !strings.Contains(ranges, " "+mathfont.RangesFile+", ") {
		t.Errorf("the notice of %s does not list the file", mathfont.RangesFile)
	}
	if tex.MathRangesFile != mathfont.RangesFile {
		t.Errorf("the notices name %s, the command writes %s", tex.MathRangesFile, mathfont.RangesFile)
	}
}

// fontFiles returns the names of the text font files.
func fontFiles() []string {
	var out []string
	for _, f := range fonts {
		out = append(out, f.file)
	}
	return out
}

// TestMathFont checks the names and the tables of the file of the
// mathematics font.
func TestMathFont(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("the full tier builds Bedern Math")
	}
	data := ran(t).files[mathfont.File]
	tables := readTables(t, mathfont.File, data)
	for _, tag := range []string{"MATH", "GSUB", "CFF ", "cmap"} {
		if len(tables[tag]) == 0 {
			t.Errorf("%s has no table %s", mathfont.File, tag)
		}
	}
	got := readFamilyFields(t, mathfont.File, data)
	for key, tab := range got.names {
		if tab.Family != mathfont.Family || tab.Subfamily != "Regular" {
			t.Errorf("%s, %s: family %q %q, want %q Regular", mathfont.File, key, tab.Family, tab.Subfamily, mathfont.Family)
		}
	}
}

// familyFields are the fields of a font file which name it and its style.
type familyFields struct {
	names     map[string]*name.Table // the name tables, by platform and language
	weight    uint16                 // usWeightClass of the OS/2 table
	selection uint16                 // fsSelection of the OS/2 table
	macStyle  uint16                 // macStyle of the head table
}

// readFamilyFields reads the fields of familyFields from a font file.
func readFamilyFields(t *testing.T, file string, data []byte) familyFields {
	t.Helper()
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	table := func(tag string, n int) []byte {
		b, err := h.ReadTableBytes(r, tag)
		if err != nil {
			t.Fatalf("%s: %s: %v", file, tag, err)
		}
		if len(b) < n {
			t.Fatalf("%s: the table %s has %d bytes", file, tag, len(b))
		}
		return b
	}
	info, err := name.Decode(table("name", 6))
	if err != nil {
		t.Fatalf("%s: name: %v", file, err)
	}
	f := familyFields{names: map[string]*name.Table{}}
	for lang, tab := range info.Mac {
		f.names["Mac "+lang] = tab
	}
	for lang, tab := range info.Windows {
		f.names["Windows "+lang] = tab
	}
	os2 := table("OS/2", 64)
	f.weight = binary.BigEndian.Uint16(os2[4:])
	f.selection = binary.BigEndian.Uint16(os2[62:])
	f.macStyle = binary.BigEndian.Uint16(table("head", 46)[44:])
	return f
}

// TestFamily checks that the font files form one family: the same family
// name, version, copyright and licence, the subfamily, full name, PostScript
// name and unique identifier of each style, and the weight class, the bits
// of fsSelection (bold 0x20, regular 0x40, italic 0x01) and of macStyle (bold
// 0x01, italic 0x02).
func TestFamily(t *testing.T) {
	t.Parallel()
	files := ran(t).files
	want := []struct {
		file, sub, ps    string
		weight, sel, mac uint16
	}{
		{"Bedern-Regular.otf", "Regular", "Bedern-Regular", 400, 0x40, 0},
		{"Bedern-Bold.otf", "Bold", "Bedern-Bold", 700, 0x20, 0x01},
		{"Bedern-Italic.otf", "Italic", "Bedern-Italic", 400, 0x01, 0x02},
	}
	versions := map[string]bool{}
	ids := map[string]string{}
	for _, w := range want {
		if !slices.Contains(fontFiles(), w.file) {
			continue
		}
		f := readFamilyFields(t, w.file, files[w.file])
		if f.weight != w.weight {
			t.Errorf("%s: weight class %d, want %d", w.file, f.weight, w.weight)
		}
		if got := f.selection & 0x61; got != w.sel {
			t.Errorf("%s: fsSelection bits %#x, want %#x", w.file, got, w.sel)
		}
		if got := f.macStyle & 0x03; got != w.mac {
			t.Errorf("%s: macStyle bits %#x, want %#x", w.file, got, w.mac)
		}
		if len(f.names) < 2 {
			t.Errorf("%s: %d name tables, want one for Mac and one for Windows", w.file, len(f.names))
		}
		full := "Bedern " + w.sub
		for key, tab := range f.names {
			check := func(field, got, want string) {
				if got != want {
					t.Errorf("%s, %s: %s %q, want %q", w.file, key, field, got, want)
				}
			}
			check("family", tab.Family, "Bedern")
			check("subfamily", tab.Subfamily, w.sub)
			check("full name", tab.FullName, full)
			check("PostScript name", tab.PostScriptName, w.ps)
			check("typographic family", tab.TypographicFamily, "")
			check("typographic subfamily", tab.TypographicSubfamily, "")
			check("copyright", tab.Copyright, bedern.Copyright)
			check("licence", tab.License, bedern.License)
			check("licence URL", tab.LicenseURL, bedern.LicenseURL)
			check("version", tab.Version, "Version "+version)
			check("unique identifier", tab.Identifier, full+"; "+version+"; "+runDate)
			if other, ok := ids[tab.Identifier]; ok && other != w.file {
				t.Errorf("%s and %s share the unique identifier %q", w.file, other, tab.Identifier)
			}
			ids[tab.Identifier] = w.file
			versions[tab.Version] = true
		}
	}
	if len(versions) != 1 {
		t.Errorf("the files carry the versions %v, want one", versions)
	}
}

// TestUndated checks that the font files without a date of the build differ
// from the dated ones only in the unique identifier, which is the full name
// and the version.
func TestUndated(t *testing.T) {
	t.Parallel()
	files := ran(t).files
	for _, f := range fonts {
		dated := files[f.file]
		undated, err := build.Undated(dated)
		if err != nil {
			t.Fatal(err)
		}
		a, b := readTables(t, f.file, dated), readTables(t, f.file, undated)
		if len(a) != len(b) {
			t.Errorf("%s: %d tables, want %d", f.file, len(b), len(a))
		}
		for tag, data := range a {
			if tag != "name" && tag != "head" && !bytes.Equal(data, b[tag]) {
				t.Errorf("%s: the table %s differs", f.file, tag)
			}
		}
		// The head table differs in checkSumAdjustment only.
		ha, hb := a["head"], b["head"]
		if len(ha) != len(hb) || !bytes.Equal(ha[:8], hb[:8]) || !bytes.Equal(ha[12:], hb[12:]) {
			t.Errorf("%s: the table head differs beyond its checksum", f.file)
		}
		full := readFamilyFields(t, f.file, undated)
		want := readFamilyFields(t, f.file, dated)
		for key, tab := range full.names {
			w := *want.names[key]
			w.Identifier = f.style().Family + " " + f.style().Subfamily + "; " + version
			if !reflect.DeepEqual(*tab, w) {
				t.Errorf("%s, %s: name table %+v, want %+v", f.file, key, *tab, w)
			}
		}
	}
}

// readTables returns the tables of a font file by tag.
func readTables(t *testing.T, file string, data []byte) map[string][]byte {
	t.Helper()
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	out := map[string][]byte{}
	for tag := range h.Toc {
		b, err := h.ReadTableBytes(r, tag)
		if err != nil {
			t.Fatalf("%s: %s: %v", file, tag, err)
		}
		out[tag] = b
	}
	return out
}

// TestTree checks the files for TeX which run writes: the fonts in both
// formats, the licence, the metric files of both fonts in both encodings and
// the text files, at their places in the tree.
func TestTree(t *testing.T) {
	t.Parallel()
	r := ran(t)
	want := []string{
		"doc/fonts/bedern/OFL.txt",
		"fonts/enc/dvips/bedern/bedern-t1.enc",
		"fonts/enc/dvips/bedern/bedern-ts1.enc",
		"fonts/map/dvips/bedern/bedern.map",
		"fonts/opentype/public/bedern/Bedern-Bold.otf",
		"fonts/opentype/public/bedern/Bedern-Regular.otf",
		"fonts/tfm/public/bedern/bedern-bold-sl-t1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-sl-ts1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-t1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-ts1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-sl-t1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-sl-ts1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-t1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-ts1.tfm",
		"fonts/type1/public/bedern/Bedern-Bold.pfb",
		"fonts/type1/public/bedern/Bedern-Regular.pfb",
		"tex/generic/bedern/bedern-glyphtounicode.tex",
		"tex/latex/bedern/bedern.sty",
		"tex/latex/bedern/t1bedern.fd",
		"tex/latex/bedern/ts1bedern.fd",
		"tex/plain/bedern/bedern.tex",
	}
	if slices.Contains(fontFiles(), "Bedern-Italic.otf") {
		want = append(want,
			"fonts/opentype/public/bedern/Bedern-Italic.otf",
			"fonts/tfm/public/bedern/bedern-italic-t1.tfm",
			"fonts/tfm/public/bedern/bedern-italic-ts1.tfm",
			"fonts/type1/public/bedern/Bedern-Italic.pfb")
		slices.Sort(want)
	}
	var got []string
	for path := range r.tree {
		got = append(got, path)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("files %v, want %v", got, want)
	}
	for _, f := range fonts {
		if !bytes.Equal(r.tree[tex.DirOpenType+"/"+f.file], r.files[f.file]) {
			t.Errorf("%s in the tree differs from the font file", f.file)
		}
	}
	if string(r.tree[tex.DirDoc+"/"+bedern.LicenseFile]) != bedern.OFL {
		t.Error("OFL.txt in the tree differs from the embedded licence")
	}
}

// needTool returns the path of the named external tool.  If the tool is
// missing, the test is skipped, unless the environment variable
// BEDERN_REQUIRE_TOOLS is 1, in which case the test fails.
func needTool(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		if os.Getenv("BEDERN_REQUIRE_TOOLS") == "1" {
			t.Fatalf("required tool %s not found (BEDERN_REQUIRE_TOOLS=1): %v", name, err)
		}
		t.Skipf("%s not found", name)
	}
	return path
}

// TestTeX compiles a plain document with pdftex and a LaTeX document with
// pdflatex against the tree which run writes, which the engines find through
// the search paths of their file types.  Both
// set every character of T1 which the font has and text in both fonts; the
// logs report no missing character and no substituted font, and the engines
// embed both fonts from the PFB files of the tree.
func TestTeX(t *testing.T) {
	t.Parallel()
	pdftex, pdflatex := needTool(t, "pdftex"), needTool(t, "pdflatex")
	r := ran(t)
	dir, tree := t.TempDir(), sharedTree(t)
	index, _ := tex.T1.Resolve(r.fonts[0].glyphs)
	var slots strings.Builder
	for c, i := range index {
		if i >= 0 {
			fmt.Fprintf(&slots, "\\char%d ", c)
		}
	}
	text := "The office staff finds five affluent waffles, ``AVATAR''---To We."
	docs := []struct {
		engine, name, src string
	}{
		{pdftex, "plain", "\\tracinglostchars=1 \\input bedern\n" + text + " caf\\'e \\ss\\ae\n\n" +
			"{\\bf " + text + "}\n\n" + slots.String() + "\n\n" +
			"\\font\\sym=bedern-regular-ts1 at 7pt \\sym \\char\"A9\n\\bye\n"},
		{pdflatex, "latex", "\\documentclass{article}\\usepackage{bedern}\\begin{document}\n" + text +
			" caf\\'e \\ss\\ae\n\n\\textbf{" + text + "}\n\n" + slots.String() +
			"\n\n\\textcopyright\\ \\textdegree\n\\end{document}\n"},
	}
	for _, d := range docs {
		out, log, ok := compile(t, dir, tree, d.engine, d.name, d.src)
		if !ok {
			continue
		}
		for _, bad := range []string{"Missing character", "Font Warning", "not loadable"} {
			if bytes.Contains(log, []byte(bad)) {
				t.Errorf("%s: the log says %q", d.name, bad)
			}
		}
		for _, pfb := range []string{"Bedern-Regular.pfb>", "Bedern-Bold.pfb>"} {
			if !bytes.Contains(bytes.ReplaceAll(out, []byte("\n"), nil), []byte(pfb)) {
				t.Errorf("%s: the output does not name %s", d.name, pfb)
			}
		}
	}
}

// compile compiles the document src, which it writes to dir as name.tex, with
// the engine against the tree, which the engine finds through the search
// paths of its file types.  It returns the output and the log of the engine
// and false if the engine failed or wrote no PDF file.
func compile(t *testing.T, dir, tree, engine, name, src string) (out, log []byte, ok bool) {
	t.Helper()
	path := filepath.Join(dir, name+".tex")
	if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(engine, "-interaction=nonstopmode", "-halt-on-error", name+".tex")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for variable, sub := range map[string]string{
		"TEXINPUTS":   "tex",
		"TFMFONTS":    tex.DirTFM,
		"T1FONTS":     tex.DirPFB,
		"ENCFONTS":    tex.DirEnc,
		"TEXFONTMAPS": tex.DirMap,
	} {
		// the empty last entry stands for the default path
		cmd.Env = append(cmd.Env, variable+"="+filepath.Join(tree, filepath.FromSlash(sub))+"//"+string(os.PathListSeparator))
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("%s: %v\n%s", name, err, out)
		return nil, nil, false
	}
	log, err = os.ReadFile(filepath.Join(dir, name+".log"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, name+".pdf")); err != nil {
		t.Errorf("%s: no PDF file: %v", name, err)
		return nil, nil, false
	}
	return out, log, true
}

// limitHeight is the largest change which the reduction of the heights makes
// in a metric file of T1 or TS1 with the lines of the design fixed, in units:
// 21 in T1 and 23 in TS1, where PLtoTF's method without fixed lines needs 17 and 19.5.
const limitHeight = 23

// beginsLigature reports whether c is the first character of a ligature of
// the metric file, the one which TeX kerns before the ligature forms (see
// TestKernBeforeLigature).
func beginsLigature(file *tfm.File, c int) bool {
	for _, s := range file.Program(c) {
		if s.Op.IsLig() {
			return true
		}
	}
	return false
}

// TestMetricFiles reads the metric files of T1 back and compares them with
// the fonts as they were built: the width of every character is the advance
// of its glyph, the kern of every pair of characters is the kern of the
// glyphs, the ligatures of f give the glyphs which stand for the letters, the
// letters x, H and p have the x-height, the height of the capitals and the
// depth of the descenders, the interword space is the advance of the space,
// the face byte names the weight and the slope,
// and the reduction of the heights and depths to the sizes of the format
// changes no value by more than a fortieth of the em.
func TestMetricFiles(t *testing.T) {
	t.Parallel()
	r := ran(t)
	_, reports := texOf(t)
	for _, rep := range reports {
		for _, red := range rep.Reductions {
			limit := 15
			if red.Table == "italic" {
				limit = 63
			}
			if red.After > limit {
				t.Errorf("%s: %d values of the %s, want at most %d", rep.TFM, red.After, red.Table, limit)
			}
			if red.Change > limitHeight {
				t.Errorf("%s: the reduction changes a %s by %g units", rep.TFM, red.Table, red.Change)
			}
		}
	}
	fix := func(v float64) tfm.Fix { return tfm.FixFromFloat(v / 1000) }
	for i, f := range fonts {
		tf := &tex.Font{Style: f.style()}
		name := tex.DirTFM + "/" + tf.TFM(tex.T1) + ".tfm"
		file, err := tfm.Read(bytes.NewReader(r.tree[name]))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		p, gs, table := r.fonts[i].params, r.fonts[i].glyphs, r.fonts[i].kern
		index, _ := tex.T1.Resolve(gs)
		kerns := 0
		for a, ia := range index {
			if file.Exists(a) != (ia >= 0) {
				t.Errorf("%s: the character %#02x exists: %t, its glyph: %t", name, a, file.Exists(a), ia >= 0)
			}
			if ia < 0 {
				continue
			}
			if len(gs[ia].Path) > 0 {
				// the reduction never makes a height or a depth zero
				_, y0, _, y1 := gs[ia].Path.BBox()
				if y1 > 0 && file.CharHeight(a) == 0 {
					t.Errorf("%s: %s reaches %g above the baseline and has the height 0", name, gs[ia].Name, y1)
				}
				if y0 < -p.OvershootUpper && file.CharDepth(a) == 0 {
					t.Errorf("%s: %s reaches %g below the baseline and has the depth 0", name, gs[ia].Name, -y0)
				}
			}
			if got, want := file.CharWidth(a), fix(gs[ia].Advance); got != want {
				t.Errorf("%s: width of %s %v, want %v", name, gs[ia].Name, got, want)
			}
			for b, ib := range index {
				if ib < 0 {
					continue
				}
				step, ok := file.Step(a, b)
				if ok && step.Op.IsLig() {
					continue
				}
				want := fix(float64(table.Kern(ia, ib)))
				if got := file.KernBetween(a, b); got != want && !beginsLigature(file, b) {
					t.Errorf("%s: kern of %s and %s %v, want %v", name, gs[ia].Name, gs[ib].Name, got, want)
				}
				if want != 0 {
					kerns++
				}
			}
		}
		if kerns == 0 {
			t.Errorf("%s has no kerns", name)
		}
		for _, g := range gs {
			if len(g.Parts) == 0 || g.Rune == 0 {
				continue
			}
			// the ligature is reached from its first letter, one letter at a time
			c := int(g.Parts[0][0])
			for _, part := range g.Parts[1:] {
				step, ok := file.Step(c, int(part[0]))
				if !ok || step.Op != tfm.OpLig {
					t.Errorf("%s: no ligature on the way to %s", name, g.Name)
					c = -1
					break
				}
				c = int(step.Result)
			}
			if c >= 0 && gs[index[c]].Name != g.Name {
				t.Errorf("%s: the letters of %s give %s", name, g.Name, gs[index[c]].Name)
			}
		}
		// the lines of the design are fixed values of the reduction
		for _, c := range []struct {
			what      string
			got, want tfm.Fix
		}{
			{"height of x", file.CharHeight('x'), fix(p.XHeight)},
			{"height of H", file.CharHeight('H'), fix(p.CapHeight)},
			{"height of d", file.CharHeight('d'), fix(p.Ascender)},
			{"depth of p", file.CharDepth('p'), fix(p.Descender)},
		} {
			if c.got != c.want {
				t.Errorf("%s: %s %v, want %v", name, c.what, c.got, c.want)
			}
		}
		space := 0.0
		for _, g := range gs {
			if g.Name == "space" {
				space = g.Advance
			}
		}
		if len(file.Param) != 16 || file.Param[1] != fix(space) || file.Param[4] != fix(p.XHeight) ||
			file.Param[7] != fix(p.CapHeight) || file.Param[8] != fix(p.Ascender) ||
			file.Param[10] != fix(p.Descender) || file.Param[15] != fix(p.TypoAscender-p.TypoDescender+p.LineGap) {
			t.Errorf("%s: parameters %v", name, file.Param)
		}
		wantFace := byte(0)
		switch {
		case f.style().Bold:
			wantFace = 2
		case f.style().Italic:
			wantFace = 1
		}
		if face, ok := file.Face(); !ok || face != wantFace {
			t.Errorf("%s: face byte %d, want %d", name, face, wantFace)
		}
	}
}

// ts1Commands lists the text commands of TS1 (without the backslash).
const ts1Commands = "textacutedbl textascendercompwordmark textasciiacute textasciibreve textasciicaron " +
	"textasciidieresis textasciigrave textasciimacron textbaht textbardbl textbigcircle " +
	"textblank textborn textbrokenbar textbullet textcapitalcompwordmark textcelsius textcent " +
	"textcentoldstyle textcircled textcolonmonetary textcopyleft textcopyright textcurrency " +
	"textdagger textdaggerdbl textdblhyphen textdblhyphenchar textdegree textdied textdiscount " +
	"textdiv textdivorced textdollar textdollaroldstyle textdong textdownarrow " +
	"texteightoldstyle textestimated texteuro textfiveoldstyle textflorin textfouroldstyle " +
	"textfractionsolidus textgravedbl textguarani textinterrobang textinterrobangdown " +
	"textlangle textlbrackdbl textleaf textleftarrow textlira textlnot textlquill textmarried " +
	"textmho textminus textmu textmusicalnote textnaira textnineoldstyle textnumero textohm " +
	"textonehalf textoneoldstyle textonequarter textonesuperior textopenbullet textordfeminine " +
	"textordmasculine textparagraph textperiodcentered textpertenthousand textperthousand " +
	"textpeso textpilcrow textpm textquotesingle textquotestraightbase textquotestraightdblbase " +
	"textrangle textrbrackdbl textrecipe textreferencemark textregistered textrightarrow " +
	"textrquill textsection textservicemark textsevenoldstyle textsixoldstyle textsterling " +
	"textsurd textthreeoldstyle textthreequarters textthreequartersemdash textthreesuperior " +
	"texttildelow texttimes texttrademark texttwelveudash texttwooldstyle texttwosuperior " +
	"textuparrow textwon textyen textzerooldstyle"

// pdfFonts returns the names of the fonts of the PDF file, as pdffonts lists
// them.
func pdfFonts(t *testing.T, pdf string) []string {
	t.Helper()
	out, err := exec.Command(needTool(t, "pdffonts"), pdf).Output()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n")[2:] {
		if f := strings.Fields(line); len(f) > 0 {
			names = append(names, f[0])
		}
	}
	return names
}

// TestTS1Commands sets every text command of TS1 with pdfLaTeX: nothing is
// missing from the metric file and every sign comes from Bedern, so that the
// subset which the font definition file of TS1 declares leaves no sign to
// another font and none empty.
func TestTS1Commands(t *testing.T) {
	t.Parallel()
	needTool(t, "pdftex")
	pdflatex := needTool(t, "pdflatex")
	r := ran(t)
	dir, tree := t.TempDir(), sharedTree(t)
	var src strings.Builder
	src.WriteString("\\documentclass{article}\\usepackage{bedern}\\begin{document}\\tracinglostchars=1\n")
	for _, c := range strings.Fields(ts1Commands) {
		if c == "textcircled" {
			src.WriteString("\\textcircled{a}\n")
		} else {
			fmt.Fprintf(&src, "x\\%s{}x\n", c)
		}
	}
	src.WriteString("\\end{document}\n")
	_, log, ok := compile(t, dir, tree, pdflatex, "ts1", src.String())
	if !ok {
		return
	}
	for _, bad := range []string{"Missing character", "Font Warning", "undefined"} {
		if bytes.Contains(log, []byte(bad)) {
			t.Errorf("the log says %q", bad)
		}
	}
	for _, name := range pdfFonts(t, filepath.Join(dir, "ts1.pdf")) {
		if !strings.Contains(name, "+Bedern-") {
			t.Errorf("the font %s is not Bedern", name)
		}
	}
	fd := string(r.tree[tex.DirLaTeX+"/"+tex.FDFile(tex.TS1)])
	if !strings.Contains(fd, "\\DeclareEncodingSubset{TS1}{bedern}{0}\n") {
		t.Errorf("the font definition file of TS1 does not declare the subset 0:\n%s", fd)
	}
}

// TestShapes sets the slanted and the italic shape with pdfLaTeX and the
// fonts of \it, \sl, \tt and of the footline with plain pdfTeX: the shape sl
// of LaTeX comes from the slanted instance and the shape it from the italic
// font, which is in the PDF file of both engines if the command writes it,
// the accented letters and the page number of plain TeX come from Bedern in
// every font, and the typewriter font is the only other font.
func TestShapes(t *testing.T) {
	t.Parallel()
	pdftex, pdflatex := needTool(t, "pdftex"), needTool(t, "pdflatex")
	needTool(t, "pdffonts")
	dir, tree := t.TempDir(), sharedTree(t)
	hasItalic := slices.Contains(fontFiles(), assembly.File)
	body := "caf\\'e Stra\\ss e \\l\\'od\\'z na\\\"\\i ve"
	_, log, ok := compile(t, dir, tree, pdflatex, "shapes",
		"\\documentclass{article}\\usepackage{bedern}\\begin{document}\n"+
			"\\textsl{"+body+"} \\textit{"+body+"} \\textbf{\\textsl{x} \\textit{x}} \\emph{x}\n\\end{document}\n")
	if ok {
		for _, bad := range []string{"Missing character", "Font Warning", "undefined"} {
			if bytes.Contains(log, []byte(bad)) {
				t.Errorf("LaTeX: the log says %q", bad)
			}
		}
		names := pdfFonts(t, filepath.Join(dir, "shapes.pdf"))
		for _, name := range names {
			if !strings.Contains(name, "+Bedern-") {
				t.Errorf("LaTeX: the font %s is not Bedern", name)
			}
		}
		if got := slices.ContainsFunc(names, func(n string) bool { return strings.HasSuffix(n, "+Bedern-Italic") }); got != hasItalic {
			t.Errorf("LaTeX: the italic font is in the PDF file: %t, want %t", got, hasItalic)
		}
	}
	// one document for the four fonts: only the typewriter font is not Bedern
	var plain strings.Builder
	plain.WriteString("\\tracinglostchars=1 \\input bedern\n")
	for _, switches := range []string{"\\rm", "\\it", "\\sl", "\\tt"} {
		plain.WriteString("{" + switches + " " + body + "}\\par\n")
	}
	plain.WriteString("\\bye\n")
	if _, log, ok := compile(t, dir, tree, pdftex, "plain", plain.String()); ok {
		if bytes.Contains(log, []byte("Missing character")) {
			t.Error("plain: the log reports a missing character")
		}
		other, italic := 0, false
		for _, name := range pdfFonts(t, filepath.Join(dir, "plain.pdf")) {
			if !strings.Contains(name, "+Bedern-") {
				other++
			}
			italic = italic || strings.HasSuffix(name, "+Bedern-Italic")
		}
		if italic != hasItalic {
			t.Errorf("plain: the italic font is in the PDF file: %t, want %t", italic, hasItalic)
		}
		if other != 1 {
			t.Errorf("plain: %d fonts other than Bedern, want 1 (the typewriter font)", other)
		}
	}
}

// TestPackageOptions loads the package with the option of the figures, which
// pdfTeX cannot use, and with an option which does not exist: both give a
// warning and no error, and a global option of the class is no option of the
// package.
func TestPackageOptions(t *testing.T) {
	t.Parallel()
	pdflatex := needTool(t, "pdflatex")
	dir, tree := t.TempDir(), sharedTree(t)
	_, log, ok := compile(t, dir, tree, pdflatex, "options",
		"\\documentclass[12pt,a4paper]{article}\\usepackage[tabular,oldstyle,nonsense]{bedern}"+
			"\\begin{document}x\\end{document}\n")
	if !ok {
		return
	}
	for _, want := range []string{"Unknown option `nonsense'", "The options oldstyle have no effect under pdfTeX"} {
		if !bytes.Contains(bytes.ReplaceAll(log, []byte("\n"), nil), []byte(want)) {
			t.Errorf("the log lacks the warning %q", want)
		}
	}
	for _, bad := range []string{"Unknown option `12pt'", "Unknown option `a4paper'", "tabular"} {
		if bytes.Contains(log, []byte(bad)) {
			t.Errorf("the log says %q", bad)
		}
	}
}

// TestSlantedMetricFiles reads the metric files of the slanted instances back:
// the slant parameter is the slant of the instance, the widths are those of the
// upright font, the face byte names an italic slope and the map file slants
// the outlines by the same amount.
func TestSlantedMetricFiles(t *testing.T) {
	t.Parallel()
	r := ran(t)
	mapFile := string(r.tree[tex.DirMap+"/"+tex.MapFile])
	for _, f := range fonts {
		if f.italic {
			continue
		}
		upright := &tex.Font{Style: f.style()}
		slanted := &tex.Font{Style: f.style(), Slant: tex.SlantRatio}
		for _, e := range tex.Encodings {
			read := func(tf *tex.Font) *tfm.File {
				name := tex.DirTFM + "/" + tf.TFM(e) + ".tfm"
				file, err := tfm.Read(bytes.NewReader(r.tree[name]))
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				return file
			}
			a, b := read(upright), read(slanted)
			if got, want := b.Param[0], tfm.FixFromFloat(tex.SlantRatio); got != want {
				t.Errorf("%s: slant %v, want %v", slanted.TFM(e), got, want)
			}
			if a.Param[0] != 0 {
				t.Errorf("%s: slant %v, want 0", upright.TFM(e), a.Param[0])
			}
			wantFace := byte(1)
			if f.style().Bold {
				wantFace = 3
			}
			if face, ok := b.Face(); !ok || face != wantFace {
				t.Errorf("%s: face byte %d, want %d", slanted.TFM(e), face, wantFace)
			}
			for c := range 256 {
				if a.Exists(c) != b.Exists(c) || a.CharWidth(c) != b.CharWidth(c) {
					t.Errorf("%s: the character %#02x differs from the upright font", slanted.TFM(e), c)
					break
				}
			}
			line := slanted.TFM(e) + " " + f.style().PostScript + " \"0.1667 SlantFont " + e.VectorName() + " ReEncodeFont\""
			if !strings.Contains(mapFile, line) {
				t.Errorf("the map file lacks the line %s", line)
			}
		}
	}
}

// TestKernBeforeLigature runs the ligature and kerning programs of the metric
// files of T1 as TeX does on a letter followed by the characters of a
// ligature, and compares the kern which TeX inserts after the letter with the
// kerns of the font.  TeX kerns the letter with the first character before it
// knows that a ligature forms, so that one kern stands for the character and
// for the ligatures which it begins.  The kern is that of the ligature for the
// less-than and the greater-than sign, that of the character for the
// exclamation and the question mark, and otherwise the one of least size
// among them; where the characters and the ligature differ by no more than
// 5 units, it is the kern of the character.
func TestKernBeforeLigature(t *testing.T) {
	t.Parallel()
	r := ran(t)
	files, _ := texOf(t)
	seqs := []struct {
		typed string
		form  int // the character which forms
		first int
	}{
		{"--", 0x15, '-'}, {"---", 0x16, '-'}, {"<<", 0x13, '<'}, {">>", 0x14, '>'},
		{",,", 0x12, ','}, {"``", 0x10, '`'}, {"''", 0x11, '\''},
		{"!`", 0xBD, '!'}, {"?`", 0xBE, '?'},
		{"ff", 0x1B, 'f'}, {"fi", 0x1C, 'f'}, {"fl", 0x1D, 'f'}, {"ffi", 0x1E, 'f'}, {"ffl", 0x1F, 'f'},
	}
	for i, f := range fonts {
		tf := &tex.Font{Style: f.style()}
		name := tex.DirTFM + "/" + tf.TFM(tex.T1) + ".tfm"
		file, err := tfm.Read(bytes.NewReader(files[name]))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		gs, table := r.fonts[i].glyphs, r.fonts[i].kern
		index, _ := tex.T1.Resolve(gs)
		units := func(k tfm.Fix) float64 { return k.Float() * 1000 }
		pairs, off, before, worst := 0, 0, 0, 0.0
		for l, il := range index {
			if il < 0 {
				continue
			}
			for _, s := range seqs {
				if step, ok := file.Step(l, s.first); ok && step.Op.IsLig() {
					continue // the letter forms a ligature with the first character
				}
				codes := []int{l}
				for _, c := range []byte(s.typed) {
					codes = append(codes, int(c))
				}
				items, err := file.Typeset(codes)
				if err != nil {
					t.Fatal(err)
				}
				// the letter, the kern if there is one, and the ligature
				applied, last := 0.0, items[len(items)-1]
				if len(items) == 3 && items[1].Char < 0 {
					applied = units(items[1].Kern)
				} else if len(items) != 2 {
					t.Fatalf("%s: %q after %#02x gives %v", name, s.typed, l, items)
				}
				if last.Char != s.form {
					t.Fatalf("%s: %q after %#02x gives the character %#02x, want %#02x", name, s.typed, l, last.Char, s.form)
				}
				lone := float64(table.Kern(il, index[s.first]))
				lig := float64(table.Kern(il, index[s.form]))
				const eps = 0.01
				if s.first != '!' && s.first != '?' {
					pairs++
					if math.Abs(applied-lig) > 5 {
						off++
						worst = math.Max(worst, math.Abs(applied-lig))
					}
					if math.Abs(lone-lig) > 5 {
						before++
					}
				}
				// the kerns of all the forms of the character
				spread, least := 0.0, lone
				for _, o := range seqs {
					if o.first == s.first {
						k := float64(table.Kern(il, index[o.form]))
						spread = math.Max(spread, math.Abs(k-lone))
						if math.Abs(k) < math.Abs(least) {
							least = k
						}
					}
				}
				want := least
				switch {
				case spread <= 5, s.first == '!', s.first == '?':
					want = lone
				case s.first == '<' || s.first == '>':
					want = lig
				}
				if math.Abs(applied-want) > eps {
					t.Errorf("%s: %q after %#02x: kern %g, want %g (character %g, ligature %g)", name, s.typed, l, applied, want, lone, lig)
				}
			}
		}
		t.Logf("%s: %d pairs of a letter and a ligature (without those of ! and ?), %d with the kern of the character more than 5 units from that of the ligature before, %d after, at most %g", name, pairs, before, off, worst)
	}
}

// inkDistance returns the least distance between the outlines p and q, where
// q is shifted by shift to the right.
func inkDistance(p, q outline.Path, shift float64) float64 {
	toSegment := func(a, b, z outline.Point) float64 {
		dx, dy := b.X-a.X, b.Y-a.Y
		t := 0.0
		if l := dx*dx + dy*dy; l > 0 {
			t = math.Max(0, math.Min(1, ((z.X-a.X)*dx+(z.Y-a.Y)*dy)/l))
		}
		return math.Hypot(a.X+t*dx-z.X, a.Y+t*dy-z.Y)
	}
	best := math.Inf(1)
	for _, cp := range p {
		a := cp.Flatten()
		for _, cq := range q {
			b := cq.Flatten()
			for i := range b {
				b[i].X += shift
			}
			for i := range a {
				for _, z := range b {
					best = math.Min(best, toSegment(a[i], a[(i+1)%len(a)], z))
				}
			}
			for i := range b {
				for _, z := range a {
					best = math.Min(best, toSegment(b[i], b[(i+1)%len(b)], z))
				}
			}
		}
	}
	return best
}

// TestFBeforeDieresis sets f before the letters i with a mark in the way of
// the metric files, with the advance of f and the kern of the file, and asks
// that the ink of f keep the clearance floor of the marks from the ink of the
// mark (the ball of f reaches over the dot of the dieresis without the kern).
func TestFBeforeDieresis(t *testing.T) {
	t.Parallel()
	r := ran(t)
	files, _ := texOf(t)
	for i, f := range fonts {
		tf := &tex.Font{Style: f.style()}
		name := tex.DirTFM + "/" + tf.TFM(tex.T1) + ".tfm"
		file, err := tfm.Read(bytes.NewReader(files[name]))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		p, gs := r.fonts[i].params, r.fonts[i].glyphs
		index, _ := tex.T1.Resolve(gs)
		floor := p.Acc.MarkFloor
		n := 0
		for c, ic := range index {
			if ic < 0 || !strings.ContainsRune("ìíîïĩīĭ", tex.T1.Slots[c].Rune) {
				continue
			}
			n++
			shift := gs[index['f']].Advance + file.KernBetween('f', c).Float()*1000
			if d := inkDistance(gs[index['f']].Path, gs[ic].Path, shift); d < floor-1 {
				t.Errorf("%s: f and %s are %.1f apart, want at least %g", name, gs[ic].Name, d, floor)
			}
		}
		if n < 4 {
			t.Errorf("%s: only %d letters i with a mark", name, n)
		}
	}
}
