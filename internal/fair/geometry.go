// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// seg is a line (Curve false, P1 and P2 unused) or a cubic from P0 to P3.
type seg struct {
	Curve          bool
	P0, P1, P2, P3 Point
}

func sub(a, b Point) Point             { return Point{X: a.X - b.X, Y: a.Y - b.Y} }
func add(a, b Point) Point             { return Point{X: a.X + b.X, Y: a.Y + b.Y} }
func mul(a Point, s float64) Point     { return Point{X: a.X * s, Y: a.Y * s} }
func dot(a, b Point) float64           { return a.X*b.X + a.Y*b.Y }
func cross(a, b Point) float64         { return a.X*b.Y - a.Y*b.X }
func norm(a Point) float64             { return math.Hypot(a.X, a.Y) }
func dist(a, b Point) float64          { return norm(sub(a, b)) }
func lerp(a, b Point, t float64) Point { return add(a, mul(sub(b, a), t)) }
func unit(a Point) Point {
	n := norm(a)
	if n == 0 {
		return Point{}
	}
	return mul(a, 1/n)
}
func deg(a Point) float64 { return math.Atan2(a.Y, a.X) * 180 / math.Pi }

// at returns the point, the first and the second derivative at t.
func (s seg) at(t float64) (p, d1, d2 Point) {
	if !s.Curve {
		return lerp(s.P0, s.P3, t), sub(s.P3, s.P0), Point{}
	}
	u := 1 - t
	p = add(add(mul(s.P0, u*u*u), mul(s.P1, 3*u*u*t)), add(mul(s.P2, 3*u*t*t), mul(s.P3, t*t*t)))
	d1 = add(add(mul(sub(s.P1, s.P0), 3*u*u), mul(sub(s.P2, s.P1), 6*u*t)), mul(sub(s.P3, s.P2), 3*t*t))
	d2 = add(mul(add(sub(s.P2, mul(s.P1, 2)), s.P0), 6*u), mul(add(sub(s.P3, mul(s.P2, 2)), s.P1), 6*t))
	return
}

// curvature returns the signed curvature and the unit tangent at t.  Where
// the derivative vanishes (a handle of length zero) they are taken a little
// way inside the segment.
func (s seg) curvature(t float64) (k float64, tan Point) {
	_, d1, d2 := s.at(t)
	if norm(d1) < 1e-7*(1+s.chord()) {
		tt := t + 1e-4
		if t > 0.5 {
			tt = t - 1e-4
		}
		_, d1, d2 = s.at(tt)
	}
	n := norm(d1)
	if n == 0 {
		return 0, Point{}
	}
	return cross(d1, d2) / (n * n * n), mul(d1, 1/n)
}

// chord returns the length of the control polygon.
func (s seg) chord() float64 {
	if !s.Curve {
		return dist(s.P0, s.P3)
	}
	return dist(s.P0, s.P1) + dist(s.P1, s.P2) + dist(s.P2, s.P3)
}

// length returns the arc length of the segment.
func (s seg) length() float64 {
	if !s.Curve {
		return dist(s.P0, s.P3)
	}
	const n = 32
	var l float64
	prev := s.P0
	for i := 1; i <= n; i++ {
		p, _, _ := s.at(float64(i) / n)
		l += dist(prev, p)
		prev = p
	}
	return l
}

// reverse returns the segment run backwards.
func (s seg) reverse() seg {
	return seg{Curve: s.Curve, P0: s.P3, P1: s.P2, P2: s.P1, P3: s.P0}
}

// segsOf returns the segments of a contour, including the closing line,
// without segments of zero length.
func segsOf(c outline.Contour) []seg {
	var out []seg
	cur := c.Start
	for _, s := range c.Segs {
		g := seg{Curve: s.IsCurve, P0: cur, P1: s.P1, P2: s.P2, P3: s.P3}
		cur = s.P3
		if g.chord() < 1e-9 {
			continue
		}
		if g.Curve && cross(sub(g.P1, g.P0), sub(g.P3, g.P0)) == 0 && cross(sub(g.P2, g.P0), sub(g.P3, g.P0)) == 0 &&
			dot(sub(g.P1, g.P0), sub(g.P3, g.P0)) >= 0 && dot(sub(g.P2, g.P3), sub(g.P0, g.P3)) >= 0 {
			// A cubic whose control points lie on the chord, between its
			// ends, is a straight segment.
			g = seg{P0: g.P0, P3: g.P3}
		}
		out = append(out, g)
	}
	if cur != c.Start && dist(cur, c.Start) > 1e-9 {
		out = append(out, seg{P0: cur, P3: c.Start})
	}
	return out
}

// toContour converts segments back to an outline contour.
func toContour(ss []seg) outline.Contour {
	var c outline.Contour
	if len(ss) == 0 {
		return c
	}
	c.Start = ss[0].P0
	for _, s := range ss {
		c.Segs = append(c.Segs, outline.Segment{IsCurve: s.Curve, P1: s.P1, P2: s.P2, P3: s.P3})
	}
	return c
}

// flatten returns the polygon of segments, n pieces for each curve.
func flatten(ss []seg, n int) []Point {
	var pts []Point
	for _, s := range ss {
		pts = append(pts, s.P0)
		if s.Curve {
			for i := 1; i < n; i++ {
				p, _, _ := s.at(float64(i) / float64(n))
				pts = append(pts, p)
			}
		}
	}
	return pts
}

// winding returns the winding number of the polygons around q.
func winding(polys [][]Point, q Point) int {
	w := 0
	for _, pts := range polys {
		for i, a := range pts {
			b := pts[(i+1)%len(pts)]
			if a.Y <= q.Y {
				if b.Y > q.Y && cross(sub(b, a), sub(q, a)) > 0 {
					w++
				}
			} else if b.Y <= q.Y && cross(sub(b, a), sub(q, a)) < 0 {
				w--
			}
		}
	}
	return w
}

// orient returns the contours of p as segments, each turned so that the ink
// lies on its left.  The ink side is found from the winding number of the
// whole outline at points a little to either side of the contour; where the
// test is not decided (overlapping contours), the orientation of the
// largest contour decides, as in the font formats.
func orient(p outline.Path) [][]seg {
	var cs [][]seg
	var polys [][]Point
	for _, c := range p {
		ss := segsOf(c)
		if len(ss) == 0 {
			continue
		}
		cs = append(cs, ss)
		polys = append(polys, flatten(ss, 16))
	}
	// The sign of the area of the largest contour.
	var big, sign float64
	for _, pts := range polys {
		var a float64
		for i, q := range pts {
			r := pts[(i+1)%len(pts)]
			a += q.X*r.Y - r.X*q.Y
		}
		if math.Abs(a) > big {
			big, sign = math.Abs(a), a
		}
	}
	for i, ss := range cs {
		vote := 0
		for _, s := range ss {
			p, d, _ := s.at(0.5)
			t := unit(d)
			if t == (Point{}) {
				continue
			}
			left := add(p, mul(Point{X: -t.Y, Y: t.X}, 0.3))
			right := add(p, mul(Point{X: t.Y, Y: -t.X}, 0.3))
			wl, wr := winding(polys, left) != 0, winding(polys, right) != 0
			switch {
			case wl && !wr:
				vote++
			case wr && !wl:
				vote--
			}
		}
		var a float64
		pts := polys[i]
		for j, q := range pts {
			r := pts[(j+1)%len(pts)]
			a += q.X*r.Y - r.X*q.Y
		}
		flip := vote < 0
		if vote == 0 {
			// Undecided: an outer contour (of the sign of the largest) is
			// made counter-clockwise, which puts the ink on its left, and a
			// counter clockwise.
			outer := a*sign > 0
			flip = outer != (a > 0)
		}
		if flip {
			r := make([]seg, len(ss))
			for j, s := range ss {
				r[len(ss)-1-j] = s.reverse()
			}
			cs[i] = r
		}
	}
	return cs
}

// sampleStep is the largest arc length between two samples of a curve.
const sampleStep = 1.5

// sampleContour samples every segment from its start to its end, so that the
// end of a segment and the start of the next are both present, at the same
// arc length.
func sampleContour(ss []seg) ([]Sample, float64) {
	var out []Sample
	var s0 float64
	for i, g := range ss {
		l := g.length()
		n := 2
		if g.Curve {
			n = max(12, min(400, int(math.Ceil(l/sampleStep))))
		} else {
			n = max(2, int(math.Ceil(l/3)))
		}
		var prev Point
		var acc float64
		for j := 0; j <= n; j++ {
			t := float64(j) / float64(n)
			p, _, _ := g.at(t)
			k, tan := g.curvature(t)
			if !g.Curve {
				k = 0
			}
			if j > 0 {
				acc += dist(prev, p)
			}
			prev = p
			out = append(out, Sample{P: p, S: s0 + acc, K: k, T: tan, Seg: i})
		}
		s0 += acc
		if g.Curve {
			smooth(out[len(out)-n-1:])
		}
	}
	return out, s0
}

// smoothLen is half the length over which the curvature of a sample is
// taken.
const smoothLen = 2.0

// smooth replaces the curvature of the samples of one curve by the turning
// of the tangent over smoothLen on either side, within the curve, divided by
// the arc length.  A handle which is nearly of length zero makes the
// curvature at the end of the curve very large on a stretch too short to
// see; the mean over a few units is what the eye sees.  At the ends of the
// curve the window is one-sided, so that a jump at a node is kept.
func smooth(ss []Sample) {
	theta := make([]float64, len(ss))
	for i, s := range ss {
		theta[i] = math.Atan2(s.T.Y, s.T.X)
		if i > 0 {
			for theta[i]-theta[i-1] > math.Pi {
				theta[i] -= 2 * math.Pi
			}
			for theta[i]-theta[i-1] < -math.Pi {
				theta[i] += 2 * math.Pi
			}
		}
	}
	ks := make([]float64, len(ss))
	lo, hi := 0, 0
	for i, s := range ss {
		for ss[lo].S < s.S-smoothLen && lo < i {
			lo++
		}
		for hi+1 < len(ss) && ss[hi+1].S <= s.S+smoothLen {
			hi++
		}
		a, b := lo, hi
		if a == b {
			ks[i] = s.K
			continue
		}
		ks[i] = (theta[b] - theta[a]) / (ss[b].S - ss[a].S)
	}
	for i := range ss {
		ss[i].K = ks[i]
	}
}

// nodesOf returns the node at the start of each segment.
func nodesOf(ss []seg, samples []Sample, th Thresholds) []Node {
	first := make([]int, len(ss)) // index of the first sample of each segment
	last := make([]int, len(ss))
	for i := len(samples) - 1; i >= 0; i-- {
		first[samples[i].Seg] = i
	}
	for i, s := range samples {
		last[s.Seg] = i
	}
	nodes := make([]Node, len(ss))
	for i := range ss {
		prev := (i + len(ss) - 1) % len(ss)
		in, out := samples[last[prev]], samples[first[i]]
		turn := 0.0
		if in.T != (Point{}) && out.T != (Point{}) {
			turn = math.Atan2(cross(in.T, out.T), dot(in.T, out.T)) * 180 / math.Pi
		}
		nodes[i] = Node{
			P: ss[i].P0, S: out.S, Turn: turn, KIn: in.K, KOut: out.K,
			Corner:  math.Abs(turn) > th.CornerDeg,
			InCurve: ss[prev].Curve, OutCurve: ss[i].Curve,
		}
	}
	return nodes
}

// runsOf cuts the contour at its corners into runs.
func runsOf(ci int, c *Contour) []Run {
	n := len(c.Segs)
	var corners []int
	for i, nd := range c.Nodes {
		if nd.Corner {
			corners = append(corners, i)
		}
	}
	segSamples := make([][]Sample, n)
	for _, s := range c.Samples {
		segSamples[s.Seg] = append(segSamples[s.Seg], s)
	}
	build := func(from, count int) Run {
		r := Run{Contour: ci}
		var offset float64
		for k := 0; k < count; k++ {
			i := (from + k) % n
			if k > 0 && i == 0 {
				offset = c.Length
			}
			if k > 0 {
				r.Nodes = append(r.Nodes, i)
			}
			for _, s := range segSamples[i] {
				s.S += offset
				r.Samples = append(r.Samples, s)
			}
		}
		r.Straight = count == 1 && !c.Segs[from].Curve
		return r
	}
	if len(corners) == 0 {
		r := build(0, n)
		r.Cyclic = true
		r.Nodes = append(r.Nodes, 0)
		return []Run{r}
	}
	var runs []Run
	for k, from := range corners {
		to := corners[(k+1)%len(corners)]
		count := (to - from + n) % n
		if count == 0 {
			count = n
		}
		runs = append(runs, build(from, count))
	}
	return runs
}

// fillets marks the curves which belong to a fillet: a group of curves
// joined at smooth nodes, bounded at both ends by a straight segment or a
// corner, and shorter than FilletLen in all.  The bracket of a serif and a
// rounded corner are fillets; the eye judges them as a joint.
func fillets(ss []seg, nodes []Node, th Thresholds) []bool {
	n := len(ss)
	out := make([]bool, n)
	// A boundary is a node at a corner or next to a straight segment.
	start := -1
	for i := range n {
		if !ss[i].Curve || nodes[i].Corner {
			start = i
			break
		}
	}
	if start < 0 {
		return out // all curves, no corner: a closed curve
	}
	for k := 0; k < n; {
		i := (start + k) % n
		if !ss[i].Curve {
			k++
			continue
		}
		// A group of curves from i on.
		var group []int
		var l float64
		for k < n {
			j := (start + k) % n
			if !ss[j].Curve || (len(group) > 0 && nodes[j].Corner) {
				break
			}
			group = append(group, j)
			l += ss[j].length()
			k++
		}
		if l < th.FilletLen {
			for _, j := range group {
				out[j] = true
			}
		}
	}
	return out
}
