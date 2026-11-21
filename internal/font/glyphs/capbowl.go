// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The bowl of the capitals B, D, P and R.
//
// A capital bowl stands on the right of a capital stem.  It leaves the stem at
// the top as a thin horizontal stroke, Bar thick, which meets the right-hand
// edge of the stem squarely, as the arms of E do, runs straight to the point
// Flat, rounds over to the right extreme, where the stroke is thickest
// (BowlRound), and returns to the stem below as a thin horizontal stroke of the
// same kind.  The bowl is symmetric about its middle height: the round is two
// quarter curves on each edge, one above and one below the extreme, and the
// curvature is continuous at the extreme.
//
// Each quarter is one cubic from the end of the straight stroke, where it
// leaves horizontally, to the extreme, where it arrives vertically.  Its
// handles are fractions of the run and of the rise of the quarter (the fields
// of a capBowlShape): the handle at the end of the straight stroke is long, and
// the curvature begins small after the straight stroke and grows evenly to its
// greatest value at the shoulder; the handle at the extreme is shorter.  The
// counter is drawn by the same construction inside the outer edge, with quarters
// of the proportion of those of the outer edge (innerFlat): the stroke is Bar
// thick where it is straight and BowlRound at the extreme, and it swells evenly
// between them.
//
// The parts return chains.  outer runs from the stem edge at the bottom of the
// bowl, along the underside of the lower stroke, round the bowl and back along
// the top of the upper stroke to the stem edge; the letters cut it where a leg
// (R) or a second bowl (B) takes over.  counter is the counter as a clockwise
// contour of its own, bounded by the stem edge.

// A capBowlShape holds the handles of the quarter curves of a capital bowl:
// for the outer edge and for the counter, the handle at the end of the straight
// stroke as a fraction of the run of the quarter, and the handle at the extreme
// as a fraction of its rise.
type capBowlShape struct {
	OuterFlat, OuterExtreme float64
	InnerFlat, InnerExtreme float64
}

// capBowlRound is the shape of the bowls of B, D, P and R.  A quarter of the
// super-ellipse of the o has the handles 0.587 at both ends.
var capBowlRound = capBowlShape{
	OuterFlat: 0.75, OuterExtreme: 0.55,
	InnerFlat: 0.75, InnerExtreme: 0.55,
}

// A capBowl describes one bowl on the right of a capital stem.
type capBowl struct {
	Edge        float64 // the right-hand edge of the stem
	Flat        float64 // where the straight strokes end and the round begins
	Right       float64 // the right extreme of the outer edge
	Bottom, Top float64 // the outer edges of the lower and the upper stroke
	Bar         float64 // the thickness of the straight strokes
	Thick       float64 // the thickness at the extreme
	Shape       capBowlShape
}

// capBowlFrom returns the bowl of the width w (from the stem edge to the
// extreme) on the stem whose right-hand edge is at edge, between the heights
// bottom and top, whose strokes run straight for the fraction flat of w.
func capBowlFrom(p design.Params, edge, w, flat, bottom, top float64) capBowl {
	return capBowl{
		Edge: edge, Flat: edge + flat*w, Right: edge + w,
		Bottom: bottom, Top: top,
		Bar: p.Caps.Bar, Thick: p.Caps.BowlRound,
		Shape: capBowlRound,
	}
}

func (b capBowl) mid() float64 { return (b.Bottom + b.Top) / 2 }

// innerFlat returns where the straight strokes of the counter end: the
// quarters of the counter have the proportion of run to rise of the quarters of
// the outer edge.  The counter is less tall than the outer edge by twice Bar but
// less wide by Thick alone, and so its straight strokes end before those of the
// outer edge, and the stroke swells from Bar to Thick along the whole of the
// round.  The length of the straight strokes is not less than zero.
func (b capBowl) innerFlat() float64 {
	half := b.Top - b.mid()
	x := b.Right - b.Thick - (half-b.Bar)*(b.Right-b.Flat)/half
	return math.Max(x, b.Edge)
}

// flatToExtreme returns the quarter from the end of a straight stroke at a,
// which it leaves horizontally, to the extreme at e, which it reaches
// vertically, with the handles fa (of the run) and fe (of the rise).
func flatToExtreme(a, e outline.Point, fa, fe float64) cubic {
	return cubic{
		P0: a,
		P1: outline.Point{X: a.X + fa*(e.X-a.X), Y: a.Y},
		P2: outline.Point{X: e.X, Y: e.Y + fe*(a.Y-e.Y)},
		P3: e,
	}
}

// outerLower returns the lower quarter of the outer edge, from the end of the
// lower stroke up to the extreme.
func (b capBowl) outerLower() cubic {
	return flatToExtreme(outline.Point{X: b.Flat, Y: b.Bottom}, outline.Point{X: b.Right, Y: b.mid()}, b.Shape.OuterFlat, b.Shape.OuterExtreme)
}

// outerUpper returns the upper quarter of the outer edge, from the extreme up
// to the end of the upper stroke.
func (b capBowl) outerUpper() cubic {
	return flatToExtreme(outline.Point{X: b.Flat, Y: b.Top}, outline.Point{X: b.Right, Y: b.mid()}, b.Shape.OuterFlat, b.Shape.OuterExtreme).reversed()
}

// innerUpper returns the upper quarter of the counter, from the end of the
// upper stroke down to the extreme of the counter.
func (b capBowl) innerUpper() cubic {
	return flatToExtreme(outline.Point{X: b.innerFlat(), Y: b.Top - b.Bar}, outline.Point{X: b.Right - b.Thick, Y: b.mid()}, b.Shape.InnerFlat, b.Shape.InnerExtreme)
}

// innerLower returns the lower quarter of the counter, from its extreme down to
// the end of the lower stroke.
func (b capBowl) innerLower() cubic {
	return flatToExtreme(outline.Point{X: b.innerFlat(), Y: b.Bottom + b.Bar}, outline.Point{X: b.Right - b.Thick, Y: b.mid()}, b.Shape.InnerFlat, b.Shape.InnerExtreme).reversed()
}

// outer returns the outer edge of the bowl: from the stem edge at Bottom along
// the lower stroke, round the bowl and along the upper stroke to the stem edge
// at Top.
func (b capBowl) outer() chain {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	return join(
		lineChain(pt(b.Edge, b.Bottom), pt(b.Flat, b.Bottom)),
		chain{curveSeg(b.outerLower()), curveSeg(b.outerUpper())},
		lineChain(pt(b.Flat, b.Top), pt(b.Edge, b.Top)),
	)
}

// counter returns the counter of the bowl as a clockwise closed chain: from the
// stem edge under the upper stroke, along it, round the inside of the bowl,
// back along the lower stroke and up the stem edge.
func (b capBowl) counter() chain {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	top, bottom := b.Top-b.Bar, b.Bottom+b.Bar
	return join(
		lineChain(pt(b.Edge, top), pt(b.innerFlat(), top)),
		chain{curveSeg(b.innerUpper()), curveSeg(b.innerLower())},
		lineChain(pt(b.innerFlat(), bottom), pt(b.Edge, bottom)),
		lineChain(pt(b.Edge, bottom), pt(b.Edge, top)),
	)
}

// capBowlStem returns the stem of B, D, P and R: the capital stem at
// BowlSideLeft with the serifs on the left at head and foot, and on the right
// at the foot if footRight is set.
func capBowlStem(p design.Params, footRight bool) capStem {
	j := p.Caps.SerifJut
	s := capStem{Left: design.CapBowlStemLeft(p), FootLeft: j, HeadLeft: j}
	if footRight {
		s.FootRight = j
	}
	return s
}

// capBowlLetter returns the outline of a bowl letter: the closed outer chain c,
// which must end at the top of the left-hand edge of the stem s, completed with
// the head serif and the left-hand edge with its foot serif, and the counters.
func capBowlLetter(p design.Params, s capStem, c chain, counters ...chain) outline.Path {
	head := capHeadLeft(p, s.Left, p.CapHeight, s.HeadLeft)
	c = join(c, lineChain(c.end(), head.start()), head)
	path := capContour(join(c, s.down(p, c.end())))
	for _, ct := range counters {
		path = append(path, capContour(ct)...)
	}
	return path
}
