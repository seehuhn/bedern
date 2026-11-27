// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['η'] = eta }

// η: a stem with a beak at the top left, an arch from the stem to the
// right leg, which descends below the baseline into a soft taper.
func eta(c *ctx) letter {
	sx := 110.0
	yt := c.x + c.o - penWidth(c.pen, 0)/2
	a, b := 80.0, 120.0
	cx1, cy1 := sx-a, yt-b
	first := ellipseNode(cx1, cy1, a, b, 30, false)
	ent := []stroke.Node{
		first,
		ellipseNode(cx1, cy1, a, b, 0, false),
		ellipseNode(cx1, cy1, a, b, -90, false),
		ne(sx+3, 0, -91, auto),
	}
	stem := stroke.Stroke{Nodes: ent, Start: cut, End: cut}
	rx := 390.0
	x0 := sx + 4
	cy := 215.0
	aa, bb := (rx-x0)/2, yt-cy
	p := ellipseAt(x0+aa, cy, aa, bb, 190, false)
	arch := stroke.Stroke{Nodes: []stroke.Node{
		seedA(p, 12), p,
		ellipseNode(x0+aa, cy, aa, bb, 0, false),
		ellipseNode(x0+aa, cy, aa, bb, -90, false),
		n(rx-34, c.desc+8, -101, 36),
	}, Start: hidden, End: stroke.End{Kind: stroke.Cut, Line: true, Angle: 8}}
	return letter{adv: rx + penWidth(c.pen, -90)/2 + 30, strokes: []stroke.Stroke{stem, arch}}
}
