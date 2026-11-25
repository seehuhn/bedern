// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// widestInscribed returns the diameter of the largest circle which lies in the
// ink of the closed contour c (a counter if c is a hole of the path) and does
// not touch any contour of the path, found on a grid of step units.
func widestInscribed(path outline.Path, c outline.Contour, step, enough float64) float64 {
	var edges [][2]outline.Point
	for _, k := range path {
		pts := k.Flatten()
		for i := range pts {
			edges = append(edges, [2]outline.Point{pts[i], pts[(i+1)%len(pts)]})
		}
	}
	poly := c.Flatten()
	x0, y0, x1, y1 := outline.Path{c}.BBox()
	best := 0.0
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			if !insidePolygon(poly, x, y) {
				continue
			}
			d := math.Inf(1)
			for _, e := range edges {
				d = math.Min(d, segDist(x, y, e[0], e[1]))
			}
			best = math.Max(best, 2*d)
			if best >= enough {
				return best
			}
		}
	}
	return best
}

func insidePolygon(poly []outline.Point, x, y float64) bool {
	in := false
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		if (a.Y > y) != (b.Y > y) && x < a.X+(y-a.Y)*(b.X-a.X)/(b.Y-a.Y) {
			in = !in
		}
	}
	return in
}

func segDist(x, y float64, a, b outline.Point) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	l := dx*dx + dy*dy
	t := 0.0
	if l > 0 {
		t = math.Max(0, math.Min(1, ((x-a.X)*dx+(y-a.Y)*dy)/l))
	}
	return math.Hypot(x-a.X-t*dx, y-a.Y-t*dy)
}

// TestDoubleStruckWhiteIsOpenEverywhere requires every white of a
// double-struck letter, every counter of its doubled strokes, to be at its
// widest at least 0.8 of the white of the level wide (the rule, well above
// dsWhiteMin); no channel closes up at a vertex or a serif, and the counters
// of the bowls and rounds, which are wider, pass too.  The widest part is the
// largest circle which fits in the counter; the search stops at the first
// circle of the diameter enough.
func TestDoubleStruckWhiteIsOpenEverywhere(t *testing.T) {
	for _, s := range derivedSets(0)[:1] {
		m := dsSet(s)
		for _, l := range dsLetters {
			for level, suffix := range []string{"", ".s1", ".s2"} {
				g := m[l.name+suffix]
				path := g.Path.Round()
				want := 0.8 * math.Round(dsGap*dsLevels[level].wide)
				for i, c := range path {
					if i == 0 || (c.SignedArea() > 0) == (path[0].SignedArea() > 0) {
						continue // the outer contour, or a part which is not a counter
					}
					// A channel, which is long against its width, is a white of a doubled
					// stroke; any other counter need only be open.
					need := dsWhiteMin[level]
					if area := math.Abs(c.SignedArea()); area > 3*want*want {
						need = want
					}
					if w := widestInscribed(path, c, 8, need); w < need {
						t.Errorf("%s: a counter of %s is %.1f wide at its widest, want at least %.0f", s.Name, g.Name, w, need)
					}
				}
			}
		}
	}
}

// dsBendMax is the greatest thickness of the ink at the bend of U and J, in
// lines: the doubled stem runs into the round as two lines which join; the
// bend is about two lines thick, 97 and 99 units at text size.
const dsBendMax = 2.2

// TestDoubleStruckBendsAreNotHeavy requires the ink of the bend of U and J,
// below the straight stem, to be nowhere thicker than dsBendMax lines.  The
// thickness is the diameter of the largest circle which fits in the ink there.
func TestDoubleStruckBendsAreNotHeavy(t *testing.T) {
	for _, s := range derivedSets(0)[:1] {
		p := s.Params
		m := dsSet(s)
		for _, name := range []string{"u1D54C", "u1D541"} {
			for level, suffix := range []string{"", ".s1", ".s2"} {
				g := m[name+suffix]
				path := g.Path.Round()
				_, y0, _, _ := path.BBox()
				yTop := y0 + 0.30*p.CapHeight
				line := math.Round(dsLine * p.CapDiag.ThinStem * dsLevels[level].weight)
				best := 0.0
				// The ink under the height yTop: the circles are taken against
				// the whole outline, from the outline contour only.
				var edges [][2]outline.Point
				for _, k := range path {
					pts := k.Flatten()
					for i := range pts {
						edges = append(edges, [2]outline.Point{pts[i], pts[(i+1)%len(pts)]})
					}
				}
				x0, _, x1, _ := path.BBox()
				for y := y0; y <= yTop; y += 5 {
					for x := x0; x <= x1; x += 5 {
						if !inkAt(path, x, y) {
							continue
						}
						d := math.Inf(1)
						for _, e := range edges {
							d = math.Min(d, segDist(x, y, e[0], e[1]))
						}
						best = math.Max(best, 2*d)
					}
				}
				if best > dsBendMax*line {
					t.Errorf("%s: the bend of %s is %.1f thick, want at most %.0f", s.Name, g.Name, best, dsBendMax*line)
				}
			}
		}
	}
}

// inkAt reports whether the point is in the ink of the path, by the even-odd
// rule.
func inkAt(path outline.Path, x, y float64) bool {
	in := false
	for _, c := range path {
		if insidePolygon(c.Flatten(), x, y) {
			in = !in
		}
	}
	return in
}

var _ = glyphs.Outline
