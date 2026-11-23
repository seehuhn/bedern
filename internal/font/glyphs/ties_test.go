// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// TestTies asks, in every parameter set, that each of the four ties is one
// contour, symmetric about the middle of its ink within a unit, whose two
// ends are its lowest ink; that it is thickest at its top, at least half a
// stem there, at least one and a half times as thick as its ends are wide
// and no less high than at a quarter of its width from an end; that the ties of the lowercase keep the gap of the lowercase accents
// from the x-height and those of the capitals the gap of the capital accents
// from the cap height; that the ink of a tie starts less than a stem to the
// left of the middle of its advance and ends beyond the advance; that the tie
// of the capitals is wider than that of the lowercase; and that a rounder
// form is as high as its tie, narrower by at least a tenth, and stands inside
// its advance with equal side spaces, within a unit.
func TestTies(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			width := map[string]float64{}
			height := map[string]float64{}
			for _, c := range []struct {
				name string
				line float64
			}{
				{"tieaccentlowercase", p.XHeight + p.Acc.Gap},
				{"tieaccentcapital", p.CapHeight + p.Acc.GapCap},
				{"tieaccentlowercase.new", p.XHeight + p.Acc.Gap},
				{"tieaccentcapital.new", p.CapHeight + p.Acc.GapCap},
			} {
				path, adv := cachedOutline(np, c.name)
				if len(path) != 1 {
					t.Errorf("%s has %d contours, want 1", c.name, len(path))
					continue
				}
				x0, y0, x1, y1 := path.BBox()
				width[c.name], height[c.name] = x1-x0, y1-y0
				if y0 < c.line-0.5 {
					t.Errorf("%s: the lowest ink is at %v, below %v", c.name, y0, c.line)
				}
				ends := runs(path, y0+1)
				if len(ends) != 2 || math.Abs(ends[0][0]-x0) > 1 || math.Abs(ends[1][1]-x1) > 1 {
					t.Errorf("%s: the ink at its foot is %v, want the two ends of the arc from %v to %v", c.name, ends, x0, x1)
					continue
				}
				end := math.Max(ends[0][1]-ends[0][0], ends[1][1]-ends[1][0])
				top := vruns(path, (x0+x1)/2)
				if len(top) != 1 {
					t.Errorf("%s: the scan through the middle has %d runs, want 1", c.name, len(top))
					continue
				}
				thick := top[0][1] - top[0][0]
				if thick < 0.5*p.Stem || thick < 1.5*end {
					t.Errorf("%s: the arc is %.1f thick at the top and its ends %.1f wide, want at least half a stem of %v and 1.5 ends", c.name, thick, end, p.Stem)
				}
				for _, f := range []float64{0.1, 0.25, 0.4} {
					l, r := vruns(path, x0+f*(x1-x0)), vruns(path, x1-f*(x1-x0))
					if len(l) != 1 || len(r) != 1 || math.Abs(l[0][0]-r[0][0]) > 1 || math.Abs(l[0][1]-r[0][1]) > 1 {
						t.Errorf("%s: the arc is %v at %v of its width from the left and %v from the right", c.name, l, f, r)
					} else if w := l[0][1] - l[0][0]; f >= 0.25 && w > thick+0.01 {
						t.Errorf("%s: the arc is %.1f high at %v of its width, more than %.1f at the top", c.name, w, f, thick)
					}
				}
				if c.name == "tieaccentlowercase" || c.name == "tieaccentcapital" {
					if d := adv/2 - x0; d < 0 || d > p.Stem {
						t.Errorf("%s: the ink starts %v to the left of the middle of the advance, want from 0 to a stem of %v", c.name, d, p.Stem)
					}
					if x1 <= adv {
						t.Errorf("%s: the ink ends at %v, inside the advance %v", c.name, x1, adv)
					}
				} else if x0 < 0 || x1 > adv || math.Abs(x0-(adv-x1)) > 1 {
					t.Errorf("%s: the side spaces are %v and %v", c.name, x0, adv-x1)
				}
			}
			if len(width) != 4 {
				return
			}
			if width["tieaccentcapital"] <= width["tieaccentlowercase"] {
				t.Errorf("the tie of the capitals is %v wide, and that of the lowercase %v", width["tieaccentcapital"], width["tieaccentlowercase"])
			}
			for _, name := range []string{"tieaccentlowercase", "tieaccentcapital"} {
				if width[name+".new"] > 0.9*width[name] || math.Abs(height[name+".new"]-height[name]) > 0.01 {
					t.Errorf("%s: the rounder form is %v by %v, and the tie %v by %v", name, width[name+".new"], height[name+".new"], width[name], height[name])
				}
			}
		})
	}
}

// TestRogate asks, in every parameter set, that the o with a rogate has, in
// both cases, the advance and the counter of its letter and one outer
// contour; that its ink is as wide as that of the letter; that its strokes
// reach equally far above and below the letter, within a unit, by at least a
// quarter of the x-height; and that each stroke stands on the middle of the
// letter within a unit and is between 0.8 hairlines and a stem wide.
func TestRogate(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, c := range [][2]string{{"Orogate", "O"}, {"orogate", "o"}} {
				path, adv := cachedOutline(np, c[0])
				letter, ladv := cachedOutline(np, c[1])
				if adv != ladv {
					t.Errorf("%s has the advance %v, and %s %v", c[0], adv, c[1], ladv)
				}
				if len(path) != 2 || len(letter) != 2 {
					t.Errorf("%s has %d contours and %s %d, want 2 each", c[0], len(path), c[1], len(letter))
					continue
				}
				var counter, own outline.Path
				for _, ct := range letter {
					if ct.SignedArea() < 0 {
						counter = append(counter, ct)
					}
				}
				for _, ct := range path {
					if ct.SignedArea() < 0 {
						own = append(own, ct)
					}
				}
				cx0, cy0, cx1, cy1 := counter.BBox()
				ox0, oy0, ox1, oy1 := own.BBox()
				if math.Abs(cx0-ox0)+math.Abs(cy0-oy0)+math.Abs(cx1-ox1)+math.Abs(cy1-oy1) > 0.01 {
					t.Errorf("%s: the counter is not that of %s", c[0], c[1])
				}
				x0, y0, x1, y1 := path.BBox()
				lx0, ly0, lx1, ly1 := letter.BBox()
				if x0 != lx0 || x1 != lx1 {
					t.Errorf("%s: the ink runs from %v to %v, and that of %s from %v to %v", c[0], x0, x1, c[1], lx0, lx1)
				}
				above, below := y1-ly1, ly0-y0
				if math.Abs(above-below) > 1 || above < 0.25*p.XHeight {
					t.Errorf("%s: the strokes reach %v above and %v below the letter, want equal and at least %v", c[0], above, below, 0.25*p.XHeight)
				}
				for _, y := range []float64{y1 - 0.5*above, y0 + 0.5*below} {
					rs := runs(path, y)
					if len(rs) != 1 {
						t.Errorf("%s: the scan at %v has %d runs, want 1", c[0], y, len(rs))
						continue
					}
					if w := rs[0][1] - rs[0][0]; w < 0.8*p.Hair || w > p.Stem || math.Abs((rs[0][0]+rs[0][1])/2-(lx0+lx1)/2) > 1 {
						t.Errorf("%s: the stroke at %v is %v, want between 0.8 hairlines and a stem wide on the middle %v", c[0], y, rs[0], (lx0+lx1)/2)
					}
				}
			}
		})
	}
}
