// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter g.
//
// The g is the two-storey form, drawn as a round upper bowl with an ear and one
// stroke below it: the stroke leaves the bowl at its lower left as the link,
// swings down and round to the right as the top of the loop, runs round the
// right end, back along the bottom and up the left end, and closes on the link.
// The letter is one outer contour (counter-clockwise) and two counters
// (clockwise): the eye of the bowl and the counter of the loop.  Nothing
// overlaps, and every curve is cut at its extremes (splitAtExtrema).
//
// The bowl.  A super-ellipse of the superness BowlOuter (0.70) in a box
// GBowlWidth wide and GBowlHeight high whose top is the top of the o, with a
// counter of the superness BowlInner; the sides are GBowlThick thick, and the top
// and the bottom ThinLower.
//
// The link and the loop.  The edges of the stroke are named for the way the pen
// runs.  The left edge of the link (a1) leaves the outer edge of the bowl at a
// notch NotchUp above its bottom, in the direction LeaveDir, and runs convex down
// to its leftmost point, the neck, GLinkLeft right of the left extreme of the loop
// and GLinkDrop below the bottom of the bowl, where it is vertical.  From there
// (a2) it curls round to the right and becomes the ceiling of the counter of the
// loop, which it reaches at the station (Station of the width of the loop from its
// left extreme) in the direction TopSlope.  The right edge of the link (b1) meets
// the bowl at the corner of the pocket, GLinkThick plus PocketIn right of the
// notch, and runs down to its own leftmost point, NeckRise above the neck and
// GLinkThick plus NeckShift right of it; from there (b2) it curls round to the
// right into the upper edge of the top of the loop, GLoopTopThick above the
// ceiling at the station and parallel to it.  The two curls have the same form
// (gSweep).  The loop then runs round as quarters (gQuarter with the handle Ring)
// through its extremes: the right end GLoopRightThick thick at RightDrop of the
// depth below the baseline, the bottom ThinLower thick at the depth GLoopDepth and
// at BottomAt of the width, the left end GLoopLeftThick thick at LeftDrop of the
// depth.  The outer edge (b6) and the inner edge (a6) of the left end rise from
// its extreme and meet the curl of the link at JoinOuterAt and JoinInnerAt of a2,
// arriving in the directions JoinOuterDir and JoinInnerDir: the outer edge at a
// crotch, the inner edge at the corner of the counter under the link.
//
// The joins.  From the curl of the link round the loop, the edges of the counter
// (a2 to a6) and the outer edge (b2 to b6) are redrawn so that their curves share
// the curvature where they join (gFittedEdges): each curve is cut in the middle,
// the two short ends on either side of the station, where the stroke turns from
// the curl of the link into the round of the loop, are one curve, and the lengths
// of the handles are fitted while the points and the directions stay.  The
// straight parts of the link, a1 and b1, are kept as they are.
//
// The ear (gEarFrame).  A thin stalk with the shared ball terminal of ball.go at
// the size of the ear: a ball GEarBall across with its centre GEarDrop below the
// x-height and its right extreme GEarReach beyond the bowl, on a stalk GEarThick
// thick (gEarPart, with the curl gEarCurl).  The part is turned so that the crest
// of the outer edge lies GEarRise above the x-height.  The two edges of the
// stalk, the underside and the upper edge, are single cubic curves from the bowl,
// which they leave at gEarLeave on either side of a stalk standing on the middle
// of its stroke at gEarRoot, to the section of the part, where they have its
// direction and curvature (edgeIn, edgeOut); the underside meets the ball at the
// wedge, which opens on the side of the bowl.
//
// The outer contour runs counter-clockwise from the top of the bowl: over the
// left extreme of the bowl to the notch, down the left edge of the link and round
// its curl to the crotch of the left end, down the outer edge of the left end,
// along the bottom, up the right end, back along the top of the loop and up the
// right edge of the link to the corner of the pocket, along the underside of the
// bowl and up its right side to the underside of the stalk, out along it, round
// the ball, back along the upper edge of the stalk to the bowl and over the bowl
// to the start.
//
// Side kinds: sideLoopG on the left (GSideLeft to the left extreme of the loop)
// and sideEarG on the right (GSideRight from the rightmost ink, the right end of
// the loop; design.GAdvance).

// gShapeSpec holds the shape constants of g.
type gShapeSpec struct {
	BowlOuter, BowlInner float64 // superness of the outer and the inner contour of the bowl

	NotchUp              float64 // the left edge of the link leaves the outer edge of the bowl at a notch this far above the bottom of the bowl
	LeaveDir             float64 // the direction in which it leaves, degrees
	LeaveStart, LeaveEnd float64 // its handles at the notch and at the neck, as fractions of the chord
	NeckRise, NeckShift  float64 // the leftmost point of the right edge of the link lies this far above that of the left edge, and this much more than GLinkThick to its right

	Station             float64 // the station on the top of the loop, as a fraction of the width of the loop from its left extreme
	TopSlope            float64 // the direction of both edges of the top stroke at the station, degrees (negative: falling to the right)
	SweepNeck, SweepTop float64 // handles of the edges from the link to the station at the neck and at the station, as fractions of the projection of the chord
	Ring                float64 // handles of the quarters of the loop, as fractions of the projection of the chord

	LeftDrop, RightDrop float64 // the heights of the left and the right extreme of the loop below the baseline, as fractions of the depth
	BottomAt            float64 // the bottom of the loop, as a fraction of its width from the left extreme
	BottomShift         float64 // the bottom of the counter lies this far right of the bottom of the outer edge

	JoinOuterAt, JoinInnerAt   float64 // parameters of the left edge of the link (from the neck to the station) at which the outer and the inner edge of the left end of the loop aim
	JoinOuterDir, JoinInnerDir float64 // directions in which they arrive there, degrees
	JoinExtra                  float64 // they are drawn this far beyond, to cross the edge
	JoinRise, JoinEnd          float64 // their handles at the extreme (fraction of the rise) and at the join (fraction of the chord)

	PocketIn, PocketDir, PocketExtra float64 // the right edge of the link aims at the point of the outer edge of the bowl GLinkThick plus PocketIn to the right of the notch, arriving in this direction (degrees), drawn on this far beyond
	PocketStart, PocketEnd           float64 // its handles at the neck and at the bowl, as fractions of the chord

}

var gShape = gShapeSpec{
	BowlOuter: 0.70,
	BowlInner: 0.715,

	NotchUp:    50,
	LeaveDir:   -115,
	LeaveStart: 0.35,
	LeaveEnd:   0.40,
	NeckRise:   20,
	NeckShift:  2,

	Station:   0.70,
	TopSlope:  -2,
	SweepNeck: 0.80,
	SweepTop:  0.45,
	Ring:      0.56,

	LeftDrop:    0.45,
	RightDrop:   0.40,
	BottomAt:    0.48,
	BottomShift: 4,

	JoinOuterAt:  0.34,
	JoinInnerAt:  0.52,
	JoinOuterDir: 25,
	JoinInnerDir: 0,
	JoinExtra:    12,
	JoinRise:     0.60,
	JoinEnd:      0.30,

	PocketIn:    -12,
	PocketDir:   80,
	PocketExtra: 12,
	PocketStart: 0.40,
	PocketEnd:   0.30,
}

// gDeg returns the unit vector in the direction a, in degrees.
func gDeg(a float64) outline.Point {
	return outline.Point{X: math.Cos(a * math.Pi / 180), Y: math.Sin(a * math.Pi / 180)}
}

// gHandled returns the cubic from p0, leaving in the direction d0 with a handle of
// length h0, to p3, arriving in the direction d3 with a handle of length h3.
func gHandled(p0, d0 outline.Point, h0 float64, p3, d3 outline.Point, h3 float64) cubic {
	return cubic{p0,
		outline.Point{X: p0.X + h0*d0.X, Y: p0.Y + h0*d0.Y},
		outline.Point{X: p3.X - h3*d3.X, Y: p3.Y - h3*d3.Y},
		p3}
}

// gQuarter returns the cubic from p0 to p3 which leaves in the direction d0 and
// arrives in the direction d3, with handles k times the projection of the chord
// on each direction: for directions along the axes the quarter of a
// super-ellipse.
func gQuarter(p0, d0, p3, d3 outline.Point, k float64) cubic {
	dx, dy := p3.X-p0.X, p3.Y-p0.Y
	h0 := k * math.Abs(dx*d0.X+dy*d0.Y)
	h3 := k * math.Abs(dx*d3.X+dy*d3.Y)
	return gHandled(p0, d0, h0, p3, d3, h3)
}

// gSweep returns the edge from the neck of the link, which it leaves downwards,
// to the station on the top of the loop, where it arrives in the direction u.
func gSweep(neck, st, u outline.Point) cubic {
	dx, dy := st.X-neck.X, st.Y-neck.Y
	return gHandled(neck, gDown, gShape.SweepNeck*math.Abs(dy), st, u, gShape.SweepTop*math.Abs(dx*u.X+dy*u.Y))
}

// gBeyond returns the point d further than q in the direction u.
func gBeyond(q, u outline.Point, d float64) outline.Point {
	return outline.Point{X: q.X + d*u.X, Y: q.Y + d*u.Y}
}

var (
	gUp    = outline.Point{X: 0, Y: 1}
	gDown  = outline.Point{X: 0, Y: -1}
	gLeft  = outline.Point{X: -1, Y: 0}
	gRight = outline.Point{X: 1, Y: 0}
)

// A gFrame holds the curves from which g is drawn.  The edges of the stroke are
// named for the way the pen runs, from the bowl down the link, along the top of
// the loop, round its right end, along the bottom and up the left end: the left
// edge of the link becomes the edge of the counter of the loop (a1 to a6), the
// right edge of the link becomes the outer edge of the loop (b1 to b6, kept here in
// the direction of the outer contour, which is the other way).
type gFrame struct {
	outerB, innerB superShape

	a1, a2, a3, a4, a5, a6 gEdge
	b1, b2, b3, b4, b5, b6 gEdge

	tJa, tJb float64 // the outer edge of the left end (b6, from its extreme) meets a2: the parameters on a2 and on b6
	tKa, tKb float64 // the inner edge of the left end (a6) meets a2: the parameters on a2 and on a6
	tPb, tPq float64 // the right edge of the link (b1) meets the bowl: the parameters on b1 and on the quarter 2 of the bowl

	tQ  float64 // the left edge of the link leaves the quarter 2 of the bowl
	ear gEarFrame
}

// newGFrame returns the frame of g for the parameters, or an error which says
// which parts do not meet.
func newGFrame(p design.Params) (gFrame, error) {
	var f gFrame
	sh := gShape
	top, bottom := design.GBowlTop(p), design.GBowlBottom(p)
	bl, br := design.GBowlLeft(p), design.GBowlRight(p)
	f.outerB = newSuperShape(bl, bottom, br, top, sh.BowlOuter)
	f.innerB = newSuperShape(bl+p.GBowlThick, bottom+p.ThinLower, br-p.GBowlThick, top-p.ThinLower, sh.BowlInner)

	ll, lr := design.GLoopLeft(p), design.GLoopRight(p)
	w, depth := p.GLoopWidth, p.GLoopDepth
	neckA := outline.Point{X: ll + p.GLinkLeft, Y: bottom - p.GLinkDrop}
	neckB := outline.Point{X: neckA.X + p.GLinkThick + sh.NeckShift, Y: neckA.Y + sh.NeckRise}
	xs := ll + sh.Station*w
	stA := outline.Point{X: xs, Y: p.GLoopFloor - p.GLoopTopThick}
	stB := outline.Point{X: xs, Y: p.GLoopFloor}
	slope := gDeg(sh.TopSlope)
	back := outline.Point{X: -slope.X, Y: -slope.Y}
	rOut := outline.Point{X: lr, Y: -sh.RightDrop * depth}
	rIn := outline.Point{X: lr - p.GLoopRightThick, Y: rOut.Y}
	bOut := outline.Point{X: ll + sh.BottomAt*w, Y: -depth}
	bIn := outline.Point{X: bOut.X + sh.BottomShift, Y: -depth + p.ThinLower}
	lOut := outline.Point{X: ll, Y: -sh.LeftDrop * depth}
	lIn := outline.Point{X: ll + p.GLoopLeftThick, Y: lOut.Y}

	q2 := f.outerB.cubic(2, true)
	var a1, a2, a3, a4, a5, a6, b1, b2, b3, b4, b5, b6 cubic
	f.tQ = q2.tAtY(bottom + sh.NotchUp)
	{
		q := q2.at(f.tQ)
		chord := math.Hypot(neckA.X-q.X, neckA.Y-q.Y)
		a1 = gHandled(q, gDeg(sh.LeaveDir), sh.LeaveStart*chord, neckA, gDown, sh.LeaveEnd*chord)
	}
	a2 = gSweep(neckA, stA, slope)
	a3 = gQuarter(stA, slope, rIn, gDown, sh.Ring)
	a4 = gQuarter(rIn, gDown, bIn, gLeft, sh.Ring)
	a5 = gQuarter(bIn, gLeft, lIn, gUp, sh.Ring)
	{
		u := gDeg(sh.JoinInnerDir)
		end := gBeyond(a2.at(sh.JoinInnerAt), u, sh.JoinExtra)
		a6 = gHandled(lIn, gUp, sh.JoinRise*(end.Y-lIn.Y), end, u, sh.JoinEnd*math.Hypot(end.X-lIn.X, end.Y-lIn.Y))
	}

	// The outer edge of the loop, in the direction of the outer contour: from the
	// left extreme over the bottom and the right end, along the top and up the
	// right edge of the link to the bowl.  b6 is drawn from the left extreme up to
	// the join.
	{
		u := gDeg(sh.JoinOuterDir)
		end := gBeyond(a2.at(sh.JoinOuterAt), u, sh.JoinExtra)
		b6 = gHandled(lOut, gUp, sh.JoinRise*(end.Y-lOut.Y), end, u, sh.JoinEnd*math.Hypot(end.X-lOut.X, end.Y-lOut.Y))
	}
	b5 = gQuarter(lOut, gDown, bOut, gRight, sh.Ring)
	b4 = gQuarter(bOut, gRight, rOut, gUp, sh.Ring)
	b3 = gQuarter(rOut, gUp, stB, back, sh.Ring)
	b2 = gSweep(neckB, stB, slope).reversed()
	{
		u := gDeg(sh.PocketDir)
		aim := q2.at(q2.tAtX(a1.P0.X + p.GLinkThick + sh.PocketIn))
		end := gBeyond(aim, u, sh.PocketExtra)
		chord := math.Hypot(end.X-neckB.X, end.Y-neckB.Y)
		b1 = gHandled(neckB, gUp, sh.PocketStart*chord, end, u, sh.PocketEnd*chord)
	}

	// The joins of the edges from the curl of the link round the loop are made
	// smooth (gFittedEdges), each edge in the direction of the pen; the straight
	// parts of the link, a1 and b1, keep their form.
	as := gFittedEdges([]cubic{a2, a3, a4, a5, a6})
	f.a1, f.a2, f.a3, f.a4, f.a5, f.a6 = gEdge{a1}, as[0], as[1], as[2], as[3], as[4]
	bs := gFittedEdges([]cubic{b2.reversed(), b3.reversed(), b4.reversed(), b5.reversed(), b6})
	f.b1, f.b2, f.b3, f.b4, f.b5, f.b6 = gEdge{b1}, bs[0].reversed(), bs[1].reversed(), bs[2].reversed(), bs[3].reversed(), bs[4]

	var ok bool
	if f.tJa, f.tJb, ok = f.a2.crossing(f.b6); !ok {
		return f, fmt.Errorf("the outer edge of the left end of the loop does not meet the link")
	}
	if f.tKa, f.tKb, ok = f.a2.crossing(f.a6); !ok {
		return f, fmt.Errorf("the inner edge of the left end of the loop does not meet the link")
	}
	if !(f.tJa < f.tKa) {
		return f, fmt.Errorf("the left end of the loop meets the link the wrong way round")
	}
	if f.tPb, f.tPq, ok = f.b1.crossing(gEdge{q2}); !ok {
		return f, fmt.Errorf("the right edge of the link does not meet the bowl")
	}

	var err error
	if f.ear, err = newGEarFrame(p, f); err != nil {
		return f, err
	}
	return f, nil
}

// gFitted returns the curves, which join end to end with a common direction,
// with the lengths of their handles changed so that they also share the
// curvature where they join: of all such handles, none shorter than
// gFitMinHandle unless it was, those closest in the least squares sense to the
// given ones and, where curve i-1 meets curve i, to the curvature target[i] if
// there is one, and at the start of the first curve and the end of the last to
// target[0] and target[n] (Levenberg and Marquardt's method, with the joins
// weighted by gFitWeight and the curvatures by gFitMid).  The points and the directions
// stay.
func gFitted(cs []cubic, target map[int]float64) []cubic {
	n := len(cs)
	h0 := make([]float64, 2*n)
	for i, c := range cs {
		h0[2*i] = math.Hypot(c.P1.X-c.P0.X, c.P1.Y-c.P0.Y)
		h0[2*i+1] = math.Hypot(c.P3.X-c.P2.X, c.P3.Y-c.P2.Y)
	}
	build := func(h []float64) []cubic {
		out := make([]cubic, n)
		for i, c := range cs {
			out[i] = gHandled(c.P0, c.tangent(0), h[2*i], c.P3, c.tangent(1), h[2*i+1])
		}
		return out
	}
	resid := func(h []float64) []float64 {
		r := make([]float64, 0, 3*n)
		for i := range h {
			r = append(r, h[i]-h0[i])
		}
		b := build(h)
		for i := 1; i < n; i++ {
			r = append(r, gFitWeight*(b[i-1].curvature(1)-b[i].curvature(0)))
		}
		for i := 1; i < n; i++ {
			if k, ok := target[i]; ok {
				r = append(r, gFitMid*(b[i-1].curvature(1)-k))
			}
		}
		if k, ok := target[0]; ok {
			r = append(r, gFitMid*(b[0].curvature(0)-k))
		}
		if k, ok := target[n]; ok {
			r = append(r, gFitMid*(b[n-1].curvature(1)-k))
		}
		return r
	}
	sq := func(r []float64) float64 {
		var s float64
		for _, v := range r {
			s += v * v
		}
		return s
	}
	h := append([]float64(nil), h0...)
	r := resid(h)
	lambda := 1e-3
	m := len(h)
	for iter := 0; iter < 200; iter++ {
		// The Jacobian, by forward differences.
		jac := make([][]float64, len(r))
		for k := range jac {
			jac[k] = make([]float64, m)
		}
		for j := 0; j < m; j++ {
			hh := append([]float64(nil), h...)
			hh[j] += 1e-5
			rr := resid(hh)
			for k := range r {
				jac[k][j] = (rr[k] - r[k]) / 1e-5
			}
		}
		a := make([][]float64, m)
		g := make([]float64, m)
		for i := 0; i < m; i++ {
			a[i] = make([]float64, m)
			for j := 0; j < m; j++ {
				for k := range r {
					a[i][j] += jac[k][i] * jac[k][j]
				}
			}
			for k := range r {
				g[i] -= jac[k][i] * r[k]
			}
		}
		improved := false
		for tries := 0; tries < 20; tries++ {
			aa := make([][]float64, m)
			for i := range a {
				aa[i] = append([]float64(nil), a[i]...)
				aa[i][i] += lambda * (1 + a[i][i])
			}
			d := gSolve(aa, append([]float64(nil), g...))
			hn := make([]float64, m)
			ok := true
			for i := range h {
				hn[i] = math.Max(h[i]+d[i], math.Min(h0[i], gFitMinHandle))
				if math.IsNaN(hn[i]) {
					ok = false
				}
			}
			if ok {
				if rn := resid(hn); sq(rn) < sq(r) {
					h, r = hn, rn
					lambda /= 3
					improved = true
					break
				}
			}
			lambda *= 4
		}
		if !improved {
			break
		}
	}
	return build(h)
}

// gFittedEdges returns the curves, which join end to end with a common
// direction, as edges whose pieces also share the curvature where they join.
// Each curve is cut in the middle.  Where two curves which bend different ways
// join, each is also cut 0.15 of its parameter from that join, and the two
// short ends next to the join are one piece, an arc which turns from the one way
// to the other and belongs to the edge of the second curve.  The lengths of the
// handles of the pieces are fitted (gFitted) to keep the curvature of the given
// curves where they are cut and at the two ends, and at a join of two curves
// which bend the same way the curvature of the one which bends less there.
func gFittedEdges(cs []cubic) []gEdge {
	n := len(cs)
	turns := func(i int) bool { return i > 0 && i < n && cs[i-1].curvature(1)*cs[i].curvature(0) < 0 }
	var hs []cubic
	var owner []int
	target := map[int]float64{}
	target[0] = cs[0].curvature(0)
	handle := func(q, r outline.Point) float64 { return math.Hypot(r.X-q.X, r.Y-q.Y) }
	for i, c := range cs {
		if i > 0 && !turns(i) {
			ka, kb := cs[i-1].curvature(1), c.curvature(0)
			target[len(hs)] = ka
			if math.Abs(kb) < math.Abs(ka) {
				target[len(hs)] = kb
			}
		}
		cuts := []float64{0}
		if turns(i) {
			cuts = append(cuts, 0.15)
		}
		cuts = append(cuts, 0.5)
		if turns(i + 1) {
			cuts = append(cuts, 0.85)
		}
		cuts = append(cuts, 1)
		for j := 1; j < len(cuts); j++ {
			h := c.sub(cuts[j-1], cuts[j])
			if j == 1 && turns(i) {
				last := hs[len(hs)-1]
				hs[len(hs)-1] = gHandled(last.P0, last.tangent(0), handle(last.P0, last.P1), h.P3, h.tangent(1), handle(h.P2, h.P3))
				owner[len(hs)-1] = i
			} else {
				hs, owner = append(hs, h), append(owner, i)
			}
			if j < len(cuts)-1 {
				target[len(hs)] = c.curvature(cuts[j])
			}
		}
	}
	target[len(hs)] = cs[n-1].curvature(1)
	f := gFitted(hs, target)
	out := make([]gEdge, n)
	for j, c := range f {
		out[owner[j]] = append(out[owner[j]], c)
	}
	return out
}

// gFitWeight is the weight of a difference of curvature at a join against a
// unit of the length of a handle in gFitted.
const gFitWeight = 1e7

// gFitMinHandle is the shortest handle which gFitted gives a curve whose
// handle is at least as long, in units: a handle shorter than that loses its
// direction when the outline is rounded to integers.
const gFitMinHandle = 15.0

// gFitMid is the weight of a difference of curvature from a target against a
// unit of the length of a handle in gFitted.
const gFitMid = 30000.0

// gSolve solves the linear system a x = b by Gaussian elimination with partial
// pivoting; a and b are overwritten.
func gSolve(a [][]float64, b []float64) []float64 {
	n := len(b)
	for c := 0; c < n; c++ {
		p := c
		for r := c + 1; r < n; r++ {
			if math.Abs(a[r][c]) > math.Abs(a[p][c]) {
				p = r
			}
		}
		a[c], a[p] = a[p], a[c]
		b[c], b[p] = b[p], b[c]
		for r := c + 1; r < n; r++ {
			f := a[r][c] / a[c][c]
			for k := c; k < n; k++ {
				a[r][k] -= f * a[c][k]
			}
			b[r] -= f * b[c]
		}
	}
	x := make([]float64, n)
	for r := n - 1; r >= 0; r-- {
		s := b[r]
		for k := r + 1; k < n; k++ {
			s -= a[r][k] * x[k]
		}
		x[r] = s / a[r][r]
	}
	return x
}

// A gEdge is an edge of g drawn as cubic curves end to end, which it walks
// with one parameter from 0 to 1, each curve taking an equal share.
type gEdge []cubic

// split returns the curve and the parameter on it of the point t of the edge.
func (e gEdge) split(t float64) (int, float64) {
	n := float64(len(e))
	i := int(math.Min(math.Floor(t*n), n-1))
	return i, t*n - float64(i)
}

// at returns the point of the edge at the parameter t.
func (e gEdge) at(t float64) outline.Point {
	i, u := e.split(t)
	return e[i].at(u)
}

// start and end return the ends of the edge.
func (e gEdge) start() outline.Point { return e[0].P0 }
func (e gEdge) end() outline.Point   { return e[len(e)-1].P3 }

// reversed returns the edge walked the other way.
func (e gEdge) reversed() gEdge {
	out := make(gEdge, len(e))
	for i, c := range e {
		out[len(e)-1-i] = c.reversed()
	}
	return out
}

// sub returns the part of the edge from the parameter t0 to t1 > t0 as a
// chain, cut at the extremes.
func (e gEdge) sub(t0, t1 float64) chain {
	i0, u0 := e.split(t0)
	i1, u1 := e.split(t1)
	if u1 == 0 && i1 > i0 {
		i1, u1 = i1-1, 1
	}
	var ch chain
	for i := i0; i <= i1; i++ {
		a, b := 0.0, 1.0
		if i == i0 {
			a = u0
		}
		if i == i1 {
			b = u1
		}
		ch = append(ch, splitAtExtrema(e[i].sub(a, b))...)
	}
	return ch
}

// crossing returns the parameters at which the edges e and o cross.
func (e gEdge) crossing(o gEdge) (float64, float64, bool) {
	for i, c := range e {
		for j, d := range o {
			if ta, tb, ok := crossing(c, d); ok {
				return (float64(i) + ta) / float64(len(e)), (float64(j) + tb) / float64(len(o)), true
			}
		}
	}
	return 0, 0, false
}

// gFrameOf returns the frame of g, which panics if the parameters do not allow the
// construction; Check turns the panic into an error.
func gFrameOf(p design.Params) gFrame {
	f, err := newGFrame(p)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph g: %v; glyphs.Check reports such parameters", err))
	}
	return f
}

// The ear: constants for the direction in which the stalk of the ear leaves the
// bowl, the point of the bowl on which it stands and the part of the ear.  The
// ball and the end of the stalk are the shared ball terminal (ball.go) at the
// size of the ear.
const (
	// gEarLeave is the direction in which the edges of the stalk leave the
	// bowl, degrees above the horizontal.
	gEarLeave = 40.0
	// gEarRoot is the angle on the bowl, counter-clockwise from the right
	// extreme, of the point in the middle of its stroke on which the stalk
	// stands.
	gEarRoot = 55.0
	// gEarCurl is the radius of the outer edge of the stalk at the section of
	// the part, as a multiple of the radius of the ball.
	gEarCurl = 2.0
)

// A gEarFrame holds the curves of the ear: the shared ball terminal at the size
// of the ear, and the two edges of the stalk which join it to the bowl.
type gEarFrame struct {
	centre outline.Point // the centre of the ball
	r      float64       // its radius
	part   placedBall    // the ball terminal in its place
	under  cubic         // the underside of the stalk, from the bowl to the part
	up     cubic         // the upper edge of the stalk, from the part to the bowl

	tLq, tUq float64 // the underside and the upper edge meet the quarter 0 of the bowl at these parameters
}

// gEarPart returns the ball terminal of the ear: the ball of GEarBall across on a
// stalk GEarThick thick.  The angle over which the outer edge blends into the ball
// grows with the square root of the diameter against the full size.
func gEarPart(p design.Params) (ballPart, error) {
	return makeBallPart(p.GEarBall/2, p.GEarThick, gEarCurl, ballSection, ballBlend*math.Sqrt(p.GEarBall/p.BallDiameter))
}

// newGEarFrame returns the curves of the ear on the bowl of f.  The ball stands
// with its right extreme GEarReach beyond the bowl and its centre GEarDrop below
// the x-height; the part is turned so that the crest of the outer edge, the top of
// the outer circle of the stalk, lies GEarRise above the x-height; the edges of the
// stalk leave the bowl at gEarLeave on either side of a stalk which stands on the
// middle of the stroke of the bowl at gEarRoot.
func newGEarFrame(p design.Params, f gFrame) (gEarFrame, error) {
	var e gEarFrame
	part, err := gEarPart(p)
	if err != nil {
		return e, err
	}
	e.r = part.R
	e.centre = outline.Point{X: design.GEarEnd(p) - e.r, Y: p.XHeight - p.GEarDrop}
	// The part in its frame has the centre of curl C = (-d, 0); turned by theta,
	// the top of the outer circle is at Cy + Ro with Cy = -d sin(theta).
	d := part.Ro - part.R
	sin := (e.centre.Y + part.Ro - design.GEarCrest(p)) / d
	if !(math.Abs(sin) < 0.5) {
		return e, fmt.Errorf("the ear cannot be turned to put its crest at %v", design.GEarCrest(p))
	}
	pl := ballPlace{Centre: e.centre, Angle: math.Asin(sin)}
	e.part = part.place(pl)

	// The stalk stands on the bowl: its straight edges, GEarThick apart, leave the
	// stroke of the bowl at the angle gEarRoot in the direction gEarLeave.
	top, bottom := design.GBowlTop(p), design.GBowlBottom(p)
	bl, br := design.GBowlLeft(p), design.GBowlRight(p)
	cx, cy := (bl+br)/2, (top+bottom)/2
	a := (br-bl)/2 - p.GBowlThick/2
	b := (top-bottom)/2 - p.ThinLower/2
	root := gDeg(gEarRoot)
	start := outline.Point{X: cx + a*root.X, Y: cy + b*root.Y}
	dir := gDeg(gEarLeave)
	normal := outline.Point{X: -dir.Y, Y: dir.X}
	half := p.GEarThick / 2
	q0 := f.outerB.cubic(0, true)
	foot := func(side float64) (stalkFoot, float64, error) {
		o := outline.Point{X: start.X + side*half*normal.X, Y: start.Y + side*half*normal.Y}
		p0 := outline.Point{X: o.X - 200*dir.X, Y: o.Y - 200*dir.Y}
		p3 := outline.Point{X: o.X + 200*dir.X, Y: o.Y + 200*dir.Y}
		line := cubic{p0, gBeyond(p0, dir, 400.0/3), gBeyond(p0, dir, 800.0/3), p3}
		_, tq, ok := crossing(line, q0)
		if !ok {
			return stalkFoot{}, 0, fmt.Errorf("an edge of the ear does not leave the bowl")
		}
		return stalkFoot{At: q0.at(tq), Dir: dir, Kappa: math.NaN()}, tq, nil
	}
	under, tL, err := foot(-1)
	if err != nil {
		return e, err
	}
	upper, tU, err := foot(1)
	if err != nil {
		return e, err
	}
	if !(tL < tU) {
		return e, fmt.Errorf("the edges of the ear leave the bowl the wrong way round")
	}
	e.tLq, e.tUq = tL, tU
	upper.Dir = neg(upper.Dir) // the contour runs back into the bowl
	if e.under, err = edgeIn(under, e.part.In); err != nil {
		return e, fmt.Errorf("the underside of the ear: %v", err)
	}
	if e.up, err = edgeOut(e.part.Out, upper); err != nil {
		return e, fmt.Errorf("the upper edge of the ear: %v", err)
	}
	return e, nil
}

// ball returns the part of the circle of the ball which is outline: from the tip
// of the wedge, counter-clockwise round the ball, to the point where the outer
// edge runs into it, in arcs of at most a quarter turn.
func (e gEarFrame) ball() chain {
	b, th := e.part.Part, e.part.Place.Angle
	return arcPieces(e.centre, e.r, b.BallW+th, b.Alpha+th+2*math.Pi)
}

// chain returns the ear as one chain of the outer contour, from the point where
// its underside leaves the bowl, along the underside into the ball terminal and
// back along the upper edge to the bowl.
func (e gEarFrame) chain() chain {
	return join(splitAtExtrema(e.under), e.part.Chain, splitAtExtrema(e.up))
}

// outer returns the outer contour as one chain, from the top of the bowl round to
// the top of the bowl.
func (f gFrame) outer(ear gEarFrame) chain {
	q0, q1, q2, q3 := f.outerB.cubic(0, true), f.outerB.cubic(1, true), f.outerB.cubic(2, true), f.outerB.cubic(3, true)
	one := func(c cubic) chain { return chain{curveSeg(c)} }
	return weld(
		one(q1),
		one(q2.sub(0, f.tQ)),
		f.a1.sub(0, 1),
		f.a2.sub(0, f.tJa),
		f.b6.sub(0, f.tJb).reversed(),
		f.b5.sub(0, 1), f.b4.sub(0, 1),
		f.b3.sub(0, 1),
		f.b2.sub(0, 1),
		f.b1.sub(0, f.tPb),
		one(q2.sub(f.tPq, 1)),
		one(q3),
		one(q0.sub(0, ear.tLq)),
		ear.chain(),
		one(q0.sub(ear.tUq, 1)),
	)
}

// counter returns the counter of the loop, clockwise, from the point where the
// inner edge of the left end meets the link.
func (f gFrame) counter() chain {
	// The contour closes: the end of the last piece is the start of the first.
	ch := weld(
		f.a2.sub(f.tKa, 1),
		f.a3.sub(0, 1),
		f.a4.sub(0, 1),
		f.a5.sub(0, 1),
		f.a6.sub(0, f.tKb),
	)
	ch[len(ch)-1].P3 = ch[0].P0
	return ch
}

// weld returns the chains one after the other as one chain, moving the start of
// each piece which begins within a thousandth of a unit of the end of the one
// before onto it.  It panics if two chains are further apart than that.
func weld(chains ...chain) chain {
	var all chain
	for _, c := range chains {
		for _, pc := range c {
			if len(all) > 0 {
				e := all[len(all)-1].P3
				if math.Abs(pc.P0.X-e.X) > 1e-3 || math.Abs(pc.P0.Y-e.Y) > 1e-3 {
					panic(fmt.Sprintf("glyphs: weld: a piece starts at %v, but the one before ends at %v", pc.P0, e))
				}
				pc.P0 = e
				if pc.line {
					pc.P1 = e
				}
			}
			all = append(all, pc)
		}
	}
	return all
}

// g draws the lowercase letter g, see the comment above.
func g(p design.Params) (outline.Path, float64) {
	f := gFrameOf(p)
	var path outline.Path
	ch := f.outer(f.ear)
	path.MoveTo(ch.start().X, ch.start().Y)
	ch.appendTo(&path)
	q := f.innerB
	path.MoveTo(q.ex[0][0], q.ex[0][1])
	for i := 0; i < 4; i++ {
		q.segment(&path, i, false)
	}
	f.counter().addContour(&path)
	return path, design.GAdvance(p)
}
