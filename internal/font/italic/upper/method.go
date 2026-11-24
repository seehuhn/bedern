// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// A round is a round part of an upright glyph whose inner edge the stress of
// the italic turns: the nodes within the super-ellipse of the inner edge,
// centre C and half axes A and B, and within the span of directions from
// From to To (degrees, anticlockwise, of the direction from C scaled to the
// circle) are turned about C by the angle of the stress, and those within
// Taper degrees outside the span by a share which falls smoothly to zero.
// The outer edge stays.  Each node turns with its two handles; a
// smooth node stays smooth.
type round struct {
	C            Point
	A, B         float64
	From, To     float64
	Taper        float64
	Tol          float64 // how far outside the inner edge a node may lie, a fraction; 0 for 0.05
	Exp          float64 // the exponent of the super-ellipse; 0 for 2.5
	NoOuterCheck bool
}

// full returns a round which turns the whole inner edge.
func full(c Point, a, b float64) round { return round{C: c, A: a, B: b, From: 0, To: 360} }

// share returns the share of the turn for a node at q: 1 on the inner edge
// within the span, falling to 0 over the taper, and 0 off the inner edge.
func (r round) share(q Point) float64 {
	u := sub(q, r.C)
	n := r.Exp
	if n == 0 {
		n = 2.5
	}
	tol := r.Tol
	if tol == 0 {
		tol = 0.05
	}
	rho := math.Pow(math.Pow(math.Abs(u.X/r.A), n)+math.Pow(math.Abs(u.Y/r.B), n), 1/n)
	if rho > 1+tol {
		return 0
	}
	if r.To-r.From >= 360 {
		return 1
	}
	th := math.Atan2(u.Y/r.B, u.X/r.A) * 180 / math.Pi
	// The distance of th from the span, in degrees.
	rel := math.Mod(th-r.From+720, 360)
	span := r.To - r.From
	if rel <= span {
		return 1
	}
	d := math.Min(rel-span, 360-rel)
	if r.Taper == 0 || d >= r.Taper {
		return 0
	}
	x := 1 - d/r.Taper
	return x * x * x * (x*(6*x-15) + 10)
}

// turned returns the path with the inner edges of the rounds turned by the
// angle a, in degrees.
func turned(path outline.Path, rounds []round, a float64) outline.Path {
	if len(rounds) == 0 || a == 0 {
		return path
	}
	// The ends of straight lines stay: stems stay vertical and bars
	// horizontal.
	fixed := map[Point]bool{}
	for _, c := range path {
		p0 := c.Start
		for _, s := range c.Segs {
			if !s.IsCurve {
				fixed[p0], fixed[s.P3] = true, true
			}
			p0 = s.P3
		}
	}
	return mapNodes(path, func(q Point) (Point, func(Point) Point) {
		if fixed[q] {
			return q, nil
		}
		var best float64
		var bestR round
		for _, r := range rounds {
			if s := r.share(q); s > best {
				best, bestR = s, r
			}
		}
		if best == 0 {
			return q, nil
		}
		b := a * best * math.Pi / 180
		return add(bestR.C, rotate(sub(q, bestR.C), b)), func(v Point) Point { return rotate(v, b) }
	})
}

// mapNodes returns the path with every node moved by f, which also returns
// the linear map of the vectors from the node to its two handles (nil for the
// identity).
func mapNodes(path outline.Path, f func(Point) (Point, func(Point) Point)) outline.Path {
	out := make(outline.Path, len(path))
	for ci, c := range path {
		segs := c.Segs
		n := len(segs)
		starts := make([]Point, n)
		for i := range segs {
			if i == 0 {
				starts[i] = c.Start
			} else {
				starts[i] = segs[i-1].P3
			}
		}
		type moved struct {
			p Point
			a func(Point) Point
		}
		apply := func(m func(Point) Point, v Point) Point {
			if m == nil {
				return v
			}
			return m(v)
		}
		ms := make([]moved, n) // the node at the end of each segment
		for i := range segs {
			p, a := f(segs[i].P3)
			ms[i] = moved{p, a}
		}
		ns := make([]outline.Segment, n)
		for i, s := range segs {
			prev := ms[(i-1+n)%n]
			if i == 0 && c.Start != segs[n-1].P3 {
				p, a := f(c.Start)
				prev = moved{p, a}
			}
			cur := ms[i]
			ns[i] = outline.Segment{IsCurve: s.IsCurve, P3: cur.p}
			if s.IsCurve {
				ns[i].P1 = add(prev.p, apply(prev.a, sub(s.P1, starts[i])))
				ns[i].P2 = add(cur.p, apply(cur.a, sub(s.P2, s.P3)))
			}
		}
		start := ms[n-1].p
		if c.Start != segs[n-1].P3 {
			start, _ = f(c.Start)
		}
		out[ci] = outline.Contour{Start: start, Segs: ns}
	}
	return out
}

// transformed returns the path with every point mapped by the affine map
// (x, y) -> (a x + b y + e, y).
func transformed(path outline.Path, a, b, e float64) outline.Path {
	m := func(q Point) Point { return pt(a*q.X+b*q.Y+e, q.Y) }
	out := make(outline.Path, len(path))
	for ci, c := range path {
		nc := outline.Contour{Start: m(c.Start), Segs: make([]outline.Segment, len(c.Segs))}
		for i, s := range c.Segs {
			nc.Segs[i] = outline.Segment{IsCurve: s.IsCurve, P1: m(s.P1), P2: m(s.P2), P3: m(s.P3)}
		}
		out[ci] = nc
	}
	return out
}

// minStem is the shortest vertical line which the lattice treats as a stem.
const minStem = 40

// onLattice returns the upright path with its stems prepared for the lattice
// of the slant num/den: every vertical line at least minStem long moves to
// the nearest whole x, its ends to the nearest heights which are multiples of
// den, and the handles of the curves which leave it along its line to such
// heights too; a node moves with its handles, and a horizontal line which
// ends at a moved node moves with it.  Nodes at the heights in fixed keep
// their height.
func onLattice(path outline.Path, den int, fixed []float64) outline.Path {
	m := float64(den)
	snapY := func(y float64) float64 {
		for _, f := range fixed {
			if math.Abs(y-f) < 1e-6 {
				return y
			}
		}
		return m * math.Round(y/m)
	}
	out := make(outline.Path, len(path))
	for ci, c := range path {
		segs := verticalLines(c.Segs)
		n := len(segs)
		closedAt := c.Start == segs[n-1].P3
		node := func(i int) Point { return segs[(i+n)%n].P3 } // the node at the end of segment i
		// move moves the node at the end of segment i by d, with its handles.
		move := func(i int, d Point) {
			i = (i + n) % n
			j := (i + 1) % n
			segs[i].P3 = add(segs[i].P3, d)
			if segs[i].IsCurve {
				segs[i].P2 = add(segs[i].P2, d)
			}
			if segs[j].IsCurve {
				segs[j].P1 = add(segs[j].P1, d)
			}
		}
		start := func(i int) Point { return node(i - 1) }
		vertical := func(i int) bool {
			i = (i + n) % n
			s := segs[i]
			p0 := start(i)
			return !s.IsCurve && math.Abs(s.P3.X-p0.X) < 1e-3 && math.Abs(s.P3.Y-p0.Y) >= minStem
		}
		horizontal := func(i int) bool {
			i = (i + n) % n
			s := segs[i]
			p0 := start(i)
			return !s.IsCurve && math.Abs(s.P3.Y-p0.Y) < 1e-6
		}
		done := map[int]bool{}
		for i := 0; i < n; i++ {
			if !vertical(i) {
				continue
			}
			x := start(i).X
			dx := math.Round(x) - x
			for _, k := range []int{i - 1, i} {
				k = (k + n) % n
				if done[k] {
					continue
				}
				done[k] = true
				q := node(k)
				d := pt(dx+x-q.X, snapY(q.Y)-q.Y)
				if vertical(k) && vertical(k+1) || d.X == 0 && d.Y == 0 {
					if d.X != 0 || d.Y != 0 {
						move(k, d)
					}
					continue
				}
				hIn, hOut := horizontal(k), horizontal(k+1)
				move(k, d)
				// A horizontal line at this node moves with it, and so does
				// the node at its far end.
				if d.Y != 0 {
					if hIn && !done[(k-1+n)%n] {
						move(k-1, pt(0, d.Y))
					}
					if hOut && !done[(k+1)%n] {
						move(k+1, pt(0, d.Y))
					}
				}
			}
			// The handles which leave the stem along its line.
			for _, k := range []int{i - 1, i + 1} {
				k = (k + n) % n
				s := &segs[k]
				if !s.IsCurve {
					continue
				}
				x0 := start(i).X
				if k == (i-1+n)%n {
					if math.Abs(s.P2.X-x0) < 1e-3 {
						y := snapY(s.P2.Y)
						if math.Abs(y-s.P3.Y) < m {
							y = s.P3.Y + math.Copysign(m, s.P2.Y-s.P3.Y)
						}
						s.P2 = pt(x0, y)
					}
				} else {
					p0 := segs[i].P3
					if math.Abs(s.P1.X-p0.X) < 1e-3 {
						y := snapY(s.P1.Y)
						if math.Abs(y-p0.Y) < m {
							y = p0.Y + math.Copysign(m, s.P1.Y-p0.Y)
						}
						s.P1 = pt(p0.X, y)
					}
				}
			}
		}
		// A smooth node between two curves where the edge runs vertically
		// and a handle is short becomes a node of the slant after the
		// shear, which the rounding turns: it moves to a lattice point,
		// and its handles to heights which are multiples of den, which
		// the rounding keeps.
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			a, b := &segs[i], &segs[j]
			if !a.IsCurve || !b.IsCurve {
				continue
			}
			q := a.P3
			if math.Abs(a.P2.X-q.X) > 1e-6 || math.Abs(b.P1.X-q.X) > 1e-6 || a.P2.Y == q.Y || b.P1.Y == q.Y {
				continue
			}
			if math.Min(dist(a.P2, q), dist(b.P1, q)) >= shortHandle {
				continue
			}
			move(i, pt(math.Round(q.X)-q.X, snapY(q.Y)-q.Y))
			q = a.P3
			for _, h := range []*Point{&a.P2, &b.P1} {
				y := m * math.Round(h.Y/m)
				if math.Abs(y-q.Y) < m {
					y = q.Y + math.Copysign(m, h.Y-q.Y)
				}
				*h = pt(q.X, y)
			}
		}
		st := c.Start
		if closedAt {
			st = segs[n-1].P3
		}
		out[ci] = outline.Contour{Start: st, Segs: segs}
	}
	return out
}

// mergeTol is the largest distance between a curve which replaces two and the
// two, when a node which is no extreme is removed.
const mergeTol = 0.6

// atExtremes returns the path with a node at every horizontal and vertical
// extreme of its curves, and without the smooth nodes between two curves
// which are no extremes where one curve can replace the two within mergeTol.
// The handles of the curves which leave a slanted stem of the slant k stay on
// the lattice: a split curve has its handle next to the stem moved along the
// stem to a height which is a multiple of den.
func atExtremes(path outline.Path, k float64, den int, cs []circle) outline.Path {
	out := make(outline.Path, len(path))
	for ci, c := range path {
		c = slidToExtremes(c)
		type piece struct {
			line    bool
			c       cubic
			extreme bool // the node at the end is an extreme
			split   bool // the piece comes from a split curve
		}
		var ps []piece
		p0 := c.Start
		nseg := len(c.Segs)
		stemLine := func(i int) bool {
			i = (i + nseg) % nseg
			s := c.Segs[i]
			q0 := c.Segs[(i-1+nseg)%nseg].P3
			return !s.IsCurve && onSlant(cubic{q0, q0, s.P3, s.P3}, k)
		}
		for si, s := range c.Segs {
			if !s.IsCurve {
				ps = append(ps, piece{line: true, c: cubic{p0, p0, s.P3, s.P3}})
				p0 = s.P3
				continue
			}
			cu := cubic{p0, s.P1, s.P2, s.P3}
			var ts []float64
			nearStem := stemLine(si-1) || stemLine(si+1)
			// An extreme within cornerReach of an end at a corner moves to
			// that end: the handle there turns to the axis.
			for _, t := range cu.extremes() {
				o, axis := overshoot(cu, t)
				if o <= extremeTol || o > cornerShift {
					continue
				}
				a, b := cu.split(t)
				switch {
				case a.length() < cornerReach && cornerAt(c.Segs, si-1):
					if axis == 0 {
						cu[1].X = cu[0].X
					} else {
						cu[1].Y = cu[0].Y
					}
				case b.length() < cornerReach && cornerAt(c.Segs, si):
					if axis == 0 {
						cu[2].X = cu[3].X
					} else {
						cu[2].Y = cu[3].Y
					}
				}
			}
			for _, t := range cu.extremes() {
				if o, axis := overshoot(cu, t); o <= extremeTol || axis == 0 && nearStem && o <= stemExtremeTol {
					continue
				}
				// No piece shorter than cornerReach next to a line, which
				// could not be merged away, and none next to a curve where
				// the extreme runs less than nearTol beyond the nodes: the
				// rounding turns such a piece.
				a, b := cu.split(t)
				o, _ := overshoot(cu, t)
				if a.length() < cornerReach && (!c.Segs[(si-1+nseg)%nseg].IsCurve || o <= nearTol) ||
					b.length() < cornerReach && (!c.Segs[(si+1)%nseg].IsCurve || o <= nearTol) {
					continue
				}
				ts = append(ts, t)
			}
			rest := cu
			prev := 0.0
			for _, t := range ts {
				u := (t - prev) / (1 - prev)
				a, b := rest.split(u)
				ps = append(ps, piece{c: a, extreme: true, split: true})
				rest, prev = b, t
			}
			ps = append(ps, piece{c: rest, extreme: isExtremeEnd(rest), split: len(ts) > 0})
			p0 = s.P3
		}
		// Remove the smooth nodes between curves which are no extremes.
		for changed := true; changed && len(ps) > 2; {
			changed = false
			bestI, bestDev := -1, mergeTol
			var bestC cubic
			for i := range ps {
				j := (i + 1) % len(ps)
				a, b := ps[i], ps[j]
				short := a.c.length() < cornerReach || b.c.length() < cornerReach
				if a.line || b.line || a.extreme && !short {
					continue
				}
				if turn(a.c, b.c) > 1 {
					continue
				}
				m, dev, ok := fit(a.c, b.c)
				if ok && maxOvershoot(m) > extremeTol {
					// An extreme close to an end at a corner moves to that
					// end: the handle there turns to the axis.
					k := (j + 1) % len(ps)
					h := (i - 1 + len(ps)) % len(ps)
					cornerAfter := ps[k].line || turn(b.c, ps[k].c) > 1
					cornerBefore := ps[h].line || turn(ps[h].c, a.c) > 1
					m, dev = toCorner(m, a.c, b.c, cornerBefore, cornerAfter)
				}
				if !ok || dev > bestDev || maxOvershoot(m) > extremeTol {
					continue
				}
				if nearCircle(cs, a.c[3]) && !sameCurvatureWithin(m, a.c, b.c, 0.1) {
					// Beside a ball the merged curve keeps the curvatures,
					// except at an end where a short piece meets a corner
					// (the tip of the wedge), which has no curvature to keep.
					k := (j + 1) % len(ps)
					cornerAfter := ps[k].line || turn(b.c, ps[k].c) > 1
					h := (i - 1 + len(ps)) % len(ps)
					cornerBefore := ps[h].line || turn(ps[h].c, a.c) > 1
					okStart := math.Abs(m.curvature(0)-a.c.curvature(0)) <= math.Max(0.003, 0.1*math.Abs(a.c.curvature(0)))
					okEnd := math.Abs(m.curvature(1)-b.c.curvature(1)) <= math.Max(0.003, 0.1*math.Abs(b.c.curvature(1)))
					if !(b.c.length() < cornerReach && cornerAfter && okStart || a.c.length() < cornerReach && cornerBefore && okEnd) {
						continue
					}
				}
				bestI, bestDev, bestC = i, dev, m
			}
			if bestI >= 0 {
				j := (bestI + 1) % len(ps)
				ps[bestI] = piece{c: bestC, extreme: ps[j].extreme}
				ps = append(ps[:j], ps[j+1:]...)
				changed = true
			}
		}
		// A smooth node which is no extreme and has a short handle slides
		// along the curve on that side until both handles are long enough.
		for i := range ps {
			j := (i + 1) % len(ps)
			a, b := ps[i], ps[j]
			if a.line || b.line || a.extreme || turn(a.c, b.c) > 1 {
				continue
			}
			if runOut(a.c) || runOut(b.c) {
				continue
			}
			if na, nb, ok := slide(a.c, b.c); ok {
				ps[i].c, ps[j].c = na, nb
			}
		}
		// Pieces with a short handle, which the rounding turns, are
		// drawn again with the handles closest to them.
		for i := range ps {
			if ps[i].split {
				ps[i].c = balanced(ps[i].c)
			}
		}
		// Handles next to a slanted stem back on the lattice.
		for i := range ps {
			if ps[i].line {
				continue
			}
			prev, next := ps[(i-1+len(ps))%len(ps)], ps[(i+1)%len(ps)]
			if prev.line && onSlant(prev.c, k) {
				ps[i].c[1] = latticeHandle(ps[i].c[0], ps[i].c[1], k, den)
				ps[i].c[2] = latticeInner(ps[i].c[0], ps[i].c[1], ps[i].c[2], k, den)
			}
			if next.line && onSlant(next.c, k) {
				ps[i].c[2] = latticeHandle(ps[i].c[3], ps[i].c[2], k, den)
				ps[i].c[1] = latticeInner(ps[i].c[3], ps[i].c[2], ps[i].c[1], k, den)
			}
		}
		nc := outline.Contour{Start: ps[len(ps)-1].c[3]}
		if len(ps) > 0 && c.Start != c.Segs[len(c.Segs)-1].P3 {
			nc.Start = c.Start
		}
		for _, q := range ps {
			if q.line {
				nc.Segs = append(nc.Segs, outline.Segment{P3: q.c[3]})
			} else {
				nc.Segs = append(nc.Segs, outline.Segment{IsCurve: true, P1: q.c[1], P2: q.c[2], P3: q.c[3]})
			}
		}
		out[ci] = nc
	}
	return out
}

// isExtremeEnd reports whether the curve runs horizontally or vertically at
// its end.
func isExtremeEnd(c cubic) bool {
	d := unit(sub(c[3], c[2]))
	return math.Abs(d.X) < 1e-6 || math.Abs(d.Y) < 1e-6
}

// turn returns the angle in degrees between the end direction of a and the
// start direction of b.
func turn(a, b cubic) float64 {
	u, v := unit(sub(a[3], a[2])), unit(sub(b[1], b[0]))
	return math.Abs(math.Atan2(cross(u, v), dot(u, v))) * 180 / math.Pi
}

// onSlant reports whether the line from c[0] to c[3] has the slant k and is
// at least minStem high.
func onSlant(c cubic, k float64) bool {
	d := sub(c[3], c[0])
	return math.Abs(d.Y) >= minStem && math.Abs(d.X-k*d.Y) < 1e-6*math.Abs(d.Y)
}

// latticeHandle returns the handle h of the node q on a stem of slant k moved
// along the stem's line to the nearest height which is a multiple of den, at
// least den from q.  A handle which does not lie on the line stays.
func latticeHandle(q, h Point, k float64, den int) Point {
	d := sub(h, q)
	if math.Abs(d.X-k*d.Y) > 1e-6*(1+math.Abs(d.Y)) {
		return h
	}
	m := float64(den)
	y := m * math.Round(h.Y/m)
	if math.Abs(y-q.Y) < m {
		y = q.Y + math.Copysign(m, d.Y)
	}
	return pt(q.X+k*(y-q.Y), y)
}

// overshoot returns how far the curve at t runs beyond its end points, and
// the axis (0 for x, 1 for y) in which it is extreme there.
func overshoot(c cubic, t float64) (float64, int) {
	d := c.deriv(t)
	axis := 0
	if math.Abs(d.Y) < math.Abs(d.X) {
		axis = 1
	}
	get := func(q Point) float64 {
		if axis == 0 {
			return q.X
		}
		return q.Y
	}
	v := get(c.at(t))
	lo, hi := math.Min(get(c[0]), get(c[3])), math.Max(get(c[0]), get(c[3]))
	return math.Max(v-hi, lo-v), axis
}

// shortened returns the upright path with the serifs on the obtuse side of a
// slanted stem shortened by d: at the foot those left of the ink, at the head
// those right of it.  A serif tip is a vertical line of at most tip which
// ends on the baseline or on the cap height; it moves towards the ink with
// the ends of the lines which meet it.
func shortened(path outline.Path, d, tip, capHeight float64) outline.Path {
	if d == 0 {
		return path
	}
	out := make(outline.Path, len(path))
	for ci, c := range path {
		segs := append([]outline.Segment(nil), c.Segs...)
		n := len(segs)
		node := func(i int) Point { return segs[(i+n)%n].P3 }
		for i := range segs {
			p0, p1 := node(i-1), node(i)
			if segs[i].IsCurve || math.Abs(p0.X-p1.X) > 1e-6 || math.Abs(p1.Y-p0.Y) > tip || math.Abs(p1.Y-p0.Y) < 5 {
				continue
			}
			lo, hi := math.Min(p0.Y, p1.Y), math.Max(p0.Y, p1.Y)
			foot := math.Abs(lo) < 1e-6
			head := math.Abs(hi-capHeight) < 1e-6
			if !foot && !head {
				continue
			}
			if !onStem(segs, i) {
				continue
			}
			mid := (lo + hi) / 2
			inkRight := inside(path, pt(p0.X+2, mid))
			inkLeft := inside(path, pt(p0.X-2, mid))
			var dx float64
			switch {
			case foot && inkRight && !inkLeft:
				dx = d
			case head && inkLeft && !inkRight:
				dx = -d
			default:
				continue
			}
			for _, k := range []int{i - 1, i} {
				k = (k + n) % n
				segs[k].P3.X += dx
				if segs[k].IsCurve {
					segs[k].P2.X += dx
				}
				if j := (k + 1) % n; segs[j].IsCurve {
					segs[j].P1.X += dx
				}
			}
		}
		out[ci] = outline.Contour{Start: segs[n-1].P3, Segs: segs}
	}
	return out
}

// inside reports whether q lies in the ink of the path, by the non-zero rule
// of the crossings of a horizontal line.
func inside(path outline.Path, q Point) bool {
	var k int
	for _, x := range path.CrossingsX(q.Y) {
		if x < q.X {
			k++
		}
	}
	return k%2 == 1
}

// cornerReach and cornerShift bound the extremes which move to a corner: no
// further along the curve from the corner, and running no further beyond the
// nodes.
const (
	nearTol     = 1
	cornerReach = 15
	cornerShift = 3
)

// cornerAt reports whether the node at the end of segment i of the closed
// list of segments is a corner.
func cornerAt(segs []outline.Segment, i int) bool {
	n := len(segs)
	i = (i + n) % n
	j := (i + 1) % n
	a, b := segs[i], segs[j]
	p0 := segs[(i-1+n)%n].P3
	u := sub(a.P3, p0)
	if a.IsCurve {
		u = sub(a.P3, a.P2)
	}
	v := sub(b.P3, a.P3)
	if b.IsCurve {
		v = sub(b.P1, a.P3)
	}
	return math.Abs(math.Atan2(cross(u, v), dot(u, v))) > math.Pi/180
}

// onStem reports whether the serif tip, segment i of the closed list, belongs
// to a serif of a vertical stem: on one side of it follow the flat of the
// serif, the bracket and a vertical line at least minStem long.
func onStem(segs []outline.Segment, i int) bool {
	n := len(segs)
	at := func(k int) (outline.Segment, Point) { k = ((k % n) + n) % n; return segs[k], segs[(k-1+n)%n].P3 }
	for _, dir := range []int{-1, 1} {
		f, f0 := at(i + dir)
		b, _ := at(i + 2*dir)
		v, v0 := at(i + 3*dir)
		if f.IsCurve || math.Abs(f.P3.Y-f0.Y) > 1e-6 || !b.IsCurve || v.IsCurve {
			continue
		}
		if math.Abs(v.P3.X-v0.X) < 1e-6 && math.Abs(v.P3.Y-v0.Y) >= minStem {
			return true
		}
	}
	return false
}

// shortHandle is the length below which balanced draws a curve again.
const shortHandle = 8

// balanced returns the curve, or, if one of its handles is shorter than
// shortHandle, the curve with the same ends and end directions which lies
// closest to it, if that has longer short handle and lies within mergeTol.
func balanced(c cubic) cubic {
	h := math.Min(dist(c[0], c[1]), dist(c[2], c[3]))
	if h >= shortHandle || h == 0 || runOut(c) {
		return c
	}
	a, b := c.split(0.5)
	m, dev, ok := fit(a, b)
	if ok && dev <= mergeTol && math.Min(dist(m[0], m[1]), dist(m[2], m[3])) >= shortHandle {
		return m
	}
	return c
}

// maxOvershoot returns how far the curve runs beyond its end points at most,
// in x or in y.
func maxOvershoot(c cubic) float64 {
	var o float64
	for _, t := range c.extremes() {
		v, _ := overshoot(c, t)
		o = math.Max(o, v)
	}
	return o
}

// axisDir reports whether v is horizontal or vertical.
func axisDir(v Point) bool { return math.Abs(v.X) < 1e-6 || math.Abs(v.Y) < 1e-6 }

// slide moves the node between the curves a and b, which meet smoothly,
// along the curve whose handle at the node is shorter than shortHandle, so
// that both handles at the node are at least shortHandle long; the curve on
// the other side takes up the piece in one cubic within mergeTol.  The result
// is false if the handles are long enough or no place is found.
func slide(a, b cubic) (cubic, cubic, bool) {
	ha, hb := dist(a[2], a[3]), dist(b[0], b[1])
	if math.Min(ha, hb) >= shortHandle {
		return a, b, false
	}
	for _, t := range []float64{0.05, 0.1, 0.15, 0.2, 0.3, 0.4} {
		if hb < ha {
			b0, b1 := b.split(t)
			m, dev, ok := fit(a, b0)
			if ok && dev <= mergeTol && dist(m[2], m[3]) >= shortHandle && dist(b1[0], b1[1]) >= shortHandle && maxOvershoot(m) <= extremeTol {
				return m, b1, true
			}
		} else {
			a0, a1 := a.split(1 - t)
			m, dev, ok := fit(a1, b)
			if ok && dev <= mergeTol && dist(m[0], m[1]) >= shortHandle && dist(a0[2], a0[3]) >= shortHandle && maxOvershoot(m) <= extremeTol {
				return a0, m, true
			}
		}
	}
	return a, b, false
}

// A circle is a circular part of an upright glyph, such as a ball terminal,
// which stays a circle in the italic.
type circle struct {
	C    Point
	R    float64
	Zone float64 // the width of the zone beyond the circle, over R, over which keptCircles lets its deformation fall to nothing
}

// keptCircles returns the upright path deformed such that the narrowing by w
// and the shear by k, which follow, only move the circles: within a circle
// (and circleTol beyond it) a point q moves to c + L(q-c), where L undoes the
// narrowing and the shear, and over the zone of one radius beyond the circle
// the move falls smoothly to nothing.  Each node moves by the deformation and
// its handles by its derivative there; smooth nodes stay smooth.
func keptCircles(path outline.Path, cs []circle, w, k float64) outline.Path {
	if len(cs) == 0 {
		return path
	}
	lin := func(v Point) Point { return pt((v.X-k*v.Y)/w, v.Y) }
	field := func(q Point) Point {
		for _, c := range cs {
			d := dist(q, c.C)
			if d >= (1+c.Zone)*c.R+circleTol {
				continue
			}
			s := 1.0
			if d > c.R+circleTol {
				x := 1 - (d-c.R-circleTol)/(c.Zone*c.R)
				s = x * x * x * (x*(6*x-15) + 10)
			}
			t := add(c.C, lin(sub(q, c.C)))
			return add(q, mul(sub(t, q), s))
		}
		return q
	}
	jac := func(q Point, v Point) Point {
		const h = 1e-3
		fx := mul(sub(field(add(q, pt(h, 0))), field(sub(q, pt(h, 0)))), 1/(2*h))
		fy := mul(sub(field(add(q, pt(0, h))), field(sub(q, pt(0, h)))), 1/(2*h))
		return add(mul(fx, v.X), mul(fy, v.Y))
	}
	onCircle := func(q Point) bool {
		for _, c := range cs {
			if math.Abs(dist(q, c.C)-c.R) <= circleTol {
				return true
			}
		}
		return false
	}
	out := make(outline.Path, len(path))
	for ci, c := range path {
		segs := c.Segs
		n := len(segs)
		// corner reports whether the node at the end of segment i is a
		// corner.
		corner := func(i int) bool { return cornerAt(segs, i) }
		ns := make([]outline.Segment, n)
		for i, sg := range segs {
			p0 := segs[(i-1+n)%n].P3
			ns[i] = outline.Segment{IsCurve: sg.IsCurve, P3: field(sg.P3)}
			if !sg.IsCurve {
				continue
			}
			// The handle at a corner on a circle of a curve which leaves the
			// circle keeps its direction in the italic: the curve, not the
			// ball, takes the shear.
			h1, h2 := sub(sg.P1, p0), sub(sg.P2, sg.P3)
			if !(corner((i-1+n)%n) && onCircle(p0) && !onCircle(sg.P3)) {
				h1 = jac(p0, h1)
			}
			if !(corner(i) && onCircle(sg.P3) && !onCircle(p0)) {
				h2 = jac(sg.P3, h2)
			}
			ns[i].P1 = add(field(p0), h1)
			ns[i].P2 = add(field(sg.P3), h2)
			// A curve which the deformation bends is drawn again closest to
			// its image, with those directions at its ends.
			if field(sg.P1) != sg.P1 || field(sg.P2) != sg.P2 || field(p0) != p0 || field(sg.P3) != sg.P3 {
				cu := cubic{p0, sg.P1, sg.P2, sg.P3}
				qs := make([]Point, 33)
				for k := range qs {
					qs[k] = field(cu.at(float64(k) / 32))
				}
				if m, dev, ok := fitPoints(field(p0), unit(h1), field(sg.P3), unit(mul(h2, -1)), qs); ok && dev <= 0.5 {
					ns[i].P1, ns[i].P2 = m[1], m[2]
				}
			}
		}
		out[ci] = outline.Contour{Start: ns[n-1].P3, Segs: ns}
	}
	return out
}

// circleTol is how far from a circle a node may lie and count as on it.
const circleTol = 0.6

// circlesOf returns the circles of radius r (within 1.5) which the upright
// path contains: the circles through the two ends and the middle of curves
// with the handles of a circular arc, at least one quarter of a circle in
// all, with their centres merged within 1 unit.
func circlesOf(path outline.Path, r float64) []circle {
	type acc struct {
		c   Point
		sum float64
	}
	var as []acc
	for _, ct := range path {
		p0 := ct.Start
		for _, s := range ct.Segs {
			if s.IsCurve {
				cu := cubic{p0, s.P1, s.P2, s.P3}
				c, rad, ok := circleThrough(cu[0], cu.at(0.5), cu[3])
				if ok && math.Abs(rad-r) < 1.5 && math.Abs(dist(cu.at(0.25), c)-rad) < 0.3 && math.Abs(dist(cu.at(0.75), c)-rad) < 0.3 {
					ang := math.Abs(math.Atan2(cross(sub(cu[0], c), sub(cu[3], c)), dot(sub(cu[0], c), sub(cu[3], c))))
					found := false
					for i := range as {
						if dist(as[i].c, c) < 1 {
							as[i].sum += ang
							found = true
						}
					}
					if !found {
						as = append(as, acc{c, ang})
					}
				}
			}
			p0 = s.P3
		}
	}
	var out []circle
	for _, a := range as {
		if a.sum >= math.Pi/2-1e-3 {
			out = append(out, circle{C: a.c, R: r})
		}
	}
	return out
}

// circleThrough returns the centre and radius of the circle through a, b, c.
func circleThrough(a, b, c Point) (Point, float64, bool) {
	d := 2 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	if math.Abs(d) < 1e-9 {
		return Point{}, 0, false
	}
	a2, b2, c2 := dot(a, a), dot(b, b), dot(c, c)
	o := pt((a2*(b.Y-c.Y)+b2*(c.Y-a.Y)+c2*(a.Y-b.Y))/d, (a2*(c.X-b.X)+b2*(a.X-c.X)+c2*(b.X-a.X))/d)
	return o, dist(o, a), true
}

// slideReach is how far along a curve from a smooth node an extreme may lie
// for the node to move there, and slideTol how far the curve which then
// replaces the rest of that curve and the curve beyond the node may lie from
// them.
const (
	slideReach = 20
	slideTol   = 0.75
)

// slidToExtremes returns the contour in which every smooth node between two
// curves which is no extreme, but has an extreme of one of its curves within
// slideReach, has moved to that extreme: the curve is split there, and its
// rest and the curve beyond the node become one curve.  This is the node at
// an extreme of the upright roman which the shear has moved off the extreme.
func slidToExtremes(c outline.Contour) outline.Contour {
	segs := append([]outline.Segment(nil), c.Segs...)
	n := len(segs)
	cub := func(i int) cubic {
		i = (i + n) % n
		s := segs[i]
		return cubic{segs[(i-1+n)%n].P3, s.P1, s.P2, s.P3}
	}
	set := func(i int, q cubic) {
		i = (i + n) % n
		segs[i] = outline.Segment{IsCurve: true, P1: q[1], P2: q[2], P3: q[3]}
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		if !segs[i].IsCurve || !segs[j].IsCurve {
			continue
		}
		a, b := cub(i), cub(j)
		if turn(a, b) > 1 || axisDir(sub(a[3], a[2])) {
			continue
		}

		done := false
		// An extreme of a near its end.
		ts := a.extremes()
		if len(ts) > 0 {
			te := ts[len(ts)-1]
			a0, a1 := a.split(te)
			if a1.length() < slideReach {
				if m, dev, ok := slideFit(a1, b); ok && dev <= slideTol && maxOvershoot(m) <= extremeTol && !shortBeside(a0, segs[(i-1+n)%n]) {
					set(i, a0)
					set(j, m)
					done = true
				}
			}
		}
		if done {
			continue
		}
		// An extreme of b near its start.
		ts = b.extremes()
		if len(ts) > 0 {
			te := ts[0]
			b0, b1 := b.split(te)
			if b0.length() < slideReach {
				if m, dev, ok := slideFit(a, b0); ok && dev <= slideTol && maxOvershoot(m) <= extremeTol && !shortBeside(b1, segs[(j+1)%n]) {
					set(i, m)
					set(j, b1)
				}
			}
		}
	}
	return outline.Contour{Start: segs[n-1].P3, Segs: segs}
}

// curvature returns the signed curvature of the curve at t.
func (c cubic) curvature(t float64) float64 {
	const h = 1e-4
	d := c.deriv(t)
	dd := mul(sub(c.deriv(math.Min(1, t+h)), c.deriv(math.Max(0, t-h))), 1/(math.Min(1, t+h)-math.Max(0, t-h)))
	l := norm(d)
	if l == 0 {
		return 0
	}
	return cross(d, dd) / (l * l * l)
}

// sameCurvature reports whether the curve m, which replaces a and b, has at
// its ends the curvatures of a at its start and of b at its end, within a
// third or 0.003.
func sameCurvature(m, a, b cubic) bool { return sameCurvatureWithin(m, a, b, 1.0/3) }

// sameCurvatureWithin is sameCurvature within the share r, or 0.003.
func sameCurvatureWithin(m, a, b cubic, r float64) bool {
	ok := func(x, y float64) bool {
		return math.Abs(x-y) <= math.Max(0.003*r*3, math.Max(math.Abs(x), math.Abs(y))*r)
	}
	return ok(m.curvature(0), a.curvature(0)) && ok(m.curvature(1), b.curvature(1))
}

// verticalLines returns the closed list of segments with every curve whose
// control points lie on one vertical line made a line, and consecutive
// vertical lines on one line joined: a stem is one line.
func verticalLines(in []outline.Segment) []outline.Segment {
	n := len(in)
	segs := make([]outline.Segment, n)
	for i, s := range in {
		p0 := in[(i-1+n)%n].P3
		if s.IsCurve && math.Abs(s.P1.X-p0.X) < 1e-3 && math.Abs(s.P2.X-p0.X) < 1e-3 && math.Abs(s.P3.X-p0.X) < 1e-3 &&
			(s.P1.Y-p0.Y)*(s.P3.Y-p0.Y) >= 0 && (s.P2.Y-p0.Y)*(s.P3.Y-p0.Y) >= 0 {
			s = outline.Segment{P3: s.P3}
		}
		segs[i] = s
	}
	vert := func(i int) bool {
		s := segs[i]
		p0 := segs[(i-1+len(segs))%len(segs)].P3
		return !s.IsCurve && math.Abs(s.P3.X-p0.X) < 1e-3
	}
	for i := 0; len(segs) > 3 && i < len(segs); {
		j := (i + 1) % len(segs)
		if vert(i) && vert(j) {
			pi := segs[(i-1+len(segs))%len(segs)].P3
			if (segs[i].P3.Y-pi.Y)*(segs[j].P3.Y-segs[i].P3.Y) > 0 {
				segs[i].P3 = segs[j].P3
				segs = append(segs[:j], segs[j+1:]...)
				if j < i {
					i--
				}
				continue
			}
		}
		i++
	}
	return segs
}

// nearCircle reports whether q lies on one of the circles or in the zone
// beyond it over which keptCircles deforms the outline.
func nearCircle(cs []circle, q Point) bool {
	for _, c := range cs {
		if dist(q, c.C) < (1+c.Zone)*c.R+circleTol {
			return true
		}
	}
	return false
}

// fitG2 returns the cubic from a[0] to b[3] with the directions and the
// curvatures of a at its start and of b at its end, and the largest distance
// between it and a and b, which meet at a smooth node.  The handle lengths
// solve the two equations of the end curvatures by Newton's method from those
// of fit.  The result is false if fit fails or Newton's method does not
// converge to positive handles.
func fitG2(a, b cubic) (cubic, float64, bool) {
	m, _, ok := fit(a, b)
	if !ok {
		return cubic{}, 0, false
	}
	p0, p3 := a[0], b[3]
	t0, t3 := unit(sub(a[1], a[0])), unit(sub(b[3], b[2]))
	k0, k3 := a.curvature(0), b.curvature(1)
	h0, h3 := dist(m[0], m[1]), dist(m[2], m[3])
	D := sub(p3, p0)
	x := cross(t0, t3)
	conv := false
	for it := 0; it < 40; it++ {
		f0 := 2.0/3*(cross(t0, D)-h3*x)/(h0*h0) - k0
		f3 := 2.0/3*(cross(D, t3)-h0*x)/(h3*h3) - k3
		if math.Abs(f0) < 1e-8 && math.Abs(f3) < 1e-8 {
			conv = true
			break
		}
		a00 := -4.0 / 3 * (cross(t0, D) - h3*x) / (h0 * h0 * h0)
		a01 := -2.0 / 3 * x / (h0 * h0)
		a10 := -2.0 / 3 * x / (h3 * h3)
		a11 := -4.0 / 3 * (cross(D, t3) - h0*x) / (h3 * h3 * h3)
		det := a00*a11 - a01*a10
		if math.Abs(det) < 1e-14 {
			break
		}
		h0 -= (f0*a11 - f3*a01) / det
		h3 -= (a00*f3 - a10*f0) / det
		if h0 <= 0 || h3 <= 0 {
			return cubic{}, 0, false
		}
	}
	if !conv {
		return cubic{}, 0, false
	}
	g := cubic{p0, add(p0, mul(t0, h0)), sub(p3, mul(t3, h3)), p3}
	var dev float64
	for _, c := range []cubic{a, b} {
		for i := 0; i <= 32; i++ {
			dev = math.Max(dev, distToCubic(g, c.at(float64(i)/32)))
		}
	}
	return g, dev, true
}

// slideFit returns the curve which replaces a and b when a node slides to an
// extreme: the curve of fitG2 if it lies within slideTol, otherwise that of
// fit if it keeps the end curvatures (sameCurvature).
func slideFit(a, b cubic) (cubic, float64, bool) {
	if m, dev, ok := fitG2(a, b); ok && dev <= slideTol && maxOvershoot(m) <= extremeTol {
		return m, dev, true
	}
	m, dev, ok := fit(a, b)
	if ok && sameCurvature(m, a, b) {
		return m, dev, true
	}
	return cubic{}, 0, false
}

// shortBeside reports whether the curve c is shorter than cornerReach and
// the segment beyond it is a line.
func shortBeside(c cubic, beyond outline.Segment) bool {
	return !beyond.IsCurve && c.length() < cornerReach
}

// circleG2Tol is how far a curve may move when circleG2 gives it the
// curvature of the circle it runs into.
const circleG2Tol = 2.5

// circleG2 returns the path in which every curve which runs smoothly into a
// curve on one of the circles, coming from off the circle, has the curvature
// of that curve at the node where they meet: its handle at the far end
// lengthens or shortens along its direction, if the curve then moves by at
// most circleG2Tol.  The shear bends the long curves which leave a ball more
// near the ball, which keeps its circle, than the roman does.
func circleG2(path outline.Path, cs []circle) outline.Path {
	if len(cs) == 0 {
		return path
	}
	on := func(q Point) bool {
		for _, c := range cs {
			if math.Abs(dist(q, c.C)-c.R) <= circleTol {
				return true
			}
		}
		return false
	}
	out := make(outline.Path, len(path))
	for ci, c := range path {
		segs := append([]outline.Segment(nil), c.Segs...)
		n := len(segs)
		cub := func(i int) cubic {
			i = (i + n) % n
			p0 := segs[(i-1+n)%n].P3
			s := segs[i]
			return cubic{p0, s.P1, s.P2, s.P3}
		}
		for i := range segs {
			if !segs[i].IsCurve {
				continue
			}
			cu := cub(i)
			for _, atEnd := range []bool{true, false} {
				j := i + 1
				node := cu[3]
				far := cu[0]
				if !atEnd {
					j = i - 1
					node, far = cu[0], cu[3]
				}
				nb := segs[(j+n)%n]
				if !nb.IsCurve || !on(node) || on(far) {
					continue
				}
				nc := cub(j)
				if !on(nc.at(0.5)) {
					continue
				}
				var want float64
				if atEnd {
					if turn(cu, nc) > 1 {
						continue
					}
					want = nc.curvature(0)
				} else {
					if turn(nc, cu) > 1 {
						continue
					}
					want = nc.curvature(1)
				}
				orig := cu
				// The handles at both ends move along their directions; of
				// the pairs of lengths which give the curvature, the one
				// which moves the curve least is taken.
				d0, d3 := unit(sub(orig[1], orig[0])), unit(sub(orig[2], orig[3]))
				l0, l3 := dist(orig[0], orig[1]), dist(orig[3], orig[2])
				try := func(a, b float64) cubic {
					return cubic{orig[0], add(orig[0], mul(d0, a)), add(orig[3], mul(d3, b)), orig[3]}
				}
				kAt := func(m cubic) float64 {
					if atEnd {
						return m.curvature(1)
					}
					return m.curvature(0)
				}
				if math.Abs(kAt(orig)-want) <= 0.1*math.Abs(want) {
					continue
				}
				best, bestDev := orig, circleG2Tol
				for f := 0.5; f <= 2.0; f += 0.05 {
					// The far handle at f times its length, the near one by
					// bisection between 0.3 and 3 times its length.
					var lo, hi float64
					var mk func(x float64) cubic
					if atEnd {
						mk = func(x float64) cubic { return try(f*l0, x) }
						lo, hi = 0.3*l3, 3*l3
					} else {
						mk = func(x float64) cubic { return try(x, f*l3) }
						lo, hi = 0.3*l0, 3*l0
					}
					flo, fhi := kAt(mk(lo))-want, kAt(mk(hi))-want
					if flo*fhi > 0 {
						continue
					}
					for it := 0; it < 40; it++ {
						mid := (lo + hi) / 2
						if fm := kAt(mk(mid)) - want; fm*flo <= 0 {
							hi = mid
						} else {
							lo, flo = mid, fm
						}
					}
					m := mk((lo + hi) / 2)
					var dev float64
					for k := 0; k <= 32; k++ {
						dev = math.Max(dev, distToCubic(orig, m.at(float64(k)/32)))
						dev = math.Max(dev, distToCubic(m, orig.at(float64(k)/32)))
					}
					if dev < bestDev && len(m.extremes()) == len(orig.extremes()) {
						best, bestDev = m, dev
					}
				}
				if best == orig {
					continue
				}
				m := best
				cu = m
				segs[i].P1, segs[i].P2 = m[1], m[2]
			}
		}
		out[ci] = outline.Contour{Start: segs[n-1].P3, Segs: segs}
	}
	return out
}

// toCorner returns m, which replaces the curves a and b, with the extreme
// which lies within cornerReach of an end at a corner moved to that end (the
// handle there turned to the axis), and its largest distance from a and b.
func toCorner(m, a, b cubic, before, after bool) (cubic, float64) {
	for _, t := range m.extremes() {
		o, axis := overshoot(m, t)
		if o <= extremeTol || o > cornerShift {
			continue
		}
		p, q := m.split(t)
		switch {
		case after && q.length() < cornerReach:
			if axis == 0 {
				m[2].X = m[3].X
			} else {
				m[2].Y = m[3].Y
			}
		case before && p.length() < cornerReach:
			if axis == 0 {
				m[1].X = m[0].X
			} else {
				m[1].Y = m[0].Y
			}
		}
	}
	var dev float64
	for _, c := range []cubic{a, b} {
		for i := 0; i <= 32; i++ {
			dev = math.Max(dev, distToCubic(m, c.at(float64(i)/32)))
		}
	}
	return m, dev
}

// runOut reports whether the curve runs out of a straight line without
// curvature: its first three or its last three control points lie on one
// line, as where a curve leaves a stem along it.
func runOut(c cubic) bool {
	col := func(a, b, d Point) bool {
		u, v := sub(b, a), sub(d, a)
		l := norm(u) * norm(v)
		return l > 0 && math.Abs(cross(u, v)) <= 1e-6*l
	}
	return col(c[0], c[1], c[2]) || col(c[3], c[2], c[1])
}

// latticeInner returns the second inner control point h of a curve which
// runs out of a slanted stem of the slant k at q, its first inner control
// point g already on the lattice of the stem's line: where h lies on that
// line too (the run-out), it moves along it to the nearest height which is a
// multiple of den beyond g; the rounding keeps the curve without
// curvature where it leaves the stem.  If not, h stays.
func latticeInner(q, g, h Point, k float64, den int) Point {
	d := sub(h, q)
	if math.Abs(d.X-k*d.Y) > 0.05 || d.Y == 0 {
		return h
	}
	m := float64(den)
	y := m * math.Round(h.Y/m)
	if (y-g.Y)*(g.Y-q.Y) <= 0 {
		y = g.Y + math.Copysign(m, g.Y-q.Y)
	}
	return pt(q.X+k*(y-q.Y), y)
}
