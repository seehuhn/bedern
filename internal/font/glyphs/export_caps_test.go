// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// samePath reports whether two paths have the same nodes and handles within
// a thousandth of a unit.
func samePath(a, b outline.Path) bool {
	if len(a) != len(b) {
		return false
	}
	near := func(p, q outline.Point) bool { return (p.X-q.X)*(p.X-q.X)+(p.Y-q.Y)*(p.Y-q.Y) < 1e-6 }
	for i := range a {
		if !near(a[i].Start, b[i].Start) || len(a[i].Segs) != len(b[i].Segs) {
			return false
		}
		for j, s := range a[i].Segs {
			t := b[i].Segs[j]
			if s.IsCurve != t.IsCurve || !near(s.P3, t.P3) || (s.IsCurve && (!near(s.P1, t.P1) || !near(s.P2, t.P2))) {
				return false
			}
		}
	}
	return true
}

// TestCapStemOutlineIsI requires the lone stem with four serifs of the full
// height to be the outline of I.
func TestCapStemOutlineIsI(t *testing.T) {
	p := design.Regular()
	j := p.Caps.SerifJut
	got := CapStemOutline(p, design.CapStemLeft(p), p.CapHeight, j, j, j, j)
	want, _ := capI(p)
	if !samePath(got, want) {
		t.Errorf("the stem with four serifs differs from I")
	}
}

// TestCapPartsAreSingleContours requires a short stem with feet only, an
// arm of each kind and a super-ellipse to be one counter-clockwise contour
// each.
func TestCapPartsAreSingleContours(t *testing.T) {
	p := design.Regular()
	c := p.Caps
	j := c.SerifJut
	left := design.CapStemLeft(p)
	r := left + c.Stem
	parts := []struct {
		name string
		path outline.Path
	}{
		{"short stem", CapStemOutline(p, left, 200, j, j, 0, 0)},
		{"top arm", CapArmOutline(p, r, r+300, p.CapHeight-c.Bar, p.CapHeight, c.TopDrop, 0)},
		{"bottom arm", CapArmOutline(p, r, r+300, 0, c.Bar, 0, c.BottomRise)},
		{"middle arm", CapArmOutline(p, r, r+200, 300, 300+c.Bar, c.MiddleReach, c.MiddleReach)},
		{"plain arm", CapArmOutline(p, r, r+200, 300, 300+c.Bar, 0, 0)},
		{"round", SuperEllipseOutline(c.SideRoundLeft, 0, c.SideRoundLeft+300, 400, p.SupernessOuter, true)},
	}
	for _, pt := range parts {
		if len(pt.path) != 1 {
			t.Errorf("%s: %d contours, want 1", pt.name, len(pt.path))
			continue
		}
		if pt.path[0].SignedArea() <= 0 {
			t.Errorf("%s: the contour does not run counter-clockwise", pt.name)
		}
	}
}
