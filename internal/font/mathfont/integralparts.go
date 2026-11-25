// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The assembly of the integral: a top hook, an extender and a bottom hook,
// for sizes above the display size.  The
// parts are cut from the integral of the height intAssemblyHeight, the
// shear of whose stem is taken out towards the axis (uprightWarp) and the
// stem is upright where the parts join.
const (
	// intAssemblyHeight is the height of the integral the parts are cut
	// from: with the straight runs endRun at the cuts and the connector, the
	// assembly without an extender is as high as the display size.
	intAssemblyHeight = intDisplayHeight - (2*endRun - connector)
	// intUprightRun is the length from the axis over which the stem turns
	// from upright to the slant of the hooks, as a part of the half height
	// less the hook.
	intUprightRun = 0.45
)

// cubicAt returns the point and the derivative of the cubic q at t.
func cubicAt(q [4]outline.Point, t float64) (pnt, der outline.Point) {
	u := 1 - t
	pnt = pt(u*u*u*q[0].X+3*u*u*t*q[1].X+3*u*t*t*q[2].X+t*t*t*q[3].X,
		u*u*u*q[0].Y+3*u*u*t*q[1].Y+3*u*t*t*q[2].Y+t*t*t*q[3].Y)
	der = pt(3*(u*u*(q[1].X-q[0].X)+2*u*t*(q[2].X-q[1].X)+t*t*(q[3].X-q[2].X)),
		3*(u*u*(q[1].Y-q[0].Y)+2*u*t*(q[2].Y-q[1].Y)+t*t*(q[3].Y-q[2].Y)))
	return pnt, der
}

// splitCubic returns the two halves of the cubic q cut at t.
func splitCubic(q [4]outline.Point, t float64) (l, r [4]outline.Point) {
	lerp := func(a, b outline.Point) outline.Point { return pt(a.X+t*(b.X-a.X), a.Y+t*(b.Y-a.Y)) }
	a, b, c := lerp(q[0], q[1]), lerp(q[1], q[2]), lerp(q[2], q[3])
	d, e := lerp(a, b), lerp(b, c)
	m := lerp(d, e)
	return [4]outline.Point{q[0], a, d, m}, [4]outline.Point{m, e, c, q[3]}
}

// uprightWarp takes the shear of the integral out near the axis: the point
// at the height u above the axis moves by -slant times u - f(u), where f
// follows u with the slope rising from 0 to 1 over the length run
// (f(u) = u|u|/(2 run) inside, |u| - run/2 beyond); the stem is
// upright at the axis and has the slant of the hooks beyond run, the hooks
// only moved sideways.
type uprightWarp struct{ axis, run, slant float64 }

// shift returns the horizontal move of a point at the height y, and its
// derivative with respect to y.
func (w uprightWarp) shift(y float64) (dx, ddx float64) {
	u := y - w.axis
	a := math.Abs(u)
	if a >= w.run {
		return -w.slant * math.Copysign(w.run/2, u), 0
	}
	return -w.slant * (u - u*a/(2*w.run)), -w.slant * (1 - a/w.run)
}

// warpPoint returns the image of q.
func (w uprightWarp) warpPoint(q outline.Point) outline.Point {
	dx, _ := w.shift(q.Y)
	return pt(q.X+dx, q.Y)
}

// warpCubic returns the image of the cubic q, which lies within the
// zone |u| <= run or outside it: the whole of it moved by the same amount
// outside, and inside as cubics which agree with the image of q in their end
// points and tangents (n of them).
func (w uprightWarp) warpCubic(q [4]outline.Point, n int) [][4]outline.Point {
	ym := (q[0].Y + q[3].Y) / 2
	if math.Abs(ym-w.axis) >= w.run {
		dx, _ := w.shift(ym)
		out := q
		for i := range out {
			out[i] = pt(q[i].X+dx, q[i].Y)
		}
		return [][4]outline.Point{out}
	}
	var out [][4]outline.Point
	jac := func(p, d outline.Point) outline.Point {
		_, ddx := w.shift(p.Y)
		return pt(d.X+ddx*d.Y, d.Y)
	}
	for i := range n {
		t0, t1 := float64(i)/float64(n), float64(i+1)/float64(n)
		p0, d0 := cubicAt(q, t0)
		p3, d3 := cubicAt(q, t1)
		j0, j3 := jac(p0, d0), jac(p3, d3)
		dt := (t1 - t0) / 3
		a, b := w.warpPoint(p0), w.warpPoint(p3)
		out = append(out, [4]outline.Point{a, pt(a.X+j0.X*dt, a.Y+j0.Y*dt), pt(b.X-j3.X*dt, b.Y-j3.Y*dt), b})
	}
	return out
}

// warped returns the contour with every segment warped; a segment which
// crosses the line at run above or below the axis is cut there first.
func (w uprightWarp) warped(c outline.Contour) outline.Contour {
	out := outline.Contour{Start: w.warpPoint(c.Start)}
	cur := c.Start
	for _, sg := range c.Segs {
		q := [4]outline.Point{cur, sg.P1, sg.P2, sg.P3}
		if !sg.IsCurve {
			q[1] = pt(cur.X+(sg.P3.X-cur.X)/3, cur.Y+(sg.P3.Y-cur.Y)/3)
			q[2] = pt(cur.X+2*(sg.P3.X-cur.X)/3, cur.Y+2*(sg.P3.Y-cur.Y)/3)
		}
		pieces := [][4]outline.Point{q}
		for _, yc := range []float64{w.axis - w.run, w.axis + w.run} {
			var next [][4]outline.Point
			for _, pc := range pieces {
				if (pc[0].Y-yc)*(pc[3].Y-yc) < 0 {
					lo, hi := 0.0, 1.0
					for range 60 {
						m := (lo + hi) / 2
						pm, _ := cubicAt(pc, m)
						p0, _ := cubicAt(pc, lo)
						if (p0.Y-yc)*(pm.Y-yc) <= 0 {
							hi = m
						} else {
							lo = m
						}
					}
					l, r := splitCubic(pc, (lo+hi)/2)
					l[3].Y, r[0].Y = yc, yc
					next = append(next, l, r)
					continue
				}
				next = append(next, pc)
			}
			pieces = next
		}
		for _, pc := range pieces {
			for _, k := range w.warpCubic(pc, 3) {
				out.Segs = append(out.Segs, outline.Segment{IsCurve: true, P1: k[1], P2: k[2], P3: k[3]})
			}
		}
		cur = sg.P3
	}
	return out
}

// integralParts returns the parts of the assembly of the integral: the top
// hook, the extender and the bottom hook, each with its bottom at 0, and the
// advance.
func integralParts(p design.Params) (top, bot, ext outline.Path, adv float64, err error) {
	return integralPartsOf(p, integralShape(p))
}

// integralPartsDrawn is what drawIntegralParts returned for one list of
// arguments.
type integralPartsDrawn struct {
	top, bot, ext outline.Path
	adv           float64
	err           error
	computed      sync.Once
}

// integralPartsCache holds the results of drawIntegralParts, which the search
// for the shape, the glyphs and the tests ask for again.
var integralPartsCache sync.Map // string to *integralPartsDrawn

// integralPartsOf is integralParts in the shape sh.  It draws the parts once
// for each list of arguments and hands out copies of the paths.
func integralPartsOf(p design.Params, sh intShape) (top, bot, ext outline.Path, adv float64, err error) {
	v, _ := integralPartsCache.LoadOrStore(fmt.Sprintf("%v|%v", p, sh), new(integralPartsDrawn))
	d := v.(*integralPartsDrawn)
	d.computed.Do(func() { d.top, d.bot, d.ext, d.adv, d.err = drawIntegralParts(p, sh) })
	return clonePath(d.top), clonePath(d.bot), clonePath(d.ext), d.adv, d.err
}

// drawIntegralParts is the drawing behind integralPartsOf.
func drawIntegralParts(p design.Params, sh intShape) (top, bot, ext outline.Path, adv float64, err error) {
	h := intAssemblyHeight
	axis := p.Punct.MathAxis
	s := integralStroke(p, h, sh).ShearedRatio(intSlantNum, intSlantDen)
	r, err := s.Build()
	if err != nil {
		return nil, nil, nil, 0, err
	}
	path, err := halfTurnOfFoot(r.Path(), axis)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	c := splitAtY(withNodesAt(path[0], axis), axis)
	w := uprightWarp{axis: axis, run: intUprightRun * h / 2, slant: float64(intSlantNum) / float64(intSlantDen)}
	path = outline.Path{w.warped(c)}
	path, adv = withSides(path, intSide, intSide)
	if len(path) != 1 {
		return nil, nil, nil, 0, fmt.Errorf("the integral has %d contours", len(path))
	}
	top, bot, ext, err = assemblyParts(path, axis, true, endRun, extender)
	return top, bot, ext, adv, err
}
