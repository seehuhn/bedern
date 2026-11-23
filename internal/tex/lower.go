// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// capitalAccents names the glyphs which stand at their real height over a
// capital and which TeX sets as accents: the capital forms of the accents,
// slots 0 to 10 of TS1, and the two ties of the capitals, slots 27 and 29.
var capitalAccents = func() map[string]bool {
	m := map[string]bool{"tieaccentcapital": true, "tieaccentcapital.new": true}
	for c := 0; c <= 10; c++ {
		m[TS1.Slots[c].Name] = true
	}
	return m
}()

// ForTeX returns the glyphs gs as the outline files and the metric files of
// TeX need them.  The accent mechanism of TeX raises an accent over a letter
// by the height of the letter less the x-height of the font, so the capital
// forms of the accents and the ties of the capitals, which stand at their real
// height over a capital, are lowered by the difference of the cap height and
// the x-height.  The other glyphs and the glyphs gs themselves stay as they
// are, since the OpenType font sets the marks by their anchors.
func ForTeX(p design.Params, gs []glyphs.Glyph) []glyphs.Glyph {
	drop := p.CapHeight - p.XHeight
	out := make([]glyphs.Glyph, len(gs))
	copy(out, gs)
	for i, g := range out {
		if !capitalAccents[g.Name] {
			continue
		}
		path := make(outline.Path, len(g.Path))
		for k, c := range g.Path {
			d := outline.Contour{Start: lowered(c.Start, drop), Segs: make([]outline.Segment, len(c.Segs))}
			for j, s := range c.Segs {
				d.Segs[j] = outline.Segment{
					IsCurve: s.IsCurve,
					P1:      lowered(s.P1, drop), P2: lowered(s.P2, drop), P3: lowered(s.P3, drop),
				}
			}
			path[k] = d
		}
		out[i].Path = path
	}
	return out
}

// lowered returns the point moved down by drop.
func lowered(pt outline.Point, drop float64) outline.Point {
	return outline.Point{X: pt.X, Y: pt.Y - drop}
}
