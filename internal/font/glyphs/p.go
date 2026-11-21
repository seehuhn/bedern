// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// pBowl returns the bowl of p: the mirrored bowl of d on the right of the stem,
// with the upper stroke of d and the lower stroke of b (see design.Params.PBowl).
func pBowl(p design.Params) bowlSpec {
	b := bowlFrom(p, p.PBowl, design.PStemRight(p), design.PBowlRight(p))
	b.OnRight = true
	return b
}

// descenderStem returns the stem of p and q: from the descender line to the flat
// top at the x-height, with a foot serif on both sides at the descender line,
// at the left edge left.  p has the head serif of n on the left at the top, q
// none.
func descenderStem(p design.Params, left, top float64, head bool) stemSpec {
	s := stemSpec{
		Left: left, Width: p.Stem, Bottom: -p.Descender, Top: top,
		FootLeft: p.SerifJut, FootRight: p.SerifJut,
	}
	if head {
		s.HeadJut, s.HeadDrop = p.SerifJut, p.HeadSerifDrop
	}
	return s
}

// p draws the lowercase letter p as one outer contour and one counter: a stem
// from the descender line to the x-height, with a head serif on the left at the
// top as n has it and a foot serif on both sides at the descender line, and on
// its right the mirrored bowl of d.  The bowl joins the stem above the baseline
// and the stem continues below it.  The outer contour starts at the left tip of
// the left foot serif, goes up the right-hand edge of the stem to the lower
// notch, round the bowl to the upper notch, on up the stem, along the head serif
// and down the left-hand edge to the foot.
func p(p design.Params) (outline.Path, float64) {
	stem := descenderStem(p, design.PStemLeft(p), p.XHeight, true)
	bowl := pBowl(p)

	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, bowl.NotchLower)
	bowlOuter(dBowlShape, bowl).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	bowlCounter(dBowlShape, bowl).addContour(&path)
	return path, design.PAdvance(p)
}
