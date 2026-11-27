// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

func init() { draw['β'] = beta }

// β: the stem from the descender up into the upper bowl, which ends at the
// waist in a hook towards the stem without reaching it, and the lower bowl,
// begun at the waist and closed on the stem at the baseline.  The stem leans
// a little to the right, and is thicker at the foot, where it ends in a soft
// cut.
func beta(c *ctx) letter {
	pen := c.pen
	yt := c.asc - penWidth(pen, 0)/2
	yb := -c.o + penWidth(pen, 180)/2
	sx := 20 + penWidth(pen, 90)/2 // the centre line of the stem
	// the upper bowl
	ua := 150.0
	wy := 352.0 // the height of the hook at the waist
	ub := (yt - (wy - 14)) / 2
	ucy := yt - ub
	ucx := sx + 8 + ua
	// the lower bowl
	la := 178.0
	ltop := wy + 42
	lb := (ltop - yb) / 2
	lcy := yb + lb
	lcx := sx + 2 + la
	foot := n(sx-6, c.desc+6, 90, 66)
	tipw := n(152, wy+12, 174, auto)
	tipw.W = 60
	up := []stroke.Node{
		foot,
		n(sx-1, 40, 90, auto),
		ellipseNode(ucx, ucy, ua, ub, 90, false),
		ellipseNode(ucx, ucy, ua, ub, 0, false),
		ellipseNode(ucx, ucy, ua, ub, -90, false),
		ellipseNode(ucx, ucy, ua, ub, -130, false),
		tipw,
	}
	low := []stroke.Node{
		ellipseAt(lcx, lcy, la, lb, 62, false),
		ellipseNode(lcx, lcy, la, lb, -90, false),
		ellipseNode(lcx, lcy, la, lb, 180, false),
		ellipseAt(lcx, lcy, la, lb, 206, false),
	}
	return letter{adv: lcx + la + penWidth(pen, 90)/2 + 20, strokes: []stroke.Stroke{
		{Nodes: up, Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: 6}, End: cut},
		{Nodes: low, Start: hidden, End: hidden},
	}}
}
