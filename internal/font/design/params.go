// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package design holds the numeric design parameters of the Bedern fonts.
// All lengths are in font units, with 1000 units per em.
package design

import (
	"fmt"
	"math"
	"reflect"
)

// A Bowl holds the values of a bowl which hangs on a stem, as in d, a, b, p and
// q: each of these letters holds a value of this type.  The free side of the
// bowl is the half of an o: the same thin parts, overshoot and squareness, and
// the extremes at top and bottom lie Width less Run from the free extreme.  A
// field which is zero takes the value of the o.
type Bowl struct {
	// Width is the distance from the free extreme of the bowl to the edge of
	// the stem it hangs on.
	Width float64
	// JoinUpper and JoinLower are the heights at which the edge of the counter
	// meets the stem above and below; NotchUpper and NotchLower those at which
	// the outer edge of the bowl meets it.
	JoinUpper, JoinLower   float64
	NotchUpper, NotchLower float64
	// Run is the distance from the extremes of the bowl at top and bottom to the
	// stem; zero if the free side is the half of the o, which has a run of its
	// own (Width less half the width of the o).
	Run float64
	// Thick is the thickness of the stroke at the free extreme; zero for that of
	// the o (CurveLower).
	Thick float64
	// ThickPos is the height of the free extreme as a fraction of the height of
	// the bowl from the bottom; zero for the middle, as in the o.
	ThickPos float64
	// Top is the height of the top of the bowl, at the extreme at the top and at
	// the stem; zero for that of the o, with its overshoot.  The bottom has the
	// overshoot of the o.
	Top float64
}

// BowlHalf returns the distance from the free extreme of the bowl b to the
// extremes at top and bottom.
func (p Params) BowlHalf(b Bowl) float64 {
	if b.Run == 0 {
		return p.OBowlWidth / 2
	}
	return b.Width - b.Run
}

// BowlThick returns the thickness of the stroke of the bowl b at its free
// extreme.
func (p Params) BowlThick(b Bowl) float64 {
	if b.Thick == 0 {
		return p.CurveLower
	}
	return b.Thick
}

// BowlTop returns the height of the top of the bowl b.
func (p Params) BowlTop(b Bowl) float64 {
	if b.Top == 0 {
		return p.XHeight + p.OvershootLower
	}
	return b.Top
}

// Params is one set of design parameters.
type Params struct {
	UnitsPerEm float64

	// Vertical dimensions.
	XHeight        float64
	CapHeight      float64
	Ascender       float64
	Descender      float64 // depth below the baseline, positive
	OvershootLower float64 // overshoot of round lowercase letters
	OvershootUpper float64 // overshoot of round capitals

	// Line metrics of the font file.
	TypoAscender  float64
	TypoDescender float64 // negative
	LineGap       float64

	// Strokes.  Stem is the stem of n and of the other letters of x-height.
	// AscenderStem is the stem of the ascender letters l, d and the left stem
	// of h; it is Stem in the regular font and may be lighter.  The lighter
	// stem stands in the cell of a stem of Stem, with the same centre line; see
	// AscenderInset in fit.go.
	Stem         float64 // lowercase stem
	AscenderStem float64 // stem of l, d and the left stem of h, at most Stem
	CurveLower   float64 // thickest part of round lowercase strokes
	ThinLower    float64 // thinnest part of round lowercase strokes
	Hair         float64 // thin strokes which are not parts of a round letter: the crossbar of f and t, and the thin diagonals of v, w, x, y and k

	// Serifs.
	SerifJut      float64 // serif length beyond the stem, lowercase
	SerifThick    float64 // serif thickness at the tip
	BracketHeight float64 // height at which the bracket leaves the stem
	BracketReach  float64 // distance from the stem at which the bracket meets the serif

	// SerifJutInner is the length of the foot serifs of n which point into
	// the counter.  HeadSerifDrop is how far the tip of the sloped head serif
	// lies below the top of the stem.
	SerifJutInner float64
	HeadSerifDrop float64

	// The arch of n.  The heights are measured from the baseline.  ArchNotch
	// is the height at which the upper edge of the arch leaves the left stem,
	// ArchJoin the height at which the underside meets it.  ArchThinJoin is
	// the thickness of the arch where it leaves the stem, ArchPeak the
	// position of its highest point as a fraction of the distance between the
	// stem centres, measured from the left stem, and ArchOvershoot the amount
	// by which that point stands above the x-height.  ShoulderStart and
	// ShoulderStartInner are the heights down from which the outer and the
	// inner edge of the right stem are straight.
	ArchNotch          float64
	ArchJoin           float64
	ArchThinJoin       float64
	ArchPeak           float64
	ArchOvershoot      float64
	ShoulderStart      float64
	ShoulderStartInner float64

	// SupernessOuter and SupernessInner are the shape parameters of the outer
	// and the inner contour of the bowls: the point of a quarter arc at 45
	// degrees lies at this fraction of the distance from the centre to the
	// corner of the bounding box.  The value 1/sqrt(2) gives an ellipse and
	// larger values give squarer shapes.
	SupernessOuter float64
	SupernessInner float64

	// The dot of i.  DotDiameter is the width and the height of the dot,
	// DotCentre the height of its centre above the baseline, DotShift the
	// fraction of the length of the head serif (SerifJut) by which the centre
	// lies to the left of the centre of the stem (see DotOffset), and
	// DotSuperness its squareness in the sense of SupernessOuter: 1/sqrt(2) =
	// 0.7071 gives a circle.
	DotDiameter  float64
	DotCentre    float64
	DotShift     float64
	DotSuperness float64

	// The bowls which hang on a stem: DBowl for d, ABowl for a, BBowl for b and
	// PBowl for p; q has the bowl of d.  b and p stand on the right of their
	// stem, as the mirror image of the bowl of d, and have their own heights of
	// the notches and joins.  BowlJoinThick is shared by all of them: the
	// thickness of the stroke, at right angles to it, where it runs into the
	// stem.  The other values of a bowl are the fields of a Bowl.  The free side
	// of the bowl is that of the o, with the same thin parts, overshoot and
	// squareness.
	BowlJoinThick float64
	DBowl         Bowl
	ABowl         Bowl
	BBowl         Bowl
	PBowl         Bowl
	QBowl         Bowl

	// The letter e.  EBarHeight is the height of the vertical centre of the
	// bar and EBarThick its thickness, so that its lower edge lies at
	// EBarHeight - EBarThick/2 and its upper edge at EBarHeight + EBarThick/2.
	// EBowlWidth is the width of the bowl from its left to its right extreme,
	// EThickPos the height of the left extreme, where the stroke is thickest,
	// as a fraction of the height of the bowl from its bottom.  EAperture is
	// the vertical distance from the underside of the bar to the tip of the
	// lower terminal, and ETerminalTip the thickness of the terminal at the cut
	// at its tip.  EEyeSuperness is the squareness of the eye above the bar, whose
	// sides are vertical where they leave the upper edge of the bar.  The thin
	// parts, the thickest part, the overshoot and the squareness of the outer
	// edge are those of the o.
	EBarHeight    float64
	EBarThick     float64
	EBowlWidth    float64
	EThickPos     float64
	EEyeSuperness float64
	EAperture     float64
	ETerminalTip  float64

	// The letter c is the curve of the o open on the right.  CBowlWidth is the
	// width of its ink, from the left extreme to the tip of the lower terminal,
	// which is the rightmost ink; the super-ellipse of the bowl is wider by the
	// few units by which the tip lies inside its right extreme.  The left
	// extreme, the thin parts, the thickest part (EThickPos of the height), the
	// overshoots and the squareness are those of e, and so is the lower
	// terminal, which ends CTipHeight above the baseline and is ETerminalTip
	// thick at its tip.  The upper stroke ends in the ball terminal, whose stalk
	// is the stroke of the o and whose rightmost ink lies CBallInset inside the
	// rightmost ink of the letter.
	CBowlWidth float64
	CTipHeight float64
	CBallInset float64

	// The letter a.  Its bowl, ABowl, hangs on the left of the stem as in d,
	// smaller: ABowl.Width is the distance from its free extreme to the
	// left-hand edge of the stem and ABowl.Top the height of its top, at the
	// extreme at the top and at the stem (the bottom has the overshoot of the
	// o).  ABowl.Thick is the thickest part of the free side, at the fraction
	// ABowl.ThickPos of the height of the bowl from the bottom, as in e.  The
	// thin parts of the bowl and the thickness at the joins (BowlJoinThick) are
	// those of the o and d.  AArcReach is the distance by
	// which the left end of the ball lies inside the free extreme of the bowl.
	// BallDiameter and BallStalk belong to the ball terminal of a, c, f, j, r and
	// y (internal/glyphs/ball.go): the diameter of the ball and the thickness of
	// the stalk where it leaves the ball.  ASideLeft is the side space of a to the free
	// extreme of the bowl.  The stem of a ends in a tail: the foot of t on a
	// smaller scale.  ATailReach is the distance from the right-hand edge of
	// the stem to the rightmost ink of the tail, ATailTipHeight the height of
	// the highest ink of its tip and ATailTipThick its thickness there;
	// ATailSideRight is the space from the tail to the edge of the advance.  The
	// other parameters of the foot are those of t.
	AArcReach    float64
	BallDiameter float64
	BallStalk    float64
	ASideLeft    float64

	ATailReach     float64
	ATailTipHeight float64
	ATailTipThick  float64
	ATailSideRight float64

	// The letter s.  The spine is SSpineThick thick, measured at right angles
	// to it at its centre, and runs down to the right at SSpineAngle degrees
	// below the horizontal there.  SArcThick is the thickness of the arcs at the
	// top and at the bottom and SSideThick the thickness at the extremes of the
	// two bowls (the left of the upper and the right of the lower one).
	// SLowerWidth and SUpperWidth are the widths of the ink of the lower and the
	// upper half, from the left to the right extreme of each, and SUpperShift how
	// far the centre of the upper half lies to the left of that of the lower.
	// The terminals are serif beaks of the family of those of S: each arc ends
	// in a corner on a vertical cut SBeakDrop from the top (bottom) of the
	// letter, where it is SBeakArc thick at right angles to its edge, and a serif
	// hangs from the corner to the end of the terminal, SUpperDrop below the top
	// (SLowerDrop above the bottom), SSerifThick wide halfway along the cut.
	// SSideLeft and SSideRight are the side spaces, measured to the leftmost and
	// rightmost ink.
	SSpineThick float64
	SSpineAngle float64
	SArcThick   float64
	// SUpperArcThick is the thickness of the upper arc at the top of the letter,
	// which runs into the upper beak.  SArcThick is that of the lower arc.
	SUpperArcThick float64
	SSideThick     float64
	SLowerWidth    float64
	SUpperWidth    float64
	SUpperShift    float64
	SUpperDrop     float64
	SLowerDrop     float64
	SBeakDrop      float64
	SBeakLift      float64
	SBeakRun       float64
	SBeakArc       float64
	SSerifThick    float64
	SSideLeft      float64
	SSideRight     float64

	// The letter m: three stems and two arches of n.  MCounter is the width of
	// each of the two counters, between the inner edges of neighbouring stems.
	// The inner foot serifs are those of n (SerifJutInner) and keep the
	// distance MCounter - 2 SerifJutInner between their tips.
	MCounter float64

	// The letter u: the arch of n turned through half a turn between two stems.
	// UCounter is the width of its counter.  The heights are measured from the
	// baseline: UJoin is the
	// height at which the inner edge of the bowl meets the right stem and UNotch
	// that at which its outer edge meets it, UShoulder the height down from
	// which the left edge of the left stem is straight, and UShoulderInner the
	// height up from which its right edge is.  The thickness of the arch at its
	// lowest point is ThinLower, its thickness at the right join ArchThinJoin and
	// its overshoot below the baseline OvershootLower.
	UCounter       float64
	UJoin          float64
	UNotch         float64
	UShoulder      float64
	UShoulderInner float64

	// The letter r: the left stem of n and an arm which leaves it like the arch
	// of n and ends in the ball terminal.  RNotch and RJoin are the heights at
	// which the outer edge and the underside of the arm meet the stem, RReach
	// the distance from the right-hand edge of the stem to the rightmost ink of
	// the ball, RFootRight the length of the right foot serif, RSideLeft the
	// space from the edge of the advance to the left-hand edge of the stem and
	// RSideRight the space from the ball to the edge of the advance.
	RNotch     float64
	RJoin      float64
	RReach     float64
	RFootRight float64
	RSideLeft  float64
	RSideRight float64

	// The letter f: an ascender stem with a foot serif on both sides (the right
	// one FFootRight long), a hook which grows out of the stem and ends in the
	// ball terminal, and a crossbar of thickness Hair at the
	// x-height.  The top of the hook stands FOvershoot above the ascender.  The
	// left edge of the stem becomes the outer edge of the hook where it leaves
	// the straight FOuterDrop below the top, and the right edge becomes its
	// underside FInnerDrop below the top.  FReach is the distance from the right
	// edge of the cell of the stem (SideStemLeft + Stem) to the rightmost ink
	// of the terminal.  The bar extends FBarLeft to the left and FBarRight to
	// the right of the stem.  FSideRight is the space from the rightmost ink to
	// the edge of the advance, negative where the hook overhangs it.
	FFootRight float64
	FOvershoot float64
	FOuterDrop float64
	FInnerDrop float64
	FReach     float64
	FBarLeft   float64
	FBarRight  float64
	FSideRight float64

	// The letter t: a stem of the thickness Stem which rises TRise above the
	// x-height, with an oblique top (see TObliqueDrop); a crossbar of
	// thickness Hair at the x-height which extends TBarLeft to the left and
	// TBarRight to the right of the stem, like the bar of f; and a
	// foot in which the stem curves to the right and ends in a tail.  TReach is
	// the distance from the right edge of the stem to the rightmost ink of the
	// tail, TTipHeight the height of the highest ink of its tip and TTipThick its
	// thickness there; the foot is ThinLower thick at its lowest point, which
	// lies OvershootLower below the baseline.  TSideRight is the space from the
	// tail to the edge of the advance.
	TRise      float64
	TBarLeft   float64
	TBarRight  float64
	TReach     float64
	TTipHeight float64
	TTipThick  float64
	TSideRight float64

	// The top of t is cut by a straight line which rises to the right: the cut
	// meets the left-hand edge of the stem TObliqueDrop below the top.  The cut
	// ends TObliqueFlat short of the right-hand edge of the stem, and a
	// horizontal flat of that length at TRise above the x-height joins it to
	// that edge.
	TObliqueDrop float64
	TObliqueFlat float64

	// The letter j: the stem, head serif and dot of i (the dot's parameters are
	// those of i), with the stem continuing below the baseline and curving to the
	// left into a tail which ends in the ball terminal, the hook of f turned
	// through half a turn.  JStemLeft is the position of the left edge of the
	// stem, JDepth the depth of the lowest ink, JOuterDrop and JInnerDrop the
	// distances from the lowest ink up to the heights at which the right and the
	// left edge of the stem begin to curve, JReach the distance from the left edge
	// of the stem to the leftmost ink of the terminal.
	JStemLeft  float64
	JDepth     float64
	JOuterDrop float64
	JInnerDrop float64
	JReach     float64

	// The letter g, the two-storey form (internal/glyphs/g.go): a round upper bowl
	// with an ear which ends in a ball, and one stroke which leaves the bowl at its
	// lower left as the link, sweeps to the right as the top of the lower loop,
	// runs round the loop and closes on the link.  The bowl is GBowlWidth wide and
	// GBowlHeight high with its top at the top of the o; its sides are GBowlThick
	// thick and its top and bottom ThinLower, and its left extreme lies GBowlInset
	// to the right of that of the loop.  The loop is GLoopWidth wide and its lowest
	// ink GLoopDepth below the baseline.  The top stroke of the loop is
	// GLoopTopThick thick, measured vertically, where its upper edge is GLoopFloor
	// above the baseline; the left and right ends of the loop are GLoopLeftThick
	// and GLoopRightThick thick and its bottom ThinLower.  The leftmost point of the
	// left edge of the link stands GLinkLeft to the right of the left extreme of
	// the loop and GLinkDrop below the bottom of the bowl; GLinkThick sets the
	// thickness of the link, the distance from its left edge to its right edge at
	// the bowl and at that leftmost point.  The ear: a ball of the diameter GEarBall
	// whose centre is GEarDrop below the x-height and whose right extreme lies
	// GEarReach beyond the right extreme of the bowl, on a stalk GEarThick thick
	// whose upper edge rises to GEarRise above the x-height.  GSideLeft is the space
	// from the edge of the advance to the left extreme of the loop and GSideRight
	// the space from the rightmost ink to the edge.
	GBowlWidth      float64
	GBowlHeight     float64
	GBowlThick      float64
	GBowlInset      float64
	GLoopWidth      float64
	GLoopDepth      float64
	GLoopFloor      float64
	GLoopTopThick   float64
	GLoopLeftThick  float64
	GLoopRightThick float64
	GLinkThick      float64
	GLinkLeft       float64
	GLinkDrop       float64
	GEarBall        float64
	GEarReach       float64
	GEarRise        float64
	GEarDrop        float64
	GEarThick       float64
	GSideLeft       float64
	GSideRight      float64

	// MaxOverhang is the greatest distance by which ink may lie beyond the edge
	// of the advance on a side which has a side kind of that name: the hook of f
	// on the right, the tail of j on the left.
	MaxOverhang float64

	// Diagonal strokes and the letters v, w and y
	// (internal/glyphs/diagonal.go).  The thick diagonal of v and y has the
	// thickness Stem, that of w WStemCut less; the thin diagonals have the
	// thickness Hair; both are measured at right angles to the stroke.  A
	// vertex is cut flat, DiagVertexFlat wide, DiagOvershoot below the
	// baseline.  The head serifs of the diagonals have a flat top on the
	// x-height and are SerifThick thick at the tip; they reach DiagOverThick
	// (the thick diagonal) or DiagOverThin (the thin one) beyond the edge on
	// each side, measured from the edge at the level of the underside of the
	// serif (see diagJutOuter).  VInk and WInk are the widths of the ink of v
	// and y and of w, from tip to tip of the serifs at the x-height; the angles
	// of the diagonals follow from them (VLean, WLean in diag.go).
	// DiagSideLeft and DiagSideRight are the distances from the edge of the
	// advance to the tips of the serifs.  The middle apex of w is cut flat,
	// WApexFlat wide, at the height WApexHeight.
	DiagOvershoot  float64
	DiagVertexFlat float64
	DiagOverThick  float64
	DiagOverThin   float64
	DiagSideLeft   float64
	DiagSideRight  float64
	VInk           float64
	WInk           float64
	WStemCut       float64
	WApexHeight    float64
	WApexFlat      float64

	// The tail of y is the tail of j (the hook of f turned round) on the thin
	// diagonal, turned further by the angle of the diagonal: YDepth, YReach,
	// YOuterDrop and YInnerDrop are the values of JDepth, JReach, JOuterDrop and
	// JInnerDrop in the frame of the tail before it is turned, in which the stroke is
	// vertical and YStrokeThick thick.
	YDepth     float64
	YReach     float64
	YOuterDrop float64
	YInnerDrop float64

	// The letter x (internal/glyphs/x.go): a thick diagonal of the thickness XThick
	// from the upper left to the lower right and a thin one of the thickness Hair from
	// the lower left to the upper right, at right angles to the stroke.  XCross is the
	// height at which the centre lines of the upper halves cross; XThinOffset is the
	// distance of the centre line of the lower half of the thin stroke to the right of
	// that of the upper half (negative: to the left), at the crossing.  XLeanDiff is
	// the lean of the thin diagonal less that of the thick one.  XInk is the width of
	// the ink at the baseline from tip to tip of the outer serifs; the leans follow
	// (XLeanSum).  XOverThick and XOverThin are the overhangs of the serifs of the
	// thick and of the thin diagonal beyond the edge on each side, measured at the
	// level of the face of the serif which carries the bracket; their thickness is
	// SerifThick.  The side spaces are DiagSideLeft and DiagSideRight.
	XThick      float64
	XInk        float64
	XCross      float64
	XLeanDiff   float64
	XThinOffset float64
	XOverThick  float64
	XOverThin   float64

	// The letter z (internal/glyphs/z.go): two bars of the thickness Hair joined by a
	// thick diagonal of the thickness ZDiag at right angles.  ZInk is the width of the
	// ink, from the corner of the diagonal at the baseline to the end of the lower bar,
	// which lies under the corner at the x-height; the lean of the diagonal follows
	// (ZLean).  ZInkTop is the width of the ink at the top, from the end of the upper
	// beak to that corner.  The beaks hang from the left end of the upper bar and rise
	// from the right end of the lower bar: height is measured from the outer edge of
	// the bar to the end of the beak, thick is the horizontal width halfway along it.
	ZDiag            float64
	ZInk             float64
	ZInkTop          float64
	ZSideLeft        float64
	ZSideRight       float64
	ZUpperBeakHeight float64
	ZUpperBeakThick  float64
	ZLowerBeakHeight float64
	ZLowerBeakThick  float64

	// The letter k (internal/glyphs/k.go): the stem of h with a thin arm of the
	// thickness Hair at the angle KArmAngle to the vertical and a thick leg of the
	// thickness KLeg at KLegAngle which leaves the arm.  The centre lines of the two
	// meet at the height KJunction.  KReach is the distance from the right-hand edge of
	// a stem of the width Stem (the cell of the stem, see AscenderInset) to the
	// rightmost ink, the end of the foot serif of the leg; the horizontal position of
	// the junction follows (KJunctionX).  The serifs of the arm and of the leg
	// overhang the edges by KOverArm and KOverLeg on each side, measured as for x.
	KLeg       float64
	KArmAngle  float64
	KLegAngle  float64
	KJunction  float64
	KReach     float64
	KSideRight float64
	KOverArm   float64
	KOverLeg   float64

	// Widths and fit.  The advance of a letter is not a parameter: it is
	// derived from the side spaces and the body of the letter, see fit.go.
	// SideStemLeft and SideStemRight are the distances from the edge of the
	// advance to the edge of a stem which carries serifs, and SideRoundLeft
	// and SideRoundRight the distances to the extreme of a bowl.  NCounter is
	// the distance between the inner edges of the stems of n, OBowlWidth the
	// width of the bowl of o.  NarrowSideCut is taken off both stem sides of
	// the narrow letters l and i, whose stem keeps the proportion of the stem sides of n.
	SideStemLeft   float64
	SideStemRight  float64
	SideRoundLeft  float64
	SideRoundRight float64
	DSideRight     float64 // distance from the right edge of the stem of d to the edge of the advance
	SideOpenRight  float64 // distance from the right extreme of the bowl of e to the edge of the advance
	BPSideLeft     float64 // distance from the edge of the advance to the left-hand edge of the stem of b and p
	DQSideLeft     float64 // distance from the edge of the advance to the left extreme of the bowl of d and q
	BPSideRight    float64 // distance from the right extreme of the bowl of b and p to the edge of the advance
	CSideLeft      float64 // distance from the edge of the advance to the left extreme of c
	CSideRight     float64 // distance from the rightmost ink of c, the tip of the lower terminal, to the edge of the advance
	QStemTop       float64 // height of the flat top of the stem of q: the x-height
	QSpur          float64 // height of the spur of q: how far the right-hand edge of the stem rises above QStemTop, before the top is cut obliquely down to the bowl
	QSideRight     float64 // distance from the right-hand edge of the stem of q to the edge of the advance
	NCounter       float64
	OBowlWidth     float64
	NarrowSideCut  float64
	SpaceAdvance   float64

	// Caps holds the parameters of the capitals, see caps.go.
	Caps Caps

	// CapDiag holds the parameters of the capitals built from diagonal strokes, see
	// capdiag.go.
	CapDiag CapDiag

	// Figures holds the parameters of the figures 0 to 9, see figures.go.
	Figures Figures
	// Punct holds the parameters of the basic punctuation and the signs of ASCII
	// mathematics, see punct.go.
	Punct Punct
	// Sym holds the parameters of the question mark, the ampersand and the
	// symbols of the basic set, see symbols.go.
	Sym Symbols

	// Kern holds the parameters of the pair kerning, see kern.go.
	Kern Kern
}

// Regular returns the parameters of Bedern Regular.
func Regular() Params {
	return Params{
		UnitsPerEm: 1000,

		XHeight:        448,
		CapHeight:      683,
		Ascender:       694,
		Descender:      194,
		OvershootLower: 12,
		OvershootUpper: 16,

		TypoAscender:  806,
		TypoDescender: -194,
		LineGap:       200,

		Stem:         81,
		AscenderStem: 81,
		CurveLower:   94,
		ThinLower:    32,
		Hair:         36,

		SerifJut:      69,
		SerifThick:    33,
		BracketHeight: 61,
		BracketReach:  45,

		SerifJutInner: 64,
		HeadSerifDrop: 12,

		ArchNotch:          372,
		ArchJoin:           322,
		ArchThinJoin:       28,
		ArchPeak:           0.65,
		ArchOvershoot:      11,
		ShoulderStart:      300,
		ShoulderStartInner: 315,

		SupernessOuter: 0.72,
		SupernessInner: 0.745,

		DotDiameter:  116,
		DotCentre:    621,
		DotShift:     0.085,
		DotSuperness: 0.7071,

		BowlJoinThick: 32,
		DBowl: Bowl{
			Width:     338,
			JoinUpper: 358, JoinLower: 100,
			NotchUpper: 413, NotchLower: 50,
		},

		BBowl: Bowl{
			Width:     332,
			JoinUpper: 358, JoinLower: 100,
			NotchUpper: 413, NotchLower: 50,
		},
		PBowl: Bowl{
			Width:     332,
			JoinUpper: 358, JoinLower: 92,
			NotchUpper: 413, NotchLower: 40,
		},

		QBowl: Bowl{
			Width:     338,
			JoinUpper: 358, JoinLower: 100,
			NotchUpper: 413, NotchLower: 50,
		},

		EBarHeight:    240,
		EBarThick:     32,
		EBowlWidth:    400,
		EThickPos:     0.48,
		EEyeSuperness: 0.76,
		EAperture:     114,
		ETerminalTip:  34,

		CBowlWidth: 388,
		CTipHeight: 98,
		CBallInset: 24,

		ABowl: Bowl{
			Width: 273, Run: ABowlRun, Top: 273, Thick: 86, ThickPos: 0.36,
			JoinUpper: 236, JoinLower: 120,
			NotchUpper: 270, NotchLower: 52,
		},
		AArcReach:    20,
		BallDiameter: 98,
		BallStalk:    34,
		ASideLeft:    33,

		ATailReach:     100,
		ATailTipHeight: 85,
		ATailTipThick:  36,
		ATailSideRight: 0,

		SSpineThick:    84,
		SSpineAngle:    20,
		SArcThick:      36,
		SUpperArcThick: 39,
		SSideThick:     74,
		SLowerWidth:    320,
		SUpperWidth:    293,
		SUpperShift:    12,
		SUpperDrop:     156,
		SLowerDrop:     172,
		SBeakDrop:      92,
		SBeakLift:      11,
		SBeakRun:       60,
		SBeakArc:       40,
		SSerifThick:    40,
		SSideLeft:      37,
		SSideRight:     29,

		MCounter: 189,

		UCounter:       189,
		UJoin:          126,
		UNotch:         72,
		UShoulder:      148,
		UShoulderInner: 133,

		RNotch:     372,
		RJoin:      322,
		RReach:     206,
		RFootRight: 80,
		RSideLeft:  94,
		RSideRight: 7,

		FFootRight: 80,
		FOvershoot: 10,
		FOuterDrop: 120,
		FInnerDrop: 100,
		FReach:     160,
		FBarLeft:   75,
		FBarRight:  112,
		FSideRight: 3,

		TRise:      178,
		TBarLeft:   88,
		TBarRight:  140,
		TReach:     172,
		TTipHeight: 110,
		TTipThick:  36,
		TSideRight: 3,

		TObliqueDrop: 90,
		TObliqueFlat: 18,

		JStemLeft:  125,
		JDepth:     200,
		JOuterDrop: 120,
		JInnerDrop: 115,
		JReach:     184,

		GBowlWidth:      322,
		GBowlHeight:     312,
		GBowlThick:      80,
		GBowlInset:      32,
		GLoopWidth:      440,
		GLoopDepth:      200,
		GLoopFloor:      60,
		GLoopTopThick:   54,
		GLoopLeftThick:  68,
		GLoopRightThick: 74,
		GLinkThick:      42,
		GLinkLeft:       46,
		GLinkDrop:       40,
		GEarBall:        72,
		GEarReach:       76,
		GEarRise:        20,
		GEarDrop:        42,
		GEarThick:       26,
		GSideLeft:       20,
		GSideRight:      18,

		MaxOverhang: 60,

		DiagOvershoot:  12,
		DiagVertexFlat: 24,
		DiagOverThick:  52,
		DiagOverThin:   64,
		DiagSideLeft:   17,
		DiagSideRight:  11,
		VInk:           480,
		WInk:           684,
		WStemCut:       5,
		WApexHeight:    444,
		WApexFlat:      30,

		YDepth:     200,
		YReach:     145,
		YOuterDrop: 100,
		YInnerDrop: 90,

		XThick:      74,
		XInk:        472,
		XCross:      232,
		XLeanDiff:   0.027,
		XThinOffset: -5,
		XOverThick:  52,
		XOverThin:   64,

		ZDiag:            79,
		ZInk:             396,
		ZInkTop:          374,
		ZSideLeft:        12,
		ZSideRight:       18,
		ZUpperBeakHeight: 118,
		ZUpperBeakThick:  36,
		ZLowerBeakHeight: 115,
		ZLowerBeakThick:  40,

		KLeg:       76,
		KArmAngle:  47,
		KLegAngle:  35,
		KJunction:  269,
		KReach:     319,
		KSideRight: 10,
		KOverArm:   60,
		KOverLeg:   50,

		SideStemLeft:   99,
		SideStemRight:  90,
		SideRoundLeft:  26,
		SideRoundRight: 27,
		DSideRight:     90,
		BPSideLeft:     95,
		DQSideLeft:     33,
		BPSideRight:    38,
		CSideLeft:      32,
		CSideRight:     9,
		QSideRight:     71,
		QStemTop:       448,
		QSpur:          10,
		SideOpenRight:  25,
		NCounter:       189,
		OBowlWidth:     433,
		NarrowSideCut:  0,
		SpaceAdvance:   323,

		Caps: regularCaps(),

		CapDiag: regularCapDiag(),

		Figures: regularFigures(),
		Punct:   regularPunct(),
		Sym:     regularSymbols(),

		Kern: regularKern(),
	}
}

// SlitRatio is the ratio of arc length to distance above which the structural
// check treats two parts of one contour as not neighbours, see
// internal/glyphs/check.go.  Two edges which meet at a notch at an angle of less
// than the slit angle, about 23 degrees, are further apart than this ratio
// allows, over their whole length, and are counted as a slit.
const SlitRatio = 5.0

// ABowlRun is the distance from the extreme at the top of the bowl of a to its
// stem; the extreme at the bottom lies further from it (internal/glyphs/a.go).
const ABowlRun = 50.0

// The constants of the arch of n and h:
// the length of the handle of the outer edge and of the underside at the left
// stem, as fractions of the horizontal distance to the peak of the edge, the
// distance by which the underside peaks left of the outer edge, and the angles
// in degrees by which the outer edge and the underside leave the stem more
// steeply than the angle which follows from ArchThinJoin.
const (
	ArchLeftHandle      = 0.34
	ArchUnderLeftHandle = 0.40
	ArchUnderPeakShift  = 16.0
	ArchOuterSteeper    = 9.0
	ArchUnderSteeper    = 10.0
	// ArchMaxLeave is the steepest angle, in degrees from the horizontal, at which
	// an edge of an arch leaves a stem: the notch between the outer edge and the
	// stem is then 24 degrees wide (see SlitRatio).
	ArchMaxLeave = 66.0
)

// ClearanceFloor returns the smallest distance allowed between a counter and
// the outer contour and between the two sides of a stroke.  It is 20 units,
// and 0.7 ThinLower if the thin strokes of the
// parameter set are thinner than 28.6.  The structural check demands it.
func ClearanceFloor(p Params) float64 {
	return math.Min(20, 0.7*p.ThinLower)
}

// Validate returns an error for a parameter set with a value which is not
// finite, a size which is not positive, a thin part which is not thinner than
// the stem, or heights which are out of order.  Whether every letter can be
// drawn from the set is checked by glyphs.Check, which builds every glyph and
// runs the structural checks on it.
func (p Params) Validate() error {
	if err := p.validateFinite(); err != nil {
		return err
	}
	for _, c := range []struct {
		name string
		v    float64
	}{
		{"UnitsPerEm", p.UnitsPerEm}, {"XHeight", p.XHeight}, {"CapHeight", p.CapHeight},
		{"Ascender", p.Ascender}, {"Descender", p.Descender}, {"Stem", p.Stem},
		{"AscenderStem", p.AscenderStem}, {"CurveLower", p.CurveLower},
		{"ThinLower", p.ThinLower}, {"Hair", p.Hair}, {"SerifJut", p.SerifJut},
		{"SerifThick", p.SerifThick}, {"NCounter", p.NCounter},
		{"BallDiameter", p.BallDiameter}, {"BallStalk", p.BallStalk}, {"DotDiameter", p.DotDiameter},
	} {
		if !(c.v > 0) {
			return fmt.Errorf("design: %s = %v, want a positive length", c.name, c.v)
		}
	}
	for _, c := range []struct {
		name, bigName string
		v, big        float64
	}{
		{"ThinLower", "CurveLower", p.ThinLower, p.CurveLower},
		{"Hair", "Stem", p.Hair, p.Stem},
		{"ArchThinJoin", "Stem", p.ArchThinJoin, p.Stem},
		{"SerifThick", "Stem", p.SerifThick, p.Stem},
		{"ArchJoin", "ArchNotch", p.ArchJoin, p.ArchNotch},
		{"ArchNotch", "XHeight", p.ArchNotch, p.XHeight},
		{"XHeight", "CapHeight", p.XHeight, p.CapHeight},
		{"XHeight", "Ascender", p.XHeight, p.Ascender},
		{"ArchPeak", "1", p.ArchPeak, 1},
	} {
		if !(c.v < c.big) {
			return fmt.Errorf("design: %s = %v is not below %s = %v", c.name, c.v, c.bigName, c.big)
		}
	}
	// AscenderStem may be smaller than Stem but not larger.
	if p.AscenderStem > p.Stem {
		return fmt.Errorf("design: AscenderStem = %v exceeds Stem = %v", p.AscenderStem, p.Stem)
	}
	return nil
}

// validateFinite returns an error if a float64 field is NaN or infinite.
func (p Params) validateFinite() error {
	v := reflect.ValueOf(p)
	for i := 0; i < v.NumField(); i++ {
		if f := v.Field(i); f.Kind() == reflect.Float64 {
			if x := f.Float(); math.IsNaN(x) || math.IsInf(x, 0) {
				return fmt.Errorf("design: %s = %v, want a finite number", v.Type().Field(i).Name, x)
			}
		}
	}
	return nil
}

// Describe returns the parameters as plain text, one `Name value` line for
// each field, with the fields of a Bowl as `Name.Field`.
func (p Params) Describe() string {
	var out string
	var walk func(prefix string, v reflect.Value)
	walk = func(prefix string, v reflect.Value) {
		for i := 0; i < v.NumField(); i++ {
			name := prefix + v.Type().Field(i).Name
			if f := v.Field(i); f.Kind() == reflect.Struct {
				walk(name+".", f)
			} else {
				out += fmt.Sprintf("%s %v\n", name, f.Interface())
			}
		}
	}
	walk("", reflect.ValueOf(p))
	return out
}
