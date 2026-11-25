// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// operatorsOf holds the big operators, the integral and the radical by the
// name of their parameter set, each built once.
var operatorsOf sync.Map

func operators(t *testing.T, s designtest.Set) map[string]Glyph {
	t.Helper()
	v, _ := operatorsOf.LoadOrStore(s.Name, sync.OnceValues(func() ([]Glyph, error) { return operatorGlyphs(s.Params) }))
	gs, err := v.(func() ([]Glyph, error))()
	if err != nil {
		t.Fatal(err)
	}
	return byName(gs)
}

// acrossAt returns the thickness across a straight slanted stroke: the run
// of ink which pick chooses at the height y, times the sine of the angle of
// the line through the middles of the runs dy below and above it.
func acrossAt(p outline.Path, y, dy float64, pick func([][2]float64) [2]float64) float64 {
	a, b, m := runsH(p, y-dy), runsH(p, y+dy), runsH(p, y)
	if len(a) == 0 || len(b) == 0 || len(m) == 0 {
		return 0
	}
	ra, rb, rm := pick(a), pick(b), pick(m)
	dx := (rb[0]+rb[1])/2 - (ra[0]+ra[1])/2
	return (rm[1] - rm[0]) * 2 * dy / math.Hypot(dx, 2*dy)
}

func first(r [][2]float64) [2]float64 { return r[0] }
func last(r [][2]float64) [2]float64  { return r[len(r)-1] }

// TestBigOperatorSizes requires the text and display sizes of ∑, ∏ and ∐
// and of ∫ to have their heights and to be centred on the axis within one
// unit, the display sizes to reach DisplayOperatorMinHeight and the text
// sizes not, and ∐ to be ∏ mirrored in the axis.
func TestBigOperatorSizes(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := operators(t, s)
		p := s.Params
		min := float64(Constants(p).DisplayOperatorMinHeight)
		for _, c := range []struct {
			name      string
			text, dsp float64
		}{
			{"uni2211", opTextHeight, opDisplayHeight},
			{"uni220F", opTextHeight, opDisplayHeight},
			{"uni2210", opTextHeight, opDisplayHeight},
			{"uni222B", intTextHeight, intDisplayHeight},
		} {
			for _, sz := range []struct {
				name string
				h    float64
			}{{c.name, c.text}, {c.name + ".d", c.dsp}} {
				_, y0, _, y1 := m[sz.name].Path.Round().BBox()
				if math.Abs(y1-y0-sz.h) > 1 {
					t.Errorf("%s: %s is %v high, want %v", s.Name, sz.name, y1-y0, sz.h)
				}
				if d := (y0+y1)/2 - p.Punct.MathAxis; math.Abs(d) > 1 {
					t.Errorf("%s: %s is centred %v from the axis", s.Name, sz.name, d)
				}
			}
			if c.dsp < min || c.text >= min {
				t.Errorf("%s: %s has %v and %v, DisplayOperatorMinHeight is %v", s.Name, c.name, c.text, c.dsp, min)
			}
		}
		for _, n := range []string{"uni220F", "uni220F.d"} {
			a := m[n].Path.Round()
			b := mirroredY(m[strings.Replace(n, "220F", "2210", 1)].Path, p.Punct.MathAxis).Round()
			for i := 1; i < 20; i++ {
				_, y0, _, y1 := a.BBox()
				y := y0 + (y1-y0)*float64(i)/20
				ra, rb := runsH(a, y), runsH(b, y)
				if len(ra) != len(rb) {
					t.Errorf("%s: ∐ is not ∏ mirrored at %v", s.Name, y)
					break
				}
			}
		}
	}
}

// TestBigOperatorWeights requires the stems of ∏ and the thick diagonal of
// ∑ to be the capital stem times (h / CapHeight) to the power opPower, and
// the stem of ∫ intWeight of that, within one unit, taken across; the
// edges of the slanted stem of ∫ lie on whole-unit points of the lattice of
// the slant, which moves each by up to half a unit, so it has one and a half.
func TestBigOperatorWeights(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := operators(t, s)
		p := s.Params
		law := func(h float64) float64 { return p.Caps.Stem * math.Pow(h/p.CapHeight, opPower) }
		for _, c := range []struct {
			name string
			h    float64
		}{{"uni220F", opTextHeight}, {"uni220F.d", opDisplayHeight}} {
			g := m[c.name].Path.Round()
			r := runsH(g, p.Punct.MathAxis)
			if len(r) != 2 {
				t.Fatalf("%s: %s has %d runs at the axis", s.Name, c.name, len(r))
			}
			for _, x := range r {
				if got, want := x[1]-x[0], law(c.h); math.Abs(got-want) > 1 {
					t.Errorf("%s: a stem of %s is %.1f, want %.1f", s.Name, c.name, got, want)
				}
			}
		}
		for _, c := range []struct {
			name string
			h    float64
		}{{"uni2211", opTextHeight}, {"uni2211.d", opDisplayHeight}} {
			g := m[c.name].Path.Round()
			y := p.Punct.MathAxis + 0.25*c.h
			if got, want := acrossAt(g, y, 0.05*c.h, first), law(c.h); math.Abs(got-want) > 1 {
				t.Errorf("%s: the thick diagonal of %s is %.1f, want %.1f", s.Name, c.name, got, want)
			}
		}
		for _, c := range []struct {
			name string
			h    float64
		}{{"uni222B", intTextHeight}, {"uni222B.d", intDisplayHeight}} {
			g := m[c.name].Path.Round()
			got := acrossAt(g, p.Punct.MathAxis-30, 20, first)
			if want := intWeight * law(c.h); math.Abs(got-want) > 1.5 {
				t.Errorf("%s: the stem of %s is %.1f, want %.1f", s.Name, c.name, got, want)
			}
		}
	}
}

// TestIntegralSlant requires the centre line of the stem of the integral to
// lean by 1/8 within 0.005 near the axis, and its italics correction to be
// positive and to leave the ink inside the advance.
func TestIntegralSlant(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := operators(t, s)
		p := s.Params
		for _, n := range []string{"uni222B", "uni222B.d"} {
			g := m[n]
			path := g.Path.Round()
			a, b := runsH(path, p.Punct.MathAxis-40), runsH(path, p.Punct.MathAxis+40)
			slope := ((b[0][0]+b[0][1])/2 - (a[0][0]+a[0][1])/2) / 80
			if math.Abs(slope-1.0/8) > 0.005 {
				t.Errorf("%s: the stem of %s leans by %.4f, want 1/8", s.Name, n, slope)
			}
			_, y0, x1, y1 := path.BBox()
			if g.Italic <= 0 || g.Advance+g.Italic < x1 {
				t.Errorf("%s: %s has the italics correction %v; its ink reaches %v beyond the advance", s.Name, n, g.Italic, x1-g.Advance)
			}
			_ = y0
			_ = y1
		}
	}
}

// TestRadicalSizes requires every size of the radical to have its height,
// its bar RadicalRuleThickness thick at the top of the glyph and its ink
// radOverlap beyond the advance, and its long stroke the rule grown by
// radLongPower, across; and the parts of its assembly to stand on the baseline with
// their advances, with connectors of at least MinConnectorOverlap on the
// upright stroke, which the extender is.
func TestRadicalSizes(t *testing.T) {
	for _, s := range derivedSets(3) {
		m := operators(t, s)
		p := s.Params
		rule := float64(Constants(p).RadicalRuleThickness.V)
		base := m["uni221A"]
		names := append([]string{"uni221A"}, base.Variants...)
		heights := append([]float64{radBaseHeight}, delimSizes...)
		for i, n := range names {
			g := m[n]
			path := g.Path.Round()
			x0, y0, x1, y1 := path.BBox()
			if math.Abs(y1-y0-heights[i]) > 1 {
				t.Errorf("%s: %s is %v high, want %v", s.Name, n, y1-y0, heights[i])
			}
			col := runsV(path, x1-5)
			if len(col) == 0 || math.Abs(col[len(col)-1][1]-y1) > 0.5 || math.Abs(col[len(col)-1][1]-col[len(col)-1][0]-rule) > 0.5 {
				t.Errorf("%s: the bar of %s is not %v thick at the top: %v", s.Name, n, rule, col)
			}
			if d := x1 - g.Advance; math.Abs(d-radOverlap) > 1 {
				t.Errorf("%s: the bar of %s reaches %v beyond the advance, want %v", s.Name, n, d, radOverlap)
			}
			want := rule * math.Pow(heights[i]/radBaseHeight, radLongPower)
			if got := acrossAt(path, y0+0.6*(y1-y0), 0.05*(y1-y0), last); math.Abs(got-want) > 1 {
				t.Errorf("%s: the long stroke of %s is %.1f across, want %.1f", s.Name, n, got, want)
			}
			if x0 < 0 {
				t.Errorf("%s: %s starts left of the origin", s.Name, n)
			}
		}
		overlap := math.Round(p.Punct.MathThick / 2)
		var ext outline.Path
		for _, pt := range base.Parts {
			g := m[pt.Name].Path.Round()
			_, y0, _, y1 := g.BBox()
			if y0 != 0 || math.Abs(y1-pt.FullAdvance) > 0.5 {
				t.Errorf("%s: %s reaches from %v to %v, want 0 to %v", s.Name, pt.Name, y0, y1, pt.FullAdvance)
			}
			for _, c := range []float64{pt.StartConn, pt.EndConn} {
				if c != 0 && c < overlap {
					t.Errorf("%s: a connector of %s is %v", s.Name, pt.Name, c)
				}
			}
			if pt.Extender {
				ext = g
			}
		}
		top := m["uni221A.top"].Path.Round()
		a, b := runsH(ext, 1), runsH(top, 1)
		c := runsH(top, connector-1)
		if len(a) != 1 || len(b) != 1 || len(c) != 1 || a[0] != b[0] || b[0] != c[0] {
			t.Errorf("%s: the extender of the radical is not the upright stroke of its top part", s.Name)
		}
		want := rule * math.Pow(radAssemblyHeight/radBaseHeight, radLongPower)
		if w := a[0][1] - a[0][0]; math.Abs(w-want) > 1 {
			t.Errorf("%s: the upright stroke of the radical is %v, want %.1f", s.Name, w, want)
		}
	}
}

// TestAngleBracketSizes requires the strokes of every size of the angle
// brackets to be the rule grown by the law of the delimiters, taken
// across, within one unit, and their width to grow as the square root of
// the height within two units.
func TestAngleBracketSizes(t *testing.T) {
	for _, s := range derivedSets(3) {
		p := s.Params
		m, variants, _ := delimOf(t, p)
		r := newRoman(p)
		basePath, _ := angleBracket(false)(r)
		bx0, _, bx1, _ := basePath.Round().BBox()
		for _, base := range []string{"uni27E8", "uni27E9"} {
			for k, n := range variants[base] {
				h := delimSizes[k]
				g := m[n].Path.Round()
				x0, y0, x1, y1 := g.BBox()
				pick := first
				if base == "uni27E9" {
					pick = last
				}
				want := p.Punct.MathThick * grow(p, h)
				if got := acrossAt(g, y0+0.75*(y1-y0), 0.05*(y1-y0), pick); math.Abs(got-want) > 1 {
					t.Errorf("%s: the stroke of %s is %.1f across, want %.1f", s.Name, n, got, want)
				}
				if want := (bx1 - bx0) * math.Sqrt(h/baseHeight(p)); math.Abs(x1-x0-want) > 2 {
					t.Errorf("%s: %s is %v wide, want %.1f", s.Name, n, x1-x0, want)
				}
			}
		}
	}
}

// TestBraceAssemblies requires the parts of the assembly of each brace to
// stand on the baseline with their advances, every connector to be at least
// MinConnectorOverlap and to lie on a straight upright stroke, the extender
// to be that stroke, the point of the middle piece to lie at its middle, so
// that it stands on the axis, and the largest size to be the bottom part,
// the middle piece and the top part joined with the full connectors, within
// one unit.
func TestBraceAssemblies(t *testing.T) {
	for _, s := range derivedSets(1) {
		p := s.Params
		m, variants, parts := delimOf(t, p)
		overlap := math.Round(p.Punct.MathThick / 2)
		for _, base := range []string{"braceleft", "braceright"} {
			ps := parts[base]
			if len(ps) != 5 {
				t.Fatalf("%s: %s has %d parts, want 5", s.Name, base, len(ps))
			}
			straight := func(g outline.Path, lo, hi float64) bool {
				a, b := runsH(g, lo+0.5), runsH(g, hi-0.5)
				if len(a) != len(b) {
					return false
				}
				for k := range a {
					if math.Abs(a[k][0]-b[k][0]) > 0.5 || math.Abs(a[k][1]-b[k][1]) > 0.5 {
						return false
					}
				}
				return true
			}
			for _, pt := range ps {
				g := m[pt.Name].Path.Round()
				_, y0, _, y1 := g.BBox()
				if y0 != 0 || math.Abs(y1-pt.FullAdvance) > 0.5 {
					t.Errorf("%s: %s reaches from %v to %v, want 0 to %v", s.Name, pt.Name, y0, y1, pt.FullAdvance)
				}
				if pt.StartConn != 0 && (pt.StartConn < overlap || !straight(g, 0, pt.StartConn)) {
					t.Errorf("%s: the start connector of %s is short or not straight", s.Name, pt.Name)
				}
				if pt.EndConn != 0 && (pt.EndConn < overlap || !straight(g, pt.FullAdvance-pt.EndConn, pt.FullAdvance)) {
					t.Errorf("%s: the end connector of %s is short or not straight", s.Name, pt.Name)
				}
			}
			bot, ext, mid, top := m[base+".bot"].Path.Round(), m[base+".ext"].Path.Round(), m[base+".mid"].Path.Round(), m[base+".top"].Path.Round()
			if a, b := runsH(ext, 1), runsH(top, 1); len(a) != 1 || len(b) != 1 || math.Abs(a[0][0]-b[0][0]) > 0.5 || math.Abs(a[0][1]-b[0][1]) > 0.5 {
				t.Errorf("%s: the extender of %s is not the arm of its top part", s.Name, base)
			}
			x0, _, x1, hm := mid.BBox()
			point := x0
			if base == "braceright" {
				point = x1
			}
			at := math.Inf(1)
			for y := 0.0; y <= hm; y++ {
				r := runsH(mid, y)
				if len(r) == 0 {
					continue
				}
				if e := r[0][0]; base == "braceleft" && math.Abs(e-point) < 0.5 || base == "braceright" && math.Abs(r[len(r)-1][1]-point) < 0.5 {
					at = math.Min(at, math.Abs(y-hm/2))
				}
			}
			if at > 1 {
				t.Errorf("%s: the point of %s.mid lies %v from its middle", s.Name, base, at)
			}
			vs := variants[base]
			largest := m[vs[len(vs)-1]].Path.Round()
			_, ly0, _, ly1 := largest.BBox()
			_, _, _, hb := bot.BBox()
			joined := join(shifted(bot, 0, ly0), shifted(mid, 0, ly0+hb-connector), shifted(top, 0, ly0+hb+hm-2*connector))
			cuts := []float64{ly0 + hb - connector, ly0 + hb + hm - 2*connector}
			for i := 1; i < 60; i++ {
				y := ly0 + (ly1-ly0)*float64(i)/60
				if y > cuts[0]-1 && y < cuts[0]+connector+1 || y > cuts[1]-1 && y < cuts[1]+connector+1 {
					continue
				}
				a, b := runsH(largest, y), mergedRuns(runsH(joined, y))
				if len(a) != len(b) {
					t.Errorf("%s: at %v the largest %s has %d runs, the joined parts %d", s.Name, y, base, len(a), len(b))
					continue
				}
				for k := range a {
					if math.Abs(a[k][0]-b[k][0]) > 1 || math.Abs(a[k][1]-b[k][1]) > 1 {
						t.Errorf("%s: at %v the largest %s differs from its joined parts", s.Name, y, base)
					}
				}
			}
		}
	}
}

// TestOperatorsPassTheChecks applies the structural checks of the regular,
// with the bounding-box rule of the glyph named by Like, and the fairness
// analysis to the big operators, the integral, the radical and their sizes
// and parts, on the regular parameters; the integral has the two inflections
// of the S-curve of its stem, near its middle, as the brace has those of its
// S-curves.  The bar of the radical reaches
// radOverlap beyond the advance, so the radical is checked with an advance
// that much wider.
func TestOperatorsPassTheChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("the operators are checked in the full tier")
	}
	p := design.Regular()
	gs, err := operatorGlyphs(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range gs {
		adv := g.Advance
		if strings.HasPrefix(g.Name, "uni221A") {
			adv += radOverlap
		}
		if err := glyphs.CheckOutline(p, g.Like, g.Path, adv); err != nil {
			t.Errorf("%s: %v", g.Name, err)
		}
		_, y0, _, y1 := g.Path.Round().BBox()
		for _, f := range fair.Analyse(g.Name, g.Path.Round(), fair.Default).Findings {
			if strings.HasPrefix(g.Name, "uni222B") && f.Kind == fair.Inflection && math.Abs(f.Y-(y0+y1)/2) < 0.15*(y1-y0) {
				continue // the S-curve of the stem of the integral
			}
			t.Errorf("%s: a finding of the fairness analysis: %v", g.Name, f)
		}
	}
}

// TestBraceCurlsKeepTheirProportions requires the hook of the brace, the
// height over which the left edge of the stroke leaves the arm, to be 0.08 to
// 0.115 of the height in every size, and the sizes to differ
// in it by at most 20 per cent of their mean; the arm of the sizes does not
// get thinner as the size grows.
func TestBraceCurlsKeepTheirProportions(t *testing.T) {
	for _, s := range derivedSets(1) {
		m, variants, _ := delimOf(t, s.Params)
		names := variants["braceleft"]
		lo, hi, prevArm := math.Inf(1), 0.0, 0.0
		for _, n := range names {
			g := m[n].Path.Round()
			x0, y0, x1, y1 := g.BBox()
			h := y1 - y0
			arm := runsH(g, y1-0.25*h)
			if len(arm) != 1 {
				t.Fatalf("%s: %s has %d runs at a quarter of its height", s.Name, n, len(arm))
			}
			a, b := arm[0][0], arm[0][1]
			depth := 0.0
			for d := 0.0; d < 0.5*h; d++ {
				r := runsH(g, y1-d-0.5)
				if len(r) == 1 && math.Abs(r[0][0]-a) < 0.03*(x1-x0) && math.Abs(r[0][1]-b) < 0.03*(x1-x0) {
					depth = d
					break
				}
			}
			if f := depth / h; f < 0.08 || f > 0.115 {
				t.Errorf("%s: the hook of %s is %.3f of its height, want 0.08 to 0.115", s.Name, n, f)
			}
			lo, hi = math.Min(lo, depth/h), math.Max(hi, depth/h)
			if b-a < prevArm-1 {
				t.Errorf("%s: the arm of %s is %v thick, thinner than the size before (%v)", s.Name, n, b-a, prevArm)
			}
			prevArm = b - a
		}
		if mean := (lo + hi) / 2; hi-lo > 0.2*mean {
			t.Errorf("%s: the hook ranges from %.3f to %.3f of the height", s.Name, lo, hi)
		}
	}
}

// TestSumAndProductLeaveAirBeforeTheirScripts requires the white after the
// ink of ∑, ∏ and ∐ at text and display size to lie within the range
// of 57 to 75, where the scripts of the operator begin.
func TestSumAndProductLeaveAirBeforeTheirScripts(t *testing.T) {
	for _, g := range bigOpGlyphs(design.Regular()) {
		_, _, x1, _ := g.Path.Round().BBox()
		if w := g.Advance - x1; w < 57 || w > 75 {
			t.Errorf("%s: %.0f units of white after the ink, want 57 to 75", g.Name, w)
		}
	}
}
