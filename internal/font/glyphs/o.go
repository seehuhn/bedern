// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// o draws the lowercase letter o: a bowl with vertical stress, thick at
// the sides and thin at the top and bottom.
func o(p design.Params) (outline.Path, float64) {
	left := design.OBowlLeft(p)
	right := design.OBowlRight(p)
	bottom := -p.OvershootLower
	top := p.XHeight + p.OvershootLower

	var path outline.Path
	superEllipse(&path, left, bottom, right, top, p.SupernessOuter, true)
	superEllipse(&path,
		left+p.CurveLower, bottom+p.ThinLower,
		right-p.CurveLower, top-p.ThinLower,
		p.SupernessInner, false)
	return path, design.OAdvance(p)
}
