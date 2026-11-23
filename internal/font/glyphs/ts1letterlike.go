// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letterlike and commercial signs of TS1: ™ ℠ ℗ № ℮ ℞ ℃ Ω ℧ and the
// copyleft sign.
//
// The trademark and the service mark are two small capitals side by side
// with their tops at the cap height; the small capitals are the capitals
// drawn with heavier strokes and scaled down, as the superior figures are.
// The published sign and the copyleft sign are the ring of © round a P and
// round a C which opens to the left.

// Shape constants of the letterlike signs.
const (
	markScale   = 0.53 // the size of the letters of ™ and ℠, in sizes of the capitals
	markGap     = 40.0 // the white added between the letters of ™ and ℠
	markSide    = 35.0 // the side spaces of ™ and ℠
	markBracket = 8.0  // the least height of the bracket of a thin stem of ™ and ℠ above the tip of its serif
)

// letterlikeList lists the letterlike and commercial signs of TS1, in glyph
// order.
var letterlikeList = []entry{
	{"trademark", '\u2122', sideOpen, sideOpen, trademark},
	{"uni2120", '\u2120', sideOpen, sideOpen, servicemark},
	{"uni2117", '\u2117', sideOpen, sideOpen, published},
	{"uni2116", '\u2116', sideCapN, sideOpen, numero},
	{"estimated", '\u212E', sideOpen, sideOpen, estimated},
	{"prescription", '\u211E', sideCapBowlStem, sideOpen, prescription},
	{"uni2103", '\u2103', sideOpen, sideCapC, celsius},
	{"Omega", '\u2126', sideCapRound, sideCapRound, ohm},
	{"uni2127", '\u2127', sideCapRound, sideCapRound, mho},
	{"u1F12F", '\U0001F12F', sideOpen, sideOpen, copyleft},
}

func init() {
	sharpCrotches["trademark"] = func(p design.Params) []sharpCrotch { return markCrotches(p, "T") }
	sharpCrotches["uni2120"] = func(p design.Params) []sharpCrotch { return markCrotches(p, "S") }
}

// heavierCapitals returns the parameters with the strokes of the capitals k
// times as thick, those of the diagonal capitals and of S and the brackets
// of the serifs among them.
func heavierCapitals(p design.Params, k float64) design.Params {
	q := heavierLetters(p, k)
	c, d, g := &q.Caps, &q.CapDiag, &q.Caps.CGSJQ
	for _, v := range []*float64{&c.BracketHeight, &c.BracketReach,
		&d.Thick, &d.Thin, &d.WThick, &d.XThick, &d.MThick, &d.NThick, &d.KLeg, &d.ZDiag, &d.ThinStem, &d.MStem,
		&g.SSpineThick, &g.SArcThick, &g.SUpperArcThick, &g.SSideThick, &g.BeakArc} {
		*v *= k
	}
	return q
}

// markParams returns the parameters with which the letters of ™ and ℠ are
// drawn before they are scaled by markScale: the strokes of the capitals
// Figures.SupWeight/markScale times as thick, so that the strokes of the
// small capitals are Figures.SupWeight of those of the capitals, as those of
// the superior figures are of those of the figures.
func markParams(p design.Params) design.Params {
	q := heavierCapitals(p, p.Figures.SupWeight/markScale)
	c := &q.Caps
	// The bracket of a thin stem leaves the stem at least markBracket above
	// the tip of the serif, in units of the scaled letter.
	c.BracketHeight = math.Max(c.BracketHeight, (c.SerifThick+markBracket/markScale)/thinStemBracket)
	return q
}

// smallCapKey names a small capital in one set of parameters.
type smallCapKey struct {
	p      design.Params
	letter string
}

// smallCaps holds the small capitals of ™ and ℠ by the parameters and the
// letter, so that the two signs draw the M once.
var smallCaps sync.Map // smallCapKey -> func() smallCap

// A smallCap is a capital drawn with markParams and scaled by markScale, with
// its leftmost ink at the abscissa 0 and its foot on the baseline, the
// distance by which it was moved to the right after the scaling, and its ink
// width.
type smallCap struct {
	path  outline.Path
	shift float64
	width float64
}

// smallCapOf returns the small capital of the letter.  The path is shared and
// must not be changed.
func smallCapOf(p design.Params, letter string) smallCap {
	f, _ := smallCaps.LoadOrStore(smallCapKey{p, letter}, sync.OnceValue(func() smallCap {
		path, _, _ := Outline(letter, markParams(p))
		path = scaledPath(path, markScale)
		x0, _, x1, _ := path.BBox()
		return smallCap{path: shiftInk(path, -x0, 0), shift: -x0, width: x1 - x0}
	}))
	return f.(func() smallCap)()
}

// markPlaces returns the distances by which the two small capitals of a mark
// are moved from their places on the baseline: the first letter to the right
// by x1 and the M by x2, both up by y.
func markPlaces(p design.Params, first string) (x1, x2, y float64) {
	a := smallCapOf(p, first)
	x1 = markSide
	x2 = math.Round(x1 + a.width + markGap)
	return x1, x2, math.Round((1 - markScale) * p.CapHeight)
}

// mark draws ™ or ℠: the small capital of the first letter and that of M,
// markGap apart, with their tops at the cap height and markSide on each side.
func mark(p design.Params, first string) (outline.Path, float64) {
	a, m := smallCapOf(p, first), smallCapOf(p, "M")
	x1, x2, y := markPlaces(p, first)
	path := append(shiftInk(a.path, x1, y), shiftInk(m.path, x2, y)...)
	return path, math.Round(x2 + m.width + markSide)
}

func trademark(p design.Params) (outline.Path, float64)   { return mark(p, "T") }
func servicemark(p design.Params) (outline.Path, float64) { return mark(p, "S") }

// markCrotches returns the sharp crotches of the M of the mark which begins
// with the given letter.
func markCrotches(p design.Params, first string) []sharpCrotch {
	m := smallCapOf(p, "M")
	_, x2, y := markPlaces(p, first)
	ks := capMCrotches(markParams(p))
	for i := range ks {
		ks[i].Tip = outline.Point{X: markScale*ks[i].Tip.X + m.shift + x2, Y: markScale*ks[i].Tip.Y + y}
	}
	return ks
}

// published draws ℗: the ring of © round a P.
func published(p design.Params) (outline.Path, float64) { return circled(p, "P") }

// copyleft draws the copyleft sign: the ring of © round a C which opens to
// the left.
func copyleft(p design.Params) (outline.Path, float64) {
	l, _, _ := Outline("C", heavierLetters(p, p.Figures.SupWeight/supScale))
	var mirrored outline.Path
	for _, c := range pathChains(l) {
		c.mirrorX(0).addContour(&mirrored)
	}
	return circledLetter(p, mirrored)
}
