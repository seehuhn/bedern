// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the question mark, the ampersand and the
// symbols of the basic set: at sign, number sign, percent, dollar, sterling,
// euro, section sign, braces, the ASCII tilde, circumflex and grave, and the
// degree.  The construction is described in the files of the glyphs in
// internal/glyphs (question.go, ampersand.go, at.go, numbersign.go, percent.go,
// currency.go, section.go, braces.go, asciiaccents.go, degree.go).

// Symbols holds the parameters of the symbols.
type Symbols struct {
	// The question mark.  QTop is the top of the hook, QInk the ink width,
	// QGap the white between the hook and the dot, QStem the thickness of the
	// stem at its foot, QSide the side spaces.
	QTop  float64
	QInk  float64
	QGap  float64
	QStem float64
	QSide float64

	// The degree: a ring of the outer diameter DegDiam and the thickness
	// DegThick with its top at DegTop, DegSideLeft to its left (against a figure
	// before it) and DegSideRight to its right (against a C or F after it).
	DegDiam, DegThick, DegTop, DegSideLeft, DegSideRight float64

	// The ASCII tilde: TildeWidth wide, TildeHeight high with its centre on the
	// maths axis, TildeThick thick at the middle and TildeThin at the crests.
	TildeWidth, TildeHeight, TildeThick, TildeThin, TildeSide float64

	// The ASCII circumflex: CircWidth wide and CircHeight high with its apex at
	// CircTop, strokes CircThick thick, CircSide on each side.
	CircWidth, CircHeight, CircTop, CircThick, CircSide float64

	// The ASCII grave: a wedge from a round head GraveHead across to a round
	// tip GraveTip across, GraveWidth wide and GraveHeight high with its top
	// at GraveTop, and the side spaces GraveSideLeft and GraveSideRight.
	GraveHead, GraveTip, GraveWidth, GraveHeight, GraveTop float64
	GraveSideLeft, GraveSideRight                          float64

	// The braces: BraceInk wide from the bottom to the top of the parentheses
	// (BracketBottom and BracketTop), BraceArm thick on the straight arms,
	// BraceEnd thick at the ends, BraceTip across at the point, which leans
	// BraceTipAngle degrees from the horizontal; BraceSide on each side.
	BraceInk, BraceArm, BraceEnd, BraceTip, BraceTipAngle, BraceSide float64

	// The percent: PctInk wide, two small zeros PctZeroW wide and PctZeroH
	// high (the upper with its top, the lower with its bottom at the overshoot
	// of the figures), PctZeroThick thick at the sides and PctZeroThin at the
	// top and bottom, and a diagonal PctSlash thick from the baseline to
	// the figure height; PctSide on each side.
	PctInk, PctZeroW, PctZeroH, PctZeroThick, PctZeroThin, PctSlash, PctSide float64

	// The number sign: two strokes NumStroke thick leaning NumLean degrees
	// from the vertical, NumPitch apart at the middle, and two bars NumBar
	// thick, NumGap apart between their inner edges, centred on NumCentre;
	// NumInk wide and NumHeight high, NumSide on each side.
	NumStroke, NumLean, NumPitch, NumBar, NumGap, NumCentre, NumInk, NumHeight, NumSide float64

	// The dollar: an S at the figure height, drawn as the capital S is drawn,
	// with the values in the meaning of the parameters of s; and a vertical
	// stroke DollarStroke thick from DollarBottom to DollarTop through it.
	DollarLowerWidth, DollarUpperWidth, DollarUpperShift         float64
	DollarSpine, DollarSpineAngle, DollarArc, DollarUpperArc     float64
	DollarSideThick, DollarUpperDrop, DollarLowerDrop, DollarTip float64
	DollarStroke, DollarTop, DollarBottom                        float64

	// The sterling: SterInk wide; a stem SterStem thick which rises from the
	// base into a hook over the top that ends in the ball terminal at the
	// upper right; a bar SterBar thick at SterBarY, from the left of the ink
	// SterBarLen long; a base SterBase thick on the baseline.
	SterInk, SterStem, SterBar, SterBarY, SterBarLen, SterBase float64

	// The euro: EuroInk wide, of which the bars reach EuroReach left of the
	// round.  The round is the capital C (the round of O, the upper beak and
	// the lower terminal) at the figure height, EuroRoundW wide with the cut of
	// the beak EuroBeakX from its left extreme and the end of the beak at
	// EuroBeakEnd, EuroThick at the left and the thin part of the capitals at
	// the top and the bottom; two bars EuroBar thick, EuroBarGap apart between
	// their centres about the middle of the figure height, reach EuroBarEnd
	// from the left of the ink.
	EuroInk, EuroReach, EuroRoundW, EuroBeakX, EuroBeakEnd, EuroThick float64
	EuroBar, EuroBarGap, EuroBarEnd                                   float64

	// The section sign: from SectBottom to SectTop; two strokes whose spines
	// are SectSpine thick across their length, with a counter between them
	// (the eye) SectEyeLong by SectEyeShort along and across the spines; the
	// side of each hook SectHookThick, its top and the arcs which close the
	// eye Figures.Thin; SectSide on each side of the ink.
	SectTop, SectBottom, SectSpine, SectEyeLong, SectEyeShort, SectHookThick, SectSide float64

	// The at sign: AtInk wide from AtBottom to AtTop; the ring AtRing thick;
	// the bowl of the inner a AtBowlW wide from AtBowlBottom to AtBowlTop,
	// AtBowlThick at its left and Figures.Thin at its top and bottom, and its stem
	// AtStem thick; AtSide on each side.
	AtInk, AtTop, AtBottom, AtRing, AtBowlW, AtBowlBottom, AtBowlTop, AtBowlThick, AtStem, AtSide float64

	// The ampersand: AmpInk wide from AmpBottom to AmpTop; the thick stroke
	// (the left of the upper loop, the diagonal and the tail) AmpThick, the
	// left side of the lower bowl AmpBowlThick, the thin strokes Figures.Thin; the
	// arm, AmpArmThick at its foot, ends at AmpArm in the head serif of a
	// diagonal of the capitals, which reaches AmpSerifLeft and AmpSerifRight
	// beyond its edges; AmpSide on each side.
	AmpInk, AmpTop, AmpBottom, AmpThick, AmpBowlThick, AmpArm, AmpArmThick, AmpSerifLeft, AmpSerifRight, AmpSide float64
}

// regularSymbols returns the parameters of the symbols of Bedern Regular.
func regularSymbols() Symbols {
	return Symbols{
		QTop:  690,
		QInk:  370,
		QGap:  72,
		QStem: 50,
		QSide: 46,

		DegDiam: 287, DegThick: 42, DegTop: 699, DegSideLeft: 22, DegSideRight: 12,

		TildeWidth: 540, TildeHeight: 132, TildeThick: 60, TildeThin: 50, TildeSide: 20,

		CircWidth: 520, CircHeight: 220, CircTop: 726, CircThick: 54, CircSide: 30,

		GraveHead: 70, GraveTip: 26, GraveWidth: 190, GraveHeight: 171, GraveTop: 698,
		GraveSideLeft: 96, GraveSideRight: 104,

		BraceInk: 280, BraceArm: 73, BraceEnd: 32, BraceTip: 44, BraceTipAngle: 32, BraceSide: 55,

		PctInk: 680, PctZeroW: 280, PctZeroH: 370, PctZeroThick: 68, PctZeroThin: 32, PctSlash: 46, PctSide: 50,

		NumStroke: 50, NumLean: 12, NumPitch: 230, NumBar: 54, NumGap: 150, NumCentre: 330, NumInk: 550, NumHeight: 690, NumSide: 35,

		DollarLowerWidth: 394, DollarUpperWidth: 368, DollarUpperShift: 10,
		DollarSpine: 80, DollarSpineAngle: 20, DollarArc: 37, DollarUpperArc: 40,
		DollarSideThick: 70, DollarUpperDrop: 215, DollarLowerDrop: 245, DollarTip: 38,
		DollarStroke: 42, DollarTop: 748, DollarBottom: -90,

		SterInk: 440, SterStem: 86, SterBar: 40, SterBarY: 300, SterBarLen: 300, SterBase: 50,

		EuroInk: 458, EuroReach: 56, EuroRoundW: 425, EuroBeakX: 402, EuroBeakEnd: 405, EuroThick: 96,
		EuroBar: 38, EuroBarGap: 110, EuroBarEnd: 300,

		SectTop: 700, SectBottom: -160, SectSpine: 76, SectEyeLong: 300, SectEyeShort: 148, SectHookThick: 66, SectSide: 48,

		AtInk: 660, AtTop: 700, AtBottom: -20, AtRing: 40, AtBowlW: 300, AtBowlBottom: 130, AtBowlTop: 500,
		AtBowlThick: 84, AtStem: 76, AtSide: 48,

		AmpInk: 670, AmpTop: 697, AmpBottom: -14, AmpThick: 88, AmpBowlThick: 88,
		AmpArm: 450, AmpArmThick: 50, AmpSerifLeft: 30, AmpSerifRight: 40, AmpSide: 40,
	}
}
