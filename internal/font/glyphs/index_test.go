// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// A bandIndex holds the edges of flattened contours by horizontal bands, so
// that a point is tested only against the edges whose height range meets its
// band.  For one polygon its inside gives what inside gives, and for the
// flattened contours of a path its leftOdd gives what inkAt gives.
type bandIndex struct {
	edges [][2]outline.Point
	y0, h float64
	bands [][]int32 // the edges which meet each band
}

const indexBands = 64

// newBandIndex returns the index of the polygons.
func newBandIndex(polys ...[]outline.Point) *bandIndex {
	ix := &bandIndex{}
	y0, y1 := math.Inf(1), math.Inf(-1)
	for _, poly := range polys {
		for i, a := range poly {
			ix.edges = append(ix.edges, [2]outline.Point{a, poly[(i+1)%len(poly)]})
			y0, y1 = math.Min(y0, a.Y), math.Max(y1, a.Y)
		}
	}
	ix.bands = make([][]int32, indexBands)
	if !(y1 > y0) {
		// One band for all edges.
		ix.y0, ix.h = 0, math.Inf(1)
	} else {
		ix.y0, ix.h = y0, (y1-y0)/indexBands
	}
	for i, e := range ix.edges {
		lo, hi := ix.band(math.Min(e[0].Y, e[1].Y)), ix.band(math.Max(e[0].Y, e[1].Y))
		for k := lo; k <= hi; k++ {
			ix.bands[k] = append(ix.bands[k], int32(i))
		}
	}
	return ix
}

// pathIndex returns the index of the flattened contours of the path.
func pathIndex(path outline.Path) *bandIndex {
	polys := make([][]outline.Point, len(path))
	for i, ct := range path {
		polys[i] = ct.Flatten()
	}
	return newBandIndex(polys...)
}

// band returns the band of the height y, clamped to the bands.
func (ix *bandIndex) band(y float64) int {
	v := math.Floor((y - ix.y0) / ix.h)
	if !(v > 0) {
		return 0
	}
	return int(min(v, indexBands-1))
}

// crossings calls f with the abscissa of each crossing of the edges with the
// horizontal line at y, by the half-open rule of outline.Path.CrossingsX.
func (ix *bandIndex) crossings(y float64, f func(x float64)) {
	for _, i := range ix.bands[ix.band(y)] {
		a, b := ix.edges[i][0], ix.edges[i][1]
		if (a.Y <= y) == (b.Y <= y) {
			continue
		}
		f(a.X + (y-a.Y)/(b.Y-a.Y)*(b.X-a.X))
	}
}

// inside reports whether q lies inside the polygons (even-odd rule), counting
// the crossings to the right of q as inside does.
func (ix *bandIndex) inside(q outline.Point) bool {
	in := false
	ix.crossings(q.Y, func(x float64) {
		if x > q.X {
			in = !in
		}
	})
	return in
}

// leftOdd reports whether an odd number of crossings lie to the left of q, as
// inkAt counts them.
func (ix *bandIndex) leftOdd(q outline.Point) bool {
	in := false
	ix.crossings(q.Y, func(x float64) {
		if x < q.X {
			in = !in
		}
	})
	return in
}

// A pointGrid holds points by square cells, for the least distance from a
// point to any of them.
type pointGrid struct {
	size   float64
	x0, y0 int // the cell of the lower left corner
	nx, ny int
	cells  [][]outline.Point // by (x-x0)*ny + (y-y0)
}

// newPointGrid returns the grid of the points with cells of the given size.
func newPointGrid(pts []outline.Point, size float64) *pointGrid {
	g := &pointGrid{size: size}
	if len(pts) == 0 {
		return g
	}
	x0, y0, x1, y1 := math.MaxInt, math.MaxInt, math.MinInt, math.MinInt
	for _, q := range pts {
		x, y := g.cell(q)
		x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
	}
	g.x0, g.y0, g.nx, g.ny = x0, y0, x1-x0+1, y1-y0+1
	g.cells = make([][]outline.Point, g.nx*g.ny)
	for _, q := range pts {
		x, y := g.cell(q)
		k := (x-x0)*g.ny + (y - y0)
		g.cells[k] = append(g.cells[k], q)
	}
	return g
}

func (g *pointGrid) cell(q outline.Point) (int, int) {
	return int(math.Floor(q.X / g.size)), int(math.Floor(q.Y / g.size))
}

// nearest returns the least value of math.Hypot(o.X-q.X, o.Y-q.Y) over the
// points o, or +Inf if there are none.  The cells are searched in square
// rings round the cell of q until the points beyond the next ring are farther
// than the best distance found, with one ring to spare.
func (g *pointGrid) nearest(q outline.Point) float64 {
	best := math.Inf(1)
	if len(g.cells) == 0 {
		return best
	}
	cx, cy := g.cell(q)
	visit := func(x, y int) {
		x, y = x-g.x0, y-g.y0
		if x < 0 || y < 0 || x >= g.nx || y >= g.ny {
			return
		}
		for _, o := range g.cells[x*g.ny+y] {
			best = math.Min(best, math.Hypot(o.X-q.X, o.Y-q.Y))
		}
	}
	for r := 0; ; r++ {
		if r == 0 {
			visit(cx, cy)
		} else {
			for d := -r; d <= r; d++ {
				visit(cx+d, cy-r)
				visit(cx+d, cy+r)
			}
			for d := -r + 1; d < r; d++ {
				visit(cx-r, cy+d)
				visit(cx+r, cy+d)
			}
		}
		if best <= float64(r-1)*g.size {
			return best
		}
	}
}
