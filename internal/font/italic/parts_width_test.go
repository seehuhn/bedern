// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/stroke"
)

// A widthSample is the thickness of a built stroke at the fraction t of the
// length of its middle path.
type widthSample struct{ t, w float64 }

// measureWidth samples the right edge of the stroke about every 3 units and
// measures the distance along its inward normal to the left edge; the middle
// of that distance traces the middle path, along which t is measured.
func measureWidth(r *stroke.Result) []widthSample {
	var left []Point
	for _, pc := range r.Left {
		for j := range 32 {
			left = append(left, pc.C.At(float64(j)/32))
		}
	}
	left = append(left, r.Left[len(r.Left)-1].C[3])
	var out []widthSample
	var mids []Point
	var arc float64
	for _, pc := range r.Right {
		n := max(1, int(math.Ceil(pc.C.Length()/3)))
		for j := range n {
			u := float64(j) / float64(n)
			p, d := pc.C.At(u), pc.C.Tangent(u)
			l := math.Hypot(d.X, d.Y)
			nx, ny := -d.Y/l, d.X/l
			best := math.Inf(1)
			for i := 0; i+1 < len(left); i++ {
				a, b := left[i], left[i+1]
				ex, ey := b.X-a.X, b.Y-a.Y
				den := nx*ey - ny*ex
				if den == 0 {
					continue
				}
				wx, wy := a.X-p.X, a.Y-p.Y
				s := (wx*ey - wy*ex) / den
				v := (wx*ny - wy*nx) / den
				if s > 0.5 && v >= 0 && v <= 1 && s < best {
					best = s
				}
			}
			if math.IsInf(best, 1) {
				continue
			}
			m := pt(p.X+nx*best/2, p.Y+ny*best/2)
			if len(mids) > 0 {
				q := mids[len(mids)-1]
				arc += math.Hypot(m.X-q.X, m.Y-q.Y)
			}
			mids = append(mids, m)
			out = append(out, widthSample{t: arc, w: best})
		}
	}
	for i := range out {
		out[i].t /= arc
	}
	return out
}

// TestWidthAlong checks that a stroke whose nodes take their thickness from a
// profile along it, with nodes added on long spans, has that thickness
// between the nodes too.
func TestWidthAlong(t *testing.T) {
	lines := map[string][]stroke.Node{
		"straight": {
			{P: pt(0, 0), Dir: 60, K: auto},
			{P: pt(200*cosd(60), 200*sind(60)), Dir: 60, K: auto},
			{P: pt(450*cosd(60), 450*sind(60)), Dir: 60, K: auto},
		},
		"curved": {
			{P: pt(0, 0), Dir: 80, K: auto},
			{P: pt(60, 220), Dir: 65, K: auto},
			{P: pt(200, 420), Dir: 45, K: auto},
		},
	}
	profiles := map[string]widthFunc{
		"linear": linearWidth(34, 20),
		"eased":  easeWidth(34, 20, 0.2, 0.8, 0.35),
	}
	for ln, nodes := range lines {
		for pn, f := range profiles {
			ns := subdivide(nodes, widthStep)
			widthAlong(ns, f)
			r, err := stroke.Stroke{Nodes: ns}.Build()
			if err != nil {
				t.Fatalf("%s %s: %v", ln, pn, err)
			}
			worst, at := 0.0, 0.0
			for _, s := range measureWidth(r) {
				if s.t < 0.02 || s.t > 0.98 {
					continue
				}
				want, _ := f(s.t)
				if d := math.Abs(s.w - want); d > worst {
					worst, at = d, s.t
				}
			}
			if worst > 1 {
				t.Errorf("%s stroke, %s profile: the thickness is %.2f off at t = %.2f", ln, pn, worst, at)
			}
		}
	}
}

// widthStep is the longest span between nodes in TestWidthAlong.
const widthStep = 60
