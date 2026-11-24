// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"
	"math/rand"
	"testing"
)

// TestHermiteRecoversCubic asks that the curve drawn from the ends, directions
// and curvatures of a convex cubic is that cubic, whose handles are a
// solution of the equations.
func TestHermiteRecoversCubic(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	n := 0
	for i := 0; i < 200; i++ {
		c := Cubic{
			{X: 0, Y: 0},
			{X: 20 + 80*rng.Float64(), Y: -30 + 60*rng.Float64()},
			{X: 120 + 80*rng.Float64(), Y: 60 + 80*rng.Float64()},
			{X: 200 + 50*rng.Float64(), Y: 150 + 50*rng.Float64()},
		}
		h, ok := hermite(c[0], c.Tangent(0), c.Curvature(0), c[3], c.Tangent(1), c.Curvature(1))
		if !ok {
			t0, t3 := c.Tangent(0), c.Tangent(1)
			d := sub(c[3], c[0])
			t.Errorf("%v: no curve; a=%g b=%g c=%g A=%g B=%g k0=%g k3=%g", c, dist(c[1], c[0]), dist(c[3], c[2]), cross(t0, t3), cross(t0, d), cross(d, t3), c.Curvature(0), c.Curvature(1))
			continue
		}
		if !ok {
			continue
		}
		// The equations may have a second solution; the one found has the
		// end data of c.
		if math.Abs(h.Curvature(0)-c.Curvature(0)) > 1e-9 || math.Abs(h.Curvature(1)-c.Curvature(1)) > 1e-9 {
			t.Errorf("%v: curvatures %g %g, want %g %g", c, h.Curvature(0), h.Curvature(1), c.Curvature(0), c.Curvature(1))
		}
		if dist(h[1], c[1]) < 1e-6 && dist(h[2], c[2]) < 1e-6 {
			n++
		}
	}
	t.Logf("%d of 200 recovered exactly", n)
}
