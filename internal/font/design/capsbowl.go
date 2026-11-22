// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import "math"

// This file holds the parameters of the capitals B, D, P, R and U and their
// advances.  They are the embedded field CapBowls of Caps.  The
// construction is described in internal/glyphs/capbowl.go and the files of the
// letters.

// CapBowls holds the parameters of the capitals B, D, P, R and U.  All widths
// of a bowl are measured from the right-hand edge of the stem to the right
// extreme of the bowl, which is the end of the ink.
type CapBowls struct {
	// Strokes.  BowlRound is the thickness of a bowl at its right extreme; the
	// thin strokes with which a bowl leaves and meets the stem are Bar thick.
	// BowlFlat is the fraction of the width of the bowls of B, P and R over
	// which these strokes run straight before the round begins, and DFlat that
	// of D.
	BowlRound float64
	BowlFlat  float64
	DFlat     float64

	// B.  BWaist is the height of the centre of the waist, the thin stroke
	// between the two bowls, which is Bar thick; BUpper and BLower are the
	// widths of the upper and the lower bowl, and BUpperRound and BLowerRound
	// their thicknesses at the right extreme.
	BWaist      float64
	BUpper      float64
	BLower      float64
	BUpperRound float64
	BLowerRound float64

	// P and R.  PBar and RBar are the heights of the centre of the thin stroke
	// under the bowl, PWidth and RWidth the widths of the bowls, and
	// RBowlRound the thickness of the bowl of R at its right extreme.
	PBar       float64
	PWidth     float64
	RBar       float64
	RWidth     float64
	RBowlRound float64

	// The leg of R.  RLegTop is the distance from the right-hand edge of the
	// stem to the left edge of the leg where it leaves the underside of the
	// bowl, RLegFoot that to the left edge of the leg on the baseline, and
	// RLeg the thickness of the leg at right angles.  The leg ends in a foot serif of the
	// diagonals, which overhangs the left edge by RLegSerifLeft and the right
	// edge by RLegSerifRight.
	RLegTop        float64
	RLegFoot       float64
	RLeg           float64
	RLegSerifLeft  float64
	RLegSerifRight float64

	// D.  DWidth is the width of its bowl.
	DWidth float64

	// U.  UThin is the width of the right stem, UCounter the distance between
	// the inner edges of the stems and UCurve the height at which the outer
	// edges leave the stems for the round at the bottom; the round is RoundThin
	// thick at its lowest point and overshoots the baseline by OvershootUpper.
	UThin    float64
	UCounter float64
	UCurve   float64

	// Side spaces.  BowlSideLeft is the distance from the left edge of the
	// advance to the left-hand edge of the stem of B, D, P and R; BowlSideRight,
	// BSideRight and DSideRight are those from the right extreme of the bowl of
	// P, B and D to the right edge of the advance, and RSideRight that from the
	// tip of the right foot serif of the leg of R.
	BowlSideLeft  float64
	BowlSideRight float64
	BSideRight    float64
	DSideRight    float64
	RSideRight    float64
}

// regularCapBowls returns the parameters of B, D, P, R and U of Bedern Regular.
func regularCapBowls() CapBowls {
	return CapBowls{
		BowlRound: 105,
		BowlFlat:  0.45,
		DFlat:     0.33,

		BWaist:      355,
		BUpper:      354,
		BLower:      376,
		BUpperRound: 96,
		BLowerRound: 98,

		PBar:       333,
		PWidth:     372,
		RBar:       347,
		RWidth:     364,
		RBowlRound: 100,

		RLegTop:        92,
		RLegFoot:       322,
		RLeg:           90,
		RLegSerifLeft:  60,
		RLegSerifRight: 60,

		DWidth: 464,

		UThin:    42,
		UCounter: 341,
		UCurve:   222,

		BowlSideLeft:  119,
		BowlSideRight: 36,
		BSideRight:    68,
		DSideRight:    58,
		RSideRight:    16,
	}
}

// CapBowlStemLeft returns the position of the left-hand edge of the stem of B,
// D, P and R.
func CapBowlStemLeft(p Params) float64 { return p.Caps.BowlSideLeft }

// capBowlAdvance returns the advance of a bowl letter whose widest bowl is w
// wide and whose right side is side.
func capBowlAdvance(p Params, w, side float64) float64 {
	c := p.Caps
	return c.BowlSideLeft + c.Stem + w + side
}

// CapBAdvance returns the advance of B: the lower bowl is the wider.
func CapBAdvance(p Params) float64 { return capBowlAdvance(p, p.Caps.BLower, p.Caps.BSideRight) }

// CapDAdvance returns the advance of D.
func CapDAdvance(p Params) float64 { return capBowlAdvance(p, p.Caps.DWidth, p.Caps.DSideRight) }

// CapPAdvance returns the advance of P.
func CapPAdvance(p Params) float64 { return capBowlAdvance(p, p.Caps.PWidth, p.Caps.BowlSideRight) }

// CapRLegLean returns the lean of the leg of R, dx/dy along it going up: the
// left edge runs from RLegFoot on the baseline to RLegTop under the bowl.
func CapRLegLean(p Params) float64 {
	c := p.Caps
	return (c.RLegTop - c.RLegFoot) / (c.RBar - c.Bar/2)
}

// CapRAdvance returns the advance of R: the stem, the leg on the baseline and
// its right foot serif, whose overhang is measured at the level of its upper
// face, with the space after the tip of the serif, rounded to a whole unit.
func CapRAdvance(p Params) float64 {
	c := p.Caps
	lean := CapRLegLean(p)
	return math.Round(c.BowlSideLeft + c.Stem + c.RLegFoot + DiagWidth(c.RLeg, lean) + lean*c.SerifThick + c.RLegSerifRight + c.RSideRight)
}

// CapUAdvance returns the advance of U: the stem sides, the two stems and the
// counter.
func CapUAdvance(p Params) float64 {
	c := p.Caps
	return c.SideStemLeft + c.Stem + c.UCounter + c.UThin + c.SideStemRight
}
