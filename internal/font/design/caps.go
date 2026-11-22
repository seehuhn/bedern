// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the capitals and what follows from them:
// the positions of the stems and the advances.  The capitals share the cap
// height (CapHeight) and the overshoot of the round capitals (OvershootUpper)
// with the rest of the font and have all their other values here.  The
// construction is described in internal/glyphs/capstem.go, caparm.go and the
// files of the letters.

// Caps holds the parameters of the capitals.  All horizontal positions which
// follow from them are in the functions below.
type Caps struct {
	// Strokes.  Stem is the stem of H, I, E, F, L and T.  Round is the
	// thickest part of O, at its left and right extremes, and RoundThin its
	// thinnest part, at the top and the bottom.  Bar is the thickness of the
	// crossbar of H and of the arms of E, F and L; TBar that of the bar of T.
	Stem      float64
	Round     float64
	RoundThin float64
	Bar       float64
	TBar      float64

	// Serifs of the stems.  SerifJut is the length of a serif beyond the stem,
	// the same on both sides and at head and foot; SerifThick the thickness
	// at the tip; BracketHeight the distance from the baseline (from the cap
	// height, at the head) at which the bracket leaves the stem; BracketReach
	// the distance from the stem at which it meets the flat of the serif.  The
	// head serif is the foot serif turned through half a turn.
	SerifJut      float64
	SerifThick    float64
	BracketHeight float64
	BracketReach  float64

	// The vertical serifs at the ends of the arms of E, F, L and T, which are
	// the beaks of z at the size of the capitals.  TopDrop is how far the serif
	// of a top arm (E, F) reaches below the underside of the arm, BottomRise
	// how far that of a bottom arm (E, L) reaches above its upper edge,
	// MiddleReach how far that of a middle arm (E, F) reaches above and below
	// it, and TDrop how far those of T reach below its bar.  ArmSerifThick is
	// the width of a serif halfway along its height.
	TopDrop       float64
	BottomRise    float64
	MiddleReach   float64
	TDrop         float64
	ArmSerifThick float64

	// Heights.  HBarHeight is the height of the centre of the bar of H,
	// EMiddleHeight and FMiddleHeight those of the middle arms of E and F.
	HBarHeight    float64
	EMiddleHeight float64
	FMiddleHeight float64

	// Widths.  The lengths of the arms are measured from the right-hand edge of
	// the stem to the vertical cut at the end of the arm, which is the end of
	// the ink; TArm is the length of each arm of T from the edge of its stem.
	// HCounter is the distance between the inner edges of the stems of H and
	// OWidth the width of O from its left to its right extreme.
	HCounter   float64
	OWidth     float64
	ETopArm    float64
	EMiddleArm float64
	EBottomArm float64
	FTopArm    float64
	FMiddleArm float64
	LArm       float64
	TArm       float64

	// Side spaces.  SideStemLeft and SideStemRight are the distances from the
	// edge of the advance to the edge of a capital stem with serifs,
	// SideRoundLeft and SideRoundRight those to the extremes of O.  SideArmE,
	// SideArmF and SideArmL are the distances from the end of the longest arm
	// of E, F and L to the right edge of the advance, and SideT the distance
	// from the end of each arm of T to the edge of the advance on its side.
	SideStemLeft   float64
	SideStemRight  float64
	SideRoundLeft  float64
	SideRoundRight float64
	SideArmE       float64
	SideArmF       float64
	SideArmL       float64
	SideT          float64

	// The letters C, G, S, J and Q; see internal/design/capcgsjq.go.
	CGSJQ CGSJQ
	// The bowl letters B, D, P and R and the letter U, see capsbowl.go.
	CapBowls
}

// regularCaps returns the parameters of the capitals of Bedern Regular.
func regularCaps() Caps {
	return Caps{
		Stem:      96,
		Round:     116,
		RoundThin: 38,
		Bar:       38,
		TBar:      48,

		SerifJut:      92,
		SerifThick:    33,
		BracketHeight: 64,
		BracketReach:  50,

		TopDrop:       115,
		BottomRise:    135,
		MiddleReach:   80,
		TDrop:         105,
		ArmSerifThick: 40,

		HBarHeight:    355,
		EMiddleHeight: 355,
		FMiddleHeight: 345,

		HCounter:   287,
		OWidth:     646,
		ETopArm:    390,
		EMiddleArm: 240,
		EBottomArm: 405,
		FTopArm:    380,
		FMiddleArm: 235,
		LArm:       345,
		TArm:       272,

		SideStemLeft:   121,
		SideStemRight:  120,
		SideRoundLeft:  50,
		SideRoundRight: 50,
		SideArmE:       19,
		SideArmF:       11,
		SideArmL:       15,
		SideT:          13,

		CGSJQ:    regularCGSJQ(),
		CapBowls: regularCapBowls(),
	}
}

// CapStemLeft returns the position of the left-hand edge of the stem of H, I,
// E, F and L: the capital stem side.
func CapStemLeft(p Params) float64 { return p.Caps.SideStemLeft }

// CapHAdvance returns the advance of H: the stem sides, two stems and the counter.
func CapHAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + 2*c.Stem + c.HCounter + c.SideStemRight
}

// CapOAdvance returns the advance of O: the round sides and the width of O.
func CapOAdvance(p Params) float64 {
	c := p.Caps
	return c.SideRoundLeft + c.OWidth + c.SideRoundRight
}

// CapIAdvance returns the advance of I: the stem between the stem sides.
func CapIAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + c.Stem + c.SideStemRight
}

// CapEAdvance returns the advance of E: the left stem side, the stem, the
// bottom arm, which is the longest, and the space after it.
func CapEAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + c.Stem + c.EBottomArm + c.SideArmE
}

// CapFAdvance returns the advance of F: the left stem side, the stem, the top
// arm and the space after it.
func CapFAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + c.Stem + c.FTopArm + c.SideArmF
}

// CapLAdvance returns the advance of L: the left stem side, the stem, the arm and
// the space after it.
func CapLAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + c.Stem + c.LArm + c.SideArmL
}

// CapTStemLeft returns the position of the left-hand edge of the stem of T.
func CapTStemLeft(p Params) float64 { return p.Caps.SideT + p.Caps.TArm }

// CapTAdvance returns the advance of T: the two arms on either side of the
// stem, each with its side space.
func CapTAdvance(p Params) float64 {
	c := p.Caps
	return 2*(c.SideT+c.TArm) + c.Stem
}
