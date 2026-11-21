// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
)

// This file holds what follows from the parameters of the diagonal strokes and of
// the letters v, w and y: the thicknesses and angles of the diagonals, the
// advances, the heights of the crotches.  The
// construction is described in internal/glyphs/diagonal.go.
//
// The widths are the parameters and the angles follow from them.  The ink width
// (VInk) and the serifs are given, and the lean of the diagonals, the tangent of
// their angle to the vertical, is the number which makes the outer edges reach
// from the serif of one side to that of the other.  The outer edges of both
// diagonals of a vertex have the same lean: they are mirror images of each other
// in the vertical line through the centre of the vertex.

// The head serifs of the diagonals have a flat top on the line of the x-height
// and are SerifThick thick at the tip.  The overhang of a serif is how far its
// tip stands from the edge of the stroke at the level of the underside of the
// serif (the lower face of a head serif, the upper face of a foot serif),
// measured horizontally.  Both sides of a stroke have the same overhang.  The
// edges of a diagonal lean: the extended edge at the line of the top is
// displaced from the edge at the underside by lean times SerifThick, and the
// reach of the tip from the extended edge at the top (the jut) is shorter than
// the overhang on the outer side, where the edge leans away from the serif
// going down, and longer on the inner side.

// DiagJutOuter returns the reach of the tip of the serif on the outer side of a
// diagonal of the given lean beyond the edge extended to the line of the top, for the
// overhang over.
func DiagJutOuter(p Params, over, lean float64) float64 {
	return over - math.Abs(lean)*p.SerifThick
}

// DiagJutInner returns the same for the serif on the inner side.
func DiagJutInner(p Params, over, lean float64) float64 {
	return over + math.Abs(lean)*p.SerifThick
}

// WThick returns the thickness at right angles of the thick diagonals of w.
func WThick(p Params) float64 { return p.Stem - p.WStemCut }

// DiagHeight returns the height of the x-height above the flat of a vertex.
func DiagHeight(p Params) float64 { return p.XHeight + p.DiagOvershoot }

// DiagWidth returns the horizontal width of a diagonal of the thickness thick
// (at right angles) whose lean, the tangent of its angle to the vertical, is lean.
func DiagWidth(thick, lean float64) float64 { return thick * math.Sqrt(1+lean*lean) }

// VLean returns the lean of the diagonals of v and y.  The outer edges stand
// VInk - DiagJutOuter(thick) - DiagJutOuter(thin) apart at the x-height and
// DiagVertexFlat apart at the flat of the vertex, DiagHeight below it, which gives
//
//	VInk = DiagVertexFlat + 2 lean DiagHeight + DiagOverThick + DiagOverThin
//	       - 2 lean SerifThick.
func VLean(p Params) float64 {
	return (p.VInk - p.DiagVertexFlat - p.DiagOverThick - p.DiagOverThin) / (2 * (DiagHeight(p) - p.SerifThick))
}

// VCentre returns the horizontal position of the middle of the flat of the vertex of
// v and y: the tips of the two outer serifs stand VInk apart, the first
// DiagSideLeft from the edge of the advance, and the outer serifs of the thick and
// of the thin diagonal differ in their overhang.
func VCentre(p Params) float64 {
	return p.DiagSideLeft + p.VInk/2 + (p.DiagOverThick-p.DiagOverThin)/2
}

// VAngle returns the angle of the diagonals of v to the vertical in degrees.
func VAngle(p Params) float64 { return math.Atan(VLean(p)) * 180 / math.Pi }

// VAdvance returns the advance of the letter v: the ink between the side spaces of
// the diagonal letters.
func VAdvance(p Params) float64 { return p.DiagSideLeft + p.VInk + p.DiagSideRight }

// YAdvance returns the advance of the letter y, which is that of v.
func YAdvance(p Params) float64 { return VAdvance(p) }

// WAdvance returns the advance of the letter w.
func WAdvance(p Params) float64 { return p.DiagSideLeft + p.WInk + p.DiagSideRight }

// VCounter returns the width of the counter of v at the x-height, between the
// extended inner edges of the two diagonals.
func VCounter(p Params) float64 {
	l := VLean(p)
	return p.VInk - DiagJutOuter(p, p.DiagOverThick, l) - DiagJutOuter(p, p.DiagOverThin, l) -
		DiagWidth(p.Stem, l) - DiagWidth(p.Hair, l)
}

// crotchHeight returns the height of the inner corner of a vertex whose outer
// edges have the lean lean and meet a flat of the width flat at DiagOvershoot below
// the baseline, for diagonals of the thicknesses thickL and thickR: where the
// right-hand edge of the left diagonal meets the left-hand edge of the right one.
func crotchHeight(p Params, thickL, thickR, lean float64) float64 {
	return (DiagWidth(thickL, lean)+DiagWidth(thickR, lean)-p.DiagVertexFlat)/(2*lean) - p.DiagOvershoot
}

// VCrotch returns the height of the crotch of v, the inner corner of the vertex.
// It is also the height at which the two diagonals of y meet.
func VCrotch(p Params) float64 { return crotchHeight(p, p.Stem, p.Hair, VLean(p)) }

// WLean returns the lean of the four diagonals of w.  The two vertices stand a
// period apart, the middle apex is cut flat WApexFlat wide at WApexHeight, and the
// outer edges of the two outer diagonals reach from the serif on one side to that of
// the other over WInk, so that the lean is the solution of
//
//	WInk - DiagOverThick - DiagOverThin + 2 lean SerifThick = WApexFlat + 2 DiagVertexFlat
//	    + 2 lean (XHeight + WApexHeight + 2 DiagOvershoot) - (Hair + WThick) sqrt(1 + lean^2)
//
// whose right-hand side grows with the lean, found by bisection.
func WLean(p Params) float64 {
	f := func(l float64) float64 {
		want := p.WInk - p.DiagOverThick - p.DiagOverThin + 2*l*p.SerifThick
		return p.WApexFlat + 2*p.DiagVertexFlat + 2*l*(p.XHeight+p.WApexHeight+2*p.DiagOvershoot) -
			(p.Hair+WThick(p))*math.Sqrt(1+l*l) - want
	}
	lo, hi := 0.05, 1.0
	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		if f(mid) < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// WAngle returns the angle of the diagonals of w to the vertical in degrees.
func WAngle(p Params) float64 { return math.Atan(WLean(p)) * 180 / math.Pi }

// WPeriod returns the distance between the centres of the two vertices of w.  At
// the apex the left-hand edge of the thin diagonal and the right-hand edge of the
// thick one are WApexFlat apart, which fixes it.
func WPeriod(p Params) float64 {
	l := WLean(p)
	return p.WApexFlat + p.DiagVertexFlat + 2*l*(p.WApexHeight+p.DiagOvershoot) -
		DiagWidth(p.Hair, l) - DiagWidth(WThick(p), l)
}

// WCrotch returns the height of the crotch of each of the two vertices of w.
func WCrotch(p Params) float64 { return crotchHeight(p, WThick(p), p.Hair, WLean(p)) }

// WApexNotch returns the height of the notch under the apex of w, where the
// right-hand edge of the thin diagonal on its left meets the left-hand edge of the
// thick diagonal on its right.
func WApexNotch(p Params) float64 {
	return (WPeriod(p)-p.DiagVertexFlat)/(2*WLean(p)) - p.DiagOvershoot
}

// DiagBracket is the size of the bracket of a serif on a diagonal: Reach is the
// distance from the edge, along the face of the serif, at which the bracket leaves
// the face, and Obtuse and Acute are the vertical distances below the face (above it
// for a foot serif) at which it meets the edge where the white corner between the
// face and the edge is obtuse (the edge leans away from the serif as it leaves the
// face) and where it is acute.
type DiagBracket struct {
	Reach, Obtuse, Acute float64
	ReachAcute           float64 // the reach on the acute side
}

// The bracket of a serif on a diagonal is that of the serif on a stem adapted to the
// inclined edge.  A stem bracket rises BracketHeight - SerifThick above the face of the
// serif and reaches BracketReach along it.  On a diagonal the bracket meets the edge
// DiagBracketObtuse times that rise from the face on the obtuse side and
// DiagBracketAcute times it on the acute side.
// The reach is BracketReach on the obtuse side and DiagBracketReachAcute of it on the
// acute side, but at most DiagBracketReachMax of the overhang.
const (
	DiagBracketObtuse     = 1.75
	DiagBracketAcute      = 1.1
	DiagBracketReachAcute = 0.8
	DiagBracketReachMax   = 0.8
)

// diagBracketFor returns the bracket of the serifs on the diagonals for a stem
// bracket of the height height and the reach reach on serifs of the thickness thick.
func diagBracketFor(height, reach, thick float64) DiagBracket {
	rise := height - thick
	return DiagBracket{
		Reach: reach, Obtuse: DiagBracketObtuse * rise,
		ReachAcute: DiagBracketReachAcute * reach, Acute: DiagBracketAcute * rise,
	}
}

// LowerDiagBracket returns the bracket of the serifs on the diagonals of the
// lowercase, from the stem bracket of the lowercase.
func LowerDiagBracket(p Params) DiagBracket {
	return diagBracketFor(p.BracketHeight, p.BracketReach, p.SerifThick)
}
