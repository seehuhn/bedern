// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/outline"
)

// segmentsMatch reports whether the outlines have the same contours with the
// same number of segments of the same kinds.
func segmentsMatch(a, b outline.Path) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i].Segs) != len(b[i].Segs) {
			return false
		}
		for j := range a[i].Segs {
			if a[i].Segs[j].IsCurve != b[i].Segs[j].IsCurve {
				return false
			}
		}
	}
	return true
}

// distanceTo returns the largest distance from a node of the outline a to the
// flattened outline b.
func distanceTo(a, b outline.Path) float64 {
	var worst float64
	var polys [][]outline.Point
	for _, c := range b {
		polys = append(polys, c.Flatten())
	}
	for _, c := range a {
		for _, s := range c.Segs {
			q := s.P3
			best := math.Inf(1)
			for _, pts := range polys {
				for k := range pts {
					u, v := pts[k], pts[(k+1)%len(pts)]
					dx, dy := v.X-u.X, v.Y-u.Y
					t := 0.0
					if l := dx*dx + dy*dy; l > 0 {
						t = math.Max(0, math.Min(1, ((q.X-u.X)*dx+(q.Y-u.Y)*dy)/l))
					}
					best = math.Min(best, math.Hypot(u.X+t*dx-q.X, u.Y+t*dy-q.Y))
				}
			}
			worst = math.Max(worst, best)
		}
	}
	return worst
}

// SetLike gives the glyphs of a perturbed set the nodes of the glyphs of the
// regular set and leaves their shapes where Set puts them, for all glyphs
// but a few, and with a set as its own model it returns what Set
// returns.
func TestSetLike(t *testing.T) {
	if testing.Short() {
		t.Skip("builds every glyph four times")
	}
	reg := design.Regular()
	want := setOf(reg)

	self, failed := SetLike(reg, reg)
	for name, err := range failed {
		t.Errorf("%s with the regular set as its own model: %v", name, err)
	}
	for i, g := range self {
		if !segmentsMatch(g.Path, want[i].Path) {
			t.Errorf("%s with the regular set as its own model has other segments than Set gives it", g.Name)
			continue
		}
		if d := distanceTo(g.Path, want[i].Path); d > 1 {
			t.Errorf("%s with the regular set as its own model lies %.2f units from the outline of Set", g.Name, d)
		}
	}

	for _, set := range designtest.Sets(true, 2) {
		plain := setOf(set.Params)
		gs, failed := SetLike(set.Params, reg)
		if n := len(failed); n > len(gs)/20 {
			t.Errorf("%s: %d glyphs cannot take the nodes of the regular set, want at most %d", set.Name, n, len(gs)/20)
		}
		for i, g := range gs {
			if failed[g.Name] != nil {
				if !segmentsMatch(g.Path, plain[i].Path) {
					t.Errorf("%s: %s fails and is not the glyph of Set", set.Name, g.Name)
				}
				continue
			}
			if !segmentsMatch(g.Path, want[i].Path) {
				t.Errorf("%s: %s has other segments than the glyph of the regular set", set.Name, g.Name)
				continue
			}
			if g.Advance != plain[i].Advance {
				t.Errorf("%s: %s has the advance %g, Set gives %g", set.Name, g.Name, g.Advance, plain[i].Advance)
			}
			// Both outlines are rounded, each node by up to 0.71 units, and a
			// merged curve lies within outline.MergeTolerance of the curves
			// which it replaces.
			if d := math.Max(distanceTo(g.Path, plain[i].Path), distanceTo(plain[i].Path, g.Path)); d > 2 {
				t.Errorf("%s: %s lies %.2f units from the outline of Set", set.Name, g.Name, d)
			}
		}
	}
}
