// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// dgCentre returns the centre of the ellipse with the half axes a and b whose
// node travelling in the direction dir lies at (px, py).
func dgCentre(a, b, dir float64, ccw bool, px, py float64) (float64, float64) {
	o := ellipseNode(0, 0, a, b, dir, ccw).P
	return px - o.X, py - o.Y
}

// dgLineX returns the abscissa at height y of the line through (x0, y0)
// travelling in the direction dir.
func dgLineX(x0, y0, dir, y float64) float64 {
	return x0 + (y-y0)/math.Tan(dir*math.Pi/180)
}

// dgHook returns the nodes of the entry hook of a letter: the upright
// ellipse about the left end of the stroke, from its start at the lower left
// over the top to the node at which the stroke travels in the direction end.
// The top of the centre line is at height top, the left of the centre line
// at x = left.
func dgHook(left, top, a, b, start, end float64) []stroke.Node {
	cy := top - b
	cx := left + a
	return []stroke.Node{
		ellipseNode(cx, cy, a, b, start, false),
		ellipseNode(cx, cy, a, b, 0, false),
		ellipseNode(cx, cy, a, b, end, false),
	}
}
