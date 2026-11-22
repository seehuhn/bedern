// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package outline

import "math"

// SmoothAngle is the angle, in degrees, below which a node is smooth: the
// directions in which the outline arrives at the node and leaves it differ by
// less than this.
const SmoothAngle = 1.0

// MinPiece is the length, in font units, below which Round merges a segment
// with a neighbour at a smooth node.
const MinPiece = 20.0

// MergeTolerance is the largest distance, in font units, between a merged
// curve of Round and the segments it replaces.
const MergeTolerance = 0.5

// mergeOvershoot and mergeOvershootExact are the largest distances by which a
// merged curve may run beyond its end points in x or in y, with its points
// rounded to the nearest integers and as it is.
const (
	mergeOvershoot      = 0.3
	mergeOvershootExact = 0.45
)

// roundOvershoot is the largest distance by which a curve at a smooth node may
// run beyond its end points in x or in y after Round, unless it runs further
// with its points rounded to the nearest integers.
const roundOvershoot = 0.4

// extremeAngle is the angle, in degrees, within which Round treats the
// direction of a smooth node as horizontal or vertical.
const extremeAngle = 0.05

// turnTie is the largest angle, in degrees, by which the turn which Round
// leaves at a smooth node exceeds the least turn of the candidates there.
const turnTie = 1.5

// turnScale (degrees), curvatureScale and distWeight (per square unit) weigh
// the turn at a smooth node, the curvature error of its two curves and the
// distance by which the node and the handles move in Round.
const (
	turnScale      = 1.0
	curvatureScale = 0.3
	distWeight     = 0.5
)

// curvatureFloor is the curvature below which Round does not tell
// curvatures apart by their ratio.
const curvatureFloor = 0.002

// curvatureGood is the curvature error (curvatureError) which Round allows at
// a smooth node where plain rounding gives less.
const curvatureGood = 0.5

// Round returns a copy of the path with all coordinates rounded to integers.
//
// First the short segments are merged.  A segment shorter than MinPiece which
// meets a neighbour at a smooth node is replaced, together with the
// neighbour, by one cubic curve with the directions of the two at their outer
// ends, if such a curve lies within MergeTolerance of them, keeps its
// curvature within the range of theirs and at its ends close to theirs, and
// runs beyond its end points by less than half a unit (mergeFits).  The curve
// is the closer of the curve with the curvatures of the two at its ends and
// the curve closest to them in the least squares sense.  The shortest segments are
// merged first, each with the neighbour which gives the closer curve.  A line
// is merged only if it is the short segment, and a straight short segment only
// between two smooth nodes and not where a straight segment lies beyond it.
//
// Then the points are rounded to the nearest integers, halves away from zero.
// At a smooth node the node and its two handles are chosen together.  The
// candidates for the node are the integer points next to it where both of its
// segments are curves and the edge there is neither horizontal nor vertical
// (within 0.05 degrees), and the rounded node otherwise; those for a handle
// are the integer points within one unit of its exact position in both
// coordinates, on the side of the node towards the exact handle, and on the
// horizontal or vertical line through the node where the edge is horizontal
// or vertical.  A choice may not
// flip the sign of the curvature of a curve, change its curvature by more
// than curvatureGood or run beyond its end points by more than roundOvershoot,
// where plain rounding does not, and its turn at the node is at most turnTie
// above the least.  Of these the choice is taken which gives the least
// (turn/turnScale)^2 + (e/curvatureScale)^2 + distWeight d^2, where e is the
// largest curvature error of the two curves along them and the error of the
// change of curvature at the node, and d the distance by which the node and
// the handles move.  Each smooth node is placed twice, the second time with
// the handles which the first pass gave its neighbours.  All other points are
// rounded to the nearest integers.
func (p Path) Round() Path {
	out := make(Path, len(p))
	for i, c := range p {
		out[i] = c.merged().round()
	}
	return out
}

// Merged returns the path with the short segments merged as Round does, not
// rounded.
func (p Path) Merged() Path {
	out := make(Path, len(p))
	for i, c := range p {
		out[i] = c.merged()
	}
	return out
}

// roundPoint returns q rounded to the nearest integers, halves away from
// zero.  A coordinate within 0.001 of a half counts as the half.
func roundPoint(q Point) Point { return Point{roundUnit(q.X), roundUnit(q.Y)} }

// roundUnit returns x rounded to the nearest integer, halves away from zero,
// after x is rounded to a multiple of 0.001.
func roundUnit(x float64) float64 { return math.Round(math.Round(x*1e3) / 1e3) }

func sub(a, b Point) Point { return Point{a.X - b.X, a.Y - b.Y} }

func unitOf(v Point) Point {
	n := math.Hypot(v.X, v.Y)
	return Point{v.X / n, v.Y / n}
}

// turnAngle returns the angle from the direction u to the direction v, in
// degrees, between -180 and 180.
func turnAngle(u, v Point) float64 {
	return math.Atan2(u.X*v.Y-u.Y*v.X, u.X*v.X+u.Y*v.Y) * 180 / math.Pi
}

// isClosed reports whether the last segment of the contour ends at its start.
func (c Contour) isClosed() bool {
	n := len(c.Segs)
	if n == 0 {
		return false
	}
	e := c.Segs[n-1].P3
	return math.Abs(e.X-c.Start.X) < 1e-6 && math.Abs(e.Y-c.Start.Y) < 1e-6
}

// startOf returns the point at which segment j of the contour starts.
func (c Contour) startOf(j int) Point {
	if j == 0 {
		return c.Start
	}
	return c.Segs[j-1].P3
}

// arriving returns the direction in which segment s, which starts at p0,
// arrives at its end, and the handle which sets it (nil for a line or a
// handle on the node).
func arriving(s *Segment, p0 Point) (Point, *Point) {
	if s.IsCurve {
		if s.P2 != s.P3 {
			return sub(s.P3, s.P2), &s.P2
		}
		if s.P1 != s.P3 {
			return sub(s.P3, s.P1), nil
		}
	}
	return sub(s.P3, p0), nil
}

// leaving returns the direction in which segment s, which starts at p0,
// leaves it, and the handle which sets it (nil for a line or a handle on the
// node).
func leaving(s *Segment, p0 Point) (Point, *Point) {
	if s.IsCurve {
		if s.P1 != p0 {
			return sub(s.P1, p0), &s.P1
		}
		if s.P2 != p0 {
			return sub(s.P2, p0), nil
		}
	}
	return sub(s.P3, p0), nil
}

// NodeTurn returns the angle, in degrees, by which the outline turns at the
// end of segment j of the contour, between -180 and 180: the angle from the
// direction in which segment j arrives to that in which the next segment
// leaves.  For the last segment of a contour which does not end at its start,
// the next segment is the straight line which closes it.
func (c Contour) NodeTurn(j int) float64 {
	segs := c.Segs
	if !c.isClosed() {
		segs = append(segs[:len(segs):len(segs)], Segment{P3: c.Start})
	}
	k := (j + 1) % len(segs)
	din, _ := arriving(&segs[j], c.startOf(j))
	dout, _ := leaving(&segs[k], segs[j].P3)
	return turnAngle(din, dout)
}

// round rounds the contour as Round says.
func (c Contour) round() Contour {
	out := Contour{Start: roundPoint(c.Start), Segs: make([]Segment, len(c.Segs))}
	for j, s := range c.Segs {
		out.Segs[j] = Segment{IsCurve: s.IsCurve, P1: roundPoint(s.P1), P2: roundPoint(s.P2), P3: roundPoint(s.P3)}
	}
	n := len(c.Segs)
	closed := c.isClosed()
	// Each node is placed twice, the second time with the handles which the
	// first pass gave its neighbours.
	for pass := 0; pass < 2; pass++ {
		for j := 0; j < n; j++ {
			// The node at the end of segment j, and the segment which leaves it.
			k := j + 1
			if k == n {
				if !closed {
					continue
				}
				k = 0
			}
			c.smoothNode(&out, j, k)
		}
	}
	return out
}

// smoothNode places the handles of the rounded contour r at the node between
// segments j and k of c, and the node itself, if the node is smooth.
func (c Contour) smoothNode(r *Contour, j, k int) {
	sj, sk := &c.Segs[j], &c.Segs[k]
	node := sj.P3
	din, hin := arriving(sj, c.startOf(j))
	dout, hout := leaving(sk, node)
	if hin == nil && hout == nil {
		return
	}
	if math.Abs(turnAngle(din, dout)) >= SmoothAngle {
		return
	}
	rj, rk := &r.Segs[j], &r.Segs[k]
	plainNode := roundPoint(node)

	// A node at which the edge is horizontal or vertical keeps its handles on
	// that line.
	ui, uo := unitOf(din), unitOf(dout)
	a := math.Abs(math.Atan2(ui.Y+uo.Y, ui.X+uo.X)) * 180 / math.Pi
	horizontal := a < extremeAngle || a > 180-extremeAngle
	vertical := math.Abs(a-90) < extremeAngle

	// The candidates for the node: the integer points next to it in both
	// coordinates where both sides are curves and the edge is neither
	// horizontal nor vertical, and the rounded node otherwise.
	nodes := []Point{plainNode}
	if hin != nil && hout != nil && !horizontal && !vertical {
		nodes = nil
		for _, x := range nextIntegers(node.X) {
			for _, y := range nextIntegers(node.Y) {
				nodes = append(nodes, Point{x, y})
			}
		}
	}

	// The candidates for a handle: on the horizontal or vertical line through
	// the node, or anywhere, the integer points within one unit of the exact
	// handle in both coordinates.
	cands := func(h *Point, nd Point) []Point {
		if h == nil {
			return []Point{{}}
		}
		e := sub(*h, node)
		var qs []Point
		for x := math.Ceil(h.X - 1); x <= math.Floor(h.X+1); x++ {
			for y := math.Ceil(h.Y - 1); y <= math.Floor(h.Y+1); y++ {
				q := Point{x, y}
				d := sub(q, nd)
				switch {
				case horizontal && y != nd.Y, vertical && x != nd.X, q == nd, d.X*e.X+d.Y*e.Y <= 0:
					continue
				}
				qs = append(qs, q)
			}
		}
		return qs
	}

	// The directions of the sides without a handle to place are those of the
	// rounded contour.
	fixedIn, _ := arriving(rj, r.startOf(j))
	fixedOut, _ := leaving(rk, plainNode)
	type side struct {
		q, dir              Point
		dist, kerr, k, over float64
		flip                bool
	}
	ej, ek := [4]Point{c.startOf(j), sj.P1, sj.P2, node}, [4]Point{node, sk.P1, sk.P2, sk.P3}
	sideIn := func(q, nd Point) side {
		if hin == nil {
			return side{dir: neg(fixedIn)}
		}
		rc := [4]Point{r.startOf(j), rj.P1, q, nd}
		e, f := profileError(rc, ej)
		return side{q, sub(q, nd), math.Hypot(q.X-hin.X, q.Y-hin.Y), e, curvatureAt(rc, 1), overshoot(rc), f}
	}
	sideOut := func(q, nd Point) side {
		if hout == nil {
			return side{dir: fixedOut}
		}
		rc := [4]Point{nd, q, rk.P2, rk.P3}
		e, f := profileError(rc, ek)
		return side{q, sub(q, nd), math.Hypot(q.X-hout.X, q.Y-hout.Y), e, curvatureAt(rc, 0), overshoot(rc), f}
	}
	plain := func(h *Point) Point {
		if h == nil {
			return Point{}
		}
		q := roundPoint(*h)
		switch {
		case horizontal:
			q.Y = plainNode.Y
		case vertical:
			q.X = plainNode.X
		}
		return q
	}
	var keIn, keOut float64
	if hin != nil {
		keIn = curvatureAt(ej, 1)
	}
	if hout != nil {
		keOut = curvatureAt(ek, 0)
	}
	jumpError := func(si, so side) float64 {
		d := math.Abs(si.k-so.k) - math.Abs(keIn-keOut)
		return math.Max(0, d) / (math.Max(math.Abs(keIn), math.Abs(keOut)) + curvatureFloor)
	}
	pi, po := sideIn(plain(hin), plainNode), sideOut(plain(hout), plainNode)
	maxKerr := math.Max(math.Max(pi.kerr, po.kerr), curvatureGood)
	maxOver := math.Max(math.Max(pi.over, po.over), roundOvershoot)

	type choice struct {
		node       Point
		in, out    side
		turn, kerr float64
		dist       float64
	}
	var all []choice
	minTurn := math.Inf(1)
	for _, nd := range nodes {
		nodeDist := math.Hypot(nd.X-node.X, nd.Y-node.Y)
		for _, qi := range cands(hin, nd) {
			si := sideIn(qi, nd)
			if si.flip && !pi.flip || si.kerr > maxKerr || si.over > maxOver {
				continue
			}
			for _, qo := range cands(hout, nd) {
				so := sideOut(qo, nd)
				if so.flip && !po.flip || so.kerr > maxKerr || so.over > maxOver {
					continue
				}
				t := math.Abs(turnAngle(neg(si.dir), so.dir))
				minTurn = math.Min(minTurn, t)
				e := math.Max(math.Max(si.kerr, so.kerr), jumpError(si, so))
				all = append(all, choice{nd, si, so, t, e, nodeDist + si.dist + so.dist})
			}
		}
	}
	if len(all) == 0 {
		return
	}
	var best choice
	bestCost := math.Inf(1)
	for _, ch := range all {
		if ch.turn > minTurn+turnTie {
			continue
		}
		cost := ch.turn*ch.turn/(turnScale*turnScale) + ch.kerr*ch.kerr/(curvatureScale*curvatureScale) + distWeight*ch.dist*ch.dist
		if cost < bestCost-1e-9 || (cost < bestCost+1e-9 && ch.dist < best.dist) {
			best, bestCost = ch, cost
		}
	}
	rj.P3 = best.node
	if k == 0 {
		r.Start = best.node
	}
	if hin != nil {
		rj.P2 = best.in.q
	}
	if hout != nil {
		rk.P1 = best.out.q
	}
}

// nextIntegers returns the integers next to x: the two on either side, or
// x rounded if it lies within 0.001 of an integer.
func nextIntegers(x float64) []float64 {
	if r := roundUnit(x); math.Abs(x-r) < 1e-3 {
		return []float64{r}
	}
	return []float64{math.Floor(x), math.Ceil(x)}
}

func neg(v Point) Point { return Point{-v.X, -v.Y} }

// profileError returns the largest curvature error (curvatureError) of the
// rounded curve r against the exact curve e, at nine points evenly spaced in
// the parameter, and whether the sign of the curvature flips at one of them.
func profileError(r, e [4]Point) (float64, bool) {
	var worst float64
	var flip bool
	for i := 0; i <= 8; i++ {
		t := float64(i) / 8
		k, ke := curvatureAt(r, t), curvatureAt(e, t)
		worst = math.Max(worst, curvatureError(k, ke))
		flip = flip || flips(k, ke)
	}
	return worst, flip
}

// curvatureError returns the difference of the curvature k from the exact
// curvature e, relative to e and to curvatureFloor.
func curvatureError(k, e float64) float64 {
	return math.Abs(k-e) / (math.Abs(e) + curvatureFloor)
}

// flips reports whether the curvature k has the opposite sign of the exact
// curvature e, both beyond curvatureFloor.
func flips(k, e float64) bool {
	return math.Abs(e) > curvatureFloor && math.Abs(k) > curvatureFloor && (k > 0) != (e > 0)
}

// asCubic returns the segment, which starts at p0, as the control points of
// a cubic curve.
func asCubic(s Segment, p0 Point) [4]Point {
	if s.IsCurve {
		return [4]Point{p0, s.P1, s.P2, s.P3}
	}
	return [4]Point{p0, {p0.X + (s.P3.X-p0.X)/3, p0.Y + (s.P3.Y-p0.Y)/3}, {p0.X + 2*(s.P3.X-p0.X)/3, p0.Y + 2*(s.P3.Y-p0.Y)/3}, s.P3}
}

func bez(c [4]Point, t float64) Point {
	u := 1 - t
	a, b, cc, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	return Point{a*c[0].X + b*c[1].X + cc*c[2].X + d*c[3].X, a*c[0].Y + b*c[1].Y + cc*c[2].Y + d*c[3].Y}
}

func bezD(c [4]Point, t float64) Point {
	u := 1 - t
	a, b, cc := 3*u*u, 6*u*t, 3*t*t
	return Point{a*(c[1].X-c[0].X) + b*(c[2].X-c[1].X) + cc*(c[3].X-c[2].X), a*(c[1].Y-c[0].Y) + b*(c[2].Y-c[1].Y) + cc*(c[3].Y-c[2].Y)}
}

func bezDD(c [4]Point, t float64) Point {
	u := 1 - t
	return Point{6*u*(c[2].X-2*c[1].X+c[0].X) + 6*t*(c[3].X-2*c[2].X+c[1].X), 6*u*(c[2].Y-2*c[1].Y+c[0].Y) + 6*t*(c[3].Y-2*c[2].Y+c[1].Y)}
}

// arcLength returns the length of the cubic, from 32 straight pieces.
func arcLength(c [4]Point) float64 {
	var l float64
	prev := c[0]
	for i := 1; i <= 32; i++ {
		q := bez(c, float64(i)/32)
		l += math.Hypot(q.X-prev.X, q.Y-prev.Y)
		prev = q
	}
	return l
}

// merged returns the contour with its short segments merged, see Round.
func (c Contour) merged() Contour {
	segs := append([]Segment(nil), c.Segs...)
	start := c.Start
	closed := c.isClosed()
	failed := map[Point]bool{} // the ends of short segments which cannot be merged
	for {
		n := len(segs)
		if n < 4 {
			break
		}
		cur := Contour{Start: start, Segs: segs}
		// The shortest segment not yet tried.
		j, shortest := -1, MinPiece
		for i := range segs {
			if failed[segs[i].P3] {
				continue
			}
			if l := arcLength(asCubic(segs[i], cur.startOf(i))); l < shortest {
				j, shortest = i, l
			}
		}
		if j < 0 {
			break
		}
		// The two pairs (a, a+1) which contain j.
		bestDev, bestA := math.Inf(1), -1
		var bestC [4]Point
		for _, a := range []int{j - 1, j} {
			b := a + 1
			if closed {
				a, b = (a+n)%n, (b+n)%n
			} else if a < 0 || b >= n {
				continue
			}
			sa, sb := segs[a], segs[b]
			if !sa.IsCurve && !sb.IsCurve {
				continue
			}
			// A line is merged only if it is the short segment, and a straight
			// short segment only between two smooth nodes and not where a
			// straight segment lies beyond it.
			if o := a + b - j; !segs[o].IsCurve && o != j {
				continue
			}
			if isStraight(segs[j], cur.startOf(j)) {
				far, farNode := (j+1)%n, j // the segment beyond j and the node between them
				if j == a {
					far, farNode = (j-1+n)%n, (j-1+n)%n
				}
				if !closed && (j == a && j == 0 || j == b && j == n-1) {
					continue
				}
				if isStraight(segs[far], cur.startOf(far)) || math.Abs(cur.NodeTurn(farNode)) >= SmoothAngle {
					continue
				}
			}
			if math.Abs(cur.NodeTurn(a)) >= SmoothAngle {
				continue
			}
			ca, cb := asCubic(sa, cur.startOf(a)), asCubic(sb, sa.P3)
			m, dev, ok := mergeCubics(ca, cb)
			if ok && dev < bestDev {
				bestDev, bestA, bestC = dev, a, m
			}
		}
		if bestA < 0 {
			failed[segs[j].P3] = true
			continue
		}
		ms := Segment{IsCurve: true, P1: bestC[1], P2: bestC[2], P3: bestC[3]}
		b := (bestA + 1) % n
		if b == 0 {
			// The merged segment runs over the start of the contour.
			start = segs[n-2].P3
			segs = append([]Segment{ms}, segs[1:n-1]...)
			continue
		}
		segs = append(segs[:bestA:bestA], append([]Segment{ms}, segs[b+1:]...)...)
	}
	return Contour{Start: start, Segs: segs}
}

// isStraight reports whether the segment, which starts at p0, is a line or a
// curve whose control points lie on the line from p0 to its end.
func isStraight(s Segment, p0 Point) bool {
	if !s.IsCurve {
		return true
	}
	d := sub(s.P3, p0)
	l := math.Hypot(d.X, d.Y)
	if l == 0 {
		return false
	}
	for _, q := range []Point{s.P1, s.P2} {
		e := sub(q, p0)
		if math.Abs(d.X*e.Y-d.Y*e.X)/l > 1e-6 {
			return false
		}
	}
	return true
}
