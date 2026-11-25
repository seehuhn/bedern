// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// An sShape holds the numbers of S: the ear of the upper terminal from the
// right edge, the position of the terminal on the ellipse of the bowl in
// degrees from the right, the height of the ellipse over the height which
// makes its tangent at dsSSpine pass through the middle, the shape p of the
// fall of the curvature of the spine (sSpine), and how much of what a doubled
// stem adds to a stem the letter is wider than its roman.
type sShape struct {
	ear, terminal, tall, profile, widen float64
}

// dsSSpine is the direction in degrees of the line from the ellipse of the
// bowl through the middle of the letter which sets the height of the
// ellipse.
const dsSSpine = 325.0

// dsSNodes is the number of nodes of the stroke tool on the centre line of
// the spine from the join on the bowl to the middle, the middle not counted.
const dsSNodes = 3

// dsSShapes are the shapes of S which dsS tries in turn: the first is the
// shape which the fairness analysis accepts on the regular parameters and on
// almost every set at every level, the others make the bowls a little lower
// or higher, let the curvature of the spine fall faster or slower, and move
// the ear and the terminal a little, for the parameter sets and the levels on
// which the rounding to whole units leaves it a finding.
var dsSShapes = func() []sShape {
	var out []sShape
	for _, f := range [][2]float64{{0.95, -0.5}, {0.95, -0.7}, {0.9, -0.5}, {1, -0.5}, {0.95, -0.6}, {0.9, -0.7}, {1, -0.7}, {0.85, -0.5}} {
		for _, e := range [][2]float64{{0, 0}, {0, -2}, {0, 2}, {-5, 0}, {5, 0}, {0, 4}, {0, -4}} {
			out = append(out, sShape{55 + e[0], 14 + e[1], f[0], f[1], 0.4})
		}
	}
	return out
}()

// dsS draws 𝕊: one stroke from the upper terminal over the top and down the
// left of the upper bowl, along the spine and round the lower bowl to the
// lower terminal, the lower half the upper turned through half a turn about
// the middle of the spine.  The stroke is one line thick in the bowls and
// grows to a doubled stroke in the middle of the spine, and its white is the
// part of the spine more than a line from both edges, a lens.
func dsS(d dsFrame) outline.Path {
	return fairestWith("S", len(dsSShapes), func(i int) outline.Path { return d.settled(dsSOf(d, dsSShapes[i])) }, sSpineInflection)
}

// sSpineInflection reports whether the finding is an inflection in the spine
// of 𝕊: an inflection between a quarter and three quarters of the height of
// the letter, where the edges of the spine and of its white run from the curve
// of one bowl into that of the other.  The fairness baseline of the regular
// judges the inflection in the spine of S intended, and 𝕊 has one on each
// edge of its spine and of its white for the same reason.
func sSpineInflection(path outline.Path, f fair.Finding) bool {
	_, y0, _, y1 := path.BBox()
	return f.Kind == fair.Inflection && f.Y > y0+0.25*(y1-y0) && f.Y < y0+0.75*(y1-y0)
}

// dsSOf draws S with the shape sh.
//
// The centre line of the upper bowl is an ellipse with the half axis a from
// the extent of the letter and the other half axis b, tall times the one which
// makes its tangent at dsSSpine pass through the middle.  At the join the
// centre line leaves the ellipse with its direction and curvature and runs to
// the middle as a curve whose curvature falls to nought there (sSpine); the
// join and the direction in the middle are those for which the curve reaches
// the middle, found by Newton's method from the best point of a grid.  The
// stroke grows over this curve from a line at the join to a doubled stroke in
// the middle as t + (d - t)(3x² - 2x³) of the share x of its length; it is
// widest in the middle and its edges bend outwards there.  The lower
// half is the upper turned through half a turn about the middle.  The edges
// are those of the stroke tool, whose nodes carry the curvature of the centre
// line and the first two derivatives of the thickness; the curvature of the
// edges runs on without a jump: each edge of the spine bends from the
// curve of the one bowl into that of the other through one inflection, the
// outer edge after the middle and the inner edge before it.
//
// The white is the part of the spine more than a line from both edges: the
// outer edge moved a line inwards from where it crosses the inner edge moved
// a line inwards, at the points of the lens, and the same turned through half
// a turn.
func dsSOf(d dsFrame, sh sShape) outline.Path {
	cap := d.p.CapHeight
	x0, _, x1, y1 := d.roman.BBox()
	x0, x1 = d.x(x0), d.x(x1)
	x1 += sh.widen * d.extra
	cx, cy := (x0+x1)/2, cap/2
	t := d.t
	top := y1 - t/2
	xb := x0 + t/2
	a := (x1 - sh.ear - xb) / 2
	ux := xb + a
	// The point of the ellipse about (ux, top - b) at the parameter phi, in
	// degrees, the direction of travel there and the curvature.
	at := func(b, phi float64) (x, y, dir, k float64) {
		ph := phi * math.Pi / 180
		s, c := math.Sin(ph), math.Cos(ph)
		x, y = ux+a*c, top-b+b*s
		dir = math.Atan2(b*c, -a*s) * 180 / math.Pi
		k = a * b / math.Pow(a*a*s*s+b*b*c*c, 1.5)
		return x, y, math.Mod(dir+360, 360), k
	}
	// phiSpine returns the parameter at which the direction is dsSSpine, by
	// bisection between 180 and 270 degrees.
	phiSpine := func(b float64) float64 {
		lo, hi := 180.0, 270.0
		for range 60 {
			m := (lo + hi) / 2
			if _, _, dir, _ := at(b, m); dir > dsSSpine {
				hi = m
			} else {
				lo = m
			}
		}
		return (lo + hi) / 2
	}
	// miss is how far the line in the direction dsSSpine through its point on
	// the ellipse passes above the middle.
	th := dsSSpine * math.Pi / 180
	miss := func(b float64) float64 {
		x, y, _, _ := at(b, phiSpine(b))
		return y + (cx-x)*math.Tan(th) - cy
	}
	lo, hi := 40.0, 320.0
	for range 80 {
		m := (lo + hi) / 2
		if miss(m) > 0 {
			lo = m
		} else {
			hi = m
		}
	}
	b := sh.tall * (lo + hi) / 2
	spineOf := func(phi, dir float64) sSpine {
		jx, jy, jdir, jk := at(b, phi)
		turn := math.Remainder((dir-jdir)*math.Pi/180, 2*math.Pi)
		return sSpine{
			p0: pt(jx, jy), a0: jdir * math.Pi / 180, k0: jk, p: sh.profile,
			len: turn / (jk * (0.5 + sh.profile/6)),
		}
	}
	missAt := func(phi, dir float64) (float64, float64) {
		q := spineOf(phi, dir).at(1, 32)
		return q.X - cx, q.Y - cy
	}
	phi, dir := 0.0, 0.0
	best := math.Inf(1)
	for ph := 120.0; ph <= 240; ph += 4 {
		for th := 315.0; th <= 350; th += 2.5 {
			if fx, fy := missAt(ph, th); math.Hypot(fx, fy) < best {
				best, phi, dir = math.Hypot(fx, fy), ph, th
			}
		}
	}
	for range 40 {
		fx, fy := missAt(phi, dir)
		if math.Hypot(fx, fy) < 1e-6 {
			break
		}
		const h = 1e-4
		ax, ay := missAt(phi+h, dir)
		bx, by := missAt(phi, dir+h)
		j11, j21 := (ax-fx)/h, (ay-fy)/h
		j12, j22 := (bx-fx)/h, (by-fy)/h
		det := j11*j22 - j12*j21
		if math.Abs(det) < 1e-15 {
			panic("mathfont: the spine of S has no centre line")
		}
		dp := (fx*j22 - fy*j12) / det
		dd := (j11*fy - j21*fx) / det
		// Damped steps keep the search near the point of the grid.
		if m := math.Max(math.Abs(dp)/4, math.Abs(dd)/2); m > 1 {
			dp, dd = dp/m, dd/m
		}
		phi, dir = phi-dp, dir-dd
	}
	sp := spineOf(phi, dir)
	if fx, fy := missAt(phi, dir); math.Hypot(fx, fy) > 1e-3 || phi < 120 || phi > 240 || dir < 310 || dir > 355 || sp.len <= 0 {
		panic("mathfont: the spine of S has no centre line")
	}
	dw, l := d.d-t, sp.len
	spineNode := func(x float64) stroke.Node {
		return stroke.Node{
			P:   sp.at(x, 64),
			Dir: sp.angle(x) * 180 / math.Pi,
			K:   sp.k(x),
			W:   t + dw*(3*x*x-2*x*x*x),
			DW:  dw * (6*x - 6*x*x) / l,
			DDW: dw * (6 - 12*x) / (l * l),
		}
	}
	var upper []stroke.Node
	for _, deg := range []float64{sh.terminal, 90, 180} {
		if deg > phi-15 {
			continue
		}
		x, y, dir, k := at(b, deg)
		upper = append(upper, stroke.Node{P: pt(x, y), Dir: dir, K: k, W: t})
	}
	for i := range dsSNodes {
		upper = append(upper, spineNode(float64(i)/dsSNodes))
	}
	nodes := append(upper, spineNode(1))
	turn := func(q outline.Point) outline.Point { return pt(2*cx-q.X, 2*cy-q.Y) }
	for i := len(upper) - 1; i >= 0; i-- {
		n := upper[i]
		n.P, n.K, n.DW = turn(n.P), -n.K, -n.DW
		nodes = append(nodes, n)
	}
	r, err := stroke.Stroke{Nodes: nodes, Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut}}.Build()
	if err != nil {
		panic(err)
	}
	// The inner edge is taken as the outer edge turned and run backwards, so
	// that the lens is turned into itself.
	outer := chainCubics(r.Right)
	var inner []stroke.Cubic
	for i := len(outer) - 1; i >= 0; i-- {
		c := outer[i]
		inner = append(inner, stroke.Cubic{turn(c[3]), turn(c[2]), turn(c[1]), turn(c[0])})
	}
	cs := offsetCrossings(outer, t, inner, -t)
	if len(cs) != 2 {
		panic(fmt.Sprintf("mathfont: the edges of the white of S cross %d times", len(cs)))
	}
	// The lens is the moved outer edge between its points and the same
	// turned; the end of the first half is put on the turned start.
	wa := offsetRun(outer, cs[0].a, cs[1].a, t, 40)
	last := &wa[len(wa)-1]
	end := turn(wa[0][0])
	last[2] = pt(last[2].X+end.X-last[3].X, last[2].Y+end.Y-last[3].Y)
	last[3] = end
	white := outline.Contour{Start: wa[0][0]}
	for _, half := range []bool{false, true} {
		for _, c := range wa {
			if half {
				c = stroke.Cubic{turn(c[0]), turn(c[1]), turn(c[2]), turn(c[3])}
			}
			white.Segs = append(white.Segs, outline.Segment{IsCurve: true, P1: c[1], P2: c[2], P3: c[3]})
		}
	}
	white.Segs[len(white.Segs)-1].P3 = white.Start
	return append(r.Path(), reversed(white))
}

// An sSpine is the centre line of the spine of S from the bowl to the middle:
// a curve from p0 in the direction a0, in radians, whose curvature at the
// share x of its length len is k0 (1 - x)(1 + p x), which falls to nought at
// its end, faster at first for a negative p.
type sSpine struct {
	p0             outline.Point
	a0, k0, p, len float64
}

func (e sSpine) k(x float64) float64 { return e.k0 * (1 - x) * (1 + e.p*x) }

// angle returns the direction of the curve at the share x of its length, in
// radians.
func (e sSpine) angle(x float64) float64 {
	x2, x3 := x*x, x*x*x
	return e.a0 + e.len*e.k0*(x+(e.p-1)*x2/2-e.p*x3/3)
}

// at returns the point of the curve at the share x of its length, in n steps
// of Simpson's rule.
func (e sSpine) at(x float64, n int) outline.Point {
	q := e.p0
	h := x / float64(n)
	for i := range n {
		s := float64(i) * h
		a, m, b := e.angle(s), e.angle(s+h/2), e.angle(s+h)
		q.X += e.len * h / 6 * (math.Cos(a) + 4*math.Cos(m) + math.Cos(b))
		q.Y += e.len * h / 6 * (math.Sin(a) + 4*math.Sin(m) + math.Sin(b))
	}
	return q
}

// An edgePlace is a place on a chain of cubics: the index of the cubic and the
// parameter on it.
type edgePlace struct {
	i int
	s float64
}

// An edgeCrossing is where two moved edges cross, as the place on each.
type edgeCrossing struct{ a, b edgePlace }

// offsetCrossings returns where the edge a moved by ha to its left crosses
// the edge b moved by hb, in the order of a.
func offsetCrossings(a []stroke.Cubic, ha float64, b []stroke.Cubic, hb float64) []edgeCrossing {
	const n = 24
	type sample struct {
		p edgePlace
		q outline.Point
	}
	samples := func(c stroke.Cubic, i int, h float64) []sample {
		out := make([]sample, n+1)
		for j := range n + 1 {
			s := float64(j) / n
			out[j] = sample{edgePlace{i, s}, offsetAt(c, s, h).p}
		}
		return out
	}
	box := func(ss []sample) [4]float64 {
		r := [4]float64{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
		for _, s := range ss {
			r[0], r[1] = math.Min(r[0], s.q.X), math.Min(r[1], s.q.Y)
			r[2], r[3] = math.Max(r[2], s.q.X), math.Max(r[3], s.q.Y)
		}
		return r
	}
	var sb [][]sample
	var bb [][4]float64
	for j, c := range b {
		s := samples(c, j, hb)
		sb, bb = append(sb, s), append(bb, box(s))
	}
	var out []edgeCrossing
	for i, c := range a {
		sa := samples(c, i, ha)
		ba := box(sa)
		for j := range b {
			if ba[2] < bb[j][0] || bb[j][2] < ba[0] || ba[3] < bb[j][1] || bb[j][3] < ba[1] {
				continue
			}
			for k := range n {
				a0, a1 := sa[k].q, sa[k+1].q
				for m := range n {
					b0, b1 := sb[j][m].q, sb[j][m+1].q
					dx1, dy1 := a1.X-a0.X, a1.Y-a0.Y
					dx2, dy2 := b1.X-b0.X, b1.Y-b0.Y
					den := dx1*dy2 - dy1*dx2
					if math.Abs(den) < 1e-12 {
						continue
					}
					r := ((b0.X-a0.X)*dy2 - (b0.Y-a0.Y)*dx2) / den
					q := ((b0.X-a0.X)*dy1 - (b0.Y-a0.Y)*dx1) / den
					if r < 0 || r >= 1 || q < 0 || q >= 1 {
						continue
					}
					out = append(out, edgeCrossing{
						edgePlace{i, sa[k].p.s + r/n},
						edgePlace{j, sb[j][m].p.s + q/n},
					})
				}
			}
		}
	}
	return out
}

// offsetRun returns the edge cs from the place p0 to the place p1 moved by h
// to its left, as cubics which meet with the direction and the curvature of
// the moved edge: one from each joint of the edge to the next, where a joint
// closer than minRun to the one before along the edge is left out.
func offsetRun(cs []stroke.Cubic, p0, p1 edgePlace, h, minRun float64) []stroke.Cubic {
	places := []edgePlace{p0}
	run := cs[p0.i].Sub(p0.s, 1).Length()
	for i := p0.i + 1; i <= p1.i; i++ {
		if run >= minRun {
			places = append(places, edgePlace{i, 0})
			run = 0
		}
		if i < p1.i {
			run += cs[i].Length()
		} else {
			run += cs[i].Sub(0, p1.s).Length()
		}
	}
	if run < minRun && len(places) > 1 {
		places = places[:len(places)-1]
	}
	places = append(places, p1)
	var out []stroke.Cubic
	a := offsetAt(cs[p0.i], p0.s, h)
	for _, q := range places[1:] {
		b := offsetAt(cs[q.i], q.s, h)
		out = append(out, hermite2(a.p, a.u, a.k, b.p, b.u, b.k))
		a = b
	}
	return out
}

// An offsetPoint is a point of an edge moved at right angles to it, with the
// direction and the curvature of the moved edge there.
type offsetPoint struct {
	p, u outline.Point
	k    float64
}

// offsetAt returns the point of the cubic c at s moved by h to its left.
func offsetAt(c stroke.Cubic, s, h float64) offsetPoint {
	q, u, k := c.At(s), c.Tangent(s), c.Curvature(s)
	return offsetPoint{pt(q.X-h*u.Y, q.Y+h*u.X), u, k / (1 - h*k)}
}

// hermite2 returns the cubic from p0 to p3 which leaves p0 in the unit
// direction u0 with the curvature k0 and reaches p3 in the direction u3 with
// the curvature k3, its handle lengths found by Newton's method from a third
// of the chord; where there is no such cubic near that start, the handles
// stay a third of the chord.
func hermite2(p0, u0 outline.Point, k0 float64, p3, u3 outline.Point, k3 float64) stroke.Cubic {
	dx, dy := p3.X-p0.X, p3.Y-p0.Y
	chord := math.Hypot(dx, dy)
	cr := u0.X*u3.Y - u0.Y*u3.X
	c0 := u0.X*dy - u0.Y*dx
	c3 := dx*u3.Y - dy*u3.X
	al, be := chord/3, chord/3
	for range 30 {
		f1 := k0*al*al - 2.0/3*(c0-be*cr)
		f2 := k3*be*be - 2.0/3*(c3-al*cr)
		j11, j12 := 2*k0*al, 2.0/3*cr
		j21, j22 := 2.0/3*cr, 2*k3*be
		det := j11*j22 - j12*j21
		if math.Abs(det) < 1e-12 {
			break
		}
		al -= (f1*j22 - f2*j12) / det
		be -= (j11*f2 - j21*f1) / det
	}
	if !(al > 0.05*chord && al < 0.9*chord && be > 0.05*chord && be < 0.9*chord) {
		al, be = chord/3, chord/3
	}
	return stroke.Cubic{p0, pt(p0.X+al*u0.X, p0.Y+al*u0.Y), pt(p3.X-be*u3.X, p3.Y-be*u3.Y), p3}
}

// chainCubics returns the cubics of an edge of a stroke.
func chainCubics(ch stroke.Chain) []stroke.Cubic {
	var out []stroke.Cubic
	for _, pc := range ch {
		out = append(out, pc.C)
	}
	return out
}

// pointed returns the contour with every straight segment shorter than max
// taken out, the ends which it joined meeting in its middle: the white of S
// begins and ends at a point and not in a cut.
func pointed(c outline.Contour, max float64) outline.Contour {
	pts := []outline.Point{c.Start}
	for _, sg := range c.Segs {
		pts = append(pts, sg.P3)
	}
	out := outline.Contour{Start: c.Start}
	for i, sg := range c.Segs {
		if !sg.IsCurve && math.Hypot(sg.P3.X-pts[i].X, sg.P3.Y-pts[i].Y) < max {
			m := pt((sg.P3.X+pts[i].X)/2, (sg.P3.Y+pts[i].Y)/2)
			if n := len(out.Segs); n > 0 {
				out.Segs[n-1].P3 = m
			} else {
				out.Start = m
			}
			pts[i+1] = m
			continue
		}
		out.Segs = append(out.Segs, sg)
	}
	// The closing segment returns to the start, which a cut may have moved.
	if n := len(out.Segs); n > 0 && !c.Segs[len(c.Segs)-1].IsCurve && pts[len(pts)-1] != c.Start {
		out.Segs[n-1].P3 = out.Start
	}
	return out
}
