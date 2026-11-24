// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"sort"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The glyphs of the default parameters, built once and shared by the tests.
var (
	glyphsOnce sync.Once
	glyphSet   map[string]*Glyph
	glyphErr   error
)

// glyph returns the glyph of the default parameters.
func glyph(t *testing.T, name string) *Glyph {
	t.Helper()
	glyphsOnce.Do(func() {
		glyphSet = map[string]*Glyph{}
		gs, err := Glyphs(Default())
		if err != nil {
			glyphErr = err
			return
		}
		for _, g := range gs {
			glyphSet[g.Name] = g
		}
	})
	if glyphErr != nil {
		t.Fatal(glyphErr)
	}
	g, ok := glyphSet[name]
	if !ok {
		t.Fatalf("no glyph %s", name)
	}
	return g
}

// polys returns the contours of the path as polygons, each cubic flattened
// into 48 pieces.
func polys(p outline.Path) [][]Point {
	var out [][]Point
	for _, c := range p {
		pts := []Point{c.Start}
		cur := c.Start
		for _, s := range c.Segs {
			if !s.IsCurve {
				pts = append(pts, s.P3)
				cur = s.P3
				continue
			}
			for i := 1; i <= 48; i++ {
				u := float64(i) / 48
				v := 1 - u
				pts = append(pts, Point{
					X: v*v*v*cur.X + 3*v*v*u*s.P1.X + 3*v*u*u*s.P2.X + u*u*u*s.P3.X,
					Y: v*v*v*cur.Y + 3*v*v*u*s.P1.Y + 3*v*u*u*s.P2.Y + u*u*u*s.P3.Y,
				})
			}
			cur = s.P3
		}
		out = append(out, pts)
	}
	return out
}

// crossings returns the parameters s at which the line q + s d crosses the
// outline, sorted.
func crossings(p outline.Path, q, d Point) []float64 {
	var ss []float64
	for _, pts := range polys(p) {
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			e := Point{X: b.X - a.X, Y: b.Y - a.Y}
			den := d.X*e.Y - d.Y*e.X
			if math.Abs(den) < 1e-12 {
				continue
			}
			w := Point{X: a.X - q.X, Y: a.Y - q.Y}
			s := (w.X*e.Y - w.Y*e.X) / den
			u := (w.X*d.Y - w.Y*d.X) / den
			if u >= 0 && u < 1 {
				ss = append(ss, s)
			}
		}
	}
	sort.Float64s(ss)
	return ss
}

// runs returns the runs of ink along the horizontal line at the height y,
// as pairs of abscissae.
func runs(p outline.Path, y float64) [][2]float64 {
	xs := crossings(p, Point{X: 0, Y: y}, Point{X: 1})
	var out [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		out = append(out, [2]float64{xs[i], xs[i+1]})
	}
	return out
}

// bbox returns the bounding box of the outline.
func bbox(p outline.Path) (x0, y0, x1, y1 float64) {
	x0, y0, x1, y1 = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, pts := range polys(p) {
		for _, q := range pts {
			x0, y0 = math.Min(x0, q.X), math.Min(y0, q.Y)
			x1, y1 = math.Max(x1, q.X), math.Max(y1, q.Y)
		}
	}
	return
}

// gapBetween returns the least distance between the outline a and the
// outline b moved right by dx.
func gapBetween(a, b outline.Path, dx float64) float64 {
	pa, pb := polys(a), polys(b)
	d := math.Inf(1)
	for _, ca := range pa {
		for _, q := range ca {
			for _, cb := range pb {
				for _, r := range cb {
					d = math.Min(d, math.Hypot(q.X-r.X-dx, q.Y-r.Y))
				}
			}
		}
	}
	return d
}

// italicX returns the abscissa in the italic of the point at the height y
// of the line which stands at x at ShearHeight.
func italicX(p Params, x, y float64) float64 { return x + (y-p.ShearHeight)*p.slant() }

// stemRun returns the run of ink at the height y which contains the centre
// line of the stem whose centre stands at c at ShearHeight.
func stemRun(t *testing.T, p Params, path outline.Path, c, y float64) [2]float64 {
	t.Helper()
	x := italicX(p, c, y)
	for _, r := range runs(path, y) {
		if r[0] <= x && x <= r[1] {
			return r
		}
	}
	t.Fatalf("no ink on the stem at %.0f,%.0f", x, y)
	return [2]float64{}
}
