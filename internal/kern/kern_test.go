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
// -short checks besides the variants.
const routineSets = 8

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

// exactDistance is the closest distance of the flattened outlines of a and b,
// with b moved by its advance and kern, computed from segments and independent
// of the sampling of Font.Distance.
func exactDistance(ga, gb glyphs.Glyph, kern float64) float64 {
	// A segment with its bounding box.
	type seg struct {
		s              [2]outline.Point
		x0, x1, y0, y1 float64
	}
	segs := func(p outline.Path, dx float64) []seg {
		var out []seg
		for _, c := range p {
			pts := c.Flatten()
			for i := range pts {
				q := pts[(i+1)%len(pts)]
				a, b := outline.Point{X: pts[i].X + dx, Y: pts[i].Y}, outline.Point{X: q.X + dx, Y: q.Y}
				out = append(out, seg{[2]outline.Point{a, b}, math.Min(a.X, b.X), math.Max(a.X, b.X), math.Min(a.Y, b.Y), math.Max(a.Y, b.Y)})
			}
		}
		return out
	}
	sa, sb := segs(ga.Path, 0), segs(gb.Path, ga.Advance+kern)
	best := math.Inf(1)
	for _, s := range sa {
		for _, u := range sb {
			if s.x0-u.x1 > best || u.x0-s.x1 > best || s.y0-u.y1 > best || u.y0-s.y1 > best {
				continue
			}
			if d := segDist(s.s, u.s); d < best {
				best = d
			}
		}
	}
	return best
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
			for _, pr := range tb.Pairs {
				k := tb.Kern(pr.Left.Index, pr.Right.Index)
				if k != pr.Kern {
					t.Fatalf("%s %s: the table gives %d, the pair %d", pr.Left.Name, pr.Right.Name, k, pr.Kern)
				}
				if d := exactDistance(gs[pr.Left.Index], gs[pr.Right.Index], float64(k)); d < floor-2 {
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
