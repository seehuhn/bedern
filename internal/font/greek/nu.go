// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

func init() { draw['ν'] = nu }

// dgArc returns the nodes of an arc of an ellipse which leaves the point
// (vx, vy) travelling in the direction d0 and ends travelling in the direction
// d1 at the point (vx+dx, vy+dy) (the ellipse is chosen to fit the offset as
// well as it can, and then moved so that the end is exact).
func dgArc(vx, vy, dx, dy, d0, d1 float64, ccw bool) []stroke.Node {
	best, ba, bb := math.Inf(1), 0.0, 0.0
	for a := 50.0; a < 2500; a *= 1.02 {
		for b := 50.0; b < 2500; b *= 1.02 {
			o0 := ellipseNode(0, 0, a, b, d0, ccw).P
			o1 := ellipseNode(0, 0, a, b, d1, ccw).P
			err := math.Hypot(o1.X-o0.X-dx, o1.Y-o0.Y-dy)
			if err < best {
				best, ba, bb = err, a, b
			}
		}
	}
	cx, cy := dgCentre(ba, bb, d0, ccw, vx, vy)
	e1 := ellipseNode(cx, cy, ba, bb, d1, ccw).P
	cx += vx + dx - e1.X
	cy += vy + dy - e1.Y
	return []stroke.Node{
		ellipseNode(cx, cy, ba, bb, d0, ccw),
		ellipseNode(cx, cy, ba, bb, (d0+d1)/2, ccw),
		ellipseNode(cx, cy, ba, bb, d1, ccw),
	}
}

// dgBranch is dgArc anticlockwise.
func dgBranch(vx, vy, dx, dy, d0, d1 float64) []stroke.Node {
	return dgArc(vx, vy, dx, dy, d0, d1, true)
}

// dgStem returns the nodes of a stem with its entry hook: the small
// hook at the left, then a clockwise arc from the hook to the point (ex, ey),
// where the stem travels in the direction d1.
func dgStem(c *ctx, ex, ey, bend float64) []stroke.Node {
	return dgStemTo(c, c.x+c.o-penWidth(c.pen, 0)/2, ex, ey, math.NaN(), bend)
}

// dgStemTo is dgStem with the final direction d1 given (unless NaN).
func dgStemTo(c *ctx, top, ex, ey, d1, bend float64) []stroke.Node {
	pen := c.pen
	left := 22 + penWidth(pen, 90)/2
	// The hook ends where the stem leaves it, travelling bend degrees
	// clockwise of the direction of the chord to the end of the stem.
	th := -60.0
	for range 60 {
		h := dgHook(left, top, 84, 70, 125, th)[2].P
		ch := math.Atan2(ey-h.Y, ex-h.X) * 180 / math.Pi
		if math.IsNaN(d1) {
			th = ch + bend
		} else {
			th = (th + 2*ch - d1) / 2
		}
	}
	hook := dgHook(left, top, 84, 70, 125, th)
	h := hook[2].P
	if math.IsNaN(d1) {
		d1 = th - 2*bend
	}
	// a circular arc from the hook to the end
	kap := -(th - d1) * math.Pi / 180 / math.Hypot(ex-h.X, ey-h.Y)
	a0 := n(h.X, h.Y, th, auto)
	a0.K = -0.008
	a1 := n(ex, ey, d1, auto)
	a1.K = kap
	arc := []stroke.Node{a0, a1}
	return append(hook[:2:2], arc...)
}

// ν: a stem from the entry hook at the top left to the baseline, and a
// curved branch from its foot to the top right.
func nu(c *ctx) letter {
	pen := c.pen
	top := c.x + c.o - penWidth(pen, 0)/2
	stem := dgStem(c, 215, 22, 8)
	vy := 70.0
	vx := centreX(stem, vy) + 10
	br := dgBranch(vx, vy, 380-vx, top-vy, 52, 88)
	br[len(br)-1].DW = 0
	end := br[len(br)-1].P
	return letter{adv: end.X + penWidth(pen, 88)/2 + 25, strokes: []stroke.Stroke{
		{Nodes: stem, Start: cut, End: cut},
		{Nodes: br, Start: hidden, End: cut},
	}}
}
