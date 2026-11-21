// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capI draws the capital I: the capital stem with its four serifs, standing
// between the stem sides.
func capI(p design.Params) (outline.Path, float64) {
	j := p.Caps.SerifJut
	s := capStem{Left: design.CapStemLeft(p), FootLeft: j, FootRight: j, HeadLeft: j, HeadRight: j}
	up := s.up(p, s.start(), p.CapHeight/2)
	top := s.top(p, up.end(), p.CapHeight/2)
	return capContour(join(up, top, s.down(p, top.end()))), design.CapIAdvance(p)
}
