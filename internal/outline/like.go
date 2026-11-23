// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package outline

import (
	"fmt"
	"math"
)

// RoundLike returns the path rounded to integer coordinates with the nodes
// of the model ref: ref.Round() and the result have the same contours, each
// with the same segments of the same kinds from the corresponding start.
// The path and the model are outlines of one shape for different measures,
// with the same contours in the same order and direction.
//
// The nodes of the path are laid over those of ref.Merged(), contour by
// contour (Like), and the result is rounded as Round rounds a path whose
// short segments are merged.
func (p Path) RoundLike(ref Path) (Path, error) {
	q, err := p.Like(ref)
	if err != nil {
		return nil, err
	}
	return q.RoundNodes(), nil
}

// Like returns the path with the nodes of ref.Merged(), the model, not
// rounded.
//
// The nodes of each contour are laid over those of the contour of the
// model, both in the frame of the bounding box of their path, in the order
// of the contour and from the start which fits best, so that the distances
// between partners are small and few nodes are left without a partner
// (likeGap), least of all a node which cannot be removed (likeFixed), and
// partners on smooth runs run in like directions (likeTurn).  A corner is
// the partner of a corner only.  A node of the path without a partner is
// removed: the two segments which meet there are replaced by the curve of
// which both are parts, by one line if they are lines in one direction, or
// by the curve with which Round replaces a short segment and its neighbour.
// A node of the model without a partner is added: the segment of the path
// which runs past it is cut near the same share of its length, where it
// runs most nearly in the direction of the model at the node (cutSegment).
// A line where the model has a curve becomes a straight curve, and a curve
// which lies within likeStraight of the line between its ends becomes that
// line where the model has a line.  The contour then starts at the partner
// of the start of the model.
//
// The shape changes only where two segments are replaced by a curve which
// Round would use, by at most MergeTolerance, and where a curve becomes a
// line, by less than likeStraight.  The result is an error if the paths
// differ in the number or the direction of their contours, if a node cannot
// be removed, if a corner has no partner, if partners on smooth runs run
// more than likeApart apart, or if a curve which is not straight stands
// where the model has a line.
func (p Path) Like(ref Path) (Path, error) {
	if len(p) != len(ref) {
		return nil, fmt.Errorf("outline: %d contours where the model has %d", len(p), len(ref))
	}
	refM := ref.Merged()
	fp, fr := frameOf(p), frameOf(refM)
	out := make(Path, len(p))
	for i := range p {
		if p[i].SignedArea()*refM[i].SignedArea() < 0 {
			return nil, fmt.Errorf("outline: contour %d runs against the direction of the model", i)
		}
		c, err := likeContour(p[i], fp, refM[i], fr)
		if err != nil {
			return nil, fmt.Errorf("outline: contour %d: %w", i, err)
		}
		out[i] = c
	}
	return out, nil
}

// likeGap is the cost of a node without a partner in Like, as a distance in
// the frame of the bounding box, whose sides have the length 1; the cost of
// a pair of partners is their distance there.  likeFixed is added to the
// cost of a node of the path without a partner which cannot be removed
// alone, and to the distance of two partners of which one is a corner and
// the other a smooth node.
const (
	likeGap   = 0.15
	likeFixed = 1.0
)

// likeApart is the largest angle, in degrees, between the directions in
// which the path and the model leave two partners on smooth runs.
const likeApart = 45.0

// likeTurn is added to the distance of two partners which are not corners
// and at which the outlines run at right angles to each other or further
// apart, and the share of it for a smaller angle.
const likeTurn = 0.10

// cutMargin is the share of the distance to the neighbouring cuts which a
// cut of cutSegment keeps from them, and cutPull the angle, in degrees, by
// which a place at the far end of the range must run closer to the wanted
// direction than the place at the share of the length.
const (
	cutMargin = 0.1
	cutPull   = 5.0
)

// likeCornerAngle is the turn, in degrees, from which a node counts as a
// corner in Like, and likeSmoothAngle the turn below which it counts as
// smooth.
const (
	likeCornerAngle = 20.0
	likeSmoothAngle = 5.0
)

// A likeFrame maps the points of a path to the unit square of its nodes.
type likeFrame struct{ x0, y0, w, h float64 }

func frameOf(p Path) likeFrame {
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, c := range p {
		for _, q := range append([]Point{c.Start}, ends(c)...) {
			x0, y0, x1, y1 = math.Min(x0, q.X), math.Min(y0, q.Y), math.Max(x1, q.X), math.Max(y1, q.Y)
		}
	}
	return likeFrame{x0, y0, math.Max(x1-x0, 1), math.Max(y1-y0, 1)}
}

func ends(c Contour) []Point {
	ps := make([]Point, len(c.Segs))
	for i, s := range c.Segs {
		ps[i] = s.P3
	}
	return ps
}

func (f likeFrame) of(q Point) Point { return Point{(q.X - f.x0) / f.w, (q.Y - f.y0) / f.h} }

// A ring is a closed contour as a cycle: segment k runs from node k to node
// k+1, and the last segment ends at node 0.  The line which closes a contour
// that does not end at its start is a segment of the ring (implicit).
type ring struct {
	nodes    []Point
	segs     []Segment
	implicit bool
}

func ringOf(c Contour) ring {
	r := ring{segs: append([]Segment(nil), c.Segs...)}
	if !c.isClosed() {
		r.segs = append(r.segs, Segment{P3: c.Start})
		r.implicit = true
	}
	n := len(r.segs)
	r.nodes = make([]Point, n)
	r.nodes[0] = c.Start
	for k := 1; k < n; k++ {
		r.nodes[k] = r.segs[k-1].P3
	}
	if n > 0 {
		r.segs[n-1].P3 = c.Start
	}
	return r
}

// The kinds of a node in Like: the outline does not turn there (by less
// than likeSmoothAngle), turns sharply (by likeCornerAngle or more), or
// turns a little.
const (
	nodeSmooth = iota
	nodeSoft
	nodeCorner
)

// kind returns the kind of node k of the ring.
func (r ring) kind(k int) int {
	n := len(r.segs)
	in, out := &r.segs[(k-1+n)%n], &r.segs[k]
	din, _ := arriving(in, r.nodes[(k-1+n)%n])
	dout, _ := leaving(out, r.nodes[k])
	if din == (Point{}) || dout == (Point{}) {
		return nodeSoft
	}
	switch a := math.Abs(turnAngle(din, dout)); {
	case a >= likeCornerAngle:
		return nodeCorner
	case a >= likeSmoothAngle:
		return nodeSoft
	}
	return nodeSmooth
}

// likeStraight is the largest distance, in font units, of the control points
// of a curve from the line between its ends at which Like takes the curve
// for that line where the model has a line.
const likeStraight = 0.25

// nearlyStraight reports whether the control points of the segment, which
// starts at p0, lie within likeStraight of the line from p0 to its end and
// between its ends.
func nearlyStraight(s Segment, p0 Point) bool {
	d := sub(s.P3, p0)
	l := math.Hypot(d.X, d.Y)
	if l == 0 {
		return false
	}
	for _, q := range []Point{s.P1, s.P2} {
		e := sub(q, p0)
		if math.Abs(d.X*e.Y-d.Y*e.X)/l > likeStraight {
			return false
		}
		if along := (d.X*e.X + d.Y*e.Y) / l; along < 0 || along > l {
			return false
		}
	}
	return true
}

// leave returns the direction in which the ring leaves node k, or the zero
// vector if it has none.
func (r ring) leave(k int) Point {
	d, _ := leaving(&r.segs[k], r.nodes[k])
	if l := math.Hypot(d.X, d.Y); l > 0 {
		return Point{d.X / l, d.Y / l}
	}
	return Point{}
}

// likeContour gives the contour c the nodes of the contour ref.
func likeContour(c Contour, fc likeFrame, ref Contour, fr likeFrame) (Contour, error) {
	if sameSegments(c, ref) {
		return Contour{Start: c.Start, Segs: append([]Segment(nil), c.Segs...)}, nil
	}
	a, b := ringOf(c), ringOf(ref)
	n, m := len(a.nodes), len(b.nodes)
	if n < 2 || m < 2 {
		return Contour{}, fmt.Errorf("a contour of fewer than two segments")
	}
	pa, pb := make([]Point, n), make([]Point, m)
	ca, cb := make([]int, n), make([]int, m)
	da, db := make([]Point, n), make([]Point, m)
	for i := range pa {
		pa[i], ca[i], da[i] = fc.of(a.nodes[i]), a.kind(i), a.leave(i)
	}
	for j := range pb {
		pb[j], cb[j], db[j] = fr.of(b.nodes[j]), b.kind(j), b.leave(j)
	}
	// A node of a which cannot be removed is dearer to leave without a
	// partner than one which can.
	gapA := make([]float64, n)
	for i := range gapA {
		gapA[i] = likeGap
		prev := (i - 1 + n) % n
		if _, err := mergeRun(a.nodes[prev], []Segment{a.segs[prev], a.segs[i]}); err != nil {
			gapA[i] += likeFixed
		}
	}
	pair := func(i, j int) float64 {
		d := math.Hypot(pa[i].X-pb[j].X, pa[i].Y-pb[j].Y)
		switch {
		case ca[i]+cb[j] == nodeSmooth+nodeCorner && ca[i] != cb[j]:
			// A corner is no partner of a smooth node.
			d += likeFixed
		case ca[i] == nodeSmooth && cb[j] == nodeSmooth && da[i] != (Point{}) && db[j] != (Point{}):
			d += likeTurn * math.Min(1, math.Abs(turnAngle(da[i], db[j]))/90)
		}
		return d
	}

	// The nodes of a from the node r on are laid over those of b in order;
	// the start r with the least cost is taken.
	cost := make([][]float64, n+1)
	for i := range cost {
		cost[i] = make([]float64, m+1)
	}
	fill := func(r int) float64 {
		for i := n; i >= 0; i-- {
			for j := m; j >= 0; j-- {
				switch {
				case i == n && j == m:
					cost[i][j] = 0
				case i == n:
					cost[i][j] = cost[i][j+1] + likeGap
				case j == m:
					cost[i][j] = cost[i+1][j] + gapA[(i+r)%n]
				default:
					cost[i][j] = math.Min(pair((i+r)%n, j)+cost[i+1][j+1],
						math.Min(cost[i+1][j]+gapA[(i+r)%n], cost[i][j+1]+likeGap))
				}
			}
		}
		return cost[0][0]
	}
	best, bestCost := 0, math.Inf(1)
	for r := 0; r < n; r++ {
		if v := fill(r); v < bestCost-1e-12 {
			best, bestCost = r, v
		}
	}
	fill(best)
	type partners struct{ i, j int } // node i of a and node j of b
	var ps []partners
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case cost[i][j] == pair((i+best)%n, j)+cost[i+1][j+1]:
			ps = append(ps, partners{(i + best) % n, j})
			i, j = i+1, j+1
		case cost[i][j] == cost[i+1][j]+gapA[(i+best)%n]:
			i++
		default:
			j++
		}
	}
	if len(ps) < 2 {
		return Contour{}, fmt.Errorf("the nodes do not correspond to those of the model")
	}
	for _, q := range ps {
		if ca[q.i]+cb[q.j] == nodeSmooth+nodeCorner && ca[q.i] != cb[q.j] {
			return Contour{}, fmt.Errorf("the node of the model at (%.1f, %.1f) has no partner of its kind, a corner or a smooth node", b.nodes[q.j].X, b.nodes[q.j].Y)
		}
		if ca[q.i] == nodeSmooth && cb[q.j] == nodeSmooth && da[q.i] != (Point{}) && db[q.j] != (Point{}) &&
			math.Abs(turnAngle(da[q.i], db[q.j])) > likeApart {
			return Contour{}, fmt.Errorf("the node of the model at (%.1f, %.1f) has no partner at which the path runs in its direction", b.nodes[q.j].X, b.nodes[q.j].Y)
		}
	}

	// Between two partners and the next two, the segments of a are replaced
	// by as many as b has there.
	var segs []Segment
	var first Point
	for k, q := range ps {
		next := ps[(k+1)%len(ps)]
		na := (next.i-q.i+n-1)%n + 1 // the number of segments of a in the run
		nb := (next.j-q.j+m-1)%m + 1
		run := make([]Segment, na)
		for l := range run {
			run[l] = a.segs[(q.i+l)%n]
		}
		from := a.nodes[q.i]
		if na != nb {
			one, err := mergeRun(from, run)
			if err != nil {
				return Contour{}, err
			}
			shares := make([]float64, nb)
			var total float64
			for l := range shares {
				total += arcLength(asCubic(b.segs[(q.j+l)%m], b.nodes[(q.j+l)%m]))
				shares[l] = total
			}
			for l := range shares {
				shares[l] /= total
			}
			dirs := make([]Point, nb-1)
			for l := range dirs {
				k := (q.j + l + 1) % m
				if cb[k] == nodeCorner {
					return Contour{}, fmt.Errorf("the corner of the model at (%.1f, %.1f) has no partner", b.nodes[k].X, b.nodes[k].Y)
				}
				if cb[k] == nodeSmooth {
					dirs[l] = db[k]
				}
			}
			run = cutSegment(from, one, shares[:nb-1], dirs)
		}
		at := from
		for l := range run {
			want := b.segs[(q.j+l)%m].IsCurve
			switch {
			case want && !run[l].IsCurve:
				cu := asCubic(run[l], at)
				run[l] = Segment{IsCurve: true, P1: cu[1], P2: cu[2], P3: cu[3]}
			case !want && run[l].IsCurve:
				if !nearlyStraight(run[l], at) {
					return Contour{}, fmt.Errorf("a curve to (%.1f, %.1f) where the model has a line", run[l].P3.X, run[l].P3.Y)
				}
				run[l] = Segment{P3: run[l].P3}
			}
			at = run[l].P3
		}
		if k == 0 {
			first = from
		}
		segs = append(segs, run...)
	}

	// The contour starts at the partner of node 0 of b.
	if len(segs) != m {
		return Contour{}, fmt.Errorf("the nodes do not correspond to those of the model")
	}
	// Segment k of the result stands for segment ps[0].j+k of b.
	s := (m - ps[0].j) % m // the segment of the result which leaves the partner of node 0 of b
	out := Contour{Segs: make([]Segment, 0, len(segs))}
	if s == 0 {
		out.Start = first
	} else {
		out.Start = segs[s-1].P3
	}
	out.Segs = append(out.Segs, segs[s:]...)
	out.Segs = append(out.Segs, segs[:s]...)
	out.Segs[len(out.Segs)-1].P3 = out.Start
	if b.implicit {
		if l := out.Segs[len(out.Segs)-1]; l.IsCurve {
			return Contour{}, fmt.Errorf("a curve closes the contour where the model has a line")
		}
		out.Segs = out.Segs[:len(out.Segs)-1]
	}
	return out, nil
}

// mergeRun returns one segment from the point from for the segments of run,
// which follow each other from there.
func mergeRun(from Point, run []Segment) (Segment, error) {
	cur := run[0]
	for _, next := range run[1:] {
		mid := cur.P3
		switch {
		case mid == from:
			cur = next
			continue
		case next.P3 == mid:
			continue
		}
		if !cur.IsCurve && !next.IsCurve {
			u, v := sub(mid, from), sub(next.P3, mid)
			if math.Abs(u.X*v.Y-u.Y*v.X) > 1e-6*math.Hypot(u.X, u.Y)*math.Hypot(v.X, v.Y) || u.X*v.X+u.Y*v.Y < 0 {
				return Segment{}, fmt.Errorf("the corner at (%.1f, %.1f) is not a node of the model", mid.X, mid.Y)
			}
			cur = Segment{P3: next.P3}
			continue
		}
		din, _ := arriving(&cur, from)
		dout, _ := leaving(&next, mid)
		if math.Abs(turnAngle(din, dout)) >= SmoothAngle {
			return Segment{}, fmt.Errorf("the corner at (%.1f, %.1f) is not a node of the model", mid.X, mid.Y)
		}
		ca, cb := asCubic(cur, from), asCubic(next, mid)
		c, _, ok := mergeCubics(ca, cb)
		if !ok {
			c, ok = parentCubic(ca, cb)
		}
		if !ok {
			return Segment{}, fmt.Errorf("the node at (%.1f, %.1f) is not a node of the model and cannot be removed", mid.X, mid.Y)
		}
		cur = Segment{IsCurve: true, P1: c[1], P2: c[2], P3: c[3]}
	}
	return cur, nil
}

// parentCubic returns the cubic curve of which a and b, which follow each
// other, are the two parts, if there is one.
func parentCubic(a, b [4]Point) ([4]Point, bool) {
	la := math.Hypot(a[3].X-a[2].X, a[3].Y-a[2].Y)
	lb := math.Hypot(b[1].X-b[0].X, b[1].Y-b[0].Y)
	if la == 0 || lb == 0 {
		return [4]Point{}, false
	}
	t := la / (la + lb)
	c := [4]Point{
		a[0],
		{a[0].X + (a[1].X-a[0].X)/t, a[0].Y + (a[1].Y-a[0].Y)/t},
		{b[3].X + (b[2].X-b[3].X)/(1-t), b[3].Y + (b[2].Y-b[3].Y)/(1-t)},
		b[3],
	}
	l, r := splitCubic(c, t)
	for k := range l {
		if math.Hypot(l[k].X-a[k].X, l[k].Y-a[k].Y) > 1e-6 || math.Hypot(r[k].X-b[k].X, r[k].Y-b[k].Y) > 1e-6 {
			return [4]Point{}, false
		}
	}
	return c, true
}

// splitCubic returns the parts of the cubic c before and after t.
func splitCubic(c [4]Point, t float64) (l, r [4]Point) {
	mix := func(p, q Point) Point { return Point{p.X + t*(q.X-p.X), p.Y + t*(q.Y-p.Y)} }
	ab, bc, cd := mix(c[0], c[1]), mix(c[1], c[2]), mix(c[2], c[3])
	abc, bcd := mix(ab, bc), mix(bc, cd)
	mid := mix(abc, bcd)
	return [4]Point{c[0], ab, abc, mid}, [4]Point{mid, bcd, cd, c[3]}
}

// cutSegment returns the segment s, which starts at from, cut in as many
// places as there are shares.  A line is cut at the shares of its length,
// which increase.  A curve is cut near them: each cut lies between the cut
// before it and the share after it, where the curve runs most nearly in the
// direction given for the cut in dirs, if that is not the zero vector.
func cutSegment(from Point, s Segment, shares []float64, dirs []Point) []Segment {
	if !s.IsCurve {
		out := make([]Segment, 0, len(shares)+1)
		for _, f := range shares {
			out = append(out, Segment{P3: Point{from.X + f*(s.P3.X-from.X), from.Y + f*(s.P3.Y-from.Y)}})
		}
		return append(out, s)
	}
	c := asCubic(s, from)
	// The length along the curve at 64 steps of the parameter.
	const steps = 64
	var ls [steps + 1]float64
	prev := c[0]
	for i := 1; i <= steps; i++ {
		q := bez(c, float64(i)/steps)
		ls[i] = ls[i-1] + math.Hypot(q.X-prev.X, q.Y-prev.Y)
		prev = q
	}
	param := func(f float64) float64 {
		want := f * ls[steps]
		for i := 1; i <= steps; i++ {
			if ls[i] >= want {
				d := ls[i] - ls[i-1]
				if d == 0 {
					return float64(i) / steps
				}
				return (float64(i-1) + (want-ls[i-1])/d) / steps
			}
		}
		return 1
	}
	out := make([]Segment, 0, len(shares)+1)
	done := 0.0 // the parameter of c at which the rest starts
	rest := c
	for k, f := range shares {
		t := param(f)
		if dirs[k] != (Point{}) {
			lo, hi := done, 1.0
			if k+1 < len(shares) {
				hi = param(shares[k+1])
			}
			lo, hi = lo+cutMargin*(t-lo), hi-cutMargin*(hi-t)
			best := math.Abs(turnAngle(bezD(c, t), dirs[k]))
			const tries = 200
			for i := 0; i <= tries; i++ {
				u := lo + (hi-lo)*float64(i)/tries
				// A place further from the share must run closer to the
				// direction, by cutPull degrees for the whole range.
				if v := math.Abs(turnAngle(bezD(c, u), dirs[k])) + cutPull*math.Abs(u-param(f)); v < best {
					best, t = v, u
				}
			}
		}
		var l [4]Point
		if done < 1 {
			l, rest = splitCubic(rest, (t-done)/(1-done))
		} else {
			l = [4]Point{rest[0], rest[0], rest[0], rest[0]}
		}
		out = append(out, Segment{IsCurve: true, P1: l[1], P2: l[2], P3: l[3]})
		done = t
	}
	return append(out, Segment{IsCurve: true, P1: rest[1], P2: rest[2], P3: rest[3]})
}

// sameSegments reports whether the contours have the same number of segments
// of the same kinds in the same order and are closed in the same way.
func sameSegments(a, b Contour) bool {
	if len(a.Segs) != len(b.Segs) || a.isClosed() != b.isClosed() {
		return false
	}
	for i := range a.Segs {
		if a.Segs[i].IsCurve != b.Segs[i].IsCurve {
			return false
		}
	}
	return true
}
