// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['γ'] = gamma }

// γ: a stem from the entry hook at the top left through the junction to a
// tail below the baseline, and an arm from the top right which joins it.
func gamma(c *ctx) letter {
	pen := c.pen
	top := c.x + c.o - penWidth(pen, 0)/2
	stem := dgStem(c, 243, c.desc+30, 8)
	stem[len(stem)-1].W = 46
	jy := 100.0
	jx := centreX(stem, jy)
	arm := dgBranch(465, top-12, jx-465, jy-(top-12), -123, -100)
	arm[0].DW = 0
	arm[2].W = 34
	ex := arm[0].P
	return letter{adv: ex.X + penWidth(pen, -123)/2 + 25, strokes: []stroke.Stroke{
		{Nodes: stem, Start: cut, End: cut},
		{Nodes: arm, Start: cut, End: hidden},
	}}
}
