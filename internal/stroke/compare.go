// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/outline"
)

// A sample is a point of an outline and whether it belongs to the shape (not
// to a line which only bounds a part).
type sample struct {
	P    Point
	Real bool
}

// samples returns points of the results, about one unit apart.
func samples(rs []*Result) []sample {
	var out []sample
	for _, r := range rs {
		for _, c := range r.Contours {
			for _, pc := range c {
				n := int(math.Ceil(pc.C.Length())) + 1
				for i := 0; i <= n; i++ {
					out = append(out, sample{pc.C.At(float64(i) / float64(n)), pc.Kind != JoinPiece})
				}
			}
		}
	}
	return out
}

// polySegs returns the sides of the flattened contours of p.
func polySegs(p outline.Path) [][2]Point {
	var out [][2]Point
	for _, c := range p {
		pts := c.Flatten()
		for i := range pts {
			out = append(out, [2]Point{pts[i], pts[(i+1)%len(pts)]})
		}
	}
	return out
}

// Distance returns the largest distance between the outlines of the results,
// rounded to whole units as the font rounds them, and the outline font: from
// every point of the shape of the results to the font, and from every point of
// the font which lies within reach units of the shape of the results, and whose
// nearest point there is not next to a line which bounds the part, to the
// results.  It also returns the place of the largest distance.
func Distance(rs []*Result, font outline.Path, reach float64) (float64, Point) {
	ds, at := distances(rs, font, reach)
	var worst float64
	var w Point
	for i, d := range ds {
		if d > worst {
			worst, w = d, at[i]
		}
	}
	return worst, w
}

// distances returns the distances which Distance takes the largest of, with
// their places.
func distances(rs []*Result, font outline.Path, reach float64) ([]float64, []Point) {
	var mine outline.Path
	for _, r := range rs {
		mine = append(mine, r.Path()...)
	}
	mineR := mine.Round()
	exact := samples(rs)
	exactPts := make([][2]Point, len(exact))
	for i, e := range exact {
		exactPts[i] = [2]Point{e.P, e.P}
	}
	exactIdx := newSegGrid(exactPts)
	fontIdx := newSegGrid(polySegs(font))
	mineSegs := polySegs(mineR)
	mineIdx := newSegGrid(mineSegs)
	// The rounded outline of the results, each point classified by the
	// nearest exact point.
	var ds []float64
	var at []Point
	var real [][2]Point
	for _, s := range mineSegs {
		q := s[0]
		_, k := exactIdx.nearest(q)
		if !exact[k].Real {
			continue
		}
		real = append(real, [2]Point{q, q})
		d, _ := fontIdx.nearest(q)
		ds, at = append(ds, d), append(at, q)
	}
	realIdx := newSegGrid(real)
	// The other way round.
	var bounds [][2]Point
	for _, e := range exact {
		if !e.Real {
			bounds = append(bounds, [2]Point{e.P, e.P})
		}
	}
	boundIdx := newSegGrid(bounds)
	for _, s := range fontIdx.segs {
		q := s[0]
		dmin, k := realIdx.nearest(q)
		if dmin > reach {
			continue
		}
		if len(bounds) > 0 {
			if db, _ := boundIdx.nearest(real[k][0]); db < 12 {
				continue
			}
		}
		d, _ := mineIdx.nearest(q)
		ds, at = append(ds, d), append(at, q)
	}
	return ds, at
}

// segGrid finds the nearest of a set of segments by a grid of square cells.
type segGrid struct {
	segs  [][2]Point
	cells map[[2]int][]int
}

const gridCell = 8.0

func cellOf(q Point) [2]int {
	return [2]int{int(math.Floor(q.X / gridCell)), int(math.Floor(q.Y / gridCell))}
}

func newSegGrid(segs [][2]Point) *segGrid {
	g := &segGrid{segs: segs, cells: map[[2]int][]int{}}
	for i, s := range segs {
		a, b := cellOf(s[0]), cellOf(s[1])
		for x := min(a[0], b[0]); x <= max(a[0], b[0]); x++ {
			for y := min(a[1], b[1]); y <= max(a[1], b[1]); y++ {
				g.cells[[2]int{x, y}] = append(g.cells[[2]int{x, y}], i)
			}
		}
	}
	return g
}

// nearest returns the distance of q from the nearest segment and its index,
// or +Inf for an empty grid.
func (g *segGrid) nearest(q Point) (float64, int) {
	best, bi := math.Inf(1), -1
	if len(g.segs) == 0 {
		return best, bi
	}
	c := cellOf(q)
	for r := 0; r < 4000; r++ {
		if float64(r-1)*gridCell > best {
			break
		}
		for x := c[0] - r; x <= c[0]+r; x++ {
			for y := c[1] - r; y <= c[1]+r; y++ {
				if max(abs(x-c[0]), abs(y-c[1])) != r {
					continue
				}
				for _, i := range g.cells[[2]int{x, y}] {
					if d := segDist(q, g.segs[i][0], g.segs[i][1]); d < best {
						best, bi = d, i
					}
				}
			}
		}
	}
	return best, bi
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// RoundedPath returns the outlines of the results as one path, rounded to
// whole units as the font rounds.
func RoundedPath(rs []*Result) outline.Path {
	var p outline.Path
	for _, r := range rs {
		p = append(p, r.Path()...)
	}
	return p.Round()
}

// Findings returns the findings of the smoothness analysis on the rounded
// outline of the results which the analysis of the font outline (the whole
// glyph) does not have within 20 units with the same kind, and those which it
// has, leaving out the findings within 30 units of a line which only bounds the
// part.
func Findings(name string, rs []*Result, font outline.Path) (extra, shared []fair.Finding) {
	mine := fair.Analyse(name, RoundedPath(rs), fair.Default)
	theirs := fair.Analyse(name, font, fair.Default)
	var bounds []Point
	for _, e := range samples(rs) {
		if !e.Real {
			bounds = append(bounds, e.P)
		}
	}
	for _, f := range mine.Findings {
		q := Point{X: f.X, Y: f.Y}
		cut := false
		for _, b := range bounds {
			if dist(b, q) < 30 {
				cut = true
				break
			}
		}
		if cut {
			continue
		}
		found := false
		for _, g := range theirs.Findings {
			if g.Kind == f.Kind && dist(q, Point{X: g.X, Y: g.Y}) < 20 {
				found = true
				break
			}
		}
		if found {
			shared = append(shared, f)
		} else {
			extra = append(extra, f)
		}
	}
	return extra, shared
}
