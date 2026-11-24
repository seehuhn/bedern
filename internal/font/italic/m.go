// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// m draws the m: the n with a second arch.  The entry and the left stem
// end in a plain foot, the first arch runs into the middle stem, which ends
// in a plain foot, and the second arch runs into the right stem, which ends
// in the exit; both arches are the arch of n between stems MCounter apart.
func m(p Params) letter {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.MCounter
	e3 := e2 + p.Stem + p.MCounter
	return letter{
		strokes: []stroke.Stroke{p.stem(e1, p.XHeight, false), p.arch(e1, e2, false), p.arch(e2, e3, true)},
		advance: e3 + p.Stem + p.SideStemRight,
	}
}
