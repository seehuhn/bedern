// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import "math"

// This file holds the parameters of the capitals built from diagonal strokes (A,
// V, W, X, Y, Z, K, M and N) and what follows from them: the leans of the diagonals,
// the positions of the vertices and the advances.  The construction is described
// in internal/glyphs/capdiag.go and the files of the letters.
//
// As for the lowercase v, w and x (diag.go, xzk.go), the widths of the ink and
// the serifs are the parameters and the angles follow from them.  Every ink
// width is measured at the level of the serifs from the outermost tip on the
// left to the outermost tip on the right, and the side spaces are the distances
// from the edge of the advance to those tips.

// CapSides is the pair of side spaces of a letter: the distances from the edges of
// the advance to the outermost tips of the serifs.
type CapSides struct {
	Left, Right float64
}

// CapDiag holds the parameters of the diagonal capitals.
type CapDiag struct {
	// Strokes.  Thick and Thin are the thick and the thin diagonals of A, V and
	// Y, the thin one also those of W, X, M and the arm of K.  The thick diagonals
	// of W, X and M are thinner where strokes cross or meet (WThick,
	// XThick, MThick), and NThick, KLeg and ZDiag are those of N, the leg of K and the
	// diagonal of Z.  ThinStem is the thin stem of M and N and MStem the thick stem
	// of M, on the right.  All are measured at right angles to the stroke.
	Thick, Thin float64
	WThick      float64
	XThick      float64
	MThick      float64
	NThick      float64
	KLeg        float64
	ZDiag       float64
	ThinStem    float64
	MStem       float64

	// Serifs.  OverThick and OverThin are the overhangs of the serifs on the
	// thick and on the thin diagonals, measured horizontally from the edge to the
	// tip at the level of the face with the bracket; both sides of a stroke have
	// the same overhang.  KOverLeg is that of the foot of the leg of K and KOverArm that
	// of the head of its arm.
	OverThick, OverThin float64
	KOverLeg, KOverArm  float64

	// Vertices.  VertexFlat is the width of the flat of the lower vertex of V and of
	// W and VertexDepth the depth of the flat below the baseline.  ApexFlat is the
	// width of the flat at the top of A and ApexRise the height by which it stands
	// above the cap height.  WApexFlat and WApexHeight are the width and the height of
	// the flat of the middle apex of W.  NFlat and NDepth are those of the vertex of
	// N at the lower right, and NJoin is the distance below the cap height at which
	// the lower edge of the diagonal of N leaves the left stem.  MFlat and MHeight are those of
	// the vertex of M, which lies above the baseline.
	VertexFlat, VertexDepth float64
	ApexFlat, ApexRise      float64
	WApexFlat, WApexHeight  float64
	NFlat, NDepth, NJoin    float64
	MFlat, MHeight          float64

	// MCrotchDepth is the distance below the cap height of the point of each white
	// wedge between a stem of M and a diagonal, where the outer edge of the diagonal
	// leaves the inner edge of the stem: the ink over it is that deep.
	MCrotchDepth float64

	// Heights.  ABar is the height of the centre of the crossbar of A, XCross the
	// height at which the centre lines of the upper halves of X cross, YJunction the
	// height of the crotch of Y, KJunction the height at which the centre lines of
	// the arm and the leg of K meet.
	ABar      float64
	XCross    float64
	YJunction float64
	KJunction float64

	// Angles and offsets.  KArmAngle and KLegAngle are the angles of the arm and the
	// leg of K to the vertical in degrees, XThinOffset the offset of the lower half
	// of the thin diagonal of X (negative: to the left) and XLeanDiff the amount by
	// which the thin diagonal of X is steeper than the thick one.
	KArmAngle, KLegAngle float64
	XThinOffset          float64
	XLeanDiff            float64

	// Widths.  The ink widths, see above; ZInkTop is the length of the upper bar of
	// Z from the right-hand corner to the cut of the beak.
	AInk, VInk, WInk, XInk, YInk, ZInk, KInk, MInk, NInk float64
	ZInkTop                                              float64

	// Side spaces of the letters.
	SideA, SideV, SideW, SideX, SideY, SideZ, SideK, SideM, SideN CapSides
}

// regularCapDiag returns the parameters of the diagonal capitals of Bedern Regular.
func regularCapDiag() CapDiag {
	return CapDiag{
		Thick:    93,
		Thin:     42,
		WThick:   91,
		XThick:   92,
		MThick:   86,
		NThick:   92,
		KLeg:     90,
		ZDiag:    92,
		ThinStem: 44,
		MStem:    94,

		OverThick: 66,
		OverThin:  80,
		KOverLeg:  60,
		KOverArm:  70,

		VertexFlat: 28, VertexDepth: 10,
		ApexFlat: 28, ApexRise: 12,
		WApexFlat: 30, WApexHeight: 679,
		NFlat: 24, NDepth: 8, NJoin: 55,
		MFlat: 24, MHeight: 36,

		MCrotchDepth: 75,

		ABar:      244,
		XCross:    362,
		YJunction: 321,
		KJunction: 403,

		KArmAngle:   46,
		KLegAngle:   34,
		XThinOffset: -6,
		XLeanDiff:   0.027,

		AInk: 664, VInk: 690, WInk: 961, XInk: 681, YInk: 705, ZInk: 489, KInk: 682, MInk: 817, NInk: 663,
		ZInkTop: 461,

		SideA: CapSides{25, 27},
		SideV: CapSides{15, 14},
		SideW: CapSides{12, 12},
		SideX: CapSides{16, 14},
		SideY: CapSides{8, 6},
		SideZ: CapSides{43, 46},
		SideK: CapSides{29, 33},
		SideM: CapSides{29, 28},
		SideN: CapSides{29, 28},
	}
}

// The advance of a letter is its ink between its side spaces.

// CapAAdvance returns the advance of A.
func CapAAdvance(p Params) float64 { d := p.CapDiag; return d.SideA.Left + d.AInk + d.SideA.Right }

// CapVAdvance returns the advance of V.
func CapVAdvance(p Params) float64 { d := p.CapDiag; return d.SideV.Left + d.VInk + d.SideV.Right }

// CapWAdvance returns the advance of W.
func CapWAdvance(p Params) float64 { d := p.CapDiag; return d.SideW.Left + d.WInk + d.SideW.Right }

// CapXAdvance returns the advance of X.
func CapXAdvance(p Params) float64 { d := p.CapDiag; return d.SideX.Left + d.XInk + d.SideX.Right }

// CapYAdvance returns the advance of Y.
func CapYAdvance(p Params) float64 { d := p.CapDiag; return d.SideY.Left + d.YInk + d.SideY.Right }

// CapZAdvance returns the advance of Z.
func CapZAdvance(p Params) float64 { d := p.CapDiag; return d.SideZ.Left + d.ZInk + d.SideZ.Right }

// CapKAdvance returns the advance of K.
func CapKAdvance(p Params) float64 { d := p.CapDiag; return d.SideK.Left + d.KInk + d.SideK.Right }

// CapMAdvance returns the advance of M.
func CapMAdvance(p Params) float64 { d := p.CapDiag; return d.SideM.Left + d.MInk + d.SideM.Right }

// CapNAdvance returns the advance of N.
func CapNAdvance(p Params) float64 { d := p.CapDiag; return d.SideN.Left + d.NInk + d.SideN.Right }

// capDiagWidth is DiagWidth.
func capDiagWidth(thick, lean float64) float64 { return DiagWidth(thick, lean) }

// capCrotchHeight returns the height of the inner corner of a vertex whose outer
// edges have the lean lean in absolute value and meet a flat of the width flat at
// the height y0 (the diagonals having the thicknesses thickL and thickR): where the
// right-hand edge of the left diagonal meets the left-hand edge of the right one.
// The same formula gives, mirrored in a horizontal line, the depth of the inner
// corner under the apex of A.
func capCrotchHeight(flat, y0, thickL, thickR, lean float64) float64 {
	return y0 + (capDiagWidth(thickL, lean)+capDiagWidth(thickR, lean)-flat)/(2*lean)
}

// CapALean returns the lean of the diagonals of A: the outer edges stand ApexFlat
// apart at the top of the apex, ApexRise above the cap height, and AInk apart over the
// tips of the foot serifs, whose faces lie SerifThick above the baseline.
func CapALean(p Params) float64 {
	d := p.CapDiag
	return (d.AInk - d.ApexFlat - d.OverThick - d.OverThin) / (2 * (p.CapHeight + d.ApexRise - p.Caps.SerifThick))
}

// CapACentre returns the horizontal position of the middle of the flat of the apex of
// A: the tip of the serif of the thin left diagonal is SideA.Left from the edge.
func CapACentre(p Params) float64 {
	d := p.CapDiag
	return d.SideA.Left + d.OverThin + CapALean(p)*(p.CapHeight+d.ApexRise-p.Caps.SerifThick) + d.ApexFlat/2
}

// CapVLean returns the lean of the diagonals of V: the outer edges stand VertexFlat
// apart at the depth VertexDepth below the baseline and VInk apart over the tips of
// the head serifs, whose undersides lie SerifThick below the cap height.
func CapVLean(p Params) float64 {
	d := p.CapDiag
	return (d.VInk - d.VertexFlat - d.OverThick - d.OverThin) / (2 * (p.CapHeight + d.VertexDepth - p.Caps.SerifThick))
}

// CapVCentre returns the horizontal position of the middle of the flat of the vertex
// of V: the tip of the serif of the thick left diagonal is SideV.Left from the edge.
func CapVCentre(p Params) float64 {
	d := p.CapDiag
	return d.SideV.Left + d.OverThick + CapVLean(p)*(p.CapHeight+d.VertexDepth-p.Caps.SerifThick) + d.VertexFlat/2
}

// CapWLean returns the lean of the four diagonals of W, which follows from the width
// of the ink in the way that of the lowercase w does (WLean):
//
//	WInk - OverThick - OverThin + 2 lean SerifThick = WApexFlat + 2 VertexFlat
//	    + 2 lean (CapHeight + WApexHeight + 2 VertexDepth) - (Thin + WThick) sqrt(1 + lean^2)
func CapWLean(p Params) float64 {
	d := p.CapDiag
	f := func(l float64) float64 {
		want := d.WInk - d.OverThick - d.OverThin + 2*l*p.Caps.SerifThick
		return d.WApexFlat + 2*d.VertexFlat + 2*l*(p.CapHeight+d.WApexHeight+2*d.VertexDepth) -
			(d.Thin+d.WThick)*math.Sqrt(1+l*l) - want
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

// CapWPeriod returns the distance between the centres of the two vertices of W; at the
// apex the left-hand edge of the thin diagonal and the right-hand edge of the thick
// one are WApexFlat apart.
func CapWPeriod(p Params) float64 {
	d := p.CapDiag
	l := CapWLean(p)
	return d.WApexFlat + d.VertexFlat + 2*l*(d.WApexHeight+d.VertexDepth) -
		capDiagWidth(d.Thin, l) - capDiagWidth(d.WThick, l)
}

// CapWCentre returns the horizontal position of the middle of the flat of the first
// vertex of W.
func CapWCentre(p Params) float64 {
	d := p.CapDiag
	return d.SideW.Left + d.OverThick + CapWLean(p)*(p.CapHeight+d.VertexDepth-p.Caps.SerifThick) + d.VertexFlat/2
}

// xTipsCap is xTips for the capital X: the horizontal positions of the tips of the four
// outer serifs relative to the crossing of the centre lines of the upper halves, for
// the sum of the leans s.
func xTipsCap(p Params, s float64) (tl, tr, bl, br float64) {
	d := p.CapDiag
	lt, ln := (s-d.XLeanDiff)/2, (s+d.XLeanDiff)/2
	wt, wn := capDiagWidth(d.XThick, lt)/2, capDiagWidth(d.Thin, ln)/2
	up, down := p.CapHeight-d.XCross, d.XCross
	jt, jn := capJutOuter(p, d.OverThick, lt), capJutOuter(p, d.OverThin, ln)
	tl = -(lt*up + wt + jt)
	tr = ln*up + wn + jn
	bl = d.XThinOffset - (ln*down + wn + jn)
	br = lt*down + wt + jt
	return tl, tr, bl, br
}

// capJutOuter is DiagJutOuter with the serif thickness of the capitals: the reach of
// the tip of the serif on the outer side of a diagonal beyond the edge extended to the
// line of the top, for the overhang over.
func capJutOuter(p Params, over, lean float64) float64 {
	return over - math.Abs(lean)*p.Caps.SerifThick
}

// capXExtent returns the left and the right end of the ink of X relative to the
// crossing.
func capXExtent(p Params, s float64) (left, right float64) {
	tl, tr, bl, br := xTipsCap(p, s)
	return math.Min(tl, bl), math.Max(tr, br)
}

// CapXLeanSum returns the sum of the leans of the two diagonals of X; the ink is XInk
// wide from the outermost tip on the left to that on the right.
func CapXLeanSum(p Params) float64 {
	f := func(s float64) float64 {
		l, r := capXExtent(p, s)
		return r - l - p.CapDiag.XInk
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

// CapXLeanThick returns the lean of the thick diagonal of X in absolute value.
func CapXLeanThick(p Params) float64 { return (CapXLeanSum(p) - p.CapDiag.XLeanDiff) / 2 }

// CapXLeanThin returns the lean of the thin diagonal of X.
func CapXLeanThin(p Params) float64 { return (CapXLeanSum(p) + p.CapDiag.XLeanDiff) / 2 }

// CapXCentre returns the horizontal position of the crossing of the centre lines of the
// upper halves of X.
func CapXCentre(p Params) float64 {
	l, _ := capXExtent(p, CapXLeanSum(p))
	return p.CapDiag.SideX.Left - l
}

// capYGeometry returns the lean of the diagonals of Y and the height at which their
// outer edges meet the edges of the stem, for the junction (the crotch) at
// YJunction and the ink YInk wide over the tips of the head serifs.  The outer
// edges pass through the ends of a flat as wide as the stem, which is where the
// stem begins, so that the crotch lies (wL + wR - Stem) / (2 lean) above it.
func capYGeometry(p Params) (lean, joint float64) {
	d := p.CapDiag
	stem := p.Caps.Stem
	f := func(l float64) float64 {
		yj := d.YJunction - (capDiagWidth(d.Thick, l)+capDiagWidth(d.Thin, l)-stem)/(2*l)
		return stem + 2*l*(p.CapHeight-p.Caps.SerifThick-yj) + d.OverThick + d.OverThin - d.YInk
	}
	lo, hi := 0.15, 1.5
	for i := 0; i < 100; i++ {
		mid := (lo + hi) / 2
		if f(mid) < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	lean = (lo + hi) / 2
	joint = d.YJunction - (capDiagWidth(d.Thick, lean)+capDiagWidth(d.Thin, lean)-stem)/(2*lean)
	return lean, joint
}

// CapYLean returns the lean of the diagonals of Y.
func CapYLean(p Params) float64 { l, _ := capYGeometry(p); return l }

// CapYJoint returns the height at which the outer edges of the diagonals of Y meet
// the edges of the stem.
func CapYJoint(p Params) float64 { _, j := capYGeometry(p); return j }

// CapYCentre returns the horizontal position of the middle of the stem of Y.
func CapYCentre(p Params) float64 {
	d := p.CapDiag
	l, j := capYGeometry(p)
	return d.SideY.Left + d.OverThick + p.Caps.Stem/2 + l*(p.CapHeight-p.Caps.SerifThick-j)
}

// CapZLean returns the lean of the diagonal of Z (see ZLean): the corner at the cap
// height lies over the end of the lower bar and the corner at the baseline is the
// left end of the ink.
func CapZLean(p Params) float64 {
	h, w, t := p.CapHeight, p.CapDiag.ZInk, p.CapDiag.ZDiag
	return (h*w - t*math.Sqrt(h*h+w*w-t*t)) / (h*h - t*t)
}

// CapKArmLean returns the lean of the arm of K, the tangent of its angle to the vertical.
func CapKArmLean(p Params) float64 { return math.Tan(p.CapDiag.KArmAngle * math.Pi / 180) }

// CapKLegLean returns the lean of the leg of K in absolute value.
func CapKLegLean(p Params) float64 { return math.Tan(p.CapDiag.KLegAngle * math.Pi / 180) }

// CapKStemLeft returns the position of the left-hand edge of the stem of K.
func CapKStemLeft(p Params) float64 { return p.CapDiag.SideK.Left + p.Caps.SerifJut }

// CapKJunctionX returns the horizontal position of the junction of the centre lines of
// the arm and the leg of K: the tip of the foot serif of the leg is KInk from the tip
// of the serif on the left of the stem.
func CapKJunctionX(p Params) float64 {
	d := p.CapDiag
	l := CapKLegLean(p)
	right := d.SideK.Left + d.KInk
	return right - d.KOverLeg - capDiagWidth(d.KLeg, l)/2 - l*(d.KJunction-p.Caps.SerifThick)
}

// CapMStemLeft returns the position of the left-hand edge of the left stem of M.
func CapMStemLeft(p Params) float64 { return p.CapDiag.SideM.Left + p.Caps.SerifJut }

// CapMStemRight returns the position of the right-hand edge of the right stem of M.
func CapMStemRight(p Params) float64 {
	return p.CapDiag.SideM.Left + p.CapDiag.MInk - p.Caps.SerifJut
}

// CapMLean returns the lean of the diagonals of M in absolute value.  The outer edges
// of the diagonals leave the inner edges of the stems MCrotchDepth below the cap
// height and end in the flat of the vertex, MFlat wide, at MHeight.
func CapMLean(p Params) float64 {
	d := p.CapDiag
	x1 := CapMStemLeft(p) + d.ThinStem
	x0 := CapMStemRight(p) - d.MStem
	return (x0 - x1 - d.MFlat) / (2 * (p.CapHeight - d.MCrotchDepth - d.MHeight))
}

// CapNStemLeft returns the position of the left-hand edge of the left stem of N.
func CapNStemLeft(p Params) float64 { return p.CapDiag.SideN.Left + p.Caps.SerifJut }

// CapNStemRight returns the position of the right-hand edge of the right stem of N.
func CapNStemRight(p Params) float64 {
	return p.CapDiag.SideN.Left + p.CapDiag.NInk - p.Caps.SerifJut
}

// CapNLean returns the lean of the diagonal of N in absolute value: its left-hand
// edge leaves the right-hand edge of the left stem NJoin below the cap height and ends
// in the left end of the flat of the vertex, NFlat from the right-hand edge of the right
// stem and NDepth below the baseline.
func CapNLean(p Params) float64 {
	d := p.CapDiag
	x1 := CapNStemLeft(p) + d.ThinStem
	return (CapNStemRight(p) - d.NFlat - x1) / (p.CapHeight - d.NJoin + d.NDepth)
}
