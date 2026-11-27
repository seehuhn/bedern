// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The broad pen of the Greek: its edge at penAngle degrees above the
// horizontal, near upright, so that horizontals are heavy and verticals light;
// penMajor and penMinor are its thickest and thinnest line.
const (
	penAngle = 80.0
	penMajor = 80.0
	penMinor = 50.0
)

// penNib returns the pen of the Greek.
func penNib(p design.Params) *stroke.Nib {
	return &stroke.Nib{Major: penMajor, Minor: penMinor, Angle: penAngle}
}

// penWidth returns the thickness which the pen gives to a stroke travelling
// in the direction dir (degrees).
func penWidth(b *stroke.Nib, dir float64) float64 {
	t := (dir + 90 - b.Angle) * math.Pi / 180
	return 2 * math.Hypot(math.Cos(t)*b.Major/2, math.Sin(t)*b.Minor/2)
}

// ellipseNode returns the node of the centre line on the ellipse about
// (cx, cy) with the half axes a and b at which the stroke travels in the
// direction dir (degrees), anticlockwise if ccw, with the curvature of the
// ellipse there and the thickness from the pen.
func ellipseNode(cx, cy, a, b, dir float64, ccw bool) stroke.Node {
	th := dir * math.Pi / 180
	var phi float64
	if ccw {
		phi = math.Atan2(-b*math.Cos(th), a*math.Sin(th))
	} else {
		phi = math.Atan2(b*math.Cos(th), -a*math.Sin(th))
	}
	s, c := math.Sin(phi), math.Cos(phi)
	k := a * b / math.Pow(a*a*s*s+b*b*c*c, 1.5)
	if !ccw {
		k = -k
	}
	return stroke.Node{P: pt(cx+a*c, cy+b*s), Dir: dir, K: k, W: auto, DW: auto}
}

// ellipseAt returns the node of the centre line on the ellipse at the angle
// phi (degrees) of its parameter, travelling anticlockwise if ccw.
func ellipseAt(cx, cy, a, b, phi float64, ccw bool) stroke.Node {
	f := phi * math.Pi / 180
	s, c := math.Sin(f), math.Cos(f)
	t := pt(-a*s, b*c)
	if !ccw {
		t = pt(a*s, -b*c)
	}
	k := a * b / math.Pow(a*a*s*s+b*b*c*c, 1.5)
	if !ccw {
		k = -k
	}
	return stroke.Node{P: pt(cx+a*c, cy+b*s), Dir: math.Atan2(t.Y, t.X) * 180 / math.Pi, K: k, W: auto, DW: auto}
}

// arcNode returns the node on the circle about (cx, cy) of the radius r at
// which the stroke travels in the direction dir, anticlockwise if ccw.
func arcNode(cx, cy, r, dir float64, ccw bool) stroke.Node {
	return ellipseNode(cx, cy, r, r, dir, ccw)
}

// centreX returns the abscissa of the centre line through the nodes at the
// height y, for a centre line which runs downwards.
func centreX(nodes []stroke.Node, y float64) float64 {
	for i := 1; i < len(nodes); i++ {
		c := stroke.CentreLine(nodes[i-1], nodes[i])
		const steps = 200
		for j := 0; j < steps; j++ {
			p, q := c.At(float64(j)/steps), c.At(float64(j+1)/steps)
			if (p.Y-y)*(q.Y-y) <= 0 && p.Y != q.Y {
				return p.X + (q.X-p.X)*(y-p.Y)/(q.Y-p.Y)
			}
		}
	}
	return math.NaN()
}

// meetEllipse returns the angle (degrees) of the parameter between lo and hi
// at which the ellipse about (cx, cy) with the half axes a and b meets the
// centre line through the nodes, coming from inside it at lo.
func meetEllipse(cx, cy, a, b, lo, hi float64, nodes []stroke.Node) float64 {
	f := func(phi float64) float64 {
		t := phi * math.Pi / 180
		return cx + a*math.Cos(t) - centreX(nodes, cy+b*math.Sin(t))
	}
	flo := f(lo)
	for range 50 {
		m := (lo + hi) / 2
		if fm := f(m); (fm < 0) == (flo < 0) {
			lo, flo = m, fm
		} else {
			hi = m
		}
	}
	return (lo + hi) / 2
}
