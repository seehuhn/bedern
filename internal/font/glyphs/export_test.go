// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// TestUnionOfCrossingBars requires the union of two bars which cross to be one
// counter-clockwise contour which covers the area of both less their overlap.
func TestUnionOfCrossingBars(t *testing.T) {
	var a, b outline.Path
	rect(&a, 0, 40, 200, 60, true)
	rect(&b, 90, 0, 110, 100, true)
	u := Union(a, b)
	if len(u) != 1 {
		t.Fatalf("the union has %d contours, want 1", len(u))
	}
	want := 200.0*20 + 20*100 - 20*20
	if got := u[0].SignedArea(); math.Abs(got-want) > 1e-6 {
		t.Errorf("the union covers %v, want %v", got, want)
	}
}

// TestCheckOutline requires the checks of an outline of the table to pass as
// those of the glyph do, and an outline which breaks the bounding-box rule of
// its parent to fail.
func TestCheckOutline(t *testing.T) {
	p := design.Regular()
	path, adv, _ := Outline("plus", p)
	if err := CheckOutline(p, "plus", path, adv); err != nil {
		t.Errorf("the plus fails: %v", err)
	}
	var moved outline.Path
	for _, c := range path {
		n := outline.Contour{Start: outline.Point{X: c.Start.X - 50, Y: c.Start.Y}}
		for _, s := range c.Segs {
			n.Segs = append(n.Segs, outline.Segment{IsCurve: s.IsCurve,
				P1: outline.Point{X: s.P1.X - 50, Y: s.P1.Y}, P2: outline.Point{X: s.P2.X - 50, Y: s.P2.Y}, P3: outline.Point{X: s.P3.X - 50, Y: s.P3.Y}})
		}
		moved = append(moved, n)
	}
	if err := CheckOutline(p, "plus", moved, adv); err == nil {
		t.Error("the plus moved out of its side space passes")
	}
	if err := CheckOutline(p, "no such glyph", path, adv); err == nil {
		t.Error("an unknown parent passes")
	}
}
