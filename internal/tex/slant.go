// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// SlantItalics sets the italic corrections of m for the glyphs gs as they
// stand when the outlines are slanted by m.Slant, a ratio: the extent of the
// slanted outline beyond the advance.  The widths, the heights, the depths and
// the kerns do not change.
func (m *Metrics) SlantItalics(gs []glyphs.Glyph) {
	for _, ch := range m.Chars {
		if ch == nil {
			continue
		}
		g := gs[ch.Glyph]
		ch.Italic = 0
		if len(g.Path) == 0 {
			continue
		}
		_, _, x1, _ := shear(g.Path, m.Slant).BBox()
		ch.Italic = math.Max(0, x1-g.Advance)
	}
}

// shear returns the path with every point moved to the right by the slant
// times its height.
func shear(p outline.Path, slant float64) outline.Path {
	pt := func(q outline.Point) outline.Point { return outline.Point{X: q.X + slant*q.Y, Y: q.Y} }
	out := make(outline.Path, len(p))
	for i, c := range p {
		out[i].Start = pt(c.Start)
		for _, s := range c.Segs {
			out[i].Segs = append(out[i].Segs, outline.Segment{
				IsCurve: s.IsCurve, P1: pt(s.P1), P2: pt(s.P2), P3: pt(s.P3),
			})
		}
	}
	return out
}
