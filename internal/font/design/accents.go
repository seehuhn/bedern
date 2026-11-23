// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the accent marks and of the letters
// composed from a base letter and a mark.  The construction is in
// internal/font/glyphs (marks.go, compose.go).
//
// A mark has a form for the lowercase and, where it is given, a flatter form
// for the capitals and for the letters with ascenders, which stands above the
// cap height or the ascender height.  Heights are measured from the x-height
// (the lowercase forms) or from the cap height or the ascender height (the
// capital forms) to the lowest ink of the mark; widths and heights are those
// of the ink.

// Accents holds the parameters of the accent marks.
type Accents struct {
	// Gap is the height of the lowest ink of a lowercase mark above the
	// x-height, GapCap that of a capital mark above the cap height and above
	// the ascender height.
	Gap, GapCap float64

	// The acute and the grave: a wedge from a round head AcuteHead across,
	// the heavy end at the top, to a round tip AcuteTip across, AcuteWidth
	// wide and AcuteHeight high (AcuteHeightCap on the capitals).
	// AcuteShift moves the acute to the right of the anchor and the grave to
	// the left.
	AcuteHead, AcuteTip, AcuteWidth, AcuteHeight, AcuteHeightCap, AcuteShift float64

	// The circumflex and the caron: two straight strokes, CircThin thick on
	// the left of the circumflex and the right of the caron and CircThick on
	// the other side, CircWidth wide and CircHeight high (CircHeightCap on the
	// capitals), meeting in a point.
	CircThin, CircThick, CircWidth, CircHeight, CircHeightCap float64

	// The tilde: TildeWidth wide and TildeHeight high (TildeHeightCap on the
	// capitals), TildeThick in the middle and TildeThin at the crest and the
	// trough.
	TildeWidth, TildeHeight, TildeHeightCap, TildeThick, TildeThin float64

	// The diaeresis and the dot above: dots DotScale times the dot of i
	// across; the two dots of the diaeresis have their centres DiaPitch apart,
	// DiaPitchNarrow on the dotless i.  The dot above is DotAboveScale times
	// the dot of i across.
	DotScale, DiaPitch, DiaPitchNarrow, DotAboveScale float64

	// The macron: a bar MacronWidth long (MacronWidthCap on the capitals) and
	// MacronThick thick.
	// MacronLift raises the macron above the other marks.
	MacronWidth, MacronWidthCap, MacronThick, MacronLift float64

	// The breve: a cup BreveWidth wide and BreveHeight high (BreveHeightCap
	// on the capitals), BreveThick thick at the bottom and BreveThin at the
	// ends.
	BreveWidth, BreveHeight, BreveHeightCap, BreveThick, BreveThin float64

	// The ring: RingDiam across outside (RingDiamCap on the capitals),
	// RingThick thick at the sides and RingThin at the top and the bottom.
	RingDiam, RingDiamCap, RingThick, RingThin float64

	// The double acute: two wedges of the acute, DblWidth wide and DblHeight
	// high (DblHeightCap on the capitals), DblHead and DblTip across at their
	// ends, with their heads DblPitch apart.
	DblWidth, DblHeight, DblHeightCap, DblHead, DblTip, DblPitch float64

	// The cedilla: a neck CedNeck thick which leaves the base CedOverlap above
	// the baseline, a hook which reaches CedReach to the right of the neck,
	// CedThick thick at its right, and a tail which ends CedTail to the left
	// of the neck, CedDepth below the baseline at its lowest, in a cut
	// CedEnd across.
	// A cedilla which is not joined to its base begins CedApart below it.
	CedNeck, CedOverlap, CedReach, CedThick, CedTail, CedDepth, CedEnd, CedApart float64

	// The ogonek: a hook which leaves the base OgoOverlap above the baseline
	// OgoNeck thick, bends OgoReach to the left of where it leaves, is
	// OgoThick thick at its left, OgoThin at its bottom and OgoDepth deep, and
	// ends in a cut OgoTip across OgoTail to the right of where it leaves.
	OgoNeck, OgoOverlap, OgoReach, OgoThick, OgoThin, OgoDepth, OgoTail, OgoTip float64

	// The comma below and the comma above of g: the comma scaled by
	// CommaScale, its top CommaGap below the baseline, or its lowest ink
	// CommaGap above the x-height.
	CommaScale, CommaGap float64

	// The caron of d, l, t and L: the right single quotation mark scaled by
	// ApoScale, its leftmost ink ApoGap to the right of the stem, with its top
	// at the top of the stem.
	ApoScale, ApoGap float64

	// MarkSide is the side space of a spacing accent on each side of its ink.
	// Overhang is the most by which the ink of a mark may reach beyond the
	// advance of its letter.
	MarkSide, Overhang float64

	// MarkFloor is the least distance between a mark and the neighbouring
	// letter which the kerning keeps.
	MarkFloor float64
}

func regularAccents() Accents {
	return Accents{
		Gap:    60,
		GapCap: 50,

		AcuteHead:      70,
		AcuteTip:       26,
		AcuteWidth:     170,
		AcuteHeight:    160,
		AcuteHeightCap: 130,
		AcuteShift:     16,

		CircThin:      46,
		CircThick:     46,
		CircWidth:     296,
		CircHeight:    150,
		CircHeightCap: 134,

		TildeWidth:     280,
		TildeHeight:    92,
		TildeHeightCap: 92,
		TildeThick:     50,
		TildeThin:      38,

		DotScale:       0.88,
		DiaPitch:       198,
		DiaPitchNarrow: 176,
		DotAboveScale:  0.92,

		MacronWidth:    270,
		MacronWidthCap: 300,
		MacronThick:    44,
		MacronLift:     40,

		BreveWidth:     260,
		BreveHeight:    130,
		BreveHeightCap: 117,
		BreveThick:     57,
		BreveThin:      32,

		RingDiam:    176,
		RingDiamCap: 176,
		RingThick:   40,
		RingThin:    34,

		DblWidth:     140,
		DblHeight:    165,
		DblHeightCap: 150,
		DblHead:      66,
		DblTip:       22,
		DblPitch:     145,

		CedNeck:    40,
		CedOverlap: 24,
		CedReach:   95,
		CedThick:   52,
		CedTail:    90,
		CedDepth:   210,
		CedEnd:     28,
		CedApart:   36,

		OgoNeck:    54,
		OgoOverlap: 26,
		OgoReach:   100,
		OgoThick:   56,
		OgoThin:    30,
		OgoTip:     22,
		OgoDepth:   210,
		OgoTail:    60,

		CommaScale: 0.92,
		CommaGap:   56,

		ApoScale: 0.9,
		ApoGap:   36,

		MarkSide: 40,
		Overhang: 60,

		MarkFloor: 40,
	}
}
