// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['φ'] = phi }

// φ: one stroke, as it is written: from the flag at the top, left round the
// bowl, over the arch at the top right and down as the stem, which crosses
// the bottom of the bowl and descends to a soft end.
func phi(c *ctx) letter {
	pen := c.pen
	ink := 485.0
	left := 20.0
	const tilt = -12.0
	a := (ink - penWidth(pen, 90)) / 2
	b := (c.x + 2*c.o - penWidth(pen, 180)) / 2
	cx, cy := left+ink/2, c.x/2
	stemX := cx - 10
	var nodes []stroke.Node
	for _, phi := range []float64{122, 180, 270, 360} {
		nodes = append(nodes, tiltAt(cx, cy, a, b, tilt, phi, true))
	}
	top := c.x + c.o - penWidth(pen, 180)/2
	nodes = append(nodes,
		n(cx+100, top, 180, auto),
		n(stemX+14, 310, -97, auto),
		n(stemX+1, 70, -91, auto),
		e(stemX-3, c.desc, -87, 52),
	)
	return letter{adv: left + ink + 20, strokes: []stroke.Stroke{
		{Nodes: nodes,
			Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: -55},
			End:   stroke.End{Kind: stroke.Cut, Line: true, Angle: 18}},
	}}
}
