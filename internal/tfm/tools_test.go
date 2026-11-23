// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// needTool returns the path of the named external tool.  If the tool is
// missing, the test is skipped, unless the environment variable
// BEDERN_REQUIRE_TOOLS is 1, in which case the test fails.
func needTool(t testing.TB, name string) string {
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

// converters runs the installed TFtoPL and PLtoTF.
type converters struct {
	tftopl, pltotf string
}

func needConverters(t testing.TB) *converters {
	t.Helper()
	return &converters{tftopl: needTool(t, "tftopl"), pltotf: needTool(t, "pltotf")}
}

// run runs a command and returns its standard output, its standard error
// output and whether it ended with the status 0.
func run(name string, args ...string) (stdout, stderr []byte, ok bool, err error) {
	cmd := exec.Command(name, args...)
	var o, e bytes.Buffer
	cmd.Stdout = &o
	cmd.Stderr = &e
	err = cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return o.Bytes(), e.Bytes(), false, nil
	}
	return o.Bytes(), e.Bytes(), err == nil, err
}

// toPL runs TFtoPL on the TFM file at path.
func (cv *converters) toPL(path string) (pl, messages []byte, ok bool, err error) {
	return run(cv.tftopl, path)
}

// toTFM runs PLtoTF on a PL text, with dir as the directory for its files.
func (cv *converters) toTFM(dir string, pl []byte) (tfm, messages []byte, ok bool, err error) {
	in := filepath.Join(dir, "in.pl")
	out := filepath.Join(dir, "out.tfm")
	if err := os.WriteFile(in, pl, 0o644); err != nil {
		return nil, nil, false, err
	}
	os.Remove(out)
	_, messages, ok, err = run(cv.pltotf, in, out)
	if err != nil {
		return nil, nil, false, err
	}
	tfm, err = os.ReadFile(out)
	if err != nil {
		return nil, messages, false, nil
	}
	return tfm, messages, ok, nil
}

// TestConverters runs the installed converters on the fonts which the tests
// build: PLtoTF applied to the output of TFtoPL gives the bytes back, and
// the PL text of the package is the output of TFtoPL.
func TestConverters(t *testing.T) {
	if testing.Short() {
		t.Skip("runs external programs")
	}
	cv := needConverters(t)
	dir := t.TempDir()
	for _, tf := range testFonts() {
		file, _, data := build(t, tf.font)
		path := filepath.Join(dir, tf.name+".tfm")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		pl, messages, ok, err := cv.toPL(path)
		if err != nil || !ok || len(messages) > 0 {
			t.Errorf("%s: TFtoPL failed: %v %s", tf.name, err, messages)
			continue
		}
		var ours bytes.Buffer
		if err := file.WritePL(&ours); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(ours.Bytes(), pl) {
			t.Errorf("%s: PL text differs from the output of TFtoPL at byte %d", tf.name, firstDiff(ours.Bytes(), pl))
		}
		back, messages, ok, err := cv.toTFM(dir, pl)
		if err != nil || !ok || len(messages) > 0 {
			t.Errorf("%s: PLtoTF failed: %v %s", tf.name, err, messages)
			continue
		}
		if !bytes.Equal(back, data) {
			t.Errorf("%s: PLtoTF gives other bytes, first at %d", tf.name, firstDiff(back, data))
		}
	}
}

// TestPLtoTF runs the installed PLtoTF on the PL texts of the tests: it
// converts the texts which ReadPL reads, reports a fault for the texts for
// which ReadPL returns an error and prints a message for the texts whose
// descriptions Build refuses.
func TestPLtoTF(t *testing.T) {
	if testing.Short() {
		t.Skip("runs external programs")
	}
	cv := needConverters(t)
	dir := t.TempDir()
	for _, src := range plSources {
		data, _, ok, err := cv.toTFM(dir, []byte(src.text))
		if err != nil || !ok {
			t.Errorf("%s: PLtoTF failed: %v", src.name, err)
		} else if !bytes.Equal(data, golden(t, "pl-"+src.name+".tfm")) {
			t.Errorf("%s: the file of testdata is not the output of PLtoTF", src.name)
		}
	}
	for _, c := range plErrors {
		if _, _, ok, err := cv.toTFM(dir, []byte(c.text)); err != nil || ok {
			t.Errorf("%q: PLtoTF reports no fault (%v)", c.text, err)
		}
	}
	for _, text := range buildErrors {
		if _, messages, ok, err := cv.toTFM(dir, []byte(text)); err != nil || !ok || len(messages) == 0 {
			t.Errorf("%q: PLtoTF: %v, success %v, messages %q", text, err, ok, messages)
		}
	}
	for _, text := range stricter {
		if _, messages, ok, err := cv.toTFM(dir, []byte(text)); err != nil || !ok || len(messages) != 0 {
			t.Errorf("%q: PLtoTF: %v, success %v, messages %q", text, err, ok, messages)
		}
	}
}

// plainName matches the file names which the test passes to TeX.
var plainName = regexp.MustCompile(`^[A-Za-z0-9/._+-]+$`)

// texLoads runs TeX in the directory dir and reports for each of the named
// TFM files whether TeX loads it.
func texLoads(t *testing.T, tex, dir string, names []string) []bool {
	t.Helper()
	loads := make([]bool, len(names))
	for start := 0; start < len(names); start += 200 {
		end := min(start+200, len(names))
		var src strings.Builder
		src.WriteString("\\catcode`\\{=1 \\catcode`\\}=2 \\batchmode\n")
		for i := start; i < end; i++ {
			fmt.Fprintf(&src, "\\message{@%d@}\\font\\f=%s \n", i, names[i])
		}
		fmt.Fprintf(&src, "\\message{@%d@}\\end\n", end)
		if err := os.WriteFile(filepath.Join(dir, "load.tex"), []byte(src.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(tex, "-ini", "-interaction=batchmode", "load.tex")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "max_print_line=100000")
		cmd.Run()
		log, err := os.ReadFile(filepath.Join(dir, "load.log"))
		if err != nil {
			t.Fatalf("TeX wrote no log: %v", err)
		}
		text := string(log)
		for i := start; i < end; i++ {
			a := strings.Index(text, fmt.Sprintf("@%d@", i))
			b := strings.Index(text, fmt.Sprintf("@%d@", i+1))
			if a < 0 || b < a {
				t.Fatalf("the log of TeX has no mark for file %d", i)
			}
			part := text[a:b]
			if strings.Contains(part, "not found") {
				t.Fatalf("TeX does not find %s", names[i])
			}
			loads[i] = !strings.Contains(part, "not loadable")
		}
	}
	return loads
}

// TestTeX checks with the installed TeX that it loads the files which the
// tests build and the odd files, and that it refuses the inputs which Read
// refuses, but for the rules which are stricter than those of TeX.
func TestTeX(t *testing.T) {
	if testing.Short() {
		t.Skip("runs external programs")
	}
	tex := needTool(t, "tex")
	dir := t.TempDir()
	var names, titles []string
	var want []int // 1: TeX loads the file, 0: it does not, -1: either
	add := func(title string, data []byte, loads int) {
		name := fmt.Sprintf("f%d", len(names))
		if err := os.WriteFile(filepath.Join(dir, name+".tfm"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
		titles = append(titles, title)
		want = append(want, loads)
	}
	for _, tf := range testFonts() {
		_, _, data := build(t, tf.font)
		add(tf.name, data, 1)
	}
	for _, of := range oddFiles(t) {
		add(of.name, of.file.bytes(), 1)
	}
	for _, src := range plSources {
		add(src.name, golden(t, "pl-"+src.name+".tfm"), 1)
	}
	for _, c := range invalidFiles(t) {
		if c.stricter {
			add(c.name, c.file.bytes(), -1)
		} else {
			add(c.name, c.file.bytes(), 0)
		}
	}
	for _, c := range notTFMInputs(t) {
		add(c.name, c.data, 0)
	}
	for i, loads := range texLoads(t, tex, dir, names) {
		switch {
		case want[i] == 1 && !loads:
			t.Errorf("%s: TeX does not load the file", titles[i])
		case want[i] == 0 && loads:
			t.Errorf("%s: TeX loads the file", titles[i])
		case want[i] == -1:
			t.Logf("%s: a rule which is stricter than TeX; TeX loads the file: %v", titles[i], loads)
		}
	}
}

// randomFont returns a description which depends on seed only.
func randomFont(seed int64) *Font {
	rng := rand.New(rand.NewSource(seed))
	f := NewFont()
	f.CodingScheme = [...]string{"BEDERN RANDOM", "TEX MATH SYMBOLS", "TEX MATH EXTENSION", ""}[rng.Intn(4)]
	f.Family = "R" + strconv.FormatInt(seed, 10)
	f.Face = uint8(rng.Intn(256))
	f.DesignSize = Unity + Fix(rng.Intn(100<<20))
	units := 1.0
	switch rng.Intn(4) {
	case 0:
		f.DesignUnits = 1000 * Unity
		units = 1000
	case 1:
		f.DesignUnits = FixFromFloat(7.3)
		units = 7.3
	}
	dim := func(lo, hi float64) Fix {
		return FixFromFloat((lo + rng.Float64()*(hi-lo)) * units)
	}
	if rng.Intn(2) == 0 {
		f.HasChecksum, f.Checksum = true, rng.Uint32()
	}
	for i := rng.Intn(4); i > 0; i-- {
		f.Header = append(f.Header, rng.Uint32())
	}
	for i := rng.Intn(25); i > 0; i-- {
		f.Params = append(f.Params, dim(-1.9, 1.9))
	}
	var codes []uint8
	n := 1 + rng.Intn(256)
	for _, c := range rng.Perm(256)[:n] {
		codes = append(codes, uint8(c))
	}
	// the number of different values of each dimension, and its range
	type kind struct {
		n      int
		lo, hi float64
	}
	pick := func(few, many int) kind {
		k := kind{n: 1 + rng.Intn(few), lo: -0.5, hi: 1.5}
		if rng.Intn(3) == 0 {
			k.n = 1 + rng.Intn(many)
		}
		switch rng.Intn(6) {
		case 0:
			k.lo, k.hi = -1.9, -0.01
		case 1:
			k.lo, k.hi = 0.01, 1.9
		}
		return k
	}
	kinds := [4]kind{pick(100, 300), pick(14, 60), pick(14, 60), pick(60, 200)}
	var values [4][]Fix
	for i, k := range kinds {
		for j := 0; j < k.n; j++ {
			values[i] = append(values[i], dim(k.lo, k.hi))
		}
		if rng.Intn(2) == 0 {
			values[i] = append(values[i], 0)
		}
	}
	for _, c := range codes {
		ch := &Char{}
		for i, d := range []*Fix{&ch.Width, &ch.Height, &ch.Depth, &ch.Italic} {
			*d = values[i][rng.Intn(len(values[i]))]
		}
		f.Chars[c] = ch
	}
	any := func() uint8 { return codes[rng.Intn(len(codes))] }
	if rng.Intn(2) == 0 {
		f.HasBoundary = true
		f.BoundaryChar = uint8(rng.Intn(256))
	}
	// lists of larger characters and recipes on some characters
	tagged := make(map[uint8]bool)
	for i := rng.Intn(6); i > 0; i-- {
		c := any()
		if tagged[c] {
			continue
		}
		tagged[c] = true
		if rng.Intn(2) == 0 {
			f.Chars[c].Tag, f.Chars[c].NextLarger = ListTag, any()
		} else {
			f.Chars[c].Tag = ExtTag
			f.Chars[c].Recipe = Recipe{Top: any(), Rep: any()}
			if rng.Intn(2) == 0 {
				f.Chars[c].Recipe.Mid, f.Chars[c].Recipe.Bot = any(), any()
			}
		}
	}
	// programs
	ops := []Op{OpLig, OpLigSlash, OpSlashLig, OpSlashLigSlash, OpLigSlashOver, OpSlashLigOver, OpSlashLigSlashOver, OpSlashLigSlashOver2}
	kerns := []Fix{dim(-0.3, 0.3), dim(-0.3, 0.3), dim(-0.3, 0.3)}
	programs := rng.Intn(60)
	if rng.Intn(4) == 0 {
		programs = 100 + rng.Intn(200)
	}
	for i := 0; i < programs; i++ {
		c := any()
		if tagged[c] {
			continue
		}
		tagged[c] = true
		f.LigTable = append(f.LigTable, Label(c))
		if i == 0 && rng.Intn(2) == 0 {
			f.LigTable = append(f.LigTable, BoundaryLabel())
		}
		steps := 1 + rng.Intn(8)
		for j := 0; j < steps; j++ {
			next := any()
			if f.HasBoundary && rng.Intn(8) == 0 {
				next = f.BoundaryChar
			}
			switch rng.Intn(12) {
			case 0:
				f.LigTable = append(f.LigTable, Lig(ops[rng.Intn(len(ops))], next, any()))
			case 1:
				f.LigTable = append(f.LigTable, Krn(next, dim(-0.5, 0.5)))
			default:
				f.LigTable = append(f.LigTable, Krn(next, kerns[rng.Intn(len(kerns))]))
			}
			if j < steps-1 && rng.Intn(10) == 0 {
				f.LigTable = append(f.LigTable, Skip(uint8(rng.Intn(steps-1-j))))
			}
		}
		if rng.Intn(5) > 0 || i == programs-1 {
			f.LigTable = append(f.LigTable, Stop())
		}
	}
	return f
}

// TestRandom compares Build with the installed PLtoTF on descriptions which
// a random number generator makes: where Build gives a file, PLtoTF gives the
// same bytes, and where Build refuses a description, PLtoTF prints a message.
func TestRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("runs external programs")
	}
	cv := needConverters(t)
	dir := t.TempDir()
	same, refused, reduced := 0, 0, 0
	for seed := int64(1); seed <= 400; seed++ {
		font := randomFont(seed)
		text := description(t, font)
		want, messages, ok, err := cv.toTFM(dir, text)
		if err != nil || !ok {
			t.Errorf("seed %d: PLtoTF failed: %v %s", seed, err, messages)
			continue
		}
		file, rep, err := font.Build()
		if err != nil {
			refused++
			if len(messages) == 0 {
				t.Errorf("seed %d: Build refuses the description (%v), PLtoTF is silent", seed, err)
			}
			continue
		}
		if len(rep.Reductions) > 0 {
			reduced++
		}
		if got := file.bytes(); !bytes.Equal(got, want) {
			t.Errorf("seed %d: other bytes than PLtoTF, first at %d (reductions %+v)", seed, firstDiff(got, want), rep.Reductions)
			continue
		}
		same++
		again, err := ReadPL(bytes.NewReader(text))
		if err != nil || !reflect.DeepEqual(again, font) {
			t.Errorf("seed %d: the PL text reads as another description (%v)", seed, err)
		}
	}
	t.Logf("400 descriptions: %d files equal to the output of PLtoTF (%d with a reduction), %d refused by Build with a message of PLtoTF", same, reduced, refused)
}

// onlyWarnings reports whether the faults which PLtoTF reports are all about
// the indentation, which ReadPL does not check.
func onlyWarnings(messages []byte) bool {
	faults, warnings := 0, 0
	for _, line := range strings.Split(string(messages), "\n") {
		if strings.Contains(line, " (line ") {
			faults++
			if strings.HasPrefix(line, "Warning: ") {
				warnings++
			}
		}
	}
	return faults > 0 && faults == warnings
}

// TestMutations compares ReadPL and Build with the installed PLtoTF on PL
// texts with random changes: ReadPL returns an error where PLtoTF reports a
// fault, Build gives the bytes of PLtoTF, and where Build refuses a
// description, PLtoTF prints a message.
func TestMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("runs external programs")
	}
	cv := needConverters(t)
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(7))
	const alphabet = "()() \n\nCDOHFR0123456789.-+/>ABLIGKRNSTOPabcxyz"
	sources := []string{formsPL, repeatPL}
	for _, tf := range testFonts()[:4] {
		sources = append(sources, string(description(t, tf.font)))
	}
	faulty, same, refused, warned := 0, 0, 0, 0
	for i := 0; i < 1500; i++ {
		src := []byte(sources[rng.Intn(len(sources))])
		for k := 1 + rng.Intn(3); k > 0; k-- {
			pos := rng.Intn(len(src))
			c := alphabet[rng.Intn(len(alphabet))]
			switch rng.Intn(3) {
			case 0:
				src[pos] = c
			case 1:
				src = slices.Delete(src, pos, pos+1)
			default:
				src = slices.Insert(src, pos, c)
			}
		}
		want, messages, ok, err := cv.toTFM(dir, src)
		if err != nil {
			t.Fatal(err)
		}
		font, err := ReadPL(bytes.NewReader(src))
		switch {
		case !ok && onlyWarnings(messages):
			warned++
			continue
		case !ok && err == nil:
			t.Errorf("case %d: PLtoTF reports a fault, ReadPL does not: %s", i, messages)
			continue
		case ok && err != nil:
			t.Errorf("case %d: ReadPL reports a fault, PLtoTF does not: %v", i, err)
			continue
		case !ok:
			faulty++
			continue
		}
		file, _, err := font.Build()
		if err != nil {
			refused++
			if len(messages) == 0 {
				t.Errorf("case %d: Build refuses the description (%v), PLtoTF is silent", i, err)
			}
			continue
		}
		if got := file.bytes(); !bytes.Equal(got, want) {
			t.Errorf("case %d: other bytes than PLtoTF, first at %d", i, firstDiff(got, want))
			continue
		}
		same++
	}
	t.Logf("1500 texts: %d faulty for both, %d with the bytes of PLtoTF, %d refused by Build with a message of PLtoTF, %d with a fault of the indentation only",
		faulty, same, refused, warned)
}

func firstDiff(a, b []byte) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// treeFiles returns the TFM files of the local TeX tree, sorted.
func treeFiles(t *testing.T) []string {
	t.Helper()
	kpsewhich := needTool(t, "kpsewhich")
	out, _, ok, err := run(kpsewhich, "--show-path=tfm")
	if err != nil || !ok {
		t.Fatalf("kpsewhich failed: %v", err)
	}
	seen := make(map[string]bool)
	var files []string
	for _, dir := range filepath.SplitList(strings.TrimSpace(string(out))) {
		dir = strings.TrimLeft(dir, "!")
		deep := strings.HasSuffix(dir, "//")
		dir = filepath.Clean(dir)
		if dir == "." {
			continue
		}
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			continue
		}
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if !deep && path != dir {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(filepath.Ext(path), ".tfm") && !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
			return nil
		})
	}
	sort.Strings(files)
	return files
}

// treeFile is what the test of the TeX tree knows about one file.
type treeFile struct {
	path     string
	data     []byte
	file     *File
	err      error
	finds    []Finding
	rebuilt  []byte // the bytes of the description of the file, built again
	buildErr error
	classes  []string // the ways in which the file differs from rebuilt
	features []string
}

// headerStringClasses names the ways in which a string of the header is not
// in the form which PLtoTF writes.
func headerStringClasses(f *File, pos, max int, classes map[string]bool) {
	n := int(f.headerByte(pos))
	if n >= max {
		classes["header: a string longer than its field"] = true
		return
	}
	for i := 1; i <= n; i++ {
		c := f.headerByte(pos + i)
		switch {
		case c >= 'a' && c <= 'z':
			classes["header: lower-case letters in a string"] = true
		case c == '(' || c == ')' || c < ' ' || c > '~':
			classes["header: a string with a parenthesis or a byte which is no printable ASCII character"] = true
		case c == ' ' && i == 1:
			classes["header: a string which starts with a blank"] = true
		}
	}
	for i := n + 1; i < max; i++ {
		if f.headerByte(pos+i) != 0 {
			classes["header: bytes after the end of a string are not zero"] = true
		}
	}
}

// tableClasses names the ways in which a table of dimensions is not in the
// form which PLtoTF writes.
func tableClasses(name string, t []Fix, used []bool, zero bool, classes map[string]bool) {
	for i := 1; i < len(t); i++ {
		if !used[i] {
			classes[name+": entries which no character uses"] = true
		}
		if i > 1 && t[i] < t[i-1] {
			classes[name+": other order"] = true
		}
		if !zero && t[i] == 0 {
			classes[name+": a second entry for zero"] = true
		}
		for j := 1; j < i; j++ {
			if t[j] == t[i] {
				classes[name+": a value twice"] = true
				break
			}
		}
	}
}

// classify names the ways in which a file differs from the file which the
// description of it gives.  Every class is a property of the file itself.  A
// difference without a class is reported as unexplained.
func classify(f, g *File, rep *Report) []string {
	classes := make(map[string]bool)
	unexplained := func(what string) {
		classes["unexplained: "+what] = true
	}

	// the header
	if !slices.Equal(f.Header, g.Header) {
		n := len(classes)
		if len(f.Header) < 18 {
			classes["header: fewer than 18 words"] = true
		}
		if len(f.Header) >= 12 {
			headerStringClasses(f, 8, 40, classes)
		}
		if len(f.Header) >= 17 {
			headerStringClasses(f, 48, 20, classes)
		}
		if len(f.Header) >= 18 {
			flag := byte(0)
			if rep.SevenBitSafe {
				flag = 128
			}
			if f.headerByte(68) != flag {
				classes["header: the seven-bit-safe flag is not the one which PLtoTF works out"] = true
			}
			if f.headerByte(69) != 0 || f.headerByte(70) != 0 {
				classes["header: the unused bytes of word 17 are not zero"] = true
			}
		}
		if len(classes) == n {
			unexplained("header")
		}
	}

	// the tables of dimensions
	nw, nh, nd, ni := len(f.Width), len(f.Height), len(f.Depth), len(f.Italic)
	usedW, usedH, usedD, usedI := make([]bool, nw), make([]bool, nh), make([]bool, nd), make([]bool, ni)
	first, last := 256, -1
	strayFields, strayTags := false, false
	for i, ci := range f.Chars {
		if ci.WidthIndex == 0 {
			if ci.HeightIndex != 0 || ci.DepthIndex != 0 || ci.ItalicIndex != 0 || (ci.Tag == NoTag && ci.Remainder != 0) {
				strayFields = true
			}
			if ci.Tag == ListTag || ci.Tag == ExtTag {
				strayTags = true
			}
			continue
		}
		first = min(first, f.First+i)
		last = f.First + i
		usedW[ci.WidthIndex] = true
		usedH[ci.HeightIndex] = true
		usedD[ci.DepthIndex] = true
		usedI[ci.ItalicIndex] = true
	}
	for _, tb := range []struct {
		name string
		a, b []Fix
		used []bool
	}{{"widths", f.Width, g.Width, usedW}, {"heights", f.Height, g.Height, usedH},
		{"depths", f.Depth, g.Depth, usedD}, {"italic corrections", f.Italic, g.Italic, usedI}} {
		if slices.Equal(tb.a, tb.b) {
			continue
		}
		n := len(classes)
		tableClasses(tb.name, tb.a, tb.used, tb.name == "widths", classes)
		if len(classes) == n {
			unexplained(tb.name)
		}
	}

	// the ligature and kerning program
	l := f.list()
	ligDiffers := !slices.Equal(f.LigKern, g.LigKern)
	kernDiffers := !slices.Equal(f.Kern, g.Kern)
	if kernDiffers {
		n := len(classes)
		seen := make([]bool, len(f.Kern))
		next := 0
		for i, s := range f.LigKern {
			if l.activity[i] != accessible || s.Skip > 128 || s.Op < 128 {
				continue
			}
			k := int(s.Op-128)<<8 | int(s.Remainder)
			if !seen[k] {
				seen[k] = true
				if k != next {
					classes["kerns: not in the order of their first use"] = true
				}
				next++
			}
		}
		for k, x := range f.Kern {
			if !seen[k] {
				classes["kerns: values which no step of a program uses"] = true
			}
			for j := 0; j < k; j++ {
				if f.Kern[j] == x {
					classes["kerns: a value twice"] = true
					break
				}
			}
		}
		if len(classes) == n {
			unexplained("kerns")
		}
	}
	if ligDiffers {
		n := len(classes)
		if l.unused > 0 {
			classes["program: steps which no program reaches"] = true
		}
		if l.oddOps > 0 {
			classes["program: a ligature operation which is none of the eight"] = true
		}
		// the words which are no steps
		var pass, passG []LigKern
		for i, s := range f.LigKern {
			if l.activity[i] == passThrough {
				pass = append(pass, s)
			} else if l.activity[i] == accessible && s.Skip > 128 {
				classes["program: a program ends at a word which is no step"] = true
			}
		}
		lg := g.list()
		for i, s := range g.LigKern {
			if lg.activity[i] == passThrough {
				passG = append(passG, s)
			}
		}
		if len(pass) != len(passG) {
			classes["program: other number of words for the starts of programs and the boundaries"] = true
		} else if !slices.Equal(pass, passG) && !kernDiffers {
			classes["program: other words for the starts of programs and the boundaries"] = true
		}
		// the steps themselves
		var steps, stepsG []LigKern
		for i, s := range f.LigKern {
			if l.activity[i] == accessible {
				steps = append(steps, s)
			}
		}
		for i, s := range g.LigKern {
			if lg.activity[i] == accessible {
				stepsG = append(stepsG, s)
			}
		}
		if len(steps) == len(stepsG) {
			for i := range steps {
				a, b := steps[i], stepsG[i]
				if a.Skip != b.Skip && l.unused == 0 {
					if a.Skip > 128 || b.Skip > 128 {
						classes["program: a program ends at a word which is no step"] = true
					} else if a.Skip == 0 && b.Skip == 128 {
						classes["program: the last step of the program has no stop"] = true
					} else {
						classes["program: other skips"] = true
					}
				}
				if (a.Op != b.Op || a.Remainder != b.Remainder) && a.Op >= 128 && !kernDiffers {
					unexplained("kern indices")
				}
			}
		}
		if kernDiffers {
			classes["program: kern indices follow the table of kerns"] = true
		}
		if len(classes) == n {
			unexplained("program")
		}
	}

	// the recipes
	if !slices.Equal(f.Exten, g.Exten) {
		n := len(classes)
		users := make([]int, len(f.Exten))
		next := 0
		for _, ci := range f.Chars {
			if ci.WidthIndex == 0 || ci.Tag != ExtTag {
				continue
			}
			r := int(ci.Remainder)
			users[r]++
			if users[r] > 1 {
				classes["recipes: a recipe which several characters share"] = true
			} else if r != next {
				classes["recipes: not in the order of their characters"] = true
			}
			next++
		}
		for _, n := range users {
			if n == 0 {
				classes["recipes: a recipe which no character uses"] = true
			}
		}
		if len(classes) == n {
			unexplained("recipes")
		}
	}

	// the parameters
	if !slices.Equal(f.Param, g.Param) {
		if len(f.Param) > 0 && f.Param[0] < -1024*Unity {
			classes["parameters: a slant below -1024"] = true
		} else {
			unexplained("parameters")
		}
	}
	if len(f.Trailer) > 0 {
		classes["bytes after the stated length"] = true
	}

	// the words of the characters
	if f.First != g.First || f.Last != g.Last {
		if first > last {
			if f.First != 1 || f.Last != 0 {
				classes["characters: a font without characters whose range is not 1 to 0"] = true
			}
		} else if f.First != first || f.Last != last {
			classes["characters: codes without a character at the ends of the range"] = true
		} else {
			unexplained("range of the characters")
		}
	} else if !slices.Equal(f.Chars, g.Chars) {
		n := len(classes)
		if strayFields {
			classes["characters: the word of a code without a character is not zero"] = true
		}
		if strayTags {
			classes["characters: a code without a character has a larger character or a recipe"] = true
		}
		for _, fd := range f.Check() {
			if fd.Kind == LargerMissing {
				classes["characters: a next larger character which does not exist"] = true
			}
		}
		// differences which follow from the tables
		for i, a := range f.Chars {
			b := g.Chars[i]
			if a.WidthIndex == 0 {
				continue
			}
			if a.WidthIndex != b.WidthIndex && !slices.Equal(f.Width, g.Width) ||
				a.HeightIndex != b.HeightIndex && !slices.Equal(f.Height, g.Height) ||
				a.DepthIndex != b.DepthIndex && !slices.Equal(f.Depth, g.Depth) ||
				a.ItalicIndex != b.ItalicIndex && !slices.Equal(f.Italic, g.Italic) {
				classes["characters: indices follow the tables of dimensions"] = true
			}
			if a.Tag == LigTag && b.Tag == LigTag && a.Remainder != b.Remainder && ligDiffers {
				classes["characters: starts follow the program"] = true
			}
			if a.Tag == ExtTag && b.Tag == ExtTag && a.Remainder != b.Remainder && !slices.Equal(f.Exten, g.Exten) {
				classes["characters: indices follow the recipes"] = true
			}
		}
		if len(classes) == n {
			unexplained("words of the characters")
		}
	}

	var list []string
	for c := range classes {
		list = append(list, c)
	}
	sort.Strings(list)
	return list
}

// features names the properties of a file by which the test selects its
// sample.
func features(f *File, finds []Finding) []string {
	var fs []string
	v := f.view()
	if v.fontType == mathSy && len(f.Param) == 22 {
		fs = append(fs, "22 parameters")
	}
	if v.fontType == mathEx && len(f.Param) == 13 {
		fs = append(fs, "13 parameters")
	}
	if len(f.Exten) > 0 {
		fs = append(fs, "recipes")
	}
	for _, ci := range f.Chars {
		if ci.Tag == ListTag {
			fs = append(fs, "lists of larger characters")
			break
		}
	}
	if v.list.bchar >= 0 {
		fs = append(fs, "boundary character")
	}
	if v.list.bstart >= 0 {
		fs = append(fs, "program of the left boundary")
	}
	if len(f.LigKern) > 255 {
		fs = append(fs, "program of more than 255 words")
		for _, ci := range f.Chars {
			if ci.Tag == LigTag && f.LigKern[ci.Remainder].Skip > 128 {
				fs = append(fs, "indirect starts")
				break
			}
		}
	}
	if len(f.Header) != 18 {
		fs = append(fs, "header of other than 18 words")
	}
	if len(f.Header) > 18 {
		fs = append(fs, "header of more than 18 words")
	}
	if v.list.unused > 0 {
		fs = append(fs, "unused steps")
	}
	for _, fd := range finds {
		if fd.Kind != UnusedSteps {
			fs = append(fs, "findings of TFtoPL")
			break
		}
	}
	if len(f.Trailer) > 0 {
		fs = append(fs, "trailing bytes")
	}
	return fs
}

var digits = regexp.MustCompile(`[0-9]+`)

// count counts strings.
type count map[string]int

func (c count) log(t *testing.T, title string) {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	t.Logf("%s:", title)
	for _, k := range keys {
		t.Logf("  %6d  %s", c[k], k)
	}
}

// reported reports whether TFtoPL prints a message for one of the findings.
func reported(finds []Finding) bool {
	for _, fd := range finds {
		if fd.Kind != UnusedSteps {
			return true
		}
	}
	return false
}

// capacity names the limit of PLtoTF which the description of a file is
// beyond, if there is one.
func capacity(f *File) string {
	d := f.Font()
	steps := 0
	kerns := make(map[Fix]bool)
	for _, in := range d.LigTable {
		if in.Kind == KindStep {
			steps++
			if in.Step.Op == OpKern {
				kerns[in.Step.Kern] = true
			}
		}
	}
	switch {
	case len(f.Header) > 250:
		return "more than 250 header words"
	case len(f.Param) > 254:
		return "more than 254 parameters"
	case steps > 32500:
		return "more than 32500 steps"
	case len(kerns) > 5000:
		return "more than 5000 kerns"
	}
	return ""
}

// TestTree reads the TFM files of the local TeX tree and compares the
// package with the installed converters on a sample of them: every n-th
// file, some files with each feature and of each class and some of the
// rejected files.  With the environment variable BEDERN_TFM_ALL set to 1 the
// sample is every file.  The test runs only with BEDERN_TFM_TREE=1 or
// BEDERN_TFM_ALL=1, since it reads the 80 000 files of the tree.
func TestTree(t *testing.T) {
	if os.Getenv("BEDERN_TFM_TREE") != "1" && os.Getenv("BEDERN_TFM_ALL") != "1" {
		t.Skip("walks the TeX tree; set BEDERN_TFM_TREE=1 to run it")
	}
	cv := needConverters(t)
	paths := treeFiles(t)
	if len(paths) == 0 {
		t.Skip("the TeX tree has no TFM files")
	}
	files := make([]*treeFile, len(paths))
	work := func(n int, fn func(i int)) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		next := 0
		for w := 0; w < runtime.GOMAXPROCS(0); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					mu.Lock()
					i := next
					next++
					mu.Unlock()
					if i >= n {
						return
					}
					fn(i)
				}
			}()
		}
		wg.Wait()
	}

	// every file: read, write back, describe and build again
	work(len(paths), func(i int) {
		tf := &treeFile{path: paths[i]}
		files[i] = tf
		tf.data, tf.err = os.ReadFile(tf.path)
		if tf.err != nil {
			return
		}
		tf.file, tf.err = Read(bytes.NewReader(tf.data))
		if tf.err != nil {
			return
		}
		if err := checkFile(tf.data, tf.file); err != nil {
			tf.classes = []string{"unexplained: " + err.Error()}
			return
		}
		var buf bytes.Buffer
		tf.finds = tf.file.Check()
		tf.features = features(tf.file, tf.finds)
		g, rep, err := tf.file.Font().Build()
		if err != nil {
			tf.buildErr = err
			return
		}
		buf.Reset()
		if err := g.Write(&buf); err != nil {
			tf.buildErr = err
			return
		}
		tf.rebuilt = buf.Bytes()
		if bytes.Equal(tf.rebuilt, tf.data) {
			return
		}
		tf.classes = classify(tf.file, g, rep)
		if len(tf.classes) == 0 {
			tf.classes = []string{"unexplained: no table differs"}
		}
	})

	var accepted, rejected []*treeFile
	reasons, classes, combos, featureCount, buildErrs, findings := count{}, count{}, count{}, count{}, count{}, count{}
	same := 0
	byFeature := make(map[string][]*treeFile)
	for _, tf := range files {
		if tf.file == nil {
			rejected = append(rejected, tf)
			var notTFM *NotTFMError
			var invalid *InvalidFontError
			switch {
			case errors.As(tf.err, &notTFM):
				first := "other"
				if len(tf.data) >= 2 {
					if v := int(tf.data[0])<<8 | int(tf.data[1]); v == 9 || v == 11 {
						first = "first length 9 or 11"
					} else if v == 0 {
						first = "first length 0"
					}
				}
				reasons["no TFM file of TeX82, "+first+": "+digits.ReplaceAllString(notTFM.Reason, "N")]++
			case errors.As(tf.err, &invalid):
				reasons["invalid tables: "+digits.ReplaceAllString(invalid.Reason, "N")]++
			default:
				reasons["cannot be read: "+digits.ReplaceAllString(tf.err.Error(), "N")]++
			}
			continue
		}
		accepted = append(accepted, tf)
		kinds := make(map[FindingKind]bool)
		for _, fd := range tf.finds {
			if !kinds[fd.Kind] {
				kinds[fd.Kind] = true
				findings[fd.Text]++
			}
		}
		for _, ft := range tf.features {
			featureCount[ft]++
			byFeature[ft] = append(byFeature[ft], tf)
		}
		switch {
		case tf.buildErr != nil:
			buildErrs[digits.ReplaceAllString(tf.buildErr.Error(), "N")]++
			byFeature["description which cannot be built"] = append(byFeature["description which cannot be built"], tf)
		case len(tf.classes) == 0:
			same++
		default:
			combos[strings.Join(tf.classes, " + ")]++
			for _, c := range tf.classes {
				classes[c]++
				byFeature[c] = append(byFeature[c], tf)
				if strings.HasPrefix(c, "unexplained") {
					t.Errorf("%s: %s", tf.path, c)
				}
			}
		}
	}
	t.Logf("files found: %d, accepted: %d, rejected: %d", len(files), len(accepted), len(rejected))
	reasons.log(t, "rejected files by reason")
	t.Logf("accepted files which write back byte for byte: %d of %d", len(accepted), len(accepted))
	t.Logf("accepted files whose description builds the file itself: %d", same)
	classes.log(t, "accepted files whose description builds another file, by class (a file can be in several)")
	t.Logf("combinations of classes: %d", len(combos))
	buildErrs.log(t, "accepted files whose description cannot be built")
	findings.log(t, "accepted files by finding of File.Check")
	featureCount.log(t, "accepted files by feature")

	// the sample
	inSample := make(map[*treeFile]bool)
	var sample []*treeFile
	add := func(tf *treeFile) {
		if !inSample[tf] {
			inSample[tf] = true
			sample = append(sample, tf)
		}
	}
	stride := max(1, len(accepted)/250)
	if os.Getenv("BEDERN_TFM_ALL") == "1" {
		stride = 1
	}
	for i := 0; i < len(accepted); i += stride {
		add(accepted[i])
	}
	var keys []string
	for k := range byFeature {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sampleFeatures := count{}
	for _, k := range keys {
		list := byFeature[k]
		step := max(1, len(list)/8)
		for i := 0; i < len(list); i += step {
			add(list[i])
		}
	}
	for _, tf := range sample {
		for _, ft := range tf.features {
			sampleFeatures[ft]++
		}
	}
	rstride := max(1, len(rejected)/60)
	if os.Getenv("BEDERN_TFM_ALL") == "1" {
		rstride = 1
	}
	var rsample []*treeFile
	for i, tf := range rejected {
		var invalid *InvalidFontError
		if i%rstride == 0 || errors.As(tf.err, &invalid) {
			rsample = append(rsample, tf)
		}
	}
	sampleFeatures.log(t, fmt.Sprintf("sample of %d accepted files by feature", len(sample)))

	// the converters on the sample
	dir := t.TempDir()
	var mu sync.Mutex
	plSame, tfmSame, normal, normalSame := 0, 0, 0, 0
	verdicts := count{}
	fail := func(tf *treeFile, format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		t.Errorf("%s: %s", tf.path, fmt.Sprintf(format, args...))
	}
	note := func(s string) {
		mu.Lock()
		defer mu.Unlock()
		verdicts[s]++
	}
	work(len(sample), func(i int) {
		tf := sample[i]
		sub := filepath.Join(dir, fmt.Sprint(i))
		if err := os.Mkdir(sub, 0o755); err != nil {
			fail(tf, "%v", err)
			return
		}
		defer os.RemoveAll(sub)
		pl, messages, ok, err := cv.toPL(tf.path)
		if err != nil {
			fail(tf, "TFtoPL: %v", err)
			return
		}
		var ours bytes.Buffer
		if err := tf.file.WritePL(&ours); err != nil {
			fail(tf, "WritePL: %v", err)
			return
		}
		if !bytes.Equal(ours.Bytes(), pl) {
			fail(tf, "PL text differs from the output of TFtoPL at byte %d", firstDiff(ours.Bytes(), pl))
		} else {
			mu.Lock()
			plSame++
			mu.Unlock()
		}
		if rep := reported(tf.finds); rep != (len(messages) > 0) {
			fail(tf, "findings %v, but TFtoPL says %q", tf.finds, messages)
		}
		loop := false
		for _, fd := range tf.finds {
			if fd.Kind == LigatureLoop || fd.Kind == TooManyPairs {
				loop = true
			}
		}
		if loop == ok {
			fail(tf, "findings %v, but TFtoPL ends with success %v", tf.finds, ok)
		}
		switch {
		case !ok:
			note("accepted by Read, TFtoPL gives up (infinite loop of ligatures)")
			return
		case len(messages) > 0:
			note("accepted by Read, TFtoPL prints a message and converts")
		default:
			note("accepted by Read, TFtoPL silent")
		}
		back, messages, ok, err := cv.toTFM(sub, pl)
		if err != nil {
			fail(tf, "PLtoTF: %v", err)
			return
		}
		switch {
		case tf.buildErr != nil:
			if len(messages) == 0 {
				fail(tf, "the description cannot be built (%v), but PLtoTF is silent", tf.buildErr)
			} else {
				note("description cannot be built, PLtoTF prints a message")
			}
			return
		case !ok:
			if c := capacity(tf.file); c != "" {
				note("PLtoTF fails: " + c)
			} else {
				fail(tf, "PLtoTF fails: %s", messages)
			}
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if bytes.Equal(back, tf.rebuilt) {
			tfmSame++
		} else {
			t.Errorf("%s: PLtoTF gives other bytes than Build, first at %d", tf.path, firstDiff(back, tf.rebuilt))
		}
		if bytes.Equal(back, tf.data) {
			normal++
			if bytes.Equal(tf.rebuilt, tf.data) {
				normalSame++
			}
		}
	})
	t.Logf("sample: %d accepted files", len(sample))
	t.Logf("  PL text equal to the output of TFtoPL: %d of %d", plSame, len(sample))
	t.Logf("  bytes of Build equal to the output of PLtoTF: %d", tfmSame)
	t.Logf("  files which PLtoTF gives back unchanged: %d, of which Build gives back unchanged: %d", normal, normalSame)

	// the verdict of TFtoPL on rejected files
	work(len(rsample), func(i int) {
		tf := rsample[i]
		_, messages, ok, err := cv.toPL(tf.path)
		if err != nil {
			fail(tf, "TFtoPL: %v", err)
			return
		}
		switch {
		case !ok:
			note("rejected by Read, TFtoPL gives up")
		case len(messages) > 0:
			note("rejected by Read, TFtoPL prints a message and converts")
		default:
			note("rejected by Read, TFtoPL silent")
			fail(tf, "rejected (%v), but TFtoPL is silent", tf.err)
		}
	})
	t.Logf("sample: %d rejected files", len(rsample))

	// the verdict of TeX on the sampled files
	if tex, err := exec.LookPath("tex"); err != nil {
		t.Logf("TeX is not installed")
	} else {
		var names []string
		var from []*treeFile
		for _, tf := range append(append([]*treeFile{}, sample...), rsample...) {
			if plainName.MatchString(tf.path) {
				names = append(names, tf.path)
				from = append(from, tf)
			}
		}
		for i, loads := range texLoads(t, tex, dir, names) {
			tf := from[i]
			switch {
			case tf.file != nil && loads:
				note("accepted by Read, loaded by TeX")
			case tf.file != nil:
				note("accepted by Read, refused by TeX")
				fail(tf, "accepted, but TeX does not load the file")
			case loads:
				note("rejected by Read, loaded by TeX")
				fail(tf, "rejected (%v), but TeX loads the file", tf.err)
			default:
				note("rejected by Read, refused by TeX")
			}
		}
	}
	verdicts.log(t, "verdicts on the sampled files")
}
