// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// scriptSources returns the glyphs which have the script forms: the figures,
// the 24 Greek capitals and the Greek lowercase of
// the maths italic (U+1D6FC to U+1D714 and the six variants).
func scriptSources(gs []Glyph) []Glyph {
	var out []Glyph
	figures, caps := scriptNames()
	want := map[string]bool{}
	for _, n := range figures {
		want[n] = true
	}
	for _, d := range caps {
		want[d.name] = true
	}
	for _, g := range gs {
		if want[g.Name] || (g.Rune >= 0x1D6FC && g.Rune <= 0x1D714) || (g.Rune >= 0x1D716 && g.Rune <= 0x1D71B) {
			out = append(out, g)
		}
	}
	return out
}

// scriptGlyphs returns the glyphs of the font with their script forms: all
// of them in the full tier, and in the routine tier the figures and the Greek
// capitals alone, built from the constructions of the regular.
func scriptGlyphs(t *testing.T) []Glyph {
	t.Helper()
	if !testing.Short() {
		gs, err := regularGlyphs()
		if err != nil {
			t.Fatal(err)
		}
		return gs
	}
	r := newRoman(design.Regular())
	figures, _ := scriptNames()
	var gs []Glyph
	for i, n := range figures {
		path, adv := r.get(n)
		gs = append(gs, Glyph{Name: n, Rune: rune('0' + i), Advance: math.Round(adv), Path: path, Class: Ord, Like: n})
	}
	gs = append(gs, greekCaps(r)...)
	gs, err := addScripts(gs, false)
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

// TestScriptFormsExist checks that every figure, Greek capital and Greek
// lowercase letter has both script forms, named and present, and that each
// has the advance of its text form times 1.10 and 1.22, the Greek lowercase
// times 1.12 and 1.27 (a little more where the overhangs of the maths italic
// need it).
func TestScriptFormsExist(t *testing.T) {
	gs := scriptGlyphs(t)
	by := map[string]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	src := scriptSources(gs)
	want := 10 + 24 + 31
	if testing.Short() {
		want = 10 + 24
	}
	if len(src) != want {
		t.Fatalf("%d glyphs have script forms, want %d", len(src), want)
	}
	for _, g := range src {
		for level := 1; level <= 2; level++ {
			name := g.Script[level-1]
			s, ok := by[name]
			if !ok {
				t.Errorf("%s has no form %q for level %d", g.Name, name, level)
				continue
			}
			f := scriptLevels[level-1].advance
			if g.Rune >= 0x1D6FC && g.Rune <= 0x1D71B {
				f = scriptLevels[level-1].greekAdvance
			}
			if r := s.Advance / g.Advance; r < f-0.005 || r > f+0.03 {
				t.Errorf("%s is %.3f times as wide as %s, want %.2f", name, r, g.Name, f)
			}
			if len(s.Path) == 0 {
				t.Errorf("%s is empty", name)
			}
		}
	}
}

// TestScriptFormsOfFiguresAndGreekAreHeavier measures, on the script forms
// of the figures, the Greek capitals and the Greek lowercase against their
// text forms, a stem at half the height (the figures and the capitals by
// their factors figStem and capStem of scriptLevels, the lowercase 1.08 to
// 1.17 and 1.15 to 1.35 as the letters of the maths italic) and the thinnest
// part at the top of the round
// letters (at least 36 at the first level and 50 at the second, the Greek
// lowercase 40 as the letters of the maths italic).
func TestScriptFormsOfFiguresAndGreekAreHeavier(t *testing.T) {
	gs := scriptGlyphs(t)
	by := map[string]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	xh := design.Regular().XHeight
	stemAt := func(p outline.Path, y float64) float64 {
		xs := p.CrossingsX(y)
		return xs[1] - xs[0]
	}
	thinTop := func(p outline.Path) float64 {
		x0, _, x1, _ := p.BBox()
		ys := p.CrossingsY((x0 + x1) / 2)
		return ys[len(ys)-1] - ys[len(ys)-2]
	}
	mid := func(p outline.Path) float64 {
		_, y0, _, y1 := p.BBox()
		return (y0 + y1) / 2
	}
	type check struct {
		name  string
		lower bool // a letter of the maths italic
		stem  bool
		thin  bool
	}
	figure := map[string]bool{"zero": true, "one": true, "nine": true}
	for _, c := range []check{
		{"zero", false, true, true}, {"one", false, true, false}, {"nine", false, false, true},
		{"uni0393", false, true, true}, {"uni03A0", false, true, true},
		{"uni039F", false, true, true}, {"uni0398", false, true, true},
		{"u1D704", true, true, false}, {"u1D702", true, true, false}, {"u1D70A", true, false, true},
	} {
		text, ok := by[c.name]
		if !ok && testing.Short() && c.lower {
			continue
		}
		for level := 1; level <= 2; level++ {
			s := by[text.Script[level-1]]
			if c.stem {
				y := mid(text.Path)
				if c.lower {
					y = xh / 2
				}
				r := stemAt(s.Path, y) / stemAt(text.Path, y)
				want := scriptLevels[level-1].capStem
				if figure[c.name] {
					want = scriptLevels[level-1].figStem
				}
				lo, hi := want-0.03, want+0.03
				if c.lower {
					lo, hi = [2]float64{1.08, 1.15}[level-1], [2]float64{1.17, 1.35}[level-1]
				}
				if r < lo || r > hi {
					t.Errorf("the stem of %s is %.3f times that of %s, want %.2f to %.2f", s.Name, r, text.Name, lo, hi)
				}
			}
			if c.thin {
				floor := scriptLevels[level-1].thin
				if c.lower {
					floor = scriptLevels[level-1].minThin
				}
				if d := thinTop(s.Path); d < floor-0.5 {
					t.Errorf("the top of %s is %.1f thick, want at least %v", s.Name, d, floor)
				}
			}
		}
	}
}

// TestScriptFormsKeepTheHeights checks that the script forms of the figures
// and the capitals have the heights of their text forms.
func TestScriptFormsKeepTheHeights(t *testing.T) {
	gs := scriptGlyphs(t)
	by := map[string]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	for _, g := range scriptSources(gs) {
		_, y0, _, y1 := g.Path.BBox()
		for level := 1; level <= 2; level++ {
			s := by[g.Script[level-1]]
			_, sy0, _, sy1 := s.Path.BBox()
			if d0, d1 := sy0-y0, sy1-y1; d0 < -25 || d0 > 25 || d1 < -25 || d1 > 25 {
				t.Errorf("%s moves the bottom by %.0f and the top by %.0f", s.Name, d0, d1)
			}
		}
	}
}

// TestScriptCapitalsAreNotCompressed requires the ink of the script forms of
// the Greek capitals which differ from the Latin capitals to be 1.08 and 1.15
// times as wide as that of the text forms at the median; their advance is
// 1.10 and 1.22 times that of the text forms.
func TestScriptCapitalsAreNotCompressed(t *testing.T) {
	gs := scriptGlyphs(t)
	by := map[string]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	for level := 1; level <= 2; level++ {
		var ratios []float64
		for _, n := range []string{"uni0393", "uni0394", "uni0398", "uni039B", "uni039E", "uni03A0",
			"uni03A3", "uni03A5", "uni03A6", "uni03A8", "uni03A9"} {
			text, ok := by[n]
			if !ok {
				t.Fatalf("no glyph %s", n)
			}
			s := by[text.Script[level-1]]
			tx0, _, tx1, _ := text.Path.BBox()
			x0, _, x1, _ := s.Path.BBox()
			ratios = append(ratios, (x1-x0)/(tx1-tx0))
		}
		slices.Sort(ratios)
		if m, want := ratios[len(ratios)/2], []float64{1.08, 1.15}[level-1]; m < want {
			t.Errorf("at level %d the ink of the Greek capitals is %.3f times as wide as the text forms' at the median, want at least %.2f", level, m, want)
		}
	}
}

// TestScriptFiguresAreNotHeavierThanLetters requires, in the full tier, the
// stem of the script forms of the figure 1 to be at most 1.10 times that of
// the script forms of the italic l, the script Greek lowercase to leave a
// median white of 0.14 and 0.20 of the x-height between two letters, and ν to
// be wider than v by at least 3 per cent.
func TestScriptFiguresAreNotHeavierThanLetters(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the script forms of the Greek lowercase and of the italic")
	}
	gs, err := regularGlyphs()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]Glyph{}
	byRune := map[rune]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
		if g.Rune != 0 {
			byRune[g.Rune] = g
		}
	}
	xh := design.Regular().XHeight
	stemAt := func(p outline.Path, y float64) float64 {
		xs := p.CrossingsX(y)
		return xs[1] - xs[0]
	}
	for level := 1; level <= 2; level++ {
		one := by[by["one"].Script[level-1]]
		l := byRune[0x1D459]
		ls := by[l.Script[level-1]]
		_, y0, _, y1 := one.Path.BBox()
		_, ly0, _, ly1 := ls.Path.BBox()
		r := stemAt(one.Path, (y0+y1)/2) / stemAt(ls.Path, (ly0+ly1)/2)
		if r > 1.10 {
			t.Errorf("at level %d the stem of the figure 1 is %.3f times that of the letter l, want at most 1.10", level, r)
		}
		var gaps []float64
		var forms []Glyph
		for r := rune(0x1D6FC); r <= 0x1D714; r++ {
			forms = append(forms, by[byRune[r].Script[level-1]])
		}
		for _, a := range forms {
			_, _, ax1, _ := a.Path.BBox()
			for _, b := range forms {
				bx0, _, _, _ := b.Path.BBox()
				gaps = append(gaps, a.Advance-ax1+bx0)
			}
		}
		slices.Sort(gaps)
		if m, want := gaps[len(gaps)/2]/xh, []float64{0.14, 0.20}[level-1]; m < want {
			t.Errorf("at level %d the median white between two script Greek letters is %.3f of the x-height, want at least %.2f", level, m, want)
		}
		nu, v := by[byRune[0x1D708].Script[level-1]], by[byRune[0x1D463].Script[level-1]]
		nx0, _, nx1, _ := nu.Path.BBox()
		vx0, _, vx1, _ := v.Path.BBox()
		if r := (nx1 - nx0) / (vx1 - vx0); r < 1.03 {
			t.Errorf("at level %d ν is %.3f times as wide as v, want at least 1.03", level, r)
		}
	}
}

// TestScriptFiguresKeepTheirFloors holds the two rules of the script figures
// which the factor of the thin strokes alone does not: the tips of their
// serifs have the floor of the level (36 and 50), and
// the ink of the figure 1 is 1.08 and 1.13 times that of its text form.
func TestScriptFiguresKeepTheirFloors(t *testing.T) {
	for level := 1; level <= 2; level++ {
		s := scriptLevels[level-1]
		if tip := scriptRomans[scriptFigures][level-1]().p.SerifThick; tip < s.thin {
			t.Errorf("level %d: the tips of the serifs of the figures are %.0f thick, want at least %v", level, tip, s.thin)
		}
		ink := func(r *roman) float64 {
			path, _ := r.get("one")
			x0, _, x1, _ := path.BBox()
			return x1 - x0
		}
		got := ink(scriptRomans[scriptFigures][level-1]()) / ink(newRoman(design.Regular()))
		if got < s.oneWide-0.005 {
			t.Errorf("level %d: the ink of the figure 1 is %.3f times that of the text form, want at least %.2f", level, got, s.oneWide)
		}
	}
}

// TestScriptXFollowsTheRule requires, in the full tier, the ink of the script
// forms of the maths italic x to be 1.10 and 1.22 times as wide as that of
// its text form, the rule of the script forms, within 0.02, the tolerance of
// the ink of the script capitals.  The script χ is then 1.13 to 1.29 times as
// wide as the script x at both levels.
func TestScriptXFollowsTheRule(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the script forms of the Greek lowercase and of the italic")
	}
	gs, err := regularGlyphs()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]Glyph{}
	byRune := map[rune]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
		if g.Rune != 0 {
			byRune[g.Rune] = g
		}
	}
	ink := func(g Glyph) float64 {
		x0, _, x1, _ := g.Path.BBox()
		return x1 - x0
	}
	x, chi := byRune[0x1D465], byRune[0x1D712]
	for level := 1; level <= 2; level++ {
		xs, cs := by[x.Script[level-1]], by[chi.Script[level-1]]
		want := scriptLevels[level-1].advance
		if r := ink(xs) / ink(x); r < want-0.02 || r > want+0.02 {
			t.Errorf("at level %d the ink of x is %.3f times that of the text form, want %.2f to %.2f", level, r, want-0.02, want+0.02)
		}
		if r := ink(cs) / ink(xs); r < 1.13 || r > 1.29 {
			t.Errorf("at level %d χ is %.3f times as wide as x, want 1.13 to 1.29", level, r)
		}
	}
}
