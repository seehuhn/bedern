// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"bytes"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/tfm"
)

// testNow is the time of the test files.
var testNow = time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)

// box returns a glyph whose outline is the rectangle from (0, bottom) to
// (right, top) and whose advance is 500.
func box(name string, r rune, bottom, top, right float64) glyphs.Glyph {
	var p outline.Path
	p.MoveTo(0, bottom)
	p.LineTo(right, bottom)
	p.LineTo(right, top)
	p.LineTo(0, top)
	return glyphs.Glyph{Name: name, Rune: r, Advance: 500, Path: p}
}

// allGlyphs returns a glyph for every position of the encodings, each a
// rectangle, with the ligatures of f.
func allGlyphs() []glyphs.Glyph {
	gs := []glyphs.Glyph{{Name: ".notdef", Advance: 500}}
	seen := make(map[Slot]bool)
	for _, e := range Encodings {
		for _, s := range e.Slots {
			if !s.used() || seen[s] {
				continue
			}
			seen[s] = true
			name := s.Name
			if name == "" {
				name = fmt.Sprintf("u%04X", s.Rune)
			}
			g := box(name, s.Rune, 0, 400, 450)
			switch s.Rune {
			case 0xFB00:
				g.Parts = []string{"u0066", "u0066"}
			case 0xFB01:
				g.Parts = []string{"u0066", "u0069"}
			case 0xFB02:
				g.Parts = []string{"u0066", "u006C"}
			case 0xFB03:
				g.Parts = []string{"u0066", "u0066", "u0069"}
			case 0xFB04:
				g.Parts = []string{"u0066", "u0066", "u006C"}
			}
			gs = append(gs, g)
		}
	}
	return gs
}

// without returns the glyphs without those of the position c of e.
func without(gs []glyphs.Glyph, e *Encoding, c int) []glyphs.Glyph {
	index, _ := e.Resolve(gs)
	return slices.Delete(slices.Clone(gs), index[c], index[c]+1)
}

// TestSlots checks the tables of the encodings: T1 uses every position, the
// letters and figures of ASCII stand at their codes, and no glyph stands at
// two positions except the hyphen of T1, the double hyphen of TS1 and the
// paragraph sign of TS1.
func TestSlots(t *testing.T) {
	for c, s := range T1.Slots {
		if !s.used() {
			t.Errorf("T1 does not use the position %#02x", c)
		}
	}
	for c := 0; c < 128; c++ {
		r := rune(c)
		if (r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') && T1.Slots[c].Rune != r {
			t.Errorf("T1 has %U at the position of %q", T1.Slots[c].Rune, r)
		}
	}
	twice := map[rune]bool{'-': true, 0x2E40: true, 0x00B6: true}
	for _, e := range Encodings {
		runes := make(map[rune]int)
		names := make(map[string]int)
		for c, s := range e.Slots {
			if d, ok := runes[s.Rune]; ok && s.Rune != 0 && !twice[s.Rune] {
				t.Errorf("%s has %U at the positions %#02x and %#02x", e.Name, s.Rune, d, c)
			}
			if d, ok := names[s.Name]; ok && s.Name != "" {
				t.Errorf("%s has %s at the positions %#02x and %#02x", e.Name, s.Name, d, c)
			}
			runes[s.Rune], names[s.Name] = c, c
		}
	}
}

// TestResolve checks that a position takes the glyph with its name, then the
// glyph with its code point, which stands in for a form the font does not
// have, and is listed as missing otherwise.
func TestResolve(t *testing.T) {
	gs := []glyphs.Glyph{
		{Name: ".notdef"},
		{Name: "a", Rune: 'a'},
		{Name: "S_S"},
		{Name: "other", Rune: 0x200C},
	}
	index, missing := T1.Resolve(gs)
	if index['a'] != 1 || index[0xDF] != 2 || index[0x17] != 3 {
		t.Errorf("a, the two letters S and the word mark resolve to %d, %d and %d", index['a'], index[0xDF], index[0x17])
	}
	if index['b'] != -1 || !slices.Contains(missing, 'b') || slices.Contains(missing, 'a') {
		t.Errorf("b resolves to %d, missing %v", index['b'], missing)
	}
	if len(missing) != 256-3 {
		t.Errorf("%d positions are missing, want %d", len(missing), 256-3)
	}
	gs = []glyphs.Glyph{
		{Name: ".notdef"},
		{Name: "quotesingle", Rune: '\''},
		{Name: "paragraph", Rune: 0x00B6},
		{Name: "paragraph.alt"},
	}
	index, _ = TS1.Resolve(gs)
	if index[0x27] != 1 || index[0xB6] != 2 || index[0x99] != 3 {
		t.Errorf("the single quote and the two paragraph signs resolve to %d, %d and %d", index[0x27], index[0xB6], index[0x99])
	}
	if got := TS1.StandIns(gs); !slices.Equal(got, []int{0x27}) {
		t.Errorf("stand-ins %v, want the position of the single quote", got)
	}
}

// TestEncFile checks that an encoding file names 256 glyphs, the glyph of
// each position or .notdef.
func TestEncFile(t *testing.T) {
	gs := without(allGlyphs(), T1, 'q')
	data := string(EncFile(T1, gs))
	body := regexp.MustCompile(`(?s)/BedernT1Encoding \[(.*)\] def`).FindStringSubmatch(data)
	if body == nil {
		t.Fatalf("the file has no vector:\n%s", data)
	}
	names := strings.Fields(body[1])
	if len(names) != 256 {
		t.Fatalf("%d names, want 256", len(names))
	}
	if names['q'] != "/.notdef" || names['a'] != "/u0061" || names[0xDF] != "/S_S" {
		t.Errorf("q, a and the two letters S are %s, %s and %s", names['q'], names['a'], names[0xDF])
	}
}

// TestTS1Subset checks the number of the subset of TS1 for fonts which lack
// one glyph.
func TestTS1Subset(t *testing.T) {
	gs := allGlyphs()
	for _, c := range []struct{ slot, want int }{
		{-1, 0},
		{0x97, 0}, // the trade mark sign: used in every font
		{0x4F, 1}, // the big circle
		{0x62, 2}, // born
		{0x3C, 3}, // the left angle bracket
		{0x18, 4}, // the left arrow
		{0x9B, 5}, // the numero sign
		{0x8C, 6}, // the florin
		{0x57, 7}, // the ohm sign
		{0x20, 8}, // the blank
		{0xB9, 9}, // the superior one
	} {
		g := gs
		if c.slot >= 0 {
			g = without(gs, TS1, c.slot)
		}
		if got := TS1Subset(g); got != c.want {
			t.Errorf("without the position %#02x: subset %d, want %d", c.slot, got, c.want)
		}
	}
}

// TestNominal checks that the overshoot of a round top or bottom is left out
// of the height and the depth, and that other extents stay.
func TestNominal(t *testing.T) {
	p := design.Regular()
	for _, c := range []struct{ top, bottom, height, depth float64 }{
		{p.XHeight, 0, p.XHeight, 0},
		{p.XHeight + p.OvershootLower, -p.OvershootLower, p.XHeight, 0},
		{p.CapHeight + p.OvershootUpper, -p.OvershootUpper, p.CapHeight, 0},
		{p.Ascender, -p.Descender, p.Ascender, p.Descender},
		{p.XHeight, -p.Descender - p.OvershootLower, p.XHeight, p.Descender},
		{p.Figures.Height + p.Figures.OverTop, 0, p.Figures.Height, 0},
		{p.CapHeight + 150, -100, p.CapHeight + 150, 100},
		{-50, -120, 0, 120},
	} {
		h, d := nominal(p, c.top, c.bottom)
		if h != c.height || d != c.depth {
			t.Errorf("outline from %g to %g: height %g and depth %g, want %g and %g", c.bottom, c.top, h, d, c.height, c.depth)
		}
	}
}

// TestMetrics checks the metrics of a font of rectangles: the widths, the
// italic corrections, the ligatures of T1 and the kerns.
func TestMetrics(t *testing.T) {
	p := design.Regular()
	gs := allGlyphs()
	index, _ := T1.Resolve(gs)
	gs[index['V']] = box("u0056", 'V', 0, 400, 530)
	gs[index['A']].Advance = 700
	kern := func(a, b int) int {
		switch {
		case a == index['A'] && b == index['V']:
			return -80
		case a == index['f'] && b == index['i']:
			return 5
		case a == index['f'] && b == index['a']:
			return 7
		}
		return 0
	}
	m := NewMetrics(T1, p, gs, kern)
	if w := m.Chars['A'].Width; w != 700 {
		t.Errorf("width of A %g, want 700", w)
	}
	if ic := m.Chars['V'].Italic; ic != 30 {
		t.Errorf("italic correction of V %g, want 30", ic)
	}
	if ic := m.Chars['A'].Italic; ic != 0 {
		t.Errorf("italic correction of A %g, want 0", ic)
	}
	lig := func(a, b int) int {
		for _, l := range m.Chars[a].Ligs {
			if l.Next == b {
				return l.Result
			}
		}
		return -1
	}
	for _, c := range [][3]int{
		{'f', 'f', 0x1B}, {'f', 'i', 0x1C}, {'f', 'l', 0x1D}, {0x1B, 'i', 0x1E}, {0x1B, 'l', 0x1F},
		{'`', '`', 0x10}, {'\'', '\'', 0x11}, {',', ',', 0x12}, {'<', '<', 0x13}, {'>', '>', 0x14},
		{'-', '-', 0x15}, {0x15, '-', 0x16}, {'-', 0x7F, 0x7F}, {'!', '`', 0xBD}, {'?', '`', 0xBE},
		{'f', 'a', -1}, {0x1C, 'i', -1},
	} {
		if got := lig(c[0], c[1]); got != c[2] {
			t.Errorf("ligature of %#02x and %#02x: %#02x, want %#02x", c[0], c[1], got, c[2])
		}
	}
	want := map[[2]int]float64{{'A', 'V'}: -80, {'f', 'a'}: 7}
	got := make(map[[2]int]float64)
	for a, ch := range m.Chars {
		for _, k := range ch.Kerns {
			got[[2]int{a, k.Next}] = k.Amount
		}
	}
	if len(got) != len(want) || got[[2]int{'A', 'V'}] != -80 || got[[2]int{'f', 'a'}] != 7 {
		t.Errorf("kerns %v, want %v", got, want)
	}

	// without the ligature of two letters f there is none of three
	m = NewMetrics(T1, p, without(gs, T1, 0x1B), kern)
	for _, l := range m.Chars['f'].Ligs {
		if l.Result == 0x1E || l.Result == 0x1F || l.Result == 0x1B {
			t.Errorf("ligature of f and %#02x gives %#02x without the ligature of two letters f", l.Next, l.Result)
		}
	}

	m = NewMetrics(TS1, p, gs, kern)
	for c, ch := range m.Chars {
		if ch != nil && len(ch.Ligs) > 0 {
			t.Errorf("the position %#02x of TS1 has ligatures", c)
		}
	}
}

// TestKernBeforeLigatures checks the kern which a letter takes before a
// character which begins a ligature (see firstKern).
func TestKernBeforeLigatures(t *testing.T) {
	gs := allGlyphs()
	index, _ := T1.Resolve(gs)
	kerns := map[[2]int]int{
		{'A', '-'}: -80, {'A', 0x15}: 0, {'A', 0x16}: -80, // the compromise is the smallest
		{'B', '-'}: -30, {'B', 0x15}: -28, {'B', 0x16}: -29, // within the tolerance
		{'T', '<'}: 0, {'T', 0x13}: -60, // the ligature
		{'f', '?'}: 13, {'f', 0xBE}: 0, // the character
		{'V', '-'}: 10, {'V', 0x15}: -40, {'V', 0x16}: 30, // of mixed signs: that of least size
	}
	kern := func(a, b int) int {
		for k, v := range kerns {
			if index[k[0]] == a && index[k[1]] == b {
				return v
			}
		}
		return 0
	}
	m := NewMetrics(T1, design.Regular(), gs, kern)
	for _, c := range []struct {
		a, b int
		want float64
	}{{'A', '-', 0}, {'B', '-', -30}, {'T', '<', -60}, {'f', '?', 13}, {'V', '-', 10}, {'A', 0x15, 0}} {
		got := 0.0
		for _, k := range m.Chars[c.a].Kerns {
			if k.Next == c.b {
				got = k.Amount
			}
		}
		if got != c.want {
			t.Errorf("kern of %#02x and %#02x: %g, want %g", c.a, c.b, got, c.want)
		}
	}
}

// bestWidth returns the least width w for which the sorted distinct values
// can be covered by at most n intervals of length w, found by dynamic
// programming over the partitions into runs of neighbours.
func bestWidth(vs []float64, n int) float64 {
	m := len(vs)
	// cost[k][i] is the least largest span of a partition of vs[:i] into k runs
	cost := make([][]float64, n+1)
	for k := range cost {
		cost[k] = make([]float64, m+1)
		for i := range cost[k] {
			cost[k][i] = math.Inf(1)
		}
	}
	cost[0][0] = 0
	for k := 1; k <= n; k++ {
		for i := 1; i <= m; i++ {
			for j := 0; j < i; j++ {
				cost[k][i] = math.Min(cost[k][i], math.Max(cost[k-1][j], vs[i-1]-vs[j]))
			}
		}
	}
	best := math.Inf(1)
	for k := 1; k <= n; k++ {
		best = math.Min(best, cost[k][m])
	}
	return best
}

// TestReduce checks the reduction of a table of dimensions on random values:
// no more values other than zero than allowed, zero kept and no other value
// made zero, the order of the values kept, the fixed values unchanged, and,
// without fixed values, the largest change no more than half of the least
// width at which the values can be covered by that many intervals, which is
// what PLtoTF achieves.
func TestReduce(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for trial := 0; trial < 300; trial++ {
		n := 2 + rng.IntN(14) // at least one value for each fixed one
		keep := []float64{448, 683}
		values := []float64{0, 448, 683}
		for i := rng.IntN(60); i > 0; i-- {
			values = append(values, float64(rng.IntN(900)))
		}
		to, change := reduce(values, keep, n)
		out := make(map[float64]bool)
		var distinct []float64
		for _, v := range values {
			w, ok := to[v]
			if !ok {
				t.Fatalf("trial %d: no value for %g", trial, v)
			}
			if (w == 0) != (v == 0) {
				t.Errorf("trial %d: %g becomes %g", trial, v, w)
			}
			if w != 0 {
				out[w] = true
				distinct = append(distinct, v)
			}
			if math.Abs(w-v) > change {
				t.Errorf("trial %d: %g becomes %g, a change above %g", trial, v, w, change)
			}
		}
		if len(out) > n {
			t.Errorf("trial %d: %d values, want at most %d", trial, len(out), n)
		}
		slices.Sort(distinct)
		distinct = slices.Compact(distinct)
		best := bestWidth(distinct, n) / 2
		if change < best-1e-9 {
			t.Errorf("trial %d: the largest change is %g, below the least possible %g", trial, change, best)
		}
		for _, k := range keep {
			if to[k] != k {
				t.Errorf("trial %d: the fixed value %g becomes %g", trial, k, to[k])
			}
		}
		if _, change := reduce(values, nil, n); math.Abs(change-best) > 1e-9 {
			t.Errorf("trial %d: without fixed values the largest change is %g, the least is %g", trial, change, best)
		}
		sorted := slices.Clone(values)
		slices.Sort(sorted)
		for i := 1; i < len(sorted); i++ {
			if to[sorted[i-1]] > to[sorted[i]] {
				t.Errorf("trial %d: %g and %g change their order", trial, sorted[i-1], sorted[i])
			}
		}
		if len(distinct) <= n && change != 0 {
			t.Errorf("trial %d: values which fit change by %g", trial, change)
		}
	}
	// values which fit stay as they are
	to, change := reduce([]float64{0, 1, 2, 3}, nil, 3)
	if change != 0 || to[1] != 1 || to[2] != 2 || to[3] != 3 {
		t.Errorf("three values reduced to three: %v, change %g", to, change)
	}
	// a small height does not become zero, and the line of the design stays fixed
	to, change = reduce([]float64{0, 3, 440, 448, 456, 700}, []float64{448}, 3)
	if to[3] != 3 || to[448] != 448 || to[440] != 448 || to[456] != 448 || to[700] != 700 || change != 8 {
		t.Errorf("small values: %v, change %g", to, change)
	}
}

// TestReduceLikePLtoTF compares the reduction of a table with that of the
// writer of metric files, which applies the method of PLtoTF to a table which
// is too large: the largest change is the same.
func TestReduceLikePLtoTF(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	for trial := 0; trial < 50; trial++ {
		var values []float64
		f := tfm.NewFont()
		f.DesignSize = DesignSize * tfm.Unity
		for c := 0; c < 100; c++ {
			v := float64(1 + rng.IntN(800))
			values = append(values, v)
			f.Chars[c] = &tfm.Char{Width: fix(500), Height: fix(v)}
		}
		var buf bytes.Buffer
		report, err := f.Write(&buf)
		if err != nil {
			t.Fatal(err)
		}
		_, change := reduce(values, nil, 15)
		got := 0.0
		for _, r := range report.Reductions {
			if r.Table == "height" {
				got = r.MaxChange.Float() * 1000
			}
		}
		if got == 0 || math.Abs(change-got) > 0.002 {
			t.Errorf("trial %d: change %g, PLtoTF %g", trial, change, got)
		}
	}
}

// TestFiles checks the files for TeX of a font of rectangles: one metric
// file for each encoding, the lines of the map file, and the mappings of the
// glyphs to Unicode.
func TestFiles(t *testing.T) {
	p := design.Regular()
	gs := allGlyphs()
	fonts := []*Font{
		{Style: design.RegularStyle(), Series: "m", Glyphs: gs},
		{Style: design.BoldStyle(), Series: "b", Glyphs: gs},
	}
	none := func(a, b int) int { return 0 }
	files, reports, err := Files(fonts, []design.Params{p, p}, []KernFunc{none, none}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"fonts/tfm/public/bedern/bedern-regular-t1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-ts1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-t1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-ts1.tfm",
		"fonts/tfm/public/bedern/bedern-regular-sl-t1.tfm",
		"fonts/tfm/public/bedern/bedern-bold-sl-ts1.tfm",
		"fonts/enc/dvips/bedern/bedern-t1.enc",
		"fonts/enc/dvips/bedern/bedern-ts1.enc",
		"fonts/map/dvips/bedern/bedern.map",
		"tex/latex/bedern/bedern.sty",
		"tex/latex/bedern/t1bedern.fd",
		"tex/latex/bedern/ts1bedern.fd",
		"tex/generic/bedern/bedern-glyphtounicode.tex",
		"tex/plain/bedern/bedern.tex",
	} {
		if len(files[name]) == 0 {
			t.Errorf("the file %s is missing or empty", name)
		}
	}
	if len(files) != 16 {
		t.Errorf("%d files, want 16", len(files))
	}
	if len(reports) != 8 {
		t.Fatalf("%d reports, want 8", len(reports))
	}
	for _, r := range reports {
		if len(r.Missing) != 0 {
			t.Errorf("%s: missing positions %v", r.TFM, r.Missing)
		}
	}
	if reports[0].Chars != 256 || reports[0].Ligs != 15 {
		t.Errorf("T1: %d characters and %d ligatures, want 256 and 15", reports[0].Chars, reports[0].Ligs)
	}
	mapFile := string(files["fonts/map/dvips/bedern/bedern.map"])
	line := `bedern-bold-ts1 Bedern-Bold "BedernTS1Encoding ReEncodeFont" <bedern-ts1.enc <Bedern-Bold.pfb`
	if !strings.Contains(mapFile, line+"\n") {
		t.Errorf("the map file lacks the line %q:\n%s", line, mapFile)
	}
	uni := string(files["tex/generic/bedern/bedern-glyphtounicode.tex"])
	for _, want := range []string{
		`\pdfglyphtounicode{tfm:bedern-regular-t1/uFB03}{0066 0066 0069}`,
		`\pdfglyphtounicode{tfm:bedern-regular-t1/u00E9}{00E9}`,
		`\pdfglyphtounicode{tfm:bedern-bold-ts1/u1F12F}{1F12F}`,
	} {
		if !strings.Contains(uni, want+"\n") {
			t.Errorf("the mappings to Unicode lack the line %s", want)
		}
	}
	again, _, err := Files(fonts, []design.Params{p, p}, []KernFunc{none, none}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range files {
		if !bytes.Equal(data, again[name]) {
			t.Errorf("two renderings of %s differ", name)
		}
	}
}

// TestPackageCommands checks that the package declares the command of a glyph
// in the Private Use Area which the font has, at its code point and at its
// position, and none for a glyph which the font lacks.
func TestPackageCommands(t *testing.T) {
	gs := []glyphs.Glyph{
		{Name: ".notdef"},
		box("perthousandzero", 0xEB4D, 0, 400, 450),
	}
	pkg := string(Package([]*Font{{Style: design.RegularStyle(), Series: "m", Glyphs: gs}}, testNow))
	for _, want := range []string{
		`\DeclareTextSymbol{\textperthousandzero}{TU}{"EB4D}`,
		`\DeclareTextSymbol{\textperthousandzero}{T1}{24}`,
		`\DeclareTextSymbolDefault{\textperthousandzero}{T1}`,
	} {
		if !strings.Contains(pkg, want+"\n") {
			t.Errorf("the package lacks the line %s", want)
		}
	}
	if strings.Contains(pkg, `\textleaf`) {
		t.Error("the package declares the command of the leaf, which the font lacks")
	}
}

// TestUnicodeOf checks the code points of glyphs which have none of their
// own: a glyph with a suffix takes those of the glyph without it (also where
// that is named uniXXXX or uXXXX, and for a form in the Private Use Area), and
// a glyph named after several glyphs takes theirs.
func TestUnicodeOf(t *testing.T) {
	gs := []glyphs.Glyph{
		{Name: ".notdef"},
		{Name: "zero", Rune: '0'},
		{Name: "zero.taboldstyle"},
		{Name: "S", Rune: 'S'},
		{Name: "S_S"},
		{Name: "odd"},
		{Name: "uni2E40", Rune: 0x2E40},
		{Name: "hyphendbl.alt", Rune: 0xEB32},
		{Name: "u1F600.alt"},
		{Name: "quotedblbase", Rune: 0x201E},
		{Name: "quotedblbase.ts1", Rune: 0xEB52},
		{Name: "uni2e40.alt"},
	}
	byName := make(map[string]int)
	for i, g := range gs {
		byName[g.Name] = i
	}
	for i, want := range map[int]string{1: "0", 2: "0", 4: "SS", 5: "", 7: "\u2E40", 8: "\U0001F600", 10: "\u201E", 11: ""} {
		if got := string(unicodeOf(gs, byName, i)); got != want {
			t.Errorf("%s stands for %q, want %q", gs[i].Name, got, want)
		}
	}
}

// TestLicence checks that every generated text file carries the notice of the
// LaTeX Project Public License with the address of its text and the list of
// the files of the work, and no other licence.
func TestLicence(t *testing.T) {
	p := design.Regular()
	gs := allGlyphs()
	fonts := []*Font{
		{Style: design.RegularStyle(), Series: "m", Glyphs: gs},
		{Style: design.BoldStyle(), Series: "b", Glyphs: gs},
	}
	none := func(a, b int) int { return 0 }
	files, _, err := Files(fonts, []design.Params{p, p}, []KernFunc{none, none}, testNow)
	if err != nil {
		t.Fatal(err)
	}
	listed := strings.Join(workFiles(), ", ")
	count := 0
	for name, data := range files {
		if strings.HasSuffix(name, ".tfm") {
			continue
		}
		count++
		text := string(data)
		for _, want := range []string{
			"SPDX-License-Identifier: LPPL-1.3c",
			"LaTeX Project Public License, either version 1.3c",
			"https://www.latex-project.org/lppl.txt",
			"maintenance status",
			"Current Maintainer",
			"This work consists of the files " + listed + ".",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s lacks %q", name, want)
			}
		}
		if strings.Contains(text, "OFL-1.1") {
			t.Errorf("%s names the licence of the fonts as its own", name)
		}
	}
	if count != 8 {
		t.Errorf("%d text files, want 8", count)
	}
	for _, f := range strings.Split(listed, ", ") {
		found := false
		for name := range files {
			found = found || strings.HasSuffix(name, "/"+f)
		}
		if !found {
			t.Errorf("the notice lists %s, which is not generated", f)
		}
	}
}

// TestTS1SlotsOfTheKernel checks that TS1 uses no position at which the
// encoding of LaTeX has no sign: the position 115 is not used.
func TestTS1SlotsOfTheKernel(t *testing.T) {
	if TS1.Slots[115].used() {
		t.Errorf("TS1 uses the position 115 for %+v", TS1.Slots[115])
	}
	enc := string(EncFile(TS1, allGlyphs()))
	if !strings.Contains(enc, "/.notdef") {
		t.Error("the encoding file has no empty position")
	}
	index, missing := TS1.Resolve(allGlyphs())
	if index[115] != -1 || len(missing) != 0 {
		t.Errorf("position 115 has the glyph %d, missing %v", index[115], missing)
	}
}

// TestSlantedFiles checks the font definition files, the map file and the
// macro file for the slanted instances: the shape sl comes from the slanted
// metric file, the shape it and the series bx are substituted, plain TeX
// loads the slanted font for \sl and \it and the typewriter font of T1 for
// \tt, and the fonts of the footline are the fonts of the family.
func TestSlantedFiles(t *testing.T) {
	gs := allGlyphs()
	fonts := []*Font{
		{Style: design.RegularStyle(), Series: "m", Glyphs: gs},
		{Style: design.BoldStyle(), Series: "b", Glyphs: gs},
	}
	all := append(fonts, Slanted(fonts)...)
	fd := string(FD(T1, all, testNow))
	for _, want := range []string{
		`\DeclareFontShape{T1}{bedern}{m}{sl}{<-> bedern-regular-sl-t1}{}`,
		`\DeclareFontShape{T1}{bedern}{b}{sl}{<-> bedern-bold-sl-t1}{}`,
		`\DeclareFontShape{T1}{bedern}{m}{it}{<->ssub * bedern/m/sl}{}`,
		`\DeclareFontShape{T1}{bedern}{b}{it}{<->ssub * bedern/b/sl}{}`,
		`\DeclareFontShape{T1}{bedern}{bx}{sl}{<->ssub * bedern/b/sl}{}`,
		`\DeclareFontShape{T1}{bedern}{bx}{it}{<->ssub * bedern/b/it}{}`,
	} {
		if !strings.Contains(fd, want+"\n") {
			t.Errorf("the font definition file lacks %s", want)
		}
	}
	plain := string(Plain(all))
	for _, want := range []string{
		`\font\bedernrm=bedern-regular-t1`, `\let\tenrm=\bedernrm`,
		`\font\bedernbf=bedern-bold-t1`, `\let\tenbf=\bedernbf`,
		`\font\bedernsl=bedern-regular-sl-t1`, `\let\tensl=\bedernsl`, `\let\tenit=\bedernsl`,
		`\def\sl{\fam\slfam\bedernsl}`, `\def\it{\fam\itfam\bedernsl}`,
		`\font\bedernttfont=ectt1000`, `\let\tentt=\bedernttfont`,
	} {
		if !strings.Contains(plain, want+"\n") {
			t.Errorf("the macro file lacks %s", want)
		}
	}
	if strings.Count(plain, "\\font\\") != 4 {
		t.Errorf("the macro file loads %d fonts, want 4", strings.Count(plain, "\\font\\"))
	}
	pkg := string(Package(all, testNow))
	if strings.Count(pkg, "\\setmainfont") != 1 {
		t.Error("the package sets the main font more than once")
	}
}

// TestSlantedItalics checks that the italic correction of a slanted instance
// is the extent of the slanted outline beyond the advance.
func TestSlantedItalics(t *testing.T) {
	gs := []glyphs.Glyph{{Name: ".notdef"}, box("A", 'A', 0, 600, 480)}
	m := &Metrics{Slant: 0.25}
	m.Chars['A'] = &Char{Glyph: 1, Width: 500}
	m.SlantItalics(gs)
	if got, want := m.Chars['A'].Italic, 480+0.25*600-500.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("italic correction %g, want %g", got, want)
	}
}

// TestPackageOptions checks the option handling of the package.
func TestPackageOptions(t *testing.T) {
	pkg := string(Package([]*Font{{Style: design.RegularStyle(), Series: "m", Glyphs: allGlyphs()}}, testNow))
	for _, want := range []string{
		`\DeclareOption{oldstyle}`, `\DeclareOption{lining}`, `\DeclareOption{proportional}`,
		`\DeclareOption{tabular}`, `\DeclareOption*{\PackageWarning{bedern}{Unknown option`,
		`\ProcessOptions\relax`,
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("the package lacks %s", want)
		}
	}
	if strings.Index(pkg, `\ProcessOptions`) > strings.Index(pkg, `\RequirePackage`) {
		t.Error("the package loads packages before it processes its options")
	}
}

// TestGlyphToUnicodeForms checks that the double hyphen of TS1, a form whose
// base is named by its code point, is mapped to the character, both under its own
// name and under the name without the suffix, which pdfTeX looks up first.
func TestGlyphToUnicodeForms(t *testing.T) {
	gs := allGlyphs()
	fonts := []*Font{{Style: design.RegularStyle(), Series: "m", Glyphs: gs}}
	text := string(GlyphToUnicode(fonts))
	for _, want := range []string{
		"{tfm:bedern-regular-ts1/hyphendbl.alt}{2E40}",
		"{tfm:bedern-regular-ts1/hyphendbl}{2E40}",
	} {
		if !strings.Contains(text, "\\pdfglyphtounicode"+want+"\n") {
			t.Errorf("the mapping lacks %s", want)
		}
	}
}

// TestForTeXLowersTheCapitalAccents checks that the capital forms of the
// accents and the ties of the capitals are lowered by the difference of the
// cap height and the x-height, that no other glyph moves and that the list
// handed in stays as it is.  The metric file of TS1 then has the heights
// which TeX needs for them, since it raises an accent over a capital by the
// height of the capital less the x-height.
func TestForTeXLowersTheCapitalAccents(t *testing.T) {
	p := design.Regular()
	drop := p.CapHeight - p.XHeight
	gs := []glyphs.Glyph{
		{Name: ".notdef", Advance: 500},
		box("grave.case", 0, 900, 1000, 100),
		box("dotaccent.case", 0, 900, 1000, 100),
		box("tieaccentcapital", 0xEB63, 900, 1000, 400),
		box("tieaccentcapital.new", 0, 900, 1000, 400),
		box("tieaccentlowercase", 0xEB65, 900, 1000, 300),
		box("grave", '`', 900, 1000, 100),
	}
	got := ForTeX(p, gs)
	for i, g := range gs {
		_, y0, _, y1 := g.Path.BBox()
		_, z0, _, z1 := got[i].Path.BBox()
		lowerIt := i >= 1 && i <= 4
		switch {
		case lowerIt && (z0 != y0-drop || z1 != y1-drop):
			t.Errorf("%s: from %v-%v to %v-%v, want %v lower", g.Name, y0, y1, z0, z1, drop)
		case !lowerIt && (z0 != y0 || z1 != y1):
			t.Errorf("%s moved from %v-%v to %v-%v", g.Name, y0, y1, z0, z1)
		}
	}
	if _, y0, _, _ := gs[1].Path.BBox(); y0 != 900 {
		t.Errorf("the list handed in changed: the grave stands at %v", y0)
	}

	noKern := func(a, b int) int { return 0 }
	m := NewMetrics(TS1, p, ForTeX(p, gs), noKern)
	n := NewMetrics(TS1, p, gs, noKern)
	for _, c := range []int{0x00, 0x0A, 0x1B, 0x1D} {
		if m.Chars[c] == nil {
			t.Errorf("slot %d of TS1 has no glyph", c)
			continue
		}
		if want := n.Chars[c].Height - drop; m.Chars[c].Height != want {
			t.Errorf("slot %d of TS1 has height %v, want %v", c, m.Chars[c].Height, want)
		}
	}
	for _, c := range []int{0x1A, 0x1C} {
		if m.Chars[c] == nil {
			continue
		}
		if m.Chars[c].Height != n.Chars[c].Height {
			t.Errorf("slot %d of TS1 changed its height", c)
		}
	}
}

// TestPackageDeclaresTS1Subset checks that the package declares the subset of
// TS1 of the family with the number which the font definition file declares.
func TestPackageDeclaresTS1Subset(t *testing.T) {
	fonts := []*Font{{Style: design.RegularStyle(), Series: "m", Glyphs: allGlyphs()}}
	want := fmt.Sprintf("\\DeclareEncodingSubset{TS1}{bedern}{%d}\n", ts1SubsetOf(fonts))
	if pkg := string(Package(fonts, testNow)); !strings.Contains(pkg, want) {
		t.Errorf("the package lacks the line %s", want)
	}
	if fd := string(FD(TS1, fonts, testNow)); !strings.Contains(fd, want) {
		t.Errorf("the font definition file lacks the line %s", want)
	}
}
