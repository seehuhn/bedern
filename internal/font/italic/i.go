// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// i draws the i: one stroke from the entry down the stem into the exit, and
// the dot, a circle whose centre lies on the centre line of the stem.
func i(p Params) letter {
	e := p.SideStemLeft
	return letter{
		strokes: []stroke.Stroke{p.stem(e, p.XHeight, true)},
		dots:    []dot{{C: pt(e+p.Stem/2, p.DotCentre), D: p.Dot}},
		advance: e + p.Stem + p.SideStemRight,
	}
}
