// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// pictureList lists the arrows, the radical and the centred asterisk of
// text, in glyph order.
var pictureList = []entry{
	{"arrowleft", '←', sideMath, sideMath, arrowleft},
	{"arrowright", '→', sideMath, sideMath, arrowright},
	{"arrowup", '↑', sideOpen, sideOpen, arrowup},
	{"arrowdown", '↓', sideOpen, sideOpen, arrowdown},
	{"radical", '√', sideOpen, sideOpen, radical},
	{"asteriskmath", '∗', sideAsterisk, sideAsterisk, asteriskmath},
}

// signList lists the musical note, the signs of genealogy and the leaf, in
// glyph order.
var signList = []entry{
	{"musicalnote", '♪', sideOpen, sideOpen, musicalnote},
	{"born", privateUseRune("born"), sideOpen, sideOpen, born},
	{"died", privateUseRune("died"), sideOpen, sideOpen, died},
	{"uni26AD", '⚭', sideOpen, sideOpen, married},
	{"uni26AE", '⚮', sideOpen, sideOpen, divorced},
	{"leaf", privateUseRune("leaf"), sideOpen, sideOpen, leaf},
}

// asteriskmath draws the centred asterisk: the asterisk with its middle on
// the maths axis, with the side spaces of the asterisk.
func asteriskmath(p design.Params) (outline.Path, float64) {
	q := p.Punct
	dy := math.Round(q.MathAxis - (q.AsteriskTop - q.AsteriskSize/2))
	return placeInk(shiftInk(asteriskPath(p), 0, dy), q.Asterisk)
}
