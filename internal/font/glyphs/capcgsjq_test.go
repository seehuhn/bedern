// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of C, G, S, J and Q measure the outlines in every variant and
// perturbed parameter set: that C, G and Q are the round of O, the overshoots,
// the beaks and apertures of C and G, the stem, bar and foot of G, the tail and
// the ball of J, the tail of Q, the S against the rules of s, and the edges of
// all five, which must have no kink, no bump and no straight stretch.

// leftRun returns the first run of the scan at y.
func leftRun(t *testing.T, path outline.Path, y float64) [2]float64 {
	t.Helper()
	r := runs(path, y)
	if len(r) == 0 {
		t.Fatalf("no ink at the height %.1f", y)
	}
	return r[0]
}

// C, G and Q are cut from the round of O: on their left side, from a tenth to
// nine tenths of the cap height, the outer and the inner edge are those of O
// within half a unit, and so are the thin parts at the top and the bottom on
// the vertical through the extremes.
func TestCapCGQAreTheRoundOfO(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		o := capOutline(np, "O")
		ox0, _, ox1, _ := o.BBox()
		cx := (ox0 + ox1) / 2
		for _, name := range []string{"C", "G", "Q"} {
			g := capOutline(np, name)
			for k := 1; k <= 9; k++ {
				y := cap * float64(k) / 10
				a, b := leftRun(t, o, y), leftRun(t, g, y)
				if math.Abs(a[0]-b[0]) > 0.5 || math.Abs(a[1]-b[1]) > 0.5 {
					t.Errorf("%s at the height %.0f: the left stroke is %.1f to %.1f, O has %.1f to %.1f", name, y, b[0], b[1], a[0], a[1])
				}
			}
			vo, vg := vruns(o, cx), vruns(g, cx)
			if len(vg) < 2 {
				t.Errorf("%s: %d runs on the vertical through the extremes", name, len(vg))
				continue
			}
			lo, hi := vg[0], vg[len(vg)-1]
			if math.Abs(lo[0]-vo[0][0]) > 0.5 || math.Abs(lo[1]-vo[0][1]) > 0.5 ||
				math.Abs(hi[0]-vo[1][0]) > 0.5 || math.Abs(hi[1]-vo[1][1]) > 0.5 {
				t.Errorf("%s: the thin parts are %v and %v, O has %v", name, lo, hi, vo)
			}
		}
	})
}

// The round capitals overshoot as O does: the top of C, G, S and Q and the bottom
// of C, G, S and J lie within one unit of those of O.
func TestCapCGSJQOvershoot(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		_, oy0, _, oy1 := capOutline(np, "O").BBox()
		for _, name := range []string{"C", "G", "S", "J", "Q"} {
			_, y0, _, y1 := capOutline(np, name).BBox()
			if name != "J" && math.Abs(y1-oy1) > 1 {
				t.Errorf("%s reaches %.1f, O %.1f", name, y1, oy1)
			}
			if name != "Q" && math.Abs(y0-oy0) > 1 {
				t.Errorf("%s reaches down to %.1f, O to %.1f", name, y0, oy0)
			}
		}
	})
}

// cBeak measures the upper beak of a letter whose rightmost ink above the middle
// is the cut of the beak: the abscissa of the cut and the height of the end of the
// beak, read on the vertical run 3 units inside the cut.
func cBeak(t *testing.T, path outline.Path, cap float64) (x, end, top float64) {
	t.Helper()
	x = math.Inf(-1)
	for y := cap / 2; y < cap; y++ {
		if r := runs(path, y); len(r) > 0 {
			x = math.Max(x, r[len(r)-1][1])
		}
	}
	v := vruns(path, x-3)
	if len(v) == 0 {
		t.Fatalf("no ink inside the cut of the beak at %.1f", x)
	}
	last := v[len(v)-1]
	return x, last[0], last[1]
}

// The upper beak of C is a vertical beak: its end lies at 0.55 to 0.68 of the cap
// height, and the vertical run 3 units inside the cut is 0.10 to 0.25 of the cap
// height.  The tail reaches 0 to 12 units beyond the vertical of the cut, and the
// aperture, the least vertical gap between the two ends in the right 35 per cent
// of the ink, is 0.25 to 0.45 of the cap height.  The serif of C ends higher than
// that of G (CBeakEnd), whose upper beak it shares in other respects.
func TestCapCBeakAndAperture(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		c := capOutline(np, "C")
		x, end, top := cBeak(t, c, cap)
		if h := (top - end) / cap; h < 0.10 || h > 0.25 {
			t.Errorf("the cut of the beak of C is %.3f of the cap height tall, want 0.10 to 0.25", h)
		}
		if e := end / cap; e < 0.55 || e > 0.68 {
			t.Errorf("the beak of C ends at %.3f of the cap height, want 0.55 to 0.68", e)
		}
		x0, _, x1, _ := c.BBox()
		if x1-x < 0 || x1-x > 12 {
			t.Errorf("the tail of C reaches %.1f, %.1f beyond the cut of the beak, want 0 to 12", x1, x1-x)
		}
		gap := math.Inf(1)
		for xs := x1 - 0.35*(x1-x0); xs < x1-1; xs++ {
			v := vruns(c, xs)
			for i := 0; i+1 < len(v); i++ {
				if v[i][1] < cap/2 && v[i+1][0] > cap/2 {
					gap = math.Min(gap, v[i+1][0]-v[i][1])
				}
			}
		}
		if a := gap / cap; a < 0.25 || a > 0.45 {
			t.Errorf("the aperture of C is %.3f of the cap height, want 0.25 to 0.45", a)
		}
		if np.name == "regular" {
			t.Logf("beak %.0f (%.3f cap), aperture %.0f (%.3f cap)", top-end, (top-end)/cap, gap, gap/cap)
		}
	})
}

// The lower end of C is a tail without a serif: right of the bottom extreme and
// below the middle the outer contour has one straight piece, the cut, and no
// corner but its two ends.  The tail tapers: its thickness (the distance from a
// point of the outer edge to the nearest point of the inner edge) is 1.1 to 1.6
// times the thin part of O where it is thickest, and it falls without rising over
// at least the last 0.3 of the outer edge to the cut, which is 0.75 to 1.1 times
// the hairline.  The top of the tail, the end of the lower arc, lies at 0.24 to
// 0.34 of the cap height.
func TestCapCTail(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		cap := capHeight(np)
		path := capOutline(np, "C")
		ct := path[0]
		n := len(ct.Segs)
		_, y0, _, y1 := path.BBox()
		mid := (y0 + y1) / 2
		start := func(i int) outline.Point { return segCubic(ct, i).P0 }
		bottom := 0 // the segment which starts at the bottom extreme
		for i := 0; i < n; i++ {
			if start(i).Y < start(bottom).Y {
				bottom = i
			}
		}
		cx := start(bottom).X
		// The segments of the tail, from the bottom extreme of the outer edge
		// round the tip to the counter.
		var tail []int
		for i := bottom; ; i = (i + 1) % n {
			c := segCubic(ct, i)
			if !(c.P3.Y < mid && c.P3.X > cx+0.5) {
				tail = append(tail, i)
				break
			}
			tail = append(tail, i)
		}
		cut, lines, corners := -1, 0, 0
		for k, i := range tail {
			if !ct.Segs[i].IsCurve {
				lines++
				cut = k
			}
			if k+1 < len(tail) {
				j := tail[k+1]
				a, b := segCubic(ct, i).tangent(1), segCubic(ct, j).tangent(0)
				if !ct.Segs[i].IsCurve || !ct.Segs[j].IsCurve || math.Abs(math.Atan2(a.X*b.Y-a.Y*b.X, a.X*b.X+a.Y*b.Y)) > 0.01 {
					corners++
				}
			}
		}
		if lines != 1 || corners != 2 {
			t.Fatalf("the tail of C has %d straight pieces and %d corners, want the cut and its two ends", lines, corners)
		}
		sample := func(is []int) []outline.Point {
			var q []outline.Point
			for _, i := range is {
				c := segCubic(ct, i)
				for s := 0; s <= 200; s++ {
					q = append(q, c.at(float64(s)/200))
				}
			}
			return q
		}
		outer, inner := sample(tail[:cut]), sample(tail[cut+1:])
		thick := make([]float64, len(outer))
		arc := make([]float64, len(outer))
		for k, q := range outer {
			d := math.Inf(1)
			for _, r := range inner {
				d = math.Min(d, math.Hypot(q.X-r.X, q.Y-r.Y))
			}
			thick[k] = d
			if k > 0 {
				arc[k] = arc[k-1] + math.Hypot(q.X-outer[k-1].X, q.Y-outer[k-1].Y)
			}
		}
		top := 0
		for k := range thick {
			if thick[k] > thick[top] {
				top = k
			}
		}
		if r := thick[top] / p.Caps.RoundThin; r < 1.1 || r > 1.6 {
			t.Errorf("the tail of C is %.1f thick where it is thickest, %.2f times the thin part of O, want 1.1 to 1.6", thick[top], r)
		}
		last := len(outer) - 1
		if l := (arc[last] - arc[top]) / arc[last]; l < 0.3 {
			t.Errorf("the tail of C is thickest %.2f of its outer edge before the cut, want at least 0.3", l)
		}
		for k := top + 1; k <= last; k++ {
			if thick[k] > thick[k-1]+0.2 {
				t.Errorf("the thickness of the tail of C rises from %.1f to %.1f at (%.0f, %.0f) before the cut", thick[k-1], thick[k], outer[k].X, outer[k].Y)
				break
			}
		}
		cc := segCubic(ct, tail[cut])
		if r := math.Hypot(cc.P3.X-cc.P0.X, cc.P3.Y-cc.P0.Y) / p.Hair; r < 0.75 || r > 1.1 {
			t.Errorf("the cut of the tail of C is %.2f times the hairline, want 0.75 to 1.1", r)
		}
		end := math.Max(cc.P0.Y, cc.P3.Y) / cap
		if end < 0.24 || end > 0.34 {
			t.Errorf("the tail of C ends at %.3f of the cap height, want 0.24 to 0.34", end)
		}
		if np.name == "regular" {
			t.Logf("thickest %.1f, %.0f of %.0f before the cut; cut %.1f; end at %.3f of the cap height", thick[top], arc[last]-arc[top], arc[last], math.Hypot(cc.P3.X-cc.P0.X, cc.P3.Y-cc.P0.Y), end)
		}
	})
}

// The stem of G stands on the vertical of the cut of the beak and is at least as
// heavy as the capital stem between the lower arc and the bar; its right-hand
// edge is straight from the junction with the lower arc to the foot of the
// bracket of the right serif; its top, the bar, lies at 0.38 to 0.50 of the cap
// height, and the opening between the bar and the beak is at least 0.15 of the
// cap height.
func TestCapGStemAndBar(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		cap := capHeight(np)
		g := capOutline(np, "G")
		x, end, _ := cBeak(t, g, cap)
		v := vruns(g, x-1)
		if len(v) < 2 {
			t.Fatalf("G: %d runs at the stem", len(v))
		}
		stem := v[0]
		if b := stem[1] / cap; b < 0.38 || b > 0.50 {
			t.Errorf("the bar of G is at %.3f of the cap height, want 0.38 to 0.50", b)
		}
		if o := (end - stem[1]) / cap; o < 0.15 {
			t.Errorf("the opening of G is %.3f of the cap height, want at least 0.15", o)
		}
		join, foot := p.Caps.CGSJQ.GJoin, p.Caps.CGSJQ.GStemTop-p.Caps.BracketHeight
		mid := (join + foot) / 2
		r := runs(g, mid)
		if w := r[len(r)-1][1] - r[len(r)-1][0]; w < p.Caps.Stem-0.5 || math.Abs(r[len(r)-1][1]-x) > 0.5 {
			t.Errorf("at the height %.0f the stem of G is %.1f to %.1f, want at least %v wide and ending at %.1f", mid, r[len(r)-1][0], r[len(r)-1][1], p.Caps.Stem, x)
		}
		for y := join + 1; y < foot; y++ {
			if r := runs(g, y); math.Abs(r[len(r)-1][1]-x) > 0.5 {
				t.Errorf("at the height %.0f the right-hand edge of the stem of G is at %.1f, want %.1f", y, r[len(r)-1][1], x)
				break
			}
		}
	})
}

// J sits on the baseline: its lowest point lies OvershootUpper below it, as the
// round capitals do, the ball is the leftmost ink, and the end of the hook, the
// top of the ball, lies at 0.15 to 0.32 of the cap height.  The white between the
// stem, the tail and the ball is a counter, at least 60 units wide and 40 high.
func TestCapJTail(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		cap := capHeight(np)
		j := capOutline(np, "J")
		x0, y0, _, _ := j.BBox()
		if math.Abs(y0+p.OvershootUpper) > 1 {
			t.Errorf("the lowest point of J is at %.1f, want %.1f", y0, -p.OvershootUpper)
		}
		stemLeft := leftRun(t, j, cap/2)[0]
		end := 0.0
		ball := np.params.Caps.CGSJQ.JBall
		for x := x0; x < math.Min(x0+ball, stemLeft-1); x++ {
			for _, r := range vruns(j, x) {
				if r[1] < cap/2 {
					end = math.Max(end, r[1])
				}
			}
		}
		if h := end / cap; h < 0.15 || h > 0.32 {
			t.Errorf("the end of the hook of J is at %.3f of the cap height, want 0.15 to 0.32", h)
		}
		// The counter: on the scan at the height of the top of the ball less 20,
		// the white between the ball and the stem.
		r := runs(j, end-20)
		if len(r) < 2 || r[1][0]-r[0][1] < 60 {
			t.Errorf("at the height %.0f the white between the ball and the stem is %v, want at least 60 wide", end-20, r)
		}
		mx := 0.0
		if len(r) >= 2 {
			mx = (r[0][1] + r[1][0]) / 2
		}
		v := vruns(j, mx)
		if len(v) == 0 || v[0][1] > end-60 {
			t.Errorf("the counter of J is less than 40 high above the tail at %.0f: %v", mx, v)
		}
	})
}

// The tail of Q goes down to 140 to 200 units below the baseline, its ink below
// the bowl is 0.22 to 0.37 of the cap height long, it ends inside the right
// extreme of the bowl, and it is 40 to 110 units thick at 0.6 of its length.  The
// notches where its edges leave the bowl are open, at least 60 degrees.
func TestCapQTail(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		q := capOutline(np, "Q")
		_, oy0, ox1, _ := capOutline(np, "O").BBox()
		_, y0, x1, _ := q.BBox()
		if d := -y0; d < 140 || d > 200 {
			t.Errorf("the tail of Q reaches %.1f below the baseline, want 140 to 200", d)
		}
		if x1 > ox1+0.5 {
			t.Errorf("the tail of Q reaches %.1f, beyond the bowl at %.1f", x1, ox1)
		}
		// The extent of the ink 6 units below the bowl.
		lo, hi := math.Inf(1), math.Inf(-1)
		for y := oy0 - 6; y > y0; y-- {
			if r := runs(q, y); len(r) > 0 {
				lo, hi = math.Min(lo, r[0][0]), math.Max(hi, r[len(r)-1][1])
			}
		}
		if l := (hi - lo) / capHeight(np); l < 0.22 || l > 0.37 {
			t.Errorf("the tail of Q is %.3f of the cap height long, want 0.22 to 0.37", l)
		}
		v := vruns(q, lo+0.6*(hi-lo))
		if len(v) == 0 {
			t.Fatalf("no tail halfway along")
		}
		if th := v[0][1] - v[0][0]; th < 40 || th > 110 {
			t.Errorf("the tail of Q is %.1f thick at 0.6 of its length, want 40 to 110", th)
		}
		if np.name == "regular" {
			t.Logf("depth %.0f, length %.0f, thickness %.0f", -y0, hi-lo, v[0][1]-v[0][0])
		}
		ct := q[0]
		n := len(ct.Segs)
		corners := 0
		for i := 0; i < n; i++ {
			a, b := segCubic(ct, i), segCubic(ct, i+1)
			if !ct.Segs[i].IsCurve || !ct.Segs[(i+1)%n].IsCurve {
				continue
			}
			ta, tb := a.tangent(1), b.tangent(0)
			d := math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y) * 180 / math.Pi
			if math.Abs(d) < 0.5 {
				continue
			}
			corners++
			if white := 180 - math.Abs(d); white < 60 {
				t.Errorf("a notch of the tail of Q at (%.0f, %.0f) is %.0f degrees, want at least 60", a.P3.X, a.P3.Y, white)
			}
		}
		if corners != 2 {
			t.Errorf("the outer contour of Q has %d corners between curves, want the 2 of the tail", corners)
		}
	})
}

// capCurveRuns returns the runs of curves of a contour between corners and
// straight pieces, each as samples with their arc lengths.
func capCurveRuns(ct outline.Contour) [][]outline.Point {
	n := len(ct.Segs)
	smooth := func(i int) bool { // the node at the end of segment i
		if !ct.Segs[i].IsCurve || !ct.Segs[(i+1)%n].IsCurve {
			return false
		}
		ta, tb := segCubic(ct, i).tangent(1), segCubic(ct, i+1).tangent(0)
		return math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) < 0.01
	}
	// Start after a node which is not smooth, if there is one.
	start := 0
	for i := 0; i < n; i++ {
		if !smooth(i) {
			start = (i + 1) % n
			break
		}
	}
	var out [][]outline.Point
	var cur []outline.Point
	for k := 0; k < n; k++ {
		i := (start + k) % n
		if ct.Segs[i].IsCurve {
			c := segCubic(ct, i)
			for s := 0; s <= edgeSamples; s++ {
				cur = append(cur, c.at(float64(s)/edgeSamples))
			}
		}
		if !smooth(i) && len(cur) > 0 {
			out = append(out, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		out = append(out, cur)
	}
	return out
}

// capWithoutHollow cuts the stretch of the outer edge from the start P of the
// hollow of the beak b to its tip out of the runs.  The curvature of that
// stretch falls through zero to a small opposite value (one inflection, nothing
// above capBeakHollowCurv, TestCapBeakHollow).  The runs are cut into the pieces
// on either side.
func capWithoutHollow(runs [][]outline.Point, b capBeakParts) [][]outline.Point {
	if b.Hollow == (cubic{}) {
		return runs
	}
	px, ty := b.Hollow.P3.X, b.Corner.Y
	var out [][]outline.Point
	for _, run := range runs {
		var cur []outline.Point
		for _, q := range run {
			if q.X >= px-0.01 && q.Y >= ty-0.01 {
				if len(cur) > 1 {
					out = append(out, cur)
				}
				cur = nil
				continue
			}
			cur = append(cur, q)
		}
		if len(cur) > 1 {
			out = append(out, cur)
		}
	}
	return out
}

// capEdgeWindow and capEdgeTurn: every stretch of capEdgeWindow units of a curved
// edge of C, G, J and Q turns by at least capEdgeTurn radians (a mean curvature of
// 0.001 per unit); the turns of the regular letters are logged.
const (
	capEdgeWindow = 50.0
	capEdgeTurn   = 0.05
)

// The curved edges of C, G, J and Q have no kink, no bump and no straight
// stretch: where two curves meet, the tangent is continuous (the corners are the
// notches of the tail of Q and the tip of the wedge of the ball of J) and the
// curvature changes by at most a factor of 2 (but at the foot of a bracket and
// on the ball, whose radii are below 40 units); along each run of curves the
// contour turns one way only, but for the tail of Q, whose upper edge is
// concave, and every stretch of capEdgeWindow units turns by capEdgeTurn.
func TestCapCGJQEdges(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		for _, name := range []string{"C", "G", "J", "Q"} {
			path := capOutline(np, name)
			for ci, ct := range path {
				n := len(ct.Segs)
				for i := 0; i < n; i++ {
					if !ct.Segs[i].IsCurve || !ct.Segs[(i+1)%n].IsCurve {
						continue
					}
					a, b := segCubic(ct, i), segCubic(ct, i+1)
					ta, tb := a.tangent(1), b.tangent(0)
					if math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) > 0.01 {
						continue // a corner; Q and J count theirs in their own tests
					}
					ka, kb := a.curvature(1), b.curvature(0)
					if math.Max(math.Abs(ka), math.Abs(kb)) > 1.0/40 {
						continue // the foot of a bracket or the ball, parts with tests of their own
					}
					if ka*kb < 0 || math.Max(math.Abs(ka), math.Abs(kb)) > 2*math.Min(math.Abs(ka), math.Abs(kb))+1e-4 {
						t.Errorf("%s contour %d: the curvature jumps from %.5f to %.5f at (%.0f, %.0f)", name, ci, ka, kb, a.P3.X, a.P3.Y)
					}
				}
				least := math.Inf(1)
				var where outline.Point
				runs := capCurveRuns(ct)
				if name == "C" || name == "G" {
					runs = capWithoutHollow(runs, capBeak(np.params, capRound(np.params), capUpperBeakSpec(np.params)))
				}
				for _, run := range runs {
					s := make([]float64, len(run))
					for k := 1; k < len(run); k++ {
						s[k] = s[k-1] + math.Hypot(run[k].X-run[k-1].X, run[k].Y-run[k-1].Y)
					}
					if s[len(s)-1] < capEdgeWindow {
						continue
					}
					if l, at := slidingTurns(run, s, capEdgeWindow); l < least {
						least, where = l, at
					}
				}
				if least < capEdgeTurn {
					t.Errorf("%s: %s contour %d turns by only %.3f over %v units at (%.0f, %.0f)", np.name, name, ci, least, capEdgeWindow, where.X, where.Y)
				}
				if np.name == "regular" {
					t.Logf("%s contour %d: least turn %.3f over %v units at (%.0f, %.0f)", name, ci, least, capEdgeWindow, where.X, where.Y)
				}
			}
		}
	})
}

// S follows the rules of s at the size of the capitals: each long edge is one
// flowing curve, whose curvature changes sign once and which has no straight
// stretch longer than 120 units and no elbow; the bowls are round, with a
// vertical tangent and a radius of at least 0.9 of a quarter of the height of
// the letter at the left extreme of the upper and the right extreme of the lower
// bowl, and no radius below 0.4 of it in the counters; and the letter stands
// upright, with the centre of its ink within 12 units of the middle of its ink
// box across and within 9 units of it up and down.
func TestCapSFollowsTheRulesOfS(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		f := newCapSFrame(p)
		for k, edge := range sLongEdges(f) {
			var s, kap []float64
			length, prev := 0.0, edge[0].at(0)
			for _, c := range edge {
				for i := 0; i <= 40; i++ {
					q := c.at(float64(i) / 40)
					length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
					prev = q
					s = append(s, length)
					kap = append(kap, sCurvature(c, float64(i)/40))
				}
			}
			changes, flat, flatFrom, jump := 0, 0.0, -1.0, 0.0
			for i, v := range kap {
				if i > 0 {
					if v*kap[i-1] < 0 {
						changes++
					}
					jump = math.Max(jump, math.Abs(v-kap[i-1]))
				}
				if math.Abs(v) < 0.0013 { // 0.002 of s scaled to the height of S
					if flatFrom < 0 {
						flatFrom = s[i]
					}
					flat = math.Max(flat, s[i]-flatFrom)
				} else {
					flatFrom = -1
				}
			}
			if changes != 1 {
				t.Errorf("edge %d: the curvature changes sign %d times, want once", k, changes)
			}
			if flat > 120 {
				t.Errorf("edge %d: a straight stretch of %.0f units, want at most 120", k, flat)
			}
			if jump > 0.004 {
				t.Errorf("edge %d: the curvature changes by %.4f between neighbouring samples, want at most 0.004", k, jump)
			}
		}
		quarter := (p.CapHeight + 2*p.OvershootUpper) / 4
		for _, c := range []cubic{f.UpperE.Outer[0], f.LowerE.Outer[0]} {
			tg := c.tangent(1)
			if math.Abs(tg.X) > 1e-9 {
				t.Errorf("the tangent at a side extreme is %v, not vertical", tg)
			}
			if r := 1 / math.Abs(sCurvature(c, 1)); r < 0.9*quarter {
				t.Errorf("radius %.0f at a side extreme, want at least %.0f", r, 0.9*quarter)
			}
		}
		for _, e := range []sEdges{f.LowerE, f.UpperE} {
			for _, c := range e.Inner[:2] {
				for i := 0; i <= 40; i++ {
					if r := 1 / math.Abs(sCurvature(c, float64(i)/40)); r < 0.4*quarter {
						t.Errorf("a counter has a radius of %.0f, want at least %.0f", r, 0.4*quarter)
						return
					}
				}
			}
		}
		path := capOutline(np, "S")
		pts, _ := sPoints(path)
		var a, cx, cy float64
		for i, q := range pts {
			r := pts[(i+1)%len(pts)]
			cr := q.X*r.Y - r.X*q.Y
			a += cr / 2
			cx += (q.X + r.X) * cr / 6
			cy += (q.Y + r.Y) * cr / 6
		}
		cx, cy = cx/a, cy/a
		x0, y0, x1, y1 := path.BBox()
		if dx := cx - (x0+x1)/2; math.Abs(dx) > 12 {
			t.Errorf("the centre of the ink of S lies %.1f units right of the middle", dx)
		}
		if dy := cy - (y0+y1)/2; math.Abs(dy) > 9 {
			t.Errorf("the centre of the ink of S lies %.1f units above the middle", dy)
		}
		if np.name == "regular" {
			t.Logf("centre of the ink %.1f right of and %.1f above the middle", cx-(x0+x1)/2, cy-(y0+y1)/2)
		}
	})
}

// C and G are a little narrower than the round of O, and C is open.  The ink of
// C is 0.90 to 0.94 of that of O; that of G is 0.99 to 1.04 of O and 1.08 to 1.14
// of C, and the advance of G does not exceed that of O.  The opening of C, the
// largest vertical white between the upper and the lower ink on the vertical at
// 0.90 of its ink width, is 0.58 to 0.68 of the cap height.
func TestCapCGWidthsAndOpening(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		c, g, o := capOutline(np, "C"), capOutline(np, "G"), capOutline(np, "O")
		c0, _, c1, _ := c.BBox()
		g0, _, g1, _ := g.BBox()
		o0, _, o1, _ := o.BBox()
		if r := (c1 - c0) / (o1 - o0); r < 0.90 || r > 0.94 {
			t.Errorf("the ink of C is %.3f of that of O, want 0.90 to 0.94", r)
		}
		if r := (g1 - g0) / (o1 - o0); r < 0.99 || r > 1.04 {
			t.Errorf("the ink of G is %.3f of that of O, want 0.99 to 1.04", r)
		}
		if r := (g1 - g0) / (c1 - c0); r < 1.08 || r > 1.14 {
			t.Errorf("the ink of G is %.3f of that of C, want 1.08 to 1.14", r)
		}
		if _, ga := cachedOutline(np, "G"); ga > func() float64 { _, oa := cachedOutline(np, "O"); return oa }() {
			t.Errorf("the advance of G is %.0f, more than that of O", ga)
		}
		var gap float64
		v := vruns(c, c0+0.90*(c1-c0))
		for i := 0; i+1 < len(v); i++ {
			gap = math.Max(gap, v[i+1][0]-v[i][1])
		}
		if a := gap / cap; a < 0.58 || a > 0.68 {
			t.Errorf("the opening of C is %.3f of the cap height, want 0.58 to 0.68", a)
		}
	})
}

// The lower arc of G runs into the stem without a spur: the outer edge from the
// bottom extreme to the right-hand edge of the stem is one curve which turns one
// way (no inflection), leaves the extreme running east, ends on the vertical of
// the edge of the stem running straight up (no corner), and the rightmost ink
// of G does not retreat on the way up to the stem (no notch).  The arc keeps its
// weight: from the bottom extreme to the stem it is nowhere thinner than the
// thin part of O.
func TestCapGFoot(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		f := capRound(p)
		foot, j := capGFoot(p, f)
		if !oneWay(foot, 1) {
			t.Errorf("the outer edge of the lower arc of G turns both ways")
		}
		if d := foot.tangent(0); d.Y < -1e-6 || d.Y > 1e-6 || d.X < 0.999 {
			t.Errorf("the outer edge of the lower arc of G leaves the bottom extreme running %v, want east", d)
		}
		if d := foot.tangent(1); d.X > 1e-6 || d.X < -1e-6 || d.Y < 0.999 || math.Abs(foot.P3.X-j.X) > 1e-6 {
			t.Errorf("the outer edge of the lower arc of G ends running %v at x = %.2f, want straight up at %.2f", d, foot.P3.X, j.X)
		}
		g := capOutline(np, "G")
		prev := math.Inf(-1)
		for y := f.bottom + 1; y < j.Y; y++ {
			r := runs(g, y)
			if e := r[len(r)-1][1]; e < prev-0.05 {
				t.Errorf("at the height %.0f the rightmost ink of G retreats from %.1f to %.1f (a notch)", y, prev, e)
				break
			} else {
				prev = e
			}
		}
		xL := j.X - p.Caps.Stem
		var in []outline.Point
		for q := 0; q < 4; q++ {
			c := f.inner.cubic(q, false)
			for i := 0; i <= 200; i++ {
				in = append(in, c.at(float64(i)/200))
			}
		}
		least := math.Inf(1)
		for i := 0; i <= 200; i++ {
			q := foot.at(float64(i) / 200)
			if q.X > xL {
				break
			}
			d := math.Inf(1)
			for _, r := range in {
				d = math.Min(d, math.Hypot(q.X-r.X, q.Y-r.Y))
			}
			least = math.Min(least, d)
		}
		if least < p.Caps.RoundThin-0.5 {
			t.Errorf("the lower arc of G is %.1f thick at its thinnest, want at least %.1f", least, p.Caps.RoundThin)
		}
		if np.name == "regular" {
			t.Logf("lower arc of G: thinnest %.1f (thin part of O %.0f), joins the stem at %.0f", least, p.Caps.RoundThin, j.Y)
		}
	})
}
