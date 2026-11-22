// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the quantities which follow from the side spaces and the
// bodies of the letters.  They are derived, not stored.

// NAdvance returns the advance of the letter n: the left side space, two
// stems, the counter and the right side space.
func NAdvance(p Params) float64 {
	return p.SideStemLeft + 2*p.Stem + p.NCounter + p.SideStemRight
}

// NGap returns the distance between the stems of two adjacent letters n, which
// is the right side space of one plus the left side space of the next.
func NGap(p Params) float64 {
	return p.SideStemRight + p.SideStemLeft
}

// OBowlLeft returns the position of the left extreme of the bowl of o.
func OBowlLeft(p Params) float64 {
	return p.SideRoundLeft
}

// OBowlRight returns the position of the right extreme of the bowl of o.
func OBowlRight(p Params) float64 {
	return p.SideRoundLeft + p.OBowlWidth
}

// OAdvance returns the advance of the letter o: the bowl between its two
// round side spaces.
func OAdvance(p Params) float64 {
	return p.SideRoundLeft + p.OBowlWidth + p.SideRoundRight
}

// NarrowAdvance returns the advance of a narrow letter with a single stem,
// such as l: the stem between the stem sides, each reduced by NarrowSideCut.
func NarrowAdvance(p Params) float64 {
	return p.SideStemLeft + p.Stem + p.SideStemRight - 2*p.NarrowSideCut
}

// NarrowStemLeft returns the position of the left edge of the stem of a
// narrow letter.  The stem keeps the proportion of the stem sides of n: it
// stands SideStemLeft less the cut from the left edge, and the right side space
// is SideStemRight less the cut.
func NarrowStemLeft(p Params) float64 {
	return p.SideStemLeft - p.NarrowSideCut
}

// DotOffset returns the distance of the centre of the dot of i and j to the
// right of the centre of the stem, which is negative.  The dot moves to the
// left by DotShift times the length of the head serif.
func DotOffset(p Params) float64 { return -p.DotShift * p.SerifJut }

// NarrowSideLeft returns the space between the left edge of the advance and
// the stem of a narrow letter, which is NarrowStemLeft.
func NarrowSideLeft(p Params) float64 { return NarrowStemLeft(p) }

// NarrowSideRight returns the space between the stem of a narrow letter and
// the right edge of the advance.
func NarrowSideRight(p Params) float64 {
	return p.SideStemRight - p.NarrowSideCut
}

// DStemLeft returns the position of the left-hand edge of the stem of d: the
// left round side space and the bowl lie to its left.
func DStemLeft(p Params) float64 {
	return DBowlLeft(p) + p.DBowl.Width
}

// DBowlLeft returns the position of the left extreme of the bowl of d, which is
// that of q as well: the side space DQSideLeft.
func DBowlLeft(p Params) float64 {
	return p.DQSideLeft
}

// DAdvance returns the advance of the letter d: the round side space, the bowl,
// the stem and the right side space of the stem of d.
func DAdvance(p Params) float64 {
	return DStemLeft(p) + p.Stem + p.DSideRight
}

// EBowlLeft returns the position of the left extreme of the bowl of e.
func EBowlLeft(p Params) float64 {
	return p.SideRoundLeft
}

// EBowlRight returns the position of the right extreme of the bowl of e.
func EBowlRight(p Params) float64 {
	return p.SideRoundLeft + p.EBowlWidth
}

// EAdvance returns the advance of the letter e: the round left side space, the
// bowl and the open right side space.
func EAdvance(p Params) float64 {
	return p.SideRoundLeft + p.EBowlWidth + p.SideOpenRight
}

// CAdvance returns the advance of the letter c: the left side space CSideLeft, the
// ink, which ends at the tip of the lower terminal, and the right side space
// CSideRight.  CBowlWidth is the width of the ink, from the left extreme to the
// tip, and CSideRight is measured to the rightmost ink.
func CAdvance(p Params) float64 {
	return CBowlLeft(p) + p.CBowlWidth + p.CSideRight
}

// CBowlLeft returns the position of the left extreme of c.
func CBowlLeft(p Params) float64 {
	return p.CSideLeft
}

// AStemLeft returns the position of the left-hand edge of the stem of a: the
// left side space and the bowl lie to its left.
func AStemLeft(p Params) float64 {
	return p.ASideLeft + p.ABowl.Width
}

// AAdvance returns the advance of the letter a: the left side space, the bowl,
// the stem, the tail and the right side space.  The right side space is
// measured from the end of the tail: the advance is made from the reach of the
// tail and ATailSideRight, as that of t is made from TReach and TSideRight.
func AAdvance(p Params) float64 {
	return AStemLeft(p) + p.Stem + p.ATailReach + p.ATailSideRight
}

// SAdvance returns the advance of the letter s: the left side space, the lower
// half, which is the wider and holds the rightmost and the leftmost ink, and the
// right side space.
func SAdvance(p Params) float64 {
	return p.SSideLeft + p.SLowerWidth + p.SSideRight
}

// AscenderInset returns how far each edge of an ascender stem (l, d, the left
// stem of h) lies inside the edge of a stem of width Stem.  The lighter stem
// stands in the cell of the regular stem with the same centre line: every
// advance is as for AscenderStem equal to Stem, and the white on both sides of
// the stem grows by this amount.
func AscenderInset(p Params) float64 {
	return (p.Stem - p.AscenderStem) / 2
}

// MAdvance returns the advance of the letter m: the left side space, three
// stems, the two counters and the right side space.
func MAdvance(p Params) float64 {
	return p.SideStemLeft + 3*p.Stem + 2*p.MCounter + p.SideStemRight
}

// MFootGap returns the clear distance between the tips of the inner foot
// serifs of two neighbouring stems of m.
func MFootGap(p Params) float64 {
	return p.MCounter - 2*p.SerifJutInner
}

// UAdvance returns the advance of the letter u: the left side space, two stems,
// the counter and the right side space.
func UAdvance(p Params) float64 {
	return p.SideStemLeft + 2*p.Stem + p.UCounter + p.SideStemRight
}

// RAdvance returns the advance of the letter r: the left side space RSideLeft, the stem,
// the reach of the arm with its ball beyond the stem and the space to the right
// of the ball.
func RAdvance(p Params) float64 {
	return RStemLeft(p) + p.Stem + p.RReach + p.RSideRight
}

// RStemLeft returns the position of the left-hand edge of the stem of r: the
// side space RSideLeft.
func RStemLeft(p Params) float64 {
	return p.RSideLeft
}

// BStemLeft returns the position of the left-hand edge of the stem of b: the
// side space BPSideLeft, which is measured to the edge of a stem of width Stem,
// and the inset of the lighter ascender stem (AscenderInset).
func BStemLeft(p Params) float64 {
	return p.BPSideLeft + AscenderInset(p)
}

// BStemRight returns the position of the right-hand edge of the stem of b, on
// which the bowl hangs.
func BStemRight(p Params) float64 {
	return BStemLeft(p) + p.AscenderStem
}

// BBowlRight returns the position of the right extreme of the bowl of b, the
// mirror image of d: the cell of the stem (a stem of width Stem) and the width
// of the bowl, as for d.
func BBowlRight(p Params) float64 {
	return p.BPSideLeft + p.Stem + p.BBowl.Width
}

// BAdvance returns the advance of the letter b: the left side space of the
// stem, the stem, the bowl and the right side space of the bowl, BPSideRight.
func BAdvance(p Params) float64 {
	return BBowlRight(p) + p.BPSideRight
}

// PStemLeft returns the position of the left-hand edge of the stem of p.
func PStemLeft(p Params) float64 {
	return p.BPSideLeft
}

// PStemRight returns the position of the right-hand edge of the stem of p,
// on which the bowl hangs.
func PStemRight(p Params) float64 {
	return p.BPSideLeft + p.Stem
}

// PBowlRight returns the position of the right extreme of the bowl of p.
func PBowlRight(p Params) float64 {
	return PStemRight(p) + p.PBowl.Width
}

// PAdvance returns the advance of the letter p: the left side space of the
// stem, the stem, the bowl and the right side space of the bowl, BPSideRight.
func PAdvance(p Params) float64 {
	return PBowlRight(p) + p.BPSideRight
}

// QStemLeft returns the position of the left-hand edge of the stem of q: the
// round left side space and the bowl, which is that of d, lie to its left.
func QStemLeft(p Params) float64 {
	return DBowlLeft(p) + p.QBowl.Width
}

// QAdvance returns the advance of the letter q: the round side space, the
// bowl, the stem and the right side space of the stem of q.
func QAdvance(p Params) float64 {
	return QStemLeft(p) + p.Stem + p.QSideRight
}
