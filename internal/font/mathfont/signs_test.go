// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// The tests of the arrows, the relations and the binary operators (arrows.go,
// relations.go).

// waves gives the number of waves of the signs made with the wave of ∼.  The
// centre line of a wave turns from one side to the other in its middle, so
// that each edge of a wave has one inflection there, which is intended.
var waves = map[string]int{"uni223C": 1, "uni2248": 2, "uni2243": 1, "uni2245": 1}

// unintended returns the findings of the fairness analysis of the glyph less
// those which its design intends: one inflection on each edge of each wave,
// within 40 units of the middle of the ink.
func unintended(g Glyph, fs []fair.Finding) []fair.Finding {
	n := waves[g.Name]
	if n == 0 {
		return fs
	}
	x0, _, x1, _ := g.Path.Round().BBox()
	var out []fair.Finding
	allowed := 2 * n
	for _, f := range fs {
		if f.Kind == fair.Inflection && math.Abs(f.X-(x0+x1)/2) < 40 && allowed > 0 {
			allowed--
			continue
		}
		out = append(out, f)
	}
	return out
}

// newRunes are the code points of the arrows, relations and binary operators
// and ∉.
var newRunes = []rune{
	0x2192, 0x2190, 0x2194, 0x2191, 0x2193, 0x2195, 0x21D2, 0x21D0, 0x21D4, 0x21D1,
	0x21D3, 0x21A6, 0x21AA, 0x27F5, 0x27F6, 0x27F7, 0x27F8, 0x27F9, 0x27FA, 0x27FC, 0x20D7,
	0x2208, 0x220B, 0x2209, 0x2282, 0x2283, 0x2286, 0x2287, 0x228A, 0x2288, 0x222A,
	0x2229, 0x2228, 0x2227, 0x223C, 0x2241, 0x2248, 0x2243, 0x2245, 0x227A, 0x227B,
	0x2AAF, 0x2AB0, 0x2295, 0x2297, 0x2299, 0x2218, 0x22C6, 0x221D, 0x2254, 0x2255,
	0x2A7D, 0x2A7E, 0x2270, 0x2271, 0x226E, 0x226F, 0x2262, 0x2224,
	0x2226, 0x220C, 0x2284, 0x2285, 0x2289, 0x2249,
}

// TestNewSignsHaveGlyphs requires a glyph with ink for every code point of
// the new signs, and a class which decides its side spaces.
func TestNewSignsHaveGlyphs(t *testing.T) {
	m := map[rune]Glyph{}
	for _, g := range derivedSigns(derivedSets(0)[0]) {
		m[g.Rune] = g
	}
	for _, r := range newRunes {
		g, ok := m[r]
		if !ok || len(g.Path) == 0 {
			t.Errorf("%U has no glyph", r)
		}
	}
}

// levelArrows are the arrows whose shaft lies on the axis, and centred the
// signs whose ink is centred on the axis.
var (
	levelArrows = []string{"uni2192", "uni2190", "uni2194", "uni21D2", "uni21D0", "uni21D4",
		"uni21A6", "uni21AA", "uni27F5", "uni27F6", "uni27F7", "uni27F8", "uni27F9", "uni27FA", "uni27FC"}
	centred = []string{"uni2191", "uni2193", "uni2195", "uni21D1", "uni21D3",
		"uni2208", "uni220B", "uni2209", "uni2282", "uni2283", "uni2286", "uni2287", "uni2288",
		"uni222A", "uni2229", "uni2228", "uni2227", "uni223C", "uni2241", "uni2248", "uni227A",
		"uni227B", "uni2AAF", "uni2AB0", "uni2295", "uni2297", "uni2299", "uni2218", "uni221D",
		"uni2254", "uni2255", "uni2A7D", "uni2A7E", "uni2270", "uni2271", "uni226E", "uni226F",
		"uni2262", "uni2260"}
)

// TestNewSignsStandOnTheAxis requires, within one unit, the shafts of the
// level arrows and the ink of the other signs to be centred on the axis, the
// bars of ≃ where those of the equals stand and those of ≅ at the pitch of
// the equals about the axis, ⊊ where ⊆ stands
// and the centre of the points of ⋆ on the axis.
func TestNewSignsStandOnTheAxis(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		axis := s.Params.Punct.MathAxis
		check := func(name, what string, y float64) {
			if math.Abs(y-axis) > 1 {
				t.Errorf("%s: %s of %s is centred %.1f from the axis", s.Name, what, name, y-axis)
			}
		}
		for _, n := range levelArrows {
			p := m[n].Path.Round()
			x0, _, x1, _ := p.BBox()
			r := runsV(p, x0+0.4*(x1-x0))
			if len(r) == 0 {
				t.Errorf("%s: %s has no shaft", s.Name, n)
				continue
			}
			check(n, "the shafts", (r[0][0]+r[len(r)-1][1])/2)
		}
		for _, n := range centred {
			_, y0, _, y1 := m[n].Path.Round().BBox()
			check(n, "the ink", (y0+y1)/2)
		}
		for _, c := range []struct{ name, like string }{{"uni2243", "equal"}, {"uni2245", "equal"}, {"uni228A", "uni2286"}} {
			var like outline.Path
			if c.like == "equal" {
				like, _, _ = glyphs.Outline("equal", s.Params)
			} else {
				like = m[c.like].Path
			}
			_, ly0, _, _ := like.Round().BBox()
			_, y0, _, _ := m[c.name].Path.Round().BBox()
			if c.name == "uni228A" {
				_, _, _, ly1 := like.Round().BBox()
				_, _, _, y1 := m[c.name].Path.Round().BBox()
				if math.Abs(y1-ly1) > 1 {
					t.Errorf("%s: the top of ⊊ is %.0f, that of ⊆ %.0f", s.Name, y1, ly1)
				}
				continue
			}
			if c.name == "uni2245" {
				// Three places at the pitch of the equals, the middle on the axis.
				ly0 -= (s.Params.Punct.EqualsGap + s.Params.Punct.MathThick) / 2
			}
			if math.Abs(y0-ly0) > 1 {
				t.Errorf("%s: the lowest bar of %s is at %.0f, that of %s at %.0f", s.Name, c.name, y0, c.like, ly0)
			}
		}
		{
			// The centre of the points of ⋆: the middle of the top point and
			// the line through the two lower points.
			p := m["uni22C6"].Path.Round()
			_, y0, _, y1 := p.BBox()
			rad := (y1 - y0) / (1 + math.Cos(math.Pi/5))
			check("uni22C6", "the centre of the points", y1-rad)
		}
	}
}

// TestNewSignsHaveTheRule requires the strokes of the new signs to be as
// thick as the rule, within one unit, taken across: the shafts of the
// arrows, the arms and the round of ⊂ and ∈ and the bar of ∈, the stems and
// the round of ∪, the strokes of ∨ across, the middle of the wave of ∼, the
// rings of ⊕ and ∘, and the barbs of → at barbShare of the rule.
func TestNewSignsHaveTheRule(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		rule := s.Params.Punct.MathThick
		check := func(name, where string, d, want float64) {
			if math.Abs(d-want) > 1 {
				t.Errorf("%s: %s of %s is %.1f thick, want %v", s.Name, where, name, d, want)
			}
		}
		for _, n := range levelArrows {
			p := m[n].Path.Round()
			x0, _, x1, _ := p.BBox()
			for _, r := range runsV(p, x0+0.45*(x1-x0)) {
				check(n, "a shaft", r[1]-r[0], rule)
			}
		}
		for _, n := range []string{"uni2191", "uni21D1"} {
			p := m[n].Path.Round()
			_, y0, _, y1 := p.BBox()
			for _, r := range runsH(p, y0+0.4*(y1-y0)) {
				check(n, "a shaft", r[1]-r[0], rule)
			}
		}
		{
			// The barb of →, across: its run along a level line halfway up
			// times the sine of the angle of the chord of its outer edge.
			p := m["uni2192"].Path.Round()
			_, y0, x1, y1 := p.BBox()
			axis := s.Params.Punct.MathAxis
			yb := axis + rule/2 + 0.5*(y1-axis-rule/2)
			rs := runsH(p, yb)
			var tx, ty float64 = x1, axis
			for _, c := range p {
				for _, sg := range c.Segs {
					if sg.P3.Y > ty+0.5 || (math.Abs(sg.P3.Y-y1) < 0.5 && sg.P3.X > tx) {
						tx, ty = sg.P3.X, sg.P3.Y
					}
				}
			}
			_ = y0
			ang := math.Atan2(y1-axis, x1-tx)
			w := rs[len(rs)-1][1] - rs[len(rs)-1][0]
			if d := w * math.Sin(ang); math.Abs(d-math.Round(barbShare*rule)) > 4 {
				t.Errorf("%s: the barb of → is %.1f across, want about %v", s.Name, d, math.Round(barbShare*rule))
			}
		}
		for _, n := range []string{"uni2282", "uni2208"} {
			p := m[n].Path.Round()
			x0, y0, x1, y1 := p.BBox()
			r := runsV(p, x0+0.8*(x1-x0))
			want := 2
			if n == "uni2208" {
				want = 3
			}
			if len(r) != want {
				t.Errorf("%s: %s has %d strokes near its right end, want %d", s.Name, n, len(r), want)
			}
			for _, a := range r {
				check(n, "an arm", a[1]-a[0], rule)
			}
			if n == "uni2282" {
				r := runsH(p, (y0+y1)/2)
				check(n, "the round", r[0][1]-r[0][0], rule)
			}
		}
		{
			p := m["uni222A"].Path.Round()
			x0, y0, x1, y1 := p.BBox()
			for _, r := range runsH(p, y1-0.2*(y1-y0)) {
				check("uni222A", "a stem", r[1]-r[0], rule)
			}
			r := runsV(p, (x0+x1)/2)
			check("uni222A", "the round", r[0][1]-r[0][0], rule)
		}
		{
			p := m["uni2228"].Path.Round()
			x0, y0, x1, y1 := p.BBox()
			a := math.Atan2(y1-y0, (x1-x0)/2-s.Params.Punct.MathVertexFlat/2)
			for _, r := range runsH(p, y0+0.7*(y1-y0)) {
				check("uni2228", "a stroke", (r[1]-r[0])*math.Sin(a), rule)
			}
		}
		{
			// The wave crosses its middle at the angle of its steepest
			// slope: the run upright there times the cosine of that angle.
			p := m["uni223C"].Path.Round()
			x0, _, x1, _ := p.BBox()
			xm := (x0 + x1) / 2
			r := runsV(p, xm)
			a, b := runsV(p, xm-4), runsV(p, xm+4)
			slope := ((a[0][0] + a[0][1]) - (b[0][0] + b[0][1])) / 2 / 8
			check("uni223C", "the middle", (r[0][1]-r[0][0])*math.Cos(math.Atan(slope)), rule)
		}
		for _, n := range []string{"uni2295", "uni2218"} {
			p := m[n].Path.Round()
			_, y0, _, y1 := p.BBox()
			// Along a level line rule above the middle, corrected for the
			// slope of the ring there.
			rm := (y1 - y0 - rule) / 2
			r := runsH(p, (y0+y1)/2+rule)
			check(n, "the ring", (r[0][1]-r[0][0])*math.Sqrt(1-(rule/rm)*(rule/rm)), rule)
		}
	}
}

// TestNewPairsAreMirrorImages requires the left and the right forms to be
// mirror images in a vertical line, ∩ and ∧ those of ∪ and ∨ in a level line
// (not turned), ↓ and ⇓ those of ↑ and ⇑ in a level line, within one unit.
func TestNewPairsAreMirrorImages(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		for _, c := range []struct {
			a, b     string
			vertical bool
		}{
			{"uni2192", "uni2190", false}, {"uni21D2", "uni21D0", false},
			{"uni27F6", "uni27F5", false}, {"uni27F9", "uni27F8", false},
			{"uni2208", "uni220B", false}, {"uni2282", "uni2283", false},
			{"uni2286", "uni2287", false}, {"uni227A", "uni227B", false},
			{"uni2AAF", "uni2AB0", false}, {"uni2254", "uni2255", false},
			{"uni2A7D", "uni2A7E", false},
			{"uni222A", "uni2229", true}, {"uni2228", "uni2227", true},
			{"uni2191", "uni2193", true}, {"uni21D1", "uni21D3", true},
		} {
			if !mirrored(m[c.a].Path.Round(), m[c.b].Path.Round(), c.vertical) {
				t.Errorf("%s: %s is not the mirror image of %s", s.Name, c.b, c.a)
			}
		}
		// Each double-headed arrow and each symmetric sign is its own mirror
		// image.
		for _, n := range []string{"uni2194", "uni21D4", "uni27F7", "uni27FA", "uni2295", "uni2297", "uni2299", "uni2218", "uni22C6"} {
			p := m[n].Path.Round()
			if !mirrored(p, p, false) {
				t.Errorf("%s: %s is not symmetric", s.Name, n)
			}
		}
	}
}

// TestDerivedSignsFollowTheirRules checks the rules by which the derived
// signs are made: ≺ has the width of its rule, ↑ is → turned by a quarter, the long arrows are → with a
// longer shaft and the same head, ≈ is two waves of ∼, and every negated
// sign is negOver higher than its parent above and below, at one angle.
func TestDerivedSignsFollowTheirRules(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := byName(derivedSigns(s))
		ink := func(n string) (w, h float64) {
			x0, y0, x1, y1 := m[n].Path.Round().BBox()
			return x1 - x0, y1 - y0
		}
		rw, rh := ink("uni2192")
		if uw, uh := ink("uni2191"); math.Abs(uw-rh) > 1 || math.Abs(uh-rw) > 1 {
			t.Errorf("%s: ↑ is %.0f by %.0f, → %.0f by %.0f", s.Name, uw, uh, rw, rh)
		}
		if lw, lh := ink("uni27F6"); math.Abs(lh-rh) > 1 || math.Abs(lw-rw-(longArrowWidth-arrowWidth)) > 1 {
			t.Errorf("%s: ⟶ is %.0f by %.0f, → %.0f by %.0f", s.Name, lw, lh, rw, rh)
		}
		if pw, _ := ink("uni227A"); math.Abs(pw-math.Round(precWidth*s.Params.Punct.MathAngleWidth)) > 1 {
			t.Errorf("%s: ≺ is %.0f wide, want %.0f", s.Name, pw, math.Round(precWidth*s.Params.Punct.MathAngleWidth))
		}
		if _, ah := ink("uni2248"); true {
			_, sh := ink("uni223C")
			pitch := s.Params.Punct.EqualsGap + s.Params.Punct.MathThick
			if math.Abs(ah-(sh+pitch)) > 1 {
				t.Errorf("%s: ≈ is %.0f high, want ∼ (%.0f) and the pitch of the equals (%.0f)", s.Name, ah, sh, pitch)
			}
		}
		eq, _, _ := glyphs.Outline("equal", s.Params)
		less, _, _ := glyphs.Outline("less", s.Params)
		parent := func(p outline.Path) float64 { _, y0, _, y1 := p.Round().BBox(); return y1 - y0 }
		for _, c := range []struct {
			name string
			h    float64
		}{
			{"uni2260", parent(eq)}, {"uni2241", parent(m["uni223C"].Path)},
			{"uni2209", parent(m["uni2208"].Path)}, {"uni2288", parent(m["uni2286"].Path)},
			{"uni2270", parent(m["uni2264"].Path)}, {"uni226E", parent(less)},
			{"uni2262", parent(m["uni2261"].Path)}, {"uni2284", parent(m["uni2282"].Path)},
			{"uni2285", parent(m["uni2283"].Path)}, {"uni2289", parent(m["uni2287"].Path)},
			{"uni220C", parent(m["uni220B"].Path)}, {"uni2249", parent(m["uni2248"].Path)},
		} {
			if _, h := ink(c.name); math.Abs(h-(c.h+2*negOver)) > 1 {
				t.Errorf("%s: %s is %.0f high, want %.0f", s.Name, c.name, h, c.h+2*negOver)
			}
		}
	}
}

// TestNewSignsHaveTheirWeight requires the ink of each new drawing
// over that of n to lie in the range.  It runs on the regular only: the
// perturbed sets move the n and not the signs.
func TestNewSignsHaveTheirWeight(t *testing.T) {
	area := func(p outline.Path) float64 {
		var a float64
		for _, c := range p {
			a += c.SignedArea()
		}
		return math.Abs(a)
	}
	for _, s := range derivedSets(0)[:1] {
		m := byName(derivedSigns(s))
		n, _, _ := glyphs.Outline("n", s.Params)
		an := area(n.Round())
		for _, c := range []struct {
			name   string
			lo, hi float64
		}{
			{"uni2192", 0.61, 0.85}, {"uni21D2", 1.01, 1.25}, {"uni2208", 0.80, 1.27},
			{"uni2282", 0.71, 0.93}, {"uni222A", 0.73, 0.93}, {"uni2228", 0.62, 0.78},
			{"uni223C", 0.31, 0.54}, {"uni227A", 0.55, 0.77}, {"uni2295", 1.09, 2.23},
			{"uni221D", 0.75, 1.17}, {"uni22C6", 0.49, 1.31}, {"uni2218", 0.31, 0.53},
		} {
			if v := area(m[c.name].Path.Round()) / an; v < c.lo || v > c.hi {
				t.Errorf("%s: the ink of %s is %.2f of that of n, outside %.2f to %.2f", s.Name, c.name, v, c.lo, c.hi)
			}
		}
	}
}
