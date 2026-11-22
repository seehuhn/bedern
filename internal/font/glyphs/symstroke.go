// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The stroke of the symbols.
//
// The question mark, the ampersand, the at sign, the braces,
// the tilde and the hooks of sterling and euro are drawn as strokes: a centre
// line given by stations, each with a point, the direction in which the stroke
// runs there and its thickness there, measured along the nib.  The nib is the
// direction across the stroke in which the thickness is laid off; it is at right
// angles to the direction of the stroke unless a station gives it.  The two edges of the
// stroke pass through the ends of the nib at every station, and run there in the
// direction of the stroke, so that an extreme of the centre line with a
// horizontal or vertical direction is an extreme of both edges, and an on-curve
// point.
//
// Between two stations an edge is one cubic curve which leaves the first point
// and meets the second with the direction of the stroke and with a curvature
// which is the same on both sides of every station: the curvature at a station
// is the mean of the curvatures of the two circles which touch the edge there
// and pass through the edge points of the stations before and after it, so that
// the edge has no corner and no jump of curvature at a station and bends the
// more where the stations turn the more (g2Cubic).  Where no such curve exists
// the edge takes the handles of a circular arc.  The curves are cut at their
// extremes.
//
// The contour of a stroke runs along its right-hand edge from the first station
// to the last, across the end, back along the left-hand edge and across the
// start, so that the ink lies on the left of the direction of travel and the
// contour runs counter-clockwise, as the conventions of doc.go demand.  The ends
// are cut along the nib by default; a glyph which ends a stroke in a ball or a
// beak replaces the end.

// A station is a point of the centre line of a stroke.  Dir is the direction
// of the stroke in degrees from the positive x axis, W its thickness and Nib,
// if it is not zero, the direction of the nib in degrees; otherwise the nib is
// at right angles to Dir.  K, if it is not zero, fixes the curvature of the
// centre line at the station (positive to the left), which a station at an
// inflection or at the end of a straight run needs: the edges then take that
// curvature, corrected for their distance from the centre line.  A station
// with Flat set has edges which are straight towards the next station: the
// stroke is a straight run there.  Between a station at an end of the stroke
// with Ease set and its neighbour, which ends a straight run, an edge whose
// curve has at the straight a shorter handle than the eased curve (easeCubic)
// is that curve instead: it leaves the straight with the curvature zero and a
// handle of a set fraction of that of a circular arc, and takes at the end the
// curvature which follows.  At a station at an end of the stroke with ArcEnd
// set, each edge has the handle of a circular arc and meets the neighbouring
// station with the curvature there (g2End), where that curve exists.
type station struct {
	X, Y, Dir, W float64
	Nib          float64
	K            float64
	Flat         bool
	Ease         bool
	ArcEnd       bool
}

// dir returns the unit vector of the direction of the stroke.
func (s station) dir() outline.Point { return dirOf(s.Dir * math.Pi / 180) }

// nib returns the unit vector from the right-hand to the left-hand end of the
// nib.
func (s station) nib() outline.Point {
	if s.Nib != 0 {
		return dirOf(s.Nib * math.Pi / 180)
	}
	d := s.dir()
	return outline.Point{X: -d.Y, Y: d.X}
}

// edgePoint returns the point of the right-hand edge (side -1) or of the
// left-hand edge (side 1) at the station.
func (s station) edgePoint(side float64) outline.Point {
	n := s.nib()
	return outline.Point{X: s.X + side*s.W/2*n.X, Y: s.Y + side*s.W/2*n.Y}
}

// strokeEdge returns the edge of the stroke on the given side, from the first
// station to the last, as a chain cut at its extremes.
func strokeEdge(sts []station, side float64) chain {
	n := len(sts)
	pts := make([]outline.Point, n)
	tgs := make([]outline.Point, n)
	for i, s := range sts {
		pts[i], tgs[i] = s.edgePoint(side), s.dir()
	}
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	// kOut and kIn are the curvatures of the circles which touch the edge at
	// station i and pass through the edge point of station i+1, or i-1.
	kOut := func(i int) float64 {
		d := outline.Point{X: pts[i+1].X - pts[i].X, Y: pts[i+1].Y - pts[i].Y}
		return 2 * cr(tgs[i], d) / (d.X*d.X + d.Y*d.Y)
	}
	kIn := func(i int) float64 {
		d := outline.Point{X: pts[i].X - pts[i-1].X, Y: pts[i].Y - pts[i-1].Y}
		return 2 * cr(d, tgs[i]) / (d.X*d.X + d.Y*d.Y)
	}
	// A straight run has the direction of its chord at both ends.
	for i := 0; i+1 < n; i++ {
		if sts[i].Flat {
			tgs[i] = unit(pts[i], pts[i+1])
			tgs[i+1] = tgs[i]
		}
	}
	fixed := make([]bool, n)
	ks := make([]float64, n)
	for i := range sts {
		switch {
		case sts[i].K != 0:
			// The curvature of the centre line, carried to the edge: an edge
			// on the inside of the turn bends more.
			r := 1 / sts[i].K
			ks[i], fixed[i] = 1/(r-side*sts[i].W/2), true
		case sts[i].Flat || (i > 0 && sts[i-1].Flat):
			ks[i], fixed[i] = 0, true
		case i == 0:
			ks[i] = kOut(i)
		case i == n-1:
			ks[i] = kIn(i)
		default:
			ks[i] = (kIn(i) + kOut(i)) / 2
		}
	}
	// segOK reports whether the curve of segment i with the curvatures ks at
	// its ends exists and turns one way where its ends bend the same way.
	segOK := func(i int) bool {
		if i < 0 || i+1 >= n || sts[i].Flat {
			return true
		}
		c, err := g2Cubic(pts[i], tgs[i], ks[i], pts[i+1], tgs[i+1], ks[i+1])
		if err != nil {
			return false
		}
		if ks[i]*ks[i+1] > 0 {
			return oneWay(c, math.Copysign(1, ks[i]))
		}
		return true
	}
	// Where a curvature does not allow the curves on both sides of a station, it
	// is moved towards the values of the circles of the two sides, and between
	// them, until both curves exist.
	for pass := 0; pass < 3; pass++ {
		for i := range sts {
			if fixed[i] || (segOK(i-1) && segOK(i)) {
				continue
			}
			lo, hi := ks[i], ks[i]
			if i > 0 {
				lo = math.Min(lo, kIn(i))
				hi = math.Max(hi, kIn(i))
			}
			if i+1 < n {
				lo = math.Min(lo, kOut(i))
				hi = math.Max(hi, kOut(i))
			}
			orig := ks[i]
			found := false
			for k := 1; k <= 40 && !found; k++ {
				for _, sgn := range []float64{1, -1} {
					v := orig + sgn*float64(k)/40*(hi-lo+math.Abs(orig)*0.5)
					ks[i] = v
					if segOK(i-1) && segOK(i) {
						found = true
						break
					}
				}
			}
			if !found {
				ks[i] = orig
			}
		}
	}
	var ch chain
	for i := 0; i+1 < n; i++ {
		p0, p3, t0, t3 := pts[i], pts[i+1], tgs[i], tgs[i+1]
		if sts[i].Flat {
			ch = append(ch, lineSeg(p0, p3))
			continue
		}
		c, err := g2Cubic(p0, t0, ks[i], p3, t3, ks[i+1])
		if err != nil {
			c = nearG2(p0, t0, ks[i], p3, t3, ks[i+1])
		}
		if i+1 == n-1 && sts[i+1].Ease && i > 0 && sts[i-1].Flat {
			if e, err := easeCubic(p0, t0, p3, t3, true); err == nil && dist(c.P0, c.P1) < dist(e.P0, e.P1) {
				c = e
			}
		}
		if i == 0 && sts[0].Ease && n > 2 && sts[1].Flat {
			if e, err := easeCubic(p0, t0, p3, t3, false); err == nil && dist(c.P2, c.P3) < dist(e.P2, e.P3) {
				c = e
			}
		}
		if i == 0 && sts[0].ArcEnd {
			if e, err := g2End(p0, t0, arcHandle(p0, t0, p3, t3), p3, t3, ks[1]); err == nil {
				c = e
			}
		}
		if i+1 == n-1 && sts[n-1].ArcEnd {
			if e, err := g2Start(p0, t0, ks[i], p3, t3, arcHandle(p0, t0, p3, t3)); err == nil {
				c = e
			}
		}
		ch = append(ch, splitAtExtrema(c)...)
	}
	return ch
}

// dist returns the distance between a and b.
func dist(a, b outline.Point) float64 { return math.Hypot(b.X-a.X, b.Y-a.Y) }

// easeHandle is the handle of an eased curve at the straight, as a fraction
// of the handle of a circular arc.
const easeHandle = 1.0 / 3

// easeCubic returns the cubic curve from p0, leaving in the direction t0, to p3,
// arriving in the direction t3, whose curvature is zero at the end which meets
// a straight run (at p0 if atStart is set, else at p3) and whose handle there
// is easeHandle times that of a circular arc between the same points and
// directions (arcHandle).  The handle at the other end is the one which gives the
// curvature zero.  It reports an error if a handle is not positive or not
// shorter than the chord.
func easeCubic(p0, t0, p3, t3 outline.Point, atStart bool) (cubic, error) {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	d := outline.Point{X: p3.X - p0.X, Y: p3.Y - p0.Y}
	chord := math.Hypot(d.X, d.Y)
	s := cr(t0, t3)
	h := easeHandle * arcHandle(p0, t0, p3, t3)
	var a, b float64
	if atStart {
		a, b = h, cr(t0, d)/s
	} else {
		a, b = cr(d, t3)/s, h
	}
	if !(a > 0 && b > 0 && a < chord && b < chord) {
		return cubic{}, fmt.Errorf("no eased curve runs from %v to %v", p0, p3)
	}
	return cubic{p0,
		outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y},
		outline.Point{X: p3.X - b*t3.X, Y: p3.Y - b*t3.Y},
		p3}, nil
}

// strokeChain returns the closed contour of the stroke with both ends cut
// along the nib.
func strokeChain(sts []station) chain {
	r := strokeEdge(sts, -1)
	l := strokeEdge(sts, 1).reversed()
	return join(r, lineChain(r.end(), l.start()), l, lineChain(l.end(), r.start()))
}

// strokeOpen returns the two edges of the stroke, the right-hand edge from the
// first station to the last and the left-hand edge from the last to the first,
// for a glyph which draws the ends itself.
func strokeOpen(sts []station) (right, left chain) {
	return strokeEdge(sts, -1), strokeEdge(sts, 1).reversed()
}

// chainsPath returns the path made of the closed chains, each a contour.
func chainsPath(cs ...chain) outline.Path {
	var path outline.Path
	for _, c := range cs {
		c.addContour(&path)
	}
	return path
}

// ringChains returns the outer contour (counter-clockwise) and the counter
// (clockwise) of a ring whose outer edge is the super-ellipse in the box from
// (x0, y0) to (x1, y1) and whose inner edge lies thickX inside it at the sides
// and thickY at the top and bottom.
func ringChains(x0, y0, x1, y1, thickX, thickY, supOuter, supInner float64) (outer, counter chain) {
	q := newSuperShape(x0, y0, x1, y1, supOuter)
	for i := 0; i < 4; i++ {
		outer = append(outer, curveSeg(q.cubic(i, true)))
	}
	r := newSuperShape(x0+thickX, y0+thickY, x1-thickX, y1-thickY, supInner)
	for i := 0; i < 4; i++ {
		counter = append(counter, curveSeg(r.cubic(i, false)))
	}
	return outer, counter
}

// circleChain returns the circle about c of the radius r as a closed
// counter-clockwise chain of four quarters, starting at its right extreme.
func circleChain(c outline.Point, r float64) chain {
	return arcPieces(c, r, 0, 2*math.Pi)
}

// dotChain returns the dot of i (i.go) centred at (cx, cy): the super-ellipse
// of the diameter DotDiameter and the squareness DotSuperness, as a closed
// counter-clockwise chain.  It is the dot of the question mark.
func dotChain(p design.Params, cx, cy float64) chain {
	r := p.DotDiameter / 2
	q := newSuperShape(cx-r, cy-r, cx+r, cy+r, p.DotSuperness)
	var c chain
	for i := 0; i < 4; i++ {
		c = append(c, curveSeg(q.cubic(i, true)))
	}
	return c
}

// pathChains returns the contours of the path as closed chains, closing each
// with a straight piece if it does not end where it started.
func pathChains(path outline.Path) []chain {
	var out []chain
	for _, ct := range path {
		var c chain
		cur := ct.Start
		for _, s := range ct.Segs {
			if s.IsCurve {
				c = append(c, curveSeg(cubic{cur, s.P1, s.P2, s.P3}))
			} else {
				c = append(c, lineSeg(cur, s.P3))
			}
			cur = s.P3
		}
		if math.Hypot(cur.X-ct.Start.X, cur.Y-ct.Start.Y) > 1e-9 {
			c = append(c, lineSeg(cur, ct.Start))
		}
		out = append(out, c)
	}
	return out
}

// rectChain returns the counter-clockwise rectangle from (x0, y0) to (x1, y1).
func rectChain(x0, y0, x1, y1 float64) chain {
	a, b := outline.Point{X: x0, Y: y0}, outline.Point{X: x1, Y: y0}
	c, d := outline.Point{X: x1, Y: y1}, outline.Point{X: x0, Y: y1}
	return chain{lineSeg(a, b), lineSeg(b, c), lineSeg(c, d), lineSeg(d, a)}
}

// rotated180 returns the chain turned through half a turn about c, which keeps
// its orientation.
func (c chain) rotated180(at outline.Point) chain {
	return c.mapPoints(func(q outline.Point) outline.Point { return outline.Point{X: 2*at.X - q.X, Y: 2*at.Y - q.Y} })
}

// ballStart returns the ball terminal (ball.go) at the start of a stroke whose
// edges are right and left (strokeOpen), with its stalk: from the end of the
// left-hand edge, which the contour of the stroke runs back along, round the
// ball to the start of the right-hand edge.  The edges of the stalk go on from
// the stroke with its direction and its curvature, or, where no place allows
// that, with its direction.  The ball's centre has the abscissa cx and one of
// the heights ys, the part is mirrored if mirror is set and turned by one of the
// angles (degrees); of these the place is taken for which the two edges of the
// stalk bend least (bending).  It
// panics with the name of the glyph if no place lets the edges reach the part.
func ballStart(glyph string, p design.Params, right, left chain, cx float64, ys []float64, angles []float64, mirror bool) chain {
	c, _ := ballStartPlaced(glyph, p, right, left, cx, ys, angles, mirror)
	return c
}

// ballStartPlaced is ballStart which also returns the placed part.
func ballStartPlaced(glyph string, p design.Params, right, left chain, cx float64, ys []float64, angles []float64, mirror bool) (chain, placedBall) {
	b := newBallPart(p)
	r0, l1 := right[0].cubic, left[len(left)-1].cubic
	in := stalkFoot{At: l1.P3, Dir: l1.tangent(1), Kappa: l1.curvature(1)}
	out := stalkFoot{At: r0.P0, Dir: r0.tangent(0), Kappa: r0.curvature(0)}
	if best, tb := ballSearch(b, in, out, cx, ys, angles, mirror); best != nil {
		return best, tb
	}
	// No place lets the edges of the stalk go on with the curvature of the
	// stroke: they take their direction only.
	in.Kappa, out.Kappa = math.NaN(), math.NaN()
	if best, tb := ballSearch(b, in, out, cx, ys, angles, mirror); best != nil {
		return best, tb
	}
	panic("glyphs: glyph " + glyph + ": the stalk cannot reach the ball terminal; glyphs.Check reports such parameters")
}

// ballSearch returns the stalk and the ball of ballStart for the place, among
// those given, for which the edges of the stalk bend least, or nil, and the
// placed part.
func ballSearch(b ballPart, in, out stalkFoot, cx float64, ys, angles []float64, mirror bool) (chain, placedBall) {
	var best chain
	var bestTB placedBall
	bestE := math.Inf(1)
	for _, deg := range angles {
		for _, y := range ys {
			pl := ballPlace{Centre: outline.Point{X: math.Round(cx), Y: math.Round(y)}, Angle: deg * math.Pi / 180, Mirror: mirror}
			tb := b.place(pl)
			a, err1 := edgeIn(in, tb.In)
			c, err2 := edgeOut(tb.Out, out)
			if err1 != nil || err2 != nil {
				continue
			}
			if e := bending(a) + bending(c); e < bestE {
				bestE = e
				best = join(splitAtExtrema(a), tb.Chain, splitAtExtrema(c))
				bestTB = tb
			}
		}
	}
	return best, bestTB
}

// span returns the values from lo to hi in steps of step.
func span(lo, hi, step float64) []float64 {
	var out []float64
	for v := lo; v <= hi+1e-9; v += step {
		out = append(out, v)
	}
	return out
}

// nearG2 returns the cubic curve from p0, leaving in the direction t0, to p3,
// arriving in the direction t3, whose curvatures at the ends come nearest to k0
// and k3, for a pair of stations between which no curve has both: the handles
// are searched from 0.05 to 0.8 of the chord, and the error is the sum of the
// squares of the differences of the curvatures times the chord.
func nearG2(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) cubic {
	chord := math.Hypot(p3.X-p0.X, p3.Y-p0.Y)
	best, bestE := cubic{}, math.Inf(1)
	for i := 0; i <= 60; i++ {
		a := (0.05 + 0.75*float64(i)/60) * chord
		for j := 0; j <= 60; j++ {
			b := (0.05 + 0.75*float64(j)/60) * chord
			c := cubic{p0, outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y}, outline.Point{X: p3.X - b*t3.X, Y: p3.Y - b*t3.Y}, p3}
			d0, d3 := (c.curvature(0)-k0)*chord, (c.curvature(1)-k3)*chord
			if e := d0*d0 + d3*d3; e < bestE {
				best, bestE = c, e
			}
		}
	}
	return best
}
