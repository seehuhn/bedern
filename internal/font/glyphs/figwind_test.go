// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the edges by which the figures run into their balls: the outer
// edge from the extreme of the round to the ball, and on 2, 6 and 9 the inner
// edge of the arch from the tip of the wedge over the top of the counter.  The
// curvature is sampled every figEdgeStep units of arc length.
const (
	figEdgeStep = 5.0

	// figWindSlack is how far, as a fraction of the curvature at the extreme of
	// the round, the curvature of an outer edge may go against its course: it
	// falls to one least value and then rises to the ball.  On 2, 3 and 5 the
	// least value is that at the extreme, within the slack, and the edge
	// tightens all the way; on 6 and 9 the hook opens first.
	figWindSlack = 0.06

	// figWindRatio is the largest ratio of the curvatures of two neighbouring
	// samples on an outer edge, and figWindBall the largest ratio between the
	// curvature where the edge meets the ball and that of the ball.
	figWindRatio = 1.15
	figWindBall  = 1.05

	// figArchTop is the largest ratio of the curvatures on the two sides of the
	// top of an arch, figArchSag the least ratio of the curvature anywhere on
	// one side of the top to the smaller of its values at the ends of that side,
	// and figArchSpike the largest ratio of a sample to the larger of its
	// neighbours.
	figArchTop   = 1.05
	figArchSag   = 0.5
	figArchSpike = 1.3

	// figMinPiece is the shortest curve on these edges, except those whose
	// curvature is everywhere at least figBallLike times that of the ball.
	figMinPiece = 15.0
	figBallLike = 0.8
)

// A figRun is a run of curves of a contour in the order in which an edge is
// walked, each walked from its P0 to its P3.
type figRun []cubic

// figSegs returns the segments of the contour as cubic curves, and whether each
// is a curve.
func figSegs(ct outline.Contour) ([]cubic, []bool) {
	var cs []cubic
	var curve []bool
	for i, s := range ct.Segs {
		cs = append(cs, segCubic(ct, i))
		curve = append(curve, s.IsCurve)
	}
	return cs, curve
}

// smoothJoin reports whether a runs into b without a corner.
func smoothJoin(a, b cubic) bool {
	ta, tb := a.tangent(1), b.tangent(0)
	return math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) < 0.01
}

// cubicLength returns the arc length of c.
func cubicLength(c cubic) float64 {
	var l float64
	prev := c.P0
	for i := 1; i <= 200; i++ {
		q := c.at(float64(i) / 200)
		l += math.Hypot(q.X-prev.X, q.Y-prev.Y)
		prev = q
	}
	return l
}

// samples returns the absolute curvature of the run every figEdgeStep units of
// arc length, from its start to its end.
func (r figRun) samples() []float64 {
	var ks []float64
	next, s := 0.0, 0.0
	for _, c := range r {
		prev := c.P0
		for i := 0; i <= 400; i++ {
			t := float64(i) / 400
			q := c.at(t)
			s += math.Hypot(q.X-prev.X, q.Y-prev.Y)
			prev = q
			if s >= next {
				ks = append(ks, math.Abs(c.curvature(t)))
				next += figEdgeStep
			}
		}
	}
	return ks
}

// reverse returns the run walked the other way.
func (r figRun) reverse() figRun {
	out := make(figRun, len(r))
	for i, c := range r {
		out[len(r)-1-i] = cubic{c.P3, c.P2, c.P1, c.P0}
	}
	return out
}

// figBallEdges finds the balls of the radius rad on the contour, as runs of
// curves with the curvature of the ball, and returns for each the outer edge,
// from the first node beyond the ball with a horizontal tangent to the ball,
// the inner edge from the tip of the wedge away from the ball up to the end of
// the contour (to be cut by the caller), and whether the outer edge meets the
// ball without a corner.
func figBallEdges(ct outline.Contour, rad float64) (outer, inner []figRun, smooth []bool) {
	cs, curve := figSegs(ct)
	n := len(cs)
	onBall := func(i int) bool {
		i = ((i % n) + n) % n
		if !curve[i] {
			return false
		}
		for _, t := range []float64{0, 0.5, 1} {
			if math.Abs(cs[i].curvature(t)*rad-1) > 0.02 {
				return false
			}
		}
		return true
	}
	at := func(i int) cubic { return cs[((i%n)+n)%n] }
	for i := 0; i < n; i++ {
		if !onBall(i) || onBall(i-1) {
			continue
		}
		j := i
		for onBall(j + 1) {
			j++
		}
		// The ball runs from segment i to segment j.  The outer edge joins it
		// smoothly on one side, the tip of the wedge is a corner on the other.
		var o, in figRun
		if smoothJoin(at(j), at(j+1)) {
			for k := j + 1; k < j+n; k++ {
				o = append(o, at(k))
				if tg := at(k).tangent(1); math.Abs(tg.Y) < 1e-6 {
					break
				}
			}
			o = o.reverse()
			for k := i - 1; k > i-n; k-- {
				in = append(in, at(k))
			}
			in = reverseEach(in) // walked away from the ball
		} else {
			if !smoothJoin(at(i-1), at(i)) {
				smooth = append(smooth, false)
			}
			for k := i - 1; k > i-n; k-- {
				o = append(o, at(k))
				if tg := at(k).tangent(0); math.Abs(tg.Y) < 1e-6 {
					break
				}
			}
			o = reverseEach(o).reverse()
			for k := j + 1; k < j+n; k++ {
				in = append(in, at(k))
			}
		}
		if len(smooth) < len(outer)+1 {
			smooth = append(smooth, true)
		}
		outer = append(outer, o)
		inner = append(inner, in)
	}
	return outer, inner, smooth
}

// reverseEach returns the run with each curve walked the other way, in the
// same order.
func reverseEach(r figRun) figRun {
	out := make(figRun, len(r))
	for i, c := range r {
		out[i] = cubic{c.P3, c.P2, c.P1, c.P0}
	}
	return out
}

// The outer edge of each stroke of 2, 3, 5, 6 and 9 which ends in a ball runs
// from the extreme of its round into the ball as a spiral: its curvature falls
// to one least value and then rises to that of the ball (on 2, 3 and 5 it rises
// all the way), within figWindSlack; no two neighbouring samples differ by more
// than the factor figWindRatio; the edge meets the circle of the ball without a
// corner and with a curvature within figWindBall of the ball's; and none of its
// curves is shorter than figMinPiece, but those in the curl of the ball.
func TestFigureOuterEdgesWindIn(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		rad := np.params.Figures.Ball / 2
		for _, c := range []struct {
			name  string
			balls int
			hook  bool
		}{{"two", 1, false}, {"three", 2, false}, {"five", 1, false}, {"six", 1, true}, {"nine", 1, true}} {
			outer, _, smooth := figBallEdges(capOutline(np, c.name)[0], rad)
			if len(outer) != c.balls {
				t.Errorf("%s: %s: %d balls on the outer contour, want %d", np.name, c.name, len(outer), c.balls)
				continue
			}
			for bi, run := range outer {
				where := run[0].P0
				ks := run.samples()
				k0 := ks[0]
				least := 0
				for i, k := range ks {
					if k < ks[least] {
						least = i
					}
				}
				if !c.hook && ks[least] < (1-figWindSlack)*k0 {
					t.Errorf("%s: %s: the outer edge from (%.0f, %.0f) falls from %.5f to %.5f before it winds in", np.name, c.name, where.X, where.Y, k0, ks[least])
				}
				for i := 1; i < len(ks); i++ {
					if i <= least && ks[i] > ks[i-1]+figWindSlack*k0 || i > least && ks[i] < ks[i-1]-figWindSlack*k0 {
						t.Errorf("%s: %s: the curvature of the outer edge from (%.0f, %.0f) turns back at %.0f units, from %.5f to %.5f", np.name, c.name, where.X, where.Y, float64(i)*figEdgeStep, ks[i-1], ks[i])
						break
					}
				}
				for i := 1; i < len(ks); i++ {
					if math.Max(ks[i], ks[i-1]) > figWindRatio*math.Min(ks[i], ks[i-1]) {
						t.Errorf("%s: %s: the curvature of the outer edge from (%.0f, %.0f) jumps from %.5f to %.5f at %.0f units", np.name, c.name, where.X, where.Y, ks[i-1], ks[i], float64(i)*figEdgeStep)
						break
					}
				}
				last := run[len(run)-1]
				if kb := math.Abs(last.curvature(1)) * rad; kb > figWindBall || kb < 1/figWindBall {
					t.Errorf("%s: %s: the outer edge meets the ball %d with %.3f times its curvature", np.name, c.name, bi, kb)
				}
				if !smooth[bi] {
					t.Errorf("%s: %s: the outer edge meets the ball %d in a corner", np.name, c.name, bi)
				}
				for _, cu := range run {
					if l := cubicLength(cu); l < figMinPiece && !ballLike(cu, rad) {
						t.Errorf("%s: %s: a curve of %.1f units on the outer edge at (%.0f, %.0f)", np.name, c.name, l, cu.P0.X, cu.P0.Y)
					}
				}
			}
		}
	})
}

// ballLike reports whether the curvature of c is everywhere at least figBallLike
// times that of a ball of the radius rad.
func ballLike(c cubic, rad float64) bool {
	for i := 0; i <= 20; i++ {
		if math.Abs(c.curvature(float64(i)/20))*rad < figBallLike {
			return false
		}
	}
	return true
}

// The inner edges of the arches of 2, 6 and 9 run from the tip of the wedge over
// the top of the counter, where a node lies at the extreme and the curvature on
// its two sides agrees within figArchTop, down to the extreme on the other side
// or to the notch of the stem.  On each side of the top the curvature nowhere
// falls below figArchSag times the smaller of its values at the ends of that
// side (no flat stretch), no sample exceeds the larger of its neighbours by more
// than figArchSpike, and no curve is shorter than figMinPiece.
func TestFigureArchesAreRound(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		rad := np.params.Figures.Ball / 2
		for _, name := range []string{"two", "six", "nine"} {
			_, inner, _ := figBallEdges(capOutline(np, name)[0], rad)
			if len(inner) != 1 {
				t.Errorf("%s: %s: %d balls on the outer contour, want 1", np.name, name, len(inner))
				continue
			}
			// The arch runs to the node at the top, the other side on to a
			// node with a vertical tangent or a corner.
			var arch, side figRun
			run := inner[0]
			k := 0
			for ; k < len(run); k++ {
				arch = append(arch, run[k])
				if math.Abs(run[k].tangent(1).Y) < 1e-6 {
					break
				}
			}
			for k++; k < len(run); k++ {
				side = append(side, run[k])
				if math.Abs(run[k].tangent(1).X) < 1e-6 || k+1 < len(run) && !smoothJoin(run[k], run[k+1]) {
					break
				}
			}
			if len(arch) == 0 || len(side) == 0 {
				t.Errorf("%s: %s: no top on the inner edge of the arch", np.name, name)
				continue
			}
			top := arch[len(arch)-1]
			ka, kb := math.Abs(top.curvature(1)), math.Abs(side[0].curvature(0))
			if math.Max(ka, kb) > figArchTop*math.Min(ka, kb) {
				t.Errorf("%s: %s: the curvature at the top of the arch, (%.0f, %.0f), is %.5f on one side and %.5f on the other", np.name, name, top.P3.X, top.P3.Y, ka, kb)
			}
			for _, part := range []figRun{arch, side} {
				ks := part.samples()
				ends := math.Min(math.Abs(part[0].curvature(0)), math.Abs(part[len(part)-1].curvature(1)))
				for i, kk := range ks {
					if kk < figArchSag*ends {
						t.Errorf("%s: %s: the inner edge from (%.0f, %.0f) flattens to %.5f at %.0f units, against %.5f at its ends", np.name, name, part[0].P0.X, part[0].P0.Y, kk, float64(i)*figEdgeStep, ends)
						break
					}
				}
				for i := 1; i+1 < len(ks); i++ {
					if ks[i] > figArchSpike*math.Max(ks[i-1], ks[i+1]) {
						t.Errorf("%s: %s: a spike of the curvature, %.5f, on the inner edge from (%.0f, %.0f)", np.name, name, ks[i], part[0].P0.X, part[0].P0.Y)
						break
					}
				}
				for _, cu := range part {
					if l := cubicLength(cu); l < figMinPiece {
						t.Errorf("%s: %s: a curve of %.1f units on the inner edge at (%.0f, %.0f)", np.name, name, l, cu.P0.X, cu.P0.Y)
					}
				}
			}
		}
	})
}
