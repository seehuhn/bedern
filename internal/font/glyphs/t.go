// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
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
// leaves the straight edge vertically tFoot.OuterStart above the baseline, runs
// round the low point, which lies OvershootLower below the baseline and has a
// horizontal tangent, and goes on in a second curve to the tail, where it arrives
// at the angle tFoot.TipAngle to the horizontal.  The tail ends in a blunt tip: a
// straight cut at right angles to its direction, TTipThick long, whose two corners
// are rounded with the radius tFoot.TipRound.  The rounded outer corner is the
// rightmost ink, TReach beyond the right-hand edge of the stem, and the rounded
// upper corner is the highest ink of the tail, TTipHeight above the baseline.
// The upper edge returns from the tip to the inner low point, which lies ThinLower
// above the outer one at the same x and is horizontal, and rises from there in a
// curve to the right-hand edge of the stem, which it meets vertically at
// tFoot.InnerTop.  The foot is ThinLower thick at its lowest point and TTipThick
// thick at the tip.
//
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
	LowRight   float64 // the low point lies this far to the right of the right-hand edge of the stem
	OuterStart float64 // height above the baseline at which the left-hand edge begins to curve
	InnerTop   float64 // height above the baseline at which the inner curve meets the right-hand edge
	TipAngle   float64 // direction of the tail at its tip, in degrees above the horizontal
	TipRound   float64 // radius of the rounding of the two corners of the cut tip
	OuterA     float64 // handles of the first outer curve: at the stem edge, as a fraction of the drop
	OuterB     float64 // and at the low point, as a fraction of the run
	TailA      float64 // handle of the second outer curve at the low point, as a fraction of the run to the tip
	TailB      float64 // and at the tip, as a fraction of the chord
	UpperA     float64 // handle of the upper edge at the tip, as a fraction of the chord to the inner low point
	UpperB     float64 // and at the inner low point, as a fraction of the run from the tip
	InnerA     float64 // handles of the inner curve at the inner low point, as a fraction of the run
	InnerB     float64 // and at the stem edge, as a fraction of the rise
	// Lean is the angle in radians by which the edges of the stem lean to the
	// right going down where the foot leaves them: zero for the upright stems of
	// t and a, the lean of the leg for R.
	Lean float64
}

var tFoot = tFootShape{
	LowRight:   55,
	OuterStart: 110,
	InnerTop:   80,
	TipAngle:   72,
	TipRound:   9,
	OuterA:     0.55,
	OuterB:     0.55,
	TailA:      0.25,
	TailB:      0.35,
	UpperA:     0.35,
	UpperB:     0.3,
	InnerA:     0.55,
	InnerB:     0.55,
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
// of the stem, at the height tFoot.OuterStart, and ends on the right-hand edge, at
// the height tFoot.InnerTop.
func tFootChain(p design.Params, sh tFootShape) chain {
	s := tStem(p)
	return footChain(p, s.Left, s.right(), p.TReach, p.TTipHeight, p.TTipThick, sh)
}

// footChain returns the foot of a stem with an upturned tail: the stem has the
// edges xl and xr, the tail reaches reach beyond xr, its highest ink is tipHeight
// above the baseline and it is tipThick thick at the blunt tip.  The foot of t
// and the tail of the alternate a use it.  The chain is that of tFootChain.
func footChain(p design.Params, xl, xr, reach, tipHeight, tipThick float64, sh tFootShape) chain {
	low := -p.OvershootLower
	xb := xr + sh.LowRight
	ib := low + p.ThinLower

	// The tip of the tail: a cut at right angles to the direction of the stroke,
	// tipThick long, whose two corners are rounded with the radius sh.TipRound.  The
	// rounded outer corner is the rightmost ink and the rounded upper corner the
	// highest.
	th := sh.TipAngle * math.Pi / 180
	dir := outline.Point{X: math.Cos(th), Y: math.Sin(th)}
	e := outline.Point{X: -dir.Y, Y: dir.X} // along the cut, from the outer corner to the upper one
	rho := sh.TipRound
	// The centre of the circle which rounds the outer corner lies at the corner less
	// rho along the direction of the stroke plus rho along the cut; its right
	// extreme is the rightmost ink, and the upper corner is higher than the outer one
	// by what the cut rises.
	cornerOuter := outline.Point{
		X: xr + reach - rho + rho*(dir.X-e.X),
		Y: tipHeight - rho - (tipThick-rho)*e.Y + rho*dir.Y,
	}
	cornerInner := outline.Point{X: cornerOuter.X + tipThick*e.X, Y: cornerOuter.Y + tipThick*e.Y}
	tipOuter := outline.Point{X: cornerOuter.X - rho*dir.X, Y: cornerOuter.Y - rho*dir.Y}
	tipInner := outline.Point{X: cornerInner.X - rho*dir.X, Y: cornerInner.Y - rho*dir.Y}
	// The four points where the straight cut and the rounded corners meet.
	outerEnd := outline.Point{X: cornerOuter.X + rho*e.X, Y: cornerOuter.Y + rho*e.Y}
	innerStart := outline.Point{X: cornerInner.X - rho*e.X, Y: cornerInner.Y - rho*e.Y}
	const kappa = 0.5522847498
	round := func(p0, d0, d1 outline.Point) cubic {
		// the quarter circle which turns from the direction d0 to d1, p0 to p0
		// + rho(d0 + d1)
		return cubic{p0,
			outline.Point{X: p0.X + kappa*rho*d0.X, Y: p0.Y + kappa*rho*d0.Y},
			outline.Point{X: p0.X + rho*(d0.X+d1.X) - kappa*rho*d1.X, Y: p0.Y + rho*(d0.Y+d1.Y) - kappa*rho*d1.Y},
			outline.Point{X: p0.X + rho*(d0.X+d1.X), Y: p0.Y + rho*(d0.Y+d1.Y)}}
	}
	negDir := outline.Point{X: -dir.X, Y: -dir.Y}
	roundOuter := round(tipOuter, dir, e)
	roundInner := round(innerStart, e, negDir)

	c1 := cubic{
		P0: outline.Point{X: xl, Y: sh.OuterStart},
		P1: outline.Point{X: xl + math.Sin(sh.Lean)*sh.OuterA*(sh.OuterStart-low), Y: sh.OuterStart - math.Cos(sh.Lean)*sh.OuterA*(sh.OuterStart-low)},
		P2: outline.Point{X: xb - sh.OuterB*(xb-xl), Y: low},
		P3: outline.Point{X: xb, Y: low},
	}
	run := tipOuter.X - xb
	chord := math.Hypot(tipOuter.X-xb, tipOuter.Y-low)
	c2 := cubic{
		P0: c1.P3,
		P1: outline.Point{X: xb + sh.TailA*run, Y: low},
		P2: outline.Point{X: tipOuter.X - sh.TailB*chord*dir.X, Y: tipOuter.Y - sh.TailB*chord*dir.Y},
		P3: tipOuter,
	}
	chordU := math.Hypot(tipInner.X-xb, tipInner.Y-ib)
	runU := tipInner.X - xb
	c3 := cubic{
		P0: tipInner,
		P1: outline.Point{X: tipInner.X - sh.UpperA*chordU*dir.X, Y: tipInner.Y - sh.UpperA*chordU*dir.Y},
		P2: outline.Point{X: xb + sh.UpperB*runU, Y: ib},
		P3: outline.Point{X: xb, Y: ib},
	}
	c4 := cubic{
		P0: c3.P3,
		P1: outline.Point{X: xb - sh.InnerA*(xb-xr), Y: ib},
		P2: outline.Point{X: xr + math.Sin(sh.Lean)*sh.InnerB*(sh.InnerTop-ib), Y: sh.InnerTop - math.Cos(sh.Lean)*sh.InnerB*(sh.InnerTop-ib)},
		P3: outline.Point{X: xr, Y: sh.InnerTop},
	}
	var ch chain
	for _, c := range []cubic{c1, c2} {
		ch = append(ch, splitAtExtrema(c)...)
	}
	ch = append(ch, splitAtExtrema(roundOuter)...)
	ch = append(ch, lineSeg(outerEnd, innerStart))
	ch = append(ch, splitAtExtrema(roundInner)...)
	for _, c := range []cubic{c3, c4} {
		ch = append(ch, splitAtExtrema(c)...)
	}
	return join(ch)
}

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
	path.LineTo(stem.Left, tFoot.OuterStart)
	tFootChain(p, tTail).appendTo(&path)
	path.LineTo(stem.right(), bar.Top-bar.Thick)
	barRight(&path, bar)
	path.LineTo(stem.right(), stem.Top)
	return path, design.TAdvance(p)
}
