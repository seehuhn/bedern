// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['ε'] = epsilon }

// ε: two strokes.  The upper runs from the top
// right over the top and down the left into the waist, along which it
// passes to the right and ends in a small upturn.  The lower begins inside
// that bar, falls to the left, rounds the bottom and ends at the lower
// right.  The bar is the top of the lower counter.
func epsilon(c *ctx) letter {
	pen := c.pen
	yt := c.x + c.o - penWidth(pen, 0)/2
	yb := -c.o + penWidth(pen, 180)/2
	lx := 20 + penWidth(pen, 90)/2
	ua := 150.0
	ucy := 322.0
	ub := yt - ucy
	ucx := lx + ua
	la := 175.0
	lcx := lx + la
	lb := 108.0
	lcy := yb + lb
	tip := n(ucx+92, ucy-88, 28, auto)
	tip.W = 46
	up := []stroke.Node{
		ellipseAt(ucx, ucy, ua, ub, 32, true),
		ellipseNode(ucx, ucy, ua, ub, 180, true),
		ellipseNode(ucx, ucy, ua, ub, -90, true),
		ellipseNode(ucx, ucy, ua, ub, -40, true),
		ellipseNode(ucx, ucy, ua, ub, -10, true),
		tip,
	}
	low := []stroke.Node{
		ellipseAt(lcx, lcy, la, lb, 124, true),
		ellipseNode(lcx, lcy, la, lb, -90, true),
		ellipseNode(lcx, lcy, la, lb, -45, true),
		ellipseNode(lcx, lcy, la, lb, 0, true),
		ellipseAt(lcx, lcy, la, lb, 337, true),
	}
	endV := cut
	return letter{adv: lcx + la*0.92 + 20, strokes: []stroke.Stroke{
		{Nodes: up, Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: 78}, End: cut},
		{Nodes: low, Start: hidden, End: endV},
	}}
}
