// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the pair kerning, which internal/kern
// computes from the outlines, and the table of exceptions.

// Kern holds the parameters of the kerning computation.
type Kern struct {
	// Band is the height of the horizontal bands in which the profile of a
	// glyph side is measured.
	Band float64

	// Depth is the cap of the white in one band: the horizontal distance
	// between two glyphs counts as at most this much, and a band without ink
	// on one side counts as this much.
	Depth float64

	// SideDepth is the cap of the depth of one glyph side in one band.
	SideDepth float64

	// Narrow is the weight of the white of the narrowest band in the white of
	// a pair of letters; the mean over the bands has the weight 1 - Narrow.
	Narrow float64

	// Mouth is the height, as a fraction of the x-height, below which an
	// opening of a glyph side counts only as deep as its entrance.
	Mouth float64

	// Allowance is the amount by which the white of a pair of lowercase
	// letters or of a letter and a mark may exceed the target before the pair
	// is kerned; the kern removes the excess over the target plus the
	// allowance.  It is an amount and not a fraction.
	Allowance float64

	// AllowanceUpper is the allowance of two capitals.
	AllowanceUpper float64

	// AllowanceMixed is the allowance of a capital next to a lowercase letter.
	AllowanceMixed float64

	// Soft is the scale of the soft limit of negative kerns.  The shift which
	// brings the white of an open pair (AV, To) to the target is not bounded.  A
	// shift s is applied as Soft times tanh(s / Soft): small shifts are
	// applied nearly in full and large ones come to a little less than Soft.
	Soft float64

	// Interplay is the fraction of the bands of the zone of a pair in which both
	// glyphs must have ink for the pair to be kerned.
	Interplay float64

	// Threshold is the size below which a kern is dropped.
	Threshold float64

	// MaxNegative is the largest negative kern, as a positive number.
	MaxNegative float64

	// Floor is the smallest distance between the outlines of two glyphs which
	// a kern may leave.  A pair which is closer than this without kerning gets
	// a positive kern.
	Floor float64

	// Approach is the fraction of the closest approach of the pair HH for
	// capitals, nn for lowercase letters and Hn for mixed pairs which a
	// negative kern leaves to the pair.  The closest approach is the smallest
	// band gap over all bands in which both glyphs have ink, over the whole
	// height.
	Approach float64

	// DashClear is the smallest distance between the outlines of a capital
	// and a hyphen or dash which a negative kern leaves; a pair closer than
	// that without a kern gets none.
	DashClear float64

	// Serif is the height at the two ends of the zone of a pair which the
	// second measure of the closest approach leaves out.
	Serif float64

	// StraightTol and StraightFrac say what a straight stem is: a glyph side
	// at which at least the fraction StraightFrac of the bands of the body of
	// the zone have ink within StraightTol of the nearest.
	StraightTol, StraightFrac float64

	// ClassTolerance is the largest difference of the capped depth of two glyph
	// sides in any band of their zone for which they share a class.
	ClassTolerance float64

	// PairTolerance is the largest difference between the kern of a pair and
	// the kern of its class pair for which the pair is left to the class.
	PairTolerance float64
}

// KernException is a kern which overrides the computation for one pair.
// Left and Right are glyph names.
type KernException struct {
	Left, Right string
	Value       float64
	Reason      string
}

// KernExceptions lists the pairs whose kern is set by hand instead of by the
// model of internal/kern.  Each entry has a Reason.
var KernExceptions = []KernException{}

// regularKern returns the kerning parameters of Bedern Regular.
func regularKern() Kern {
	return Kern{
		Band:           10,
		Depth:          400,
		SideDepth:      300,
		Narrow:         0.5,
		Mouth:          0.5,
		Allowance:      70,
		AllowanceUpper: 40,
		AllowanceMixed: 20,
		Soft:           100,
		Interplay:      0.4,
		Threshold:      8,
		MaxNegative:    135,
		Floor:          20,
		Approach:       0.85,
		DashClear:      50,
		Serif:          40,
		StraightTol:    12,
		StraightFrac:   0.8,
		ClassTolerance: 8,
		PairTolerance:  3,
	}
}
