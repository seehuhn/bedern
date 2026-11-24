// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// n draws the n: the entry and the left stem down to a plain foot, and the
// arch from the left stem into the right stem, which ends in the exit.
func n(p Params) letter {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.NCounter
	return letter{
		strokes: []stroke.Stroke{p.stem(e1, p.XHeight, false), p.arch(e1, e2, true)},
		advance: e2 + p.Stem + p.SideStemRight,
	}
}
