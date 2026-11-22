// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// at draws the at sign: a round a in a ring.  The a is a bowl on a stem, as d
// without its ascender: the stem AtStem thick stands at the right with a flat
// top at the top of the bowl, and the bowl is one stroke (symstroke.go),
// AtBowlThick at the left extreme and Thin at the top and the bottom, which
// leaves the stem near its top and meets it again near its foot; its counter is
// bounded by the stem.  The stem and the ring are one stroke: from the top of
// the stem down to its foot, where it turns to the right in a small hook and
// rises into the ring, AtRing thick, which sweeps up the right, over the top,
// down the left and round the bottom and ends at the lower right in a cut,
// below and beyond the hook.  The outline is the union of the two strokes
// (symunion.go).
//
// Side kinds: open, AtSide to the ring.
func at(p design.Params) (outline.Path, float64) {
	s := p.Sym
	w, top, bottom := s.AtInk, s.AtTop, s.AtBottom
	x0, x1 := s.AtSide, s.AtSide+w
	h := top - bottom
	cx, cy := (x0+x1)/2, (top+bottom)/2
	t, rt := p.Figures.Thin, s.AtRing
	ov := p.OvershootLower
	bt, bb := s.AtBowlTop, s.AtBowlBottom
	bh := bt - bb
	// The a is centred a little left of the ring.
	bx0 := cx - 0.03*w - (s.AtBowlW+s.AtStem*0.55)/2
	stemL := bx0 + s.AtBowlW - 0.45*s.AtStem
	sx := stemL + s.AtStem/2
	stemR := stemL + s.AtStem
	outer := newSuperShape(bx0, bb-ov, stemR-4, bt+ov, p.SupernessOuter)
	// The flat top of the stem lies where the outer edge of the bowl crosses
	// the left edge of the stem, so that the two meet in a corner.
	stemTop := bt - ov
	for i := 0; i < 4; i++ {
		c := outer.cubic(i, true)
		if (c.P0.X-stemL)*(c.P3.X-stemL) <= 0 && c.P0.Y+c.P3.Y > bt+bb {
			stemTop = c.at(c.tAtX(stemL)).Y
		}
	}
	stroke := []station{
		{X: sx, Y: stemTop, Dir: -90, W: s.AtStem, Flat: true},
		{X: sx, Y: bb + 0.22*bh, Dir: -90, W: s.AtStem},
		{X: sx + 0.95*s.AtStem, Y: bb - ov + 0.6*t, Dir: 0, W: 1.2 * t},
		{X: x1 - rt/2, Y: cy + 0.08*h, Dir: 90, W: rt},
		{X: cx, Y: top - rt/2, Dir: 180, W: rt},
		{X: x0 + rt/2, Y: cy, Dir: -90, W: rt},
		{X: cx + 0.03*w, Y: bottom + rt/2, Dir: 0, W: rt},
		{X: x1 - 0.08*w, Y: bottom + 0.09*h, Dir: 32, W: 0.9 * rt},
	}
	// The bowl: the two super-ellipses of o, the counter AtBowlThick inside the
	// outer edge at the left and Thin at the top and the bottom; at the right both
	// end inside the stem.
	inner := newSuperShape(bx0+s.AtBowlThick, bb-ov+t, stemL+16, bt+ov-t, p.SupernessInner)
	var bo, bc chain
	for i := 0; i < 4; i++ {
		bo = append(bo, curveSeg(outer.cubic(i, true)))
		bc = append(bc, curveSeg(inner.cubic(i, false)))
	}
	return chainsPath(union(strokeChain(stroke), bo, bc)...), x1 + s.AtSide
}
