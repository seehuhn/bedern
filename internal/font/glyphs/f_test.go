// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The values checked for f: the stem and the feet of l (81 at 99, feet 69 and
// 80), a hook which rises to the ascender with its overshoot (704) and ends 160
// beyond the stem in the ball terminal, a crossbar of the hairline thickness 36
// with its top at the x-height and extents of 75 and 112, and the advance 340
// with the hook ending at the advance (the overhang is 0).

// fGlyph returns the outline and the advance of f in the parameters p, and
// fails the test if it is not a single contour.
func fGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("f", p)
	if !ok {
		t.Fatal("glyph f not found")
	}
	if len(path) != 1 {
		t.Fatalf("f has %d contours, want one: the bar and the stem are one outline", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of f is not counter-clockwise")
	}
	return path, advance
}

func TestFAdvanceAndStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance := fGlyph(t, p)
	if advance != design.FAdvance(p) {
		t.Errorf("advance %v, want %v", advance, design.FAdvance(p))
	}
	x0, y0, x1, y1 := path.BBox()
	near(t, "bottom", y0, 0)
	near(t, "top", y1, p.Ascender+p.FOvershoot)
	near(t, "left ink", x0, design.FInkLeft(p))
	near(t, "rightmost ink beyond the cell of the stem", x1-design.FCellRight(p), p.FReach)
	near(t, "overhang of the advance", x1-advance, -p.FSideRight)
	// The stem, measured between the foot and the bar.
	xs := path.CrossingsX(300)
	if len(xs) != 2 {
		t.Fatalf("crossings at height 300: %v", xs)
	}
	near(t, "left edge of the stem", xs[0], 99)
	near(t, "stem width", xs[1]-xs[0], p.AscenderStem)
	// It is the stem of l: the same crossings below the bar.
	l, _, _ := outlineOf("l", p)
	for y := p.BracketHeight + 5; y < p.XHeight-p.Hair-5; y += 7 {
		if fs, ls := path.CrossingsX(y), l.CrossingsX(y); len(fs) != len(ls) || math.Abs(fs[0]-ls[0]) > 0.01 || math.Abs(fs[1]-ls[1]) > 0.01 {
			t.Errorf("at height %.0f the stem of f has the crossings %v, that of l %v", y, fs, ls)
		}
	}
}

func TestFFeet(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := fGlyph(t, p)
	xl, xr := design.FStemLeft(p), design.FStemLeft(p)+p.AscenderStem
	checkFoot(t, p, path, xr, true, p.FFootRight, p.SerifThick)
	checkFoot(t, p, path, xl, false, p.SerifJut, p.SerifThick)
}

// The crossbar is the hairline thick and has its top at the x-height; its
// ends are plain cuts; it extends 75 to the left of the stem and 112 to the
// right.
func TestFCrossbar(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := fGlyph(t, p)
	xl, xr := design.FStemLeft(p), design.FStemLeft(p)+p.AscenderStem
	for _, c := range []struct {
		name string
		x    float64
	}{{"left part", xl - 40}, {"right part", xr + 50}} {
		var ys []float64
		for _, y := range path.CrossingsY(c.x) {
			if y > p.BracketHeight+p.SerifThick && y < p.XHeight+50 { // neither the foot serif nor the hook
				ys = append(ys, y)
			}
		}
		if len(ys) != 2 {
			t.Fatalf("%s: crossings %v", c.name, ys)
		}
		near(t, c.name+": top of the bar", ys[1], p.XHeight)
		near(t, c.name+": thickness of the bar", ys[1]-ys[0], p.Hair)
	}
	// The extents: the bar at its middle height.
	xs := path.CrossingsX(p.XHeight - p.Hair/2)
	if len(xs) != 2 {
		t.Fatalf("crossings through the bar: %v", xs)
	}
	near(t, "extent to the left", xl-xs[0], p.FBarLeft)
	near(t, "extent to the right", xs[1]-xr, p.FBarRight)
	// The ends are plain: vertical cuts as thick as the bar.
	for _, x := range []float64{xs[0] + 1, xs[1] - 1} {
		var ys []float64
		for _, y := range path.CrossingsY(x) {
			if y > p.BracketHeight+p.SerifThick && y < p.XHeight+50 {
				ys = append(ys, y)
			}
		}
		if len(ys) != 2 {
			t.Fatalf("crossings near the end of the bar: %v", ys)
		}
		near(t, "thickness of the bar 1 unit in from its end", ys[1]-ys[0], p.Hair)
	}
}

// The left edge of the stem becomes the outer edge of the hook and the right edge
// its underside, each in a curve which leaves the straight edge vertically,
// without a corner, above the crossbar.
func TestFEdgesLeaveTheStemSmoothly(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := fGlyph(t, p)
	ct := path[0]
	xl, xr := design.FStemLeft(p), design.FStemLeft(p)+p.AscenderStem
	for _, c := range []struct {
		name string
		up   bool // the contour runs up into the curve
		vert float64
	}{
		{"left edge", false, xl},
		{"right edge", true, xr},
	} {
		i := stemCurveNode(t, ct, c.vert, p.XHeight, design.FApex(p))
		before, after := segCubic(ct, i+len(ct.Segs)-1), segCubic(ct, i)
		near(t, c.name+": the node is on the edge of the stem", ct.Segs[(i+len(ct.Segs)-1)%len(ct.Segs)].P3.X, c.vert)
		// The straight edge runs into the curve, and both are vertical there.
		if c.up {
			if before.P0.X != before.P3.X || after.P1.X != c.vert {
				t.Errorf("%s: the edge below the node is %v, the handle of the curve %v, want both vertical", c.name, before, after.P1)
			}
		} else if before.P2.X != c.vert || after.P0.X != after.P3.X {
			t.Errorf("%s: the handle of the curve before the node is %v, the edge after it %v, want both vertical", c.name, before.P2, after)
		}
	}
}

// stemCurveNode returns the index of the node of the contour on the vertical x
// between the heights y0 and y1 at which a straight segment and a curve meet.
func stemCurveNode(t *testing.T, ct outline.Contour, x, y0, y1 float64) int {
	t.Helper()
	n := len(ct.Segs)
	nodes := contourNodes(ct)
	found := -1
	for i := 0; i < n; i++ {
		q := nodes[i]
		if math.Abs(q.X-x) > 1e-9 || q.Y <= y0 || q.Y >= y1 {
			continue
		}
		if ct.Segs[(i+n-1)%n].IsCurve != ct.Segs[i].IsCurve {
			if found >= 0 {
				t.Fatalf("two nodes on the edge x = %v where a line meets a curve", x)
			}
			found = i
		}
	}
	if found < 0 {
		t.Fatalf("no node on the edge x = %v where a line meets a curve", x)
	}
	return found
}

// The outer edge of the hook is smooth through its highest point, with
// horizontal handles of the two curves pointing away from each other and a
// continuous curvature.
func TestFCrestIsSmooth(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := fGlyph(t, p)
	ct := path[0]
	var crest outline.Point
	for _, q := range contourNodes(ct) {
		if q.Y > crest.Y {
			crest = q
		}
	}
	near(t, "height of the crest", crest.Y, design.FApex(p))
	i := nodeNear(t, ct, crest)
	before, after := segCubic(ct, i+len(ct.Segs)-1), segCubic(ct, i)
	if math.Abs(before.P2.Y-crest.Y) > crestLevel || math.Abs(after.P1.Y-crest.Y) > crestLevel {
		t.Errorf("the handles at the crest are not horizontal: %v and %v", before.P2, after.P1)
	}
	if !(before.P2.X > crest.X && after.P1.X < crest.X) {
		t.Errorf("the handles at the crest do not point away from each other: %v and %v on the node %v", before.P2, after.P1, crest)
	}
	k0, k1, d0, d1 := curvaturesAround(t, ct, crest, crestDistance)
	if k0 <= 0 || k1 <= 0 || d0 <= 0 || d1 <= 0 {
		t.Errorf("the edge is not convex at the crest: curvatures %.4f %.4f %.4f %.4f", k0, k1, d0, d1)
	}
	t.Logf("curvature before %.4f, after %.4f (ratio %.2f); %v away: %.4f, %.4f (ratio %.2f)", k0, k1, k1/k0, crestDistance, d0, d1, d1/d0)
	for _, r := range []float64{k1 / k0, d1 / d0} {
		if r < crestRatioLow || r > crestRatioHigh {
			t.Errorf("the curvature after the crest is %.2f times that before it, want %v to %v", r, crestRatioLow, crestRatioHigh)
		}
	}
}

// The hook of f overhangs the advance by 60 at most, and in the regular
// parameters by at least fOverhangMin.
func TestFOverhang(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance := fGlyph(t, p)
	_, _, x1, _ := path.BBox()
	const fOverhangMin = 40.0
	if over := x1 - advance; over < fOverhangMin-0.5 || over > p.MaxOverhang+0.5 {
		t.Errorf("the hook overhangs the advance by %.1f, want from %v to %v", over, fOverhangMin, p.MaxOverhang)
	}
}

// minDistanceBetween returns the least distance between the flattened outlines
// of two paths, with the nearest points.
func minDistanceBetween(a, b outline.Path) (d float64, pa, pb outline.Point) {
	d = math.Inf(1)
	// A flattened contour with its bounding box.
	type flatContour struct {
		pts            []outline.Point
		x0, y0, x1, y1 float64
	}
	flat := func(p outline.Path) []flatContour {
		var all []flatContour
		for _, ct := range p {
			c := flatContour{pts: ct.Flatten(), x0: math.Inf(1), y0: math.Inf(1), x1: math.Inf(-1), y1: math.Inf(-1)}
			for _, q := range c.pts {
				c.x0, c.y0, c.x1, c.y1 = min(c.x0, q.X), min(c.y0, q.Y), max(c.x1, q.X), max(c.y1, q.Y)
			}
			all = append(all, c)
		}
		return all
	}
	// far says whether the box from (lo, bot) to (hi, top) is at least d from
	// the box of c.
	far := func(lo, bot, hi, top float64, c flatContour) bool {
		gx := max(0, max(lo, c.x0)-min(hi, c.x1))
		gy := max(0, max(bot, c.y0)-min(top, c.y1))
		return gx*gx+gy*gy >= d*d
	}
	fb := flat(b)
	for _, ca := range flat(a) {
		pa0 := ca.pts
		for _, cb := range fb {
			pb0 := cb.pts
			if far(ca.x0, ca.y0, ca.x1, ca.y1, cb) {
				continue
			}
			for i, q := range pa0 {
				q2 := pa0[(i+1)%len(pa0)]
				qlo, qhi := min(q.X, q2.X), max(q.X, q2.X)
				qbot, qtop := min(q.Y, q2.Y), max(q.Y, q2.Y)
				if far(qlo, qbot, qhi, qtop, cb) {
					continue
				}
				for j, r := range pb0 {
					r2 := pb0[(j+1)%len(pb0)]
					// A segment whose bounding box is at least d from that of
					// the other is skipped.
					gx := max(0, max(qlo, min(r.X, r2.X))-min(qhi, max(r.X, r2.X)))
					gy := max(0, max(qbot, min(r.Y, r2.Y))-min(qtop, max(r.Y, r2.Y)))
					if gx*gx+gy*gy >= d*d {
						continue
					}
					for _, c := range []struct {
						pt     outline.Point
						s0, s1 outline.Point
						first  bool
					}{{q, r, r2, true}, {q2, r, r2, true}, {r, q, q2, false}, {r2, q, q2, false}} {
						if dd := segDist(c.pt, c.s0, c.s1); dd < d {
							d = dd
							if c.first {
								pa, pb = c.pt, nearestOnSeg(c.pt, c.s0, c.s1)
							} else {
								pb, pa = c.pt, nearestOnSeg(c.pt, c.s0, c.s1)
							}
						}
					}
				}
			}
		}
	}
	return d, pa, pb
}

// nearestOnSeg returns the point of the segment from a to b nearest to q.
func nearestOnSeg(q, a, b outline.Point) outline.Point {
	dx, dy := b.X-a.X, b.Y-a.Y
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return a
	}
	t := math.Max(0, math.Min(1, ((q.X-a.X)*dx+(q.Y-a.Y)*dy)/l2))
	return outline.Point{X: a.X + t*dx, Y: a.Y + t*dy}
}

// shiftedPath returns the path moved by dx to the right.
func shiftedPath(p outline.Path, dx float64) outline.Path {
	q := make(outline.Path, len(p))
	for i, ct := range p {
		q[i] = outline.Contour{Start: outline.Point{X: ct.Start.X + dx, Y: ct.Start.Y}}
		for _, s := range ct.Segs {
			q[i].Segs = append(q[i].Segs, outline.Segment{IsCurve: s.IsCurve,
				P1: outline.Point{X: s.P1.X + dx, Y: s.P1.Y},
				P2: outline.Point{X: s.P2.X + dx, Y: s.P2.Y},
				P3: outline.Point{X: s.P3.X + dx, Y: s.P3.Y}})
		}
	}
	return q
}

// The nearest approach of f to what may follow it, measured between outlines
// with the advance of f as the only space: the hook overhangs the advance and
// stands level with the head serifs of the letters with ascenders, which it
// meets, and f keeps fClearance from every other lowercase letter.  The kerning
// keeps the floor where the hook meets a letter, and fl is a ligature.

// fClearance is the least distance which f keeps from every lowercase letter
// that may follow it without a kern, the floor of the kerning.
const fClearance = 20.0

// followers lists every lowercase letter.
var followers = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

// fMeets lists the letters whose head serifs the hook of f meets.
var fMeets = map[string]bool{"b": true, "h": true, "k": true, "l": true}

// f keeps its nearest approach of at least fClearance to every lowercase letter
// which exists except those of fMeets, which come closer, in the regular
// parameters and in every variant.  The distance is measured between the
// outlines with the advance of f as the only space.
func TestFClearsEveryFollowingLetter(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		np := namedParams{name: v.Name, params: v.Params}
		f, adv := fGlyph(t, np.params)
		for _, name := range followers {
			next, _, ok := outlineOf(name, np.params)
			if !ok {
				continue
			}
			d, a, b := minDistanceBetween(f, shiftedPath(next, adv))
			if np.name == "regular" {
				t.Logf("f followed by %s: nearest approach %.1f, at (%.0f, %.0f) and (%.0f, %.0f)", name, d, a.X, a.Y, b.X, b.Y)
			}
			switch {
			case fMeets[name] && d >= fClearance:
				t.Errorf("%s: f followed by %s: nearest approach %.1f, want the hook to come closer than %v", np.name, name, d, fClearance)
			case !fMeets[name] && d < fClearance:
				t.Errorf("%s: f followed by %s: nearest approach %.1f, want at least %v", np.name, name, d, fClearance)
			}
		}
	}
}

// The tail of j reaches beyond its left edge and keeps its distance from what
// precedes it; the tail of t keeps its distance from what follows it.  Both are
// measured between outlines with the advances as the only space.

// pairClearance is the least distance which the tails of j and t keep from their
// neighbours.
const pairClearance = 20.0

func TestJTailKeepsAwayFromWhatPrecedes(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		j, _, _, _ := jGlyph(t, v.Params)
		for _, name := range followers {
			prev, adv, ok := outlineOf(name, v.Params)
			if !ok {
				continue
			}
			d, a, b := minDistanceBetween(prev, shiftedPath(j, adv))
			if v.Name == "regular" {
				t.Logf("%s followed by j: nearest approach %.1f, at (%.0f, %.0f) and (%.0f, %.0f)", name, d, a.X, a.Y, b.X, b.Y)
			}
			if name == "q" || name == "g" {
				// The tail of j touches the foot serif of the descender of q and
				// the lower right of the loop of g.
				if d > 1 {
					t.Errorf("%s: %s followed by j: %.1f apart, want the pair to touch", v.Name, name, d)
				}
				continue
			}
			if d < pairClearance {
				t.Errorf("%s: %s followed by j: nearest approach %.1f, want at least %v", v.Name, name, d, pairClearance)
			}
		}
	}
}

func TestTTailKeepsAwayFromWhatFollows(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		tt, adv, ok := outlineOf("t", v.Params)
		if !ok {
			t.Fatal("glyph t not found")
		}
		for _, name := range followers {
			next, _, ok := outlineOf(name, v.Params)
			if !ok {
				continue
			}
			d, a, b := minDistanceBetween(tt, shiftedPath(next, adv))
			if v.Name == "regular" {
				t.Logf("t followed by %s: nearest approach %.1f, at (%.0f, %.0f) and (%.0f, %.0f)", name, d, a.X, a.Y, b.X, b.Y)
			}
			if d < pairClearance {
				t.Errorf("%s: t followed by %s: nearest approach %.1f, want at least %v", v.Name, name, d, pairClearance)
			}
		}
	}
}
