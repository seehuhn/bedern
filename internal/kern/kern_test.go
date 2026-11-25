// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
	"seehuhn.de/go/bedern/internal/outline"
)

// These tests check properties of the kerning.  None of them pins a computed
// kern.

// routineSets is the number of parameter sets of designtest.Sets which go test
// -short checks besides the variants.  Without -short the tests check every
// set and the bold parameters.
const routineSets = 0

type named struct {
	name string
	p    design.Params
}

func allSets() []named {
	var sets []named
	for _, v := range design.Variants() {
		sets = append(sets, named{v.Name, v.Params})
	}
	for _, s := range designtest.Sets(testing.Short(), routineSets) {
		sets = append(sets, named{s.Name, s.Params})
	}
	if !testing.Short() {
		sets = append(sets, named{"bold", design.Bold()})
	}
	return sets
}

var (
	computedMu sync.Mutex
	computedBy = map[design.Params]*computedEntry{}
)

type computedEntry struct {
	once sync.Once
	tb   *Table
}

// computed returns Compute for the parameters and their glyphs, once per test
// binary.  The table is shared and must not be changed.
func computed(p design.Params) *Table {
	computedMu.Lock()
	e := computedBy[p]
	if e == nil {
		e = new(computedEntry)
		computedBy[p] = e
	}
	computedMu.Unlock()
	e.once.Do(func() { e.tb = Compute(p, glyphstest.Set(p)) })
	return e.tb
}

// A seg is a segment of a flattened outline with its bounding box.
type seg struct {
	s              [2]outline.Point
	x0, x1, y0, y1 float64
}

// A flat is a flattened outline: its segments and their bounding box.
type flat struct {
	segs           []seg
	x0, x1, y0, y1 float64
}

// flatten returns the segments of the flattened outline p.
func flatten(p outline.Path) flat {
	f := flat{x0: math.Inf(1), x1: math.Inf(-1), y0: math.Inf(1), y1: math.Inf(-1)}
	for _, c := range p {
		pts := c.Flatten()
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			s := seg{[2]outline.Point{a, b}, math.Min(a.X, b.X), math.Max(a.X, b.X), math.Min(a.Y, b.Y), math.Max(a.Y, b.Y)}
			f.segs = append(f.segs, s)
			f.x0, f.x1, f.y0, f.y1 = math.Min(f.x0, s.x0), math.Max(f.x1, s.x1), math.Min(f.y0, s.y0), math.Max(f.y1, s.y1)
		}
	}
	return f
}

// distanceBelow returns the closest distance of a and of b moved by dx, if it
// is less than limit, and limit otherwise.
func (a flat) distanceBelow(b flat, dx, limit float64) float64 {
	best := limit
	if a.x0-(b.x1+dx) >= best || (b.x0+dx)-a.x1 >= best || a.y0-b.y1 >= best || b.y0-a.y1 >= best {
		return best
	}
	for _, s := range a.segs {
		if s.x0-(b.x1+dx) > best || (b.x0+dx)-s.x1 > best || s.y0-b.y1 > best || b.y0-s.y1 > best {
			continue
		}
		for _, u := range b.segs {
			ux0, ux1 := u.x0+dx, u.x1+dx
			if s.x0-ux1 > best || ux0-s.x1 > best || s.y0-u.y1 > best || u.y0-s.y1 > best {
				continue
			}
			v := [2]outline.Point{{X: u.s[0].X + dx, Y: u.s[0].Y}, {X: u.s[1].X + dx, Y: u.s[1].Y}}
			if d := segDist(s.s, v); d < best {
				best = d
			}
		}
	}
	return best
}

// flats holds the flattened outlines of a set of glyphs, each made when it is
// first asked for.  It is not safe for concurrent use.
type flats struct {
	gs   []glyphs.Glyph
	done []bool
	fs   []flat
}

func newFlats(gs []glyphs.Glyph) *flats {
	return &flats{gs: gs, done: make([]bool, len(gs)), fs: make([]flat, len(gs))}
}

// of returns the flattened outline of glyph i.
func (f *flats) of(i int) flat {
	if !f.done[i] {
		f.fs[i], f.done[i] = flatten(f.gs[i].Path), true
	}
	return f.fs[i]
}

func segDist(s, u [2]outline.Point) float64 {
	if segCross(s, u) {
		return 0
	}
	d := math.Inf(1)
	for _, p := range s {
		d = math.Min(d, ptSeg(p, u))
	}
	for _, p := range u {
		d = math.Min(d, ptSeg(p, s))
	}
	return d
}

func segCross(s, u [2]outline.Point) bool {
	o := func(a, b, c outline.Point) float64 { return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X) }
	return o(s[0], s[1], u[0])*o(s[0], s[1], u[1]) < 0 && o(u[0], u[1], s[0])*o(u[0], u[1], s[1]) < 0
}

func ptSeg(p outline.Point, s [2]outline.Point) float64 {
	dx, dy := s[1].X-s[0].X, s[1].Y-s[0].Y
	l := dx*dx + dy*dy
	t := 0.0
	if l > 0 {
		t = math.Max(0, math.Min(1, ((p.X-s[0].X)*dx+(p.Y-s[0].Y)*dy)/l))
	}
	return math.Hypot(p.X-s[0].X-t*dx, p.Y-s[0].Y-t*dy)
}

// TestKernKeepsTheFloor: no pair of the font, kerned as the font will do it,
// comes closer than the floor, in every variant and parameter set.  The
// distance of the computation may be 1.5 too large; the test allows 2.
func TestKernKeepsTheFloor(t *testing.T) {
	t.Parallel()
	for _, s := range allSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			gs := glyphstest.Set(s.p)
			tb := computed(s.p)
			if len(tb.Pairs) == 0 {
				t.Fatal("no kerns")
			}
			floor := s.p.Kern.Floor
			fl := newFlats(gs)
			for _, pr := range tb.Pairs {
				k := tb.Kern(pr.Left.Index, pr.Right.Index)
				if k != pr.Kern {
					t.Fatalf("%s %s: the table gives %d, the pair %d", pr.Left.Name, pr.Right.Name, k, pr.Kern)
				}
				if d := fl.of(pr.Left.Index).distanceBelow(fl.of(pr.Right.Index), gs[pr.Left.Index].Advance+float64(k), floor-2); d < floor-2 {
					t.Errorf("%s %s with kern %d: closest distance %.1f, floor %v", pr.Left.Name, pr.Right.Name, k, d, floor)
				}
			}
		})
	}
}

// TestStraightPairsAreNotKerned: stems against stems keep their side spaces.
func TestStraightPairsAreNotKerned(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	tb := computed(p)
	for _, s := range []string{"nn", "HH", "ll", "II", "ln", "nl", "HI", "IH"} {
		a, b := tb.Font.glyphOf(rune(s[0])), tb.Font.glyphOf(rune(s[1]))
		if k := tb.Kern(a.Index, b.Index); k != 0 {
			t.Errorf("%s has the kern %d", s, k)
		}
	}
}

// TestKernMovesTowardsTheTarget: a kern never leaves the white of its pair
// further from the target (plus the allowance) than it was, except by the
// pair tolerance, and except where the floor sets the kern.
func TestKernMovesTowardsTheTarget(t *testing.T) {
	t.Parallel()
	for _, s := range allSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			tol := s.p.Kern.PairTolerance + 1
			for _, pr := range tb.Pairs {
				if pr.Floor || pr.Reason != "" {
					continue
				}
				tg := pr.Target + pr.Allow
				if math.Abs(pr.White1-tg) > math.Abs(pr.White0-tg)+tol {
					t.Errorf("%s %s: white %.1f, with kern %d %.1f, target %.1f", pr.Left.Name, pr.Right.Name, pr.White0, pr.Kern, pr.White1, tg)
				}
			}
		})
	}
}

// TestProfilesOfSymmetricGlyphs: the left and the right profile of a glyph
// which is its own mirror image agree within the class tolerance, apart from
// the difference of its two side spaces.  The diagonal letters (v, w, x, A and
// V) have a thick and a thin stroke; for them three times the tolerance is
// allowed.
func TestProfilesOfSymmetricGlyphs(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	f := computed(p).Font
	for _, r := range "ovwxAHOTV" {
		g := f.glyphOf(r)
		tol := p.Kern.ClassTolerance
		if r == 'v' || r == 'w' || r == 'x' || r == 'A' || r == 'V' {
			tol *= 3
		}
		z := f.profileZone(g)
		// Compare the shapes, after taking the difference of the smallest
		// depths away.
		var l0, r0 = math.Inf(1), math.Inf(1)
		for i := z.Lo; i < z.Hi; i++ {
			if !math.IsNaN(g.Left[i]) {
				l0 = math.Min(l0, g.Left[i])
			}
			if !math.IsNaN(g.Right[i]) {
				r0 = math.Min(r0, g.Right[i])
			}
		}
		for i := z.Lo; i < z.Hi; i++ {
			l, rr := g.Left[i], g.Right[i]
			if math.IsNaN(l) != math.IsNaN(rr) {
				t.Errorf("%c: band %d has ink on one side only", r, i)
				continue
			}
			if math.IsNaN(l) {
				continue
			}
			if d := math.Abs((l - l0) - (rr - r0)); d > tol {
				t.Errorf("%c: band %d: left %.1f, right %.1f", r, i, l-l0, rr-r0)
				break
			}
		}
	}
}

// TestKernKeepsTheClosestApproach: no negative kern takes the closest approach
// of a pair (the smallest band gap over the whole height) below the fraction
// design.Kern.Approach of that of the pair named by Ref, and the limit
// at the stems of the body of the zone holds for the kern as written.
func TestKernKeepsTheClosestApproach(t *testing.T) {
	t.Parallel()
	for _, s := range allSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			f := tb.Font
			for _, pr := range tb.Pairs {
				k := tb.Kern(pr.Left.Index, pr.Right.Index)
				if k >= 0 || pr.Reason != "" {
					continue
				}
				c := f.Closest(pr.Left, pr.Right, float64(k), 0, f.Rows())
				need := s.p.Kern.Approach * f.refClosest(pr.Left, pr.Right, pr.Ref, 0, f.Rows())
				if c < need-1 {
					t.Errorf("%s %s with kern %d: closest approach %.1f, least %.1f", pr.Left.Name, pr.Right.Name, k, c, need)
				}
				if f.limitApproach(pr.Left, pr.Right, pr.Zone, pr.Ref, k) != k {
					t.Errorf("%s %s: the kern %d goes beyond the limit of the closest approach", pr.Left.Name, pr.Right.Name, k)
				}
			}
		})
	}
}

// TestOverhangsAgainstStems: an overhanging part facing a full-height straight
// stem (the bar of T and the ascender of h, the arm of L and the foot of H, the
// bowl of P and the stem of H) is not pulled in by a kern, while the pairs in
// which the second letter can move into the white under the arm (To, Ta, LT,
// LY, AV) are.  The test runs for the variants.
func TestOverhangsAgainstStems(t *testing.T) {
	t.Parallel()
	var sets []named
	for _, v := range design.Variants() {
		sets = append(sets, named{v.Name, v.Params})
	}
	for _, s := range sets {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			kern := func(pair string) int {
				a, b := tb.Font.glyphOf(rune(pair[0])), tb.Font.glyphOf(rune(pair[1]))
				return tb.Kern(a.Index, b.Index)
			}
			for _, pair := range []string{"Th", "LH", "PH", "FE", "Tl"} {
				if k := kern(pair); k < 0 && float64(-k) >= s.p.Kern.Threshold {
					t.Errorf("%s has the kern %d, but its closest approach is that of two stems", pair, k)
				}
			}
			for _, pair := range []string{"To", "Ta", "LT", "LY", "AV"} {
				if k := kern(pair); k >= 0 {
					t.Errorf("%s has the kern %d, but its second letter can move into the white", pair, k)
				}
			}
		})
	}
}

// TestDepthTellsPairsApart checks the cap of the white in one band: it lies
// above the white between the stems of HH in every band, so that a pair of
// capitals responds to a shift, and above the depth of the open side of e and
// c, so that these count in full and not as much as a stem gap.
func TestDepthTellsPairsApart(t *testing.T) {
	p := design.Regular()
	f := computed(p).Font
	depth := p.Kern.Depth
	z := Zone{f.Row(0), f.Row(0) + int(p.CapHeight/p.Kern.Band)}
	H := f.glyphOf('H')
	var gap float64
	for i := z.Lo; i < z.Hi; i++ {
		if !math.IsNaN(H.Right[i]) && !math.IsNaN(H.Left[i]) {
			gap = math.Max(gap, H.Right[i]+H.Left[i])
		}
	}
	if depth <= gap {
		t.Errorf("Kern.Depth %.0f is not above the largest band gap of HH, %.0f", depth, gap)
	}
	for _, r := range "ec" {
		var d float64
		for i := f.Row(0); i < f.Row(0)+int(p.XHeight/p.Kern.Band); i++ {
			if v := f.glyphOf(r).Right[i]; !math.IsNaN(v) {
				d = math.Max(d, v)
			}
		}
		if depth < d {
			t.Errorf("Kern.Depth %.0f is below the depth of the open side of %c, %.0f", depth, r, d)
		}
	}
}

// TestMarksKeepTheFloor: no lowercase letter comes closer than the floor to a
// mark on either side, the brackets and the raised marks included, kerned as
// the font does it, in the regular font and the variants.
func TestMarksKeepTheFloor(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			gs := glyphstest.Set(v.Params)
			tb := computed(v.Params)
			f := tb.Font
			floor := v.Params.Kern.Floor
			fl := newFlats(gs)
			for _, a := range f.Glyphs {
				for _, b := range f.Glyphs {
					if !(a.Kind == Punct && b.Kind == Lower || a.Kind == Lower && b.Kind == Punct) {
						continue
					}
					k := tb.Kern(a.Index, b.Index)
					if f.Distance(a, b, float64(k)) > floor+10 {
						continue
					}
					if d := fl.of(a.Index).distanceBelow(fl.of(b.Index), gs[a.Index].Advance+float64(k), floor-2); d < floor-2 {
						t.Errorf("%s %s with kern %d: closest distance %.1f, floor %v", a.Name, b.Name, k, d, floor)
					}
				}
			}
		})
	}
}

// kernRanges gives for some pairs the range of the kern, as fractions of the
// em.
var kernRanges = []struct {
	pair   string
	lo, hi float64
}{
	{"Ta", -0.105, -0.020}, {"Te", -0.105, -0.025}, {"To", -0.105, -0.025}, {"Tr", -0.100, 0.020}, {"Tu", -0.110, 0.020}, {"Ty", -0.110, 0.025},
	{"Va", -0.100, -0.056}, {"Ve", -0.100, -0.065}, {"Vo", -0.100, -0.065}, {"Vr", -0.090, -0.035}, {"Vu", -0.083, 0.000}, {"Vy", -0.100, -0.013},
	{"Wa", -0.100, -0.056}, {"We", -0.100, -0.061}, {"Wo", -0.100, -0.061}, {"Wr", -0.090, -0.035}, {"Wu", -0.083, 0.000}, {"Wy", -0.100, -0.013},
	{"Ya", -0.100, -0.054}, {"Ye", -0.105, -0.071}, {"Yo", -0.105, -0.071}, {"Yu", -0.083, 0.000}, {"Yr", -0.083, 0.000},
	{"AT", -0.095, -0.030}, {"AV", -0.140, -0.056}, {"AW", -0.140, -0.055}, {"AY", -0.090, -0.042},
	{"Av", -0.111, -0.037}, {"Aw", -0.083, -0.037}, {"Ay", -0.090, -0.023}, {"A’", -0.120, 0.000}, {"A”", -0.125, 0.000},
	{"LT", -0.120, 0.000}, {"LV", -0.125, 0.000}, {"LW", -0.111, 0.000}, {"LY", -0.108, 0.000}, {"L’", -0.125, 0.000}, {"L”", -0.130, 0.000},
	{"Pa", -0.050, 0.000}, {"Po", -0.055, 0.000}, {"P,", -0.188, -0.024}, {"P.", -0.188, -0.024},
	{"Fa", -0.083, 0.000}, {"Fo", -0.083, 0.000}, {"F,", -0.169, 0.000}, {"F.", -0.169, 0.000},
	{"Ko", -0.030, 0.000}, {"Ro", -0.040, 0.015},
	{"r,", -0.111, 0.000}, {"r.", -0.111, 0.000}, {"v,", -0.120, 0.000}, {"v.", -0.120, 0.000},
	{"w,", -0.120, 0.000}, {"w.", -0.120, 0.000}, {"y,", -0.133, -0.059}, {"y.", -0.133, -0.059},
	{"f)", 0.000, 0.094}, {"f]", 0.000, 0.094}, {"f?", 0.000, 0.070}, {"f’", 0.000, 0.101}, {"f”", 0.000, 0.101},
	{"(j", 0.000, 0.102}, {"[j", 0.000, 0.102}, {"{j", 0.000, 0.102},
	{"ov", -0.035, 0.000}, {"ev", -0.025, 0.000}, {"ey", -0.030, 0.000}, {"cv", -0.020, 0.000}, {"cy", -0.020, 0.000},
	{"ox", -0.035, 0.000}, {"rv", 0.000, 0.025},
	{"T.", -0.103, 0.000}, {"V,", -0.140, 0.000}, {"“A", -0.130, 0.000}, {"‘A", -0.130, 0.000},
}

// kernRangeMargin widens each range of kernRanges on both sides, as a fraction
// of the em.
const kernRangeMargin = 0.015

// TestKernRanges: in the regular font the kern of each pair of kernRanges lies
// within its range, widened by kernRangeMargin.
func TestKernRanges(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	tb := computed(p)
	for _, c := range kernRanges {
		rs := []rune(c.pair)
		a, b := tb.Font.glyphOf(rs[0]), tb.Font.glyphOf(rs[1])
		if a == nil || b == nil {
			t.Errorf("%s: no glyph", c.pair)
			continue
		}
		k := float64(tb.Kern(a.Index, b.Index)) / p.UnitsPerEm
		if k < c.lo-kernRangeMargin || k > c.hi+kernRangeMargin {
			t.Errorf("%s: kern %.3f em, want %.3f to %.3f", c.pair, k, c.lo-kernRangeMargin, c.hi+kernRangeMargin)
		}
	}
}

// dashKernSpread is the largest difference in units between the kerns of a
// capital before the hyphen, the en dash and the em dash.
const dashKernSpread = 10

// capitalsC holds the glyph names of C and its accented forms.
var capitalsC = map[string]bool{"C": true, "Cacute": true, "Ccircumflex": true, "Cdotaccent": true, "Ccaron": true, "Ccedilla": true}

// TestCapitalsKernAlikeBeforeDashes: a capital has the same kern, within
// dashKernSpread, before the hyphen, the en dash and the em dash, and C and
// its accented forms, with a mouth at the height of the dashes, have none.
func TestCapitalsKernAlikeBeforeDashes(t *testing.T) {
	t.Parallel()
	for _, s := range allSets() {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			tb := computed(s.p)
			f := tb.Font
			for _, a := range f.Glyphs {
				if a.Kind != Upper {
					continue
				}
				lo, hi := math.Inf(1), math.Inf(-1)
				for _, r := range dashes {
					k := float64(tb.Kern(a.Index, f.glyphOf(r).Index))
					lo, hi = math.Min(lo, k), math.Max(hi, k)
				}
				if hi-lo > dashKernSpread {
					t.Errorf("%s: kerns before the dashes range from %g to %g", a.Name, lo, hi)
				}
				if capitalsC[a.Name] && (lo != 0 || hi != 0) {
					t.Errorf("%s: kerns before the dashes range from %g to %g, want none", a.Name, lo, hi)
				}
			}
		})
	}
}
