// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// m draws the lowercase letter m as one counter-clockwise contour: three stems
// and two arches of n.  The first stem has the head serif and both feet, the
// last one both feet, the middle one the two inner feet; the inner feet are
// those of n, SerifJutInner long, and leave the clear gap MFootGap between the
// tips of neighbouring ones.  Both counters are MCounter wide and the arches use
// the arch of n between the stems they join (archBetween): its handles are
// fractions of the distance to the peak.
//
// The contour is that of n with the middle stem added.  It runs
//
//	the left stem up to the join of the first arch (stemUp)
//	the underside of the first arch to the middle stem (archUnder)
//	the middle stem down its left-hand edge and along its left foot (stemDown),
//	  and up its right-hand foot and edge to the join of the second arch (stemUp)
//	the underside of the second arch to the right stem (archUnder)
//	the right stem down and up again, as in n, to the shoulder of its outer edge
//	the outer edge of the second arch to the notch on the middle stem
//	  (archOuter), and from there the outer edge of the first arch to the notch on
//	  the left stem (archOuter)
//	the top and the left-hand edge of the left stem (stemCap, stemDown)
//
// The two outer edges meet at the notch on the right-hand edge of the middle
// stem.  The outer edge of the second arch ends there at the angle at which it
// leaves the stem, and the first arch starts there, as the right stem of n
// starts at its shoulder, vertically: the first arch leaves the middle stem at
// the height of the notch (ArchNotch) and not at ShoulderStart.  The two edges
// meet at the angle of the notch of n.
func m(p design.Params) (outline.Path, float64) {
	first := stemSpec{
		Left: p.SideStemLeft, Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJut, FootRight: p.SerifJutInner,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	middle := stemSpec{
		Left: first.right() + p.MCounter, Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJutInner, FootRight: p.SerifJutInner,
	}
	last := stemSpec{
		Left: middle.right() + p.MCounter, Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJutInner, FootRight: p.SerifJut,
	}
	archL := archBetween(p, first, middle)
	archL.Shoulder = archL.Notch // the first arch ends at the notch of the second
	archL.AtNotch = true
	archR := archBetween(p, middle, last)

	var path outline.Path
	path.MoveTo(first.start())
	stemUp(&path, p, first, archL.Join)
	archUnder(nArchShape, archL).appendTo(&path)
	stemDown(&path, p, middle)
	stemUp(&path, p, middle, archR.Join)
	archUnder(nArchShape, archR).appendTo(&path)
	stemDown(&path, p, last)
	stemUp(&path, p, last, archR.Shoulder)
	archOuter(nArchShape, archR).appendTo(&path)
	archOuter(nArchShape, archL).appendTo(&path)
	stemCap(&path, p, first)
	stemDown(&path, p, first)
	return path, design.MAdvance(p)
}
