// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The diagonal strokes of v, w and y.
//
// A diagonal is a straight band: a centre line with a direction and a thickness
// at right angles to it.  Its two edges are straight lines, the left-hand edge
// and the right-hand edge in the sense of the x axis (at any height the left-hand
// edge has the smaller x).  The lean of a diagonal is the tangent of its angle to
// the vertical, dx/dy along the centre line: positive for a stroke which leans to
// the right going up (the thin diagonal of v), negative for one which leans to the
// left (the thick one).  Horizontal lines cut a diagonal in a segment of the width
// Thick sqrt(1 + Lean^2), which is more than the thickness at right angles.
//
// Two diagonals meet at a vertex.  Their outer edges, the left-hand edge of the
// left diagonal and the right-hand edge of the right diagonal, meet in a point
// below the baseline; a flat, a horizontal line of the width Flat at the depth
// Depth below the baseline, cuts off the point.  The outer edges pass through the
// two ends of the flat and have the same lean in absolute value; the vertex is a
// mirror image of itself in the vertical line through the middle of the flat,
// apart from the thicknesses.  The inner edges, the right-hand edge of the left
// diagonal and the left-hand edge of the right one, meet in the crotch, a sharp
// inner corner above the flat whose angle is twice that of a diagonal.
//
// The head serif of a diagonal sits at its top, where a horizontal line at the
// height top cuts it off.  It is a slab on one side of the diagonal: its top lies on
// the line from the edge to the tip, the tip is Thick thick, and the underside runs
// flat from the tip and then into the edge along a bracket.  Over is the overhang,
// the horizontal distance from the edge to the tip at the level of the underside.
// The bracket is of the family of the bracket of a stem serif (footUp), adapted to
// the inclined edge.  The corner between the underside and the edge is obtuse where
// the edge leans away from the serif going down and acute where it leans towards
// it; on the acute side the bracket turns through the vertical and has an extreme
// in x, which is an on-curve point.  In v, w, x, y, k and the diagonal capitals the
// bracket is fairBracket: it leaves the underside the reach of design.DiagBracket
// from the edge, tangent to it, and meets the edge, tangent to it and with a
// curvature close to zero, a few times the rise of the stem bracket below the
// underside, one cubic on the obtuse side and two on the acute side.  Its reach is
// Reach on the obtuse side and ReachAcute on the acute side, lengthened or
// shortened as far as the flat of the underside at the tip needs to be between
// design.DiagBracketFlatMin and DiagBracketFlatMax long.  In the other letters which
// carry such serifs the bracket is stemBracket: one cubic which meets the edge a
// vertical run below the underside (Obtuse or Acute), with each handle half-way to
// the corner, and the reach is at most design.DiagBracketReachMax of the overhang.
//
// A serif on the left-hand edge extends to the left and one on the right-hand
// edge to the right.  headServif returns the chain from the point where the top
// meets the edge to the end of the bracket on the edge, in the direction down the
// edge; the contour of a letter which goes up the edge uses the chain reversed.
// The parts of a letter are the serif chains, the straight edges between them
// (lineChain), the flat, and lines at the top.

// An edge names one of the two edges of a diagonal.
type edge int

const (
	edgeLeft  edge = iota // the edge with the smaller x at every height
	edgeRight             // the edge with the larger x
)

// diagBracketHandle is the length of each handle of the bracket of a serif on a
// diagonal as a fraction of the distance from its end to the corner between the
// underside and the edge: one half, as in the bracket of a stem serif (footUp).
const diagBracketHandle = 0.5

// A diagonal is a straight stroke.
type diagonal struct {
	X, Y  float64 // a point on the centre line
	Lean  float64 // dx/dy along the centre line
	Thick float64 // the thickness at right angles to the centre line
}

// width returns the horizontal width of the diagonal.
func (d diagonal) width() float64 { return design.DiagWidth(d.Thick, d.Lean) }

// centreAt returns the position of the centre line at the height y.
func (d diagonal) centreAt(y float64) float64 { return d.X + d.Lean*(y-d.Y) }

// edgeAt returns the point of the edge e at the height y.
func (d diagonal) edgeAt(e edge, y float64) outline.Point {
	h := d.width() / 2
	if e == edgeLeft {
		h = -h
	}
	return outline.Point{X: d.centreAt(y) + h, Y: y}
}

// down returns the unit vector which points down the diagonal.
func (d diagonal) down() outline.Point {
	n := math.Hypot(d.Lean, 1)
	return outline.Point{X: -d.Lean / n, Y: -1 / n}
}

// diagFromEdge returns the diagonal of the lean lean and the thickness thick whose
// edge e passes through the point (x, y).
func diagFromEdge(e edge, x, y, lean, thick float64) diagonal {
	d := diagonal{Lean: lean, Thick: thick}
	h := d.width() / 2
	if e == edgeLeft {
		d.X = x + h
	} else {
		d.X = x - h
	}
	d.Y = y
	return d
}

// edgeMeet returns the point where the edge ea of the diagonal a meets the edge eb
// of b.  It panics if the edges are parallel.
func edgeMeet(a diagonal, ea edge, b diagonal, eb edge) outline.Point {
	// Each edge is x = x0 + lean y.
	pa, pb := a.edgeAt(ea, 0), b.edgeAt(eb, 0)
	if a.Lean == b.Lean {
		panic("glyphs: the edges of two parallel diagonals do not meet")
	}
	y := (pb.X - pa.X) / (a.Lean - b.Lean)
	return a.edgeAt(ea, y)
}

// lineChain returns the straight piece from a to b as a chain.
func lineChain(a, b outline.Point) chain { return chain{lineSeg(a, b)} }

// A vertex is a pair of diagonals which meet at the bottom, see above.
type vertex struct {
	Left, Right diagonal // the diagonal which leans to the right going down, and the other
	Flat, Depth float64  // the width of the flat and its depth below the baseline
	Centre      float64  // the middle of the flat
}

// newVertex returns the vertex at the centre cx, with a flat of the width flat
// depth below the baseline, whose outer edges have the lean lean (in absolute
// value) and whose diagonals have the thicknesses thickL and thickR.
func newVertex(cx, flat, depth, lean, thickL, thickR float64) vertex {
	return vertex{
		Left:  diagFromEdge(edgeLeft, cx-flat/2, -depth, -lean, thickL),
		Right: diagFromEdge(edgeRight, cx+flat/2, -depth, lean, thickR),
		Flat:  flat, Depth: depth, Centre: cx,
	}
}

// outerLeft returns the left end of the flat, where the outer edge of the left
// diagonal ends.
func (v vertex) outerLeft() outline.Point {
	return outline.Point{X: v.Centre - v.Flat/2, Y: -v.Depth}
}

// outerRight returns the right end of the flat.
func (v vertex) outerRight() outline.Point {
	return outline.Point{X: v.Centre + v.Flat/2, Y: -v.Depth}
}

// crotch returns the inner corner of the vertex, where the right-hand edge of the
// left diagonal meets the left-hand edge of the right one.
func (v vertex) crotch() outline.Point { return edgeMeet(v.Left, edgeRight, v.Right, edgeLeft) }

// flat returns the flat as a chain, from its left end to its right end.
func (v vertex) flat() chain { return lineChain(v.outerLeft(), v.outerRight()) }

// A diagSerif describes a head serif of a diagonal.
type diagSerif struct {
	Over    float64            // the overhang at the level of the underside
	Thick   float64            // the thickness at the tip
	Bracket design.DiagBracket // the size of the bracket
	Fair    bool               // the bracket is fairBracket (see above)
}

// The sizes of fairBracket: fairRunObtuse and fairRunAcute are the runs, the
// distances below the underside at which the bracket meets the edge, as multiples
// of the rise of the stem bracket; fairHandleObtuse is the place of the handle on
// the edge on the obtuse side, as a fraction of the way from the corner to the end;
// fairOffsetObtuse and fairOffsetAcute are the distances of the other handle from
// the line of the edge, on the white side; on the acute side that handle lies at
// most fairOffsetAcuteMax of the way up from the edge to the extreme.
const (
	fairRunObtuse      = 2.5
	fairRunAcute       = 2.2
	fairHandleObtuse   = 0.7
	fairOffsetObtuse   = 0.5
	fairOffsetAcute    = 0.3
	fairOffsetAcuteMax = 0.5
)

// stemBracket returns the bracket of the stem type from the point leave on the
// underside to the point end on the edge, where corner is the point where the
// underside meets the edge: one cubic with each handle half-way to the corner.
func stemBracket(leave, corner, end outline.Point) cubic {
	h := diagBracketHandle
	return cubic{
		P0: leave,
		P1: outline.Point{X: leave.X + h*(corner.X-leave.X), Y: leave.Y},
		P2: outline.Point{X: end.X + h*(corner.X-end.X), Y: end.Y + h*(corner.Y-end.Y)},
		P3: end,
	}
}

// fairBracket returns the bracket of a serif on the edge e of the diagonal d from
// the point leave on the underside to the edge, where corner is the point where the
// underside meets the edge, see the comment above.  On the obtuse side it is one
// cubic from leave to the edge fairRunObtuse times rise below the underside, whose
// first handle lies on the underside fairOffsetObtuse from the line of the edge and
// whose second handle lies on the edge fairHandleObtuse of the way from the corner.
// On the acute side it is two cubics which meet at the extreme X in x of the stem
// bracket of the reach of leave and the run acute: the first from leave to X, with
// each handle half-way to the corner of the underside and the vertical through X,
// and the second from X to the edge fairRunAcute times rise below the underside,
// whose first handle lies on the vertical through X fairOffsetAcute from the line
// of the edge and whose second handle lies on the edge, placed so that the two
// cubics have the same curvature at X.  Both brackets leave the underside tangent
// to it and meet the edge tangent to it, with a curvature of less than 1/400.
func (d diagonal) fairBracket(e edge, leave, corner outline.Point, acute bool, rise, run float64) chain {
	under := leave.Y
	if !acute {
		end := d.edgeAt(e, under-fairRunObtuse*rise)
		f := fairHandleObtuse
		u, v := unit(corner, leave), unit(corner, end)
		off := fairOffsetObtuse / math.Abs(u.X*v.Y-u.Y*v.X)
		return splitAtExtrema(cubic{
			P0: leave,
			P1: outline.Point{X: corner.X + off*u.X, Y: corner.Y + off*u.Y},
			P2: outline.Point{X: corner.X + f*(end.X-corner.X), Y: corner.Y + f*(end.Y-corner.Y)},
			P3: end,
		})
	}
	sb := stemBracket(leave, corner, d.edgeAt(e, under-run))
	ts := outline.CubicExtrema(sb.P0.X, sb.P1.X, sb.P2.X, sb.P3.X)
	if len(ts) == 0 {
		panic("glyphs: the bracket on the acute side of a diagonal has no extreme")
	}
	x := sb.at(ts[0])
	k := outline.Point{X: x.X, Y: under}
	first := cubic{
		P0: leave,
		P1: outline.Point{X: (leave.X + k.X) / 2, Y: under},
		P2: outline.Point{X: x.X, Y: (x.Y + under) / 2},
		P3: x,
	}
	below := d.edgeAt(e, corner.Y+(x.X-corner.X)/d.Lean)
	end := d.edgeAt(e, under-fairRunAcute*rise)
	off := min(fairOffsetAcute/math.Abs(unit(below, end).X), fairOffsetAcuteMax*(x.Y-below.Y))
	handle := outline.Point{X: below.X, Y: below.Y + off}
	on := func(beta float64) cubic {
		return cubic{P0: x, P1: handle, P2: outline.Point{X: below.X + beta*(end.X-below.X), Y: below.Y + beta*(end.Y-below.Y)}, P3: end}
	}
	k0, k1 := on(0).curvature(0), on(1).curvature(0)
	second := on((first.curvature(1) - k0) / (k1 - k0))
	return join(chain{curveSeg(first)}, splitAtExtrema(second))
}

// headServif returns the head serif on the edge e of the diagonal d, whose top is
// cut off at the height top: the chain starts where the top meets the edge, runs
// along the flat top to the tip, down the tip, and along the bracket to the edge.
func (d diagonal) headServif(e edge, top float64, s diagSerif) chain {
	side := 1.0
	if e == edgeLeft {
		side = -1
	}
	down := d.down()
	run, reach := s.Bracket.Obtuse, s.Bracket.Reach
	acute := side*down.X > 0 // the edge goes down towards the serif
	if acute {
		run, reach = s.Bracket.Acute, s.Bracket.ReachAcute
	}
	if s.Fair {
		reach = max(min(reach, s.Over-design.DiagBracketFlatMin), s.Over-design.DiagBracketFlatMax)
	} else {
		reach = math.Min(reach, design.DiagBracketReachMax*s.Over)
	}
	topPt := d.edgeAt(e, top)
	under := top - s.Thick
	corner := d.edgeAt(e, under)
	tipLower := outline.Point{X: corner.X + side*s.Over, Y: under}
	tipUpper := outline.Point{X: tipLower.X, Y: top}
	leave := outline.Point{X: corner.X + side*reach, Y: under}
	ch := chain{lineSeg(topPt, tipUpper), lineSeg(tipUpper, tipLower), lineSeg(tipLower, leave)}
	if s.Fair {
		return join(ch, d.fairBracket(e, leave, corner, acute, s.Bracket.Rise, run))
	}
	return join(ch, splitAtExtrema(stemBracket(leave, corner, d.edgeAt(e, under-run))))
}

// footServif returns the foot serif on the edge e of the diagonal d, whose foot is
// cut off by the baseline at the height base.  It is the head serif turned upside
// down with a flat upper side: the chain starts on the edge a run above the upper
// corner of the tip, runs along the bracket to the tip, down the tip, and back
// along the baseline to the edge, ending on it at the baseline.  A contour which
// comes down the edge uses the chain as it is, and one which goes up the edge uses
// it reversed.  The part is made from headServif for the diagonal mirrored in the
// baseline.  The overhang is measured at the level of the upper face, the one with
// the bracket.
func (d diagonal) footServif(e edge, base float64, s diagSerif) chain {
	m := diagonal{X: d.X, Y: 2*base - d.Y, Lean: -d.Lean, Thick: d.Thick}
	return m.headServif(e, base, s).flipY(base)
}
