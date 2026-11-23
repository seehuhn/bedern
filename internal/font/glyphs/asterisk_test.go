// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// asteriskRing returns, for the circle of the radius rho about the centre
// (cx, cy), the ink at n equally spaced angles from the upward direction.
func asteriskRing(poly []outline.Point, cx, cy, rho float64, n int) []bool {
	ix := newBandIndex(poly)
	ink := make([]bool, n)
	for i := range ink {
		a := math.Pi/2 + 2*math.Pi*float64(i)/float64(n)
		ink[i] = ix.inside(outline.Point{X: cx + rho*math.Cos(a), Y: cy + rho*math.Sin(a)})
	}
	return ink
}

// asteriskStretches returns the lengths of the stretches of ink and of white
// along the ring, in units of the arc.
func asteriskStretches(ink []bool, rho float64) (inks, whites []float64) {
	n := len(ink)
	start := 0
	for start < n && ink[start] == ink[(start+n-1)%n] {
		start++
	}
	if start == n {
		return nil, nil
	}
	step := 2 * math.Pi * rho / float64(n)
	run := 0
	for k := 0; k < n; k++ {
		i := (start + k) % n
		run++
		if ink[(i+1)%n] != ink[i] {
			if ink[i] {
				inks = append(inks, float64(run)*step)
			} else {
				whites = append(whites, float64(run)*step)
			}
			run = 0
		}
	}
	return inks, whites
}

// The asterisk is a star of six arms, symmetric to within one and a half units
// under a sixth of a turn and under the mirror in the vertical axis.  At a
// third of the radius the circle about the centre crosses six separate arms,
// and the white between neighbouring arms is at least 0.6 of an arm there.
// An arm, measured as the arc of ink on circles about the centre, is at its
// thickest (between 0.3 and 0.9 of the radius) at least 0.75 lowercase stems
// and at most 0.75 of the full stop wide, and at 0.3 of the radius 0.25 to 0.55
// stems and at most 0.6 of its thickest.
func TestAsteriskArms(t *testing.T) {
	t.Parallel()
	for _, s := range punctSets() {
		p := s.Params
		path, _ := punctOutline(t, "asterisk", p)
		if len(path) != 1 {
			t.Errorf("%s: the asterisk has %d contours, want 1", s.Name, len(path))
			continue
		}
		poly := path[0].Flatten()
		r := p.Punct.AsteriskSize / 2
		x0, _, x1, _ := path.BBox()
		cx, cy := (x0+x1)/2, p.Punct.AsteriskTop-r
		stem := p.Stem
		const n = 3600

		for _, f := range []float64{0.2, 0.4, 0.6, 0.8} {
			ink := asteriskRing(poly, cx, cy, f*r, n)
			// m samples of the ring are one and a half units of arc.
			m := int(math.Ceil(1.5 * float64(n) / (2 * math.Pi * f * r)))
			near := func(v bool, j int) bool {
				for d := -m; d <= m; d++ {
					if ink[((j+d)%n+n)%n] == v {
						return true
					}
				}
				return false
			}
			for i := range ink {
				if !near(ink[i], i+n/6) {
					t.Errorf("%s: at %.1f of the radius the ink is not repeated a sixth of a turn on", s.Name, f)
					break
				}
				if !near(ink[i], n-i) {
					t.Errorf("%s: at %.1f of the radius the ink is not mirrored in the vertical", s.Name, f)
					break
				}
			}
		}

		inks, whites := asteriskStretches(asteriskRing(poly, cx, cy, r/3, n), r/3)
		if len(inks) != 6 {
			t.Errorf("%s: at a third of the radius the circle crosses %d arms, want 6", s.Name, len(inks))
		} else {
			for i, w := range whites {
				if w < 0.6*inks[i] {
					t.Errorf("%s: at a third of the radius the white between two arms is %.1f, an arm %.1f, want at least 0.6 of it", s.Name, w, inks[i])
					break
				}
			}
		}

		thickest := 0.0
		for f := 0.3; f <= 0.9+1e-9; f += 0.01 {
			inks, _ := asteriskStretches(asteriskRing(poly, cx, cy, f*r, n), f*r)
			for _, w := range inks {
				thickest = max(thickest, w)
			}
		}
		inks, _ = asteriskStretches(asteriskRing(poly, cx, cy, 0.3*r, n), 0.3*r)
		thin := 0.0
		for _, w := range inks {
			thin = max(thin, w)
		}
		if thickest < 0.75*stem {
			t.Errorf("%s: an arm is %.1f wide at its thickest, %.2f stems, want 0.75 or more", s.Name, thickest, thickest/stem)
		}
		if thickest > 0.75*p.DotDiameter {
			t.Errorf("%s: an arm is %.1f wide at its thickest, %.2f of the full stop, want at most 0.75", s.Name, thickest, thickest/p.DotDiameter)
		}
		if thin < 0.25*stem || thin > 0.55*stem {
			t.Errorf("%s: an arm is %.1f wide at 0.3 of the radius, %.2f stems, want 0.25 to 0.55", s.Name, thin, thin/stem)
		}
		if thin > 0.6*thickest {
			t.Errorf("%s: an arm is %.1f wide at 0.3 of the radius and %.1f at its thickest, want at most 0.6 of it", s.Name, thin, thickest)
		}
	}
}
