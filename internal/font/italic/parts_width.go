// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// A widthFunc gives the thickness of a stroke and its rate of change at the
// fraction t of the length of its centre line: dw is the change of w per unit
// of t.
type widthFunc func(t float64) (w, dw float64)

// centreLengths returns the length of the centre line from the first node to
// each node, measured along the cubics of stroke.CentreLine.
func centreLengths(nodes []stroke.Node) []float64 {
	s := make([]float64, len(nodes))
	for i := 1; i < len(nodes); i++ {
		s[i] = s[i-1] + stroke.CentreLine(nodes[i-1], nodes[i]).Length()
	}
	return s
}

// widthAlong sets the thickness W and its rate of change DW (per unit of
// length of the centre line) of every node from f, at the fraction of the
// length of the centre line from the first node to the last at which the node
// lies.
func widthAlong(nodes []stroke.Node, f widthFunc) {
	s := centreLengths(nodes)
	total := s[len(s)-1]
	for i := range nodes {
		w, dw := f(s[i] / total)
		nodes[i].W, nodes[i].DW = w, dw/total
	}
}

// subdivide returns the nodes with nodes added on the centre line between
// any two of them which lie further apart along it than maxLen, evenly by
// length: on the cubic of stroke.CentreLine, with its direction there, the
// curvature by rule and the thickness and its rate of change interpolated
// linearly.
func subdivide(nodes []stroke.Node, maxLen float64) []stroke.Node {
	out := []stroke.Node{nodes[0]}
	for i := 1; i < len(nodes); i++ {
		a, b := nodes[i-1], nodes[i]
		c := stroke.CentreLine(a, b)
		l := c.Length()
		n := int(math.Ceil(l / maxLen))
		for j := 1; j < n; j++ {
			f := float64(j) / float64(n)
			t := paramAt(c, f*l)
			d := c.Tangent(t)
			out = append(out, stroke.Node{
				P:   c.At(t),
				Dir: math.Atan2(d.Y, d.X) * 180 / math.Pi,
				K:   auto,
				W:   a.W + f*(b.W-a.W),
				DW:  a.DW + f*(b.DW-a.DW),
			})
		}
		out = append(out, b)
	}
	return out
}

// paramAt returns the parameter of the cubic at which its length from the
// start is s.
func paramAt(c stroke.Cubic, s float64) float64 {
	lo, hi := 0.0, 1.0
	for range 40 {
		m := (lo + hi) / 2
		if c.Sub(0, m).Length() < s {
			lo = m
		} else {
			hi = m
		}
	}
	return (lo + hi) / 2
}

// constantWidth is the thickness w all along the stroke.
func constantWidth(w float64) widthFunc {
	return func(float64) (float64, float64) { return w, 0 }
}

// linearWidth changes evenly from a at the start of the stroke to b at its
// end.
func linearWidth(a, b float64) widthFunc {
	return func(t float64) (float64, float64) { return a + t*(b-a), b - a }
}

// easeWidth is a up to t0, b from t1 on, and between them changes from a to
// b monotonically, smoothly from no change at t0 and t1 to the fastest
// change at tm.  The rate of change is a parabola on either side of tm.
func easeWidth(a, b, t0, t1, tm float64) widthFunc {
	span := t1 - t0
	um := (tm - t0) / span
	return func(t float64) (float64, float64) {
		u := (t - t0) / span
		switch {
		case u <= 0:
			return a, 0
		case u >= 1:
			return b, 0
		}
		// g' = 3/2 (1 - x²) with x the distance from um as a fraction of
		// the side, so that g(0) = 0 and g(1) = 1.
		var g, dg float64
		if u < um {
			x := u / um
			g = 1.5 * um * (x*x - x*x*x/3)
			dg = 1.5 * x * (2 - x)
		} else {
			y := (u - um) / (1 - um)
			g = 1.5 * (um*2/3 + (1-um)*(y-y*y*y/3))
			dg = 1.5 * (1 - y*y)
		}
		return a + (b-a)*g, (b - a) * dg / span
	}
}
