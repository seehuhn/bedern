// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"strings"

	"seehuhn.de/go/bedern/internal/outline"
)

// SVGPath returns the path as SVG path data, with y pointing down, moved by
// (dx, dy) after the flip.
func SVGPath(p outline.Path, dx, dy float64) string {
	var b strings.Builder
	f := func(q Point) string { return fmt.Sprintf("%.2f %.2f", q.X+dx, dy-q.Y) }
	for _, c := range p {
		fmt.Fprintf(&b, "M%s", f(c.Start))
		for _, s := range c.Segs {
			if s.IsCurve {
				fmt.Fprintf(&b, "C%s %s %s", f(s.P1), f(s.P2), f(s.P3))
			} else {
				fmt.Fprintf(&b, "L%s", f(s.P3))
			}
		}
		b.WriteString("Z")
	}
	return b.String()
}

// SVGNodes returns small circles at the on-curve points of the path.
func SVGNodes(p outline.Path, dx, dy, r float64, colour string) string {
	var b strings.Builder
	for _, c := range p {
		for _, s := range c.Segs {
			fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.1f" fill="%s"/>`, s.P3.X+dx, dy-s.P3.Y, r, colour)
		}
	}
	return b.String()
}

// SVGCentre returns the centre lines of the strokes as polylines through
// their nodes, with the pen drawn across at every node.
func SVGCentre(ss []Stroke, dx, dy float64) string {
	var b strings.Builder
	for _, s := range ss {
		ns, err := s.resolved()
		if err != nil {
			continue
		}
		for _, nd := range ns {
			n := rot90(dirOf(nd.Dir))
			a, c := add(nd.P, mul(n, nd.W/2)), sub(nd.P, mul(n, nd.W/2))
			fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1a7f37" stroke-width="2"/>`, a.X+dx, dy-a.Y, c.X+dx, dy-c.Y)
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#1a7f37"/>`, nd.P.X+dx, dy-nd.P.Y)
		}
	}
	return b.String()
}
