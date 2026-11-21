// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the beaks of C, G and S, of the balance of C and of the spur and
// the stem of G measure the outlines in every variant and perturbed parameter
// set.

// mapPath returns the path with every point mapped by f.
func mapPath(path outline.Path, f func(outline.Point) outline.Point) outline.Path {
	out := make(outline.Path, len(path))
	for i, ct := range path {
		c := outline.Contour{Start: f(ct.Start)}
		for _, sg := range ct.Segs {
			c.Segs = append(c.Segs, outline.Segment{IsCurve: sg.IsCurve, P1: f(sg.P1), P2: f(sg.P2), P3: f(sg.P3)})
		}
		out[i] = c
	}
	return out
}

// A hangingSerif holds the measurements of a serif which hangs down from a bar
// or an arc on the right, with its vertical cut at the right: the height of its
// end, its width halfway along the straight part of its inner face, the length
// of that straight part from the end, and the height of the bracket above it,
// up to where the ink is twice as wide as the serif.
type hangingSerif struct {
	End, Thick, Face, Bracket float64
}

// measureHangingSerif measures the serif whose cut is at x, in the path turned so
// that the serif hangs at the top on the right.  The straight part of the face
// ends where the serif is more than 2 units wider than the line of the face
// (design.ZBeakTilt from the vertical) through its width 20 units above the end.
func measureHangingSerif(t *testing.T, path outline.Path, x float64) hangingSerif {
	t.Helper()
	v := vruns(path, x-3)
	if len(v) == 0 {
		t.Fatalf("no ink inside the cut at %.1f", x)
	}
	var m hangingSerif
	m.End = v[len(v)-1][0]
	width := func(y float64) float64 {
		r := runs(path, y)
		for i := len(r) - 1; i >= 0; i-- {
			if math.Abs(r[i][1]-x) < 1 {
				return r[i][1] - r[i][0]
			}
		}
		return math.NaN()
	}
	w20 := width(m.End + 20)
	tilt := math.Tan(design.ZBeakTilt)
	s := 21.0
	for ; s < 400; s++ {
		if !(width(m.End+s) <= w20+(s-20)*tilt+2) {
			break
		}
	}
	m.Face = s
	m.Thick = width(m.End + s/2)
	for ; s < 600; s++ {
		if !(width(m.End+s) < 2*m.Thick) {
			break
		}
	}
	m.Bracket = s - m.Face
	return m
}

// capBeakSerifs returns the serifs of the beaks of C, G and S, each measured in
// the frame in which it hangs at the top on the right, and the top serif of E.
func capBeakSerifs(t *testing.T, np namedParams) (e hangingSerif, beaks map[string]hangingSerif) {
	t.Helper()
	p := np.params
	cap := capHeight(np)
	beaks = map[string]hangingSerif{}
	ep := capOutline(np, "E")
	ex := math.Inf(-1)
	for y := 0.9 * cap; y < cap; y++ {
		if r := runs(ep, y); len(r) > 0 {
			ex = math.Max(ex, r[len(r)-1][1])
		}
	}
	e = measureHangingSerif(t, ep, ex)
	for _, name := range []string{"C", "G"} {
		path := capOutline(np, name)
		x, _, _ := cBeak(t, path, cap)
		beaks[name] = measureHangingSerif(t, path, x)
	}
	f := newCapSFrame(p)
	sp := capOutline(np, "S")
	beaks["S"] = measureHangingSerif(t, sp, f.C.X+f.Upper.Lx)
	turned := mapPath(sp, func(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} })
	beaks["S lower"] = measureHangingSerif(t, turned, -(f.C.X - f.Lower.Lx))
	return e, beaks
}

// The beaks of C, G and S are serifs of the family of the serifs on the arms of
// E: each is as wide as the top serif of E, within 4 units, halfway along the
// straight part of its face; that straight part is 0.6 to 1.6 times as long as
// on E, and each is bracketed, the ink widening to twice the serif within 4
// times the height in which the bracket of E does so, and not at once.
func TestCapBeaksAreArmSerifs(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		e, beaks := capBeakSerifs(t, np)
		for name, b := range beaks {
			if math.Abs(b.Thick-e.Thick) > 4 {
				t.Errorf("the serif of %s is %.1f wide, that of E %.1f", name, b.Thick, e.Thick)
			}
			if r := b.Face / e.Face; r < 0.6 || r > 1.6 {
				t.Errorf("the straight face of the serif of %s is %.1f long, %.2f times that of E, want 0.6 to 1.6", name, b.Face, r)
			}
			if b.Bracket < 5 || b.Bracket > 4*e.Bracket {
				t.Errorf("the bracket of the serif of %s is %.1f high, that of E %.1f: want 5 to %.1f", name, b.Bracket, e.Bracket, 4*e.Bracket)
			}
			if np.name == "regular" {
				t.Logf("%s: thick %.1f face %.1f bracket %.1f (E %.1f %.1f %.1f)", name, b.Thick, b.Face, b.Bracket, e.Thick, e.Face, e.Bracket)
			}
		}
	})
}

// The arcs which carry the beaks of C and G are thin: on every vertical from 90
// to 250 units inside the cut the upper arc is at most 1.5 times as thick as
// the thin part of O at the top.
func TestCapBeakArcsAreThin(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		cap := capHeight(np)
		limit := 1.5 * p.Caps.RoundThin
		for _, name := range []string{"C", "G"} {
			path := capOutline(np, name)
			x, _, _ := cBeak(t, path, cap)
			for d := 90.0; d <= 250; d += 5 {
				v := vruns(path, x-d)
				if len(v) < 2 {
					t.Fatalf("%s: %d runs %v inside the cut", name, len(v), d)
				}
				if r := v[len(v)-1]; r[1]-r[0] > limit {
					t.Errorf("%s: %.0f inside the cut the upper arc is %.1f thick, want at most %.1f", name, d, r[1]-r[0], limit)
				}
			}
		}
	})
}

// C is balanced: the ink above the middle of its height exceeds the ink below
// by 0 to 5 per cent of the ink above.
func TestCapCIsBalanced(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		c := capOutline(np, "C")
		_, y0, _, y1 := c.BBox()
		mid := (y0 + y1) / 2
		var above, below float64
		for y := y0 + 0.5; y < y1; y++ {
			var w float64
			for _, r := range runs(c, y) {
				w += r[1] - r[0]
			}
			if y > mid {
				above += w
			} else {
				below += w
			}
		}
		if d := (above - below) / above; d < 0 || d > 0.05 {
			t.Errorf("the ink of C above the middle exceeds that below by %.3f of it, want 0 to 0.05", d)
		}
		if np.name == "regular" {
			t.Logf("ink above %.0f below %.0f: %.3f more above", above, below, (above-below)/above)
		}
	})
}

// The left-hand edge of the stem stands straight on the round: it lies
// within half a unit of its line from the foot of the bracket down to 15 units
// above the point where that line meets the counter, and the curve which takes
// the corner off below reaches at most 8 units left of the line where the
// counter lies to the right of that.
func TestCapGStemEdge(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		g := capOutline(np, "G")
		f := capRound(p)
		xR := f.beakX(p)
		xL := xR - p.Caps.Stem
		qi := f.inner.cubic(0, false)
		cross := qi.at(qi.tAtX(xL)).Y
		foot := p.Caps.CGSJQ.GStemTop - p.Caps.BracketHeight
		for y := foot - 1; y > cross+15; y-- {
			r := runs(g, y)
			if e := r[len(r)-1][0]; math.Abs(e-xL) > 0.5 {
				t.Errorf("at the height %.0f the left-hand edge of the stem of G is at %.1f, want %.1f", y, e, xL)
				break
			}
		}
		for y := cross + 15; y > cross-40; y-- {
			if qi.at(qi.tAtY(y)).X < xL-8 {
				break
			}
			r := runs(g, y)
			if e := r[len(r)-1][0]; e < xL-8 {
				t.Errorf("at the height %.0f the stem of G reaches %.1f, more than 8 left of its edge at %.1f", y, e, xL)
				break
			}
		}
	})
}
