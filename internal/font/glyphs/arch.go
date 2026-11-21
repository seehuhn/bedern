// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// An archShape holds the constants which shape an arch: how far the
// underside peaks to the left of the outer edge, and the lengths of the
// handles of the Bézier curves as fractions of the extent of their segment.
// The outer edge and the underside have constants of their own.
type archShape struct {
	UnderPeakShift    float64 // the underside peaks this far left of the top
	LeftHandle        float64 // outer edge: handle at the stem, as a fraction of the segment length
	PeakHandleL       float64 // outer edge: handle at the peak, on the left of it
	PeakHandleR       float64 // outer edge, shoulder at a notch: handle at the peak, on the right of it, as a fraction of the width of the shoulder
	ShoulderHandle    float64 // outer edge, shoulder at a notch: handle at the foot of the shoulder, as a fraction of its height
	ShoulderLow       float64 // outer edge: the shoulder spline's second control point, as a fraction of the height from the shoulder to the peak
	ShoulderHigh      float64 // outer edge: its third control point, the corner of the control polygon on the line of the stem edge
	UnderLeftHandle   float64 // underside: handle at the stem
	UnderPeakHandleL  float64 // underside: handle at the peak, on the left of it
	UnderShoulderLow  float64 // underside: the same as ShoulderLow, from the inner shoulder to the top of the underside
	UnderShoulderHigh float64 // underside: the same as ShoulderHigh
	OuterSteeper      float64 // outer edge: degrees by which it leaves the stem more steeply than the angle of the stroke
	UnderSteeper      float64 // underside: the same
	OuterBend         float64 // outer edge: least angle in degrees by which the control polygon turns right at the handle at the stem
	UnderBend         float64 // underside: the same
}

// An archSpec describes an arch between two stems.  It joins the right-hand
// edge X1 of the left stem to the left-hand edge X2 and the right-hand edge
// X3 of the right stem.  The heights are measured from the baseline.
type archSpec struct {
	X1, X2, X3    float64 // right edge of the left stem, left and right edge of the right stem
	XPeak         float64 // position of the highest point of the outer edge
	Peak          float64 // height of the highest point of the outer edge
	Thick         float64 // vertical thickness of the arch at the peak
	Notch         float64 // height at which the outer edge meets the left stem
	Join          float64 // height at which the underside leaves the left stem
	ThinJoin      float64 // thickness at right angles to the stroke where it leaves the stem
	Shoulder      float64 // height at which the outer edge leaves the right stem
	ShoulderInner float64 // height at which the underside meets the right stem

	// AtNotch is set if the outer edge ends at the shoulder in a notch, where
	// the outer edge of the next arch leaves the stem, and does not run on
	// into the edge of the stem.
	AtNotch bool
}

// leaveAngle returns the cosine and the sine of the angle at which both
// edges leave the left stem.  The thickness at right angles to the stroke is
// ThinJoin: the vertical gap g between the notch and the join gives
// cos(angle) = thin/g.
func (a archSpec) leaveAngle() (cos, sin float64) {
	g := a.Notch - a.Join
	cos = math.Min(1, a.ThinJoin/g)
	return cos, math.Sqrt(1 - cos*cos)
}

// steeper returns the cosine and the sine of the angle which is the leaving angle
// plus the given number of degrees, but at most design.ArchMaxLeave.
func (a archSpec) steeper(deg float64) (cos, sin float64) {
	c, s := a.leaveAngle()
	th := math.Min(math.Atan2(s, c)+deg*math.Pi/180, design.ArchMaxLeave*math.Pi/180)
	return math.Cos(th), math.Sin(th)
}

// archUnder returns the underside of the arch as a chain of two curves.  It
// starts on the left stem, at the join.  The chain runs up to the highest point
// of the underside, which lies Thick below the peak of the outer edge and
// UnderPeakShift to its left, and down into the right stem.  It ends on the
// left-hand edge of the right stem, at the inner shoulder.
func archUnder(sh archShape, a archSpec) chain {
	xu := a.XPeak - sh.UnderPeakShift
	yu := a.Peak - a.Thick
	cos, sin := a.steeper(sh.UnderSteeper)
	hl := sh.UnderLeftHandle * (xu - a.X1)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	p1 := pt(a.X1+hl*cos, a.Join+hl*sin)
	x2 := bendLimit(p1, math.Atan2(sin, cos), sh.UnderBend, yu, xu-sh.UnderPeakHandleL*(xu-a.X1), xu-0.2*(xu-a.X1))
	branch := cubic{pt(a.X1, a.Join), p1, pt(x2, yu), pt(xu, yu)}
	s, q := pt(a.X2, a.ShoulderInner), pt(xu, yu)
	sh2 := shoulderSpline(s, q, shoulderHandle(s, q, math.Abs(branch.curvature(1)), sh.UnderShoulderHigh), sh.UnderShoulderLow, sh.UnderShoulderHigh)
	return join(chain{curveSeg(branch)}, sh2.reversed())
}

// archOuter returns the outer edge of the arch as a chain of two curves.  It
// starts on the right-hand edge of the right stem, at the outer shoulder.  The
// chain runs over the shoulder to the peak and down to the left stem.  It ends
// on the right-hand edge of the left stem, at the notch.
func archOuter(sh archShape, a archSpec) chain {
	cos, sin := a.steeper(sh.OuterSteeper)
	hn := sh.LeftHandle * (a.XPeak - a.X1)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	p1 := pt(a.X1+hn*cos, a.Notch+hn*sin)
	x2 := bendLimit(p1, math.Atan2(sin, cos), sh.OuterBend, a.Peak, a.XPeak-sh.PeakHandleL*(a.XPeak-a.X1), a.XPeak-0.2*(a.XPeak-a.X1))
	branch := cubic{pt(a.XPeak, a.Peak), pt(x2, a.Peak), p1, pt(a.X1, a.Notch)}
	if a.AtNotch {
		// The edge ends in the notch: one round curve which arrives
		// vertically.
		return chain{
			curveSeg(cubic{pt(a.X3, a.Shoulder),
				pt(a.X3, a.Shoulder+sh.ShoulderHandle*(a.Peak-a.Shoulder)),
				pt(a.XPeak+sh.PeakHandleR*(a.X3-a.XPeak), a.Peak),
				pt(a.XPeak, a.Peak)}),
			curveSeg(branch),
		}
	}
	s, q := pt(a.X3, a.Shoulder), pt(a.XPeak, a.Peak)
	sh2 := shoulderSpline(s, q, shoulderHandle(s, q, math.Abs(branch.curvature(0)), sh.ShoulderHigh), sh.ShoulderLow, sh.ShoulderHigh)
	return join(sh2, chain{curveSeg(branch)})
}

// shoulderSpline returns a shoulder: the edge from the point s on the edge of a
// stem, where it leaves the stem vertically, to the top q, where it runs
// horizontally.  It is a cubic B-spline with one inner knot, written as two
// Bézier curves, with continuous curvature.  Its control polygon is
//
//	s, (s.X, s.Y + low h), (s.X, s.Y + high h), (q.X + b, q.Y), q
//
// with h the height from s to q, and b towards the stem.  The first three points
// lie on the line of the stem edge, and the edge leaves the stem with the
// tangent of the stem and with zero curvature.  The fourth lies at the height
// of the top, and the edge is horizontal there.
func shoulderSpline(s, q outline.Point, b, low, high float64) chain {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	h := q.Y - s.Y
	q1 := pt(s.X, s.Y+low*h)
	q2 := pt(s.X, s.Y+high*h)
	q3 := pt(q.X+math.Copysign(b, s.X-q.X), q.Y)
	mid := func(u, v outline.Point) outline.Point { return lerp(u, v, 0.5) }
	a, c := mid(q1, q2), mid(q2, q3)
	m := mid(a, c)
	return chain{
		curveSeg(cubic{s, q1, a, m}),
		curveSeg(cubic{m, c, q3, q}),
	}
}

// shoulderHandle returns the distance b of the fourth control point of a
// shoulder spline from the top q (see shoulderSpline) for which the spline has
// the curvature k (a magnitude) at q: the last Bézier curve of the spline is
// m, (q2+q3)/2, q3, q, whose curvature at q is 2/3 (q.Y-q2.Y)/2 / b^2.  The
// handle is at most 0.9 of the width of the shoulder.
func shoulderHandle(s, q outline.Point, k, high float64) float64 {
	b := 0.9 * math.Abs(s.X-q.X)
	if k > 0 {
		b = math.Min(b, math.Sqrt((1-high)*(q.Y-s.Y)/(3*k)))
	}
	return b
}

// shoulderK returns the curvature at the top q of the shoulder spline from s
// with the handle b, see shoulderHandle.
func shoulderK(s, q outline.Point, b, high float64) float64 {
	return (1 - high) * (q.Y - s.Y) / (3 * b * b)
}

// bendLimit returns the x coordinate of the second control point of an edge which
// leaves the stem through the first control point p1, at the angle theta
// (radians), and arrives horizontally at the height y: the second control point
// lies at the height y, and x is its default value, moved to the right as far as
// is needed for the control polygon to turn right by at least bend degrees at p1,
// but no further than max.
func bendLimit(p1 outline.Point, theta, bend, y, x, max float64) float64 {
	psi := theta - bend*math.Pi/180
	if psi <= 0.05 || y <= p1.Y {
		return x
	}
	if need := p1.X + (y-p1.Y)/math.Tan(psi); need > x {
		x = math.Min(need, math.Max(x, max))
	}
	return x
}
