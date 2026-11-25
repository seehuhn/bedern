// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// signGlyphs returns the signs, ℓ and the primes with their script forms:
// the glyphs of the font in the full tier, and in the routine tier the text
// forms of these glyphs alone, drawn by the constructions of the regular.
func signGlyphs(t *testing.T) map[string]Glyph {
	t.Helper()
	var gs []Glyph
	if !testing.Short() {
		var err error
		if gs, err = regularGlyphs(); err != nil {
			t.Fatal(err)
		}
	} else {
		r := newRoman(design.Regular())
		names := append(append(slices.Clone(scriptSignNames()), "uni2113"), scriptPrimeNames...)
		for _, n := range names {
			path, adv := drawSign(r, n)
			gs = append(gs, Glyph{Name: n, Rune: 1, Advance: math.Round(adv), Path: path, Class: Ord, Like: n})
		}
		var err error
		if gs, err = addScripts(gs, false); err != nil {
			t.Fatal(err)
		}
	}
	by := map[string]Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	return by
}

// TestScriptSignsExist checks that the 33 signs, ℓ and the four primes have
// both script forms, named and present, with the advance of their text form
// times 1.10 and 1.22 (a little more where the overhangs need it), and that
// the signs and ℓ keep the heights of their text forms within 25 units (the
// head of the comma, which grows about its centre with the weight, within 30).
func TestScriptSignsExist(t *testing.T) {
	by := signGlyphs(t)
	names := slices.Clone(scriptSignNames())
	if len(names) != 33 {
		t.Fatalf("%d signs, want 33", len(names))
	}
	names = append(append(names, "uni2113"), scriptPrimeNames...)
	for _, n := range names {
		g, ok := by[n]
		if !ok {
			t.Errorf("no glyph %s", n)
			continue
		}
		_, y0, _, y1 := g.Path.BBox()
		for level := 1; level <= 2; level++ {
			s, ok := by[g.Script[level-1]]
			if !ok || g.Script[level-1] == "" {
				t.Errorf("%s has no form for level %d", n, level)
				continue
			}
			if len(s.Path) == 0 {
				t.Errorf("%s is empty", s.Name)
			}
			f := scriptLevels[level-1].advance
			if r := s.Advance / g.Advance; r < f-0.005 || r > f+0.03 {
				t.Errorf("%s is %.3f times as wide as %s, want %.2f", s.Name, r, n, f)
			}
			if slices.Contains(scriptPrimeNames, n) {
				continue
			}
			_, sy0, _, sy1 := s.Path.BBox()
			limit := 25.0
			if n == "comma" {
				limit = 30 // the head grows about its centre with the weight of the level
			}
			if d0, d1 := sy0-y0, sy1-y1; math.Abs(d0) > limit || math.Abs(d1) > limit {
				t.Errorf("%s moves the bottom by %.0f and the top by %.0f", s.Name, d0, d1)
			}
		}
	}
}

// TestScriptSignsAreHeavier measures the rule of the signs against that of
// their text forms: the minus, the bars of the equivalence sign, the arm of
// the plus and the bar are 1.13 and 1.24 times as thick (within 0.03), and
// none is thinner than the floors of 36 and 50.
func TestScriptSignsAreHeavier(t *testing.T) {
	by := signGlyphs(t)
	thickV := func(g Glyph) float64 {
		x0, _, x1, _ := g.Path.BBox()
		ys := g.Path.CrossingsY((x0 + x1) / 2)
		return ys[1] - ys[0]
	}
	thickH := func(y float64) func(Glyph) float64 {
		return func(g Glyph) float64 {
			xs := g.Path.CrossingsX(y)
			return xs[1] - xs[0]
		}
	}
	axis := design.Regular().Punct.MathAxis
	for _, c := range []struct {
		name  string
		thick func(Glyph) float64
	}{
		{"minus", thickV}, {"plus", thickH(axis + 100)}, {"bar", thickH(axis)}, {"uni2261", thickV},
	} {
		text := by[c.name]
		for level := 1; level <= 2; level++ {
			s := by[text.Script[level-1]]
			sc := scriptLevels[level-1]
			if r := c.thick(s) / c.thick(text); r < sc.stem-0.03 || r > sc.stem+0.03 {
				t.Errorf("the rule of %s is %.3f times that of %s, want %.2f", s.Name, r, c.name, sc.stem)
			}
			if d := c.thick(s); d < sc.thin-0.5 {
				t.Errorf("the rule of %s is %.1f thick, want at least %v", s.Name, d, sc.thin)
			}
		}
	}
}

// TestScriptPrimesStandLow checks the vertical extent of the script forms of
// the primes (about 100 to 530 at the first level and
// 90 to 630 at the second) and that a prime in a superscript, shifted by 376
// and scaled by 0.70, ends between 740 and 760.
func TestScriptPrimesStandLow(t *testing.T) {
	by := signGlyphs(t)
	for _, n := range scriptPrimeNames {
		for level := 1; level <= 2; level++ {
			s := by[by[n].Script[level-1]]
			_, y0, _, y1 := s.Path.BBox()
			want := scriptPrimes[level-1]
			if math.Abs(y0-want.bottom) > 5 || math.Abs(y1-want.top) > 5 {
				t.Errorf("%s stands from %.0f to %.0f, want %.0f to %.0f", s.Name, y0, y1, want.bottom, want.top)
			}
			if level == 1 {
				if top := 376 + 0.70*y1; top < 740 || top > 760 {
					t.Errorf("%s in a superscript ends at %.0f, want about 750", s.Name, top)
				}
			}
		}
	}
}
