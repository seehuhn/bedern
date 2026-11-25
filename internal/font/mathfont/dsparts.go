// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The parts from which the double-struck capitals with diagonals, hooks and
// two bowls are drawn (doublestruck.go, dscaps.go): straight edges as
// functions of the height, the serifs of a leaning stroke, convex polygons cut
// by lines for the white of a doubled stroke, and the crossing of two open
// contours.

// An xline is the straight edge x = x0 + s y, which leans by s units to the
// right for each unit of height.
type xline struct{ x0, s float64 }

func (e xline) at(y float64) float64 { return e.x0 + e.s*y }

// shift returns the line moved by dx to the right.
func (e xline) shift(dx float64) xline { return xline{e.x0 + dx, e.s} }

// lineThrough returns the line through two points of different height.
func lineThrough(a, b outline.Point) xline {
	s := (b.X - a.X) / (b.Y - a.Y)
	return xline{a.X - s*a.Y, s}
}

// crossY returns the height at which two lines which do not run parallel meet.
func crossY(e, f xline) float64 { return (f.x0 - e.x0) / (e.s - f.s) }

// across returns the width along a level line of a stroke of the width t
// across it, leaning by s.
func across(t, s float64) float64 { return t * math.Sqrt(1+s*s) }

// edgeOf returns the edge of the roman, widened, at the side 0 (left) or 1
// (right) of the run which pick chooses, through its points at the heights y0
// and y1.
func (d dsFrame) edgeOf(y0, y1 float64, pick func([][2]float64) [2]float64, side int) xline {
	a := d.x(d.romanRun(y0, pick)[side])
	b := d.x(d.romanRun(y1, pick)[side])
	return lineThrough(pt(a, y0), pt(b, y1))
}

// centreOf returns the centre line of the run of the roman at the index k
// through its points at the heights y0 and y1, widened.
func (d dsFrame) centreOf(y0, y1 float64, k0, k1 int) xline {
	r0, r1 := runsOf(d.roman, y0), runsOf(d.roman, y1)
	a := d.x((r0[k0][0] + r0[k0][1]) / 2)
	b := d.x((r1[k1][0] + r1[k1][1]) / 2)
	return lineThrough(pt(a, y0), pt(b, y1))
}

// stretchWith returns the ends of the longest stretch of heights between lo and
// hi (fractions of the cap height) at which the roman has n runs of ink, so
// that a stroke can be read where the strokes of the letter are apart.
func (d dsFrame) stretchWith(n int, lo, hi float64) (start, end float64) {
	cap := d.p.CapHeight
	best := -1.0
	runStart := math.NaN()
	for y := lo * cap; y <= hi*cap+1e-9; y += 4 {
		if len(runsOf(d.roman, y)) == n {
			if math.IsNaN(runStart) {
				runStart = y
			}
			if l := y - runStart; l > best {
				best, start, end = l, runStart, y
			}
		} else {
			runStart = math.NaN()
		}
	}
	if best < 0 {
		panic(fmt.Sprintf("mathfont: the roman has no height with %d runs between %.0f and %.0f", n, lo*cap, hi*cap))
	}
	return start, end
}

// heightsWith returns two heights, a tenth in from the ends of the longest
// stretch at which the roman has n runs of ink, and heightWith the middle of
// that stretch.
func (d dsFrame) heightsWith(n int, lo, hi float64) (y0, y1 float64) {
	a, b := d.stretchWith(n, lo, hi)
	return a + 0.1*(b-a), b - 0.1*(b-a)
}

func (d dsFrame) heightWith(n int, lo, hi float64) float64 {
	a, b := d.stretchWith(n, lo, hi)
	return (a + b) / 2
}

// cornerNear returns the corner of the roman outline nearest to q, before
// widening.
func (d dsFrame) cornerNear(q outline.Point) outline.Point {
	best, bd := q, math.Inf(1)
	for _, c := range d.roman {
		pts := []outline.Point{c.Start}
		for _, sg := range c.Segs {
			pts = append(pts, sg.P3)
		}
		for _, v := range pts {
			if dist := math.Hypot(d.x(v.X)-q.X, v.Y-q.Y); dist < bd {
				best, bd = v, dist
			}
		}
	}
	return best
}

// romanMaxX returns the rightmost ink of the roman between the heights y0 and
// y1.
func (d dsFrame) romanMaxX(y0, y1 float64) float64 {
	m := math.Inf(-1)
	for y := y0; y <= y1; y += 2 {
		if r := runsOf(d.roman, y); len(r) > 0 {
			m = math.Max(m, r[len(r)-1][1])
		}
	}
	return m
}

// The serifs of a leaning stroke whose edge e is the line of the edge: the
// serifs of a capital stem, level at the end, whose bracket meets the edge at
// the height of the bracket of the stem, where the bracket leaves the edge
// upright and so makes a corner with a leaning edge.  jut is the reach of
// the serif beyond that point.
func (d dsFrame) diagFootLeft(c *pen, e xline, jut float64) {
	c.run(glyphs.CapFootLeft(d.q, e.at(d.q.Caps.BracketHeight), jut))
}

func (d dsFrame) diagFootRight(c *pen, e xline, jut float64) {
	c.run(glyphs.CapFootRight(d.q, e.at(d.q.Caps.BracketHeight), jut))
}

func (d dsFrame) diagHeadLeft(c *pen, e xline, top, jut float64) {
	c.run(glyphs.CapHeadLeft(d.q, e.at(top-d.q.Caps.BracketHeight), top, jut))
}

func (d dsFrame) diagHeadRight(c *pen, e xline, top, jut float64) {
	c.run(glyphs.CapHeadRight(d.q, e.at(top-d.q.Caps.BracketHeight), top, jut))
}

// diagJut is the reach of the serifs of a leaning stroke: 0.78 of that of a
// stem, but at least the reach of the bracket and 16 units of flat.
func (d dsFrame) diagJut() float64 {
	return math.Max(0.78*d.p.Caps.SerifJut, d.p.Caps.BracketReach+16)
}

// serifs says which serifs a straight stroke has, as the reach of each: at
// the foot on the left and on the right, at the head on the left and on the
// right (zero for none).
type serifs struct{ footL, footR, headL, headR float64 }

// rod returns the outline of a straight stroke between the left edge l and the
// right edge r, from the baseline to the height top, with the serifs sf; the
// stroke is cut level where it has no serif.
func (d dsFrame) rod(l, r xline, y0, top float64, sf serifs) outline.Contour {
	var c pen
	if sf.footR > 0 {
		d.diagFootRight(&c, r, sf.footR)
	} else {
		c.lineTo(pt(r.at(y0), y0))
	}
	if sf.headR > 0 {
		d.diagHeadRight(&c, r, top, sf.headR)
	} else {
		c.lineTo(pt(r.at(top), top))
	}
	if sf.headL > 0 {
		d.diagHeadLeft(&c, l, top, sf.headL)
	} else {
		c.lineTo(pt(l.at(top), top))
	}
	if sf.footL > 0 {
		d.diagFootLeft(&c, l, sf.footL)
	} else {
		c.lineTo(pt(l.at(y0), y0))
	}
	return c.close()
}

// slotBetween returns the white inside a doubled stroke whose left and right
// edges of the white are l and r, between the heights y0 and y1, cut by the
// half-planes which keep it inside the letter, as a clockwise contour; nil if
// nothing is left.
func slotBetween(l, r xline, y0, y1 float64, cuts ...halfPlane) []outline.Contour {
	poly := []outline.Point{pt(l.at(y0), y0), pt(r.at(y0), y0), pt(r.at(y1), y1), pt(l.at(y1), y1)}
	for _, h := range cuts {
		poly = h.clip(poly)
	}
	poly = dedupe(poly, 1)
	if len(poly) < 3 {
		return nil
	}
	return []outline.Contour{reversed(polygon(poly...))}
}

// A halfPlane keeps the points with a x + b y + c at least zero.
type halfPlane struct{ a, b, c float64 }

// rightOf keeps what lies right of the edge e, and leftOf what lies left of
// it; overLevel and underLevel keep what lies over or under a level.
func rightOf(e xline) halfPlane      { return halfPlane{1, -e.s, -e.x0} }
func leftOf(e xline) halfPlane       { return halfPlane{-1, e.s, e.x0} }
func overLevel(y float64) halfPlane  { return halfPlane{0, 1, -y} }
func underLevel(y float64) halfPlane { return halfPlane{0, -1, y} }

func (h halfPlane) f(q outline.Point) float64 { return h.a*q.X + h.b*q.Y + h.c }

// clip returns the convex polygon cut by the half-plane.
func (h halfPlane) clip(poly []outline.Point) []outline.Point {
	var out []outline.Point
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		fa, fb := h.f(a), h.f(b)
		if fa >= 0 {
			out = append(out, a)
		}
		if (fa >= 0) != (fb >= 0) {
			t := fa / (fa - fb)
			out = append(out, pt(a.X+t*(b.X-a.X), a.Y+t*(b.Y-a.Y)))
		}
	}
	return out
}

// dedupe returns the points without those closer than min to the one before.
func dedupe(pts []outline.Point, min float64) []outline.Point {
	var out []outline.Point
	for _, q := range pts {
		if n := len(out); n > 0 && math.Hypot(out[n-1].X-q.X, out[n-1].Y-q.Y) < min {
			continue
		}
		out = append(out, q)
	}
	if n := len(out); n > 1 && math.Hypot(out[n-1].X-out[0].X, out[n-1].Y-out[0].Y) < min {
		out = out[:n-1]
	}
	return out
}

// chainEdge returns the edge of a stroke as an open contour.
func chainEdge(c stroke.Chain) outline.Contour {
	out := outline.Contour{Start: c[0].C[0]}
	for _, pc := range c {
		if pc.C.IsStraight() {
			out.Segs = append(out.Segs, outline.Segment{P3: pc.C[3]})
			continue
		}
		out.Segs = append(out.Segs, outline.Segment{IsCurve: true, P1: pc.C[1], P2: pc.C[2], P3: pc.C[3]})
	}
	return out
}

// crossing returns where the open contours a and b cross: the index of the
// segment of each and the parameter on it.
func crossing(a, b outline.Contour) (ia int, ta float64, ib int, tb float64, ok bool) {
	const n = 48
	sample := func(p0 outline.Point, sg outline.Segment, t float64) outline.Point {
		if !sg.IsCurve {
			return pt(p0.X+t*(sg.P3.X-p0.X), p0.Y+t*(sg.P3.Y-p0.Y))
		}
		u := 1 - t
		return pt(
			u*u*u*p0.X+3*u*u*t*sg.P1.X+3*u*t*t*sg.P2.X+t*t*t*sg.P3.X,
			u*u*u*p0.Y+3*u*u*t*sg.P1.Y+3*u*t*t*sg.P2.Y+t*t*t*sg.P3.Y)
	}
	starts := func(c outline.Contour) []outline.Point {
		out := []outline.Point{c.Start}
		for _, sg := range c.Segs[:len(c.Segs)-1] {
			out = append(out, sg.P3)
		}
		return out
	}
	sa, sb := starts(a), starts(b)
	for i, ga := range a.Segs {
		for j, gb := range b.Segs {
			for p := 0; p < n; p++ {
				t0, t1 := float64(p)/n, float64(p+1)/n
				a0, a1 := sample(sa[i], ga, t0), sample(sa[i], ga, t1)
				for q := 0; q < n; q++ {
					u0, u1 := float64(q)/n, float64(q+1)/n
					b0, b1 := sample(sb[j], gb, u0), sample(sb[j], gb, u1)
					// The point where a0 + s (a1 - a0) meets b0 + r (b1 - b0).
					dx1, dy1 := a1.X-a0.X, a1.Y-a0.Y
					dx2, dy2 := b1.X-b0.X, b1.Y-b0.Y
					den := dx1*dy2 - dy1*dx2
					if math.Abs(den) < 1e-12 {
						continue
					}
					s := ((b0.X-a0.X)*dy2 - (b0.Y-a0.Y)*dx2) / den
					r := ((b0.X-a0.X)*dy1 - (b0.Y-a0.Y)*dx1) / den
					if s < 0 || s > 1 || r < 0 || r > 1 {
						continue
					}
					sa0, sb0 := t0+s*(t1-t0), u0+r*(u1-u0)
					// Newton's method on the curves themselves.
					for range 8 {
						pa, pb := sample(sa[i], ga, sa0), sample(sb[j], gb, sb0)
						const h = 1e-6
						da := sample(sa[i], ga, sa0+h)
						db := sample(sb[j], gb, sb0+h)
						ax, ay := (da.X-pa.X)/h, (da.Y-pa.Y)/h
						bx, by := (db.X-pb.X)/h, (db.Y-pb.Y)/h
						det := ax*(-by) - (-bx)*ay
						if math.Abs(det) < 1e-12 {
							break
						}
						fx, fy := pb.X-pa.X, pb.Y-pa.Y
						sa0 += (fx*(-by) - (-bx)*fy) / det
						sb0 += (ax*fy - ay*fx) / det
					}
					return i, math.Max(0, math.Min(1, sa0)), j, math.Max(0, math.Min(1, sb0)), true
				}
			}
		}
	}
	return 0, 0, 0, 0, false
}

// splitSegment returns the two parts of the segment which ends at sg.P3 and
// starts at p0, cut at the parameter t.
func splitSegment(p0 outline.Point, sg outline.Segment, t float64) (left, right outline.Segment, at outline.Point) {
	lerp := func(a, b outline.Point) outline.Point { return pt(a.X+t*(b.X-a.X), a.Y+t*(b.Y-a.Y)) }
	if !sg.IsCurve {
		m := lerp(p0, sg.P3)
		return outline.Segment{P3: m}, outline.Segment{P3: sg.P3}, m
	}
	a, b, c := lerp(p0, sg.P1), lerp(sg.P1, sg.P2), lerp(sg.P2, sg.P3)
	e, f := lerp(a, b), lerp(b, c)
	m := lerp(e, f)
	return outline.Segment{IsCurve: true, P1: a, P2: e, P3: m}, outline.Segment{IsCurve: true, P1: f, P2: c, P3: sg.P3}, m
}

// startOf returns the start point of the segment i of the contour.
func startOf(c outline.Contour, i int) outline.Point {
	if i == 0 {
		return c.Start
	}
	return c.Segs[i-1].P3
}

// headTo returns the part of the open contour from its start to the point at
// the parameter t of its segment i.
func headTo(c outline.Contour, i int, t float64) outline.Contour {
	l, _, _ := splitSegment(startOf(c, i), c.Segs[i], t)
	return outline.Contour{Start: c.Start, Segs: append(append([]outline.Segment(nil), c.Segs[:i]...), l)}
}

// tailFrom returns the part of the open contour from the point at the
// parameter t of its segment i to its end.
func tailFrom(c outline.Contour, i int, t float64) outline.Contour {
	_, r, m := splitSegment(startOf(c, i), c.Segs[i], t)
	return outline.Contour{Start: m, Segs: append([]outline.Segment{r}, c.Segs[i+1:]...)}
}

// fairest returns the outline of the first of the n candidate shapes which the
// fairness analysis finds nothing in, rounded as the font is, or the one with
// the fewest findings if none is free of them; build draws the candidate with
// the given index and may panic where the stroke tool refuses its numbers.  An
// inflection on an upright edge, between the bracket of a serif and a round,
// is no finding: it is the edge of the stem of u, which the fairness baseline of
// the regular judges intended.
func fairest(name string, n int, build func(i int) outline.Path) outline.Path {
	return fairestWith(name, n, build, nil)
}

// fairestWith is fairest with the findings for which intended reports true,
// on the rounded outline, left out as well.
func fairestWith(name string, n int, build func(i int) outline.Path, intended func(outline.Path, fair.Finding) bool) outline.Path {
	var best outline.Path
	bestN := -1
	for i := range n {
		path, ok := tryBuild(build, i)
		if !ok {
			continue
		}
		k := 0
		rounded := path.Round()
		for _, f := range fair.Analyse(name, rounded, fair.Default).Findings {
			if f.Kind == fair.Inflection && math.Abs(math.Mod(f.Dir+360, 180)-90) < 1 {
				continue
			}
			if intended != nil && intended(rounded, f) {
				continue
			}
			k++
		}
		if bestN < 0 || k < bestN {
			best, bestN = path, k
		}
		if k == 0 {
			break
		}
	}
	if bestN < 0 {
		panic("mathfont: no candidate shape of " + name + " can be drawn")
	}
	return best
}

// tryBuild returns the outline which build draws with the index i, and false
// if it panics.
func tryBuild(build func(i int) outline.Path, i int) (path outline.Path, ok bool) {
	defer func() {
		if recover() != nil {
			path, ok = nil, false
		}
	}()
	return build(i), true
}
