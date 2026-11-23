// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// Tests of the section sign: the spines are not long and straight, the middle
// is not pinched, there is no ring in the middle and there are no kinks where a
// round meets a spine.

// sectSign holds the outline of the section sign of a parameter set with what
// the tests measure on it.
type sectSign struct {
	path  outline.Path
	balls []outline.Point
	r     float64
}

func newSectSign(t *testing.T, np namedParams) sectSign {
	path, _, _ := outlineOf("section", np.params)
	if len(path) != 2 {
		t.Fatalf("%s: %d contours, want the outline and the eye", np.name, len(path))
	}
	r := np.params.Figures.Ball / 2
	return sectSign{path: path, balls: findBalls(path, r), r: r}
}

// nearBall reports whether q lies within the reach of a ball terminal: the
// ball and its stalk, which the tests of the ball part cover.
func (g sectSign) nearBall(q outline.Point) bool {
	for _, b := range g.balls {
		if math.Hypot(q.X-b.X, q.Y-b.Y) < sectBallReach*g.r {
			return true
		}
	}
	return false
}

const sectBallReach = 2.6

// The middle is the widest part of the sign.  The width of the ink of each
// hook (the part above sectHookBand of the height, or below it from the
// bottom) is between sectHookMin and sectHookMax of the width of the middle
// (the ink between sectMidBand and 1 - sectMidBand of the height).  The
// extremes of each hook lie inside those of the middle, within
// sectHookOut.
const (
	sectHookBand = 0.70
	sectMidBand  = 0.35
	sectHookMin  = 0.70
	sectHookMax  = 0.88
	sectHookOut  = 2.0
)

func TestSectionMiddleWidest(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		g := newSectSign(t, np)
		_, y0, _, y1 := contourBox(g.path[0])
		h := y1 - y0
		band := func(lo, hi float64) (float64, float64) {
			a, b := math.Inf(1), math.Inf(-1)
			for _, q := range sampleContour(g.path[0], 1) {
				if q.Y >= y0+lo*h && q.Y <= y0+hi*h {
					a, b = math.Min(a, q.X), math.Max(b, q.X)
				}
			}
			return a, b
		}
		ml, mr := band(sectMidBand, 1-sectMidBand)
		for _, hk := range []struct {
			name   string
			lo, hi float64
		}{{"top", sectHookBand, 1}, {"bottom", 0, 1 - sectHookBand}} {
			l, r := band(hk.lo, hk.hi)
			if f := (r - l) / (mr - ml); f < sectHookMin || f > sectHookMax {
				t.Errorf("%s: the %s hook is %.2f of the width of the middle, want %v to %v", np.name, hk.name, f, sectHookMin, sectHookMax)
			}
			if l < ml-sectHookOut || r > mr+sectHookOut {
				t.Errorf("%s: the %s hook reaches from %.0f to %.0f, outside the middle (%.0f to %.0f)", np.name, hk.name, l, r, ml, mr)
			}
		}
	}
}

// The sign has point symmetry: every point of the outline, turned through half
// a turn about the middle of the ink, lies within sectSymTol of the outline.
const sectSymTol = 1.0

func TestSectionSymmetry(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		g := newSectSign(t, np)
		x0, y0, x1, y1 := contourBox(g.path[0])
		cx, cy := (x0+x1)/2, (y0+y1)/2
		var all []outline.Point
		for _, ct := range g.path {
			all = append(all, sampleContour(ct, 0.5)...)
		}
		worst := 0.0
		grid := newPointGrid(all, 8)
		for i := 0; i < len(all); i += 7 {
			q := outline.Point{X: 2*cx - all[i].X, Y: 2*cy - all[i].Y}
			worst = math.Max(worst, grid.nearest(q))
		}
		if worst > sectSymTol {
			t.Errorf("%s: the outline turned through half a turn lies up to %.2f from itself, want at most %v", np.name, worst, sectSymTol)
		}
	}
}

// The eye is one convex counter, at least sectEyeMinW wide and sectEyeMinH
// high; the spines carry the weight and the arcs which close the eye are thin
// strokes.  The spine is at least sectSpineMin thick at right angles to the
// eye's long side.  The arc on the left of the eye (and so, by the symmetry,
// the one on the right) is measured along rays from the middle of the eye,
// every sectArcStep degrees from the lower spine (sectArcFrom) round towards
// the upper one: from the lower spine the ink along the ray grows no thicker
// until the arc is at its thinnest, it is at most sectArcThin there and has
// come down to sectArcThin by sectArcShare of the sweep from sectArcFrom to
// the join, where the rays reach the upper spine; and between its thinnest place
// and the join it stays at most sectArcThin, so that it meets the upper spine
// without a swelling.  The join is the first ray beyond the thinnest place
// along which the ink is more than twice sectArcThin.
const (
	sectEyeMinW  = 245.0
	sectEyeMinH  = 230.0
	sectSpineMin = 64.0
	sectArcThin  = 40.0
	sectArcFrom  = 200.0
	sectArcStep  = 2.0
	sectArcShare = 0.6
)

func TestSectionEye(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		g := newSectSign(t, np)
		for _, run := range smoothRuns(outline.Path{g.path[1]}) {
			if n := signChanges(run); n > 0 {
				t.Errorf("%s: the eye is not convex (%d changes of the bending)", np.name, n)
			}
		}
		x0, y0, x1, y1 := contourBox(g.path[1])
		if x1-x0 < sectEyeMinW || y1-y0 < sectEyeMinH {
			t.Errorf("%s: the eye is %.0f by %.0f, want at least %v by %v", np.name, x1-x0, y1-y0, sectEyeMinW, sectEyeMinH)
		}
		eye := sampleContour(g.path[1], 1)
		outer := sampleContour(g.path[0], 0.5)
		// The spine: across the eye's upper right side where it runs at
		// sectSpineAngle, from the point of the eye furthest along the normal.
		th := sectSpineAngle * math.Pi / 180
		nrm := outline.Point{X: math.Sin(th), Y: math.Cos(th)}
		far := eye[0]
		for _, q := range eye {
			if q.X*nrm.X+q.Y*nrm.Y > far.X*nrm.X+far.Y*nrm.Y {
				far = q
			}
		}
		if spine := rayHit(outer, far, nrm); spine < sectSpineMin {
			t.Errorf("%s: the spine is %.0f thick across the eye's side, want at least %v", np.name, spine, sectSpineMin)
		}
		// The arc on the left.
		c := outline.Point{X: (x0 + x1) / 2, Y: (y0 + y1) / 2}
		var angles, ink []float64
		for deg := sectArcFrom; deg > sectArcFrom-180; deg -= sectArcStep {
			a := deg * math.Pi / 180
			dir := outline.Point{X: math.Cos(a), Y: math.Sin(a)}
			e := rayHit(eye, c, dir)
			from := outline.Point{X: c.X + e*dir.X, Y: c.Y + e*dir.Y}
			angles, ink = append(angles, deg), append(ink, rayHit(outer, from, dir))
		}
		iMin := 0
		for i := range ink {
			if ink[i] < ink[iMin] {
				iMin = i
			}
		}
		iJoin := -1
		for i := iMin; i < len(ink); i++ {
			if ink[i] > 2*sectArcThin {
				iJoin = i
				break
			}
		}
		if iJoin < 0 {
			t.Errorf("%s: the arc does not meet the upper spine", np.name)
			continue
		}
		for i := 1; i <= iMin; i++ {
			if ink[i] > ink[i-1]+0.5 {
				t.Errorf("%s: the arc grows thicker at %.0f degrees (%.1f after %.1f)", np.name, angles[i], ink[i], ink[i-1])
				break
			}
		}
		for i := iMin; i < iJoin; i++ {
			if ink[i] > sectArcThin {
				t.Errorf("%s: the arc is %.0f thick at %.0f degrees, before it joins the spine at %.0f, want at most %v", np.name, ink[i], angles[i], angles[iJoin], sectArcThin)
				break
			}
		}
		if i := int(sectArcShare * float64(iJoin)); ink[i] > sectArcThin {
			t.Errorf("%s: the arc is %.0f thick at %v of its sweep, want at most %v", np.name, ink[i], sectArcShare, sectArcThin)
		}
	}
}

// rayHit returns the distance from p along the unit vector dir to the nearest
// crossing of the closed polygon pts.
func rayHit(pts []outline.Point, p, dir outline.Point) float64 {
	best := math.Inf(1)
	n := len(pts)
	for i := 0; i < n; i++ {
		a, b := pts[i], pts[(i+1)%n]
		ex, ey := b.X-a.X, b.Y-a.Y
		den := dir.X*ey - dir.Y*ex
		if den == 0 {
			continue
		}
		wx, wy := a.X-p.X, a.Y-p.Y
		s := (wx*ey - wy*ex) / den
		u := (wx*dir.Y - wy*dir.X) / den
		if s > 1e-6 && u >= 0 && u <= 1 {
			best = math.Min(best, s)
		}
	}
	return best
}

// The edges are fair.  Away from the ball terminals, along every edge between
// corners, sampled every sectStep units of arc length, the curvatures of two
// neighbouring samples differ by at most sectKFactor - 1 times the smaller of
// them, plus the curvature sectKFloor (a radius of 650 units); an edge is nowhere
// nearly straight (a curvature below kFlat) over more than sectFlatLen units
// except across an inflection, which lies in a spine; and every cubic segment
// is at least sectMinSeg long.
const (
	sectStep    = 5.0
	sectKFactor = 1.5
	sectKFloor  = 1.0 / 650
	sectFlatLen = 40.0
	sectMinSeg  = 15.0
)

func TestSectionFairEdges(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		g := newSectSign(t, np)
		for _, run := range smoothRuns(g.path) {
			var samp []kSample
			next := 0.0
			for _, q := range run {
				if q.s >= next {
					samp = append(samp, q)
					next = q.s + sectStep
				}
			}
			for i := 1; i < len(samp); i++ {
				a, b := samp[i-1], samp[i]
				if g.nearBall(a.at) || g.nearBall(b.at) {
					continue
				}
				lo := math.Min(math.Abs(a.k), math.Abs(b.k))
				if math.Abs(a.k-b.k) > (sectKFactor-1)*lo+sectKFloor {
					t.Errorf("%s: the curvature jumps from %.2e to %.2e at (%.0f, %.0f)", np.name, a.k, b.k, b.at.X, b.at.Y)
					break
				}
			}
			// Flats: runs of nearly straight samples not across an inflection.
			for i := 0; i < len(run); {
				if math.Abs(run[i].k) >= kFlat || g.nearBall(run[i].at) {
					i++
					continue
				}
				j := i
				for j < len(run) && math.Abs(run[j].k) < kFlat {
					j++
				}
				if run[min(j, len(run)-1)].s-run[i].s > sectFlatLen {
					before, after := 0.0, 0.0
					if i > 0 {
						before = run[i-1].k
					}
					if j < len(run) {
						after = run[j].k
					}
					if before*after >= 0 {
						t.Errorf("%s: the edge is straight for %.0f units from (%.0f, %.0f)", np.name, run[min(j, len(run)-1)].s-run[i].s, run[i].at.X, run[i].at.Y)
					}
				}
				i = j
			}
		}
		for _, ct := range g.path {
			for i := range ct.Segs {
				c := segCubic(ct, i)
				if g.nearBall(c.P0) || g.nearBall(c.P3) {
					continue
				}
				if l := cubicLength(c); l < sectMinSeg {
					t.Errorf("%s: a segment from (%.0f, %.0f) is %.1f long, want at least %v", np.name, c.P0.X, c.P0.Y, l, sectMinSeg)
				}
			}
		}
	}
}

// The joints are clean: the outline has two crotches away from the balls,
// where the arc of one stroke runs into the spine of the other at a small
// angle, and each opens at most sectCrotchMax degrees; and the white between
// each ball and the spine below or above it is at least sectBallWhite straight
// towards the middle.
const (
	sectCrotchMax = 50.0
	sectBallWhite = 120.0
)

func TestSectionJoints(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		g := newSectSign(t, np)
		at, ang := crotches(g.path)
		n := 0
		for i := range at {
			if g.nearBall(at[i]) {
				continue
			}
			n++
			if ang[i] > sectCrotchMax {
				t.Errorf("%s: the crotch at (%.0f, %.0f) opens %.0f degrees, want at most %v", np.name, at[i].X, at[i].Y, ang[i], sectCrotchMax)
			}
		}
		if n != 2 {
			t.Errorf("%s: %d crotches away from the balls, want 2", np.name, n)
		}
		if len(g.balls) != 2 {
			t.Fatalf("%s: %d balls", np.name, len(g.balls))
		}
		_, y0, _, y1 := contourBox(g.path[0])
		outer := sampleContour(g.path[0], 0.5)
		for _, b := range g.balls {
			// The white straight from the ball towards the middle.
			dir := outline.Point{Y: math.Copysign(1, (y0+y1)/2-b.Y)}
			from := outline.Point{X: b.X, Y: b.Y + dir.Y*(g.r+0.5)}
			best := rayHit(outer, from, dir) + 0.5
			if best < sectBallWhite {
				t.Errorf("%s: the ball at (%.0f, %.0f) is %.0f from the spine, want at least %v", np.name, b.X, b.Y, best, sectBallWhite)
			}
		}
	}
}
