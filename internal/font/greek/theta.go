// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['θ'] = theta }

// θ: a tall closed oval, drawn as ο, and a bar across it.
func theta(c *ctx) letter {
	pen := c.pen
	ink := 360.0
	left := 20.0
	a := (ink - penWidth(pen, 90)) / 2
	b := (c.asc + 2*c.o - penWidth(pen, 180)) / 2
	cx, cy := left+ink/2, c.asc/2
	var bowl []stroke.Node
	for _, phi := range []float64{125, 150, 180, 225, 270, 315, 360, 405, 450, 485} {
		bowl = append(bowl, tiltAt(cx, cy, a, b, -3, phi, true))
	}
	bar := []stroke.Node{
		e(left+20, cy+4, -3, 56),
		e(left+ink-20, cy-8, -3, 56),
	}
	bar[0].Dir, bar[1].Dir = -3, -3
	return letter{adv: left + ink + 20, strokes: []stroke.Stroke{
		{Nodes: bowl, Start: cut, End: cut},
		{Nodes: bar, Start: hidden, End: hidden},
	}}
}
