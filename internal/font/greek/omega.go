// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ω'] = omega }

// ω: two half-bowls which balance each other.  The left one comes down from
// the top left, round the bottom and up into the middle stem; the right one
// begins at the top of that stem, goes down it, round the bottom and up to
// the top right.  The stem leans a little to the right at its top.
func omega(c *ctx) letter {
	pen := c.pen
	x, o := c.x, c.o
	left := 25.0
	yb := -o + penWidth(pen, 180)/2
	lx := left + penWidth(pen, 90)/2
	a, b := 126.0, 192.0
	by := yb + b
	cx := lx + 1.82*a - 4 // the middle stem, where the half-bowls meet
	lbx := lx + a
	rbx := cx - 8 + a
	reach, dir := 78.0, 114.0 // the arms at the top
	yp := 0.72 * x            // the top of the middle stem
	right := []stroke.Node{
		n(cx+12, yp, -100, 80),
		ellipseNode(rbx, by, a, b, -90, true),
		ellipseNode(rbx, by, a, b, 0, true),
		ellipseNode(rbx, by, a, b, 90, true),
		n(rbx+a-reach, x+o-14, dir, auto),
	}
	right[len(right)-1].DW = 0
	leftH := []stroke.Node{
		n(lx+reach+16, x+o-14, -dir, auto),
		ellipseNode(lbx, by, a, b, -90, true),
		ellipseNode(lbx, by, a, b, 0, true),
		ellipseAt(lbx, by, a, b, 325, true),
	}
	leftH[0].DW = 0
	rx := rbx + a
	return letter{adv: rx + penWidth(pen, 90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: leftH, Start: cut, End: hidden},
		{Nodes: right, Start: cut, End: cut},
	}}
}
