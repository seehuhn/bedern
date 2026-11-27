// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import "seehuhn.de/go/bedern/internal/stroke"

// The descending outstroke of ζ, ξ and ς: from
// the left extreme of the lower bowl, travelling down, in a long flat bend
// to the right, then round a hook at the bottom with a blunt end.
const (
	outRight = 312.0 // the right of the centre line of the bend
	outHook  = 88.0  // the half width of the hook
)

// outstroke returns the nodes after the left extreme (lx, ly) of the lower
// bowl, which the caller has placed with the direction -90.
func outstroke(c *ctx, lx, ly float64) []stroke.Node {
	yb := c.desc + 12 + penWidth(c.pen, 180)/2
	ry := 92.0
	hcx, hcy := outRight-outHook, yb+ry
	a, b := 150.0, ly-62
	cx := lx + a
	end := ellipseNode(hcx, hcy, outHook, ry, 120, false)
	return []stroke.Node{
		ellipseNode(cx, ly, a, b, -42, true),
		ellipseNode(hcx, hcy, outHook, ry, -45, false),
		ellipseNode(hcx, hcy, outHook, ry, -90, false),
		ellipseNode(hcx, hcy, outHook, ry, 180, false),
		end,
	}
}
