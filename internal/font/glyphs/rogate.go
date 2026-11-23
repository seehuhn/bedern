// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The o with a rogate, in both cases.
//
// The letter is the O or the o with a straight upright stroke above and
// below it, on the middle of the letter: each stroke leaves the thin part of
// the round at its top or its bottom and reaches rogateReach beyond the
// letter.  The strokes have the thin stroke of the round capitals
// (Caps.RoundThin) on O and the hairline (Hair) on o.  The glyph is the
// union of the letter and the strokes and has the advance of the letter.

// rogateReach is the length of a stroke of the rogate beyond the letter.
const rogateReach = 153.0

// rogate returns the letter with the strokes of the rogate, thick wide; thin
// is the thickness of the round of the letter at its top and its bottom.
func rogate(p design.Params, letter string, thick, thin float64) (outline.Path, float64) {
	path, adv, _ := Outline(letter, p)
	x0, y0, x1, y1 := path.BBox()
	cx := (x0 + x1) / 2
	parts := append(pathChains(path),
		rectChain(cx-thick/2, y1-thin/2, cx+thick/2, y1+rogateReach),
		rectChain(cx-thick/2, y0-rogateReach, cx+thick/2, y0+thin/2))
	return chainsPath(union(parts...)...), adv
}

func capORogate(p design.Params) (outline.Path, float64) {
	return rogate(p, "O", p.Caps.RoundThin, p.Caps.RoundThin)
}

func oRogate(p design.Params) (outline.Path, float64) {
	return rogate(p, "o", p.Hair, p.ThinLower)
}
