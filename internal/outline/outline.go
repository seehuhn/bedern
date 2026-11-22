// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package outline provides a minimal path type for glyph outlines, with
// helpers to measure the shapes in tests.
package outline

import (
	"math"
	"sort"
	"sync"
)

// Point is a point in font units.
type Point struct{ X, Y float64 }

// Segment is a line or a cubic Bézier curve.  It starts at the end of the
// previous segment and ends at P3.  P1 and P2 are used only for curves.
type Segment struct {
	IsCurve    bool
	P1, P2, P3 Point
}

// Contour is a closed contour.  The last segment is joined to Start by a
// straight line if it does not end there.
type Contour struct {
	Start Point
	Segs  []Segment
}

// Path is a list of contours.
type Path []Contour

// MoveTo starts a new contour.
func (p *Path) MoveTo(x, y float64) {
	*p = append(*p, Contour{Start: Point{x, y}})
}

// LineTo adds a straight line to the current contour.
func (p *Path) LineTo(x, y float64) {
	c := &(*p)[len(*p)-1]
	c.Segs = append(c.Segs, Segment{P3: Point{x, y}})
}

// CurveTo adds a cubic Bézier curve to the current contour.
func (p *Path) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	c := &(*p)[len(*p)-1]
	c.Segs = append(c.Segs, Segment{
		IsCurve: true,
		P1:      Point{x1, y1}, P2: Point{x2, y2}, P3: Point{x3, y3},
	})
}

// flattenSteps is the number of straight pieces per curve used for measuring.
const flattenSteps = 64

// Flatten returns the contour as a closed polygon.  The first point is not
// repeated at the end.
func (c Contour) Flatten() []Point {
	n := 1
	for _, s := range c.Segs {
		if s.IsCurve {
			n += flattenSteps
		} else {
			n++
		}
	}
	return c.appendFlat(make([]Point, 0, n))
}

// appendFlat is Flatten with the points appended to pts, which is empty.
func (c Contour) appendFlat(pts []Point) []Point {
	pts = append(pts, c.Start)
	cur := c.Start
	for _, s := range c.Segs {
		if !s.IsCurve {
			pts = append(pts, s.P3)
			cur = s.P3
			continue
		}
		for i := 1; i <= flattenSteps; i++ {
			t := float64(i) / flattenSteps
			u := 1 - t
			a, b, cc, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
			pts = append(pts, Point{
				X: a*cur.X + b*s.P1.X + cc*s.P2.X + d*s.P3.X,
				Y: a*cur.Y + b*s.P1.Y + cc*s.P2.Y + d*s.P3.Y,
			})
		}
		cur = s.P3
	}
	if n := len(pts); n > 1 && pts[n-1] == pts[0] {
		pts = pts[:n-1]
	}
	return pts
}

// scratch holds the polygons of the scans.
var scratch = sync.Pool{New: func() any { return new([]Point) }}

// SignedArea returns the area enclosed by the contour.  It is positive for
// a counter-clockwise contour.
func (c Contour) SignedArea() float64 {
	buf := scratch.Get().(*[]Point)
	defer scratch.Put(buf)
	pts := c.appendFlat((*buf)[:0])
	*buf = pts
	var a float64
	for i, p := range pts {
		q := pts[(i+1)%len(pts)]
		a += p.X*q.Y - q.X*p.Y
	}
	return a / 2
}

// CrossingsX returns the x coordinates at which the horizontal line at
// height y crosses the path, in increasing order.
//
// A segment of the flattened outline is counted when exactly one of its end
// points lies at or below the scan line (a half-open rule: points on the
// line count as below it).  Consequently, a scan line that passes exactly
// through a local extreme of the outline, for example the baseline edge of a
// shape sitting on the baseline or its top edge, reports either a doubled
// crossing or none at all.
func (p Path) CrossingsX(y float64) []float64 {
	var xs []float64
	buf := scratch.Get().(*[]Point)
	defer scratch.Put(buf)
	for _, c := range p {
		pts := c.appendFlat((*buf)[:0])
		*buf = pts
		for i, a := range pts {
			b := pts[(i+1)%len(pts)]
			if (a.Y <= y) == (b.Y <= y) {
				continue
			}
			t := (y - a.Y) / (b.Y - a.Y)
			xs = append(xs, a.X+t*(b.X-a.X))
		}
	}
	sort.Float64s(xs)
	return xs
}

// CrossingsY returns the y coordinates at which the vertical line at x
// crosses the path, in increasing order.
//
// A segment of the flattened outline is counted when exactly one of its end
// points lies at or to the left of the scan line (a half-open rule: points on
// the line count as left of it).  Consequently, a scan line that passes
// exactly through a local extreme of the outline, for example the left or
// right edge of a shape, reports either a doubled crossing or none at all.
func (p Path) CrossingsY(x float64) []float64 {
	var ys []float64
	buf := scratch.Get().(*[]Point)
	defer scratch.Put(buf)
	for _, c := range p {
		pts := c.appendFlat((*buf)[:0])
		*buf = pts
		for i, a := range pts {
			b := pts[(i+1)%len(pts)]
			if (a.X <= x) == (b.X <= x) {
				continue
			}
			t := (x - a.X) / (b.X - a.X)
			ys = append(ys, a.Y+t*(b.Y-a.Y))
		}
	}
	sort.Float64s(ys)
	return ys
}

// BBox returns the bounding box of the path, computed from the flattened
// contours.  An empty path gives infinite values: +Inf for the minima and
// -Inf for the maxima.
func (p Path) BBox() (x0, y0, x1, y1 float64) {
	x0, y0 = math.Inf(1), math.Inf(1)
	x1, y1 = math.Inf(-1), math.Inf(-1)
	buf := scratch.Get().(*[]Point)
	defer scratch.Put(buf)
	for _, c := range p {
		pts := c.appendFlat((*buf)[:0])
		*buf = pts
		for _, q := range pts {
			x0, x1 = min(x0, q.X), max(x1, q.X)
			y0, y1 = min(y0, q.Y), max(y1, q.Y)
		}
	}
	return
}

// segEps is the distance below which SelfIntersections treats a point as
// lying on a line.
const segEps = 1e-9

type flatSeg struct {
	a, b           Point
	contour, index int
	minX, maxX     float64
	minY, maxY     float64
	adjacent       func(o *flatSeg) bool
}

// SelfIntersections returns the points at which the flattened contours of
// the path cross or touch, in increasing order of x and then y.  The
// segments of a contour are those of Flatten, including the closing segment
// from the last point to the first.  Two segments are compared if they are
// not neighbours in the same contour: neighbours share an end point, which is
// not an intersection.  Segments of zero length are ignored.
//
// What counts as an intersection:
//
//   - A proper crossing of two segments is reported at the crossing point.
//   - A vertex which lies on another segment, or a vertex of one contour
//     which coincides with a vertex of another contour (or a non-neighbouring
//     vertex of the same one), is a touching and is reported at that vertex.
//   - Two collinear segments which overlap in more than a point are reported
//     at the two ends of the overlap.  A contour traced twice is thus
//     reported at each of its vertices.
//   - Where a segment returns along its neighbour (a spike with an angle
//     of zero), the tip is reported, although the segments are neighbours.
//
// Points are on a line when they lie within 1e-9 units of it.  The result
// depends on the flattening, which approximates curves by 64 straight pieces.
// Curves which cross at a shallow angle, or touch without crossing, are not
// reliably distinguished.  Duplicate points are reported once.
func (p Path) SelfIntersections() []Point {
	var segs []*flatSeg
	for ci, c := range p {
		pts := dedupe(c.Flatten())
		n := len(pts)
		if n < 2 {
			continue
		}
		first := len(segs)
		for i := range pts {
			if n == 2 && i == 1 {
				break
			}
			a, b := pts[i], pts[(i+1)%n]
			segs = append(segs, &flatSeg{
				a: a, b: b, contour: ci, index: i,
				minX: min(a.X, b.X), maxX: max(a.X, b.X),
				minY: min(a.Y, b.Y), maxY: max(a.Y, b.Y),
			})
		}
		count := len(segs) - first
		for _, s := range segs[first:] {
			s.adjacent = func(o *flatSeg) bool {
				if o.contour != s.contour {
					return false
				}
				d := (o.index - s.index + count) % count
				return d == 1 || d == count-1
			}
		}
	}

	var out []Point
	// Spikes: neighbouring segments which fold back onto each other.
	for _, s := range segs {
		for _, o := range segs {
			if s.contour == o.contour && s.b == o.a && s != o && foldsBack(s, o) {
				out = append(out, s.b)
			}
		}
	}

	// Sweep over the segments sorted by their left end.
	sort.Slice(segs, func(i, j int) bool { return segs[i].minX < segs[j].minX })
	for i, s := range segs {
		for _, o := range segs[i+1:] {
			if o.minX > s.maxX+segEps {
				break
			}
			if o.minY > s.maxY+segEps || o.maxY < s.minY-segEps || s.adjacent(o) {
				continue
			}
			out = append(out, meet(s, o)...)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].X != out[j].X {
			return out[i].X < out[j].X
		}
		return out[i].Y < out[j].Y
	})
	var res []Point
	for _, q := range out {
		if n := len(res); n > 0 && math.Abs(res[n-1].X-q.X) < 1e-6 && math.Abs(res[n-1].Y-q.Y) < 1e-6 {
			continue
		}
		res = append(res, q)
	}
	return res
}

// dedupe removes consecutive repeated points, also across the closing edge.
func dedupe(pts []Point) []Point {
	var out []Point
	for _, q := range pts {
		if n := len(out); n == 0 || out[n-1] != q {
			out = append(out, q)
		}
	}
	for len(out) > 1 && out[len(out)-1] == out[0] {
		out = out[:len(out)-1]
	}
	return out
}

// foldsBack reports whether o, which starts where s ends, runs back along s.
func foldsBack(s, o *flatSeg) bool {
	ux, uy := s.b.X-s.a.X, s.b.Y-s.a.Y
	vx, vy := o.b.X-o.a.X, o.b.Y-o.a.Y
	cross := ux*vy - uy*vx
	return math.Abs(cross)/math.Hypot(ux, uy) < segEps && ux*vx+uy*vy < 0
}

// side returns the signed distance of q from the line through a and b.
func side(a, b, q Point) float64 {
	return ((b.X-a.X)*(q.Y-a.Y) - (b.Y-a.Y)*(q.X-a.X)) / math.Hypot(b.X-a.X, b.Y-a.Y)
}

func sgn(d float64) int {
	switch {
	case d > segEps:
		return 1
	case d < -segEps:
		return -1
	}
	return 0
}

// meet returns the points where the closed segments s and o have a point in
// common, following the rules of SelfIntersections.
func meet(s, o *flatSeg) []Point {
	d1, d2 := sgn(side(s.a, s.b, o.a)), sgn(side(s.a, s.b, o.b))
	d3, d4 := sgn(side(o.a, o.b, s.a)), sgn(side(o.a, o.b, s.b))

	if d1 == 0 && d2 == 0 {
		// Collinear: project onto the longer coordinate axis.
		key := func(q Point) float64 { return q.X }
		if s.maxY-s.minY > s.maxX-s.minX {
			key = func(q Point) float64 { return q.Y }
		}
		pts := []Point{s.a, s.b, o.a, o.b}
		lo, hi := pts[0], pts[0]
		for _, q := range pts[:2] {
			if key(q) < key(lo) {
				lo = q
			}
			if key(q) > key(hi) {
				hi = q
			}
		}
		olo, ohi := o.a, o.b
		if key(olo) > key(ohi) {
			olo, ohi = ohi, olo
		}
		start, end := olo, ohi
		if key(lo) > key(olo) {
			start = lo
		}
		if key(hi) < key(ohi) {
			end = hi
		}
		switch {
		case key(start) > key(end)+segEps:
			return nil
		case key(end)-key(start) <= segEps:
			return []Point{start}
		}
		return []Point{start, end}
	}
	if d1*d2 > 0 || d3*d4 > 0 {
		return nil
	}
	switch {
	case d1 == 0:
		return []Point{o.a}
	case d2 == 0:
		return []Point{o.b}
	case d3 == 0:
		return []Point{s.a}
	case d4 == 0:
		return []Point{s.b}
	}
	// A proper crossing.
	t := side(o.a, o.b, s.a) / (side(o.a, o.b, s.a) - side(o.a, o.b, s.b))
	return []Point{{s.a.X + t*(s.b.X-s.a.X), s.a.Y + t*(s.b.Y-s.a.Y)}}
}

// CubicExtrema returns the parameters in (0, 1) at which one coordinate of a
// cubic Bézier curve with the control values p0, p1, p2 and p3 is extreme,
// from the roots of its derivative.  The roots are found with the stable form
// of the quadratic formula, which does not lose the small root when the
// leading coefficient is tiny.  A double root, where the coordinate has no
// extreme, is not returned.
func CubicExtrema(p0, p1, p2, p3 float64) []float64 {
	// The derivative is 3 (a t^2 + b t + c).
	a := -p0 + 3*p1 - 3*p2 + p3
	b := 2 * (p0 - 2*p1 + p2)
	c := p1 - p0
	var ts []float64
	switch {
	case a == 0:
		if b != 0 {
			ts = append(ts, -c/b)
		}
	default:
		disc := b*b - 4*a*c
		if disc <= 0 {
			break
		}
		// q has the sign of -b; the roots are q/a and c/q.
		q := -(b + math.Copysign(math.Sqrt(disc), b)) / 2
		ts = append(ts, q/a)
		if q != 0 {
			ts = append(ts, c/q)
		}
	}
	var out []float64
	for _, t := range ts {
		if t > 0 && t < 1 {
			out = append(out, t)
		}
	}
	sort.Float64s(out)
	return out
}

// pieceLength is the length to which MinDistance cuts the sides of the
// flattened contours.
const pieceLength = 2.0

type piece struct {
	a, b       Point
	contour    int
	s0, s1     float64 // arc length along the contour at the ends of the piece
	total      float64 // length of the contour
	minX, maxX float64
	minY, maxY float64
}

// MinDistance returns the smallest distance between two parts of the outline
// which are not neighbours, and the two points at which it is attained.  Parts
// of different contours are never neighbours.  Two parts of one contour are
// neighbours unless the arc length between them along the contour, measured in
// the shorter direction around, is more than ratio times their distance.  If
// there is no pair, the distance is +Inf.  The distances are those of the
// flattened contours, and the arc length is exact to about 2 units.
//
// The two sides of a stroke which is at least a few widths long are far apart
// along the outline compared with their distance, and the thickness of the
// stroke is found.  Parts on both sides of a corner, and the parts of a smooth
// curve, are close along the outline: for a corner with an opening angle a the
// ratio of arc length to distance is 1 / sin(a/2), which is 1.41 for a right
// angle and 3.4 for an angle of 34 degrees, and for a curve it is 1.  A ratio
// of 5 ignores every corner wider than 23 degrees.  A stroke of width w and
// length L has sides whose arc length apart is up to w + L, and the stroke
// is found if L is more than (ratio - 1) w, which is four widths for the
// ratio 5.  A narrower or shorter place is not found.
func (p Path) MinDistance(ratio float64) (d float64, a, b Point) {
	return p.MinDistanceExcept(ratio, nil)
}

// MinDistanceExcept is MinDistance without the pairs of straight pieces, each from
// a to b, for which skip returns true.  A nil skip leaves out no pair.  The pieces
// are those of the flattened contours, at most about 2 units long.
func (p Path) MinDistanceExcept(ratio float64, skip func(a0, a1, b0, b1 Point) bool) (d float64, a, b Point) {
	var ps []piece
	for ci, c := range p {
		pts := dedupe(c.Flatten())
		n := len(pts)
		if n < 2 {
			continue
		}
		first := len(ps)
		var s float64
		for i := range pts {
			from, to := pts[i], pts[(i+1)%n]
			if n == 2 && i == 1 {
				break
			}
			l := math.Hypot(to.X-from.X, to.Y-from.Y)
			k := int(math.Ceil(l / pieceLength))
			for j := 0; j < k; j++ {
				t0, t1 := float64(j)/float64(k), float64(j+1)/float64(k)
				pa := Point{from.X + t0*(to.X-from.X), from.Y + t0*(to.Y-from.Y)}
				pb := Point{from.X + t1*(to.X-from.X), from.Y + t1*(to.Y-from.Y)}
				ps = append(ps, piece{
					a: pa, b: pb, contour: ci, s0: s + t0*l, s1: s + t1*l,
					minX: min(pa.X, pb.X), maxX: max(pa.X, pb.X),
					minY: min(pa.Y, pb.Y), maxY: max(pa.Y, pb.Y),
				})
			}
			s += l
		}
		for i := first; i < len(ps); i++ {
			ps[i].total = s
		}
	}

	// The pieces are visited in the order of their left ends.  For each piece
	// the search stops at the first piece which starts more than d to the
	// right of it.  Among pairs at the smallest distance, the first pair in
	// the order of the indices i < j is returned.
	order := make([]int, len(ps))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(x, y int) bool {
		if ps[order[x]].minX != ps[order[y]].minX {
			return ps[order[x]].minX < ps[order[y]].minX
		}
		return order[x] < order[y]
	})
	d = math.Inf(1)
	bi, bj := len(ps), len(ps)
	for x, ix := range order {
		pi := &ps[ix]
		for _, iy := range order[x+1:] {
			po := &ps[iy]
			if po.minX-pi.maxX > d {
				break
			}
			if po.minY-pi.maxY > d || pi.minY-po.maxY > d {
				continue
			}
			i, j, s, o := ix, iy, pi, po
			if i > j {
				i, j, s, o = j, i, o, s
			}
			if o.contour == s.contour {
				// The distance of the boxes of two pieces is a lower bound of
				// the distance of the pieces.  Pieces which are neighbours by
				// this bound need no exact distance.
				gx := max(0, max(po.minX-pi.maxX, pi.minX-po.maxX))
				gy := max(0, max(po.minY-pi.maxY, pi.minY-po.maxY))
				lb := math.Sqrt(gx*gx+gy*gy) * (1 - 1e-9)
				if min(o.s0-s.s1, s.total-(o.s1-s.s0)) <= ratio*lb {
					continue
				}
			}
			if skip != nil && skip(s.a, s.b, o.a, o.b) {
				continue
			}
			dd, pa, pb := segmentDistance(s.a, s.b, o.a, o.b)
			if dd > d || math.IsInf(dd, 1) || math.IsNaN(dd) {
				continue
			}
			if dd == d && (i > bi || (i == bi && j > bj)) {
				continue
			}
			if o.contour == s.contour {
				// The pieces are in order along the contour.
				gap := min(o.s0-s.s1, s.total-(o.s1-s.s0))
				if gap <= ratio*dd {
					continue
				}
			}
			d, a, b, bi, bj = dd, pa, pb, i, j
		}
	}
	return d, a, b
}

// segmentDistance returns the distance between the segments ab and cd and a
// pair of points, one on each, which attain it.
func segmentDistance(a, b, c, d Point) (float64, Point, Point) {
	best := math.Inf(1)
	var pa, pb Point
	try := func(p, q Point) {
		if dd := math.Hypot(p.X-q.X, p.Y-q.Y); dd < best {
			best, pa, pb = dd, p, q
		}
	}
	// If the segments cross, the distance is zero.  Nearly collinear segments
	// do not count as crossing, whatever the rounding errors say.
	if d1, d2 := side(a, b, c), side(a, b, d); sgn(d1)*sgn(d2) < 0 {
		if d3, d4 := side(c, d, a), side(c, d, b); sgn(d3)*sgn(d4) < 0 {
			t := d3 / (d3 - d4)
			x := Point{a.X + t*(b.X-a.X), a.Y + t*(b.Y-a.Y)}
			return 0, x, x
		}
	}
	// Otherwise the minimum is attained at an end point of one of them.
	try(a, closest(c, d, a))
	try(b, closest(c, d, b))
	try(closest(a, b, c), c)
	try(closest(a, b, d), d)
	return best, pa, pb
}

// closest returns the point of the segment ab which lies closest to q.
func closest(a, b, q Point) Point {
	dx, dy := b.X-a.X, b.Y-a.Y
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return a
	}
	t := max(0, min(1, ((q.X-a.X)*dx+(q.Y-a.Y)*dy)/l2))
	return Point{a.X + t*dx, a.Y + t*dy}
}
