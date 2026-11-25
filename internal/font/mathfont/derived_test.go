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
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// derivedSets are the parameter sets on which the derived signs are tested:
// the regular alone in the routine tier, and the regular and every
// perturbed set in the full tier.
func derivedSets(n int) []designtest.Set {
	if testing.Short() {
		n = 0
	}
	return append([]designtest.Set{{Name: "regular", Params: design.Regular()}}, designtest.Sets(testing.Short(), n)...)
}

// derivedOf holds the derived signs by the name of their parameter set, each
// built once.
var derivedOf sync.Map

// derivedSigns returns the derived signs of the set.
func derivedSigns(s designtest.Set) []Glyph {
	v, _ := derivedOf.LoadOrStore(s.Name, sync.OnceValue(func() []Glyph { return Derived(s.Params) }))
	return v.(func() []Glyph)()
}

// warm calls build for each of the sets at the same time and fills the
// caches of the builds.
func warm(sets []designtest.Set, build func(designtest.Set)) {
	var wg sync.WaitGroup
	for _, s := range sets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			build(s)
		}()
	}
	wg.Wait()
}

func byName(gs []Glyph) map[string]Glyph {
	m := map[string]Glyph{}
	for _, g := range gs {
		m[g.Name] = g
	}
	return m
}

// TestDerivedSignsPassTheChecks applies the structural checks of the regular
// (with the bounding-box rule of each sign's parent) and the fairness
// analysis to every derived sign, unrounded and rounded; a derived sign has no
// finding of the fairness analysis.  The routine tier runs the regular and
// the first perturbed set, which moves the proportions; the structural checks
// take most of the time of the tests of the package.
func TestDerivedSignsPassTheChecks(t *testing.T) {
	for _, s := range derivedSets(1) {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			for _, g := range derivedSigns(s) {
				if italicPen[g.Name] {
					continue // checked as the italic is (TestItalicSymbolsAreSmooth)
				}
				path := g.Path
				if g.Mirror {
					path = mirroredX(path, g.Advance/2)
				}
				if err := glyphs.CheckOutline(s.Params, g.Like, path, g.Advance); err != nil {
					t.Errorf("%s: %v", g.Name, err)
				}
				if fs := unintended(g, fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings); len(fs) > 0 {
					t.Errorf("%s: %d findings of the fairness analysis, the first %v", g.Name, len(fs), fs[0])
				}
			}
		})
	}
}

// onAxis are the derived signs whose ink is centred on the axis.
var onAxis = []string{
	"uni2236", "uni2217", "uni2219", "uni22C5", "uni22EF", "uni22EE", "uni22F1", "uni22F0",
	"uni2264", "uni2265", "uni226A", "uni226B", "uni2260", "uni2261", "uni2216",
}

// TestDerivedSignsStandOnTheAxis requires the ink of the signs of onAxis to
// be centred on the axis within one unit, rounded.
func TestDerivedSignsStandOnTheAxis(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		for _, n := range onAxis {
			_, y0, _, y1 := m[n].Path.Round().BBox()
			if d := (y0+y1)/2 - s.Params.Punct.MathAxis; math.Abs(d) > 1 {
				t.Errorf("%s: %s is centred %.1f from the axis", s.Name, n, d)
			}
		}
	}
}

// runs returns the runs of ink along the vertical line at x.
func runsV(p outline.Path, x float64) [][2]float64 {
	ys := p.CrossingsY(x)
	slices.Sort(ys)
	var out [][2]float64
	for i := 0; i+1 < len(ys); i += 2 {
		out = append(out, [2]float64{ys[i], ys[i+1]})
	}
	return out
}

// runsH returns the runs of ink along the horizontal line at y.
func runsH(p outline.Path, y float64) [][2]float64 {
	xs := p.CrossingsX(y)
	slices.Sort(xs)
	var out [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		out = append(out, [2]float64{xs[i], xs[i+1]})
	}
	return out
}

// TestDerivedSignsHaveTheRule requires the bars of the derived relations and
// symbols to be as thick as the rule of the signs, within one unit: the bar of
// ≤ and ≥, the three bars of ≡, both bars of ⊥, ⊤, ⊺ and ⟂, the four sides
// of □ and the arms of ∃, measured across, rounded.
func TestDerivedSignsHaveTheRule(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		rule := s.Params.Punct.MathThick
		check := func(name, where string, r [2]float64) {
			if d := r[1] - r[0]; math.Abs(d-rule) > 1 {
				t.Errorf("%s: %s of %s is %.1f thick, want %v", s.Name, where, name, d, rule)
			}
		}
		for _, n := range []string{"uni2264", "uni2265"} {
			p := m[n].Path.Round()
			x0, _, x1, _ := p.BBox()
			r := runsV(p, (x0+x1)/2)
			check(n, "the bar", r[0])
		}
		{
			p := m["uni2261"].Path.Round()
			x0, _, x1, _ := p.BBox()
			r := runsV(p, (x0+x1)/2)
			if len(r) != 3 {
				t.Errorf("%s: ≡ has %d bars", s.Name, len(r))
			}
			for _, b := range r {
				check("uni2261", "a bar", b)
			}
		}
		for _, n := range []string{"uni22A5", "uni22A4", "uni22BA", "uni27C2"} {
			p := m[n].Path.Round()
			x0, y0, x1, y1 := p.BBox()
			check(n, "the horizontal bar", runsV(p, x0+0.1*(x1-x0))[0])
			r := runsH(p, (y0+y1)/2)
			if len(r) != 1 {
				t.Errorf("%s: %s has %d strokes at half its height", s.Name, n, len(r))
				continue
			}
			check(n, "the vertical bar", r[0])
		}
		{
			p := m["uni25A1"].Path.Round()
			x0, y0, x1, y1 := p.BBox()
			for _, r := range append(runsV(p, (x0+x1)/2), runsH(p, (y0+y1)/2)...) {
				check("uni25A1", "a side", r)
			}
		}
		{
			p := m["uni2203"].Path.Round()
			x0, _, x1, _ := p.BBox()
			r := runsV(p, x0+0.4*(x1-x0))
			if len(r) != 3 {
				t.Errorf("%s: ∃ has %d arms", s.Name, len(r))
			}
			for _, a := range r {
				check("uni2203", "an arm", a)
			}
		}
	}
}

// mirrorRuns compares the runs of ink of a along horizontal (or vertical)
// lines with those of b mirrored in the middle of the ink of a, within one
// unit.
func mirrored(a, b outline.Path, vertical bool) bool {
	ax0, ay0, ax1, ay1 := a.BBox()
	bx0, by0, bx1, by1 := b.BBox()
	if math.Abs((ax1-ax0)-(bx1-bx0)) > 1 || math.Abs((ay1-ay0)-(by1-by0)) > 1 {
		return false
	}
	for i := 1; i < 20; i++ {
		var ra, rb [][2]float64
		var lo, hi float64
		if vertical {
			f := ax0 + (ax1-ax0)*float64(i)/20
			ra, rb = runsV(a, f), runsV(b, bx0+(f-ax0))
			lo, hi = ay0+by1, 0
		} else {
			f := ay0 + (ay1-ay0)*float64(i)/20
			ra, rb = runsH(a, f), runsH(b, by0+(f-ay0))
			lo, hi = ax0+bx1, 0
		}
		_ = hi
		if len(ra) != len(rb) {
			return false
		}
		for k := range ra {
			m := rb[len(rb)-1-k]
			if math.Abs(ra[k][0]-(lo-m[1])) > 1 || math.Abs(ra[k][1]-(lo-m[0])) > 1 {
				return false
			}
		}
	}
	return true
}

// TestDerivedPairsAreMirrorImages requires ≥, ≫, ⌉, ⌋ and ⋰ to be the mirror
// images of ≤, ≪, ⌈, ⌊ and ⋱ in a vertical line, ⊤ and ⌊ those of ⊥ and ⌈
// in a horizontal line, and ∓ to be ± turned, within one unit.
func TestDerivedPairsAreMirrorImages(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		pm, _, _ := glyphs.Outline("plusminus", s.Params)
		for _, c := range []struct {
			a, b     string
			vertical bool
		}{
			{"uni2264", "uni2265", false}, {"uni226A", "uni226B", false},
			{"uni2308", "uni2309", false}, {"uni230A", "uni230B", false},
			{"uni22F1", "uni22F0", false},
			{"uni22A5", "uni22A4", true}, {"uni2308", "uni230A", true},
		} {
			if !mirrored(m[c.a].Path.Round(), m[c.b].Path.Round(), c.vertical) {
				t.Errorf("%s: %s is not the mirror image of %s", s.Name, c.b, c.a)
			}
		}
		// ∓ is ± mirrored in a vertical line and then in a horizontal one.
		if !mirrored(mirroredX(pm, 0).Round(), m["uni2213"].Path.Round(), true) {
			t.Errorf("%s: ∓ is not ± turned", s.Name)
		}
	}
}
