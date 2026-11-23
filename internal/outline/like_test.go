// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package outline

import (
	"math"
	"testing"
)

// oval returns a counter-clockwise oval of four curves about the origin
// with the half axes rx and ry, which starts at its right extreme.
func oval(rx, ry float64) Contour {
	const k = 0.5523
	var p Path
	p.MoveTo(rx, 0)
	p.CurveTo(rx, k*ry, k*rx, ry, 0, ry)
	p.CurveTo(-k*rx, ry, -rx, k*ry, -rx, 0)
	p.CurveTo(-rx, -k*ry, -k*rx, -ry, 0, -ry)
	p.CurveTo(k*rx, -ry, rx, -k*ry, rx, 0)
	return p[0]
}

// cutAt returns the contour with segment j cut at the parameter t.
func cutAt(c Contour, j int, t float64) Contour {
	l, r := splitCubic(asCubic(c.Segs[j], c.startOf(j)), t)
	out := Contour{Start: c.Start}
	out.Segs = append(out.Segs, c.Segs[:j]...)
	out.Segs = append(out.Segs,
		Segment{IsCurve: true, P1: l[1], P2: l[2], P3: l[3]},
		Segment{IsCurve: true, P1: r[1], P2: r[2], P3: r[3]})
	out.Segs = append(out.Segs, c.Segs[j+1:]...)
	return out
}

// startAt returns the closed contour with its start at the end of segment j.
func startAt(c Contour, j int) Contour {
	out := Contour{Start: c.Segs[j].P3}
	out.Segs = append(out.Segs, c.Segs[j+1:]...)
	out.Segs = append(out.Segs, c.Segs[:j+1]...)
	return out
}

// apart returns the largest distance from a point of the flattened path a to
// the flattened path b.  The flattening of a curve of some hundred units
// lies up to 0.03 units from the curve.
func apart(a, b Path) float64 {
	var worst float64
	for i := range a {
		pb := b[i].Flatten()
		for _, q := range a[i].Flatten() {
			best := math.Inf(1)
			for k := range pb {
				c := closest(pb[k], pb[(k+1)%len(pb)], q)
				best = math.Min(best, math.Hypot(c.X-q.X, c.Y-q.Y))
			}
			worst = math.Max(worst, best)
		}
	}
	return worst
}

// sameNodes fails the test unless got has the segments of want: as many, of
// the same kinds, closed in the same way.
func sameNodes(t *testing.T, got, want Path) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%d contours, want %d", len(got), len(want))
	}
	for i := range got {
		if !sameSegments(got[i], want[i]) {
			t.Errorf("contour %d has %d segments, want the %d of the model with their kinds", i, len(got[i].Segs), len(want[i].Segs))
		}
	}
}

// A node which the model lacks is removed and a node which the path
// lacks is added, and the shape stays.
func TestLikeTakesTheNodesOfTheModel(t *testing.T) {
	ref := Path{oval(300, 200)}
	wide := oval(360, 210)
	for _, tc := range []struct {
		name string
		p    Path
	}{
		{"one node more", Path{cutAt(wide, 1, 0.4)}},
		{"two nodes more", Path{cutAt(cutAt(wide, 3, 0.7), 0, 0.3)}},
		{"another start", Path{startAt(cutAt(wide, 2, 0.5), 3)}},
	} {
		got, err := tc.p.Like(ref)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		sameNodes(t, got, ref)
		if d := math.Max(apart(got, tc.p), apart(tc.p, got)); d > 0.1 {
			t.Errorf("%s: the shape moves by %.3f units", tc.name, d)
		}
		if got[0].Start != (Point{360, 0}) {
			t.Errorf("%s: the contour starts at %v, want the right extreme, where the model starts", tc.name, got[0].Start)
		}
	}

	// The model has a node more than the path.
	more := Path{cutAt(oval(300, 200), 2, 0.6)}
	p := Path{wide}
	got, err := p.Like(more)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, got, more)
	if d := math.Max(apart(got, p), apart(p, got)); d > 0.1 {
		t.Errorf("a node added: the shape moves by %.3f units", d)
	}
	for j := range got[0].Segs {
		if a := math.Abs(got[0].NodeTurn(j)); a > 0.01 {
			t.Errorf("a node added: the outline turns by %.3f degrees at node %d", a, j)
		}
	}
}

// A line stays a line and takes the place of a curve as a straight curve; a
// corner which the model lacks is an error, and so is a contour which
// runs the other way round.
func TestLikeKeepsCorners(t *testing.T) {
	var box, cut, house Path
	box.MoveTo(0, 0)
	box.LineTo(400, 0)
	box.LineTo(400, 300)
	box.LineTo(0, 300)
	// The same with the top edge in two parts.
	cut.MoveTo(0, 0)
	cut.LineTo(500, 0)
	cut.LineTo(500, 300)
	cut.LineTo(200, 300)
	cut.LineTo(0, 300)
	// A corner in the top edge.
	house.MoveTo(0, 0)
	house.LineTo(500, 0)
	house.LineTo(500, 300)
	house.LineTo(200, 380)
	house.LineTo(0, 300)

	got, err := cut.Like(box)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, got, box)
	if d := apart(cut, got); d > 1e-9 {
		t.Errorf("the shape moves by %g units", d)
	}
	got, err = box.Like(cut)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, got, cut)
	if _, err := house.Like(box); err == nil {
		t.Error("a corner which the model lacks is removed")
	}

	// A curve of the model where the path has a line.
	var bent Path
	bent.MoveTo(0, 0)
	bent.LineTo(400, 0)
	bent.LineTo(400, 300)
	bent.CurveTo(300, 310, 100, 310, 0, 300)
	got, err = box.Like(bent)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, got, bent)
	if d := apart(box, got); d > 1e-9 {
		t.Errorf("the straight curve moves the shape by %g units", d)
	}
	if _, err := bent.Like(box); err == nil {
		t.Error("a curve takes the place of a line of the model")
	}

	back := Path{Contour{Start: box[0].Start, Segs: []Segment{{P3: Point{0, 300}}, {P3: Point{400, 300}}, {P3: Point{400, 0}}}}}
	if _, err := back.Like(box); err == nil {
		t.Error("a contour which runs the other way round passes")
	}
}

// RoundLike follows the merging of the short segments of the model: a
// short segment which Round merges in the model is merged in the path
// although it is longer there, and the path against itself is rounded as
// Round rounds it.
func TestRoundLikeMergesWhatTheModelMerges(t *testing.T) {
	short := Path{cutAt(cutAt(oval(300, 200), 0, 0.5), 1, 0.06)}
	long := Path{cutAt(cutAt(oval(600, 400), 0, 0.5), 1, 0.06)}
	want := short.Round()
	if len(want[0].Segs) != 5 {
		t.Fatalf("Round leaves %d segments, want 5: the short segment merged", len(want[0].Segs))
	}
	if n := len(long.Round()[0].Segs); n != 6 {
		t.Fatalf("Round leaves the larger oval %d segments, want 6: its segment is not short", n)
	}
	got, err := long.RoundLike(short)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, got, want)
	if d := math.Max(apart(got, long), apart(long, got)); d > MergeTolerance+0.75 {
		t.Errorf("the rounded outline lies %.2f units from the path", d)
	}

	self, err := short.RoundLike(short)
	if err != nil {
		t.Fatal(err)
	}
	sameNodes(t, self, want)
	for j, s := range self[0].Segs {
		if s != want[0].Segs[j] {
			t.Errorf("segment %d of the path rounded like itself is %v, Round gives %v", j, s, want[0].Segs[j])
		}
	}
}

// RoundNodes keeps every segment, where Round merges the short ones, and
// rounds what Round leaves as Round does.
func TestRoundNodesKeepsTheSegments(t *testing.T) {
	short := Path{cutAt(cutAt(oval(300.3, 200.4), 0, 0.5), 1, 0.06)}
	got := short.RoundNodes()
	if len(got[0].Segs) != len(short[0].Segs) {
		t.Fatalf("RoundNodes leaves %d segments of %d", len(got[0].Segs), len(short[0].Segs))
	}
	for j, s := range got[0].Segs {
		for _, q := range []Point{s.P1, s.P2, s.P3} {
			if q.X != math.Round(q.X) || q.Y != math.Round(q.Y) {
				t.Errorf("segment %d has the point %v, which is not whole", j, q)
			}
		}
		if a := math.Abs(got[0].NodeTurn(j)); a >= turnTie+SmoothAngle {
			t.Errorf("the rounded outline turns by %.2f degrees at node %d, which is smooth", a, j)
		}
	}
	plain := Path{oval(300.3, 200.4)}
	want := plain.Round()
	for j, s := range plain.RoundNodes()[0].Segs {
		if s != want[0].Segs[j] {
			t.Errorf("segment %d of an outline without short segments is %v, Round gives %v", j, s, want[0].Segs[j])
		}
	}
}
