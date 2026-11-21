// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// degSuperness is the squareness of the outer edge of the degree ring.
const degSuperness = 0.7071

// degree draws the degree sign: a small ring of even thickness (DegThick, a
// little above the thin part of the round symbols) with its top at DegTop, 16
// above the cap height; its diameter is 0.42 of the cap height.  The ring is
// round, not a small o.  The outer edge is a circle (degSuperness) and the
// counter a hundredth squarer.  The extremes of the ring are on the integer
// grid.
//
// Side kinds: open on both sides, DegSideLeft and DegSideRight to the ring.
func degree(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0, top := s.DegSideLeft, s.DegTop
	outer, counter := ringChains(x0, top-s.DegDiam, x0+s.DegDiam, top, s.DegThick, s.DegThick, degSuperness, degSuperness+0.01)
	return chainsPath(outer, counter), s.DegSideLeft + s.DegDiam + s.DegSideRight
}
