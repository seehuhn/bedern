// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the rules of the signs, the radical, the horizontal constructions and the primes.

func inkArea(p outline.Path) float64 {
	var a float64
	for _, c := range p {
		a += c.SignedArea()
	}
	return math.Abs(a)
}

// TestTransposeAndPerpendicularReadApart requires ⊺ to be 0.65 of the width
// of ⊤ and ⟂ to be centred on the axis where ⊥ stands on the baseline.
func TestTransposeAndPerpendicularReadApart(t *testing.T) {
	for _, s := range derivedSets(1) {
		m := byName(derivedSigns(s))
		width := func(n string) float64 {
			x0, _, x1, _ := m[n].Path.Round().BBox()
			return x1 - x0
		}
		if r := width("uni22BA") / width("uni22A4"); math.Abs(r-transposeWidth) > 0.02 {
			t.Errorf("%s: ⊺ is %.2f of the width of ⊤, want %.2f", s.Name, r, transposeWidth)
		}
		_, by0, _, _ := m["uni22A5"].Path.Round().BBox()
		_, iy0, _, iy1 := m["uni27C2"].Path.Round().BBox()
		if by0 != 0 || math.Abs((iy0+iy1)/2-s.Params.Punct.MathAxis) > 1 {
			t.Errorf("%s: ⊥ stands at %v and ⟂ is centred at %v, want 0 and the axis %v", s.Name, by0, (iy0+iy1)/2, s.Params.Punct.MathAxis)
		}
	}
}

// TestCeilingHooksAreLong requires the arms of ⌈ ⌉ ⌊ ⌋ to reach hookLength
// times as far as those of the bracket.
func TestCeilingHooksAreLong(t *testing.T) {
	for _, s := range derivedSets(1) {
		m := byName(derivedSigns(s))
		bracket, _, _ := glyphs.Outline("bracketleft", s.Params)
		_, by0, _, by1 := bracket.Round().BBox()
		brace := runsOf(bracket.Round(), by1-5)
		stem := runsOf(bracket.Round(), (by0+by1)/2)
		if len(brace) != 1 || len(stem) != 1 {
			t.Fatalf("%s: the bracket has unexpected runs", s.Name)
		}
		reach := (brace[0][1] - brace[0][0]) - (stem[0][1] - stem[0][0])
		for _, n := range []string{"uni2308", "uni230A"} {
			path := m[n].Path.Round()
			_, y0, _, y1 := path.BBox()
			y := y1 - 5
			if n == "uni230A" {
				y = y0 + 5
			}
			arm := runsOf(path, y)
			st := runsOf(path, (y0+y1)/2)
			if len(arm) != 1 || len(st) != 1 {
				t.Fatalf("%s: %s has unexpected runs", s.Name, n)
			}
			got := (arm[0][1] - arm[0][0]) - (st[0][1] - st[0][0])
			if math.Abs(got-hookLength*reach) > 2 {
				t.Errorf("%s: the hook of %s reaches %.0f, want %.0f (%.1f of the bracket's %.0f)", s.Name, n, got, hookLength*reach, hookLength, reach)
			}
		}
	}
}

// TestPrecedesHasAPointedVertex requires the vertex of ≺ and ≻ to be cut
// narrower than that of < and >.
func TestPrecedesHasAPointedVertex(t *testing.T) {
	for _, s := range derivedSets(1) {
		m := byName(derivedSigns(s))
		less, _, _ := glyphs.Outline("less", s.Params)
		flat := func(p outline.Path) float64 {
			x0, _, _, _ := p.Round().BBox()
			r := runsV(p.Round(), x0+1)
			if len(r) != 1 {
				t.Fatalf("%s: the vertex has %d runs", s.Name, len(r))
			}
			return r[0][1] - r[0][0]
		}
		if f := flat(m["uni227A"].Path); f > precPoint+4 {
			t.Errorf("%s: the vertex of ≺ is %.0f wide, want at most %v", s.Name, f, precPoint+4)
		}
		if fp, fl := flat(m["uni227A"].Path), flat(less); fp > 0.7*fl {
			t.Errorf("%s: the vertex of ≺ (%.0f) is not pointed against that of < (%.0f)", s.Name, fp, fl)
		}
	}
}

// TestNegationsLeanAlike requires the solidus of ≠ ≢ ∉ ∌ ⊄ ⊅ ⊉ ≉ to lean at
// the angle of the slash.
func TestNegationsLeanAlike(t *testing.T) {
	for _, s := range derivedSets(1) {
		m := byName(derivedSigns(s))
		for _, n := range []string{"uni2260", "uni2262", "uni2209", "uni220C", "uni2284", "uni2285", "uni2289", "uni2249"} {
			g, ok := m[n]
			if !ok {
				t.Errorf("%s: no glyph %s", s.Name, n)
				continue
			}
			path := g.Path.Round()
			_, y0, _, y1 := path.BBox()
			top, bottom := runsH(path, y1-3), runsH(path, y0+3)
			if len(top) == 0 || len(bottom) == 0 {
				t.Errorf("%s: %s has no ink at its ends", s.Name, n)
				continue
			}
			xt := (top[len(top)-1][0] + top[len(top)-1][1]) / 2
			xb := (bottom[0][0] + bottom[0][1]) / 2
			lean := math.Atan2(xt-xb, (y1-3)-(y0+3)) * 180 / math.Pi
			if math.Abs(lean-s.Params.Punct.SlashAngle) > 1.5 {
				t.Errorf("%s: the solidus of %s leans %.1f degrees, want %.1f", s.Name, n, lean, s.Params.Punct.SlashAngle)
			}
		}
	}
}

// TestRadicalThickStrokeStaysLight requires the thick stroke of the
// largest radical to be at most 1.15 times that of the base
// and the stub of the bar radStub rules long.
func TestRadicalThickStrokeStaysLight(t *testing.T) {
	for _, s := range derivedSets(0) {
		gs, err := radicalGlyphs(s.Params)
		if err != nil {
			t.Fatal(err)
		}
		thick := func(g Glyph) float64 {
			path := g.Path.Round()
			_, y0, _, y1 := path.BBox()
			h := y1 - y0
			ya, yb := y0+0.12*h, y0+0.24*h
			ra, rb := runsH(path, ya), runsH(path, yb)
			if len(ra) == 0 || len(rb) == 0 {
				t.Fatalf("%s: %s has no thick stroke at the foot", s.Name, g.Name)
			}
			slope := ((rb[0][0] + rb[0][1]) - (ra[0][0] + ra[0][1])) / 2 / (yb - ya)
			return (ra[0][1] - ra[0][0]) / math.Hypot(1, slope)
		}
		base := gs[0]
		largest := byName(gs)[base.Variants[len(base.Variants)-1]]
		if r := thick(largest) / thick(base); r > 1.15 {
			t.Errorf("%s: the thick stroke of %s is %.2f times that of the base, want at most 1.15", s.Name, largest.Name, r)
		}
		for _, g := range []Glyph{base, largest} {
			path := g.Path.Round()
			_, _, x1, y1 := path.BBox()
			bar := runsV(path, x1-10)
			rule := bar[len(bar)-1][1] - bar[len(bar)-1][0]
			rs := runsH(path, y1-rule-6)
			if stub := (x1 - rs[len(rs)-1][1]) / rule; math.Abs(stub-radStub) > 0.15 {
				t.Errorf("%s: the stub of %s is %.2f rules, want %.2f", s.Name, g.Name, stub, radStub)
			}
		}
	}
}

// TestCaronKeepsItsV requires the largest caron to be taller than the
// largest circumflex by its larger growth.
func TestCaronKeepsItsV(t *testing.T) {
	var circ, caron float64
	for _, a := range wideAccents {
		switch a.name {
		case "uni0302":
			circ = a.height
		case "uni030C":
			caron = a.height
		}
	}
	if caron <= circ || caron > 1.58 {
		t.Errorf("the height growth of the caron is %v and of the circumflex %v, want the caron larger and at most 1.58", caron, circ)
	}
}

// TestScriptPrimesHaveTheirWeight requires the script primes to
// lean 22 and 20 degrees, to have a stroke of at least 85 and 98 across, and
// to be at most 275 and 325 wide.
func TestScriptPrimesHaveTheirWeight(t *testing.T) {
	for level := 1; level <= 2; level++ {
		r := newRoman(scriptPrimeParams(level))
		path, _ := scriptPrime(r, level, 1)
		path = path.Round()
		x0, y0, x1, y1 := path.BBox()
		mid := runsH(path, (y0+y1)/2)
		if len(mid) != 1 {
			t.Fatalf("level %d: the prime has %d runs", level, len(mid))
		}
		lean := scriptPrimeLean[level-1] * math.Pi / 180
		stroke := (mid[0][1] - mid[0][0]) * math.Cos(lean)
		maxW := []float64{275, 325}[level-1]
		minStroke := []float64{85, 98}[level-1]
		if x1-x0 > maxW {
			t.Errorf("level %d: the prime is %.0f wide, want at most %.0f", level, x1-x0, maxW)
		}
		if stroke < minStroke {
			t.Errorf("level %d: the prime has a stroke of %.0f, want at least %.0f", level, stroke, minStroke)
		}
	}
}

// TestDoubleStruckRAndQAreNotPale requires the ink of ℝ and ℚ over that of R
// and Q to reach 0.85 and 0.83.
func TestDoubleStruckRAndQAreNotPale(t *testing.T) {
	for _, s := range derivedSets(0)[:1] {
		m := dsSet(s)
		for _, c := range []struct {
			name, roman string
			lo          float64
		}{{"uni211D", "R", 0.85}, {"uni211A", "Q", 0.83}} {
			g, ok := m[c.name]
			if !ok {
				t.Fatalf("no %s", c.name)
			}
			rp, _, _ := glyphs.Outline(c.roman, s.Params)
			if r := inkArea(g.Path.Round()) / inkArea(rp.Round()); r < c.lo-0.005 {
				t.Errorf("%s: the ink of %s is %.3f of that of %s, want at least %.2f", s.Name, c.name, r, c.roman, c.lo)
			}
		}
	}
}
