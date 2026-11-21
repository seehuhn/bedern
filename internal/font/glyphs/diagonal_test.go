// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the diagonal part use made-up values and no letter: a thick
// diagonal which leans to the right going down (lean -0.30) and a thin one which
// leans to the left (lean +0.30) meeting at a vertex at x = 300 with a flat of 20
// which dips 10 below the baseline, under a top at 400, with head serifs of the
// lengths 40 outside and 60 inside.

const (
	tdThick   = 70.0
	tdThin    = 30.0
	tdLean    = 0.30
	tdFlat    = 20.0
	tdDepth   = 10.0
	tdTop     = 400.0
	tdCentre  = 300.0
	tdSerifTh = 33.0
)

var tdBracket = design.DiagBracket{Reach: 35, Obtuse: 55, Acute: 32, ReachAcute: 28}
var tdOut = diagSerif{Over: 45, Thick: tdSerifTh, Bracket: tdBracket}
var tdIn = diagSerif{Over: 55, Thick: tdSerifTh, Bracket: tdBracket}

func tdVertex() vertex { return newVertex(tdCentre, tdFlat, tdDepth, tdLean, tdThick, tdThin) }

// tight checks that got is within one hundredth of a unit of want.
func tight(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s = %.4f, want %.4f", what, got, want)
	}
}

// tdContour draws the made-up v from the parts, as the letter v does.
func tdContour(p design.Params) outline.Path {
	v := tdVertex()
	l, r := v.Left, v.Right
	outL := l.headServif(edgeLeft, tdTop, tdOut)
	outR := r.headServif(edgeRight, tdTop, tdOut)
	inR := r.headServif(edgeLeft, tdTop, tdIn)
	inL := l.headServif(edgeRight, tdTop, tdIn)
	c := join(
		outL,
		lineChain(outL.end(), v.outerLeft()),
		v.flat(),
		lineChain(v.outerRight(), outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), v.crotch()),
		lineChain(v.crotch(), inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	var path outline.Path
	path.MoveTo(c.start().X, c.start().Y)
	c.appendTo(&path)
	return path
}

func TestDiagonalGeometry(t *testing.T) {
	t.Parallel()
	d := diagonal{X: 100, Y: 0, Lean: -tdLean, Thick: tdThick}
	// The thickness at right angles: the distance of a point of the right-hand edge
	// from the left-hand edge line.
	a, b := d.edgeAt(edgeLeft, 0), d.edgeAt(edgeLeft, 300)
	q := d.edgeAt(edgeRight, 120)
	dist := math.Abs((b.X-a.X)*(a.Y-q.Y)-(a.X-q.X)*(b.Y-a.Y)) / math.Hypot(b.X-a.X, b.Y-a.Y)
	tight(t, "thickness at right angles", dist, tdThick)
	// The angle of each edge to the vertical.
	for _, e := range []edge{edgeLeft, edgeRight} {
		a, b := d.edgeAt(e, 0), d.edgeAt(e, 300)
		tight(t, "angle of the edge in degrees", math.Atan2(b.X-a.X, b.Y-a.Y)*180/math.Pi, -math.Atan(tdLean)*180/math.Pi)
	}
	tight(t, "horizontal width", d.width(), tdThick*math.Hypot(1, tdLean))
	tight(t, "centre line", d.centreAt(200), 100-tdLean*200)
	if u := d.down(); math.Abs(math.Hypot(u.X, u.Y)-1) > 1e-12 || u.Y >= 0 || u.X <= 0 {
		t.Errorf("down is %v, want a unit vector which leans to the right and points down", u)
	}
	// The edge is the one through the given point.
	e := diagFromEdge(edgeRight, 250, 40, 0.2, 50)
	tight(t, "edge through a point", e.edgeAt(edgeRight, 40).X, 250)
	tight(t, "other edge", e.edgeAt(edgeLeft, 40).X, 250-e.width())
}

func TestEdgeMeet(t *testing.T) {
	t.Parallel()
	a := diagonal{X: 100, Y: 0, Lean: 0.4, Thick: 60}
	b := diagonal{X: 300, Y: 0, Lean: -0.2, Thick: 20}
	q := edgeMeet(a, edgeRight, b, edgeLeft)
	tight(t, "meeting on the first edge", q.X, a.edgeAt(edgeRight, q.Y).X)
	tight(t, "meeting on the second edge", q.X, b.edgeAt(edgeLeft, q.Y).X)
	defer func() {
		if recover() == nil {
			t.Error("the meeting of two parallel edges does not panic")
		}
	}()
	edgeMeet(a, edgeLeft, diagonal{X: 0, Y: 0, Lean: 0.4, Thick: 10}, edgeRight)
}

// The vertex: the outer edges meet the flat at its two ends, which are the given
// depth below the baseline and the given width apart; the crotch is where the
// inner edges meet, at the angle twice that of a diagonal.
func TestVertex(t *testing.T) {
	t.Parallel()
	v := tdVertex()
	ol, or := v.outerLeft(), v.outerRight()
	tight(t, "left end of the flat", ol.X, tdCentre-tdFlat/2)
	tight(t, "right end of the flat", or.X, tdCentre+tdFlat/2)
	tight(t, "depth of the flat", ol.Y, -tdDepth)
	tight(t, "height of the right end", or.Y, -tdDepth)
	tight(t, "left outer edge at the flat", v.Left.edgeAt(edgeLeft, -tdDepth).X, ol.X)
	tight(t, "right outer edge at the flat", v.Right.edgeAt(edgeRight, -tdDepth).X, or.X)
	c := v.crotch()
	tight(t, "crotch on the left diagonal", c.X, v.Left.edgeAt(edgeRight, c.Y).X)
	tight(t, "crotch on the right diagonal", c.X, v.Right.edgeAt(edgeLeft, c.Y).X)
	tight(t, "crotch in the middle of the inner edges", c.X, tdCentre+(v.Left.width()-v.Right.width())/2)
	// Half the angle of the crotch is the angle of a diagonal.
	l, r := v.Left.edgeAt(edgeRight, c.Y+100), v.Right.edgeAt(edgeLeft, c.Y+100)
	tight(t, "angle of the crotch in degrees", 2*math.Atan2((r.X-l.X)/2, 100)*180/math.Pi, 2*math.Atan(tdLean)*180/math.Pi)
	// The crotch lies above the flat, at the height the geometry gives.
	want := (v.Left.width()+v.Right.width()-tdFlat)/(2*tdLean) - tdDepth
	tight(t, "height of the crotch", c.Y, want)
}

// The flat of the vertex is its lowest ink and as wide as it is said to be.
func TestVertexFlat(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path := tdContour(p)
	x0, y0, _, _ := path.BBox()
	_ = x0
	tight(t, "depth", y0, -tdDepth)
	xs := path.CrossingsX(-tdDepth + 0.01)
	if len(xs) != 2 {
		t.Fatalf("crossings just above the flat: %v", xs)
	}
	tight(t, "width of the flat", xs[1]-xs[0], tdFlat+2*tdLean*0.01)
	// The flat is a straight line between two straight edges.
	ct := path[0]
	nodes := contourNodes(ct)
	var flat int = -1
	for i := range ct.Segs {
		if nodes[i].Y == -tdDepth && nodes[i+1].Y == -tdDepth && nodes[i+1].X-nodes[i].X > 0 {
			flat = i
		}
	}
	if flat < 0 || ct.Segs[flat].IsCurve {
		t.Fatalf("no straight flat segment at the bottom")
	}
}

// The serif has a flat top on the line of the top, a tip of the thickness Thick whose
// distance from the edge at the level of the underside is the overhang, and a flat
// underside at the tip.
func TestHeadServifOverhangAndThickness(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path := tdContour(p)
	v := tdVertex()
	for _, c := range []struct {
		name  string
		d     diagonal
		e     edge
		serif diagSerif
	}{
		{"outer serif of the left diagonal", v.Left, edgeLeft, tdOut},
		{"inner serif of the left diagonal", v.Left, edgeRight, tdIn},
		{"inner serif of the right diagonal", v.Right, edgeLeft, tdIn},
		{"outer serif of the right diagonal", v.Right, edgeRight, tdOut},
	} {
		side := 1.0
		if c.e == edgeLeft {
			side = -1
		}
		under := tdTop - c.serif.Thick
		tipX := c.d.edgeAt(c.e, under).X + side*c.serif.Over
		// The tip is cut vertically at the overhang from the edge at the underside.
		got := nearest(t, c.name+": tip", path.CrossingsX(tdTop-c.serif.Thick/2), tipX)
		tight(t, c.name+": overhang", side*(got-c.d.edgeAt(c.e, under).X), c.serif.Over)
		// Five units in from the tip the serif is as thick as at the tip, up to the
		// fall of the bracket, and its top is on the line.
		const inset = 5.0
		ys := path.CrossingsY(tipX - side*inset)
		upper := nearest(t, c.name+": top", ys, tdTop)
		tight(t, c.name+": top of the serif", upper, tdTop)
		lower := math.Inf(-1)
		for _, y := range ys {
			if y < upper-1 && y > lower {
				lower = y
			}
		}
		if math.Abs(lower-under) > 0.15 {
			t.Errorf("%s: the underside is at %.3f five units from the tip, want %.3f", c.name, lower, under)
		}
	}
}

// The brackets are tangent to the underside of the serif at the tip and to the
// edge of the diagonal where they meet it, and they turn one way only (concave).
func TestHeadServifBracket(t *testing.T) {
	t.Parallel()
	v := tdVertex()
	for _, c := range []struct {
		name string
		d    diagonal
		e    edge
		s    diagSerif
	}{
		{"outer left", v.Left, edgeLeft, tdOut},
		{"inner left", v.Left, edgeRight, tdIn},
		{"inner right", v.Right, edgeLeft, tdIn},
		{"outer right", v.Right, edgeRight, tdOut},
	} {
		ch := c.d.headServif(c.e, tdTop, c.s)
		if err := ch.connected(); err != nil {
			t.Fatal(err)
		}
		if ch[0].P0 != c.d.edgeAt(c.e, tdTop) {
			t.Errorf("%s: the serif starts at %v, want the edge at the top %v", c.name, ch[0].P0, c.d.edgeAt(c.e, tdTop))
		}
		// Lines first (top, tip, flat underside), then the bracket.
		if !ch[0].line || !ch[1].line || !ch[2].line || ch[3].line {
			t.Fatalf("%s: want three lines then the bracket", c.name)
		}
		br := ch[3:]
		// Tangent to the underside, which is horizontal, where it leaves it.
		st := br.startTangent()
		if math.Abs(st.Y) > 1e-9 {
			t.Errorf("%s: the bracket leaves the tip at the slope %v, want a horizontal tangent", c.name, st)
		}
		// Tangent to the edge where it meets it: parallel to the edge, pointing down.
		et := br.endTangent()
		down := c.d.down()
		if cross := et.X*down.Y - et.Y*down.X; math.Abs(cross) > 1e-6 || et.X*down.X+et.Y*down.Y < 0 {
			t.Errorf("%s: the bracket arrives with the direction %v, want that of the edge %v", c.name, et, down)
		}
		// It ends on the edge.
		end := br.end()
		tight(t, c.name+": end on the edge", end.X, c.d.edgeAt(c.e, end.Y).X)
		// It meets the edge below the underside of the tip, by the run of its side: the
		// smaller one where the white corner is acute.
		run := c.s.Bracket.Obtuse
		if (c.e == edgeRight) == (c.d.down().X > 0) {
			run = c.s.Bracket.Acute
		}
		tight(t, c.name+": meeting height", end.Y, tdTop-c.s.Thick-run)
		// Concave: the curve turns the same way throughout, which is to the right
		// (clockwise) when it comes from a serif on the left of its edge and to the
		// left on the right of its edge.
		var heads []float64
		for _, pc := range br {
			for k := 0; k <= 20; k++ {
				tg := pc.tangent(float64(k) / 20)
				heads = append(heads, math.Atan2(tg.Y, tg.X))
			}
		}
		for k := 1; k < len(heads); k++ {
			d := turn(heads[k-1], heads[k])
			if c.e == edgeRight {
				d = -d
			}
			if d > 1e-9 {
				t.Errorf("%s: the bracket turns the wrong way by %.4f at sample %d", c.name, d, k)
				break
			}
		}
	}
}

func TestDiagonalContourIsSimpleAndCounterClockwise(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path := tdContour(p)
	if len(path) != 1 {
		t.Fatalf("%d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Errorf("the contour is clockwise (area %v)", path[0].SignedArea())
	}
	if q := path.SelfIntersections(); len(q) > 0 {
		t.Errorf("the contour crosses itself at %v", q)
	}
	// The inherited structural checks hold for the made-up letter as well: the
	// extremes of the brackets are on-curve points, the winding is simple and the
	// parts keep the clearance floor (the crotch is wider than the slit angle).
	c := structureCase{params: p, glyph: "x", path: path, advance: 600}
	for _, ch := range structureChecks {
		if ch.name == "bbox" {
			continue
		}
		for _, msg := range ch.run(c) {
			t.Errorf("check %s: %s", ch.name, msg)
		}
	}
	// The edges between the serifs and the flat are single straight segments.
	ct := path[0]
	nodes := contourNodes(ct)
	var edges int
	for i, s := range ct.Segs {
		if s.IsCurve {
			continue
		}
		a, b := nodes[i], s.P3
		if math.Abs(a.Y-b.Y) > 1e-9 && math.Abs((b.X-a.X)/(b.Y-a.Y)) > tdLean-1e-9 && math.Abs(math.Abs((b.X-a.X)/(b.Y-a.Y))-tdLean) < 1e-9 {
			edges++
		}
	}
	if edges != 4 {
		t.Errorf("%d straight segments with the lean of a diagonal, want 4 (two outer edges and two inner edges)", edges)
	}
}

// The bracket lies in the corner between the underside and the edge: within the
// height of its run and on the white side of the edge.
func TestHeadServifBracketStaysInsideTheCorner(t *testing.T) {
	t.Parallel()
	v := tdVertex()
	// The curve lies between the underside line and the edge: above the line
	// y = lower corner of the tip - the longer run and not beyond the edge.
	for _, e := range []edge{edgeLeft, edgeRight} {
		ch := v.Left.headServif(e, tdTop, tdOut)
		side := 1.0
		if e == edgeLeft {
			side = -1
		}
		for _, pc := range ch[3:] {
			for k := 0; k <= 50; k++ {
				q := pc.at(float64(k) / 50)
				if q.Y > tdTop-tdOut.Thick+1e-9 || q.Y < tdTop-tdOut.Thick-tdOut.Bracket.Obtuse-1e-9 {
					t.Fatalf("bracket point %v outside the height range", q)
				}
				// On the white side of the edge (outside the diagonal).
				if side*(q.X-v.Left.edgeAt(e, q.Y).X) < -0.3 {
					t.Errorf("bracket point %v lies inside the diagonal", q)
				}
			}
		}
	}
}
