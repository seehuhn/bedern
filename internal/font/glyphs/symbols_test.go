// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// symbolGlyphs are the question mark, the ampersand and the symbols of the
// basic set, which are drawn as strokes (symstroke.go) and unions of parts
// (symunion.go).
var symbolGlyphs = []string{"question", "braceleft", "braceright", "asciitilde", "asciicircum", "grave",
	"degree", "percent", "numbersign", "dollar", "sterling", "Euro", "section", "at", "ampersand"}

// The nodes of the symbols are of two kinds: smooth nodes, where the outline
// goes on in the same direction, and corners, where two edges meet on purpose (a
// cut, the point of a brace, a crossing of two strokes, the tip of a wedge).  A
// node which turns by more than smoothTurn and less than cornerTurn is neither:
// a kink.  At a smooth node between two curves the curvature does not jump: where
// both sides bend the same way by more than a radius of 2000 units, the ratio of
// the curvatures is within curvatureJump; and the edge does not change the
// direction of its turning at a node unless it is almost straight there (a
// radius of more than flatRadius units on one side).
const (
	smoothTurn    = 0.5  // degrees
	cornerTurn    = 12.0 // degrees
	curvatureJump = 2.5
	flatRadius    = 600.0
)

// nodeTurns calls f for every node of the contour with the turn there (in
// degrees, positive to the left) and the segments before and after it.
func nodeTurns(ct outline.Contour, f func(at outline.Point, turn float64, before, after cubic, curves bool)) {
	n := len(ct.Segs)
	for i := 0; i < n; i++ {
		prev, next := segCubic(ct, i+n-1), segCubic(ct, i)
		tin, tout := prev.tangent(1), next.tangent(0)
		if !ct.Segs[(i+n-1)%n].IsCurve {
			tin = unit(prev.P0, prev.P3)
		}
		if !ct.Segs[i].IsCurve {
			tout = unit(next.P0, next.P3)
		}
		turn := math.Atan2(tin.X*tout.Y-tin.Y*tout.X, tin.X*tout.X+tin.Y*tout.Y) * 180 / math.Pi
		f(prev.P3, turn, prev, next, ct.Segs[(i+n-1)%n].IsCurve && ct.Segs[i].IsCurve)
	}
}

func TestSymbolNodes(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		if !boldApplies(t, np) {
			continue
		}
		for _, name := range symbolGlyphs {
			path, _, _ := outlineOf(name, np.params)
			for _, ct := range path {
				nodeTurns(ct, func(at outline.Point, turn float64, before, after cubic, curves bool) {
					a := math.Abs(turn)
					if a > smoothTurn && a < cornerTurn {
						t.Errorf("%s, %s: a kink of %.1f degrees at (%.1f, %.1f)", np.name, name, turn, at.X, at.Y)
						return
					}
					if a > smoothTurn || !curves {
						return
					}
					k0, k1 := before.curvature(1), after.curvature(0)
					if math.Abs(k0) < 1/flatRadius || math.Abs(k1) < 1/flatRadius {
						return
					}
					if k0*k1 < 0 {
						t.Errorf("%s, %s: the edge turns the other way at (%.1f, %.1f): radii %.0f and %.0f", np.name, name, at.X, at.Y, 1/k0, 1/k1)
						return
					}
					if r := k1 / k0; r > curvatureJump || r < 1/curvatureJump {
						if math.Abs(k0) > 1/2000.0 && math.Abs(k1) > 1/2000.0 {
							t.Errorf("%s, %s: the curvature jumps by %.2f at (%.1f, %.1f): radii %.0f and %.0f", np.name, name, r, at.X, at.Y, 1/k0, 1/k1)
						}
					}
				})
			}
		}
	}
}

// contourBox returns the bounding box of one contour of the path.
func contourBox(ct outline.Contour) (x0, y0, x1, y1 float64) {
	return outline.Path{ct}.BBox()
}

// The dot of the question mark stands under the foot of the hook, and the two
// stand left of the middle of the ink, between qBalanceLow and qBalanceHigh of
// its width.  The white between the foot and the dot is at least half the dot.
const (
	qBalanceLow  = 0.35
	qBalanceHigh = 0.50
)

func TestQuestionBalance(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		if !boldApplies(t, np) {
			continue
		}
		path, _, _ := outlineOf("question", np.params)
		if len(path) != 2 {
			t.Fatalf("%s: %d contours, want the hook and the dot", np.name, len(path))
		}
		hx0, hy0, hx1, _ := contourBox(path[0])
		dx0, _, dx1, dy1 := contourBox(path[1])
		dot := (dx0 + dx1) / 2
		xs := outline.Path{path[0]}.CrossingsX(hy0 + 1)
		if len(xs) != 2 {
			t.Fatalf("%s: the foot of the hook has %d crossings", np.name, len(xs))
		}
		if foot := (xs[0] + xs[1]) / 2; math.Abs(foot-dot) > 1 {
			t.Errorf("%s: the dot is centred at %.1f, the foot at %.1f", np.name, dot, foot)
		}
		x0, x1 := math.Min(hx0, dx0), math.Max(hx1, dx1)
		if u := (dot - x0) / (x1 - x0); u < qBalanceLow || u > qBalanceHigh {
			t.Errorf("%s: the dot stands at %.2f of the ink, want %v to %v", np.name, u, qBalanceLow, qBalanceHigh)
		}
		if gap := hy0 - dy1; gap < np.params.DotDiameter/2 {
			t.Errorf("%s: the white between the hook and the dot is %.1f", np.name, gap)
		}
	}
}

// The round parts of the percent and the euro overshoot the figure height and
// the baseline as the round letters do, by between ovLow and ovHigh units.
const (
	ovLow  = 5.0
	ovHigh = 25.0
)

func TestSymbolOvershoot(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		for _, name := range []string{"percent", "Euro"} {
			path, _, _ := outlineOf(name, np.params)
			_, y0, _, y1 := path.BBox()
			fh := np.params.Figures.Height
			if o := y1 - fh; o < ovLow || o > ovHigh {
				t.Errorf("%s, %s: the top overshoots the figure height by %.1f", np.name, name, o)
			}
			if o := -y0; o < ovLow || o > ovHigh {
				t.Errorf("%s, %s: the bottom overshoots the baseline by %.1f", np.name, name, o)
			}
		}
	}
}

// The bars of the euro are centred on the round: the middle between them lies
// within euroBarTolerance of the middle of the height of the ink.
const euroBarTolerance = 8.0

func TestEuroBarsCentred(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		path, _, _ := outlineOf("Euro", np.params)
		x0, y0, _, y1 := path.BBox()
		ys := path.CrossingsY(x0 + 2)
		if len(ys) != 4 {
			t.Fatalf("%s: %d crossings at the left ends of the bars, want 4", np.name, len(ys))
		}
		if d := (ys[0]+ys[3])/2 - (y0+y1)/2; math.Abs(d) > euroBarTolerance {
			t.Errorf("%s: the bars are centred %.1f from the middle of the ink", np.name, d)
		}
	}
}

// The euro is the capital C scaled, with two bars.  Measured on the outlines,
// the round of the euro (from the left extreme of its outer edge, between the
// bars, to the tip of the tail) has the proportion of the ink of C, width over
// height, times euroRoundFactor within euroRoundBand, and the whole sign that
// of C times a factor between euroSignLow and euroSignHigh.  In the frame
// scaled by the widths and the heights of the two, the cut of the beak stands
// at the place of the cut of C within euroCutBand of the width, and the top and
// the bottom of the cut and the tip of the tail are those of C within
// euroFeatureTolerance units.
const (
	euroRoundFactor      = 0.78
	euroRoundBand        = 0.02
	euroSignLow          = 0.84
	euroSignHigh         = 0.90
	euroCutBand          = 0.01
	euroFeatureTolerance = 4.0
)

// A cShape holds the features of a C measured on its outline: the left
// extreme of the round, the tip of the tail (right), the bottom and the top,
// the vertical cut of the beak, the top and the bottom of the cut and the
// height of the tip of the tail.
type cShape struct {
	left, right, bottom, top  float64
	cut, cutTop, cutEnd, tipY float64
}

// measureC measures the C-shaped outline path whose round has its left
// extreme at left.
func measureC(t *testing.T, name string, path outline.Path, left float64) cShape {
	_, y0, x1, y1 := path.BBox()
	s := cShape{left: left, right: x1, bottom: y0, top: y1, cut: math.Inf(-1)}
	mid := (y0 + y1) / 2
	for y := mid + 120; y < y1-5; y += 8 {
		if xs := path.CrossingsX(y); len(xs) > 0 {
			s.cut = math.Max(s.cut, xs[len(xs)-1])
		}
	}
	var up []float64
	for _, y := range path.CrossingsY(s.cut - 1.5) {
		if y > mid {
			up = append(up, y)
		}
	}
	ys := path.CrossingsY(x1 - 1.5)
	if len(up) < 2 || len(ys) < 2 {
		t.Fatalf("%s: no cut of the beak or no tip of the tail", name)
	}
	s.cutEnd, s.cutTop = up[0], up[len(up)-1]
	s.tipY = (ys[0] + ys[1]) / 2
	return s
}

func TestEuroIsCScaled(t *testing.T) {
	t.Parallel()
	for _, np := range derivedParams() {
		cp, _, _ := outlineOf("C", np.params)
		cx0, _, _, _ := cp.BBox()
		c := measureC(t, np.name+", C", cp, cx0)

		ep, _, _ := outlineOf("Euro", np.params)
		ex0, ey0, ex1, ey1 := ep.BBox()
		mid := np.params.Figures.Height / 2
		left := math.Inf(1)
		for y := mid - 20; y <= mid+20; y += 4 {
			if xs := ep.CrossingsX(y); len(xs) > 0 {
				left = math.Min(left, xs[0])
			}
		}
		e := measureC(t, np.name+", Euro", ep, left)

		sx := (e.right - e.left) / (c.right - c.left)
		sy := (e.top - e.bottom) / (c.top - c.bottom)
		if f := sx / sy; math.Abs(f-euroRoundFactor) > euroRoundBand {
			t.Errorf("%s: the round of the euro has %.3f of the proportion of C", np.name, f)
		}
		sign := ((ex1 - ex0) / (ey1 - ey0)) / ((c.right - c.left) / (c.top - c.bottom))
		if sign < euroSignLow || sign > euroSignHigh {
			t.Errorf("%s: the euro has %.3f of the proportion of C", np.name, sign)
		}
		if d := (e.cut-e.left)/(e.right-e.left) - (c.cut-c.left)/(c.right-c.left); math.Abs(d) > euroCutBand {
			t.Errorf("%s: the cut of the beak of the euro stands %.3f of the width from that of C", np.name, d)
		}
		for _, f := range []struct {
			what string
			e, c float64
		}{
			{"the top of the cut below the top", (e.top - e.cutTop) / sy, c.top - c.cutTop},
			{"the end of the beak below the top", (e.top - e.cutEnd) / sy, c.top - c.cutEnd},
			{"the tip of the tail above the bottom", (e.tipY - e.bottom) / sy, c.tipY - c.bottom},
		} {
			if d := f.e - f.c; math.Abs(d) > euroFeatureTolerance {
				t.Errorf("%s: %s of the euro, scaled to C, is %.1f from that of C", np.name, f.what, d)
			}
		}
	}
}

// The diagonal of the percent keeps clear of the two zeros by at least
// pctClearance.
const pctClearance = 30.0

func TestPercentClearance(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("percent", np.params)
		if len(path) != 5 {
			t.Fatalf("%s: %d contours, want 5", np.name, len(path))
		}
		slash := path[2]
		for _, zero := range []outline.Contour{path[0], path[3]} {
			d, a, _ := outline.Path{slash, zero}.MinDistance(clearanceRatio)
			if d < pctClearance {
				t.Errorf("%s: the diagonal comes within %.1f of a zero at (%.1f, %.1f)", np.name, d, a.X, a.Y)
			}
		}
	}
}

// The point of a brace lies at the middle of its height, and the brace reaches
// the top and the bottom of the parentheses.
func TestBracePoint(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		for _, name := range []string{"braceleft", "braceright"} {
			path, adv, _ := outlineOf(name, np.params)
			x0, y0, x1, y1 := path.BBox()
			if math.Abs(y1-np.params.BracketTop()) > 0.5 || math.Abs(y0-np.params.BracketBottom()) > 0.5 {
				t.Errorf("%s, %s: the brace runs from %.1f to %.1f", np.name, name, y0, y1)
			}
			tipX := x0
			if name == "braceright" {
				tipX = x1
			}
			ys := path.CrossingsY(tipX + (adv/2-tipX)*0.02)
			if len(ys) != 2 {
				t.Fatalf("%s, %s: %d crossings next to the point", np.name, name, len(ys))
			}
			if m := (ys[0] + ys[1]) / 2; math.Abs(m-(y0+y1)/2) > 0.5 {
				t.Errorf("%s, %s: the point is at %.1f, the middle at %.1f", np.name, name, m, (y0+y1)/2)
			}
		}
	}
}
