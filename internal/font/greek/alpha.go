// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['α'] = alpha }

// α: one stroke which crosses itself on the right, as α is written:
// from the top right down to the left through the crossing, round the
// bowl, over the top and down to the right through the crossing again,
// and at the baseline into a short exit.  The falling line passes the
// crossing nearly upright, the rising line carries the slant.
func alpha(c *ctx) letter {
	pen := c.pen
	x, o := c.x, c.o
	h := x + 2*o        // the height of α
	left := 25.0        // the left of the ink
	cx := left + 0.65*h // the crossing
	cy := -o + 0.45*h   //
	const leanF = -8.0  // the lean of the falling line at the crossing, top to the left
	const leanR = 24.0  // the lean of the rising line, top to the right
	yt := x + o - penWidth(pen, 0)/2
	yb := -o + penWidth(pen, 180)/2
	lx := left + penWidth(pen, 90)/2 // the left of the centre line of the bowl
	// The bowl: half an ellipse on the left, its right half drawn by
	// the lines into the crossing.
	a := (cx - lx) * 0.48
	bx, by := lx+a, (yt+yb)/2
	b := (yt - yb) / 2
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dirF, dirR := -90-leanF, -90-leanR // travelling down
	cross1 := stroke.Node{P: pt(cx, cy), Dir: dirR, K: 0, W: auto, DW: auto}
	cross2 := stroke.Node{P: pt(cx, cy), Dir: dirF, K: 0, W: auto, DW: auto}
	// Below the crossing the falling line bends at once into the exit,
	// along a quarter ellipse which reaches the baseline d to the right.
	d := 100.0
	ex0, ey0, eb := cx+d, cy, cy-yb
	top := n(cx+(x+2-cy)*math.Tan(rad(leanR))-14*leanR/22, x+2, -90-leanR*10/22, auto)
	nodes := []stroke.Node{
		top,
		cross1,
		ellipseNode(bx, by, a, b, 180, false),
		ellipseNode(bx, by, a, b, 90, false),
		ellipseNode(bx, by, a, b, 0, false),
		cross2,
		ellipseNode(ex0, ey0, d, eb, -45, true),
		ellipseNode(ex0, ey0, d, eb, 55, true),
	}
	// The exit ends in the turn.
	nodes[len(nodes)-1].DW = 0
	ex := nodes[len(nodes)-1].P
	return letter{adv: ex.X + penWidth(pen, 55)/2 + 30, strokes: []stroke.Stroke{
		{Nodes: nodes, Start: cut, End: cut},
	}}
}
