// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The helpers which measure a run of the outer edge of a contour: the sum of
// its turns in the wrong direction (to the right, along a counter-clockwise
// contour) and the curvature on both sides of a node.  They sample the curves
// of the contour exactly, and not its flattening.

// edgeSamples is the number of samples per segment of an edge.
const edgeSamples = 64

// contourNodes returns the points between the segments of the contour: node i
// is where segment i starts, and node len(Segs) is node 0 again.
func contourNodes(ct outline.Contour) []outline.Point {
	nodes := []outline.Point{ct.Start}
	for _, s := range ct.Segs {
		nodes = append(nodes, s.P3)
	}
	return nodes
}

// nodeNear returns the index of the node of the contour nearest to q, which
// must lie within one unit of it.
func nodeNear(t *testing.T, ct outline.Contour, q outline.Point) int {
	t.Helper()
	nodes := contourNodes(ct)
	best, bd := 0, math.Inf(1)
	for i := 0; i < len(ct.Segs); i++ {
		if d := math.Hypot(nodes[i].X-q.X, nodes[i].Y-q.Y); d < bd {
			best, bd = i, d
		}
	}
	if bd > 1 {
		t.Fatalf("no node of the contour within 1 of (%.1f, %.1f), the nearest is %.1f away", q.X, q.Y, bd)
	}
	return best
}

// segCubic returns segment i of the contour as a cubic.
func segCubic(ct outline.Contour, i int) cubic {
	i %= len(ct.Segs)
	a := ct.Start
	if i > 0 {
		a = ct.Segs[i-1].P3
	}
	s := ct.Segs[i]
	if !s.IsCurve {
		return cubic{a, a, s.P3, s.P3}
	}
	return cubic{a, s.P1, s.P2, s.P3}
}

// headings returns the direction of the contour, in radians, at the samples of
// the run from node a to node b (walking along the contour, round the end if
// need be).
func headings(ct outline.Contour, a, b int) []float64 {
	var h []float64
	n := len(ct.Segs)
	for i := a; i%n != b%n || i == a; i++ {
		c := segCubic(ct, i)
		for k := 0; k <= edgeSamples; k++ {
			tg := c.tangent(float64(k) / edgeSamples)
			h = append(h, math.Atan2(tg.Y, tg.X))
		}
		if i-a > n {
			break
		}
	}
	return h
}

// wrongWayTurn returns the sum, in radians and as a positive number, of the
// turns to the right which the contour makes on its way from the node nearest
// to from to the node nearest to to.  An edge of a counter-clockwise outline
// which is convex turns to the left only, and the sum is then zero.
func wrongWayTurn(t *testing.T, ct outline.Contour, from, to outline.Point) float64 {
	t.Helper()
	h := headings(ct, nodeNear(t, ct, from), nodeNear(t, ct, to))
	if len(h) < 2*edgeSamples {
		t.Fatalf("only %d samples between (%.1f, %.1f) and (%.1f, %.1f)", len(h), from.X, from.Y, to.X, to.Y)
	}
	var wrong float64
	for k := 1; k < len(h); k++ {
		if d := turn(h[k-1], h[k]); d < 0 {
			wrong -= d
		}
	}
	return wrong
}

// curvaturesAround returns the curvature of the contour at the node q and at the
// arc length dist before and after it: the one-sided curvatures at the node
// itself (before, after) and the curvatures dist away (before2, after2).
func curvaturesAround(t *testing.T, ct outline.Contour, q outline.Point, dist float64) (before, after, before2, after2 float64) {
	t.Helper()
	i := nodeNear(t, ct, q)
	n := len(ct.Segs)
	prev, next := segCubic(ct, i+n-1), segCubic(ct, i)
	before, after = prev.curvature(1), next.curvature(0)
	// Walk away from the node by the arc length dist.
	walk := func(c cubic, from float64, dir float64) float64 {
		s := 0.0
		last := c.at(from)
		for k := 1; k <= 4000; k++ {
			tt := from + dir*float64(k)/4000
			if tt < 0 || tt > 1 {
				return c.curvature(math.Max(0, math.Min(1, tt)))
			}
			p := c.at(tt)
			s += math.Hypot(p.X-last.X, p.Y-last.Y)
			last = p
			if s >= dist {
				return c.curvature(tt)
			}
		}
		return c.curvature(from + dir)
	}
	return before, after, walk(prev, 1, -1), walk(next, 0, 1)
}

// The outer edge of the terminal of a, c and r is one convex curve, from the
// highest point of the stroke to the far side of the ball (a and c) or from the
// far side of the ball to the notch on the stem (r).  The contour turns to the
// left only; the sum of its turns to the right is below edgeTolerance for the
// regular parameters and below edgeToleranceAll in the other parameter sets of
// the tests.
const (
	edgeTolerance    = 0.005
	edgeToleranceAll = 0.03
)

// The crest of the arm of r.  The outer edge of the arm is one curve through its
// highest point, which is the node at XHeight + ArchOvershoot where the curve of
// the terminal is cut at its extreme: the handles there are horizontal and point
// away from each other, and the curvature does not jump: the ratio of the
// curvature after the crest to the curvature before it, at the crest and 5 units
// from it, is within crestRatioLow and crestRatioHigh.
const (
	crestRatioLow  = 0.8
	crestRatioHigh = 1.25
	crestDistance  = 5.0
	// crestLevel is how far the handles at the crest may differ from its height.
	crestLevel = 1e-3
)

func TestRCrestIsSmooth(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _, _ := outlineOf("r", p)
	ct := path[0]
	var crest outline.Point
	for _, q := range contourNodes(ct) {
		if q.Y > crest.Y {
			crest = q
		}
	}
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
