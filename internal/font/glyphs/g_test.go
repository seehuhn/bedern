// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of g check what can go wrong in the construction of
// internal/glyphs/g.go: that the stroke from the link into the loop runs without
// a corner or a straight stretch, that the link swells steadily, that the ear
// stays a light circle on a thin stalk, that the pocket and the wedge stay open,
// that the loop stands balanced under the bowl, and that g keeps clear of its
// neighbours.  The structural checks are inherited (structure_test.go).

// gTestFrame returns the frame of g in the parameter set.
func gTestFrame(t *testing.T, np namedParams) gFrame {
	t.Helper()
	f, err := newGFrame(np.params)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// gSample returns n+1 points of the curve, with the tangent and the curvature at
// each.
func gSample(c cubic, n int) (pts, tans []outline.Point, kappa []float64) {
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		pts = append(pts, c.at(t))
		tans = append(tans, c.tangent(t))
		kappa = append(kappa, c.curvature(t))
	}
	return
}

// gRunChecks checks a run of pieces which the outline joins one after the other:
// that each piece starts in the direction in which the one before ends (within
// gJoinTol degrees), and that no stretch longer than gFlatMax is straight
// (curvature below gFlatCurv).
func gRunChecks(t *testing.T, set, what string, run chain) {
	t.Helper()
	for i := 1; i < len(run); i++ {
		a := run[i-1].cubic.tangent(1)
		b := run[i].cubic.tangent(0)
		if d := math.Abs(math.Atan2(a.X*b.Y-a.Y*b.X, a.X*b.X+a.Y*b.Y)) * 180 / math.Pi; d > gJoinTol {
			t.Errorf("%s: the %s has a corner of %.2f degrees at (%.0f, %.0f)", set, what, d, run[i].P0.X, run[i].P0.Y)
		}
	}
	flat, longest := 0.0, 0.0
	var prev outline.Point
	first := true
	for _, pc := range run {
		if pc.line {
			t.Errorf("%s: the %s has a straight piece from (%.0f, %.0f) to (%.0f, %.0f)", set, what, pc.P0.X, pc.P0.Y, pc.P3.X, pc.P3.Y)
			continue
		}
		pts, _, kappa := gSample(pc.cubic, 64)
		for i, q := range pts {
			if !first {
				if math.Abs(kappa[i]) < gFlatCurv {
					flat += math.Hypot(q.X-prev.X, q.Y-prev.Y)
					longest = math.Max(longest, flat)
				} else {
					flat = 0
				}
			}
			prev, first = q, false
		}
	}
	if longest > gFlatMax {
		t.Errorf("%s: the %s runs straight (curvature below %v) for %.0f units, want at most %v", set, what, gFlatCurv, longest, gFlatMax)
	}
}

const (
	gJoinTol  = 0.5        // degrees by which the tangent may turn where two pieces of a smooth run meet
	gFlatCurv = 1.0 / 3000 // a stretch with less curvature than this (radius above 3000 units) counts as straight
	gFlatMax  = 40.0       // the longest straight stretch allowed in the stroke from the link into the loop
)

// The stroke from the link into the loop is one pen movement: the outer edge
// from the notch under the bowl down the link and round the curl to the crotch
// where the left end of the loop meets it, the edge from the pocket down the
// link, along the top of the loop and round the loop to that crotch, and the
// edge of the counter from its corner under the link round to it again have no
// corner and no straight stretch.
func TestGStrokeFlows(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f := gTestFrame(t, np)
			gRunChecks(t, np.name, "left edge of the link", weld(
				f.a1.sub(0, 1),
				f.a2.sub(0, f.tJa),
			))
			gRunChecks(t, np.name, "edge from the pocket round the loop", weld(
				f.b1.sub(0, f.tPb).reversed(),
				f.b2.reversed().sub(0, 1),
				f.b3.reversed().sub(0, 1),
				f.b4.reversed().sub(0, 1),
				f.b5.reversed().sub(0, 1),
				f.b6.sub(0, f.tJb),
			))
			ct := f.counter()
			gRunChecks(t, np.name, "counter of the loop", ct)
		})
	}
}

// gPolyline returns points along the pieces, about one per unit.
func gPolyline(ch chain) []outline.Point {
	var pts []outline.Point
	for _, pc := range ch {
		n := int(math.Max(8, math.Hypot(pc.P3.X-pc.P0.X, pc.P3.Y-pc.P0.Y)))
		for i := 0; i <= n; i++ {
			pts = append(pts, pc.cubic.at(float64(i)/float64(n)))
		}
	}
	return pts
}

// gNearest returns the distance from q to the nearest of the points.
func gNearest(q outline.Point, pts []outline.Point) float64 {
	d := math.Inf(1)
	for _, r := range pts {
		d = math.Min(d, math.Hypot(q.X-r.X, q.Y-r.Y))
	}
	return d
}

// The link is thin where it leaves the bowl and swells as it turns into the top
// of the loop: along the edge from the pocket down the link to the top of the
// loop, the distance to the other edge of the stroke (the left edge of the link
// and the ceiling of the counter) never falls by more than gLinkDip, and it is
// at least a sixth more at the top of the loop than at the pocket.  At the pocket
// it is about the hairline, from Hair - 4 to Hair + 10.
func TestGLinkSwells(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f := gTestFrame(t, np)
			other := gPolyline(weld(f.a1.sub(0, 1), f.a2.sub(0, 1)))
			edge := gPolyline(weld(
				f.b1.sub(0, f.tPb).reversed(),
				f.b2.reversed().sub(0, 1),
			))
			// The thickness is taken from gLinkSkip below the corner of the pocket,
			// where the edges have left the bowl.
			var th []float64
			var at []outline.Point
			run := 0.0
			for i := 4; i < len(edge); i += 4 {
				run += math.Hypot(edge[i].X-edge[i-4].X, edge[i].Y-edge[i-4].Y)
				if run >= gLinkSkip {
					th = append(th, gNearest(edge[i], other))
					at = append(at, edge[i])
				}
			}
			first, last := th[0], th[len(th)-1]
			peak := th[0]
			for i, d := range th {
				if d < peak-gLinkDip {
					t.Errorf("%s: the link thins from %.3f to %.3f at (%.0f, %.0f)", np.name, peak, d, at[i].X, at[i].Y)
					break
				}
				peak = math.Max(peak, d)
			}
			if lo, hi := np.params.Hair-4, np.params.Hair+10; first < lo || first > hi {
				t.Errorf("%s: the link is %.1f thick at the pocket, want %v to %v", np.name, first, lo, hi)
			}
			if last < first*7/6 {
				t.Errorf("%s: the link is %.1f thick at the pocket and %.1f at the top of the loop, want at least a sixth more", np.name, first, last)
			}
			if np.name == "regular" {
				t.Logf("link: %.1f at the pocket, %.1f at the top of the loop", first, last)
			}
		})
	}
}

// gArea returns the area enclosed by the points, positive counter-clockwise.
func gArea(pts []outline.Point) float64 {
	var a float64
	for i, q := range pts {
		r := pts[(i+1)%len(pts)]
		a += q.X*r.Y - r.X*q.Y
	}
	return a / 2
}

// The ball of the ear is a circle; the ear is lighter than the bowl it hangs on
// and stays within the right end of the loop.  Its area outside the bowl, stalk
// and ball, is at most gEarAreaMax of the area of the ring of the bowl, the
// diameter of the ball at most gEarBallMax of the thickness of the side of the
// bowl, and its rightmost ink is not right of the right extreme of the loop.
func TestGEar(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f := gTestFrame(t, np)
			e := f.ear
			right := math.Inf(-1)
			for _, q := range gPolyline(e.ball()) {
				if d := math.Hypot(q.X-e.centre.X, q.Y-e.centre.Y) - e.r; math.Abs(d) > 0.05 {
					t.Errorf("%s: the ball is not a circle: (%.1f, %.1f) is %.2f off", np.name, q.X, q.Y, d)
					break
				}
				right = math.Max(right, q.X)
			}
			if lr := f.b4.end().X; right > lr {
				t.Errorf("%s: the ball reaches to %.1f, right of the right extreme of the loop at %.1f", np.name, right, lr)
			}
			q0 := f.outerB.cubic(0, true)
			ear := gPolyline(append(e.chain(), curveSeg(q0.sub(e.tLq, e.tUq).reversed())))
			earArea := gArea(ear)
			ring := gArea(gPolyline(chain{curveSeg(f.outerB.cubic(0, true)), curveSeg(f.outerB.cubic(1, true)), curveSeg(f.outerB.cubic(2, true)), curveSeg(f.outerB.cubic(3, true))})) -
				gArea(gPolyline(chain{curveSeg(f.innerB.cubic(0, true)), curveSeg(f.innerB.cubic(1, true)), curveSeg(f.innerB.cubic(2, true)), curveSeg(f.innerB.cubic(3, true))}))
			if earArea > gEarAreaMax*ring {
				t.Errorf("%s: the ear has an area of %.0f, more than %v of the ring of the bowl (%.0f)", np.name, earArea, gEarAreaMax, ring)
			}
			side := f.innerB.ex[2][0] - f.outerB.ex[2][0]
			if 2*e.r > gEarBallMax*side {
				t.Errorf("%s: the ball is %.0f across against a side of the bowl of %.0f, want at most %v of it", np.name, 2*e.r, side, gEarBallMax)
			}
			if np.name == "regular" {
				t.Logf("ear: area %.0f, %.3f of the ring of the bowl (%.0f); ball %.0f against the side %.0f", earArea, earArea/ring, ring, 2*e.r, side)
			}
		})
	}
}

// gLinkSkip is the length of the edge below the corner of the pocket which the
// measurement of the link leaves out: there the two edges still part into the
// bowl.
const gLinkSkip = 8.0

// gLinkDip is the greatest fall of the thickness of the link along its edge
// which still counts as steady growth: one unit.
const gLinkDip = 1.0

const (
	gEarAreaMax = 0.18 // the greatest area of the ear outside the bowl, as a fraction of the ring of the bowl
	gEarBallMax = 0.90 // the greatest diameter of the ball, as a fraction of the side of the bowl
)

// The pocket between the bowl and the top of the loop stays open at small sizes:
// the underside of the bowl, from the width of the link beyond the corner of the
// pocket round to its right extreme, keeps at least gPocketMin from the top of
// the loop.  The wedge of white between the stalk of the ear and its ball opens
// by at least gWedgeMin.
func TestGPocketAndWedge(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f := gTestFrame(t, np)
			q2, q3 := f.outerB.cubic(2, true), f.outerB.cubic(3, true)
			under := gPolyline(chain{curveSeg(q2.sub(f.tPq, 1)), curveSeg(q3)})
			top := gPolyline(append(f.b2.sub(0, 1), f.b3.sub(0, 1)...))
			gap, where := math.Inf(1), outline.Point{}
			corner := q2.at(f.tPq)
			for _, q := range under {
				if q.X < corner.X+np.params.GLinkThick {
					continue
				}
				if d := gNearest(q, top); d < gap {
					gap, where = d, q
				}
			}
			if gap < gPocketMin {
				t.Errorf("%s: the pocket is %.1f high at (%.0f, %.0f), want at least %v", np.name, gap, where.X, where.Y, gPocketMin)
			}
			e := f.ear
			open := e.part.Part.wedge()
			if open < gWedgeMin {
				t.Errorf("%s: the wedge between the stalk and the ball opens by %.1f degrees, want at least %v", np.name, open, gWedgeMin)
			}
			if np.name == "regular" {
				t.Logf("pocket %.1f, wedge %.1f degrees", gap, open)
			}
		})
	}
}

const (
	gPocketMin = 55.0 // the least height of the pocket between the bowl and the top of the loop
	gWedgeMin  = 20.0 // the least opening of the wedge between the stalk and the ball, degrees
)

// The loop stands under the bowl and the letter on its advance: the middle of
// the loop lies right of the middle of the bowl by no more than gLoopShiftMax of
// the width of the loop (the ear balances it), and the centre of the ink lies
// within gInkShiftMax of the advance from the middle of the advance.
func TestGBalance(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f := gTestFrame(t, np)
			path, adv := cachedOutline(np, "g")
			loopMid := (f.b5.start().X + f.b4.end().X) / 2
			bowlMid := (f.outerB.ex[0][0] + f.outerB.ex[2][0]) / 2
			w := f.b4.end().X - f.b5.start().X
			if s := loopMid - bowlMid; s < 0 || s > gLoopShiftMax*w {
				t.Errorf("%s: the middle of the loop is %.1f right of that of the bowl, want 0 to %.0f", np.name, s, gLoopShiftMax*w)
			}
			var area, mx float64
			for _, ct := range path {
				pts := ct.Flatten()
				for i, q := range pts {
					r := pts[(i+1)%len(pts)]
					c := q.X*r.Y - r.X*q.Y
					area += c
					mx += (q.X + r.X) * c
				}
			}
			cx := mx / (3 * area)
			if d := cx - adv/2; math.Abs(d) > gInkShiftMax*adv {
				t.Errorf("%s: the centre of the ink is %.1f from the middle of the advance, want at most %.1f", np.name, d, gInkShiftMax*adv)
			}
			if np.name == "regular" {
				t.Logf("loop %.1f right of the bowl; ink centre %.1f from the middle", loopMid-bowlMid, cx-adv/2)
			}
		})
	}
}

const (
	gLoopShiftMax = 0.10 // how far right of the bowl the middle of the loop may lie, as a fraction of its width
	gInkShiftMax  = 0.04 // how far from the middle of the advance the centre of the ink may lie, as a fraction of the advance
)

// g keeps clear of the descender of a y which follows it and of the tail of a j
// before it.  The tail of a following j reaches under the loop and touches it
// (TestJTailKeepsAwayFromWhatPrecedes); that pair needs a kern.
func TestGNeighbours(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			g, gAdv := cachedOutline(np, "g")
			y, _ := cachedOutline(np, "y")
			j, jAdv := cachedOutline(np, "j")
			for _, c := range []struct {
				pair        string
				left, right outline.Path
				adv         float64
			}{{"gy", g, y, gAdv}, {"jg", j, g, jAdv}} {
				d, a, _ := minDistanceBetween(c.left, shiftedPath(c.right, c.adv))
				if d < pairClearance {
					t.Errorf("%s: %s: nearest approach %.1f at (%.0f, %.0f), want at least %v", np.name, c.pair, d, a.X, a.Y, pairClearance)
				}
				if np.name == "regular" {
					t.Logf("%s: nearest approach %.1f", c.pair, d)
				}
			}
			if np.name == "regular" {
				d, a, _ := minDistanceBetween(g, shiftedPath(j, gAdv))
				t.Logf("gj: nearest approach %.1f at (%.0f, %.0f)", d, a.X, a.Y)
			}
		})
	}
}
