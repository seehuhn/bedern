// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestLoopNodesRotate checks that a loop with the axis turned by 90 degrees
// is the loop turned by 90 degrees about the crossing, before the shear.
func TestLoopNodesRotate(t *testing.T) {
	p := Default()
	k := p.slant()
	g := p.Gk
	cross := pt(300, 200)
	make := func(axis float64) []Point {
		var pts []Point
		for _, n := range p.loopNodes(cross, axis, g.ZetaLoop, g.ZetaEye, g.ZetaTurnIn, p.nibWidth(0),
			g.ZetaLoopW, g.ZetaLoopTaper, g.ZetaTurnDir, g.ZetaCrossIn, g.ZetaCrossOut, g.ZetaCrossShift, 10) {
			pts = append(pts, pt(n.P.X+k*n.P.Y, n.P.Y)) // undo the shear
		}
		return pts
	}
	a, b := make(0), make(90)
	for i := range a {
		dx, dy := a[i].X-cross.X, a[i].Y-cross.Y
		wx, wy := cross.X-dy, cross.Y+dx
		if math.Hypot(b[i].X-wx, b[i].Y-wy) > 1e-9 {
			t.Errorf("node %d: got %v, want (%g, %g)", i, b[i], wx, wy)
		}
	}
}
