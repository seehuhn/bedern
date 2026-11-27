// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// footA returns the nodes of the foot of a stem on the line x = sx which
// bends at the height ey0 into the exit of ι: a quarter ellipse reaching
// the baseline d to the right, ending in a turn which tapers.  The first
// node is the straight stem at ey0, the last the end of the exit; the
// exit's end x is the abscissa of the last node.
func footA(c *ctx, sx, ey0, d float64) []stroke.Node {
	yb := -c.o + penWidth(c.pen, 180)/2
	eb := ey0 - yb
	ex0 := sx + d
	nodes := []stroke.Node{
		ellipseAt(ex0, ey0, d, eb, 200, true),
		ellipseNode(ex0, ey0, d, eb, -40, true),
		ellipseNode(ex0, ey0, d, eb, 55, true),
	}
	nodes[len(nodes)-1].DW = 0
	return nodes
}

// advA returns the advance of a letter whose ink ends with the exit node ex.
func advA(c *ctx, ex stroke.Node) float64 {
	return ex.P.X + penWidth(c.pen, 55)/2 + 30
}

// leadA returns a straight node t before the node nd along its tangent.
func leadA(nd stroke.Node, t float64) stroke.Node {
	r := nd.Dir * math.Pi / 180
	return s(nd.P.X-t*math.Cos(r), nd.P.Y-t*math.Sin(r), nd.Dir, auto)
}

// barA returns the nodes of a bar of height yt (top of the ink) whose left
// end curls down, from x = left to the right end xr: one smooth stroke.
func barA(c *ctx, left, xr, yt float64) []stroke.Node {
	return []stroke.Node{
		n(left+12, yt-48, 30, auto),
		n(left+150, yt, 2, auto),
		e(xr, yt, 0, auto),
	}
}

// ellipseCentreA returns the centre of the ellipse with the half axes a and
// b on which the node at p travels in the direction dir.
func ellipseCentreA(px, py, a, b, dir float64, ccw bool) (float64, float64) {
	q := ellipseNode(0, 0, a, b, dir, ccw)
	return px - q.P.X, py - q.P.Y
}

// seedA returns the first node of a stroke which starts inside another: a
// node at p travelling in the direction dir with a thin end, so that the
// corners of its cut do not show.
func seedA(nd stroke.Node, t float64) stroke.Node {
	l := leadA(nd, t)
	l.W = 16
	return l
}

// topA is the end of a stem at the x-height: a soft dome.
var topA = stroke.End{Kind: stroke.Cut}

// stemTopA returns the top node of a stem whose foot is at sx, leaning
// by lean degrees (the top to the right), with its dome at the x-height.
func stemTopA(c *ctx, sx, lean float64) stroke.Node {
	w := penWidth(c.pen, -90-lean)
	_ = w
	return e(sx+lean*4.5, c.x, -90-lean, auto)
}
