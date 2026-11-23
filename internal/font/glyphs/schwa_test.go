// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// TestSchwaForm asks of ə against o and of Ə against O, in every parameter
// set: that the letter reaches as high and as low as the round letter; that
// the ink of Ə is 0.88 to 1.00 of that of O wide and ə as wide as e; that the
// bar is one run of ink across the letter with its middle between 0.31 and
// 0.51 of the height; that the white between the bar and the terminal, on
// the vertical through the tip of the terminal, is 0.14 to 0.46 of the
// height; that halfway up that white a horizontal scan from
// the left meets no ink before the middle of the letter; and that the stroke
// which it meets there is 0.9 to 1.15 times the stroke of the round letter
// at its left extreme.
func TestSchwaForm(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, c := range []struct {
				name, round string
				height      float64
				minW, maxW  float64
			}{
				{"uni0259", "o", p.XHeight, 0, 0},
				{"uni018F", "O", p.CapHeight, 0.88, 1.00},
			} {
				path, _ := cachedOutline(np, c.name)
				round, _ := cachedOutline(np, c.round)
				x0, y0, x1, y1 := path.BBox()
				rx0, ry0, rx1, ry1 := round.BBox()
				if math.Abs(y0-ry0) > 1 || math.Abs(y1-ry1) > 1 {
					t.Errorf("%s: from %.1f to %.1f, and %s from %.1f to %.1f", c.name, y0, y1, c.round, ry0, ry1)
				}
				w := x1 - x0
				if c.maxW > 0 {
					if r := w / (rx1 - rx0); r < c.minW || r > c.maxW {
						t.Errorf("%s: the ink is %.3f of that of %s wide, want %v to %v", c.name, r, c.round, c.minW, c.maxW)
					}
				} else {
					e, _ := cachedOutline(np, "e")
					ex0, _, ex1, _ := e.BBox()
					if math.Abs(w-(ex1-ex0)) > 1 {
						t.Errorf("%s: the ink is %.1f wide, and that of e %.1f", c.name, w, ex1-ex0)
					}
				}

				// The vertical through the tip of the terminal meets the
				// lower stroke, the bar and the terminal.
				var v [][2]float64
				for f := 0.1; f < 0.5 && len(v) != 3; f += 0.01 {
					v = vruns(path, x0+f*w)
				}
				if len(v) != 3 {
					t.Errorf("%s: no vertical in the left half meets the lower stroke, the bar and the terminal", c.name)
					continue
				}
				bar, tip := v[1], v[2][0]
				mid := (bar[0] + bar[1]) / 2
				if r := mid / c.height; r < 0.31 || r > 0.51 {
					t.Errorf("%s: the middle of the bar is at %.3f of the height, want 0.31 to 0.51", c.name, r)
				}
				if h := runs(path, mid); len(h) != 1 || h[0][0] > x0+0.02*w || h[0][1] < x1-0.1*w {
					t.Errorf("%s: the bar is not one run of ink across the letter: %v", c.name, h)
				}
				if r := (tip - bar[1]) / c.height; r < 0.14 || r > 0.46 {
					t.Errorf("%s: the white between the bar and the terminal is %.3f of the height, want 0.14 to 0.46", c.name, r)
				}
				h := runs(path, (bar[1]+tip)/2)
				if len(h) != 1 || h[0][0] < (x0+x1)/2 {
					t.Errorf("%s: the scan between the bar and the terminal meets ink before the middle: %v", c.name, h)
					continue
				}
				left := runs(round, (ry0+ry1)/2)[0]
				if r := (h[0][1] - h[0][0]) / (left[1] - left[0]); r < 0.9 || r > 1.15 {
					t.Errorf("%s: the right stroke is %.3f of the stroke of %s, want 0.9 to 1.15", c.name, r, c.round)
				}
			}
		})
	}
}

// TestDotBelow asks, in every parameter set, that the dot below of ḍ and Ḍ is
// as wide and as high as the dot above of ċ within a unit; that its top
// stands between 0.30 and 0.56 of the descender below the baseline and its
// lowest ink no lower than 1.35 descenders; that under ḍ its middle stands
// within 3 of the middle of the ink of d above the baseline; and that under
// Ḍ it stands left of the middle of the ink of D, by at most 0.08 of the
// width of that ink.
func TestDotBelow(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		if !composedRoutine(np.name) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			ax0, ay0, ax1, ay1 := markPart(np, "cdotaccent")
			for _, c := range []struct{ name, base string }{{"uni1E0D", "d"}, {"uni1E0C", "D"}} {
				x0, y0, x1, y1 := markPart(np, c.name)
				if math.Abs((x1-x0)-(ax1-ax0)) > 1 || math.Abs((y1-y0)-(ay1-ay0)) > 1 {
					t.Errorf("%s: the dot is %.0f by %.0f, and the dot above %.0f by %.0f", c.name, x1-x0, y1-y0, ax1-ax0, ay1-ay0)
				}
				if r := -y1 / p.Descender; r < 0.30 || r > 0.56 {
					t.Errorf("%s: the top of the dot is %.3f descenders below the baseline, want 0.30 to 0.56", c.name, r)
				}
				if r := -y0 / p.Descender; r > 1.35 {
					t.Errorf("%s: the dot reaches %.3f descenders below the baseline, want at most 1.35", c.name, r)
				}
				b := cachedRounded(np, c.base)
				bx0, bx1 := inkSpan(b.Path, 0, p.CapHeight)
				off := (x0+x1)/2 - (bx0+bx1)/2
				switch c.base {
				case "d":
					if math.Abs(off) > 3 {
						t.Errorf("%s: the dot stands %.1f from the middle of the ink, want at most 3", c.name, off)
					}
				case "D":
					if r := -off / (bx1 - bx0); r < 0 || r > 0.08 {
						t.Errorf("%s: the dot stands %.3f of the width of the ink left of its middle, want 0 to 0.08", c.name, r)
					}
				}
			}
		})
	}
}
