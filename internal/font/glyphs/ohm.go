// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ohm sign Ω and the mho sign ℧.
//
// The ohm sign is an open round of the weights of O which stands on two
// feet: its sides are the round of O, its top is the thin of O, and its ends
// run inwards and down as thin legs to the feet.  Each foot is a bar with the
// serif of the bottom arm of E rising at its outer end.  The mho sign is the
// ohm sign turned through half a turn.

// Shape constants of the ohm sign.
const (
	ohmInk    = 0.969 // the ink width, in ink widths of H
	ohmRound  = 0.913 // the round at its widest, in ink widths
	ohmFoot   = 0.380 // the length of a foot, in ink widths
	ohmGap    = 0.175 // the white between the legs at the feet, in ink widths
	ohmBar    = 1.55  // the thickness of a foot, in bars of the arms of E
	ohmMiddle = 0.56  // the height of the widest place of the round, in cap heights
	ohmLean   = 10.0  // the legs where they meet the feet, degrees from the vertical
)

// ohmOutline returns the ink of the ohm sign with its leftmost ink at the
// abscissa 0.
func ohmOutline(p design.Params) outline.Path {
	c := p.Caps
	w := math.Round(ohmInk * design.CapHInk(p))
	cx := w / 2
	round := ohmRound * w
	foot := math.Round(ohmFoot * w)
	t := math.Round(ohmBar * c.Bar)
	thin := c.RoundThin
	top := p.CapHeight + p.OvershootUpper
	mid := ohmMiddle * p.CapHeight
	legX := ohmGap*w/2 + thin/2
	// The left half of the round, from the foot of the leg to the top; the
	// right half is its mirror image.
	in, out := strokeOpen([]station{
		{X: cx - legX, Y: t / 2, Dir: 90 + ohmLean, W: thin},
		{X: cx - round/2 + c.Round/2, Y: mid, Dir: 90, W: c.Round},
		{X: cx, Y: top - thin/2, Dir: 0, W: thin},
	})
	inR, outR := in.mirrorX(cx), out.mirrorX(cx)
	stroke := join(in, inR, lineChain(inR.end(), outR.start()), outR, out, lineChain(out.end(), in.start()))
	parts := []chain{stroke}
	for _, f := range pathChains(CapArmOutline(p, math.Min(w-foot, cx+legX-thin), w, 0, t, 0, c.BottomRise)) {
		parts = append(parts, f, f.mirrorX(cx))
	}
	return chainsPath(union(parts...)...)
}

// ohm draws Ω, with the side spaces of O.
func ohm(p design.Params) (outline.Path, float64) {
	c := p.Caps
	return placeInk(ohmOutline(p), design.PunctSides{Left: c.SideRoundLeft, Right: c.SideRoundRight})
}

// mho draws ℧: the ohm sign turned through half a turn about the middle of
// the cap height, in the advance of the ohm sign.
func mho(p design.Params) (outline.Path, float64) {
	path, adv := ohm(p)
	return turned(path, adv/2, p.CapHeight/2), adv
}
