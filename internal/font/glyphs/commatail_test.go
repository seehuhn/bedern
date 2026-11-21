// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// tailThickness returns, at the fractions of the length of the inner edge of the
// tail (from the corner at the head to the cut), the distance from the point of
// the inner edge to the nearest point of the outer edge.
func tailThickness(t *testing.T, ch chain, fracs []float64, head, r float64) []float64 {
	t.Helper()
	var inner, outer []cubic
	seenCut := false
	for _, pc := range ch {
		if pc.line {
			seenCut = true
			continue
		}
		if math.Hypot(pc.P0.X, pc.P0.Y-head) > r+0.01 || math.Hypot(pc.P3.X, pc.P3.Y-head) > r+0.01 {
			if seenCut {
				outer = append(outer, pc.cubic)
			} else {
				inner = append(inner, pc.cubic)
			}
		}
	}
	if len(inner) == 0 || len(outer) == 0 {
		t.Fatal("no tail found")
	}
	sample := func(cs []cubic) []outline.Point {
		var out []outline.Point
		for _, c := range cs {
			for k := 0; k <= 400; k++ {
				out = append(out, c.at(float64(k)/400))
			}
		}
		return out
	}
	ip, op := sample(inner), sample(outer)
	cum := make([]float64, len(ip))
	for i := 1; i < len(ip); i++ {
		cum[i] = cum[i-1] + math.Hypot(ip[i].X-ip[i-1].X, ip[i].Y-ip[i-1].Y)
	}
	var out []float64
	for _, f := range fracs {
		i := 0
		for i < len(ip)-1 && cum[i] < f*cum[len(ip)-1] {
			i++
		}
		d := math.Inf(1)
		for _, q := range op {
			d = math.Min(d, math.Hypot(q.X-ip[i].X, q.Y-ip[i].Y))
		}
		out = append(out, d)
	}
	return out
}

// The tail of the comma does not thin to a wire early: over the first half of
// the inner edge from the head its thickness stays at 42 or more (the tail is
// about 50 thick over its first half).
func TestCommaTailKeepsWeight(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		if s.Name != "regular" {
			continue
		}
		p := s.Params
		cy := commaTop(p) - p.Punct.CommaHead/2
		ch := commaAt(p, 0, commaTop(p))
		fr := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		th := tailThickness(t, ch, fr, cy, p.Punct.CommaHead/2)
		for i, d := range th {
			if d < 42 {
				t.Errorf("%s: the tail is %.1f thick at %.0f%% of its length", s.Name, d, 100*fr[i])
			}
		}
	}
}
