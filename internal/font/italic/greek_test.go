// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"sync"
	"testing"
)

// The Greek letters of the default parameters, built once and shared by the
// tests.
var (
	greekOnce sync.Once
	greekSet  map[string]*Glyph
	greekErr  error
)

// greek returns the Greek letter of the default parameters.
func greek(t *testing.T, name string) *Glyph {
	t.Helper()
	greekOnce.Do(func() {
		greekSet = map[string]*Glyph{}
		gs, err := GreekGlyphs(Default())
		if err != nil {
			greekErr = err
			return
		}
		for _, g := range gs {
			greekSet[g.Name] = g
		}
	})
	if greekErr != nil {
		t.Fatal(greekErr)
	}
	g, ok := greekSet[name]
	if !ok {
		t.Fatalf("no Greek letter %s", name)
	}
	return g
}

// TestGreekLettersAreSmooth asks every Greek letter for a clean build: no
// warning of the stroke tool and no finding of the italic profile of the
// smoothness check which is not marked as intended.
func TestGreekLettersAreSmooth(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	for _, name := range GreekNames() {
		g := greek(t, name)
		for _, w := range g.Warnings {
			t.Errorf("%s: warning: %s", name, w)
		}
		for _, f := range g.Findings {
			t.Errorf("%s: %s at %.0f,%.0f: %s", name, f.Kind, f.X, f.Y, f.Detail)
		}
	}
}

// TestGreekStemsOnTheLattice asks of every straight stem of the slant of a
// Greek letter what TestStemsOnTheLattice asks of the Latin ones.
func TestGreekStemsOnTheLattice(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	p := Default()
	for _, name := range GreekNames() {
		checkLattice(t, p, name, greek(t, name).Path)
	}
}

// TestGreekWeights asks that no run of ink of a Greek letter, measured
// horizontally at a quarter, a half and three quarters of the x-height, is
// thinner than the floor of the hairline at 8 pt, and that every run at
// half the x-height away from a crossing is no heavier than the thick of
// the pen measured horizontally.
func TestGreekWeights(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	p := Default()
	// crossed are the letters whose strokes cross or join near half the
	// x-height, where a run holds two strokes.
	crossed := map[string]bool{"alpha": true, "theta": true, "phi": true, "phiLoop": true, "psi": true, "chi": true, "epsilon": true, "lambda": true, "omega": true, "gamma": true, "beta": true, "kappa": true, "kappaVar": true, "epsilonLunate": true}
	for _, name := range GreekNames() {
		g := greek(t, name)
		for _, f := range []float64{0.25, 0.5, 0.75} {
			y := f*p.XHeight + 0.5
			for _, r := range runs(g.Path, y) {
				if w := r[1] - r[0]; w < greekHairFloor {
					t.Errorf("%s: a run of ink of %.1f at %.0f, want at least %v", name, w, y, greekHairFloor)
				}
				if w := r[1] - r[0]; f == 0.5 && !crossed[name] && w > greekRunCeiling {
					t.Errorf("%s: a run of ink of %.1f at %.0f, want at most %v", name, w, y, greekRunCeiling)
				}
			}
		}
	}
}

// greekHairFloor is the thinnest run of ink at 8 pt, 24 units, a little below
// the thin of the pen; greekRunCeiling is the thick of the pen measured
// horizontally on a stroke which leans with the slant, with a margin.
const (
	greekHairFloor  = 24.0
	greekRunCeiling = 100.0
)

// TestGreekCounters asks that the white inside every Greek letter, measured
// horizontally between its runs of ink at the heights listed, is at least
// 40 units.
func TestGreekCounters(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	p := Default()
	heights := map[string][]float64{
		"omicron": {0.3, 0.5, 0.7}, "sigma": {0.3, 0.5, 0.7}, "theta": {0.25, 1.15},
		"phi": {0.3, 0.7}, "rho": {0.3, 0.5, 0.7}, "alpha": {0.3, 0.7},
		"pi": {0.3, 0.6}, "eta": {0.3, 0.6}, "mu": {0.4, 0.7}, "upsilon": {0.4, 0.7},
		"nu": {0.5, 0.8}, "delta": {0.3, 0.5}, "psi": {0.4, 0.7}, "omega": {0.3},
		"beta": {0.3, 1.2}, "phiLoop": {0.3, 0.7},
	}
	for name, fs := range heights {
		if _, ok := greekLetters[name]; !ok {
			continue
		}
		g := greek(t, name)
		for _, f := range fs {
			y := f*p.XHeight + 0.5
			rs := runs(g.Path, y)
			if len(rs) < 2 {
				t.Errorf("%s: no counter at %.0f", name, y)
			}
			for i := 1; i < len(rs); i++ {
				if w := rs[i][0] - rs[i-1][1]; w < 40 {
					t.Errorf("%s: a white of %.1f at %.0f, want at least 40", name, w, y)
				}
			}
		}
	}
}

// greekInk returns the leftmost and the rightmost ink of the outline at the
// height y, or NaN if there is none.
func greekInk(g *Glyph, y float64) (float64, float64) {
	rs := runs(g.Path, y)
	if len(rs) == 0 {
		return math.NaN(), math.NaN()
	}
	return rs[0][0], rs[len(rs)-1][1]
}

// TestGreekApartFromLatin asks of each Greek letter which can be taken for
// a Latin italic letter in a formula the difference which keeps them apart,
// as a measure of the outline.
func TestGreekApartFromLatin(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	p := Default()
	x := p.XHeight
	has := func(n string) bool { _, ok := greekLetters[n]; return ok }
	// ν has a pointed foot and v a round one: just above the baseline the
	// ink of ν is less than two thirds as wide as that of v.
	if has("nu") {
		a0, a1 := greekInk(greek(t, "nu"), 20)
		b0, b1 := greekInk(glyph(t, "v"), 20)
		if a1-a0 > 2.0/3*(b1-b0) {
			t.Errorf("ν is %.0f wide at 20, v %.0f: the foot of ν is not pointed", a1-a0, b1-b0)
		}
	}
	// υ has no right stem and no exit: its ink near the baseline stays
	// left of its ink at half the x-height, while u reaches out to the right.
	if has("upsilon") {
		g := greek(t, "upsilon")
		_, low := greekInk(g, 40)
		_, mid := greekInk(g, x/2)
		if low >= mid {
			t.Errorf("υ reaches %.0f at 40 and %.0f at half the x-height: it has an exit", low, mid)
		}
		_, ulow := greekInk(glyph(t, "u"), 40)
		_, umid := greekInk(glyph(t, "u"), x/2)
		if ulow <= umid {
			t.Errorf("u reaches %.0f at 40 and %.0f at half the x-height", ulow, umid)
		}
	}
	// ω has no stroke to the x-height in its middle: no ink at 0.85 of the
	// x-height in the middle third of its ink, where w has its second stem.
	if has("omega") {
		g := greek(t, "omega")
		x0, _, x1, _ := bbox(g.Path)
		for _, r := range runs(g.Path, 0.85*x) {
			if r[1] > x0+(x1-x0)/3 && r[0] < x1-(x1-x0)/3 {
				t.Errorf("ω has ink at %.0f..%.0f in its middle third at 0.85 of the x-height", r[0], r[1])
			}
		}
	}
	// χ descends to the depth of the tails, x stands on the baseline.
	if has("chi") {
		_, y0, _, _ := bbox(greek(t, "chi").Path)
		if y0 > -0.8*p.TailDepth {
			t.Errorf("χ reaches down to %.0f, want below %.0f", y0, -0.8*p.TailDepth)
		}
	}
	// ρ has no foot serif and no entry: at the descender its ink is the
	// stem alone, and at 0.85 of the x-height it does not reach left of its
	// round.
	if has("rho") {
		g := greek(t, "rho")
		a0, a1 := greekInk(g, -p.TailDepth+5)
		if a1-a0 > p.Stem+10 {
			t.Errorf("the descender of ρ is %.0f wide at its foot, want at most %v", a1-a0, p.Stem+10)
		}
		hi, _ := greekInk(g, 0.85*x)
		mid, _ := greekInk(g, 0.5*x)
		if hi < mid {
			t.Errorf("ρ reaches %.0f at 0.85 of the x-height, left of %.0f at half of it: it has an entry", hi, mid)
		}
	}
	// κ stays at the x-height, k rises to the ascender.
	if has("kappa") {
		_, _, _, y1 := bbox(greek(t, "kappa").Path)
		if y1 > x+p.Overshoot+1 {
			t.Errorf("κ reaches up to %.0f, above the x-height", y1)
		}
	}
	// ι has no entry: at 0.85 of the x-height its ink is the stem alone,
	// where the entry of ı reaches far to the left.
	if has("iota") {
		a0, a1 := greekInk(greek(t, "iota"), 0.85*x)
		if a1-a0 > p.Stem+10 {
			t.Errorf("ι is %.0f wide at 0.85 of the x-height, want at most %v", a1-a0, p.Stem+10)
		}
	}
	// α has no stem rising to the x-height in a flat cut: at 0.9 of the
	// x-height its rightmost ink is the thin tip, less than two thirds of a
	// stem wide, where a has the full stem.
	if has("alpha") {
		last := func(g *Glyph) float64 {
			rs := runs(g.Path, 0.9*x)
			r := rs[len(rs)-1]
			return r[1] - r[0]
		}
		if w := last(greek(t, "alpha")); w > 2.0/3*p.Stem {
			t.Errorf("the tip of α is %.0f wide at 0.9 of the x-height, want at most %.0f", w, 2.0/3*p.Stem)
		}
		if w := last(glyph(t, "a")); w < p.Stem-1 {
			t.Errorf("the stem of a is %.0f wide at 0.9 of the x-height", w)
		}
	}
}
