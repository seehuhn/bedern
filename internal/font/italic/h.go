// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// h draws the h: the ascender with the entry at its top, the stem down to a
// plain foot, and the arch and right stem of n with the exit.  Its left side
// is HLeftLess less than that of n.
func h(p Params) letter {
	e1 := p.SideStemLeft - p.HLeftLess
	e2 := e1 + p.Stem + p.NCounter
	return letter{
		strokes: []stroke.Stroke{p.stem(e1, p.XHeight+p.AscenderEntryRise, false), p.arch(e1, e2, true)},
		advance: e2 + p.Stem + p.SideStemRight,
	}
}
