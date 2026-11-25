// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"slices"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs"

	"seehuhn.de/go/bedern/internal/fair"
)

// greekCapsOf holds the Greek capitals by the name of their parameter set,
// each built once.
var greekCapsOf sync.Map

func greekCapsFor(s designtest.Set) map[string]Glyph {
	v, _ := greekCapsOf.LoadOrStore(s.Name, sync.OnceValue(func() map[string]Glyph { return byName(GreekCapitals(s.Params)) }))
	return v.(func() map[string]Glyph)()
}

// drawnCaps are the eleven capitals drawn new; the other thirteen are the
// Latin capitals of the regular, which its own tests check.
var drawnCaps = func() map[string]bool {
	m := map[string]bool{}
	for _, d := range greekCapitals[:11] {
		m[d.name] = true
	}
	return m
}()

func regularGreekCaps() map[string]Glyph {
	return greekCapsFor(designtest.Set{Name: "regular", Params: design.Regular()})
}

// TestGreekCapitalsPassTheChecks applies the structural checks of the
// regular, with the bounding-box rule of the capital each is made like, to
// every Greek capital and the fairness analysis to the eleven drawn new, on the regular parameters
// and the first perturbed set; a finding of the analysis is allowed only
// where intendedCapFinding gives it a reason.
func TestGreekCapitalsPassTheChecks(t *testing.T) {
	for _, s := range derivedSets(1) {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			for _, g := range greekCapsFor(s) {
				if err := glyphs.CheckOutline(s.Params, g.Like, g.Path, g.Advance); err != nil {
					t.Errorf("%s: %v", g.Name, err)
				}
				if !drawnCaps[g.Name] {
					continue
				}
				for _, f := range fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings {
					if intendedCapFinding(g.Name, s.Params, f) == "" {
						t.Errorf("%s: %s at %.0f,%.0f: %s", g.Name, f.Kind, f.X, f.Y, f.Detail)
					}
				}
			}
		})
	}
}

// TestGreekCapitalsWithinTheirRanges requires the ink width of every Greek
// capital over that of H to lie inside its range, and its ink to stand on the baseline and
// reach the cap height, the rounds of Θ and Ω with the overshoot of O and
// the apex of Δ and Λ with that of A.
func TestGreekCapitalsWithinTheirRanges(t *testing.T) {
	p := design.Regular()
	h := design.CapHInk(p)
	ranges := map[string][2]float64{
		"uni0393": {0.652, 0.804}, "uni0394": {0.795, 1.081}, "uni0398": {0.904, 0.974},
		"uni039B": {0, 9}, "uni039E": {0.773, 0.852}, "uni03A0": {0.924, 1.025},
		"uni03A3": {0.745, 0.892}, "uni03A5": {0.824, 0.991}, "uni03A6": {0.856, 0.951},
		"uni03A8": {0.942, 1.116}, "uni03A9": {0.876, 1.030},
	}
	tops := map[string]float64{
		"uni0398": p.CapHeight + p.OvershootUpper, "uni03A9": p.CapHeight + p.OvershootUpper,
		"uni0394": p.CapHeight + p.CapDiag.ApexRise, "uni039B": p.CapHeight + p.CapDiag.ApexRise,
	}
	a, _ := newRoman(p).get("A")
	ax0, _, ax1, _ := a.Round().BBox()
	lx0, _, lx1, _ := regularGreekCaps()["uni039B"].Path.Round().BBox()
	// Λ is A without its bar, 0.92 to 1.03 of A.
	if r := (lx1 - lx0) / (ax1 - ax0); r < 0.92 || r > 1.03 {
		t.Errorf("Λ: the ink is %.3f of A wide, outside the range 0.92 to 1.03", r)
	}
	for name, g := range regularGreekCaps() {
		r, ok := ranges[name]
		if !ok {
			continue
		}
		x0, y0, x1, y1 := g.Path.Round().BBox()
		if w := (x1 - x0) / h; w < r[0] || w > r[1] {
			t.Errorf("%s: the ink is %.3f of H wide, outside the range %.3f to %.3f", name, w, r[0], r[1])
		}
		top, bottom := p.CapHeight, 0.0
		if v, ok := tops[name]; ok {
			top = v
		}
		if name == "uni0398" {
			bottom = -p.OvershootUpper
		}
		if math.Abs(y1-top) > 1 || math.Abs(y0-bottom) > 1 {
			t.Errorf("%s: the ink reaches from %.0f to %.0f, want %.0f to %.0f", name, y0, y1, bottom, top)
		}
	}
}

// TestGreekCapitalStems requires the stems of Γ, Π, Υ, Φ and Ψ to be the
// capital stem within one unit, measured horizontally away from the serifs
// and the joins, and the sides of the rounds of Θ, Φ and Ω and the thick
// diagonal of Δ to be no heavier than the round of O.
func TestGreekCapitalStems(t *testing.T) {
	p := design.Regular()
	c := p.Caps
	gs := regularGreekCaps()
	stems := []struct {
		name string
		y    float64
		i    int // the run of the stem among the runs at y
	}{
		{"uni0393", 300, 0}, {"uni03A0", 300, 0}, {"uni03A0", 300, 1},
		{"uni03A5", 150, 0}, {"uni03A6", 90, 0}, {"uni03A8", 150, 0},
		{"uni03A6", 0.5 * p.CapHeight, 1}, {"uni03A8", 0.75 * p.CapHeight, 1},
		{"uni03A8", 0.75 * p.CapHeight, 0}, {"uni03A8", 0.75 * p.CapHeight, 2},
	}
	for _, s := range stems {
		runs := runAt(gs[s.name].Path.Round(), s.y)
		if s.i >= len(runs) {
			t.Errorf("%s: %d runs at %.0f", s.name, len(runs), s.y)
			continue
		}
		if w := runs[s.i][1] - runs[s.i][0]; math.Abs(w-c.Stem) > 1 {
			t.Errorf("%s: the stem %d at %.0f is %.1f wide, want %v", s.name, s.i, s.y, w, c.Stem)
		}
	}
	for _, name := range []string{"uni0398", "uni03A6", "uni03A9"} {
		runs := runAt(gs[name].Path.Round(), 0.5*p.CapHeight)
		for _, r := range []([2]float64){runs[0], runs[len(runs)-1]} {
			if w := r[1] - r[0]; w > c.Round+1 || w < c.Stem {
				t.Errorf("%s: a side is %.1f wide at half the cap height, want %v to %v", name, w, c.Stem, c.Round)
			}
		}
	}
}

// TestGreekCapitalsKeepTheirWhites requires the whites which close first at
// small sizes to be at least 40 units wide: between the bar of Θ and the
// round, between the serifs of the bars of Ξ, between the legs of Ω, between
// the inner serifs of the arms of Ψ and the serifs of its stem, and inside
// the arms of Υ below their crests.
func TestGreekCapitalsKeepTheirWhites(t *testing.T) {
	p := design.Regular()
	gs := regularGreekCaps()
	gap := func(name string, y float64) float64 {
		runs := runAt(gs[name].Path.Round(), y)
		m := math.Inf(1)
		for i := 0; i+1 < len(runs); i++ {
			m = math.Min(m, runs[i+1][0]-runs[i][1])
		}
		return m
	}
	vgap := func(name string, x float64) float64 {
		path := gs[name].Path.Round()
		ys := path.CrossingsY(x)
		slices.Sort(ys)
		m := math.Inf(1)
		for i := 1; i+1 < len(ys); i += 2 {
			m = math.Min(m, ys[i+1]-ys[i])
		}
		return m
	}
	cases := []struct {
		what string
		d    float64
	}{
		{"Θ: bar to round", gap("uni0398", thetaBarHeight*p.CapHeight)},
		{"Ξ: serifs of the middle and the top bar", vgap("uni039E", 0.5*design.CapHInk(p)*xiInk*(1-xiMiddle/2)+p.Caps.SideArmE)},
		{"Ω: the legs", gap("uni03A9", 0.2*p.CapHeight)},
		{"Ψ: the head serifs", gap("uni03A8", p.CapHeight-p.Caps.SerifThick/2)},
		{"Υ: below the crests", gap("uni03A5", 0.85*p.CapHeight)},
	}
	for _, cs := range cases {
		if cs.d < 40 {
			t.Errorf("%s: the white is %.1f, want at least 40", cs.what, cs.d)
		}
	}
}

// TestGreekCapitalsAreNotLatin measures what keeps Γ from F and E and Υ from
// Y: Γ has no ink right of its stem between the foot serif and the serif of
// its arm, where F has its middle arm; the arms of Υ hang down at their ends;
// at nine tenths of the cap height a line crosses four runs of ink
// (two ends and two arms) where it crosses two on Y, and its arms are curves
// which leave the stem without the corner of the crotch of Y below half the
// height of the arms.
func TestGreekCapitalsAreNotLatin(t *testing.T) {
	p := design.Regular()
	c := p.Caps
	gs := regularGreekCaps()
	gamma := gs["uni0393"].Path.Round()
	for _, y := range []float64{c.FMiddleHeight - c.Bar, c.FMiddleHeight, c.FMiddleHeight + c.Bar} {
		if runs := runAt(gamma, y); len(runs) != 1 {
			t.Errorf("Γ: %d runs of ink at %.0f, want the stem alone", len(runs), y)
		}
	}
	ups := gs["uni03A5"].Path.Round()
	if runs := runAt(ups, upsilonEnd*p.CapHeight+15); len(runs) != 4 {
		t.Errorf("Υ: %d runs at the ends of the arms, want 4", len(runs))
	}
	y, _ := newRoman(p).get("Y")
	if runs := runAt(y.Round(), upsilonEnd*p.CapHeight+15); len(runs) != 2 {
		t.Errorf("Y: %d runs at the height of the ends of the arms of Υ, want 2", len(runs))
	}
	// The crotch of Υ, the lowest white between the arms, lies inside the
	// range.
	for h := 0.40; h < 0.80; h += 0.005 {
		if len(runAt(ups, h*p.CapHeight)) > 1 {
			if h < 0.457 || h > 0.712 {
				t.Errorf("Υ: the crotch is at %.3f of the cap height, outside the range 0.457 to 0.712", h)
			}
			break
		}
	}
}
