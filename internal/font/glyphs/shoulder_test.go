// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// A shoulderEdge is an edge of a shoulder: it runs from the join, where it
// meets the edge of a stem, to the top, where it is horizontal.  Forward says
// whether the contour runs from the join to the top.
type shoulderEdge struct {
	what      string
	join, top outline.Point
	forward   bool
}

// shoulderEdges returns the edges of the shoulders of the letter: of n, h and
// the second arch of m the outer edge and the underside where they run into
// the right stem, of the first arch of m the underside (its outer edge ends in
// a notch), of u the same as of n, turned, and of a the outer edge and the
// underside of the top, which run into the stem from the crest.
func shoulderEdges(name string, p design.Params) []shoulderEdge {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	var res []shoulderEdge
	add := func(a archSpec, m func(outline.Point) outline.Point) {
		if !a.AtNotch {
			res = append(res, shoulderEdge{"outer edge", m(pt(a.X3, a.Shoulder)), m(pt(a.XPeak, a.Peak)), true})
		}
		res = append(res, shoulderEdge{"underside", m(pt(a.X2, a.ShoulderInner)), m(pt(a.XPeak-nArchShape.UnderPeakShift, a.Peak-a.Thick)), false})
	}
	id := func(q outline.Point) outline.Point { return q }
	switch name {
	case "n":
		left := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		right := stemSpec{Left: left.right() + p.NCounter, Width: p.Stem}
		add(archBetween(p, left, right), id)
	case "h":
		left := stemSpec{Left: p.SideStemLeft + design.AscenderInset(p), Width: p.AscenderStem}
		inset := (p.Stem - left.Width) / 2
		right := stemSpec{Left: left.Left - inset + p.Stem + p.NCounter, Width: p.Stem}
		add(archBetween(p, left, right), id)
	case "m":
		first := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		middle := stemSpec{Left: first.right() + p.MCounter, Width: p.Stem}
		last := stemSpec{Left: middle.right() + p.MCounter, Width: p.Stem}
		archL := archBetween(p, first, middle)
		archL.Shoulder, archL.AtNotch = archL.Notch, true
		add(archL, id)
		add(archBetween(p, middle, last), id)
	case "u":
		left := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		right := stemSpec{Left: left.right() + p.UCounter, Width: p.Stem}
		a, cx := uArch(p, left, right)
		add(a, func(q outline.Point) outline.Point { return pt(2*cx-q.X, p.XHeight-q.Y) })
	case "a":
		left := design.AStemLeft(p)
		x, _ := aCrestAt(p, aTerminal(p))
		crest := pt(x, p.XHeight+p.OvershootLower)
		res = append(res,
			shoulderEdge{"outer edge", pt(left+p.Stem, p.ShoulderStart-aOuterDrop), crest, true},
			shoulderEdge{"underside", pt(left, p.ShoulderStartInner+aInnerRaise), pt(crest.X, crest.Y-p.BallStalk), false})
	default:
		panic("shoulderEdges: " + name)
	}
	return res
}

// A shoulderSample is a point of a shoulder edge, with the arc length from
// the join, the direction of the contour and the magnitude of its curvature.
type shoulderSample struct {
	at    outline.Point
	s     float64
	angle float64
	k     float64
}

// sampleShoulder returns the edge sampled from the join to the top, and the
// jumps of the magnitude of the curvature at the nodes between them and at the
// top, where the edge continues into the next curve.
func sampleShoulder(t *testing.T, path outline.Path, e shoulderEdge) (samples []shoulderSample, jumps []float64) {
	t.Helper()
	ct := path[0]
	n := len(ct.Segs)
	first, last := e.join, e.top
	if !e.forward {
		first, last = e.top, e.join
	}
	ia, ib := nodeNear(t, ct, first), nodeNear(t, ct, last)
	var segs []cubic
	for i := ia; i%n != ib; i++ {
		segs = append(segs, segCubic(ct, i))
	}
	// The curve on the far side of the top.
	var beyond cubic
	if e.forward {
		beyond = segCubic(ct, ib)
	} else {
		beyond = segCubic(ct, ia+n-1)
	}
	if !e.forward {
		for i, j := 0, len(segs)-1; i < j; i, j = i+1, j-1 {
			segs[i], segs[j] = segs[j], segs[i]
		}
		for i := range segs {
			c := segs[i]
			segs[i] = cubic{c.P3, c.P2, c.P1, c.P0}
		}
		c := beyond
		beyond = cubic{c.P3, c.P2, c.P1, c.P0}
	}
	length := 0.0
	prev := segs[0].P0
	for i, c := range segs {
		for j := 0; j <= 100; j++ {
			tt := float64(j) / 100
			q := c.at(tt)
			length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
			prev = q
			d := c.tangent(tt)
			samples = append(samples, shoulderSample{q, length, math.Atan2(d.Y, d.X), math.Abs(c.curvature(tt))})
		}
		next := beyond
		if i+1 < len(segs) {
			next = segs[i+1]
		}
		jumps = append(jumps, math.Abs(math.Abs(c.curvature(1))-math.Abs(next.curvature(0))))
	}
	return samples, jumps
}

// The shoulder runs into the stem without a corner.  At the join with the edge
// of the stem the edge is vertical within shoulderTangent; within shoulderNear
// units of the join its curvature is below shoulderJoinK (a radius of a
// thousand units); over the last shoulderStretch units before the join it falls
// monotonically; and from the join to beyond the top its curvature jumps
// nowhere by more than shoulderJump.
const (
	shoulderTangent = 0.5 * math.Pi / 180
	shoulderNear    = 5.0
	shoulderJoinK   = 0.001
	shoulderStretch = 40.0
	shoulderJump    = 0.0005
)

func TestShoulderRunsIntoTheStem(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range []string{"n", "h", "m", "u", "a"} {
				path, _ := cachedOutline(np, name)
				for _, e := range shoulderEdges(name, np.params) {
					smp, jumps := sampleShoulder(t, path, e)
					where := np.name + ", " + name + " " + e.what
					if d := math.Abs(math.Abs(math.Remainder(smp[0].angle, math.Pi)) - math.Pi/2); d > shoulderTangent {
						t.Errorf("%s: at the join the edge is %.2f degrees off the vertical", where, d*180/math.Pi)
					}
					for i := 1; i < len(smp) && smp[i].s < shoulderStretch; i++ {
						if smp[i].s < shoulderNear && smp[i].k > shoulderJoinK {
							t.Errorf("%s: %.1f units from the join the curvature is %.5f, want at most %v", where, smp[i].s, smp[i].k, shoulderJoinK)
							break
						}
						if smp[i].k < smp[i-1].k-1e-6 {
							t.Errorf("%s: the curvature falls from %.5f to %.5f %.1f units from the join, want it to grow monotonically from the join", where, smp[i-1].k, smp[i].k, smp[i].s)
							break
						}
					}
					for i, j := range jumps {
						if j > shoulderJump {
							t.Errorf("%s: the curvature jumps by %.5f at node %d from the join", where, j, i+1)
						}
					}
				}
			}
		})
	}
}

// In the font the curve at each join has both control points on the vertical
// through the join.
func TestShoulderJoinsAreVerticalWhenRounded(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range []string{"n", "h", "m", "u", "a"} {
				g := cachedRounded(np, name)
				if len(g.Path) == 0 {
					continue
				}
				ct := g.Path[0]
				n := len(ct.Segs)
				for _, e := range shoulderEdges(name, np.params) {
					i := nodeNear(t, ct, e.join)
					c := segCubic(ct, i)
					if !e.forward {
						d := segCubic(ct, i+n-1)
						c = cubic{d.P3, d.P2, d.P1, d.P0}
					}
					if c.P1.X != c.P0.X || c.P2.X != c.P0.X {
						t.Errorf("%s, %s %s: the curve at the join %v has the control points %v and %v off its vertical", np.name, name, e.what, c.P0, c.P1, c.P2)
					}
				}
			}
		})
	}
}

// normalisedShoulder returns the edge from the join to the top in the frame
// in which the top is at (0, 1) and the join at (1, 0).
func normalisedShoulder(smp []shoulderSample, e shoulderEdge) []outline.Point {
	var res []outline.Point
	for _, q := range smp {
		res = append(res, outline.Point{
			X: (q.at.X - e.top.X) / (e.join.X - e.top.X),
			Y: (q.at.Y - e.join.Y) / (e.top.Y - e.join.Y),
		})
	}
	return res
}

// The shoulder of a is the shoulder of n in kind: scaled to the same height
// and width, each edge of a lies within maxShoulderApart (as a fraction of the
// height and the width) of the same edge of n.
const maxShoulderApart = 0.01

func TestAShoulderIsTheShoulderOfN(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			pathN, _ := cachedOutline(np, "n")
			pathA, _ := cachedOutline(np, "a")
			en, ea := shoulderEdges("n", np.params), shoulderEdges("a", np.params)
			for k := range en {
				sn, _ := sampleShoulder(t, pathN, en[k])
				sa, _ := sampleShoulder(t, pathA, ea[k])
				qn, qa := normalisedShoulder(sn, en[k]), normalisedShoulder(sa, ea[k])
				worst := 0.0
				for _, q := range qa {
					best := math.Inf(1)
					for i := 1; i < len(qn); i++ {
						best = math.Min(best, math.Hypot(q.X-nearestOnSeg(q, qn[i-1], qn[i]).X, q.Y-nearestOnSeg(q, qn[i-1], qn[i]).Y))
					}
					worst = math.Max(worst, best)
				}
				if np.name == "regular" {
					t.Logf("%s: the shoulders of a and n are at most %.3f apart", en[k].what, worst)
				}
				if worst > maxShoulderApart {
					t.Errorf("parameter set %s, %s: scaled to the same size the shoulder of a lies %.3f from that of n, want at most %v", np.name, en[k].what, worst, maxShoulderApart)
				}
			}
		})
	}
}
