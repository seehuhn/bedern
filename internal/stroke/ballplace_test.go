// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/fair"
)

// ballRadii are the radii of the balls of TestBallRule: 45, the ball of the
// roman, and 40, the ball of the italic.
var ballRadii = []float64{45, 40}

// curl returns an arm which rises from a stem as the arch of an n and curls
// at its crest into a ball of the radius r, the whole turned by rot degrees
// about the origin: the stroke leaves (0, 0) in the direction 60 degrees, 34
// thick, and reaches its crest, span to the right and 200 above, in the
// direction 0, 31 thick.  With ahead zero the ball is placed by its depth
// below the tangent of the outer edge at the crest (Depth); otherwise its
// centre lies ahead of the crest by ahead and depth below the outer edge, so
// that the power of the spiral or the slide of the crest takes it up.  depth
// and ahead are given for the radius 45 and scale with r.
func curl(span, depth, ahead, rot, r float64) Stroke {
	rt := func(q Point) Point { return rotate(q, rot) }
	depth *= r / 45
	ahead *= r / 45
	b := BallSpec{R: r, Neck: 35, Open: 53, TipK: Auto, Power: [2]float64{1, 2.5}}
	if ahead == 0 {
		b.Depth = depth
	} else {
		b.Centre = rt(Point{X: span + ahead, Y: 200 + 15.5 - depth})
	}
	return Stroke{
		Nib: &Nib{Major: 81.6, Minor: 28.5, Angle: 6},
		Nodes: []Node{
			{P: rt(Point{}), Dir: 60 + rot, K: Auto, W: 34, DW: 0},
			{P: rt(Point{X: span, Y: 200}), Dir: rot, K: Auto, W: 31, DW: 0},
		},
		End: End{Kind: Ball, Ball: b},
	}
}

// ballPlacements are the placements of TestBallRule: the depth of the centre
// from 2.0 to 2.4 radii, the crest at three distances from the start of the
// arm, the ball by its depth and by its centre 0.9 and 1.6 radii ahead of the
// crest, and six directions of the stroke, for each radius of ballRadii; the
// depth and the distance are those of the radius 45 and scale with the radius.
// The short tier takes three placements of each radius.
func ballPlacements(short bool) [][5]float64 {
	var out [][5]float64
	for _, r := range ballRadii {
		for _, pl := range ballPlacements45(short) {
			out = append(out, [5]float64{pl[0], pl[1], pl[2], pl[3], r})
		}
	}
	return out
}

func ballPlacements45(short bool) [][4]float64 {
	if short {
		return [][4]float64{{140, 99, 0, 0}, {110, 90, 70, 180}, {170, 108, 40, 0}}
	}
	var out [][4]float64
	for _, rot := range []float64{0, 180} {
		for _, span := range []float64{140, 170, 200} {
			for _, d := range []float64{99, 108} {
				for _, ahead := range []float64{0, 40, 70} {
					out = append(out, [4]float64{span, d, ahead, rot})
				}
			}
		}
	}
	return out
}

// TestBallRule builds the curl over the range of placements of the rule of
// the spiral and asks of each terminal: no warning; the outer edge from the
// end node to the circle has a curvature which rises steadily from that of
// the edge at the end node to that of the ball; the free arc of the ball is
// a true circle within half a unit after the rounding of the font; the wedge
// of white between the ball and the inner edge is at least ballWedgeMin wide
// 10 units from its tip; and the smoothness check finds nothing on the
// terminal.
func TestBallRule(t *testing.T) {
	for _, pl := range ballPlacements(testing.Short()) {
		span, depth, ahead, rot, r := pl[0], pl[1], pl[2], pl[3], pl[4]
		name := fmt.Sprintf("r%v-span%v-depth%v-ahead%v-rot%v", r, span, depth, ahead, rot)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkCurl(t, curl(span, depth, ahead, rot, r))
		})
	}
}

// ballWedgeMin is the least width of the white of the wedge 10 units from
// its tip: an opening of 40 degrees.
const ballWedgeMin = 6.8

func checkCurl(t *testing.T, s Stroke) {
	t.Helper()
	r, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	// The warnings about the ball; those about the arm, a stand-in for the
	// strokes of the letters, are not the rule's.
	for _, w := range r.Warnings {
		if strings.HasPrefix(w, "ball:") {
			t.Errorf("warning: %s", w)
		}
	}
	placed, err := s.placeBalls()
	if err != nil {
		t.Fatal(err)
	}
	ns, sides, err := placed.edges()
	if err != nil {
		t.Fatal(err)
	}
	n := len(ns)
	b := placed.End.Ball
	var rr Result
	cp, err := rr.ball(b, ns[n-1], sides[0][n-1], sides[1][n-1])
	if err != nil {
		t.Fatal(err)
	}
	// The ball lies on the right: the chain runs along the inner edge to
	// the tip, round the circle and back along the spiral.
	var stalk, arcs, spiral Chain
	for _, pc := range cp {
		switch {
		case pc.Kind == EndPiece:
			arcs = append(arcs, pc)
		case len(arcs) == 0:
			stalk = append(stalk, pc)
		default:
			spiral = append(spiral, pc)
		}
	}
	if len(arcs) == 0 || len(spiral) == 0 || len(stalk) == 0 {
		t.Fatalf("the terminal has %d pieces of stalk, %d of circle, %d of spiral", len(stalk), len(arcs), len(spiral))
	}
	// The spiral, walked from the end node to the circle.
	spiral = spiral.reversed()
	prev := math.Abs(sides[1][n-1].K)
	if k := math.Abs(spiral[0].C.Curvature(0)); math.Abs(k-prev) > 1e-3*(1+prev) {
		t.Errorf("the spiral leaves the end node with the curvature %.5f, the edge has %.5f", k, prev)
	}
	for _, pc := range spiral {
		for i := 1; i <= 32; i++ {
			k := math.Abs(pc.C.Curvature(float64(i) / 32))
			if k < prev-0.02/b.R {
				t.Errorf("the curvature of the spiral falls from %.5f to %.5f at %.0f", prev, k, pc.C.At(float64(i)/32))
				break
			}
			prev = math.Max(prev, k)
		}
	}
	if k := math.Abs(spiral[len(spiral)-1].C.Curvature(1)); math.Abs(k*b.R-1) > 0.02 {
		t.Errorf("the spiral reaches the ball with the curvature %.5f, the ball has %.5f", k, 1/b.R)
	}
	// The wedge: the white between the circle and the inner edge 10 units
	// from the tip.
	tip := stalk[len(stalk)-1].C[3]
	onStalk := pointBack(stalk, 10)
	onBall := add(b.Centre, rotate(sub(tip, b.Centre), -10/b.R*180/math.Pi))
	if w := dist(onStalk, onBall); w < ballWedgeMin {
		t.Errorf("the wedge is %.1f wide 10 from its tip, want at least %v", w, ballWedgeMin)
	}
	// The free arc after rounding: the arcs run counter-clockwise from a0
	// to a1.
	path := (&LetterResult{Contours: r.Contours}).Path().Round()
	a0 := angleOf(sub(arcs[0].C[0], b.Centre))
	a1 := angleOf(sub(arcs[len(arcs)-1].C[3], b.Centre))
	inArc := func(q Point) bool {
		a := angleOf(sub(q, b.Centre))
		from, span := math.Mod(a-a0+720, 360), math.Mod(a1-a0+720, 360)
		return from > 5 && from < span-5
	}
	for _, c := range path {
		cur := c.Start
		for _, sg := range c.Segs {
			for i := 0; i <= 16; i++ {
				u := float64(i) / 16
				v := 1 - u
				q := Point{X: v*cur.X + u*sg.P3.X, Y: v*cur.Y + u*sg.P3.Y}
				if sg.IsCurve {
					q = Point{X: v*v*v*cur.X + 3*v*v*u*sg.P1.X + 3*v*u*u*sg.P2.X + u*u*u*sg.P3.X, Y: v*v*v*cur.Y + 3*v*v*u*sg.P1.Y + 3*v*u*u*sg.P2.Y + u*u*u*sg.P3.Y}
				}
				if d := dist(q, b.Centre); math.Abs(d-b.R) < 3 && inArc(q) && math.Abs(d-b.R) > 0.5 {
					t.Errorf("the ball is %.2f off its circle at %.1f after rounding", d-b.R, q)
				}
			}
			cur = sg.P3
		}
	}
	for _, f := range fair.Analyse("curl", path, fair.Italic).Findings {
		if dist(Point{X: f.X, Y: f.Y}, b.Centre) < 1.6*b.R {
			t.Errorf("%s at %.0f,%.0f: %s", f.Kind, f.X, f.Y, f.Detail)
		}
	}
}

// pointBack returns the point of the chain the arc length l before its end.
func pointBack(ch Chain, l float64) Point {
	for i := len(ch) - 1; i >= 0; i-- {
		c := ch[i].C
		const n = 200
		prev := c.At(1)
		for j := n - 1; j >= 0; j-- {
			q := c.At(float64(j) / n)
			l -= dist(q, prev)
			if l <= 0 {
				return q
			}
			prev = q
		}
	}
	return ch.start()
}
