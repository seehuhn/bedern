// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// A built is a stroke on which the properties are tested, with its outline.
type built struct {
	name   string
	stroke Stroke
	result *Result
}

var testBuilt struct {
	once sync.Once
	all  []built
	err  error
}

// testStrokes returns the strokes on which the properties are tested, built
// once for all tests: the reproduced roman parts and the strokes of the italic
// trial, sheared.
func testStrokes(t *testing.T) []built {
	testBuilt.once.Do(func() {
		p := design.Regular()
		var all []built
		for _, pt := range RomanParts(p) {
			for i, s := range pt.Strokes {
				all = append(all, built{name: fmt.Sprintf("%s/%d", pt.Name, i), stroke: s})
			}
		}
		lt := italicLetters(p)
		var names []string
		for name := range lt {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			for i, s := range lt[name].Strokes {
				all = append(all, built{name: fmt.Sprintf("italic %s/%d", name, i), stroke: s.ShearedRatio(SlantNum, SlantDen)})
			}
		}
		errs := make([]error, len(all))
		var wg sync.WaitGroup
		for i := range all {
			wg.Add(1)
			go func() {
				defer wg.Done()
				all[i].result, errs[i] = all[i].stroke.Build()
			}()
		}
		wg.Wait()
		for i, err := range errs {
			if err != nil {
				testBuilt.err = fmt.Errorf("%s: %w", all[i].name, err)
				return
			}
		}
		testBuilt.all = all
	})
	if testBuilt.err != nil {
		t.Fatal(testBuilt.err)
	}
	return testBuilt.all
}

// smoothJoins calls f for every join of two pieces of a contour at which the
// outline is meant to be smooth: two edge pieces, or an edge and the spiral of
// a ball.
func smoothJoins(r *Result, f func(a, b Piece)) {
	for _, c := range r.Contours {
		for i := range c {
			a, b := c[i], c[(i+1)%len(c)]
			if a.Kind == EdgePiece && b.Kind == EdgePiece && dist(a.C[3], b.C[0]) < 1e-9 {
				if math.Abs(angleBetween(a.C.Tangent(1), b.C.Tangent(0))) < 5 {
					f(a, b)
				}
			}
		}
	}
}

// TestJoinsAreCurvatureContinuous asks that at every join of two edge pieces
// the direction agrees within 0.01 degrees and the curvature within 2 per cent
// of the larger or 1e-5 per unit, for every stroke whose curves the rules
// could draw with both curvatures (a stroke with a warning is reported).
func TestJoinsAreCurvatureContinuous(t *testing.T) {
	for _, b := range testStrokes(t) {
		name, r := b.name, b.result
		if len(r.Warnings) > 0 {
			t.Logf("%s: %d warnings, first: %s", name, len(r.Warnings), r.Warnings[0])
		}
		smoothJoins(r, func(a, b Piece) {
			if d := math.Abs(angleBetween(a.C.Tangent(1), b.C.Tangent(0))); d > 0.01 {
				t.Errorf("%s: at %.1f,%.1f the direction turns by %.3f degrees", name, a.C[3].X, a.C[3].Y, d)
			}
			ka, kb := a.C.Curvature(1), b.C.Curvature(0)
			if math.Abs(ka-kb) > math.Max(1e-5, 0.02*math.Max(math.Abs(ka), math.Abs(kb))) && len(r.Warnings) == 0 {
				t.Errorf("%s: at %.1f,%.1f the curvature jumps from %.5f to %.5f", name, a.C[3].X, a.C[3].Y, ka, kb)
			}
		})
	}
}

// TestExtremesAreNodes asks that no curve of an edge has a horizontal or
// vertical extreme farther than one unit from its ends.  The arcs of circles
// keep an extreme within arcMinPiece of their ends inside, as the font does,
// and an edge one within MinPiece of a corner.
func TestExtremesAreNodes(t *testing.T) {
	for _, b := range testStrokes(t) {
		name, r := b.name, b.result
		for _, c := range r.Contours {
			for i, pc := range c {
				// A run-out from a stem on the lattice keeps its handles on whole
				// units rather than a node at the extreme next to the stem.
				if pc.C.IsStraight() || pc.Kind == EndPiece || pc.Fixed {
					continue
				}
				// Next to a corner an extreme may lie within MinPiece.
				lo, hi := 1.0, 1.0
				if prev := c[(i+len(c)-1)%len(c)]; prev.Kind != EdgePiece {
					lo = MinPiece
				}
				if next := c[(i+1)%len(c)]; next.Kind != EdgePiece {
					hi = MinPiece
				}
				l := pc.C.Length()
				for _, u := range pc.C.Extrema() {
					if s := pc.C.lengthTo(u); s > lo && l-s > hi {
						t.Errorf("%s: the curve from %.1f,%.1f has an extreme %.1f units inside", name, pc.C[0].X, pc.C[0].Y, math.Min(s, l-s))
					}
				}
			}
		}
	}
}

// TestNoShortPieces asks that every piece of an edge is at least MinPiece
// long.  The arcs of circles (balls, dots, round ends) may be shorter, as the
// smoothness gate allows.
func TestNoShortPieces(t *testing.T) {
	for _, b := range testStrokes(t) {
		name, r := b.name, b.result
		for _, c := range r.Contours {
			for _, pc := range c {
				if pc.Kind != EdgePiece {
					continue
				}
				if l := pc.C.Length(); l < MinPiece {
					t.Errorf("%s: a piece of %.1f units at %.1f,%.1f", name, l, pc.C[0].X, pc.C[0].Y)
				}
			}
		}
	}
}

// A turnAt is the turn of the rounded outline at a node, the node and the
// shorter of its two handles.
type turnAt struct {
	turn   float64
	at     Point
	handle float64
}

// roundedTurns returns, for each smooth node of the exact path whose two
// handles are at least 8 units long, the angle by which the rounded path turns
// there; the rounded path must have the same number of segments.
func roundedTurns(p outline.Path) ([]turnAt, bool) {
	r := p.Round()
	m := p.Merged()
	var out []turnAt
	for ci, c := range m {
		if len(r[ci].Segs) != len(c.Segs) {
			return nil, false
		}
		for j := range c.Segs {
			if math.Abs(c.NodeTurn(j)) >= outline.SmoothAngle {
				continue
			}
			k := (j + 1) % len(c.Segs)
			in, out2 := c.Segs[j], c.Segs[k]
			if !in.IsCurve || !out2.IsCurve {
				continue
			}
			if dist(in.P2, in.P3) < 8 || dist(out2.P1, in.P3) < 8 {
				continue
			}
			out = append(out, turnAt{math.Abs(r[ci].NodeTurn(j)), in.P3, math.Min(dist(in.P2, in.P3), dist(out2.P1, in.P3))})
		}
	}
	return out, true
}

// TestRoundingKeepsSmoothNodesSmooth perturbs the positions, directions and
// thicknesses of the nodes of every stroke by up to half a unit and a fifth of
// a degree (once in the short tier, five times otherwise), and asks that after the rounding of the font no smooth node with
// handles of at least 8 units turns by more than 2.5 degrees, the bound which
// the font asks of its own glyphs.
func TestRoundingKeepsSmoothNodesSmooth(t *testing.T) {
	trials := 5
	if testing.Short() {
		trials = 1
	}
	for k, b := range testStrokes(t) {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()
			rng := rand.New(rand.NewSource(int64(7 + k)))
			for trial := 0; trial <= trials; trial++ {
				r := b.result
				if trial > 0 {
					q := b.stroke
					q.Nodes = append([]Node(nil), q.Nodes...)
					for i := range q.Nodes {
						q.Nodes[i].P.X += rng.Float64() - 0.5
						q.Nodes[i].P.Y += rng.Float64() - 0.5
						q.Nodes[i].Dir += 0.4 * (rng.Float64() - 0.5)
						if !isAuto(q.Nodes[i].W) && q.Nodes[i].W > 1 {
							q.Nodes[i].W += rng.Float64() - 0.5
						}
					}
					var err error
					r, err = q.Build()
					if err != nil {
						t.Errorf("trial %d: %v", trial, err)
						continue
					}
				}
				turns, ok := roundedTurns(r.Path())
				if !ok {
					t.Errorf("trial %d: the rounding changes the structure", trial)
					continue
				}
				for _, a := range turns {
					if a.turn > 2.5 {
						t.Errorf("trial %d: the smooth node at %.1f,%.1f (handle %.1f) turns by %.2f degrees after rounding", trial, a.at.X, a.at.Y, a.handle, a.turn)
					}
				}
			}
		})
	}
}

// TestShearKeepsThicknessAcross asks that a straight upright stroke sheared
// by the slant of the trial is still as thick at right angles to itself as its
// pen, and leans by the slant.
func TestShearKeepsThicknessAcross(t *testing.T) {
	s := Stroke{Nodes: []Node{
		{P: Point{X: 100, Y: 0}, Dir: 90, W: 81},
		{P: Point{X: 100, Y: 448}, Dir: 90, W: 81},
	}}.Sheared(ItalicSlant)
	r, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	a, b := r.Right[0].C, r.Left[0].C
	d := unit(sub(a[3], a[0]))
	if lean := math.Atan2(d.X, d.Y) * 180 / math.Pi; math.Abs(lean-ItalicSlant) > 1e-9 {
		t.Errorf("the stroke leans by %.3f degrees", lean)
	}
	if w := math.Abs(cross(d, sub(b[0], a[0]))); math.Abs(w-81) > 1e-9 {
		t.Errorf("the stroke is %.3f thick across", w)
	}
}

// TestEnds asks of each kind of end that it closes the outline where it says:
// a cut along a stated line ends both edges on it, a taper ends them in one
// point, a wedge in its tip, and a clip ends them on the line of the other
// stroke with a join line there.
func TestEnds(t *testing.T) {
	base := func(w float64, e End) Stroke {
		return Stroke{Nodes: []Node{
			{P: Point{X: 0, Y: 0}, Dir: 0, W: 60},
			{P: Point{X: 200, Y: 40}, Dir: 20, K: Auto, W: w, DW: Auto},
		}, End: e}
	}
	r, err := base(60, End{Kind: Cut, Line: true, Angle: 90}).Build()
	if err != nil {
		t.Fatal(err)
	}
	if x0, x1 := r.Right.end().X, r.Left.end().X; math.Abs(x0-200) > 1e-6 || math.Abs(x1-200) > 1e-6 {
		t.Errorf("cut: the edges end at x = %.3f and %.3f, not 200", x0, x1)
	}
	r, err = base(0, End{}).Build()
	if err != nil {
		t.Fatal(err)
	}
	if d := dist(r.Right.end(), r.Left.end()); d > 1e-9 {
		t.Errorf("taper: the edges end %.3g apart", d)
	}
	r, err = base(60, End{Kind: Wedge, Length: 50}).Build()
	if err != nil {
		t.Fatal(err)
	}
	c := r.Contours[0]
	tip := add(Point{X: 200, Y: 40}, mul(dirOf(20), 50))
	found := false
	for _, pc := range c {
		if dist(pc.C[3], tip) < 1e-9 {
			found = true
		}
	}
	if !found {
		t.Errorf("wedge: no node at the tip %v", tip)
	}
	r, err = base(60, End{Kind: Clip, Q: Point{X: 180}, Angle: 90}).Build()
	if err != nil {
		t.Fatal(err)
	}
	joins := 0
	for _, pc := range r.Contours[0] {
		if pc.Kind == JoinPiece {
			joins++
			if math.Abs(pc.C[0].X-180) > 1e-6 || math.Abs(pc.C[3].X-180) > 1e-6 {
				t.Errorf("clip: the join line runs from %v to %v, off the line x = 180", pc.C[0], pc.C[3])
			}
		}
	}
	if joins != 1 {
		t.Errorf("clip: %d join lines", joins)
	}
}

// TestBallIsTheCircle asks that the ball of a reproduced part is a true circle
// of the stated radius, that the outer edge runs into it without a turn and
// that the inner edge meets it at the stated opening.
func TestBallIsTheCircle(t *testing.T) {
	p := design.Regular()
	for _, pt := range RomanParts(p) {
		for _, s := range pt.Strokes {
			if s.End.Kind != Ball {
				continue
			}
			b := s.End.Ball
			r, err := s.Build()
			if err != nil {
				t.Fatal(err)
			}
			var arcs []Piece
			for _, pc := range r.Contours[0] {
				if pc.Kind == EndPiece {
					arcs = append(arcs, pc)
				}
			}
			if len(arcs) == 0 {
				t.Fatalf("%s: no ball", pt.Name)
			}
			for _, a := range arcs {
				for i := 0; i <= 8; i++ {
					if e := math.Abs(dist(a.C.At(float64(i)/8), b.Centre) - b.R); e > 0.03 {
						t.Errorf("%s: the ball is off its circle by %.3f", pt.Name, e)
					}
				}
			}
			c := r.Contours[0]
			for i, pc := range c {
				next := c[(i+1)%len(c)]
				if pc.Kind == EndPiece && next.Kind != EndPiece || pc.Kind != EndPiece && next.Kind == EndPiece {
					turn := math.Abs(angleBetween(pc.C.Tangent(1), next.C.Tangent(0)))
					if turn > 0.01 && math.Abs(turn-(180-b.Open)) > 0.01 {
						t.Errorf("%s: the ball meets an edge at %.2f degrees", pt.Name, turn)
					}
				}
			}
		}
	}
}

// TestItalicLettersJoin builds every letter of the italic trial as the union
// of its strokes and asks for one clean outline: no end which another stroke
// should cover shows, no sliver or needle corner is left where edges cross,
// and the rounded outline does not cross itself.
func TestItalicLettersJoin(t *testing.T) {
	p := design.Regular()
	for name, l := range italicLetters(p) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var sl Letter
			for _, s := range l.Strokes {
				sl.Strokes = append(sl.Strokes, s.ShearedRatio(SlantNum, SlantDen))
			}
			sl.Extra = l.Extra
			r, err := sl.Build()
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range r.Warnings {
				if strings.Contains(w, "hidden end") || strings.Contains(w, "sliver") || strings.Contains(w, "a corner of") {
					t.Error(w)
				}
			}
			if x := r.Path().Round().SelfIntersections(); len(x) > 0 {
				t.Errorf("the rounded outline crosses itself at %v", x[0])
			}
		})
	}
}
