// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the capitals C, G, S, J and Q and the
// advances which follow from them.  C, G and Q are cut from the round of O
// (OWidth, Round, RoundThin and the squareness of o), and S is drawn as the
// lowercase s is drawn, with the values of the capitals.  The construction is
// described in internal/glyphs/capbeak.go and the files of the letters.

// CGSJQ holds the parameters of C, G, S, J and Q.
type CGSJQ struct {
	// The beaks of C and G.  BeakX is the distance from the left extreme of the
	// round to the vertical cut of the upper beak, which is also the cut of the
	// lower serif of C and the right-hand edge of the stem of G.  The upper arc
	// ends in a corner on the cut BeakDrop below the top of the round, where it
	// is BeakArc thick at right angles to its edge, and carries a serif which
	// hangs down to BeakEnd.  The beaks of S are those of C and G, with
	// BeakDrop and BeakArc.
	BeakX    float64
	BeakDrop float64
	BeakArc  float64
	BeakEnd  float64

	// The serif of the upper beak of C ends at CBeakEnd, higher than
	// that of G (BeakEnd); the cut BeakX, the corner
	// BeakDrop and the arc BeakArc are shared.
	CBeakEnd float64

	// The tip of the serif of the upper beak stands BeakLift above the place
	// where the round (BeakDrop below the top) meets the cut; the outer edge
	// leaves the round BeakRun (of arc length) before the cut.  S has the same
	// values.
	BeakLift float64
	BeakRun  float64

	// The tail of C, the lower end without a serif.  The outer corner of its
	// cut lies CTailReach right of the vertical of the upper cut and
	// CTailHeight above the baseline; the cut is CTailCut long, at right
	// angles to the stroke, and the tail is CTailArc thick where it is
	// thickest.
	CTailHeight float64
	CTailReach  float64
	CTailCut    float64
	CTailArc    float64

	// G.  GStemTop is the height of the flat top of the short stem at the lower
	// right, which carries a serif on each side: GSerifLeft into the counter and
	// GSerifRight beyond the stem.  GJoin is the height above the baseline at
	// which the outer edge of the lower arc of the round runs into the
	// right-hand edge of the stem.
	GStemTop    float64
	GSerifLeft  float64
	GSerifRight float64
	GJoin       float64

	// S, in the meaning of the parameters of s.
	SLowerWidth, SUpperWidth  float64
	SUpperShift               float64
	SSpineThick, SSpineAngle  float64
	SArcThick, SUpperArcThick float64
	SSideThick                float64
	SUpperDrop, SLowerDrop    float64

	// J.  JReach is the distance from the leftmost ink, the ball, to the
	// left-hand edge of the stem; JOuterDrop and JInnerDrop are the heights
	// above the lowest point at which the right-hand and the left-hand edge of
	// the stem begin to curve; JHeadRight is the length of the head serif on
	// the right (the left one is SerifJut); JBall is the diameter of the ball.
	JReach     float64
	JOuterDrop float64
	JInnerDrop float64
	JHeadRight float64
	JBall      float64

	// Q.  The tail leaves the bowl at the bottom QTailStart from the left
	// extreme of the round and meets it again on the lower right QTailJoin
	// from it; it ends QTailDepth below the baseline QTailEnd from the left
	// extreme, in a cut QTailTip long.
	QTailStart float64
	QTailJoin  float64
	QTailEnd   float64
	QTailDepth float64
	QTailTip   float64

	// Side spaces, to the rightmost ink of C (the tip of the tail), of G (the
	// tip of the right serif), to the leftmost and rightmost ink of S and to
	// the ball of J.
	CSideRight float64
	GSideRight float64
	SSideLeft  float64
	SSideRight float64
	JSideLeft  float64
}

// regularCGSJQ returns the parameters of C, G, S, J and Q of Bedern Regular.
func regularCGSJQ() CGSJQ {
	return CGSJQ{
		BeakX:    590,
		BeakDrop: 135,
		BeakLift: 10,
		BeakRun:  90,
		BeakArc:  44,
		BeakEnd:  420,
		CBeakEnd: 444,

		CTailHeight: 164,
		CTailReach:  4,
		CTailCut:    32,
		CTailArc:    50,

		GStemTop:    295,
		GSerifLeft:  92,
		GSerifRight: 66,
		GJoin:       200,

		SLowerWidth:    430,
		SUpperWidth:    402,
		SUpperShift:    12,
		SSpineThick:    100,
		SSpineAngle:    20,
		SArcThick:      40,
		SUpperArcThick: 43,
		SSideThick:     86,
		SUpperDrop:     240,
		SLowerDrop:     255,

		JReach:     250,
		JOuterDrop: 255,
		JInnerDrop: 225,
		JHeadRight: 60,
		JBall:      110,

		QTailStart: 368,
		QTailJoin:  445,
		QTailEnd:   590,
		QTailDepth: 150,
		QTailTip:   20,

		CSideRight: 28,
		GSideRight: 21,
		SSideLeft:  56,
		SSideRight: 47,
		JSideLeft:  10,
	}
}

// CapCAdvance returns the advance of C: the round side, the width to the cut of
// the beak, the reach of the tail beyond it and the space after the tail.
func CapCAdvance(p Params) float64 {
	c := p.Caps
	return c.SideRoundLeft + c.CGSJQ.BeakX + c.CGSJQ.CTailReach + c.CGSJQ.CSideRight
}

// CapGAdvance returns the advance of G: the round side, the width to the
// right-hand edge of the stem, the right serif and the space after it.
func CapGAdvance(p Params) float64 {
	c := p.Caps
	return c.SideRoundLeft + c.CGSJQ.BeakX + c.CGSJQ.GSerifRight + c.CGSJQ.GSideRight
}

// CapSAdvance returns the advance of S: the lower width between the side
// spaces.
func CapSAdvance(p Params) float64 {
	g := p.Caps.CGSJQ
	return g.SSideLeft + g.SLowerWidth + g.SSideRight
}

// CapJStemLeft returns the position of the left-hand edge of the stem of J.
func CapJStemLeft(p Params) float64 {
	g := p.Caps.CGSJQ
	return g.JSideLeft + g.JReach
}

// CapJAdvance returns the advance of J: the ball side, the reach, the stem, the
// head serif on the right and the space from its tip to the edge, which is that
// of H (SideStemRight less SerifJut).
func CapJAdvance(p Params) float64 {
	c := p.Caps
	return CapJStemLeft(p) + c.Stem + c.CGSJQ.JHeadRight + c.SideStemRight - c.SerifJut
}

// CapQAdvance returns the advance of Q, that of O: the tail ends inside the
// right extreme of the bowl.
func CapQAdvance(p Params) float64 { return CapOAdvance(p) }
