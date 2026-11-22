// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of this file compare the rounded outline of every glyph with its
// unrounded outline after the merging of short segments (cachedMerged), which
// has the same segments.

// roundSmoothTurn is the largest turn, in degrees, of the rounded outline at a
// node which is smooth before rounding and whose curves have handles at least
// roundLongHandle units long; roundAnyTurn is the largest turn there whatever
// the handles.
const (
	roundSmoothTurn = 2.5
	roundLongHandle = 8.0
	roundAnyTurn    = 20.0
)

// handleLength returns the length of the handle of segment j of the contour
// at its end (end set) or at its start, and the length of the segment if it is
// a line.
func handleLength(ct outline.Contour, j int, end bool) float64 {
	c := segCubic(ct, j)
	switch {
	case ct.Segs[j%len(ct.Segs)].IsCurve && end:
		return math.Hypot(c.P3.X-c.P2.X, c.P3.Y-c.P2.Y)
	case ct.Segs[j%len(ct.Segs)].IsCurve:
		return math.Hypot(c.P1.X-c.P0.X, c.P1.Y-c.P0.Y)
	}
	return math.Hypot(c.P3.X-c.P0.X, c.P3.Y-c.P0.Y)
}

// smoothNodes calls f for every node of the path which is smooth
// (outline.SmoothAngle): the contour, the segment which ends at the node, and
// the segment which leaves it.
func smoothNodes(path outline.Path, f func(ci, j, k int)) {
	for ci, ct := range path {
		n := len(ct.Segs)
		for j := 0; j < n; j++ {
			if math.Abs(ct.NodeTurn(j)) < outline.SmoothAngle {
				f(ci, j, (j+1)%n)
			}
		}
	}
}

// At every node which is smooth before rounding the rounded outline turns by
// at most roundAnyTurn degrees, and by at most roundSmoothTurn degrees where
// the handles of both of its segments are at least roundLongHandle long, in
// every parameter set.
func TestRoundKeepsSmoothNodesSmooth(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, e := range table {
				exact, rounded := cachedMerged(np, e.name), cachedRounded(np, e.name).Path
				smoothNodes(exact, func(ci, j, k int) {
					ct := exact[ci]
					turn := math.Abs(rounded[ci].NodeTurn(j))
					long := math.Min(handleLength(ct, j, true), handleLength(ct, k, false)) >= roundLongHandle
					if turn > roundAnyTurn || long && turn > roundSmoothTurn {
						q := ct.Segs[j].P3
						t.Errorf("%s: the node at (%.0f, %.0f) turns by %.2f degrees after rounding", e.name, q.X, q.Y, turn)
					}
				})
			}
		})
	}
}

// At every node which is smooth before rounding and at which the outline runs
// horizontally or vertically, within 0.05 degrees, the handles of the rounded
// outline lie exactly on the horizontal or vertical line through the node, in
// every parameter set.
func TestRoundKeepsExtremesLevel(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, e := range table {
				exact, rounded := cachedMerged(np, e.name), cachedRounded(np, e.name).Path
				smoothNodes(exact, func(ci, j, k int) {
					a, b := segCubic(exact[ci], j).tangent(1), segCubic(exact[ci], k).tangent(0)
					deg := math.Abs(math.Atan2(a.Y+b.Y, a.X+b.X)) * 180 / math.Pi
					horizontal := deg < 0.05 || deg > 180-0.05
					vertical := math.Abs(deg-90) < 0.05
					if !horizontal && !vertical {
						return
					}
					rc := rounded[ci]
					in, out := segCubic(rc, j), segCubic(rc, k)
					var hs []outline.Point
					if rc.Segs[j].IsCurve {
						hs = append(hs, in.P2)
					}
					if rc.Segs[k].IsCurve {
						hs = append(hs, out.P1)
					}
					q := in.P3
					for _, h := range hs {
						if horizontal && h.Y != q.Y || vertical && h.X != q.X {
							t.Errorf("%s: the handle %v of the extreme at %v is not level with it", e.name, h, q)
						}
					}
				})
			}
		})
	}
}

// circleRun is a run of consecutive curves of one contour.
type circleRun struct {
	contour int
	segs    []int
	centre  outline.Point
	r       float64
}

// circleRuns returns the runs of consecutive curves of the path whose
// curvature at five points of each curve lies within circleShare of the
// curvature at the middle of the first curve, and which turn by more than
// circleTurn degrees in all: the balls of the terminals and the dots.  The
// centre and the radius are those of the circle fitted to nine points of each
// curve (fitCircle).
func circleRuns(path outline.Path) []circleRun {
	var out []circleRun
	for ci, ct := range path {
		n := len(ct.Segs)
		k := make([]float64, n)
		arc := make([]bool, n)
		for j := 0; j < n; j++ {
			if !ct.Segs[j].IsCurve {
				continue
			}
			c := segCubic(ct, j)
			k[j], arc[j] = c.curvature(0.5), true
			for _, u := range []float64{0, 0.25, 0.75, 1} {
				if math.Abs(c.curvature(u)/k[j]-1) > circleShare {
					arc[j] = false
				}
			}
		}
		same := func(a, b int) bool { return arc[a] && arc[b] && math.Abs(k[a]/k[b]-1) < circleShare }
		for j := 0; j < n; j++ {
			if !arc[j] || same((j+n-1)%n, j) {
				continue
			}
			run := []int{j}
			for l := (j + 1) % n; l != j && same(j, l); l = (l + 1) % n {
				run = append(run, l)
			}
			var turned float64
			for _, l := range run {
				c := segCubic(ct, l)
				a, b := c.tangent(0), c.tangent(1)
				turned += math.Abs(math.Atan2(a.X*b.Y-a.Y*b.X, a.X*b.X+a.Y*b.Y))
			}
			if turned*180/math.Pi <= circleTurn {
				continue
			}
			var pts []outline.Point
			for _, l := range run {
				c := segCubic(ct, l)
				for i := 0; i <= 8; i++ {
					pts = append(pts, c.at(float64(i)/8))
				}
			}
			centre, r := fitCircle(pts)
			out = append(out, circleRun{ci, run, centre, r})
		}
	}
	return out
}

// circleShare, circleTurn: see circleRuns.  circleTrue is the distance within
// which the unrounded run lies on its circle, circleRounded that within which
// the rounded run lies on it, and circleBend the share within which the
// curvature of each rounded curve at least circleLong units long stays of that
// of the circle.
const (
	circleShare   = 0.03
	circleTurn    = 150.0
	circleTrue    = 0.1
	circleRounded = 1.2
	circleBend    = 0.25
	circleLong    = 30.0
)

// The glyphs with a ball terminal.
var ballGlyphs = []string{"a", "c", "f", "g", "j", "r", "y", "J", "two", "three", "five", "six", "nine", "question", "sterling", "section"}

// The visible arc of every ball and every dot is a circle: a run of curves
// with the curvature of the circle within circleShare which turns by more than
// circleTurn degrees and lies within circleTrue of the circle, and rounded
// within circleRounded of it, with the curvature of its curves at least
// circleLong long within circleBend of the circle's.  Every glyph with a ball
// terminal has at least one such run, in every parameter set.
func TestBallsAreCircles(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			found := map[string]bool{}
			for _, e := range table {
				exact, rounded := cachedMerged(np, e.name), cachedRounded(np, e.name).Path
				for _, run := range circleRuns(exact) {
					found[e.name] = true
					for _, l := range run.segs {
						for _, c := range []struct {
							what string
							c    cubic
							tol  float64
						}{{"unrounded", segCubic(exact[run.contour], l), circleTrue}, {"rounded", segCubic(rounded[run.contour], l), circleRounded}} {
							long := math.Hypot(c.c.P3.X-c.c.P0.X, c.c.P3.Y-c.c.P0.Y) >= circleLong
							for i := 0; i <= 8; i++ {
								u := float64(i) / 8
								q := c.c.at(u)
								if d := math.Hypot(q.X-run.centre.X, q.Y-run.centre.Y) - run.r; math.Abs(d) > c.tol {
									t.Errorf("%s, %s: the circle of the radius %.1f at %v is %.2f away at %v", e.name, c.what, run.r, run.centre, d, q)
								}
								if k := math.Abs(c.c.curvature(u)) * run.r; c.what == "rounded" && long && math.Abs(k-1) > circleBend {
									t.Errorf("%s, rounded: the curvature of the circle of the radius %.1f at %v is %.2f times its own at %v", e.name, run.r, run.centre, k, q)
								}
							}
						}
					}
				}
			}
			for _, name := range ballGlyphs {
				if !found[name] {
					t.Errorf("%s: no ball found", name)
				}
			}
		})
	}
}

// shortSeg is the length below which a segment between two smooth nodes is
// short.
const shortSeg = 15.0

// shortSegments lists, for the regular parameters, the glyphs which keep
// segments shorter than shortSeg between two smooth nodes, and how many.
var shortSegments = map[string]int{
	"five": 1,
}

// In the regular parameters no segment of the rounded outline between two
// nodes which are smooth before rounding is shorter than shortSeg, except in
// the glyphs and as many times as shortSegments lists.
func TestNoShortSegmentsBetweenSmoothNodes(t *testing.T) {
	t.Parallel()
	np := namedParams{name: "regular", params: design.Regular()}
	got := map[string]int{}
	for _, e := range table {
		exact, rounded := cachedMerged(np, e.name), cachedRounded(np, e.name).Path
		for ci, ct := range exact {
			n := len(ct.Segs)
			for j := 0; j < n; j++ {
				if math.Abs(ct.NodeTurn((j+n-1)%n)) >= outline.SmoothAngle || math.Abs(ct.NodeTurn(j)) >= outline.SmoothAngle {
					continue
				}
				c := segCubic(rounded[ci], j)
				if math.Hypot(c.P3.X-c.P0.X, c.P3.Y-c.P0.Y) < shortSeg {
					got[e.name]++
				}
			}
		}
	}
	var diff []string
	for name := range mergeKeys(got, shortSegments) {
		if got[name] != shortSegments[name] {
			diff = append(diff, fmt.Sprintf("%s: %d, want %d", name, got[name], shortSegments[name]))
		}
	}
	sort.Strings(diff)
	if len(diff) > 0 {
		t.Errorf("short segments between smooth nodes:\n%s", strings.Join(diff, "\n"))
	}
}

// mergeKeys returns the keys of both maps.
func mergeKeys(a, b map[string]int) map[string]bool {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	return keys
}
