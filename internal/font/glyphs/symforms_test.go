// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// Tests of the section sign, the ampersand, the terminal of 6 and 9 and the
// question mark.

// findBalls returns the centres of the circles of the radius r on the outline:
// runs of the contours, as long as ballArc of the circumference, which lie within
// half a unit of a circle whose radius is within a unit of r.
func findBalls(path outline.Path, r float64) []outline.Point {
	var out []outline.Point
	for _, ct := range path {
		pts := sampleContour(ct, 0.5)
		n := len(pts)
		m := int(ballArc * 2 * math.Pi * r / 0.5)
		if n < m {
			continue
		}
		for i := 0; i < n; i += 4 {
			win := make([]outline.Point, m)
			for k := range win {
				win[k] = pts[(i+k)%n]
			}
			c, rr := fitCircle(win)
			if math.Abs(rr-r) > 1 {
				continue
			}
			dev := 0.0
			for _, q := range win {
				dev = math.Max(dev, math.Abs(math.Hypot(q.X-c.X, q.Y-c.Y)-rr))
			}
			if dev > 0.5 {
				continue
			}
			seen := false
			for _, o := range out {
				if math.Hypot(o.X-c.X, o.Y-c.Y) < 5 {
					seen = true
				}
			}
			if !seen {
				out = append(out, c)
			}
		}
	}
	return out
}

// ballArc is the fraction of the circumference over which the free part of a
// ball terminal follows its circle: the shared part is attached over 99 degrees,
// so that 261 degrees are free.
const ballArc = 0.65

// The balls of the section sign, the question mark, 6 and 9 are the ball
// terminal of the lowercase (ball.go), at the size of the lowercase on the
// question mark and at the size of the figures on 6, 9 and the section sign:
// each is a circle of that diameter, within half a unit, over ballArc of its
// circumference, and each glyph has as many as it has terminals.
func TestSymbolBalls(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		p := np.params
		for _, c := range []struct {
			name string
			d    float64
			n    int
		}{
			{"section", p.Figures.Ball, 2},
			{"question", p.BallDiameter, 1},
			{"six", p.Figures.Ball, 1},
			{"nine", p.Figures.Ball, 1},
		} {
			path, _, _ := outlineOf(c.name, p)
			if got := len(findBalls(path, c.d/2)); got != c.n {
				t.Errorf("%s, %s: %d circles of the diameter %v on the outline, want %d", np.name, c.name, got, c.d, c.n)
			}
		}
	}
}

// kSample is a point of a contour with its arc length along a run and the
// curvature there.
type kSample struct {
	at outline.Point
	s  float64
	k  float64
}

// smoothRuns cuts each contour at its corners (a turn of more than two degrees
// between two segments, or a straight segment) into runs of curves, sampled 100
// times on each segment.
func smoothRuns(path outline.Path) [][]kSample {
	var out [][]kSample
	for _, ct := range path {
		n := len(ct.Segs)
		corner := make([]bool, n)
		start := -1
		for i := 0; i < n; i++ {
			prev, next := segCubic(ct, i+n-1), segCubic(ct, i)
			tin, tout := prev.tangent(1), next.tangent(0)
			if !ct.Segs[(i+n-1)%n].IsCurve {
				tin = unit(prev.P0, prev.P3)
			}
			if !ct.Segs[i].IsCurve {
				tout = unit(next.P0, next.P3)
			}
			a := math.Abs(math.Atan2(tin.X*tout.Y-tin.Y*tout.X, tin.X*tout.X+tin.Y*tout.Y)) * 180 / math.Pi
			corner[i] = a > 2 || !ct.Segs[i].IsCurve || !ct.Segs[(i+n-1)%n].IsCurve
			if corner[i] && start < 0 {
				start = i
			}
		}
		if start < 0 {
			start = 0
		}
		var cur []kSample
		s := 0.0
		for j := 0; j < n; j++ {
			i := (start + j) % n
			if corner[i] && len(cur) > 0 {
				out = append(out, cur)
				cur, s = nil, 0
			}
			if !ct.Segs[i].IsCurve {
				continue
			}
			c := segCubic(ct, i)
			prev := c.at(0)
			for m := 0; m <= 100; m++ {
				u := float64(m) / 100
				q := c.at(u)
				s += math.Hypot(q.X-prev.X, q.Y-prev.Y)
				prev = q
				cur = append(cur, kSample{q, s, c.curvature(u)})
			}
		}
		if len(cur) > 0 {
			out = append(out, cur)
		}
	}
	return out
}

// signChanges returns the number of changes of the sign of the curvature along
// the samples, counting only curvatures of more than kFlat either way, and the
// turning (radians) of the parts which turn against the sign of their run of
// samples by less than reverseRun: the size of the small reverse bends.
func signChanges(run []kSample) int {
	count, sign := 0, 0.0
	for _, q := range run {
		if math.Abs(q.k) < kFlat {
			continue
		}
		sg := math.Copysign(1, q.k)
		if sign != 0 && sg != sign {
			count++
		}
		sign = sg
	}
	return count
}

// kFlat is the curvature below which an edge counts as straight (a radius of
// 3000 units).
const kFlat = 1.0 / 3000

// crotches returns the corners of the path where the white enters the ink (a
// turn to the right of more than cornerTurn on a contour which has the ink on
// its left) with the angle of the white there in degrees.
func crotches(path outline.Path) (at []outline.Point, angle []float64) {
	for _, ct := range path {
		nodeTurns(ct, func(q outline.Point, turn float64, _, _ cubic, _ bool) {
			if turn < -cornerTurn {
				at = append(at, q)
				angle = append(angle, 180+turn)
			}
		})
	}
	return at, angle
}

// The ampersand crosses itself twice, and every crotch where its strokes cross
// is a clean V: at least ampCrotchMin, and at least ampUpperCrotchMin at the
// upper crossing (above ampUpperY of the height), where the thin stroke crosses
// the diagonal in nearly a right angle.  The counters of the loop and of the
// bowl stay open: at least ampLoopMinW and ampBowlMinW wide.
const (
	ampCrotchMin      = 45.0
	ampUpperCrotchMin = 65.0
	ampUpperY         = 0.45
	ampLoopMinW       = 130.0
	ampBowlMinW       = 190.0
)

func TestAmpersandCrossings(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		s := np.params.Sym
		path, _, _ := outlineOf("ampersand", np.params)
		at, ang := crotches(path)
		if len(at) < 8 {
			t.Errorf("%s: %d crotches, want the four of each crossing", np.name, len(at))
		}
		for i := range at {
			min := ampCrotchMin
			if at[i].Y > s.AmpBottom+ampUpperY*(s.AmpTop-s.AmpBottom) {
				min = ampUpperCrotchMin
			}
			if ang[i] < min {
				t.Errorf("%s: the crotch at (%.0f, %.0f) is %.1f degrees, want at least %v", np.name, at[i].X, at[i].Y, ang[i], min)
			}
		}
		if len(path) != 3 {
			t.Fatalf("%s: %d contours, want the outline, the bowl and the loop", np.name, len(path))
		}
		var loopW, bowlW float64
		for _, ct := range path[1:] {
			x0, y0, x1, y1 := contourBox(ct)
			if (y0+y1)/2 > (s.AmpTop+s.AmpBottom)/2 {
				loopW = x1 - x0
			} else {
				bowlW = x1 - x0
			}
		}
		if loopW < ampLoopMinW || bowlW < ampBowlMinW {
			t.Errorf("%s: the loop is %.0f wide and the bowl %.0f, want at least %v and %v", np.name, loopW, bowlW, ampLoopMinW, ampBowlMinW)
		}
	}
}

// The ball of 6 stands clear of the bowl, with at least sixBallWhite of white
// below it on the vertical through its centre; the 9 is the same turned.
const sixBallWhite = 45.0

func TestSixBallStandsClear(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		r := np.params.Figures.Ball / 2
		for _, name := range []string{"six", "nine"} {
			path, _, _ := outlineOf(name, np.params)
			balls := findBalls(path, r)
			if len(balls) != 1 {
				t.Fatalf("%s, %s: %d balls", np.name, name, len(balls))
			}
			b := balls[0]
			// Away from the stalk: down for 6, up for 9.
			dir := -1.0
			if name == "nine" {
				dir = 1
			}
			edge := b.Y + dir*r
			gap := math.Inf(1)
			for _, y := range path.CrossingsY(b.X) {
				if d := (y - edge) * dir; d > 0.5 && d < gap {
					gap = d
				}
			}
			if gap < sixBallWhite {
				t.Errorf("%s, %s: %.1f of white beyond the ball, want at least %v", np.name, name, gap, sixBallWhite)
			}
		}
	}
}

// The hook of the question mark has the vertical stress of the face: its
// thickest part lies at its right extreme, and the sweep down to the stem is at
// most qSweepRatio of that.  From the right extreme down to the foot of the stem
// each edge bends one way, passes through at most one inflection, where the
// sweep turns into the stem, and bends the other way, with no second change of
// bending; any bend against the run it lies in turns by at most qReverse
// radians in all.
const (
	qSweepRatio = 0.74
	qReverse    = 0.01
)

func TestQuestionHook(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("question", np.params)
		hook := path[0]
		pts := sampleContour(hook, 0.5)
		n := len(pts)
		_, hy0, _, _ := contourBox(hook)
		ext := 0
		for i, q := range pts {
			if q.X > pts[ext].X {
				ext = i
			}
		}
		yExt := pts[ext].Y
		// The thickness at a point of the outer edge: the distance to the
		// nearest point of the contour more than 100 units along it.
		thick := func(i int) float64 {
			best := math.Inf(1)
			for j := range pts {
				d := (j - i + n) % n
				if d < 200 || n-d < 200 {
					continue
				}
				best = math.Min(best, math.Hypot(pts[j].X-pts[i].X, pts[j].Y-pts[i].Y))
			}
			return best
		}
		// Walk the outer edge from the right extreme down to the foot; the
		// contour runs counter-clockwise, so down the right side is backwards.
		tExt := thick(ext)
		sweepMax := 0.0
		for k := 4; k < n; k += 4 {
			i := (ext - k + n) % n
			q := pts[i]
			if q.Y < hy0+5 {
				break
			}
			th := thick(i)
			if q.Y < yExt-0.4*(yExt-hy0) && q.Y > hy0+0.4*(yExt-hy0) {
				sweepMax = math.Max(sweepMax, th)
			}
			if th > tExt+1 {
				t.Errorf("%s: the hook is %.1f thick at (%.0f, %.0f), more than %.1f at its right extreme", np.name, th, q.X, q.Y, tExt)
				break
			}
		}
		if sweepMax > qSweepRatio*tExt {
			t.Errorf("%s: the sweep is %.1f thick, the right extreme %.1f", np.name, sweepMax, tExt)
		}
		// The two edges between the height of the right extreme and the foot.
		for _, run := range smoothRuns(outline.Path{hook}) {
			var part []kSample
			for _, q := range run {
				if q.at.Y < yExt && q.at.Y > hy0+1 {
					part = append(part, q)
				}
			}
			if len(part) < 10 {
				continue
			}
			if c := signChanges(part); c > 1 {
				t.Errorf("%s: an edge of the hook from (%.0f, %.0f) changes its bending %d times below the right extreme, want at most once", np.name, part[0].at.X, part[0].at.Y, c)
			}
			var rev float64
			sign := 0.0
			for k := 1; k < len(part); k++ {
				if math.Abs(part[k].k) >= 10*kFlat {
					if sign != 0 && math.Copysign(1, part[k].k) != sign {
						rev = 0 // the one inflection: a new run
					}
					sign = math.Copysign(1, part[k].k)
				} else if sign != 0 && part[k].k*sign < 0 {
					rev += math.Abs(part[k].k) * (part[k].s - part[k-1].s)
				}
			}
			if rev > qReverse {
				t.Errorf("%s: an edge of the hook turns back by %.3f radians, want at most %v", np.name, rev, qReverse)
			}
		}
	}
}
