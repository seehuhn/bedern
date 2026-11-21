// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ball terminal.
//
// The tests measure the terminal on the built glyphs of a, c, f, j, r and y,
// unrounded and rounded.  The placement of the part in each letter is used only
// to find the terminal and to say which way it faces; the measurements are taken
// from the outline: the circle of the ball is fitted to the outline, the tip of the
// wedge is the sharpest concave corner near the ball, and the thickness of the
// stalk is the distance between its edges.

// ballLetters are the letters with a ball terminal.
var ballLetters = []string{"a", "c", "f", "j", "r", "y"}

// terminalOf returns the ball terminal placed in the letter for p.
func terminalOf(p design.Params, letter string) placedBall {
	switch letter {
	case "a":
		return aTerminal(p)
	case "c":
		tb, _ := cTerminal(p, newCFrame(p))
		return tb
	case "f":
		return fTerminal(p)
	case "j":
		return jTerminal(p)
	case "r":
		return rTerminal(p)
	case "y":
		return yTerminal(p)
	case "g":
		return gFrameOf(p).ear.part
	}
	panic("terminalOf: " + letter)
}

// turn returns the angle from a to b, in (-pi, pi].
func turn(a, b float64) float64 {
	d := math.Mod(b-a, 2*math.Pi)
	if d > math.Pi {
		d -= 2 * math.Pi
	} else if d <= -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

// A terminalSample is the outer contour of a letter sampled along its arc length,
// in the frame of its ball terminal: the outline is mapped back by the inverse of
// the placement which the measured centre and tip give.
type terminalSample struct {
	pts []outline.Point // the contour, every 0.25 units of arc length, in the frame of the part
	tip int             // the index of the tip of the wedge
	r   float64         // the radius of the circle fitted to the ball
}

// sampleContour returns the points of the contour every step units of arc length.
func sampleContour(ct outline.Contour, step float64) []outline.Point {
	var pts []outline.Point
	cur := ct.Start
	for _, s := range ct.Segs {
		c := cubic{cur, s.P1, s.P2, s.P3}
		if !s.IsCurve {
			c = cubic{cur, cur, s.P3, s.P3}
		}
		const n = 400
		prev := c.at(0)
		acc := 0.0
		if len(pts) == 0 {
			pts = append(pts, prev)
		}
		for i := 1; i <= n; i++ {
			q := c.at(float64(i) / n)
			acc += math.Hypot(q.X-prev.X, q.Y-prev.Y)
			for acc >= step {
				// The point is placed by linear interpolation within the small
				// step.
				f := 1 - (acc-step)/math.Max(1e-12, math.Hypot(q.X-prev.X, q.Y-prev.Y))
				pts = append(pts, outline.Point{X: prev.X + f*(q.X-prev.X), Y: prev.Y + f*(q.Y-prev.Y)})
				acc -= step
			}
			prev = q
		}
		cur = s.P3
	}
	return pts
}

// fitCircle returns the centre and radius of the circle which fits the points
// best in the least squares sense (the algebraic fit of Kasa).
func fitCircle(pts []outline.Point) (outline.Point, float64) {
	var sx, sy, sxx, syy, sxy, sxz, syz, sz float64
	n := float64(len(pts))
	for _, q := range pts {
		z := q.X*q.X + q.Y*q.Y
		sx += q.X
		sy += q.Y
		sxx += q.X * q.X
		syy += q.Y * q.Y
		sxy += q.X * q.Y
		sxz += q.X * z
		syz += q.Y * z
		sz += z
	}
	// Solve [sxx sxy sx; sxy syy sy; sx sy n] [a b c] = [sxz syz sz] for the
	// circle x^2 + y^2 = a x + b y + c.
	m := [3][4]float64{{sxx, sxy, sx, sxz}, {sxy, syy, sy, syz}, {sx, sy, n, sz}}
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			f := m[j][i] / m[i][i]
			for k := i; k < 4; k++ {
				m[j][k] -= f * m[i][k]
			}
		}
	}
	var v [3]float64
	for i := 2; i >= 0; i-- {
		v[i] = m[i][3]
		for k := i + 1; k < 3; k++ {
			v[i] -= m[i][k] * v[k]
		}
		v[i] /= m[i][i]
	}
	c := outline.Point{X: v[0] / 2, Y: v[1] / 2}
	return c, math.Sqrt(v[2] + c.X*c.X + c.Y*c.Y)
}

// measureTerminal samples the outer contour of the letter, finds the ball and the
// tip of the wedge on it and maps it into the frame of the part.  The ball is the
// run of the contour which lies within 2 units of the circle of the placed part
// and is longer than half its circumference; the circle is fitted to that run.  The
// tip is the point of the contour within 3 units of the circle where the contour
// turns most to the right over 3 units on either side, within 6 units of the tip of
// the placed part.
func measureTerminal(t *testing.T, p design.Params, letter string, path outline.Path) terminalSample {
	t.Helper()
	tb := terminalOf(p, letter)
	pts := sampleContour(path[0], 0.25)
	n := len(pts)
	R := tb.Part.R
	near := func(q outline.Point, tol float64) bool {
		return math.Abs(math.Hypot(q.X-tb.Centre.X, q.Y-tb.Centre.Y)-R) < tol
	}
	var arc []outline.Point
	for _, q := range pts {
		if near(q, 2) {
			arc = append(arc, q)
		}
	}
	if float64(len(arc))*0.25 < math.Pi*R {
		t.Fatalf("%s: only %.0f units of the contour lie on the circle of the ball", letter, float64(len(arc))*0.25)
	}
	centre, r := fitCircle(arc)
	// A second fit to the points which lie on the first circle leaves out the
	// beginning of the stalk.
	arc = arc[:0]
	for _, q := range pts {
		if math.Abs(math.Hypot(q.X-centre.X, q.Y-centre.Y)-r) < 0.75 {
			arc = append(arc, q)
		}
	}
	centre, r = fitCircle(arc)
	// The tip: the sharpest turn to the right near the circle.
	const k = 12 // 3 units on either side
	tip, worst := -1, 0.0
	for i := range pts {
		q := pts[i]
		if math.Abs(math.Hypot(q.X-centre.X, q.Y-centre.Y)-r) > 3 || math.Hypot(q.X-tb.Tip.X, q.Y-tb.Tip.Y) > 6 {
			continue
		}
		a, b, c := pts[(i-k+n)%n], q, pts[(i+k)%n]
		tr := turn(math.Atan2(b.Y-a.Y, b.X-a.X), math.Atan2(c.Y-b.Y, c.X-b.X))
		if tr < worst {
			tip, worst = i, tr
		}
	}
	if tip < 0 {
		t.Fatalf("%s: no corner of the wedge on the circle of the ball", letter)
	}
	// The frame of the part: the centre at the origin, the tip at the angle of W, and
	// mirrored for a letter which mirrors the part.
	w := tb.Part.BallW
	at := math.Atan2(pts[tip].Y-centre.Y, pts[tip].X-centre.X)
	mirror := tb.Place.Mirror
	rot := w - at
	if mirror {
		rot = at - w
	}
	out := make([]outline.Point, n)
	for i, q := range pts {
		dx, dy := q.X-centre.X, q.Y-centre.Y
		if mirror {
			// Undo the turn, then the mirror.
			c, s := math.Cos(-at+(math.Pi-w)), math.Sin(-at+(math.Pi-w))
			x, y := c*dx-s*dy, s*dx+c*dy
			out[i] = outline.Point{X: -x, Y: y}
		} else {
			c, s := math.Cos(rot), math.Sin(rot)
			out[i] = outline.Point{X: c*dx - s*dy, Y: s*dx + c*dy}
		}
	}
	// The mirrored contour runs the other way round the ball.
	if mirror {
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
		tip = n - 1 - tip
	}
	return terminalSample{pts: out, tip: tip, r: r}
}

// at returns the sample k steps along the contour from the tip (negative: before
// it), in the frame of the part.  Along the contour, in the frame of the part, the
// inner edge of the stalk comes before the tip, and the ball after it.
func (s terminalSample) at(k int) outline.Point {
	n := len(s.pts)
	return s.pts[((s.tip+k)%n+n)%n]
}

// terminalMeasures holds the measures of the terminal.
type terminalMeasures struct {
	diameter float64 // the diameter of the circle fitted to the ball
	maxDev   float64 // the largest distance of the free part of the ball from the circle
	attach   float64 // the angle round the ball, seen from its centre, from the point where the outer edge leaves the circle to the tip of the wedge, in degrees
	wedge    float64 // the opening of the wedge at its tip, in degrees
	stalk    float64 // the thickness of the stalk at the tip of the wedge
}

// measures returns the measures of the terminal.  The outer edge leaves the circle
// where, walking from the tip round the ball, the contour first lies more than three
// units beyond it; the free part of
// the ball is the run before that, and its deviation is measured where it lies
// within half a unit outside the circle or inside it.  The opening of the wedge is
// the angle between the chords from the tip to the points 12 units along the contour
// on either side.  The thickness of the stalk at the tip is the
// distance from the tip to the outer edge, the contour after the point where it
// leaves the circle.
func (s terminalSample) measures() terminalMeasures {
	var m terminalMeasures
	m.diameter = 2 * s.r
	n := len(s.pts)
	leave := -1
	for k := 8; k < n; k++ {
		q := s.at(k)
		d := math.Hypot(q.X, q.Y) - s.r
		if d > 3 {
			leave = k
			break
		}
		if k > 8 && d < 0.5 {
			m.maxDev = math.Max(m.maxDev, math.Abs(d))
		}
	}
	tip, a := s.at(0), s.at(leave)
	m.attach = 360 - math.Mod(turn(math.Atan2(tip.Y, tip.X), math.Atan2(a.Y, a.X))*180/math.Pi+360, 360)
	b, c := s.at(-48), s.at(48)
	u, v := math.Atan2(b.Y-tip.Y, b.X-tip.X), math.Atan2(c.Y-tip.Y, c.X-tip.X)
	m.wedge = math.Abs(turn(u, v)) * 180 / math.Pi
	m.stalk = math.Inf(1)
	for k := leave; k < leave+int(4*s.r/0.25); k++ {
		q := s.at(k)
		m.stalk = math.Min(m.stalk, math.Hypot(q.X-tip.X, q.Y-tip.Y))
	}
	return m
}

// ballTolerances are the differences which TestBallTerminalIsOnePart allows between
// two letters, for the outline as drawn and rounded: the diameter and the thickness
// of the stalk in units, the attachment angle and the opening of the wedge in
// degrees.
var ballTolerances = map[bool][4]float64{false: {0.3, 0.5, 1.5, 0.3}, true: {0.6, 3, 4, 1}}

// TestBallTerminalIsOnePart measures the terminal on the six letters and requires
// that it is the same object: the same diameter, the same attachment angle, the
// same opening of the wedge and the same thickness of the stalk at the tip, and
// that the outlines near the ball coincide after the turn, mirror and shift which
// map the measured centre and tip onto each other.
func TestBallTerminalIsOnePart(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, round := range []bool{false, true} {
		var ref terminalMeasures
		var refS terminalSample
		for i, letter := range ballLetters {
			path, _, _ := outlineOf(letter, p)
			if round {
				path = path.Round()
			}
			s := measureTerminal(t, p, letter, path)
			m := s.measures()
			t.Logf("rounded %v, %s: diameter %.2f (deviation %.2f), attachment %.1f degrees, wedge %.1f degrees, stalk %.2f", round, letter, m.diameter, m.maxDev, m.attach, m.wedge, m.stalk)
			if i == 0 {
				ref, refS = m, s
				continue
			}
			check := func(what string, got, want, tol float64) {
				if math.Abs(got-want) > tol {
					t.Errorf("rounded %v, %s: %s %.2f, but %.2f on %s", round, letter, what, got, want, ballLetters[0])
				}
			}
			// The rounding of the outline to integers moves the tip and the
			// nodes by up to 0.7 units.
			tol := ballTolerances[round]
			check("diameter", m.diameter, ref.diameter, tol[0])
			check("attachment angle", m.attach, ref.attach, tol[1])
			check("opening of the wedge", m.wedge, ref.wedge, tol[2])
			check("thickness of the stalk at the tip", m.stalk, ref.stalk, tol[3])
			// The outline from a third of the way back along the section to the far
			// side of the ball coincides with that of the first letter.
			worst := 0.0
			for k := -int(0.5 * s.r / 0.25); k <= int((1.48*math.Pi+0.8)*s.r/0.25); k++ {
				q := s.at(k)
				best := math.Inf(1)
				for j := k - 40; j <= k+40; j++ {
					o := refS.at(j)
					best = math.Min(best, math.Hypot(q.X-o.X, q.Y-o.Y))
				}
				worst = math.Max(worst, best)
			}
			if limit := map[bool]float64{false: 0.75, true: 1.5}[round]; worst > limit {
				t.Errorf("rounded %v, %s: the outline near the ball lies up to %.2f from that of %s, want at most %v", round, letter, worst, ballLetters[0], limit)
			}
		}
	}
}

// The ear of g is the same part as the terminals of the six letters, built by the
// same construction at the size of the ear: the ball is 72 across on a stalk 26
// thick, with the angle of the blend scaled with the root of the diameter.  The
// test measures the ear on the outline, rounded and not: the diameter, the stalk
// at the tip and the opening of the wedge, which is that of the shared part.  The
// stalk is 0.36 of the diameter thick (0.35 on the other letters) and is
// attached over 66 degrees of the ball (99 on the other letters); the shape
// constant of the ear is gEarCurl.
func TestGEarIsTheBallTerminal(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	shared := newBallPart(p)
	for _, round := range []bool{false, true} {
		path, _, _ := outlineOf("g", p)
		if round {
			path = path.Round()
		}
		m := measureTerminal(t, p, "g", path).measures()
		tol := ballTolerances[round]
		t.Logf("rounded %v, g: diameter %.2f, attachment %.1f degrees, wedge %.1f degrees, stalk %.2f", round, m.diameter, m.attach, m.wedge, m.stalk)
		if math.Abs(m.diameter-p.GEarBall) > tol[0] {
			t.Errorf("rounded %v: diameter %.2f, want %v", round, m.diameter, p.GEarBall)
		}
		if math.Abs(m.stalk-p.GEarThick) > tol[1]+0.5 {
			t.Errorf("rounded %v: stalk %.2f, want %v", round, m.stalk, p.GEarThick)
		}
		if math.Abs(m.wedge-shared.wedge()) > tol[2]+1 {
			t.Errorf("rounded %v: wedge %.1f degrees, the shared part has %.1f", round, m.wedge, shared.wedge())
		}
	}
	ear, err := gEarPart(p)
	if err != nil {
		t.Fatal(err)
	}
	if d := math.Abs(ear.wedge() - shared.wedge()); d > 3 {
		t.Errorf("wedge of the ear %.2f degrees, of the shared part %.2f", ear.wedge(), shared.wedge())
	}
	if ear.R != p.GEarBall/2 || ear.T != p.GEarThick {
		t.Errorf("the part of the ear has the radius %v and the stalk %v", ear.R, ear.T)
	}
}

// The ball is a circle: the free part of its outline lies within half a unit of
// the fitted circle, unrounded and rounded, on every letter.
func TestBallIsACircle(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, letter := range ballLetters {
		path, _, _ := outlineOf(letter, p)
		for _, pa := range []outline.Path{path, path.Round()} {
			m := measureTerminal(t, p, letter, pa).measures()
			if m.maxDev > 0.5 {
				t.Errorf("%s: the ball lies up to %.2f from its circle, want at most 0.5", letter, m.maxDev)
			}
			if math.Abs(m.diameter-p.BallDiameter) > 1 {
				t.Errorf("%s: the ball has the diameter %.2f, want %v", letter, m.diameter, p.BallDiameter)
			}
		}
	}
}

// outerEdge returns the part of the contour from the node a to the node b (the
// nodes are named by the points near them) as a polyline, sampled finely, with the
// arc length of each point.
func outerEdge(t *testing.T, path outline.Path, a, b outline.Point) (pts []outline.Point, s []float64) {
	t.Helper()
	ct := path[0]
	n := len(ct.Segs)
	ia, ib := nodeNear(t, ct, a), nodeNear(t, ct, b)
	length := 0.0
	for i := ia; ; i++ {
		c := segCubic(ct, i)
		for k := 0; k < 200; k++ {
			q := c.at(float64(k) / 200)
			if len(pts) > 0 {
				length += math.Hypot(q.X-pts[len(pts)-1].X, q.Y-pts[len(pts)-1].Y)
			}
			pts = append(pts, q)
			s = append(s, length)
		}
		if (i+1)%n == ib%n {
			break
		}
	}
	q := ct.Segs[(ib+n-1)%n].P3
	length += math.Hypot(q.X-pts[len(pts)-1].X, q.Y-pts[len(pts)-1].Y)
	return append(pts, q), append(s, length)
}

// slidingTurns returns the least turn of the tangent, in radians, over a stretch
// of the polyline whose arc length is window, and where it starts.
func slidingTurns(pts []outline.Point, s []float64, window float64) (least float64, at outline.Point) {
	heading := make([]float64, len(pts))
	for i := 1; i < len(pts); i++ {
		heading[i] = math.Atan2(pts[i].Y-pts[i-1].Y, pts[i].X-pts[i-1].X)
	}
	heading[0] = heading[1]
	cum := make([]float64, len(pts))
	for i := 1; i < len(pts); i++ {
		cum[i] = cum[i-1] + turn(heading[i-1], heading[i])
	}
	least = math.Inf(1)
	j := 0
	for i := range pts {
		for j < len(pts) && s[j]-s[i] < window {
			j++
		}
		if j == len(pts) {
			break
		}
		if d := math.Abs(cum[j] - cum[i]); d < least {
			least, at = d, pts[i]
		}
	}
	return least, at
}

// stalkEdges returns the two edges of the stalk of the letter as the ends of runs of
// the contour, in the order along the contour: the inner edge from the point where
// it leaves the stem, stroke or o to the tip of the wedge, and the outer edge from
// the point where it leaves the circle of the ball to the point where it meets the
// stem (on c: the top extreme of the o, which the outer edge of the stalk continues).
// On a the contour runs the other way: the outer edge from the stem to the circle,
// and the inner edge from the tip to the stem; the runs are given in the order of
// the contour.
func stalkEdges(p design.Params, letter string) (inner, outer [2]outline.Point) {
	tb := terminalOf(p, letter)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	switch letter {
	case "r":
		x1 := rStem(p).right()
		return [2]outline.Point{pt(x1, p.RJoin), tb.Tip}, [2]outline.Point{tb.Tangent, pt(x1, p.RNotch)}
	case "f":
		h := fHook(p)
		return [2]outline.Point{pt(h.XR, h.InnerJoin), tb.Tip}, [2]outline.Point{tb.Tangent, pt(h.XL, h.OuterJoin)}
	case "c":
		f := newCFrame(p)
		_, tI := cInner(f, tb)
		return [2]outline.Point{f.innerU.cubic(3, false).at(tI), tb.Tip}, [2]outline.Point{tb.Tangent, pt(f.cx, p.XHeight+p.OvershootLower)}
	case "a":
		left := design.AStemLeft(p)
		return [2]outline.Point{tb.Tip, pt(left, p.ShoulderStartInner+aInnerRaise)}, [2]outline.Point{pt(left+p.Stem, p.ShoulderStart-aOuterDrop), tb.Tangent}
	case "j":
		s, h := jStem(p), jHook(p)
		return [2]outline.Point{pt(s.Left, -h.InnerJoin), tb.Tip}, [2]outline.Point{tb.Tangent, pt(s.right(), -h.OuterJoin)}
	case "y":
		h, r := yHook(p), yFrame(p).Right
		return [2]outline.Point{r.edgeAt(edgeLeft, -h.InnerJoin), tb.Tip}, [2]outline.Point{tb.Tangent, r.edgeAt(edgeRight, -h.OuterJoin)}
	}
	panic("stalkEdges: " + letter)
}

// stalkProfile returns the thickness of the stalk at the arc lengths s from the tip
// of the wedge, back along the inner edge: the distance from the point of the inner
// edge to the outer edge.
func stalkProfile(t *testing.T, p design.Params, letter string, path outline.Path) (s, th []float64) {
	t.Helper()
	in, out := stalkEdges(p, letter)
	ipts, is := outerEdge(t, path, in[0], in[1])
	opts, _ := outerEdge(t, path, out[0], out[1])
	tipFirst := letter == "a" // on a the inner edge starts at the tip
	total := is[len(is)-1]
	for i, q := range ipts {
		d := total - is[i]
		if tipFirst {
			d = is[i]
		}
		best := math.Inf(1)
		for _, o := range opts {
			best = math.Min(best, math.Hypot(q.X-o.X, q.Y-o.Y))
		}
		s = append(s, d)
		th = append(th, best)
	}
	return s, th
}

// edgeCurvature returns the signed curvature of the contour from the node a to the
// node b, sampled 50 times on each segment, with the arc length.
func edgeCurvature(t *testing.T, path outline.Path, a, b outline.Point) (s, k []float64) {
	t.Helper()
	ct := path[0]
	n := len(ct.Segs)
	ia, ib := nodeNear(t, ct, a), nodeNear(t, ct, b)
	length := 0.0
	for i := ia; ; i++ {
		c := segCubic(ct, i)
		prev := c.at(0)
		for j := 0; j <= 50; j++ {
			tt := float64(j) / 50
			q := c.at(tt)
			length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
			prev = q
			s = append(s, length)
			k = append(k, c.curvature(tt))
		}
		if (i+1)%n == ib%n {
			break
		}
	}
	return s, k
}

// ballClearance returns the narrowest white between the ball and the rest of the
// letter: the least distance from a point of the free part of the ball, more than
// a radius along the contour from the tip of the wedge, to a point of the contours
// which is not the ball, its stalk within the part, or the wedge: on the outer
// contour, a point more than ballApart along the contour from the tip on the side
// of the inner edge, and more than the length of the free part of the ball and the
// part's outer edge on the other side.
func ballClearance(t *testing.T, p design.Params, letter string, path outline.Path) (float64, outline.Point) {
	t.Helper()
	s := measureTerminal(t, p, letter, path)
	tb := terminalOf(p, letter)
	pts := sampleContour(path[0], 0.25)
	n := len(pts)
	tip := 0
	for i, q := range pts {
		if math.Hypot(q.X-tb.Tip.X, q.Y-tb.Tip.Y) < math.Hypot(pts[tip].X-tb.Tip.X, pts[tip].Y-tb.Tip.Y) {
			tip = i
		}
	}
	dir := 1
	if tb.Place.Mirror {
		dir = -1
	}
	// The samples of the outer contour which count, by their index steps from the
	// tip along the contour towards the ball (positive) or along the inner edge.
	ball := int((2*math.Pi*s.r*(1-tb.Part.attach()/(2*math.Pi)) + tb.Part.Ro*tb.Part.PhiS) / 0.25)
	apart := int(ballApart * s.r / 0.25)
	var others []outline.Point
	for k := -n / 2; k < n/2; k++ {
		if k > -apart && k < ball {
			continue
		}
		others = append(others, pts[((tip+dir*k)%n+n)%n])
	}
	for _, ct := range path[1:] {
		others = append(others, sampleContour(ct, 1)...)
	}
	best, at := math.Inf(1), outline.Point{}
	for k := int(s.r / 0.25); k < int(1.4*math.Pi*s.r/0.25); k += 4 {
		q := pts[((tip+dir*k)%n+n)%n]
		for _, o := range others {
			if d := math.Hypot(q.X-o.X, q.Y-o.Y); d < best {
				best, at = d, q
			}
		}
	}
	return best, at
}

// ballApart is the length of the inner edge, from the tip of the wedge, which
// belongs to the wedge, as a multiple of the radius of the ball.
const ballApart = 1.0

// The white between the ball and the rest of the letter, outside the wedge, is at
// least twice the clearance floor in every parameter set.  In the regular
// parameters it is 47.4 (y) to 52.7 (c).
func TestBallStandsClear(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, letter := range ballLetters {
				path, _ := cachedOutline(np, letter)
				d, at := ballClearance(t, np.params, letter, path)
				if want := 2 * design.ClearanceFloor(np.params); d < want {
					t.Errorf("parameter set %s, %s: the white next to the ball at (%.0f, %.0f) is %.1f wide, want at least %.1f", np.name, letter, at.X, at.Y, d, want)
				}
			}
		})
	}
}

// The stalk has an even thickness.  Measured from the tip of the wedge back along
// the inner edge, its thickness stays within stalkBand of BallStalk over the
// length stalkEven of each letter, where the stalk runs on its own (on f and j it
// grows out of the stem sooner, and on r it thins into the join of the arm as the
// arch of n does), and it changes nowhere by more than stalkRate units per unit of
// length, up to the stem.  In the regular parameters the thickness lies between
// 27.5 (r, near the join) and 38.2 (y, near the diagonal), and the largest rate
// is 0.88 (r, where the arm leaves the stem).
var (
	stalkBand = [2]float64{-8, 5}
	stalkEven = map[string]float64{"a": 100, "c": 120, "f": 45, "j": 45, "r": 110, "y": 140}
	stalkRate = 1.0
)

func TestStalkIsEven(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, letter := range ballLetters {
				path, _ := cachedOutline(np, letter)
				s, th := stalkProfile(t, p, letter, path)
				lo, hi := p.BallStalk+stalkBand[0], p.BallStalk+stalkBand[1]
				for i := range s {
					if s[i] <= stalkEven[letter] && (th[i] < lo || th[i] > hi) {
						t.Errorf("parameter set %s, %s: the stalk is %.1f thick %.0f from the tip of the wedge, want %.1f to %.1f", np.name, letter, th[i], s[i], lo, hi)
						break
					}
					if i > 0 {
						if ds := math.Abs(s[i] - s[i-1]); ds > 0.3 && math.Abs(th[i]-th[i-1])/ds > stalkRate {
							t.Errorf("parameter set %s, %s: the thickness of the stalk changes by %.2f per unit %.0f from the tip, want at most %v", np.name, letter, math.Abs(th[i]-th[i-1])/ds, s[i], stalkRate)
							break
						}
					}
				}
			}
		})
	}
}

// The edges of the stalk are arcs of evenly changing curvature.  On both edges,
// from the stem, stroke or o to the ball, no stretch of stretchLength units turns by
// less than minStretchTurn (no straight stretch), the contour turns one way only
// (the sum of its turns the wrong way is below edgeWrongWay; no bump), and the
// absolute curvature has at most one local maximum inside the edge (no local flat
// between two bends), see curvatureMaxima.
const (
	minStretchTurn      = 0.045
	stretchLength       = 40.0
	edgeWrongWay        = 0.01
	curvatureHysteresis = 0.0002
	curvatureRelative   = 0.06
)

func TestStalkEdgesAreEven(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, letter := range ballLetters {
				path, _ := cachedOutline(np, letter)
				in, out := stalkEdges(p, letter)
				if letter == "a" {
					// The stalk of a ends at the crest; from there to the stem
					// both edges are the shoulder
					// (TestShoulderRunsIntoTheStem).
					x, _ := aCrestAt(p, terminalOf(p, letter))
					top := p.XHeight + p.OvershootLower
					in[1] = outline.Point{X: x, Y: top - p.BallStalk}
					out[0] = outline.Point{X: x, Y: top}
				}
				for e, ends := range [][2]outline.Point{in, out} {
					what := map[int]string{0: "inner edge", 1: "outer edge"}[e]
					pts, ss := outerEdge(t, path, ends[0], ends[1])
					if least, at := slidingTurns(pts, ss, stretchLength); least < minStretchTurn {
						t.Errorf("parameter set %s, %s, %s: a stretch %.0f long at (%.0f, %.0f) turns by only %.3f radians, want at least %v", np.name, letter, what, stretchLength, at.X, at.Y, least, minStretchTurn)
					}
					s, k := edgeCurvature(t, path, ends[0], ends[1])
					sign := 1.0
					if k[len(k)/2] < 0 {
						sign = -1
					}
					var wrong float64
					prev := pts[0]
					h0 := math.NaN()
					for _, q := range pts[1:] {
						if math.Hypot(q.X-prev.X, q.Y-prev.Y) < 1e-9 {
							continue
						}
						h := math.Atan2(q.Y-prev.Y, q.X-prev.X)
						if !math.IsNaN(h0) {
							if d := turn(h0, h) * sign; d < 0 {
								wrong -= d
							}
						}
						h0, prev = h, q
					}
					if wrong > edgeWrongWay {
						t.Errorf("parameter set %s, %s, %s: the edge turns the wrong way by %.3f radians in all, want at most %v", np.name, letter, what, wrong, edgeWrongWay)
					}
					if m := curvatureMaxima(s, k); m > 1 {
						t.Errorf("parameter set %s, %s, %s: the curvature has %d local maxima, want at most one", np.name, letter, what, m)
					}
				}
			}
		})
	}
}

// curvatureMaxima returns the number of local maxima of the absolute curvature
// inside the run, by more than curvatureHysteresis plus curvatureRelative times the
// curvature on either side: a wave of a few per cent in an edge which is nearly a
// circular arc does not count.
func curvatureMaxima(s, k []float64) int {
	count := 0
	low := math.Abs(k[0])
	high, rising := low, false
	slack := func(v float64) float64 { return curvatureHysteresis + curvatureRelative*v }
	for _, v := range k[1:] {
		v = math.Abs(v)
		switch {
		case !rising && v > low+slack(low):
			rising, high = true, v
		case rising && v > high:
			high = v
		case rising && v < high-slack(high):
			count++
			rising, low = false, v
		case !rising && v < low:
			low = v
		}
	}
	return count
}

// The outer edge passes from the stalk into the ball without a corner and with the
// curvature of the ball: at the node where it runs into the circle the direction
// does not change, and the curvature on the two sides differs by less than 2 per
// cent.
func TestOuterEdgeRunsIntoTheBall(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, letter := range ballLetters {
		path, _, _ := outlineOf(letter, p)
		tb := terminalOf(p, letter)
		ct := path[0]
		i := nodeNear(t, ct, tb.Tangent)
		n := len(ct.Segs)
		before, after := segCubic(ct, i+n-1), segCubic(ct, i)
		a, b := before.tangent(1), after.tangent(0)
		if d := math.Abs(turn(math.Atan2(a.Y, a.X), math.Atan2(b.Y, b.X))); d > 1e-6 {
			t.Errorf("%s: the outer edge turns by %.2g radians where it runs into the ball", letter, d)
		}
		k0, k1 := math.Abs(before.curvature(1)), math.Abs(after.curvature(0))
		if math.Abs(k0-k1) > 0.02*math.Max(k0, k1) {
			t.Errorf("%s: the curvature changes from %.5f to %.5f where the outer edge runs into the ball", letter, k0, k1)
		}
	}
}

// The highest point of the top of a lies near the middle of the letter: between
// aCrestLow and aCrestHigh of the way from the leftmost ink to the right-hand edge
// of the stem.
const aCrestLow, aCrestHigh = 0.38, 0.55

func TestATopPeaksNearTheMiddle(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "a")
			x0, _, _, y1 := path.BBox()
			var crest outline.Point
			for _, q := range contourNodes(path[0]) {
				if q.Y > crest.Y {
					crest = q
				}
			}
			right := design.AStemLeft(p) + p.Stem
			f := (crest.X - x0) / (right - x0)
			if np.name == "regular" {
				t.Logf("the top of a peaks at %.3f of the way across, at the height %.1f (bounding box %.1f)", f, crest.Y, y1)
			}
			if f < aCrestLow || f > aCrestHigh {
				t.Errorf("parameter set %s: the top of a peaks at %.3f of the way from the left to the stem, want %v to %v", np.name, f, aCrestLow, aCrestHigh)
			}
		})
	}
}

// The inner edge of the hook of j and y, from the stem or the diagonal to the tip of
// the wedge, has no dent: its curvature grows to that of the part, where the edge
// joins it, and does not rise above it and fall back.  The edges of a, r and c
// show no such hump; that of f has its own profile.  The sets which move the joins
// and the reach of a hook (f-t-j and v-w-y) are left out for that letter.
func TestHookInnerEdgeHasNoDent(t *testing.T) {
	t.Parallel()
	const hump = 0.08 // the curvature may exceed that of the part by this fraction
	reg := design.Regular()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, letter := range []string{"j", "y"} {
				if letter == "j" && (p.JStemLeft != reg.JStemLeft || p.JReach != reg.JReach || p.JDepth != reg.JDepth || p.JOuterDrop != reg.JOuterDrop || p.JInnerDrop != reg.JInnerDrop) ||
					letter == "y" && (p.YDepth != reg.YDepth || p.YReach != reg.YReach || p.YOuterDrop != reg.YOuterDrop || p.YInnerDrop != reg.YInnerDrop) {
					continue
				}
				path, _ := cachedOutline(np, letter)
				in, _ := stalkEdges(np.params, letter)
				s, k := edgeCurvature(t, path, in[0], in[1])
				end := math.Abs(k[len(k)-1])
				for i := range k {
					if v := math.Abs(k[i]); v > end*(1+hump) {
						t.Errorf("parameter set %s, %s: the curvature of the inner edge is %.4f at %.0f units along it, %.0f per cent above that of the part (%.4f), want at most %.0f per cent", np.name, letter, v, s[i], 100*(v/end-1), end, 100*hump)
						break
					}
				}
			}
		})
	}
}
