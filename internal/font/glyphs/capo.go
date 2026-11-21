// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capO draws the capital O as the o is drawn: two super-ellipses, the outer
// one counter-clockwise and the counter clockwise, with the squareness of the
// o (SupernessOuter and SupernessInner).  The O is OWidth wide and reaches
// OvershootUpper above the cap height and below the baseline; its sides are
// Round thick and its top and bottom RoundThin.
func capO(p design.Params) (outline.Path, float64) {
	c := p.Caps
	left := c.SideRoundLeft
	right := left + c.OWidth
	bottom := -p.OvershootUpper
	top := p.CapHeight + p.OvershootUpper

	var path outline.Path
	superEllipse(&path, left, bottom, right, top, p.SupernessOuter, true)
	superEllipse(&path,
		left+c.Round, bottom+c.RoundThin,
		right-c.Round, top-c.RoundThin,
		p.SupernessInner, false)
	return path, design.CapOAdvance(p)
}
