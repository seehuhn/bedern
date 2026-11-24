// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// d draws the d: the bowl and the stem of a, with the stem carried up to the
// ascender into the entry of l.
func d(p Params) letter {
	e := p.bowlStemEdge()
	return letter{
		strokes: []stroke.Stroke{p.bowl(e), p.stem(e, p.XHeight+p.AscenderEntryRise, true)},
		advance: e + p.Stem + p.SideStemRight,
	}
}
