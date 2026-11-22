// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ampersand.
//
// The form is a figure of eight drawn in one stroke, with an arm.  A round
// loop stands at the top, left of the middle.  Its left side, ampLoopLeftW
// times AmpThick thick, runs on without a corner as the heavy diagonal, which
// thickens to AmpThick, crosses the sign from the upper left to the lower right
// and lifts at the lower right into a tail, which thins to AmpTail and ends in
// a cut at right angles to it, rising ampTailDir degrees, above the baseline at
// the right end of the ink.  The right side of the loop is thin; the thin
// stroke comes down from it to the left, crosses the diagonal and swings into
// the lower bowl, the widest part of the sign, with its weight, AmpBowlThick,
// at the extreme of its left side.  From the bottom of the bowl the stroke
// rises to the right, crosses the diagonal again at the lower crossing and runs
// on as the arm, which thickens from ampArmFootW times AmpArmThick to
// AmpArmThick, to the head serif of a diagonal of the capitals (diagonal.go),
// whose top lies at AmpArm, whose right tip lies ampSerifIn inside the right
// end of the ink and which reaches AmpSerifLeft and AmpSerifRight beyond the
// edges of the arm.  The bracket of the serif meets both edges of the arm at
// the mean of the runs of the obtuse and the acute side, with the reach of the
// acute side.
//
// The part from the serif round the bowl and the loop to the left side of the
// loop is one stroke of stations (symstroke.go): down the arm, across the
// diagonal, round the bottom of the bowl and up its left side, across the
// diagonal to the right side of the loop, over its top and down to its left
// side.  Each station has the direction of the stroke and its thickness; the
// edges between the stations are cubics with the curvatures which the
// neighbouring stations ask for, cut at their extremes.  From the left side of
// the loop to the cut the stroke follows the spine, a centre line given by its
// curvature along its length, with a thickness which changes smoothly along it;
// its edges are cubics between nodes of the spine, each with the point, the
// direction and the curvature of the edge there.  The places follow from a few
// rules:
//
//   - The guide line of the diagonal passes through the point (ampUpX,
//     ampUpY) at ampDiagAngle; the lower crossing is the point of the
//     guide line at ampLowY.
//   - The arm is straight under the serif over the part ampArmStraight of the
//     height from the lower crossing to the end of the bracket, with each edge
//     on the edge of a serif diagonal; below that its centre line bends evenly
//     more, from no curvature to the curvature at the lower crossing, which it
//     leaves ampArmBow degrees flatter than the straight part.  The thickness
//     of the arm grows evenly with the height.
//   - The stroke leaves the lower crossing and turns into the bottom of the
//     bowl, which it reaches at the overshoot, where a circular arc in the
//     direction of the arm at the lower crossing would; the extreme of the left
//     side lies on the quarter circle from the bottom, which gives the
//     curvature at the bottom.
//   - The right side of the loop lies where a circular arc from (ampUpX,
//     ampUpY), rising ampUpDir degrees, turns upright at ampLoopRightY; its
//     left side lies at ampLoopLeftY where a curve whose curvature falls evenly
//     to nothing turns from upright into the guide line.  The curvatures at
//     the left side, at the right side and at the top are those of quarters of
//     ellipses through the top and the sides, the one at the top the mean of
//     those of the two quarters.
//   - The curvature of the spine falls evenly from that at the left side of
//     the loop, at the rate of the turn into the guide line, to the
//     curvature of the diagonal, keeps that over ampSpinePlain and rises evenly
//     to the cut.  The curvature of the diagonal, the length of the spine and
//     the curvature at the cut are those with which the spine ends at the cut
//     in the direction ampTailDir.  The thickness of the spine is ampLoopLeftW
//     times AmpThick at the left side of the loop, ampDiagStartW times AmpThick
//     where the guide line meets the turn of the loop, AmpThick at the
//     height ampDiagEndY and AmpTail at the cut (ampWidth).
//
// The stroke crosses itself twice, and the outline is the union of it and of
// itself (symunion.go).  The places are fractions of the ink width AmpInk
// (abscissae) and of the height from AmpBottom to AmpTop (heights), and the
// directions are in degrees from the positive x axis.
//
// Side kinds: open, AmpSideLeft and AmpSideRight to the ink.
const (
	// The point on the guide line of the diagonal from which the right
	// side of the loop is placed, and the direction of the arc from it.
	ampUpX   = 0.35
	ampUpY   = 0.54
	ampUpDir = 20.0

	// The loop: the heights of the extremes of its right and its left side,
	// the thickness of its left side as a fraction of AmpThick, and the
	// abscissa of its top as a fraction of the way from its left to its right.
	ampLoopRightY = 0.80
	ampLoopLeftY  = 0.81
	ampLoopLeftW  = 0.8
	ampLoopTopAt  = 0.5

	// The guide line of the diagonal: its angle below the horizontal, the
	// thickness where it meets the turn of the loop as a fraction of AmpThick,
	// and the height where the thickness is AmpThick.
	ampDiagAngle  = 55.0
	ampDiagStartW = 0.9
	ampDiagEndY   = 0.15

	// The length of the spine over which it keeps the curvature of the
	// diagonal (units).
	ampSpinePlain = 300.0

	// The height of the lower crossing.
	ampLowY = 0.19

	// The arm: the part of its height which is straight, the turn of its
	// centre line below that (degrees), and its thickness at the lower crossing
	// as a fraction of AmpArmThick.
	ampArmStraight = 0.3
	ampArmBow      = 8.0
	ampArmFootW    = 0.88

	// The tail: the height of the middle of the cut above the baseline
	// (units), and the direction in which it ends.
	ampTailEndY = 34.0
	ampTailDir  = 22.0

	// How far the right tip of the serif of the arm lies inside the right end
	// of the ink.
	ampSerifIn = 20.0
)

// ampLoopTurn returns the length over which a curve whose curvature falls
// evenly from k to nothing turns from upright into the guide line of the
// diagonal.
func ampLoopTurn(k float64) float64 { return 2 * (90 - ampDiagAngle) * math.Pi / 180 / k }

// ampFrame holds the frame of the ampersand.
type ampFrame struct {
	x0, w, top, bottom, h float64
}

func newAmpFrame(p design.Params) ampFrame {
	s := p.Sym
	f := ampFrame{x0: s.AmpSideLeft, w: s.AmpInk, top: s.AmpTop, bottom: s.AmpBottom}
	f.h = f.top - f.bottom
	return f
}

// x and y return the abscissa at the fraction u of the ink width and the height
// at the fraction v of the height.
func (f ampFrame) x(u float64) float64 { return f.x0 + u*f.w }
func (f ampFrame) y(v float64) float64 { return f.bottom + v*f.h }

// diagX returns the abscissa of the guide line of the diagonal at the
// height y.
func (f ampFrame) diagX(y float64) float64 {
	return f.x(ampUpX) + (f.y(ampUpY)-y)/math.Tan(ampDiagAngle*math.Pi/180)
}

// ampArm returns the straight part of the arm as a diagonal, AmpArmThick thick,
// whose right edge lies AmpSerifRight left of the right tip of the serif at the
// underside of the serif, and the direction (degrees) of its centre line going
// up.  The chord from the point low to the point of the centre line at the
// height yS has the direction of the straight part less a third of ampArmBow,
// the chord of a curve whose curvature grows evenly from nothing.
func ampArm(p design.Params, f ampFrame, low outline.Point, yS float64) (diagonal, float64) {
	s := p.Sym
	tip := f.x0 + f.w - ampSerifIn
	under := s.AmpArm - p.Caps.SerifThick
	d := diagonal{Y: under, Thick: s.AmpArmThick}
	dir := math.Atan2(under-low.Y, tip-s.AmpSerifRight-low.X) * 180 / math.Pi
	for i := 0; i < 30; i++ {
		d.Lean = 1 / math.Tan(dir*math.Pi/180)
		d.X = tip - s.AmpSerifRight - d.width()/2
		chord := math.Atan2(yS-low.Y, d.centreAt(yS)-low.X) * 180 / math.Pi
		dir = chord + ampArmBow/3
	}
	d.Lean = 1 / math.Tan(dir*math.Pi/180)
	d.X = tip - s.AmpSerifRight - d.width()/2
	return d, dir
}

// ampTurnRadius returns the radius of the circular arc which turns from the
// direction dir (degrees, a descent) to the horizontal over the drop d.
func ampTurnRadius(dir, d float64) float64 {
	return d / (1 - math.Cos(dir*math.Pi/180))
}

// ampSpiral returns the displacement along a curve of the length l which
// leaves in the direction from (degrees) and turns through turn degrees with a
// curvature which falls evenly to nothing at its end; reversed, it rises from
// nothing.
func ampSpiral(from, turn, l float64) outline.Point {
	const n = 200
	k0 := 2 * turn * math.Pi / 180 / l
	var d outline.Point
	for i := 0; i < n; i++ {
		u := (float64(i) + 0.5) / n * l
		a := from*math.Pi/180 + k0*(u-u*u/(2*l))
		d.X += math.Cos(a) * l / n
		d.Y += math.Sin(a) * l / n
	}
	return d
}

// An ampSpine is the spine of the ampersand: a curve from p0, which it leaves
// straight down, whose curvature falls evenly from k0 to km at the rate
// k0/fall, stays at km over the length ampSpinePlain and rises evenly from
// there to k1 at its end, the length l.
type ampSpine struct {
	p0                  outline.Point
	k0, fall, km, k1, l float64
}

// knots returns the lengths and the curvatures between which the curvature of
// the spine changes evenly.
func (c ampSpine) knots() [4][2]float64 {
	s1 := c.fall * (1 - c.km/c.k0)
	return [4][2]float64{{0, c.k0}, {s1, c.km}, {s1 + ampSpinePlain, c.km}, {c.l, c.k1}}
}

// kappa returns the curvature at the length s.
func (c ampSpine) kappa(s float64) float64 {
	kn := c.knots()
	for i := 1; i < len(kn); i++ {
		if s <= kn[i][0] || i == len(kn)-1 {
			a, b := kn[i-1], kn[i]
			return a[1] + (b[1]-a[1])*(s-a[0])/(b[0]-a[0])
		}
	}
	return 0
}

// angle returns the direction (radians) at the length s.
func (c ampSpine) angle(s float64) float64 {
	kn := c.knots()
	a := -math.Pi / 2
	for i := 1; i < len(kn); i++ {
		u := math.Min(s, kn[i][0]) - kn[i-1][0]
		if u <= 0 {
			break
		}
		k0, k1 := kn[i-1][1], kn[i][1]
		a += k0*u + (k1-k0)*u*u/(2*(kn[i][0]-kn[i-1][0]))
	}
	return a
}

// at returns the point at the length s.
func (c ampSpine) at(s float64) outline.Point {
	n := int(math.Ceil(s/2)) + 1
	h := s / float64(n)
	q := c.p0
	for i := 0; i < n; i++ {
		a := c.angle((float64(i) + 0.5) * h)
		q.X += math.Cos(a) * h
		q.Y += math.Sin(a) * h
	}
	return q
}

// lengthAt returns the length at which the spine first comes down to the
// height y.
func (c ampSpine) lengthAt(y float64) float64 {
	lo, hi := 0.0, c.l
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2
		if c.at(mid).Y > y {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// newAmpSpine returns the spine from the left extreme of the loop p0, with the
// curvature k0 there, falling at the rate k0/fall, which ends at p1 in the
// direction d1 (degrees): km, l and k1 are solved for.  It panics if there is
// none.
func newAmpSpine(p0 outline.Point, k0, fall float64, p1 outline.Point, d1 float64) ampSpine {
	chord := math.Hypot(p1.X-p0.X, p1.Y-p0.Y)
	c := ampSpine{p0: p0, k0: k0, fall: fall, l: 1.1 * chord}
	c.k1 = 2 * (d1*math.Pi/180 - c.angle(c.knots()[2][0])) / (c.l - c.knots()[2][0])
	res := func(c ampSpine) [3]float64 {
		q := c.at(c.l)
		return [3]float64{(c.angle(c.l) - d1*math.Pi/180) * chord, q.X - p1.X, q.Y - p1.Y}
	}
	ptr := func(c *ampSpine, i int) *float64 { return [3]*float64{&c.km, &c.l, &c.k1}[i] }
	for it := 0; it < 60; it++ {
		r := res(c)
		if math.Abs(r[0])+math.Abs(r[1])+math.Abs(r[2]) < 1e-7 {
			return c
		}
		var j [3][3]float64
		for i := 0; i < 3; i++ {
			v := *ptr(&c, i)
			h := 1e-4 * math.Max(math.Abs(v), 1e-3)
			e := c
			*ptr(&e, i) = v + h
			ri := res(e)
			for k := 0; k < 3; k++ {
				j[k][i] = (ri[k] - r[k]) / h
			}
		}
		dx, ok := solve3(j, [3]float64{-r[0], -r[1], -r[2]})
		if !ok {
			break
		}
		// The step is shortened until the knots stay in order.
		for damp := 1.0; damp > 1e-3; damp /= 2 {
			e := c
			for i := 0; i < 3; i++ {
				*ptr(&e, i) += damp * dx[i]
			}
			if kn := e.knots(); e.km < e.k0 && kn[2][0] < e.l {
				c = e
				break
			}
		}
	}
	panic("glyphs: glyph ampersand: no spine reaches the cut of the tail; glyphs.Check reports such parameters")
}

// kslope returns the rate of change of the curvature at the length s.
func (c ampSpine) kslope(s float64) float64 {
	kn := c.knots()
	for i := 1; i < len(kn); i++ {
		if s < kn[i][0] || i == len(kn)-1 {
			return (kn[i][1] - kn[i-1][1]) / (kn[i][0] - kn[i-1][0])
		}
	}
	return 0
}

// ampWidth is the thickness along the spine: between the knots, pairs of a
// length and a thickness, the quintic with the values and the slopes of the
// knots and no second derivative there.  The slope is nothing at every knot
// but the last, where it is that of the last chord.
type ampWidth struct {
	s, w, d []float64
}

// newAmpWidth returns the thickness through the knots.
func newAmpWidth(knots [][2]float64) ampWidth {
	n := len(knots)
	w := ampWidth{s: make([]float64, n), w: make([]float64, n), d: make([]float64, n)}
	for i, k := range knots {
		w.s[i], w.w[i] = k[0], k[1]
	}
	w.d[n-1] = (w.w[n-1] - w.w[n-2]) / (w.s[n-1] - w.s[n-2])
	return w
}

// at returns the thickness and its first and second derivatives at the length
// s.
func (w ampWidth) at(s float64) (v, d1, d2 float64) {
	i := 0
	for i < len(w.s)-2 && s > w.s[i+1] {
		i++
	}
	h := w.s[i+1] - w.s[i]
	t := (s - w.s[i]) / h
	t2, t3, t4, t5 := t*t, t*t*t, t*t*t*t, t*t*t*t*t
	y0, y1, a, b := w.w[i], w.w[i+1], h*w.d[i], h*w.d[i+1]
	v = y0*(1-10*t3+15*t4-6*t5) + a*(t-6*t3+8*t4-3*t5) + b*(-4*t3+7*t4-3*t5) + y1*(10*t3-15*t4+6*t5)
	d1 = (y0*(-30*t2+60*t3-30*t4) + a*(1-18*t2+32*t3-15*t4) + b*(-12*t2+28*t3-15*t4) + y1*(30*t2-60*t3+30*t4)) / h
	d2 = (y0*(-60*t+180*t2-120*t3) + a*(-36*t+96*t2-60*t3) + b*(-24*t+84*t2-60*t3) + y1*(60*t-180*t2+120*t3)) / (h * h)
	return v, d1, d2
}

// edge returns the edge of the stroke on the spine with the thickness w on
// the given side (-1 for the right-hand edge, 1 for the left-hand one), from
// the start to the end: cubics between the nodes at the lengths us, each with
// the point, the direction and the curvature of the edge at its ends.
func (c ampSpine) edge(w ampWidth, side float64, us []float64) chain {
	type node struct {
		p, t outline.Point
		k    float64
	}
	at := func(s float64) node {
		q, a := c.at(s), c.angle(s)
		k, dk := c.kappa(s), c.kslope(s)
		wv, w1, w2 := w.at(s)
		t, n := dirOf(a), dirOf(a+math.Pi/2)
		// The edge is q + side w/2 n; its derivative is A t + B n.
		aa, bb := 1-side*k*wv/2, side*w1/2
		da, db := -side*(dk*wv+k*w1)/2, side*w2/2
		cr := aa*(aa*k+db) - bb*(da-bb*k)
		norm := math.Hypot(aa, bb)
		return node{
			p: outline.Point{X: q.X + side*wv/2*n.X, Y: q.Y + side*wv/2*n.Y},
			t: outline.Point{X: (aa*t.X + bb*n.X) / norm, Y: (aa*t.Y + bb*n.Y) / norm},
			k: cr / (norm * norm * norm),
		}
	}
	var ch chain
	prev := at(us[0])
	for _, u := range us[1:] {
		next := at(u)
		cu, err := g2Cubic(prev.p, prev.t, prev.k, next.p, next.t, next.k)
		if err != nil {
			cu = nearG2(prev.p, prev.t, prev.k, next.p, next.t, next.k)
		}
		ch = append(ch, splitAtExtrema(cu)...)
		prev = next
	}
	return ch
}

// solve3 solves the linear system a x = b by Cramer's rule.
func solve3(a [3][3]float64, b [3]float64) ([3]float64, bool) {
	det := func(m [3][3]float64) float64 {
		return m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
			m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
			m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
	}
	d := det(a)
	var x [3]float64
	if d == 0 || math.IsNaN(d) {
		return x, false
	}
	for i := 0; i < 3; i++ {
		m := a
		for k := 0; k < 3; k++ {
			m[k][i] = b[k]
		}
		x[i] = det(m) / d
	}
	return x, true
}

func ampersand(p design.Params) (outline.Path, float64) {
	s := p.Sym
	f := newAmpFrame(p)
	t := p.Caps.RoundThin
	st := func(q outline.Point, dir, w float64) station { return station{X: q.X, Y: q.Y, Dir: dir, W: w} }
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }

	up := pt(f.x(ampUpX), f.y(ampUpY))
	low := pt(f.diagX(f.y(ampLowY)), f.y(ampLowY))

	// The arm.
	br := design.CapDiagBracket(p)
	br.Obtuse = (br.Obtuse + br.Acute) / 2
	br.Acute, br.Reach = br.Obtuse, br.ReachAcute
	brY := s.AmpArm - p.Caps.SerifThick - br.Obtuse
	yS := brY - ampArmStraight*(brY-low.Y)
	arm, topDir := ampArm(p, f, low, yS)
	lowDir := topDir - ampArmBow
	armTurn := math.Hypot(arm.centreAt(yS)-low.X, yS-low.Y)
	kLow := -2 * ampArmBow * math.Pi / 180 / armTurn
	// The thickness of the arm grows evenly with the height from the lower
	// crossing to the underside of the serif; the edges of the straight part
	// are the edges of a serif diagonal each.
	under := s.AmpArm - p.Caps.SerifThick
	wLow := ampArmFootW * s.AmpArmThick
	wArm := func(y float64) float64 { return wLow + (s.AmpArmThick-wLow)*(y-low.Y)/(under-low.Y) }
	hw := math.Hypot(1, arm.Lean)
	dw := (s.AmpArmThick - wLow) / (under - low.Y) / 2 * hw
	armR := diagFromEdge(edgeRight, arm.centreAt(under)+s.AmpArmThick/2*hw, under, arm.Lean+dw, s.AmpArmThick)
	armL := diagFromEdge(edgeLeft, arm.centreAt(under)-s.AmpArmThick/2*hw, under, arm.Lean-dw, s.AmpArmThick)

	// The bowl.
	botY := f.bottom + t/2
	rb := ampTurnRadius(lowDir, low.Y-botY)
	bot := pt(low.X-rb*math.Sin(lowDir*math.Pi/180), botY)
	leftX := f.x0 + s.AmpBowlThick/2
	side := pt(leftX, botY+(bot.X-leftX))
	kBowl := -1 / (bot.X - leftX)

	// The loop.
	ry := f.y(ampLoopRightY)
	rx := up.X + math.Tan((90-ampUpDir)/2*math.Pi/180)*(ry-up.Y)
	ly := f.y(ampLoopLeftY)
	lw := ampLoopLeftW * s.AmpThick
	// The curvature at the left extreme is that of the quarter of an ellipse
	// from the top, and the turn into the guide line falls evenly from it.
	var qy, lx, kLeft, kTop, kRight float64
	var top outline.Point
	kLeft = 0.008
	for i := 0; i < 30; i++ {
		dl := ampSpiral(-90, 90-ampDiagAngle, ampLoopTurn(kLeft))
		qy = ly + dl.Y
		lx = f.diagX(qy) - dl.X
		top = pt(lx-lw/2+ampLoopTopAt*(rx+t/2-(lx-lw/2)), f.top-t/2)
		a, b := top.X-lx, top.Y-ly
		c, d := rx-top.X, top.Y-ry
		kLeft, kTop, kRight = a/(b*b), (b/(a*a)+d/(c*c))/2, c/(d*d)
	}

	// The spine, from the left side of the loop to the cut.
	ta := ampTailDir * math.Pi / 180
	end := pt(f.x0+f.w-s.AmpTail/2*math.Sin(ta), ampTailEndY)
	sp := newAmpSpine(pt(lx, ly), kLeft, ampLoopTurn(kLeft), end, ampTailDir)
	wd := newAmpWidth([][2]float64{
		{0, lw},
		{sp.lengthAt(qy), ampDiagStartW * s.AmpThick},
		{sp.lengthAt(f.y(ampDiagEndY)), s.AmpThick},
		{sp.l, s.AmpTail},
	})
	kn := sp.knots()
	s1, s2 := kn[1][0], kn[2][0]
	spineR := sp.edge(wd, -1, []float64{0, (s1 + s2) / 2, s2, (s2 + sp.l) / 2, sp.l})
	spineL := sp.edge(wd, 1, []float64{0, 0.6 * s1, (s1 + s2) / 2, s2, (s2 + sp.l) / 2, sp.l}).reversed()

	// The stroke starts where the brackets of the serif end, with a
	// horizontal nib, and is straight down to yS.
	sts := []station{
		{X: arm.centreAt(brY), Y: brY, Dir: topDir + 180, W: wArm(brY) * hw, Nib: 360, Flat: true},
		st(pt(arm.centreAt(yS), yS), topDir+180, wArm(yS)),
		{X: low.X, Y: low.Y, Dir: lowDir + 180, W: wLow, K: kLow},
		{X: bot.X, Y: bot.Y, Dir: 180, W: t, K: kBowl},
		{X: side.X, Y: side.Y, Dir: 90, W: s.AmpBowlThick},
		{X: rx, Y: ry, Dir: 90, W: t, K: kRight},
		{X: top.X, Y: top.Y, Dir: 180, W: t, K: kTop},
		{X: lx, Y: ly, Dir: -90, W: lw, K: kLeft},
	}
	right, left := strokeOpen(sts)
	right, left = join(right, spineR), join(spineL, left)
	outR := armR.headServif(edgeRight, s.AmpArm, diagSerif{Over: s.AmpSerifRight, Thick: p.Caps.SerifThick, Bracket: br}).reversed()
	outL := armL.headServif(edgeLeft, s.AmpArm, diagSerif{Over: s.AmpSerifLeft, Thick: p.Caps.SerifThick, Bracket: br})
	// The top of the serif is one straight piece from tip to tip.
	outR, outL = outR[:len(outR)-1], outL[1:]
	// The edges of the stroke begin where the brackets end.
	stroke := join(right, lineChain(right.end(), left.start()), left, outR,
		lineChain(outR.end(), outL.start()), outL)
	path := chainsPath(union(stroke)...)
	_, _, x1, _ := path.BBox()
	return path, math.Round(x1 + s.AmpSideRight)
}
