// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The shared parts of the letters: the stem, the entry, the exit, the foot and
// the arch.  Positions are in the upright description, in which a position at
// the height ShearHeight is the position in the italic.

// Point is a point in font units.
type Point = stroke.Point

// auto marks a value which the rules of the stroke tool supply.
var auto = stroke.Auto

func pt(x, y float64) Point { return Point{X: x, Y: y} }

// across returns the thickness at right angles of a stem whose width
// measured horizontally in the italic is w.
func (p Params) across(w float64) float64 {
	k := p.slant()
	return w / math.Sqrt(1+k*k)
}

// stemNode returns a node of a straight stem whose centre line passes
// through x at the height y, travelling down.
func (p Params) stemNode(x, y float64) stroke.Node {
	return stroke.Node{P: pt(x, y), Dir: -90, W: p.across(p.Stem)}
}

// nib returns the pen of the curves, its major axis turned by Stress
// upright, which the shear turns to the angle of the sheared direction.
func (p Params) nib() *stroke.Nib {
	k := p.slant()
	a := p.Stress * math.Pi / 180
	return &stroke.Nib{Major: p.Curve, Minor: p.Thin, Angle: math.Atan2(math.Sin(a), math.Cos(a)+k*math.Sin(a)) * 180 / math.Pi}
}

// nibWidth returns the thickness which the pen gives to a stroke whose
// direction upright is dir (degrees).
func (p Params) nibWidth(dir float64) float64 {
	k := p.slant()
	d := dir * math.Pi / 180
	th := math.Atan2(math.Sin(d), math.Cos(d)+k*math.Sin(d))
	b := p.nib()
	u := b.Angle * math.Pi / 180
	// The normal of the stroke against the axes of the ellipse.
	s, c := math.Sin(th-u), math.Cos(th-u)
	return 2 * math.Hypot(s*b.Major/2, c*b.Minor/2)
}

// entry returns the nodes of the entry stroke into the top of a stem whose
// left edge stands at edge and whose top (the x-height, or the ascender) is
// top: the start left of the stem, which rises to the right, the round top
// and the join on the centre line of the stem, from where the stem runs
// straight.  The stroke starts in a round end.
func (p Params) entry(edge, top float64) []stroke.Node {
	up := top - p.XHeight
	top += p.TopLift
	return []stroke.Node{
		{P: pt(edge-p.EntryReach, p.EntryLow+up), Dir: p.EntryStartDir, K: auto, W: p.EndW, DW: auto},
		{P: pt(edge+p.EntryTopShift, p.EntryTop+up-p.EntryTopW/2), Dir: 0, K: auto, W: p.EntryTopW, DW: auto},
		p.vertNode(edge+p.Stem/2, top-p.EntryJoin, top-p.EntryVert, p.across(p.Stem)),
		p.stemNode(edge+p.Stem/2, top-p.EntryJoin),
	}
}

// vertical is the direction, upright, of a stroke which runs straight down
// in the italic.
func (p Params) vertical() float64 { return -90 + degrees(math.Atan(p.slant())) }

// vertNode returns the node where a curve which leaves or enters a stem
// whose centre line stands at c, at the height ys, runs vertically in the
// italic, at the height y, so that the extremes in x of both edges are
// nodes.  Its thickness lies between the stem's and w, which the stroke has
// at the height yw, as the height lies between; it is not changing there,
// so that both edges run vertically too.  In the italic the curve lies a
// little right of the stem's line below the stem and left of it above.
func (p Params) vertNode(c, ys, y, w float64) stroke.Node {
	return stroke.Node{P: pt(c-(y-ys)/12, y), Dir: p.vertical(), K: auto, W: w, DW: 0}
}

// exit returns the nodes of the exit at the foot of a stem whose left edge
// stands at edge: the stem straight down to ExitTurn, the turn at the
// baseline, a hairline whose outer edge overshoots the baseline, and the end
// node, without curvature, which ends in a round end.
func (p Params) exit(edge float64) []stroke.Node {
	c := edge + p.Stem/2
	return []stroke.Node{
		p.stemNode(c, p.ExitTurn),
		p.vertNode(c, p.ExitTurn, p.ExitVert, p.across(p.Stem)),
		{P: pt(c+p.ExitLowShift, -p.Overshoot+p.ExitLowW/2), Dir: 0, K: auto, W: p.ExitLowW, DW: 0},
		{P: pt(edge+p.Stem+p.ExitReach-p.ExitEndW/2, p.ExitTipHeight-p.ExitEndW/2), Dir: p.ExitAngle, K: p.ExitEndK, W: p.ExitEndW, DW: 0},
	}
}

// round is the end of an entry and of an exit: a blunt cut at right angles
// to the stroke.
var round = stroke.End{Kind: stroke.Cut}

// foot is the plain cut along the baseline which ends a stem without exit.
var foot = stroke.End{Kind: stroke.Cut, Line: true, Angle: 0}

// flatTop is the flat cut along the x-height at the top of a stem without
// entry.
var flatTop = stroke.End{Kind: stroke.Cut, Line: true, Angle: 0}

// hidden is the end of a stroke which lies inside another stroke.
var hidden = stroke.End{Kind: stroke.Cut, Hidden: true}

// stem returns the stroke of a stem whose left edge stands at edge, from the
// entry at top down to the exit if exit is set and to a plain foot otherwise.
func (p Params) stem(edge, top float64, exit bool) stroke.Stroke {
	s := stroke.Stroke{Nodes: p.entry(edge, top), Start: round}
	if exit {
		s.Nodes = append(s.Nodes, p.exit(edge)...)
		s.End = round
	} else {
		s.Nodes = append(s.Nodes, p.stemNode(edge+p.Stem/2, 0))
		s.End = foot
	}
	return s
}

// archNodes returns the nodes of the arch from the stem whose left edge
// stands at edge1 to the stem whose left edge stands at edge2, down to the
// node where it joins the right stem.  The stroke starts inside the left
// stem, so placed that its underside meets the stem's right edge at Branch,
// rises thinnest a little left of its peak,
// passes the shoulder node at ShoulderTop, where it runs vertically in the
// italic, and joins the stem at Shoulder; the pen gives the thickness
// between.
func (p Params) archNodes(edge1, edge2 float64) []stroke.Node {
	c1, c2 := edge1+p.Stem/2, edge2+p.Stem/2
	dir := p.ArchStartDir
	d := dir * math.Pi / 180
	t, n := pt(math.Cos(d), math.Sin(d)), pt(math.Sin(d), -math.Cos(d))
	// The underside, near the start a straight line in the direction of the
	// start, passes through the branch; the start lies 20 units back along it.
	under := pt(edge1+p.Stem, p.Branch)
	start := pt(under.X-n.X*p.ArchStartW/2-p.ArchStartBack*t.X, under.Y-n.Y*p.ArchStartW/2-p.ArchStartBack*t.Y)
	peak := pt(c1+p.ArchPeak*(c2-c1), p.XHeight+p.TopLift+p.ArchOver-p.nibWidth(0)/2)
	return []stroke.Node{
		{P: start, Dir: dir, K: auto, W: p.ArchStartW, DW: 0},
		{P: peak, Dir: 0, K: auto, W: auto, DW: 0},
		p.vertNode(c2, p.Shoulder, p.ShoulderTop, p.ShoulderW),
		p.stemNode(c2, p.Shoulder),
	}
}

// arch returns the arch from the stem whose left edge stands at edge1 to the
// stem whose left edge stands at edge2 (archNodes), and that stem down to its
// exit or its foot.
func (p Params) arch(edge1, edge2 float64, exit bool) stroke.Stroke {
	c2 := edge2 + p.Stem/2
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.archNodes(edge1, edge2), Start: hidden}
	if exit {
		s.Nodes = append(s.Nodes, p.exit(edge2)...)
		s.End = round
	} else {
		s.Nodes = append(s.Nodes, p.stemNode(c2, 0))
		s.End = foot
	}
	return s
}

// ballEnd returns the ball terminal of the italic about centre, placed by
// the rule of the spiral of the stroke tool: the end node slides along its
// direction until the spiral of the outer edge runs into the circle.  With
// side set (-1 left, +1 right) the ball moves sideways until the ink of the
// terminal reaches inkX.
func (p Params) ballEnd(centre Point, side int, inkX float64) stroke.End {
	return stroke.End{Kind: stroke.Ball, Ball: stroke.BallSpec{
		Centre: centre, R: p.Ball / 2,
		Neck: p.BallNeck, Open: p.BallOpen, TipK: auto,
		InkSide: side, InkX: inkX,
		Power: [2]float64{p.BallPowerMin, p.BallPowerMax},
	}}
}

// ballAtDepth returns the ball terminal of the italic placed by the end
// node of its stroke: its centre lies depth from the tangent of the outer
// edge there, on the right of the stroke (or on its left, for a negative
// depth), and as far ahead as the spiral of the outer edge puts it.
func (p Params) ballAtDepth(depth float64) stroke.End {
	e := p.ballEnd(Point{}, 0, 0)
	e.Ball.Depth = depth
	return e
}
