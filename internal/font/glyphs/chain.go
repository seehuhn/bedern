// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// A piece is one segment of a chain: a straight line from P0 to P3 if line is
// set (P1 and P2 are then not used), and a cubic curve otherwise.
type piece struct {
	cubic
	line bool
}

// lineSeg returns the straight piece from a to b.
func lineSeg(a, b outline.Point) piece { return piece{cubic{a, a, b, b}, true} }

// curveSeg returns the curved piece c.
func curveSeg(c cubic) piece { return piece{cubic: c} }

// A chain is a run of connected pieces, each starting where the previous one
// ended, which a part of a letter returns instead of writing it to a path.  The
// chain can be shifted, mirrored in a vertical line or flipped in a horizontal
// line before it is written.
//
// The pieces are walked in the direction of a counter-clockwise contour.
// mirrorX and flipY reflect the chain and reverse it, so that the result is
// again walked counter-clockwise and starts at the image of the end of the
// original.  reflectX and reflectY reflect the chain and keep the order.
type chain []piece

// start returns the point where the chain starts.
func (c chain) start() outline.Point { return c[0].P0 }

// end returns the point where the chain ends.
func (c chain) end() outline.Point { return c[len(c)-1].P3 }

// startTangent returns the unit vector in which the chain leaves its start.
func (c chain) startTangent() outline.Point {
	p := c[0]
	if p.line {
		return unit(p.P0, p.P3)
	}
	for _, q := range []outline.Point{p.P1, p.P2, p.P3} {
		if q != p.P0 {
			return unit(p.P0, q)
		}
	}
	return outline.Point{}
}

// endTangent returns the unit vector in which the chain arrives at its end.
func (c chain) endTangent() outline.Point {
	p := c[len(c)-1]
	if p.line {
		return unit(p.P0, p.P3)
	}
	for _, q := range []outline.Point{p.P2, p.P1, p.P0} {
		if q != p.P3 {
			return unit(q, p.P3)
		}
	}
	return outline.Point{}
}

func unit(a, b outline.Point) outline.Point {
	n := math.Hypot(b.X-a.X, b.Y-a.Y)
	return outline.Point{X: (b.X - a.X) / n, Y: (b.Y - a.Y) / n}
}

// appendTo adds the pieces to the path.  The current point of the path must be
// the start of the chain.
func (c chain) appendTo(path *outline.Path) {
	for _, p := range c {
		if p.line {
			path.LineTo(p.P3.X, p.P3.Y)
		} else {
			path.CurveTo(p.args())
		}
	}
}

// addContour starts a new contour of the path at the start of the chain and
// adds the pieces to it.  It is for a chain which closes on itself, such as the
// counter of a bowl.
func (c chain) addContour(path *outline.Path) {
	path.MoveTo(c.start().X, c.start().Y)
	c.appendTo(path)
}

// connected reports whether every piece starts where the one before ended.
func (c chain) connected() error {
	for i := 1; i < len(c); i++ {
		if a, b := c[i-1].P3, c[i].P0; math.Abs(a.X-b.X) > 1e-9 || math.Abs(a.Y-b.Y) > 1e-9 {
			return fmt.Errorf("chain: piece %d starts at %v, but piece %d ends at %v", i, b, i-1, a)
		}
	}
	return nil
}

// join returns the chains one after the other as one chain.  It panics if a
// chain does not start where the one before ended.
func join(chains ...chain) chain {
	var all chain
	for _, c := range chains {
		all = append(all, c...)
	}
	if err := all.connected(); err != nil {
		panic(err)
	}
	return all
}

// mapPoints returns the chain with every point replaced by f of it, in the
// same order.
func (c chain) mapPoints(f func(outline.Point) outline.Point) chain {
	r := make(chain, len(c))
	for i, p := range c {
		r[i] = piece{cubic{f(p.P0), f(p.P1), f(p.P2), f(p.P3)}, p.line}
	}
	return r
}

// reversed returns the chain walked from its end to its start.
func (c chain) reversed() chain {
	r := make(chain, len(c))
	for i, p := range c {
		r[len(c)-1-i] = piece{cubic{p.P3, p.P2, p.P1, p.P0}, p.line}
	}
	return r
}

// shifted returns the chain moved by dx to the right and dy up.
func (c chain) shifted(dx, dy float64) chain {
	return c.mapPoints(func(q outline.Point) outline.Point { return outline.Point{X: q.X + dx, Y: q.Y + dy} })
}

// reflectX returns the mirror image of the chain in the vertical line x = at,
// with the pieces in the same order.  It reverses the orientation of a contour.
func (c chain) reflectX(at float64) chain {
	return c.mapPoints(func(q outline.Point) outline.Point { return outline.Point{X: 2*at - q.X, Y: q.Y} })
}

// reflectY returns the image of the chain in the horizontal line y = at, with
// the pieces in the same order.  It reverses the orientation of a contour.
func (c chain) reflectY(at float64) chain {
	return c.mapPoints(func(q outline.Point) outline.Point { return outline.Point{X: q.X, Y: 2*at - q.Y} })
}

// mirrorX returns the mirror image of the chain in the vertical line x = at.
// The image is reversed, so that it starts at the image of the end of the
// original and is counter-clockwise if the original is.
func (c chain) mirrorX(at float64) chain { return c.reflectX(at).reversed() }

// flipY returns the image of the chain in the horizontal line y = at.
// The image is reversed, so that it starts at the image of the end of the
// original and is counter-clockwise if the original is.
func (c chain) flipY(at float64) chain { return c.reflectY(at).reversed() }

// splitAtExtrema returns the curve as a chain, cut at its interior extremes in
// x and in y.  Every extreme is an on-curve point of the result, and the
// pieces join with a continuous tangent.
func splitAtExtrema(c cubic) chain {
	var ts []float64
	for _, axis := range []func(q outline.Point) float64{
		func(q outline.Point) float64 { return q.X },
		func(q outline.Point) float64 { return q.Y },
	} {
		for _, t := range outline.CubicExtrema(axis(c.P0), axis(c.P1), axis(c.P2), axis(c.P3)) {
			// Extremes within a unit of an end of the curve are ignored.
			if t > 1e-6 && t < 1-1e-6 && math.Hypot(c.at(t).X-c.P0.X, c.at(t).Y-c.P0.Y) >= 1 && math.Hypot(c.at(t).X-c.P3.X, c.at(t).Y-c.P3.Y) >= 1 {
				ts = append(ts, t)
			}
		}
	}
	sort.Float64s(ts)
	var ch chain
	prev := 0.0
	for _, t := range append(ts, 1) {
		if t-prev < 1e-6 {
			continue
		}
		ch = append(ch, curveSeg(c.sub(prev, t)))
		prev = t
	}
	return ch
}
