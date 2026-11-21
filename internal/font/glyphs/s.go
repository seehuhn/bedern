// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter s.
//
// The s is one counter-clockwise contour without a counter: an upper arc, a
// spine and a lower arc.  Its two long edges are smooth S-shaped curves, and
// the letter is built from them.
//
// The centre.  The spine has a centre C, which is the inflection point of the
// middle line of the letter.  The spine runs through C at SSpineAngle below the
// horizontal, falling to the right, and is SSpineThick thick, measured at right
// angles to itself.  Its two edges pass through C - (S/2) n and C + (S/2) n,
// where n is the unit normal pointing up and to the right.  The upper half is
// smaller than the lower in the ratio r = SUpperWidth / SLowerWidth, and the
// construction makes the upper half the image of the lower half under a
// rotation through 180 degrees about C, with the distances from C scaled by r:
// the height of C above the bottom of the letter (the overshoot below the
// baseline included) is 1 / (1 + r) of the height of the letter, and the upper
// half, from C to the top, is r times the lower.  The distance of C from the
// left end of the lower half is chosen so that the centre of the upper half
// lies SUpperShift to the left of that of the lower.  Not scaled is everything
// which is a thickness or a distance from an extreme: the thin arcs, the sides
// of the bowls, the spine and the drops of the terminals.
//
// A half.  In the frame of the lower half, with C at the origin and the y axis
// pointing up, a half has two edges which run from the spine to the terminal.
//
//	The outer edge starts at C + (S/2) n, leaves it at the angle of the spine and
//	turns clockwise to the right extreme of the bowl, at x = Rx and at the
//	fraction sShape.SideY of the depth H of the half below C, where it runs
//	straight down, and turns on to the bottom extreme at y = -H, where it runs
//	to the left.
//	The inner edge starts at C - (S/2) n at the same angle and turns clockwise to
//	the extreme of the inner side, SSideThick left of the outer one and at the
//	fraction sShape.InnerSideY of the depth, and passes through the bottom
//	extreme SArcThick above the outer one, at the same x.
//	From the bottom extremes on, the half ends in a beak (sBeak): the outer
//	edge runs on as a thin arc to a corner on the vertical cut at x = -Lx,
//	SBeakDrop above the bottom, and a serif hangs from the corner to the end of
//	the terminal at the height of the drop (y = -H + Drop).
//
// Each edge is three cubic segments, with on-curve points at all extremes:
// spine to side extreme, side extreme to bottom, bottom to the corner (outer)
// or the face of the serif (inner).  The segment from the spine to the side extreme has its
// second control point at the corner where the line of the spine meets the
// vertical through the extreme, and its first on the same line, the fraction
// SpineHandle (InnerSpine) of the way to the corner.  The edge leaves the spine
// with zero curvature: the inflection of each long edge lies at the spine,
// where the edges of the two halves join with equal tangent and equal (zero)
// curvature, and the curvature grows steadily from there into the bowl.  The
// quarter from the side extreme to the bottom has the handle of the o at the
// bottom (Quarter, a fraction of the width), and at the side the handle which
// gives it the radius of curvature of the segment before it.  The side
// extremes lie a little below the middle of each half (SideY 0.45); the inner
// extremes lie lower (InnerSideY 0.56).
//
// The beaks.  Each terminal is the serif beak of S (capbeak.go, capss.go) at the
// size of the lowercase: the arc stays thin past the bottom (top) extreme and
// ends in a true corner on the vertical cut, SBeakDrop from the bottom (top),
// where it is SBeakArc thick; the outer edge comes into the corner as the
// fairest curve which goes on from the bowl with its direction and curvature.
// The serif hangs from the corner: the cut runs on to the end of the terminal,
// the flat end and the straight inner face leave it at design.ZBeakTilt, the
// face is SSerifThick from the cut halfway along the cut, and the inner edge of
// the arc leaves the counter at its bottom (top) extreme and turns into the face
// as a bracket, which leaves the face as that of an arm under a bar Hair thick.
// The sizes are those of S scaled by the ratio of the x-height to the cap
// height, with the serif and the arc thickened.  The s is a small S and has no
// ball terminal.
//
// The contour.  From C - (S/2) n it runs along the inner edge of the lower half
// to the face of its serif, down the face, across the flat end to the cut, along
// the cut, and back along the outer edge to C + (S/2) n.  It then does the same
// in the upper half, which is the lower half turned through 180 degrees: along
// its inner edge (the edge above the lower aperture) and back along its outer
// edge, which ends at C - (S/2) n where the contour started.  The long edge of
// the letter which starts at the top (the outer edge of the upper half) goes on
// along the inner edge of the lower half and is one S; the other long edge is
// the outer edge of the lower half and the inner edge of the upper.
//
// The sides.  The leftmost ink is the cut of the lower terminal and the
// rightmost ink the right extreme of the lower bowl, and the upper half lies
// inside them.  The advance is the lower width between the two side spaces:
// design.SAdvance.  Both sides are open sides.
//
// Shape constants.
type sShapeSpec struct {
	SpineHandle float64 // outer edge: handle at the spine, as a fraction of the way to the corner point
	InnerSpine  float64 // the same for the inner edge
	SideY       float64 // height of the outer side extreme below C, as a fraction of the depth H of the half
	InnerSideY  float64 // the same for the inner side extreme
	Quarter     float64 // handles of the quarters between extremes, as a fraction of their extents
	BottomPos   float64 // x of the bottom extreme: fraction of the way from the cut to the side extreme
}

var sShape = sShapeSpec{
	SpineHandle: 0.35,
	InnerSpine:  0.45,
	SideY:       0.45,
	InnerSideY:  0.56,
	Quarter:     0.58,
	BottomPos:   0.50,
}

// An sHalf gives the extents of one half in its own frame, in which C is the
// origin and the half hangs below it: the reach Rx of the outer edge to the
// right of C, the distance Lx of the cut to the left of C, the depth H of the
// bottom below C, and the drop of the end of the terminal below the top
// (above the bottom), which puts the end at y = -H + Drop.
type sHalf struct {
	Rx, Lx, H, Drop float64
}

// sEdges holds the curves of one half, in the frame of the half.
type sEdges struct {
	Inner    [3]cubic      // from C - (S/2) n to the top of the face of the serif
	Outer    [4]cubic      // from C + (S/2) n to the corner on the cut: the two quarters, the round to P and the hollow
	TipInner outline.Point // the inner end of the flat end, the foot of the face
	TipOuter outline.Point // the other end of the flat end, on the cut
	Corner   outline.Point // the end of Outer, on the cut
}

// quarter returns the curve from p0, which it leaves along the unit vector u0,
// to p3, which it meets along u3, with handles which are the fraction k of the
// extents of the displacement along u0 and u3.
func quarter(p0, p3, u0, u3 outline.Point, k float64) cubic {
	dx, dy := p3.X-p0.X, p3.Y-p0.Y
	e0 := math.Abs(dx*u0.X + dy*u0.Y)
	e3 := math.Abs(dx*u3.X + dy*u3.Y)
	return cubic{
		P0: p0,
		P1: outline.Point{X: p0.X + k*e0*u0.X, Y: p0.Y + k*e0*u0.Y},
		P2: outline.Point{X: p3.X - k*e3*u3.X, Y: p3.Y - k*e3*u3.Y},
		P3: p3,
	}
}

// sideQuarter returns the quarter from the side extreme p0, which it leaves
// running down, to the bottom extreme p3, which it meets running left, with
// the handle at the bottom the fraction k of the width and the handle at the
// side chosen so that the radius of curvature at p0 is r: the curve goes on
// from the segment before it without a jump of curvature.  For the handles
// k0 a and k b of the height a and the width b, the radius at p0 is
// 1.5 k0² a² / ((1 - k) b).
func sideQuarter(p0, p3 outline.Point, k, r float64) cubic {
	a, b := p0.Y-p3.Y, p0.X-p3.X
	k0 := math.Sqrt(r*b*(1-k)/1.5) / a
	k0 = math.Min(math.Max(k0, 0.3), 0.9)
	return cubic{
		P0: p0,
		P1: outline.Point{X: p0.X, Y: p0.Y - k0*a},
		P2: outline.Point{X: p3.X + k*b, Y: p3.Y},
		P3: p3,
	}
}

// endRadius returns the radius of curvature of c at its end.
func endRadius(c cubic) float64 {
	ax, ay := c.P3.X-c.P2.X, c.P3.Y-c.P2.Y
	bx, by := c.P2.X-c.P1.X, c.P2.Y-c.P1.Y
	return 1.5 * math.Pow(ax*ax+ay*ay, 1.5) / math.Abs(ax*by-ay*bx)
}

// edges returns the curves of a half.
func (sh sShapeSpec) edges(p design.Params, h sHalf, arc float64) sEdges {
	th := p.SSpineAngle * math.Pi / 180
	d := outline.Point{X: math.Cos(th), Y: -math.Sin(th)} // along the spine, falling to the right
	n := outline.Point{X: math.Sin(th), Y: math.Cos(th)}  // at right angles to it, up
	half := p.SSpineThick / 2
	var e sEdges

	xb := -h.Lx + sh.BottomPos*(h.Lx+h.Rx)

	// From the spine to the side extreme.  The segment leaves p0 along the
	// spine and arrives at the side extreme vertically.  Its second control
	// point is the corner where the line of the spine meets the vertical
	// through the extreme, so that the three first points lie on the line of
	// the spine and the curvature is zero at p0, where the edge has its inflection
	// at the centre of the spine.  The first handle is the fraction hs of the
	// way to the corner.
	fromSpine := func(p0 outline.Point, x, y, hs float64) cubic {
		corner := outline.Point{X: x, Y: p0.Y + (x-p0.X)*d.Y/d.X}
		return cubic{P0: p0, P1: lerp(p0, corner, hs), P2: corner, P3: outline.Point{X: x, Y: y}}
	}
	o0 := outline.Point{X: half * n.X, Y: half * n.Y}

	// The outer edge.
	e.Outer[0] = fromSpine(o0, h.Rx, -sh.SideY*h.H, sh.SpineHandle)
	bottom := outline.Point{X: xb, Y: -h.H}
	e.Outer[1] = sideQuarter(e.Outer[0].P3, bottom, sh.Quarter, endRadius(e.Outer[0]))

	// The inner edge.
	i0 := outline.Point{X: -half * n.X, Y: -half * n.Y}
	e.Inner[0] = fromSpine(i0, h.Rx-p.SSideThick, -sh.InnerSideY*h.H, sh.InnerSpine)
	bottomIn := outline.Point{X: xb, Y: -h.H + arc}
	e.Inner[1] = sideQuarter(e.Inner[0].P3, bottomIn, sh.Quarter, endRadius(e.Inner[0]))
	return e
}

// sBeak gives a half, whose edges end at the bottom extremes, its beak of the
// size s (s.Drop is the height of the corner above the bottom, s.Arc the
// thickness of the arc there; s.End is set from the drop of the half), with a
// serif thick wide and a bracket for a bar of the thickness bar (beakSized): the
// outer edge runs from the bottom extreme, with the direction and the
// curvature of the quarter before it, to the corner on the cut, and the inner
// edge leaves the counter at its bottom extreme with its direction and
// curvature and runs into the face of the serif.  The beak is drawn in the
// frame of the beaks of C, which is that of the half turned through half a
// turn.
func sBeak(e *sEdges, h sHalf, s capBeakSpec, thick, bar float64) error {
	turn := func(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} }
	turnC := func(c cubic) cubic { return cubic{turn(c.P0), turn(c.P1), turn(c.P2), turn(c.P3)} }
	in := e.Inner[1]
	a := beakArc{
		Top:  turn(e.Outer[1].P3),
		KTop: -e.Outer[1].curvature(1),
		Leave: func(float64) (outline.Point, outline.Point, float64) {
			return turn(in.P3), turn(in.tangent(1)), in.curvature(1)
		},
		Us: []float64{1},
		X:  h.Lx,
	}
	s.End = h.H - h.Drop
	b, err := a.beakSized(s, thick, bar)
	if err != nil {
		return err
	}
	e.Outer[2] = turnC(b.Outer).reversed()
	e.Outer[3] = turnC(b.Hollow).reversed()
	if b.Hollow == (cubic{}) {
		e.Outer[3] = cubic{}
	}
	e.Inner[2] = turnC(b.Inner)
	e.Corner, e.TipOuter, e.TipInner = turn(b.Corner), turn(b.TipOut), turn(b.TipIn)
	return nil
}

// An sFrame holds the centre of the spine and the two halves of s, from which
// the letter and the tests of its parts are made.
type sFrame struct {
	C            outline.Point
	Lower, Upper sHalf
	LowerE       sEdges
	UpperE       sEdges
}

// newSFrame returns the frame of s for the parameters.
func newSFrame(p design.Params) sFrame {
	var f sFrame
	top := p.XHeight + p.OvershootLower
	bottom := -p.OvershootLower
	r := p.SUpperWidth / p.SLowerWidth
	// The distance of C from the left end of the lower half, such that the
	// upper half, which reaches r times as far from C on both sides, is
	// centred SUpperShift to the left of the lower half.
	u := p.SLowerWidth/2 - p.SUpperShift/(1+r)
	f.C = outline.Point{X: p.SSideLeft + u, Y: (top + r*bottom) / (1 + r)}
	f.Lower = sHalf{Rx: p.SLowerWidth - u, Lx: u, H: f.C.Y - bottom, Drop: p.SLowerDrop}
	f.Upper = sHalf{Rx: r * (p.SLowerWidth - u), Lx: r * u, H: top - f.C.Y, Drop: p.SUpperDrop}
	f.LowerE = sShape.edges(p, f.Lower, p.SArcThick)
	f.UpperE = sShape.edges(p, f.Upper, p.SUpperArcThick)
	spec := capBeakSpec{Drop: p.SBeakDrop, Arc: p.SBeakArc, Lift: p.SBeakLift, Run: p.SBeakRun}
	for _, half := range []struct {
		e *sEdges
		h sHalf
	}{{&f.LowerE, f.Lower}, {&f.UpperE, f.Upper}} {
		if err := sBeak(half.e, half.h, spec, p.SSerifThick, p.Hair); err != nil {
			panic(fmt.Sprintf("glyphs: a beak of s: %v; glyphs.Check reports such parameters", err))
		}
	}
	return f
}

// reversed returns the curve run backwards.
func (c cubic) reversed() cubic { return cubic{c.P3, c.P2, c.P1, c.P0} }

// circuit adds one half to the path: along the inner edge to the top of the
// face of the serif, down the face, across the flat end, along the cut, and
// back along the outer edge.
// The upper half is the lower half turned through 180 degrees about C, which
// is the sign -1.
func (f sFrame) circuit(path *outline.Path, e sEdges, sign float64) {
	at := func(q outline.Point) (float64, float64) { return f.C.X + sign*q.X, f.C.Y + sign*q.Y }
	curve := func(c cubic) {
		x1, y1 := at(c.P1)
		x2, y2 := at(c.P2)
		x3, y3 := at(c.P3)
		path.CurveTo(x1, y1, x2, y2, x3, y3)
	}
	for _, c := range e.Inner {
		curve(c)
	}
	path.LineTo(at(e.TipInner))
	path.LineTo(at(e.TipOuter))
	path.LineTo(at(e.Corner))
	for i := 3; i >= 0; i-- {
		if e.Outer[i] == (cubic{}) {
			continue
		}
		curve(e.Outer[i].reversed())
	}
}

// s draws the lowercase letter s, see the comment above.
func s(p design.Params) (outline.Path, float64) {
	f := newSFrame(p)
	var path outline.Path
	x, y := f.C.X+f.LowerE.Inner[0].P0.X, f.C.Y+f.LowerE.Inner[0].P0.Y
	path.MoveTo(x, y)
	f.circuit(&path, f.LowerE, 1)
	f.circuit(&path, f.UpperE, -1)
	return path, design.SAdvance(p)
}
