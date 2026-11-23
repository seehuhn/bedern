// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// widthAt returns the length of the run of the scan which contains t, or 0.
func widthAt(rs [][2]float64, t float64) float64 {
	for _, r := range rs {
		if r[0] <= t && t <= r[1] {
			return r[1] - r[0]
		}
	}
	return 0
}

// TestDaggerWaist asks in every parameter set that the limbs of † and ‡ are at
// least a hairline (ThinLower) wide where they leave a crossing: the stem
// above and below each crossing, from the edge of the arms to the middle of the
// limb, and each arm from the edge of the stem to its middle.
func TestDaggerWaist(t *testing.T) {
	t.Parallel()
	h := daggerTop - daggerBottom
	y1 := daggerBottom + daggerCross*h
	y0 := daggerTop + daggerBottom - y1
	for _, np := range allParams() {
		p := np.params
		for glyph, crossings := range map[string][]float64{"dagger": {y1}, "daggerdbl": {y1, y0}} {
			path, _ := cachedOutline(np, glyph)
			x0, _, x1, _ := path.BBox()
			cx := (x0 + x1) / 2
			reach := daggerReach * h
			for _, yc := range crossings {
				// The stem, upwards and downwards from the band of the arms.
				for _, dir := range []float64{1, -1} {
					least, inBand := math.Inf(1), true
					for d := 0.0; d < 0.5*(daggerTop-y1); d++ {
						w := widthAt(runs(path, yc+dir*d), cx)
						if inBand && w > reach {
							continue
						}
						inBand = false
						least = math.Min(least, w)
					}
					if least < p.ThinLower {
						t.Errorf("%s: %s: the stem at the crossing %.0f is %.1f wide (dir %v), want at least %.1f", np.name, glyph, yc, least, dir, p.ThinLower)
					}
				}
				// The arms, outwards from the stem.
				for _, dir := range []float64{1, -1} {
					least, inStem := math.Inf(1), true
					for d := 0.0; d < 0.5*reach; d++ {
						w := widthAt(vruns(path, cx+dir*d), yc)
						if inStem && w > 0.5*(daggerTop-y1) {
							continue
						}
						inStem = false
						least = math.Min(least, w)
					}
					if least < p.ThinLower {
						t.Errorf("%s: %s: the arm at the crossing %.0f is %.1f wide (dir %v), want at least %.1f", np.name, glyph, yc, least, dir, p.ThinLower)
					}
				}
			}
		}
	}
}

// TestDaggersShareTheirParts asks in every parameter set that ‡ above its upper
// crossing is the head of †, which crosses at the same height.
func TestDaggersShareTheirParts(t *testing.T) {
	t.Parallel()
	y1 := daggerBottom + daggerCross*(daggerTop-daggerBottom)
	for _, np := range allParams() {
		a, _ := cachedOutline(np, "dagger")
		b, _ := cachedOutline(np, "daggerdbl")
		for y := y1 + 0.5*(daggerTop-y1); y < daggerTop; y += 7 {
			ra, rb := runs(a, y), runs(b, y)
			if len(ra) != len(rb) {
				t.Errorf("%s: at %.0f † has %d runs and ‡ %d", np.name, y, len(ra), len(rb))
				continue
			}
			for i := range ra {
				if math.Abs(ra[i][0]-rb[i][0]) > 0.5 || math.Abs(ra[i][1]-rb[i][1]) > 0.5 {
					t.Errorf("%s: the heads of † and ‡ differ at %.0f: %v against %v", np.name, y, ra[i], rb[i])
				}
			}
		}
	}
}
