// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter t.
//
// The t is the stem of n, which rises TRise above the x-height and has an
// oblique top, the crossbar of f at the x-height (barLeft, barRight) and a foot
// in which the stem curves to the right and ends in an upturned tail.  It is one
// counter-clockwise contour without a counter and without foot serifs.  It runs
//
//	the top of the stem, from the top right corner of the stem along the flat of
//	TObliqueFlat and the oblique cut down to the left-hand edge
//	down the left-hand edge to the top of the bar, the left-hand part of the bar
//	(barLeft), and down the left-hand edge to the foot
//	the foot: the outer edge round the low point and up to the tip of the tail, the
//	cut tip, and the upper edge of the tail back and round the inside of the curve to
//	the right-hand edge of the stem (tFootChain)
//	up the right-hand edge to the underside of the bar, the right-hand part of the
//	bar (barRight), and up the right-hand edge to the top right corner
//
// The foot.  The left-hand edge of the stem is the outer edge of the curve: it
// leaves the straight edge vertically, tFoot.OuterStart plus tFoot.StemLead
// above the baseline, runs round the low point, which lies OvershootLower below
// the baseline and has a horizontal tangent, and goes on to the tail.  The tail
// ends in a blunt tip, whose rightmost point lies TReach beyond the right-hand
// edge of the stem and whose highest point lies TTipHeight above the baseline,
// where a straight cut at right angles to the direction tTail.TipAngle,
// TTipThick long, with its corners rounded with the radius tFoot.TipRound, has
// them.  The upper edge returns from the tip to the inner low point, which lies
// ThinLower above the outer one at the same x and is horizontal, and rises from
// there to the right-hand edge of the stem, which it meets vertically at
// tFoot.InnerTop plus tFoot.StemLead.  The curves meet with continuous
// direction and curvature (footChain).

// The top of the stem is cut by one straight line, which runs from a point
// TObliqueFlat short of the top right corner (TRise above the x-height) down to
// the left-hand edge of the stem, TObliqueDrop below the top; a horizontal flat
// of TObliqueFlat joins the upper end of the cut to the right-hand edge.  The
// cut is at 54 degrees to the horizontal, and no angle of the top is below 90
// degrees.
//
// Side kinds: the left side is measured to the end of the bar
// (design.TInkLeft), the right side is open: the advance is made from the reach of
// the tail and TSideRight.

// A tFootShape holds the constants which shape the foot of t.
type tFootShape struct {
	LowRight    float64 // the low point lies this far to the right of the right-hand edge of the stem
	OuterStart  float64 // height above the baseline at which the guide of the outer curve leaves the left-hand edge
	InnerTop    float64 // height above the baseline at which the guide of the inner curve meets the right-hand edge
	TipAngle    float64 // direction of the tail at its tip, in degrees above the horizontal
	TipRound    float64 // radius of the rounded corners of the cut which place the rightmost and the highest point of the tip
	RightRadius float64 // radius of the edge at the rightmost point of the tip
	TopRadius   float64 // radius of the edge at the highest point of the tip
	LowRadius   float64 // radius of the outer edge at the low point
	InnerRadius float64 // radius of the inner edge at the inner low point
	InnerShift  float64 // the inner low point lies this far to the left of the low point
	CornerRun   float64 // the upper edge is straight where it runs in the direction of the tail, this far beyond the rounded upper corner of the cut
	StemBend    float64 // curvature of the two curves where they leave the stem
	StemLead    float64 // the two curves leave the stem this far above their guides
	OuterCorner bool    // the outer curve starts at a corner, where its guide starts, with the curvature of the guide
	OuterA      float64 // handles of the guide of the outer curve: at the stem edge, as a fraction of the drop
	OuterB      float64 // and at the low point, as a fraction of the run
	InnerA      float64 // handles of the guide of the inner curve: at the inner low point, as a fraction of the run
	InnerB      float64 // and at the stem edge, as a fraction of the rise
}

var tFoot = tFootShape{
	LowRight:    55,
	OuterStart:  110,
	InnerTop:    80,
	TipAngle:    72,
	TipRound:    9,
	RightRadius: 6,
	TopRadius:   9,
	LowRadius:   50,
	InnerRadius: 35,
	CornerRun:   10,
	StemBend:    0.003,
	StemLead:    10,
	OuterA:      0.55,
	OuterB:      0.55,
	InnerA:      0.55,
	InnerB:      0.55,
}

// tTail is the foot of t: the constants of tFoot with the tail angle 62.
var tTail = func() tFootShape {
	sh := tFoot
	sh.TipAngle = 62
	return sh
}()

// tStem returns the stem of t; the top is drawn by t, the foot by tFootChain,
// and the stem has no serifs at the bottom.
func tStem(p design.Params) stemSpec {
	return stemSpec{
		Left: design.TStemLeft(p), Width: p.Stem, Bottom: 0, Top: design.TTop(p),
	}
}

// tBar returns the crossbar of t as a barSpec.
func tBar(p design.Params) barSpec {
	s := tStem(p)
	return barSpec{
		Top: p.XHeight, Thick: p.Hair, End: p.Hair,
		LeftEdge: s.Left, RightEdge: s.right(),
		Left: p.TBarLeft, Right: p.TBarRight,
	}
}

// tFootChain returns the foot of t as a chain which starts on the left-hand edge
// of the stem, at the height sh.OuterStart plus sh.StemLead, and ends on the
// right-hand edge, at the height sh.InnerTop plus sh.StemLead.  The radii of
// the tip are those of sh times TTipScale.
func tFootChain(p design.Params, sh tFootShape) chain {
	s := tStem(p)
	if k := p.TTipScale; k > 0 {
		sh.TipRound *= k
		sh.RightRadius *= k
		sh.TopRadius *= k
		sh.CornerRun *= k
	}
	return footChain(p, s.Left, s.right(), p.TReach, p.TTipHeight, p.TTipThick, sh)
}

// footChain returns the foot of a stem with an upturned tail: the stem has the
// edges xl and xr, the tail reaches reach beyond xr, its highest ink is tipHeight
// above the baseline and it is tipThick thick at the blunt tip.  The foot of t
// and the tail of a use it.  The chain is that of tFootChain.  It consists of
// cubic curves which meet with continuous direction and curvature (g2Cubic):
//
//	the outer curve, from the left-hand edge of the stem, which it leaves with
//	the curvature sh.StemBend, through the point outerGuideAt of its guide, with
//	the direction and the curvature of the guide there, to the low point, which
//	it reaches with the radius sh.LowRadius; or, if sh.OuterCorner is set, one
//	curve from the start of the guide, with its curvature, to the low point
//	the outer edge of the tail, from the low point to the rightmost point of
//	the tip, which it reaches vertically with the radius sh.RightRadius
//	the end of the tail, from there to the highest point of the tip, which it
//	reaches horizontally with the radius sh.TopRadius
//	the upper corner, from the highest point to the point sh.CornerRun beyond
//	the rounded upper corner of the cut, where the edge runs straight in the
//	direction opposite to sh.TipAngle
//	the upper edge, from there to the inner low point, which it reaches
//	horizontally with the radius sh.InnerRadius
//	the inner curve, from the inner low point through the point innerGuideAt of
//	its guide to the right-hand edge of the stem, which it meets with the
//	curvature sh.StemBend
//
// The rightmost and the highest point of the tip are those of a straight cut
// at right angles to the direction sh.TipAngle, tipThick long, whose corners
// are rounded with the radius sh.TipRound.  The guides are the cubic curves
// from the stem edges, sh.OuterStart and sh.InnerTop above the baseline, to the
// low points, with the handles sh.OuterA, sh.OuterB, sh.InnerA and sh.InnerB.
// Where the upper edge or the inner curve cannot reach the inner low point
// with the radius sh.InnerRadius, both take the nearest curvature with which
// they can; where no curvature allows both, the inner low point moves, by the
// least whole number of units up to footShiftMax, to the left or, where that
// does not help, to the right, so that both reach it with the radius
// sh.InnerRadius and each handle of the upper edge is at least footHandleMin
// of its chord.
func footChain(p design.Params, xl, xr, reach, tipHeight, tipThick float64, sh tFootShape) chain {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	low := -p.OvershootLower
	xb := xr + sh.LowRight
	ib := low + p.ThinLower

	// The centres of the two rounded corners of the cut: the outer one, whose
	// right extreme is the rightmost ink, and the upper one, tipThick less twice
	// the radius further along the cut, whose top is the highest ink.
	th := sh.TipAngle * math.Pi / 180
	dir := pt(math.Cos(th), math.Sin(th))
	e := pt(-dir.Y, dir.X) // along the cut, from the outer corner to the upper one
	rho := sh.TipRound
	w := tipThick - 2*rho
	cOuter := pt(xr+reach-rho, tipHeight-rho-w*e.Y)
	cUpper := pt(cOuter.X+w*e.X, cOuter.Y+w*e.Y)
	right := pt(cOuter.X+rho, cOuter.Y)
	top := pt(cUpper.X, cUpper.Y+rho)
	turn := pt(cUpper.X+rho*e.X-sh.CornerRun*dir.X, cUpper.Y+rho*e.Y-sh.CornerRun*dir.Y)

	up, down, east, west := pt(0, 1), pt(0, -1), pt(1, 0), pt(-1, 0)
	lowOuter := pt(xb, low)

	// The guide of the outer curve and its point through which the curve
	// passes; the guide of the inner curve is drawn below for each place of
	// the inner low point.
	og := cubic{pt(xl, sh.OuterStart), pt(xl, sh.OuterStart-sh.OuterA*(sh.OuterStart-low)), pt(xb-sh.OuterB*(xb-xl), low), lowOuter}
	om, ot, ok := og.at(outerGuideAt), og.tangent(outerGuideAt), og.curvature(outerGuideAt)
	innerStem := pt(xr, sh.InnerTop+sh.StemLead)

	// The curvature at the inner low point is 1/sh.InnerRadius or, where the
	// upper edge or the inner curve cannot take it, the nearest one which both
	// can take.  Where none can, the inner low point moves to the left or,
	// where that does not help, to the right, by the least whole number of
	// units at which both reach it with the radius sh.InnerRadius and each
	// handle of the upper edge is at least footHandleMin of its chord.
	kInner := 1 / sh.InnerRadius
	var upper, inner cubic
	var lowInner, im, it outline.Point
	var ik float64
	found := false
	place := func(shift float64) {
		lowInner = pt(xb-sh.InnerShift-shift, ib)
		ig := cubic{pt(xr, sh.InnerTop), pt(xr, sh.InnerTop-sh.InnerB*(sh.InnerTop-ib)), pt(lowInner.X-sh.InnerA*(lowInner.X-xr), ib), lowInner}
		im, it, ik = ig.at(innerGuideAt), ig.tangent(innerGuideAt), ig.curvature(innerGuideAt)
	}
	place(0)
	for i := 0; i < 100 && !found; i++ {
		for _, f := range []float64{1 + 0.01*float64(i), 1 - 0.01*float64(i)} {
			var err1, err2 error
			upper, err1 = g2Cubic(turn, neg(dir), 0, lowInner, west, -f*kInner)
			inner, err2 = g2Cubic(im, it, ik, lowInner, east, f*kInner)
			if err1 == nil && err2 == nil {
				found = true
				break
			}
		}
	}
	for d := 1.0; d <= footShiftMax && !found; d++ {
		for _, shift := range []float64{d, -d} {
			place(shift)
			for i := 0; i < 100 && !found; i++ {
				for _, f := range []float64{1 + 0.01*float64(i), 1 - 0.01*float64(i)} {
					var err1, err2 error
					upper, err1 = g2Cubic(turn, neg(dir), 0, lowInner, west, -f*kInner)
					inner, err2 = g2Cubic(im, it, ik, lowInner, east, f*kInner)
					if err1 == nil && err2 == nil && minHandle(upper) >= footHandleMin {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}
	}
	if !found {
		panic("glyphs: the foot of a stem: no upper edge of the tail reaches the inner low point; glyphs.Check reports such parameters")
	}

	curve := func(name string, p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) chain {
		c, err := g2Cubic(p0, t0, k0, p3, t3, k3)
		if err != nil {
			panic(fmt.Sprintf("glyphs: the foot of a stem: %s: %v; glyphs.Check reports such parameters", name, err))
		}
		return splitAtExtrema(c)
	}
	var outer chain
	if sh.OuterCorner {
		outer = curve("the outer edge of the foot", og.P0, down, og.curvature(0), lowOuter, east, 1/sh.LowRadius)
	} else {
		outer = join(
			curve("the outer edge of the foot", pt(xl, sh.OuterStart+sh.StemLead), down, sh.StemBend, om, ot, ok),
			curve("the outer edge of the foot", om, ot, ok, lowOuter, east, 1/sh.LowRadius),
		)
	}
	ch := join(
		outer,
		curve("the outer edge of the tail", lowOuter, east, 1/sh.LowRadius, right, up, 1/sh.RightRadius),
		curve("the end of the tail", right, up, 1/sh.RightRadius, top, west, 1/sh.TopRadius),
		curve("the upper corner of the tail", turn, dir, 0, top, east, -1/sh.TopRadius).reversed(),
		splitAtExtrema(upper),
		splitAtExtrema(inner).reversed(),
		curve("the inner edge of the foot", innerStem, down, sh.StemBend, im, it, ik).reversed(),
	)
	return ch
}

// footShiftMax is how far, in units, the inner low point of the foot may move
// to either side.
const footShiftMax = 40.0

// footHandleMin is the least length of each handle of the upper edge of the
// tail, as a fraction of its chord, where the inner low point of the foot
// moves.
const footHandleMin = 0.25

// minHandle returns the length of the shorter handle of c as a fraction of its
// chord.
func minHandle(c cubic) float64 {
	ch := math.Hypot(c.P3.X-c.P0.X, c.P3.Y-c.P0.Y)
	h0 := math.Hypot(c.P1.X-c.P0.X, c.P1.Y-c.P0.Y)
	h1 := math.Hypot(c.P3.X-c.P2.X, c.P3.Y-c.P2.Y)
	return math.Min(h0, h1) / ch
}

// outerGuideAt and innerGuideAt are the parameters of the points of the guides
// of the outer and the inner curve of the foot through which the curves pass.
const (
	outerGuideAt = 0.3
	innerGuideAt = 0.4
)

// t draws the lowercase letter t, see the comment above.
func t(p design.Params) (outline.Path, float64) {
	stem, bar := tStem(p), tBar(p)
	var path outline.Path
	path.MoveTo(stem.right(), stem.Top)
	// The flat at the top, then the oblique cut down to the left-hand edge.
	if p.TObliqueFlat > 0 {
		path.LineTo(stem.right()-p.TObliqueFlat, stem.Top)
	}
	path.LineTo(stem.Left, stem.Top-p.TObliqueDrop)
	path.LineTo(stem.Left, bar.Top)
	barLeft(&path, bar)
	foot := tFootChain(p, tTail)
	path.LineTo(foot.start().X, foot.start().Y)
	foot.appendTo(&path)
	path.LineTo(stem.right(), bar.Top-bar.Thick)
	barRight(&path, bar)
	path.LineTo(stem.right(), stem.Top)
	return path, design.TAdvance(p)
}
