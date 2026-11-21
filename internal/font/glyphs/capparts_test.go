// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The parts of the capitals, tested with made-up values: the four serifs of a
// stem are one shape, the arms are connected chains whose serif ends in a
// vertical cut, and mergeLines leaves no node in a straight edge.

// chainArea returns the area which the chain encloses together with the straight
// lines from its end to the corner and from there back to its start.
func chainArea(c chain, corner outline.Point) float64 {
	var path outline.Path
	path.MoveTo(c.start().X, c.start().Y)
	c.appendTo(&path)
	path.LineTo(corner.X, corner.Y)
	return path[0].SignedArea()
}

func TestCapSerifsAreOneShape(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	p.Caps.SerifJut, p.Caps.BracketReach, p.Caps.BracketHeight, p.Caps.SerifThick = 80, 44, 58, 30
	const x, top = 300.0, 640.0
	// The ink of each serif beyond the stem edge: the chain closed by the edge
	// and by the line along the baseline or the cap height, which meet at the
	// corner (x, 0) or (x, top).
	feet := []chain{capFootRight(p, x, 80), capFootLeft(p, x, 80)}
	heads := []chain{capHeadRight(p, x, top, 80), capHeadLeft(p, x, top, 80)}
	corners := []outline.Point{{X: x}, {X: x}, {X: x, Y: top}, {X: x, Y: top}}
	want := math.Abs(chainArea(feet[0], corners[0]))
	for i, c := range append(feet, heads...) {
		if err := c.connected(); err != nil {
			t.Fatal(err)
		}
		if a := math.Abs(chainArea(c, corners[i])); math.Abs(a-want) > 1e-6 {
			t.Errorf("serif %d has the area %.3f, the right foot serif %.3f", i, a, want)
		}
	}
	// The ends: on the stem edge at the bracket height, and on the baseline or the
	// cap height at the tip or at the edge.
	for _, c := range []struct {
		name       string
		got, other outline.Point
		want       outline.Point
	}{
		{"right foot", feet[0].end(), feet[0].start(), outline.Point{X: x, Y: 58}},
		{"left foot", feet[1].start(), feet[1].end(), outline.Point{X: x, Y: 58}},
		{"right head", heads[0].start(), heads[0].end(), outline.Point{X: x, Y: top - 58}},
		{"left head", heads[1].end(), heads[1].start(), outline.Point{X: x, Y: top - 58}},
	} {
		if math.Hypot(c.got.X-c.want.X, c.got.Y-c.want.Y) > 1e-9 {
			t.Errorf("%s serif meets the stem at %v, want %v", c.name, c.got, c.want)
		}
		if c.other.Y != 0 && c.other.Y != top {
			t.Errorf("%s serif ends at %v, not on the baseline or the cap height", c.name, c.other)
		}
	}
}

func TestCapArmsEndInAVerticalCut(t *testing.T) {
	t.Parallel()
	a := capArm{Edge: 200, End: 520, Lower: 300, Upper: 336, Drop: 110, Rise: 90, SerifThick: 38}
	for name, c := range map[string]chain{
		"top": capArmTop(a), "bottom": capArmBottom(a), "middle": capArmMiddle(a),
		"top left": capArmTopLeft(capArm{Edge: 200, End: -100, Lower: 300, Upper: 336, Drop: 110, SerifThick: 38}),
	} {
		if err := c.connected(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		// The cut is one vertical line at End, the length of the serif and the bar.
		end, length := 520.0, 0.0
		want := 36 + 110.0
		switch name {
		case "bottom":
			want = 36 + 90
		case "middle":
			want = 36 + 110 + 90
		case "top left":
			end = -100
		}
		for _, q := range c {
			if q.line && q.P0.X == end && q.P3.X == end {
				length += math.Abs(q.P3.Y - q.P0.Y)
			}
		}
		if math.Abs(length-want) > 1e-9 {
			t.Errorf("%s: the cut at %v is %.2f long, want %.2f", name, end, length, want)
		}
	}
}

func TestMergeLines(t *testing.T) {
	t.Parallel()
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	c := chain{
		lineSeg(pt(50, 0), pt(100, 0)),
		lineSeg(pt(100, 0), pt(100, 0)),
		lineSeg(pt(100, 0), pt(100, 100)),
		lineSeg(pt(100, 100), pt(0, 100)),
		lineSeg(pt(0, 100), pt(0, 0)),
		lineSeg(pt(0, 0), pt(50, 0)),
	}
	m := mergeLines(c)
	if len(m) != 4 {
		t.Fatalf("got %d pieces, want 4: %v", len(m), m)
	}
	if m.start() != pt(0, 0) || m.end() != pt(0, 0) {
		t.Errorf("the merged contour runs from %v to %v, want from and to (0, 0)", m.start(), m.end())
	}
}
