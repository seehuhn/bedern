// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/outline"
)

// A Letter is a set of strokes which together make a glyph, and further closed
// contours (a dot) which are not strokes.  Strokes may overlap and cross;
// Build returns the union.
type Letter struct {
	Strokes []Stroke
	Extra   []Chain
}

// A LetterResult is a letter as outline.
type LetterResult struct {
	Strokes  []*Result
	Contours []Chain
	Warnings []string
	Findings []string // the findings of internal/fair on the rounded outline
}

// sliverArea is the area in square units below which a contour of the union
// is a sliver, left where two edges cross at a shallow angle; it is dropped
// with a warning.
const sliverArea = 40.0

// sharpCorner is the angle in degrees below which a corner of the union is
// reported: such a wedge of ink or of white closes up at small sizes.
const sharpCorner = 25.0

// Build returns the union of the strokes and the extra contours.
func (l Letter) Build() (*LetterResult, error) {
	res := &LetterResult{}
	var cs []Chain
	for i, s := range l.Strokes {
		r, err := s.Build()
		if err != nil {
			return nil, fmt.Errorf("stroke %d: %w", i, err)
		}
		res.Strokes = append(res.Strokes, r)
		cs = append(cs, oriented(r.Contours, s.Closed)...)
	}
	cs = append(cs, l.Extra...)
	u, err := Union(cs)
	if err != nil {
		return nil, err
	}
	// A warning about the shape of an edge at a place which another stroke
	// covers, inside the ink of the union and more than coveredMargin from
	// its outline, does not concern the letter.
	covered := coverTest(u)
	for i, r := range res.Strokes {
		for j, w := range r.Warnings {
			if j < len(r.WarnAt) && covered(r.WarnAt[j]) {
				continue
			}
			res.Warnings = append(res.Warnings, fmt.Sprintf("stroke %d: %s", i, w))
		}
	}
	for _, c := range u {
		if a := math.Abs(signedArea(c)); a < sliverArea {
			res.Warnings = append(res.Warnings, fmt.Sprintf("a sliver of %.0f square units at %.0f,%.0f is dropped", a, c[0].C[0].X, c[0].C[0].Y))
			continue
		}
		var r Result
		for _, sr := range res.Strokes {
			r.hairpins = append(r.hairpins, sr.hairpins...)
		}
		c = r.mergeClosed(c)
		res.Warnings = append(res.Warnings, r.Warnings...)
		for i, pc := range c {
			if pc.Kind == HiddenPiece {
				res.Warnings = append(res.Warnings, fmt.Sprintf("a hidden end shows at %.0f,%.0f", pc.C[0].X, pc.C[0].Y))
			}
			next := c[(i+1)%len(c)]
			turn := math.Abs(angleBetween(pc.C.Tangent(1), next.C.Tangent(0)))
			if 180-turn < sharpCorner {
				res.Warnings = append(res.Warnings, fmt.Sprintf("a corner of %.0f degrees at %.0f,%.0f", 180-turn, pc.C[3].X, pc.C[3].Y))
			}
		}
		res.Contours = append(res.Contours, c)
	}
	// The findings of the smoothness gate on the rounded outline, so that the
	// drawer sees what the gate will see, with the place.
	an := fair.Analyse("letter", res.Path().Round(), fair.Default)
	for _, f := range an.Findings {
		res.Findings = append(res.Findings, fmt.Sprintf("%s at %.0f,%.0f (%s)", f.Kind, f.X, f.Y, f.Detail))
	}
	return res, nil
}

// coveredMargin is the least distance from the outline of the union of a
// place inside the ink which counts as covered.
const coveredMargin = 1.0

// coverTest returns a function which reports whether a place lies inside the
// ink of the contours and more than coveredMargin from their outline.  A
// place with NaN coordinates is never covered.
func coverTest(u []Chain) func(Point) bool {
	var polys [][]Point
	for _, c := range u {
		var pts []Point
		for _, pc := range c {
			pts = append(pts, pieceSamples(pc.C)...)
		}
		polys = append(polys, pts)
	}
	return func(q Point) bool {
		if math.IsNaN(q.X) {
			return false
		}
		wind := 0
		for _, pts := range polys {
			for i := range pts {
				a, b := pts[i], pts[(i+1)%len(pts)]
				if segDist(q, a, b) < coveredMargin {
					return false
				}
				if (a.Y <= q.Y) != (b.Y <= q.Y) {
					x := a.X + (q.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
					if x > q.X {
						if b.Y > a.Y {
							wind++
						} else {
							wind--
						}
					}
				}
			}
		}
		return wind != 0
	}
}

// oriented returns the contours of a stroke with the ink on their left: the
// contour of an open stroke counter-clockwise, and of a closed stroke the
// larger counter-clockwise and the smaller clockwise.
func oriented(cs []Chain, closed bool) []Chain {
	out := append([]Chain(nil), cs...)
	if !closed {
		for i, c := range out {
			if signedArea(c) < 0 {
				out[i] = c.reversed()
			}
		}
		return out
	}
	a0, a1 := signedArea(out[0]), signedArea(out[1])
	big, small := 0, 1
	if math.Abs(a1) > math.Abs(a0) {
		big, small = 1, 0
	}
	if signedArea(out[big]) < 0 {
		out[big] = out[big].reversed()
	}
	if signedArea(out[small]) > 0 {
		out[small] = out[small].reversed()
	}
	return out
}

// mergeClosed merges the short pieces of a closed chain, as merge does, with
// the pieces round the start of the chain as neighbours too.
func (r *Result) mergeClosed(c Chain) Chain {
	if len(c) < 3 {
		return c
	}
	// Start the chain at a corner, if there is one, so that every smooth join
	// lies inside it.
	for i := range c {
		prev := c[(i+len(c)-1)%len(c)]
		if math.Abs(angleBetween(prev.C.Tangent(1), c[i].C.Tangent(0))) > 1 {
			c = append(append(Chain(nil), c[i:]...), c[:i]...)
			break
		}
	}
	return r.merge(c, true, true)
}

// Path returns the contours of the letter as an outline path, exact.
func (l *LetterResult) Path() outline.Path {
	var p outline.Path
	for _, c := range l.Contours {
		p = append(p, chainContour(c))
	}
	return p
}

// Disc returns a circle as a closed counter-clockwise chain with nodes at its
// extremes, for a dot.
func Disc(c Point, r float64) Chain {
	return circleArc(c, r, 0, 2*math.Pi, EndPiece)
}
