// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
)

// This file holds what follows from the parameters of the letters x, z and k: the
// leans of the diagonals, the position of the crossing of x, the advances, the
// position of the junction of k.  The construction is
// described in internal/glyphs/x.go, z.go and k.go.
//
// As for v, the widths and the serifs are the parameters and the leans follow.

// xTips returns the horizontal positions of the tips of the four outer serifs of x
// relative to the crossing of the centre lines of the upper halves, for the sum of
// the leans s: the top left and top right and the bottom left and bottom right.  At
// the top the diagonals are XHeight - XCross above the crossing and at the bottom
// XCross below it; the lower half of the thin stroke is moved by XThinOffset.
func xTips(p Params, s float64) (tl, tr, bl, br float64) {
	lt, ln := (s-p.XLeanDiff)/2, (s+p.XLeanDiff)/2
	wt, wn := DiagWidth(p.XThick, lt)/2, DiagWidth(p.Hair, ln)/2
	up, down := p.XHeight-p.XCross, p.XCross
	jt, jn := DiagJutOuter(p, p.XOverThick, lt), DiagJutOuter(p, p.XOverThin, ln)
	tl = -(lt*up + wt + jt)
	tr = ln*up + wn + jn
	bl = p.XThinOffset - (ln*down + wn + jn)
	br = lt*down + wt + jt
	return tl, tr, bl, br
}

// xExtent returns the left and the right end of the ink of x relative to the
// crossing.  The tips at the top and at the bottom are almost in line, and which
// of the two stands further out depends on the parameters.
func xExtent(p Params, s float64) (left, right float64) {
	tl, tr, bl, br := xTips(p, s)
	return math.Min(tl, bl), math.Max(tr, br)
}

// XLeanSum returns the sum of the leans of the two diagonals of x, the tangents of
// their angles to the vertical.  The ink of x, from the tip of the outermost
// serif on the left to that on the right (the bottom ones, up to a unit), is XInk
// wide, and its width grows with the sum of the leans, which follows from
// XInk by bisection.  The thin diagonal is steeper than the thick one by XLeanDiff.
// The ink at the top is narrower than at the bottom.
func XLeanSum(p Params) float64 {
	f := func(s float64) float64 {
		l, r := xExtent(p, s)
		return r - l - p.XInk
	}
	lo, hi := 0.3, 3.0
	for i := 0; i < 100; i++ {
		mid := (lo + hi) / 2
		if f(mid) < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// XLeanThick returns the lean of the thick diagonal of x in absolute value; the
// diagonal leans to the left going up.
func XLeanThick(p Params) float64 { return (XLeanSum(p) - p.XLeanDiff) / 2 }

// XLeanThin returns the lean of the thin diagonal of x, which leans to the right
// going up.
func XLeanThin(p Params) float64 { return (XLeanSum(p) + p.XLeanDiff) / 2 }

// XAngleThick returns the angle of the thick diagonal of x to the vertical in degrees.
func XAngleThick(p Params) float64 { return math.Atan(XLeanThick(p)) * 180 / math.Pi }

// XAngleThin returns the angle of the thin diagonal of x to the vertical in degrees.
func XAngleThin(p Params) float64 { return math.Atan(XLeanThin(p)) * 180 / math.Pi }

// XCentre returns the horizontal position of the crossing of the centre lines of
// the upper halves, which is where the centre line of the thick diagonal is at the
// height XCross: the ink begins DiagSideLeft from the edge of the advance.
func XCentre(p Params) float64 {
	l, _ := xExtent(p, XLeanSum(p))
	return p.DiagSideLeft - l
}

// XInkTop returns the width of the ink at the x-height, from the tip of the outer
// serif of the thick diagonal to that of the thin one.
func XInkTop(p Params) float64 {
	tl, tr, _, _ := xTips(p, XLeanSum(p))
	return tr - tl
}

// XInkBottom returns the width of the ink at the baseline, from the tip of the outer
// serif of the thin diagonal to that of the thick one.
func XInkBottom(p Params) float64 {
	_, _, bl, br := xTips(p, XLeanSum(p))
	return br - bl
}

// XAdvance returns the advance of the letter x: the ink at the baseline, the wider
// end, between the side spaces of the diagonal letters.
func XAdvance(p Params) float64 { return p.DiagSideLeft + p.XInk + p.DiagSideRight }

// ZLean returns the lean of the diagonal of z, the tangent of its angle to the
// vertical.  The corner of the diagonal at the x-height lies over the end of the
// lower bar and the corner at the baseline is the left end of the ink, so that the
// ink is ZInk wide: with the thickness t of the diagonal and the height h,
//
//	ZInk = h lean + t sqrt(1 + lean^2),
//
// whose smaller root is
//
//	lean = (h ZInk - t sqrt(h^2 + ZInk^2 - t^2)) / (h^2 - t^2).
func ZLean(p Params) float64 {
	h, w, t := p.XHeight, p.ZInk, p.ZDiag
	return (h*w - t*math.Sqrt(h*h+w*w-t*t)) / (h*h - t*t)
}

// ZAngle returns the angle of the diagonal of z to the vertical in degrees.
func ZAngle(p Params) float64 { return math.Atan(ZLean(p)) * 180 / math.Pi }

// ZAdvance returns the advance of the letter z: the ink between its side spaces.
func ZAdvance(p Params) float64 { return p.ZSideLeft + p.ZInk + p.ZSideRight }

// KArmLean returns the lean of the arm of k, the tangent of its angle to the vertical.
func KArmLean(p Params) float64 { return math.Tan(p.KArmAngle * math.Pi / 180) }

// KLegLean returns the lean of the leg of k in absolute value; the leg leans to the
// left going up.
func KLegLean(p Params) float64 { return math.Tan(p.KLegAngle * math.Pi / 180) }

// KStemRight returns the right-hand edge of the cell of the stem of k: the edge of a
// stem of the width Stem.  The ascender stem is the stem of h and may be lighter,
// standing in the cell with the same centre line (AscenderInset); the arm and the
// leg are measured from the cell.
func KStemRight(p Params) float64 { return p.SideStemLeft + p.Stem }

// KJunctionX returns the horizontal position of the junction of the arm and the
// leg, the point where their centre lines meet, at the height KJunction.  The foot
// serif of the leg ends KReach to the right of the cell of the stem, which fixes
// the junction: the leg runs from the junction to the baseline, KJunction lean to
// the right, has half its horizontal width on each side of its centre line, and has
// a serif which stands KOverLeg beyond its outer edge at the level of its upper face.
func KJunctionX(p Params) float64 {
	wl := DiagWidth(p.KLeg, KLegLean(p))
	return KStemRight(p) + p.KReach - DiagJutOuter(p, p.KOverLeg, KLegLean(p)) - wl/2 - p.KJunction*KLegLean(p)
}

// KAdvance returns the advance of the letter k: the side space and the cell of the
// stem, the reach of the leg and the right side space.
func KAdvance(p Params) float64 { return p.SideStemLeft + p.Stem + p.KReach + p.KSideRight }

// The shape constants of the beaks of z: the angle from the vertical at which
// the inner edge leaves the flat end (in radians) and the horizontal extent of
// the curve of the inner edge into the underside of the bar.
const (
	ZBeakTilt = 5 * math.Pi / 180
	ZBeakRun  = 56.0
)
