// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// Contours made of straight lines, circular arcs and lead spirals between them.
//
// A lead spiral leads a straight line into a circle: it leaves the line with
// the curvature zero, its curvature rises with the power leadPower of the length
// run, and it ends on the circle, after the turn leadTurn, with the direction
// and the curvature of the circle.  Such a contour is described exactly
// (exactPath) and drawn as cubic curves, each with the points, the directions
// and the curvatures of the exact contour at its ends (g2Cubic).

// leadTurn is the angle (radians) by which a lead spiral turns.
const leadTurn = 15 * math.Pi / 180

// leadPower is the power of the length run with which the curvature of a lead
// spiral rises.
const leadPower = 2.0

// leadUnit returns the lead spiral into the circle of radius 1, in the frame
// in which it starts at the origin in the direction 0 and turns to the left:
// its length and the centre of the circle.
func leadUnit() (length float64, centre outline.Point) {
	length = (leadPower + 1) * leadTurn
	e := leadPoint(length, length)
	return length, outline.Point{X: e.X - math.Sin(leadTurn), Y: e.Y + math.Cos(leadTurn)}
}

// leadHeading returns the turn of the lead spiral of the length L after the
// length s.
func leadHeading(L, s float64) float64 {
	return leadTurn * math.Pow(s/L, leadPower+1)
}

// leadPoint returns the point of the lead spiral of the length L after the
// length s, in the frame of leadUnit, by Simpson's rule.
func leadPoint(L, s float64) outline.Point {
	const n = 400
	h := s / n
	var q outline.Point
	for i := 0; i <= n; i++ {
		w := 2.0
		if i == 0 || i == n {
			w = 1
		} else if i%2 == 1 {
			w = 4
		}
		phi := leadHeading(L, float64(i)*h)
		q.X += w * h / 3 * math.Cos(phi)
		q.Y += w * h / 3 * math.Sin(phi)
	}
	return q
}

// An exactPiece is one part of an exactPath: a line, an arc or a lead spiral.
type exactPiece interface {
	length() float64
	// at returns the point, the direction (radians) and the signed curvature
	// after the length s.
	at(s float64) (outline.Point, float64, float64)
}

// exactLine is the line from P in the direction Phi of the length Len.
type exactLine struct {
	P        outline.Point
	Phi, Len float64
}

func (l exactLine) length() float64 { return l.Len }

func (l exactLine) at(s float64) (outline.Point, float64, float64) {
	return outline.Point{X: l.P.X + s*math.Cos(l.Phi), Y: l.P.Y + s*math.Sin(l.Phi)}, l.Phi, 0
}

// exactArc is the counter-clockwise arc of the circle about C of the radius R
// from the angle A0 over the angle Sweep.
type exactArc struct {
	C            outline.Point
	R, A0, Sweep float64
}

func (a exactArc) length() float64 { return a.R * a.Sweep }

func (a exactArc) at(s float64) (outline.Point, float64, float64) {
	t := a.A0 + s/a.R
	return outline.Point{X: a.C.X + a.R*math.Cos(t), Y: a.C.Y + a.R*math.Sin(t)}, t + math.Pi/2, 1 / a.R
}

// exactLead is the lead spiral into the circle of the radius R which starts at
// Q in the direction Phi and turns to the left if Sign is 1 and to the right if
// it is -1; if Back is set, it is walked from the circle to Q.
type exactLead struct {
	Q       outline.Point
	Phi, R  float64
	Sign    float64
	Back    bool
	unitLen float64
}

func newExactLead(q outline.Point, phi, r, sign float64, back bool) exactLead {
	L, _ := leadUnit()
	return exactLead{Q: q, Phi: phi, R: r, Sign: sign, Back: back, unitLen: L}
}

func (l exactLead) length() float64 { return l.R * l.unitLen }

func (l exactLead) at(s float64) (outline.Point, float64, float64) {
	if l.Back {
		s = l.length() - s
	}
	v := leadPoint(l.unitLen, s/l.R)
	v.X, v.Y = v.X*l.R, l.Sign*v.Y*l.R
	c, sn := math.Cos(l.Phi), math.Sin(l.Phi)
	q := outline.Point{X: l.Q.X + v.X*c - v.Y*sn, Y: l.Q.Y + v.X*sn + v.Y*c}
	h := l.Phi + l.Sign*leadHeading(l.unitLen, s/l.R)
	k := l.Sign * math.Pow(s/(l.R*l.unitLen), leadPower) / l.R
	if l.Back {
		return q, h + math.Pi, -k
	}
	return q, h, k
}

// exactPath is a closed contour of exact pieces, each starting where the one
// before ends.
type exactPath []exactPiece

// at returns the point, the direction and the curvature of the contour after
// the length s from its start.
func (p exactPath) at(s float64) (outline.Point, float64, float64) {
	for i, pc := range p {
		if s <= pc.length() || i == len(p)-1 {
			return pc.at(math.Min(s, pc.length()))
		}
		s -= pc.length()
	}
	panic("unreachable")
}

// leadTolerance is the largest distance of a cubic curve of chain from the
// exact contour, unless the curve is shorter than twice leadMinPiece.
const (
	leadTolerance = 0.5
	leadMinPiece  = 16.0
)

// chain returns the contour as straight lines and cubic curves.  The nodes are
// the extremes, where the direction is a multiple of a quarter turn, the points
// a quarter of the length of a line from its ends, between which the line is
// drawn straight, and the ends of the arcs, except those within leadMinPiece of
// another node; a curve between two nodes which strays from the exact contour by more
// than leadTolerance is cut in two halves of equal length, as long as they are
// at least leadMinPiece long.
func (p exactPath) chain() (chain, error) {
	var cuts, soft []float64
	start := 0.0
	for _, pc := range p {
		L := pc.length()
		switch pc.(type) {
		case exactLine:
			cuts = append(cuts, start+L/4, start+3*L/4)
		default:
			// The extremes.
			_, h0, _ := pc.at(0)
			_, h1, _ := pc.at(L)
			lo, hi := math.Min(h0, h1), math.Max(h0, h1)
			for q := math.Ceil(lo/(math.Pi/2)) * math.Pi / 2; q < hi; q += math.Pi / 2 {
				a, b := 0.0, L
				for i := 0; i < 60; i++ {
					m := (a + b) / 2
					_, h, _ := pc.at(m)
					if (h-q)*(h1-h0) < 0 {
						a = m
					} else {
						b = m
					}
				}
				cuts = append(cuts, start+(a+b)/2)
			}
		}
		if _, ok := pc.(exactArc); ok {
			soft = append(soft, start, start+L)
		}
		start += L
	}
	total := start
	// The ends of the arcs, except those near another node.
	hard := append([]float64(nil), cuts...)
	for _, s := range soft {
		near := false
		for _, c := range hard {
			d := math.Abs(c - s)
			if math.Min(d, total-d) < leadMinPiece {
				near = true
			}
		}
		if !near {
			cuts = append(cuts, s)
		}
	}
	sort.Float64s(cuts)
	if len(cuts) < 2 {
		return nil, fmt.Errorf("the contour has too few nodes")
	}
	at := func(s float64) (outline.Point, float64, float64) { return p.at(math.Mod(s, total)) }
	var ch chain
	var add func(a, b float64) error
	add = func(a, b float64) error {
		pa, ha, ka := at(a)
		pb, hb, kb := at(b)
		ta, tb := dirOf(ha), dirOf(hb)
		var c cubic
		var err error
		switch {
		case ka == 0 && kb == 0:
			ch = append(ch, lineSeg(pa, pb))
			return nil
		case kb != 0:
			c, err = g2Cubic(pa, ta, ka, pb, tb, kb)
		default:
			c, err = g2Cubic(pb, neg(tb), -kb, pa, neg(ta), -ka)
			c = cubic{c.P3, c.P2, c.P1, c.P0}
		}
		if err != nil {
			c, err = nearG2(pa, ta, ka, pb, tb, kb), nil
		}
		if err == nil {
			// The distance of the curve from the exact contour.
			worst := 0.0
			for i := 1; i < 16; i++ {
				q := c.at(float64(i) / 16)
				best := math.Inf(1)
				for j := 0; j <= 200; j++ {
					r, _, _ := at(a + (b-a)*float64(j)/200)
					best = math.Min(best, math.Hypot(q.X-r.X, q.Y-r.Y))
				}
				worst = math.Max(worst, best)
			}
			if worst <= leadTolerance {
				ch = append(ch, curveSeg(c))
				return nil
			}
		}
		if b-a >= 2*leadMinPiece {
			if err := add(a, (a+b)/2); err != nil {
				return err
			}
			return add((a+b)/2, b)
		}
		if err != nil {
			return err
		}
		ch = append(ch, curveSeg(c))
		return nil
	}
	for i := range cuts {
		a, b := cuts[i], cuts[(i+1)%len(cuts)]
		if b <= a {
			b += total
		}
		if err := add(a, b); err != nil {
			return nil, err
		}
	}
	for i := 1; i < len(ch); i++ {
		ch[i].P0 = ch[i-1].P3
	}
	ch[len(ch)-1].P3 = ch[0].P0
	return ch, nil
}

// leadWedge returns the exact contour of the circle about c0 of the radius r0
// and the circle about c1 of the radius r1, joined by two lines, each of which
// runs into both circles on lead spirals.  The contour runs counter-clockwise
// and starts on the first circle.
func leadWedge(c0 outline.Point, r0 float64, c1 outline.Point, r1 float64) exactPath {
	_, uc := leadUnit()
	d := outline.Point{X: c1.X - c0.X, Y: c1.Y - c0.Y}
	dl := math.Hypot(d.X, d.Y)
	psi := math.Atan2(d.Y, d.X)
	y0, y1 := uc.Y*r0, uc.Y*r1
	x0, x1 := uc.X*r0, uc.X*r1
	// The side from c0 to c1, with the circles on its left.
	phi := psi - math.Asin((y1-y0)/dl)
	u := dirOf(phi)
	n := outline.Point{X: -u.Y, Y: u.X}
	q0 := outline.Point{X: c0.X + x0*u.X - y0*n.X, Y: c0.Y + x0*u.Y - y0*n.Y}
	q1 := outline.Point{X: c1.X - x1*u.X - y1*n.X, Y: c1.Y - x1*u.Y - y1*n.Y}
	// The side from c1 to c0.
	phiB := psi + math.Pi - math.Asin((y0-y1)/dl)
	w := dirOf(phiB)
	m := outline.Point{X: -w.Y, Y: w.X}
	q1b := outline.Point{X: c1.X + x1*w.X - y1*m.X, Y: c1.Y + x1*w.Y - y1*m.Y}
	q0b := outline.Point{X: c0.X - x0*w.X - y0*m.X, Y: c0.Y - x0*w.Y - y0*m.Y}
	sweep := func(a0, a1 float64) float64 {
		for a1 <= a0 {
			a1 += 2 * math.Pi
		}
		for a1 > a0+2*math.Pi {
			a1 -= 2 * math.Pi
		}
		return a1 - a0
	}
	a0s, a0e := phiB+leadTurn-math.Pi/2, phi-leadTurn-math.Pi/2
	a1s, a1e := phi+leadTurn-math.Pi/2, phiB-leadTurn-math.Pi/2
	return exactPath{
		exactArc{C: c0, R: r0, A0: a0s, Sweep: sweep(a0s, a0e)},
		newExactLead(q0, phi+math.Pi, r0, -1, true),
		exactLine{P: q0, Phi: phi, Len: math.Hypot(q1.X-q0.X, q1.Y-q0.Y)},
		newExactLead(q1, phi, r1, 1, false),
		exactArc{C: c1, R: r1, A0: a1s, Sweep: sweep(a1s, a1e)},
		newExactLead(q1b, phiB+math.Pi, r1, -1, true),
		exactLine{P: q1b, Phi: phiB, Len: math.Hypot(q0b.X-q1b.X, q0b.Y-q1b.Y)},
		newExactLead(q0b, phiB, r0, 1, false),
	}
}
