// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestSHalves asks that, upright, the lower half of s is 1.02 to 1.08 times
// as wide in the ink as the upper (the roman 1.05), and that s keeps to the overshoot at the top and bottom.
func TestSHalves(t *testing.T) {
	p := Default()
	g := glyph(t, "s")
	up := unsheared(p, g.Path)
	width := func(y0, y1 float64) float64 {
		lo, hi := math.Inf(1), math.Inf(-1)
		for y := y0 + 0.5; y < y1; y += 2 {
			for _, r := range runs(up, y) {
				lo, hi = math.Min(lo, r[0]), math.Max(hi, r[1])
			}
		}
		return hi - lo
	}
	lower, upper := width(-p.Overshoot, p.XHeight/2-20), width(p.XHeight/2+20, p.XHeight+p.Overshoot)
	if f := lower / upper; f < 1.02 || f > 1.08 {
		t.Errorf("s: the lower half is %.0f, the upper %.0f wide, ratio %.3f, want 1.02 to 1.08", lower, upper, f)
	}
	_, y0, _, y1 := bbox(g.Path)
	if math.Abs(y0+p.Overshoot) > 1 || math.Abs(y1-p.XHeight-p.Overshoot) > 1 {
		t.Errorf("s: the ink runs from %.1f to %.1f, want %v to %v", y0, y1, -p.Overshoot, p.XHeight+p.Overshoot)
	}
}

// TestSApertures asks that the white between each end of s and the spine,
// on the vertical line through the end upright, is at least 40.
func TestSApertures(t *testing.T) {
	p := Default()
	g := glyph(t, "s")
	up := unsheared(p, g.Path)
	w := p.SEndW * p.Thin
	mid := p.SWidth / 2
	for _, cs := range []struct {
		name string
		x    float64
	}{{"upper", mid + p.SUpper/2 + p.SUpperShift - w/2}, {"lower", mid - p.SLower/2 + w/2}} {
		ys := crossings(up, Point{X: cs.x, Y: 0}, Point{Y: 1})
		best := 0.0
		for i := 1; i+1 < len(ys); i += 2 {
			best = math.Max(best, ys[i+1]-ys[i])
		}
		if best < 40 {
			t.Errorf("s: the %s aperture is %.1f, want at least 40 (crossings %v)", cs.name, best, ys)
		}
	}
}
