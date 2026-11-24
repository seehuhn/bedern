// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"flag"
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// fitFlag names the part whose shape constants TestFit fits to the outline of
// the font and prints.  The descriptions are drawn by rule, without a search.
var fitFlag = flag.String("fit", "", "fit the shape constants of the part (n, s, j, 9) to the font and print them")

// fitCost returns the largest distance of the part from the font, or a large
// value if it cannot be built.
func fitCost(pt Part, font outline.Path) float64 {
	rs, err := pt.Build()
	if err != nil {
		return 1e6
	}
	ds, _ := distances(rs, font, 10)
	var sum float64
	for _, x := range ds {
		sum += math.Pow(x, 8)
	}
	var worst float64
	for _, x := range ds {
		worst = math.Max(worst, x)
	}
	d := 0.5*math.Pow(sum/float64(len(ds)), 1.0/8) + 0.5*worst
	for _, r := range rs {
		d += 3 * float64(len(r.Warnings))
	}
	return d
}

// nelderMead minimises f from x0 with the initial steps h.
func nelderMead(f func([]float64) float64, x0, h []float64, iters int) []float64 {
	n := len(x0)
	type pt struct {
		x []float64
		v float64
	}
	s := []pt{{append([]float64(nil), x0...), f(x0)}}
	for i := 0; i < n; i++ {
		x := append([]float64(nil), x0...)
		x[i] += h[i]
		s = append(s, pt{x, f(x)})
	}
	comb := func(a []float64, b []float64, t float64) []float64 {
		out := make([]float64, n)
		for i := range out {
			out[i] = a[i] + t*(b[i]-a[i])
		}
		return out
	}
	for it := 0; it < iters; it++ {
		sort.Slice(s, func(i, j int) bool { return s[i].v < s[j].v })
		c := make([]float64, n)
		for _, q := range s[:n] {
			for i := range c {
				c[i] += q.x[i] / float64(n)
			}
		}
		worst := s[n]
		xr := comb(c, worst.x, -1)
		vr := f(xr)
		switch {
		case vr < s[0].v:
			xe := comb(c, worst.x, -2)
			if ve := f(xe); ve < vr {
				s[n] = pt{xe, ve}
			} else {
				s[n] = pt{xr, vr}
			}
		case vr < s[n-1].v:
			s[n] = pt{xr, vr}
		default:
			xc := comb(c, worst.x, 0.5)
			if vc := f(xc); vc < worst.v {
				s[n] = pt{xc, vc}
			} else {
				for i := 1; i <= n; i++ {
					s[i].x = comb(s[0].x, s[i].x, 0.5)
					s[i].v = f(s[i].x)
				}
			}
		}
	}
	sort.Slice(s, func(i, j int) bool { return s[i].v < s[j].v })
	return s[0].x
}

func TestFit(t *testing.T) {
	if *fitFlag == "" {
		t.Skip("no -fit")
	}
	p := design.Regular()
	var build func(x []float64) Part
	var x0, h []float64
	switch *fitFlag {
	case "n":
		a := nArch
		x0 = []float64{a.Steeper, a.DWStem, a.StemK.Right, a.StemK.Left, a.TopShift, a.DWTop, a.TopK.Right, a.TopK.Left}
		h = []float64{2, 0.1, 0.002, 0.002, 4, 0.1, 0.003, 0.003}
		build = func(x []float64) Part {
			return romanNWith(p, archShape{x[0], x[1], EdgeK{x[2], x[3]}, x[4], x[5], EdgeK{x[6], x[7]}})
		}
	case "s":
		a := sSpine
		x0 = []float64{a.Side.Right, a.Side.Left, a.DW}
		h = []float64{0.003, 0.003, 0.1}
		build = func(x []float64) Part { return romanSWith(p, spineShape{EdgeK{x[0], x[1]}, x[2]}) }
	case "j":
		a := jHook
		x0 = []float64{a.BottomX, a.BottomW, a.BottomK.Right, a.BottomK.Left, a.BallUp}
		h = []float64{0.05, 3, 0.003, 0.002, 4}
		build = func(x []float64) Part {
			return romanJWith(p, hookShape{x[0], x[1], EdgeK{x[2], x[3]}, x[4]})
		}
	case "9":
		a := nineTail
		x0 = []float64{a.TopW, a.TopDir, a.TopK.Right, a.TopK.Left, a.BottomX, a.BottomW, a.BottomK.Right, a.BottomK.Left, a.TipK}
		h = []float64{4, 3, 0.002, 0.002, 8, 4, 0.003, 0.002, 0.004}
		build = func(x []float64) Part {
			return romanNineWith(p, tailShape{x[0], x[1], EdgeK{x[2], x[3]}, x[4], x[5], EdgeK{x[6], x[7]}, x[8]})
		}
	case "d":
		x0 = []float64{dJoinK[0].Right, dJoinK[0].Left, dJoinK[1].Right, dJoinK[1].Left}
		h = []float64{0.003, 0.003, 0.003, 0.003}
		build = func(x []float64) Part {
			return romanDWith(p, [2]EdgeK{{Right: x[0], Left: x[1]}, {Right: x[2], Left: x[3]}})
		}
	default:
		t.Fatalf("unknown part %q", *fitFlag)
	}
	glyph := build(x0).Glyph
	font, _, _ := glyphs.Outline(glyph, p)
	fr := font.Round()
	f := func(x []float64) float64 { return fitCost(build(x), fr) }
	x := x0
	for round := 0; round < 3; round++ {
		x = nelderMead(f, x, h, 400)
		t.Logf("round %d: cost %.3f at %v", round, f(x), x)
	}
}
