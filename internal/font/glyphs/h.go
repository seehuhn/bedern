// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// h draws the lowercase letter h: the n with its left stem carried up to the
// ascender height, where the stem has the head serif of l.  The arch is that
// of n.  The left stem is drawn in two pieces, up to the join and from the
// notch to the top, as doc.go describes for a stem which an arch joins
// part-way up.
func h(p design.Params) (outline.Path, float64) {
	left := stemSpec{
		Left: p.SideStemLeft + design.AscenderInset(p), Width: p.AscenderStem, Top: p.Ascender,
		FootLeft: p.SerifJut, FootRight: p.SerifJutInner,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	return archWithStem(p, left)
}
