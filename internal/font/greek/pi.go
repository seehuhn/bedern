// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['π'] = pi_ }

// π: a bar with arms, a left leg which slants and a right leg with the
// exit of ι.
func pi_(c *ctx) letter {
	yt := c.x + c.o - penWidth(c.pen, 0)/2
	rx := 395.0
	bar := stroke.Stroke{Nodes: barA(c, 25, 545, yt), Start: cut, End: cut}
	left := stroke.Stroke{Nodes: []stroke.Node{
		n(212, yt-30, -94, auto), n(112, 0, -114, auto),
	}, Start: hidden, End: stroke.End{Kind: stroke.Cut, Line: true, Angle: 0}}
	nodes := append([]stroke.Node{e(rx, yt-20, -90, auto)}, footA(c, rx, 120, 72)...)
	return letter{adv: 555 + 30, strokes: []stroke.Stroke{
		bar, left, {Nodes: nodes, Start: hidden, End: cut},
	}}
}
