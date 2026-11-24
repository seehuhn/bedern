// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the old-style figures and of the
// proportional figures.  A figure has four forms: lining or old-style, each
// tabular or proportional.  The tabular forms stand on the advance of the
// figures, the old-style ones on the advance less AdvanceCut; a proportional form is the outline of the tabular form between side
// spaces of its own.  The construction is described in
// internal/glyphs/variantfigures.go.

// OldFigures holds the parameters of the old-style figures and of the
// proportional figures.
type OldFigures struct {
	// 0.  ZeroRound is the thickness of the sides and ZeroThin that of the
	// top and the bottom.
	ZeroRound float64
	ZeroThin  float64

	// 1.  The fields are those of the lining 1 (Figures.OneFlag,
	// OneFlagDrop, OneFlagLean, OneFoot and OneRise); OneRise is the height
	// of the top above the x-height.
	OneFlag     float64
	OneFlagDrop float64
	OneFlagLean float64
	OneFoot     float64
	OneRise     float64

	// 2.  TwoInk is the width of the ink; the other fields are those of
	// the lining 2 (Figures.TwoBowlMid, TwoShoulder, TwoOuterDrop,
	// TwoDiagFoot, TwoDiagAngle, TwoBase, TwoBaseEnd, TwoRise, TwoBallLeft
	// and TwoBowlLeft).
	// TwoBall is the diameter of the ball and TwoHeavy the thickness of the
	// base.
	TwoInk       float64
	TwoBowlMid   float64
	TwoShoulder  float64
	TwoOuterDrop float64
	TwoDiagFoot  float64
	TwoDiagAngle float64
	TwoBallLeft  float64
	TwoBowlLeft  float64
	TwoBase      float64
	TwoBaseEnd   float64
	TwoRise      float64
	TwoBall      float64
	TwoHeavy     float64

	// AdvanceCut is how much narrower the advance of the tabular old-style
	// figures is than Figures.Advance.  The ink of every figure keeps its
	// place in the middle of the advance, and the ink of 4, which has the
	// least white to the edges of the advance, is narrower by AdvanceCut as
	// well, so that 4 keeps its distance from the edges.
	AdvanceCut float64

	// Left and Right are the side spaces of the proportional old-style
	// figures, from the edge of the advance to the ink, and LiningLeft and
	// LiningRight those of the proportional lining figures.
	Left, Right             [10]float64
	LiningLeft, LiningRight [10]float64
}

// regularOldFigures returns the parameters of the old-style and the
// proportional figures of Bedern Regular.
func regularOldFigures() OldFigures {
	return OldFigures{
		ZeroRound: 88,
		ZeroThin:  44,

		OneFlag:     119,
		OneFlagDrop: 70,
		OneFlagLean: 20,
		OneFoot:     112,
		OneRise:     10,

		TwoInk:       374,
		TwoBowlMid:   300,
		TwoShoulder:  330,
		TwoOuterDrop: 14,
		TwoDiagFoot:  144,
		TwoDiagAngle: 56,
		TwoBallLeft:  18,
		TwoBowlLeft:  6,
		TwoBase:      390,
		TwoBaseEnd:   16,
		TwoRise:      76,
		TwoBall:      72,
		TwoHeavy:     66,

		AdvanceCut: 0,

		//                        0   1   2   3   4   5   6   7   8   9
		Left:        [10]float64{36, 24, 38, 26, 16, 27, 38, 37, 35, 31},
		Right:       [10]float64{36, 30, 30, 28, 33, 18, 32, 31, 38, 36},
		LiningLeft:  [10]float64{44, 48, 37, 39, 18, 47, 55, 42, 48, 47},
		LiningRight: [10]float64{44, 43, 60, 49, 17, 46, 46, 27, 50, 56},
	}
}

// OldFigureDepth returns the depth below the baseline of the round bottoms
// of the descending old-style figures 3, 5 and 9, which is the depth of the
// descenders; the flat feet of 4 and 7 stand OvershootLower above it.  The
// flat tops of the descending figures and of the figures 0, 1 and 2 are at
// the x-height and the round tops OvershootLower above it.
func OldFigureDepth(p Params) float64 { return p.Descender }

// FigureSides returns the side spaces of the proportional form of the figure
// d, old-style if old is set and lining otherwise.
func FigureSides(p Params, d int, old bool) PunctSides {
	o := p.Figures.Old
	if old {
		return PunctSides{Left: o.Left[d], Right: o.Right[d]}
	}
	return PunctSides{Left: o.LiningLeft[d], Right: o.LiningRight[d]}
}
