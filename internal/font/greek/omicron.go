// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['ο'] = omicron }

// omicronInk is the width of the ink of the bowl of ο, σ and δ; omicronTilt
// the turn of its axes (degrees, anticlockwise), which carries the top to
// the right.
const (
	omicronInk  = 350.0
	omicronTilt = -7.0
)

// omicronBowl returns the centre of the bowl and its half axes on the centre
// line.
func omicronBowl(c *ctx) (cx, cy, a, b float64) {
	pen := c.pen
	left := 20.0
	a = (omicronInk - penWidth(pen, 90)) / 2
	b = (c.x + 2*c.o - penWidth(pen, 180)) / 2
	return left + omicronInk/2, c.x / 2, a, b
}

// tiltAt returns the node of the centre line on the ellipse about (cx, cy)
// with the half axes a and b, turned by th degrees, at the angle phi of its
// parameter, travelling anticlockwise if ccw.
func tiltAt(cx, cy, a, b, th, phi float64, ccw bool) stroke.Node {
	nd := ellipseAt(0, 0, a, b, phi, ccw)
	s, c := math.Sin(th*math.Pi/180), math.Cos(th*math.Pi/180)
	nd.P = pt(cx+nd.P.X*c-nd.P.Y*s, cy+nd.P.X*s+nd.P.Y*c)
	nd.Dir += th
	return nd
}

// ο: one stroke round a tilted ellipse, from the top anticlockwise.
func omicron(c *ctx) letter {
	cx, cy, a, b := omicronBowl(c)
	var nodes []stroke.Node
	for _, phi := range []float64{125, 180, 270, 360, 450, 485} {
		nodes = append(nodes, tiltAt(cx, cy, a, b, omicronTilt, phi, true))
	}
	return letter{adv: 20 + omicronInk + 20, strokes: []stroke.Stroke{{Nodes: nodes, Start: cut, End: cut}}}
}
