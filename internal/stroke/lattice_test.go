// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"
	"strings"
	"testing"
)

// TestFoldIsRefused asks that a turn too tight for the pen is refused with the
// radius the pen needs, instead of an inner edge which folds back.
func TestFoldIsRefused(t *testing.T) {
	s := Stroke{Nodes: []Node{
		{P: Point{X: 0, Y: 100}, Dir: 270, W: 81},
		{P: Point{X: 30, Y: 70}, Dir: 0, K: Auto, W: 60},
	}}
	_, err := s.Build()
	if err == nil || !strings.Contains(err.Error(), "needs a radius of at least") {
		t.Errorf("got %v, want the refusal of a fold", err)
	}
}

// TestSlantedStemsStayOnTheirLattice asks of every stroke of the italic trial
// that after the rounding of the font each handle of a curve which leaves or
// enters a straight edge on the slant lies exactly on the line of that edge,
// num (x1 - x0) = den (y1 - y0) in whole units, so that the edge leaves the
// stem in its direction.
func TestSlantedStemsStayOnTheirLattice(t *testing.T) {
	for _, b := range testStrokes(t) {
		if b.stroke.Slant[1] == 0 {
			continue
		}
		num, den := float64(b.stroke.Slant[0]), float64(b.stroke.Slant[1])
		onSlant := func(a, c Point) bool { return den*(c.X-a.X) == num*(c.Y-a.Y) }
		for _, c := range b.result.Path().Round() {
			n := len(c.Segs)
			start := func(j int) Point {
				if j == 0 {
					return c.Start
				}
				return c.Segs[j-1].P3
			}
			for j, s := range c.Segs {
				p0 := start(j)
				if s.IsCurve || p0 == s.P3 || !onSlant(p0, s.P3) {
					continue
				}
				// A line on the slant: the curves before and after it, where
				// they join it smoothly.
				if prev := c.Segs[(j+n-1)%n]; prev.IsCurve && math.Abs(angleBetween(sub(p0, prev.P2), sub(s.P3, p0))) < 2 {
					if !onSlant(prev.P2, p0) {
						t.Errorf("%s: the handle %v of the curve into the stem at %v is off the line of the stem", b.name, prev.P2, p0)
					}
				}
				if next := c.Segs[(j+1)%n]; next.IsCurve && math.Abs(angleBetween(sub(s.P3, p0), sub(next.P1, s.P3))) < 2 {
					if !onSlant(s.P3, next.P1) {
						t.Errorf("%s: the handle %v of the curve out of the stem at %v is off the line of the stem", b.name, next.P1, s.P3)
					}
				}
			}
		}
	}
}
