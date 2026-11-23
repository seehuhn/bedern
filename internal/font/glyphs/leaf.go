// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The leaf.
//
// The leaf is drawn in outline with a stroke of leafStroke strokes of the
// mathematical signs (Punct.MathThick).  Its edge is two equal arcs between
// a tip at the lower left and a tip at the upper right, one on each side of
// the line through the tips, each a cubic curve which leaves that line by
// leafTip degrees at both tips.  The inner edge runs at the distance of the
// stroke inside the outer one.  A midrib of the stroke runs from tip to tip,
// and four veins of the stroke leave it in turn: upright ones to the edge
// above it, level ones to the edge on its right.  The outline is the union
// of these parts: one outer contour and six counters.

// Shape constants of the leaf.
const (
	leafStroke = 0.85  // the stroke of the leaf, in strokes of the mathematical signs
	leafLength = 617.0 // the tips from the middle of the leaf
	leafBulge  = 0.5   // the edge from the midrib at its middle, in distances of a tip
	leafTip    = 48.0  // the angle between an edge and the midrib at a tip, degrees
	leafLean   = 39.0  // the angle of the midrib, degrees above the horizontal
	leafBottom = -60.0 // the lowest ink
	leafSide   = 20.0  // the side spaces
	leafRib    = 0.7   // the end of the midrib inside the tip of the outer edge, in the distance of the tip of the inner edge

	// leafVeins are the places at which the veins leave the midrib, from
	// its middle, in distances of a tip; the first and the third are
	// upright, the second and the fourth level.
	leafVeinA, leafVeinB, leafVeinC, leafVeinD = -0.4, -0.17, 0.12, 0.27

	// leafPieces is the number of curves of half an inner edge.
	leafPieces = 2
)

// leafEdge is an arc of the edge of the leaf from the lower tip to the upper
// one, with the side on which the inside of the leaf lies: 1 on the left of
// the direction of the arc, -1 on its right.
type leafEdge struct {
	c    cubic
	side float64
}

// inset returns the point at the distance d inside the arc at the parameter
// s, the direction of the arc there and the curvature of the curve at that
// distance.
func (e leafEdge) inset(s, d float64) (q, dir outline.Point, k float64) {
	at := e.c.at(s)
	dir = e.c.tangent(s)
	k = e.c.curvature(s)
	q = outline.Point{X: at.X - e.side*d*dir.Y, Y: at.Y + e.side*d*dir.X}
	return q, dir, k / (1 - e.side*k*d)
}

// find returns the parameter between lo and hi at which f changes its sign.
func (e leafEdge) find(lo, hi float64, f func(s float64) float64) float64 {
	neg := f(lo) < 0
	for range 50 {
		mid := (lo + hi) / 2
		if (f(mid) < 0) == neg {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// inner returns the curve at the distance d inside the arc, from the place
// where it leaves the midrib, the line through a in the direction u, to the
// place where it meets it again.
func (e leafEdge) inner(d float64, a, u outline.Point) chain {
	off := func(s float64) float64 {
		q, _, _ := e.inset(s, d)
		return u.X*(q.Y-a.Y) - u.Y*(q.X-a.X)
	}
	s0 := e.find(0, 0.5, off)
	var ch chain
	const n = 2 * leafPieces
	for i := range n {
		sa := s0 + (1-2*s0)*float64(i)/n
		sb := s0 + (1-2*s0)*float64(i+1)/n
		pa, ta, ka := e.inset(sa, d)
		pb, tb, kb := e.inset(sb, d)
		c, err := g2Cubic(pa, ta, ka, pb, tb, kb)
		if err != nil {
			panic(fmt.Sprintf("glyphs: the leaf: %v", err))
		}
		ch = append(ch, curveSeg(c))
	}
	return ch
}

// leaf draws the leaf.
func leaf(p design.Params) (outline.Path, float64) {
	t := math.Round(leafStroke * p.Punct.MathThick)
	beta, phi := degrees(leafLean), degrees(leafTip)
	u := dirOf(beta)
	a, b := xy(-leafLength*u.X, -leafLength*u.Y), xy(leafLength*u.X, leafLength*u.Y)
	// The handle with which a cubic curve bulges leafBulge tips.
	k := leafBulge * leafLength / (0.75 * math.Sin(phi))
	arc := func(sign float64) cubic {
		d0, d1 := dirOf(beta+sign*phi), dirOf(beta-sign*phi)
		return cubic{a, xy(a.X+k*d0.X, a.Y+k*d0.Y), xy(b.X-k*d1.X, b.Y-k*d1.Y), b}
	}
	upper, lower := leafEdge{arc(1), -1}, leafEdge{arc(-1), 1}

	// Each arc of the outer edge has a node at its middle.
	back := upper.c.reversed()
	outer := chain{
		curveSeg(lower.c.sub(0, 0.5)), curveSeg(lower.c.sub(0.5, 1)),
		curveSeg(back.sub(0, 0.5)), curveSeg(back.sub(0.5, 1))}
	// The two halves of the inner edge meet on the midrib.
	inUpper, inLower := upper.inner(t, a, u), lower.inner(t, a, u).reversed()
	inLower[0].P0 = inUpper.end()
	inLower[len(inLower)-1].P3 = inUpper.start()
	in := join(inUpper, inLower)

	// The midrib ends between the tips of the inner and the outer edge.
	rib := leafLength - leafRib*t/math.Sin(phi)
	parts := []chain{outer, in,
		taperBar(xy(-rib*u.X, -rib*u.Y), xy(rib*u.X, rib*u.Y), t)}

	// A vein ends on the middle line of the midrib and on the middle line
	// of the stroke of the edge.
	slope := math.Tan(beta)
	for _, f := range []float64{leafVeinA, leafVeinC} {
		x := f * leafLength * u.X
		top := func(x float64) float64 {
			s := upper.find(0, 1, func(s float64) float64 {
				q, _, _ := upper.inset(s, t/2)
				return q.X - x
			})
			q, _, _ := upper.inset(s, t/2)
			return q.Y
		}
		x0, x1 := x-t/2, x+t/2
		parts = append(parts, polygonChain([]outline.Point{
			xy(x0, x0*slope), xy(x1, x1*slope), xy(x1, top(x1)), xy(x0, top(x0))}))
	}
	for _, f := range []float64{leafVeinB, leafVeinD} {
		y := f * leafLength * u.Y
		right := func(y float64) float64 {
			s := lower.find(0, 1, func(s float64) float64 {
				q, _, _ := lower.inset(s, t/2)
				return q.Y - y
			})
			q, _, _ := lower.inset(s, t/2)
			return q.X
		}
		y0, y1 := y-t/2, y+t/2
		parts = append(parts, polygonChain([]outline.Point{
			xy(y0/slope, y0), xy(right(y0), y0), xy(right(y1), y1), xy(y1/slope, y1)}))
	}
	path := chainsPath(union(parts...)...)
	_, bottom, _, _ := path.BBox()
	path = shiftInk(path, 0, math.Round(leafBottom-bottom))
	return placeInk(path, design.PunctSides{Left: leafSide, Right: leafSide})
}
