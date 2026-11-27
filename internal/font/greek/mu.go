// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['μ'] = mu }

// μ: a stem which descends into a soft taper, a bowl from it to the
// right stem, and the right stem with the exit of ι.
func mu(c *ctx) letter {
	sx := 30 + penWidth(c.pen, -90)/2
	rx := 335.0
	yb := -c.o + penWidth(c.pen, 180)/2
	left := stroke.Stroke{Nodes: []stroke.Node{
		stemTopA(c, sx, 1.5),
		n(sx-8, c.desc+10, -93, 36),
	}, Start: topA, End: stroke.End{Kind: stroke.Cut, Line: true, Angle: 8}}
	cy := 190.0
	a := (rx + 6 - sx) / 1.966
	b := cy - yb
	p := ellipseAt(sx+a, cy, a, b, 205, true)
	bowl := stroke.Stroke{Nodes: []stroke.Node{
		seedA(p, 22), p,
		ellipseNode(sx+a, cy, a, b, 0, true),
		ellipseAt(sx+a, cy, a, b, 345, true),
	}, Start: hidden, End: hidden}
	nodes := append([]stroke.Node{stemTopA(c, rx, 0)}, footA(c, rx, 120, 72)...)
	return letter{adv: advA(c, nodes[len(nodes)-1]), strokes: []stroke.Stroke{
		left, bowl, {Nodes: nodes, Start: topA, End: cut},
	}}
}
