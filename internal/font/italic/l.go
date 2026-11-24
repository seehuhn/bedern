// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// l draws the l: one stroke from the entry at the ascender down the stem
// into the exit.
func l(p Params) letter {
	e := p.SideStemLeft
	return letter{
		strokes: []stroke.Stroke{p.stem(e, p.XHeight+p.AscenderEntryRise, true)},
		advance: e + p.Stem + p.SideStemRight,
	}
}
