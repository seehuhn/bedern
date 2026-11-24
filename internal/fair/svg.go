// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"seehuhn.de/go/bedern/internal/outline"
)

// Sizes of the picture.  The glyph is drawn at one pixel per font unit; the
// plots are in pixels.
const (
	combMedian = 30.0  // length of the tooth of the comb at the median curvature of the curves of the glyph
	combMax    = 110.0 // longest tooth; a longer one is cut and its tip drawn red
	combStep   = 3.0   // arc length between two teeth
	glyphPad   = 190.0 // space round the glyph, for the comb and the labels
	plotW      = 560.0 // width of a plot of curvature
	plotKH     = 96.0  // height of the curvature plot
	plotTH     = 44.0  // height of the thickness plot
	plotGap    = 34.0  // space above a plot for its title
	plotsInCol = 8     // plots in one column
	plotMaxLen = 420.0 // a run longer than this is shown in several plots
	listLine   = 17.0  // line height of the list of findings
)

// Colours.
const (
	colConvex  = "#1f5fbf"
	colConcave = "#d9730d"
	colFinding = "#d00"
	colNote    = "#8a2be2"
	colNode    = "#117"
	colThick   = "#2a8a3a"
)

// view maps font units to the drawing: x stays, y is negated.
func vx(p Point) float64 { return p.X }
func vy(p Point) float64 { return -p.Y }

// toothEnd returns the tip of the tooth of the comb at s, and whether it was
// cut at combMax.
func toothEnd(s Sample, combScale float64) (Point, bool) {
	l := math.Abs(s.K) * combScale
	cut := l > combMax
	l = min(l, combMax)
	out := Point{X: s.T.Y, Y: -s.T.X} // away from the ink
	return add(s.P, mul(out, l)), cut
}

// CombScale returns the length of a tooth of the comb per unit of
// curvature: the median curvature of the curved stretches of the glyph gives
// a tooth of combMedian units.  One glyph has one scale, so that the comb can
// be compared along the outline.
func (a *Analysis) CombScale() float64 {
	if a.combScale > 0 {
		return a.combScale
	}
	var ks []float64
	for _, r := range a.Runs {
		for _, s := range r.Samples {
			if k := math.Abs(s.K); k > 0.0005 {
				ks = append(ks, k)
			}
		}
	}
	a.combScale = 4000
	if len(ks) > 0 {
		sort.Float64s(ks)
		a.combScale = min(20000, max(1000, combMedian/ks[len(ks)/2]))
	}
	return a.combScale
}

// labelPos returns a place for the label of a finding: outside the ink,
// beyond the comb.
func (a *Analysis) labelPos(f Finding) Point {
	t := Point{X: math.Cos(f.Dir * math.Pi / 180), Y: math.Sin(f.Dir * math.Pi / 180)}
	out := Point{X: t.Y, Y: -t.X}
	l := 0.0
	c := a.Contours[f.Contour]
	for _, s := range c.Samples {
		if math.Abs(s.S-f.S) < 6 {
			l = max(l, min(math.Abs(s.K)*a.CombScale(), combMax))
		}
	}
	return add(Point{X: f.X, Y: f.Y}, mul(out, l+22))
}

// layerOptions select what the glyph layer shows.
type layerOptions struct {
	comb     bool
	nodes    bool
	runs     bool
	findings bool
	notes    bool
	labelPx  float64 // height of the labels, in font units
}

// writeLayer writes the glyph with its comb, nodes, runs and findings, in
// font units with y negated.
func (a *Analysis) writeLayer(w io.Writer, o layerOptions) {
	fmt.Fprintf(w, `<path class="outline" d="%s" fill="#e6e6e6" fill-rule="nonzero" stroke="#222" stroke-width="1.2"/>`+"\n", pathData(a.Path))
	if o.comb {
		for _, r := range a.Runs {
			if r.Straight {
				continue
			}
			var tips []string
			next := math.Inf(-1)
			for i, s := range r.Samples {
				end, cut := toothEnd(s, a.CombScale())
				tips = append(tips, fmt.Sprintf("%.1f,%.1f", vx(end), vy(end)))
				if s.S < next && i != len(r.Samples)-1 {
					continue
				}
				next = s.S + combStep
				col := colConvex
				if s.K < 0 {
					col = colConcave
				}
				if math.Abs(s.K)*a.CombScale() < 0.5 {
					continue
				}
				fmt.Fprintf(w, `<line class="tooth" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="0.7" stroke-opacity="0.55"/>`+"\n",
					vx(s.P), vy(s.P), vx(end), vy(end), col)
				if cut {
					fmt.Fprintf(w, `<circle cx="%.1f" cy="%.1f" r="2" fill="%s"/>`+"\n", vx(end), vy(end), colFinding)
				}
			}
			fmt.Fprintf(w, `<polyline class="envelope" points="%s" fill="none" stroke="#333" stroke-width="1"/>`+"\n", strings.Join(tips, " "))
		}
	}
	if o.nodes {
		for _, c := range a.Contours {
			for _, s := range c.Segs {
				if s.Curve {
					for _, pr := range [][2]Point{{s.P0, s.P1}, {s.P3, s.P2}} {
						fmt.Fprintf(w, `<line class="handle" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#d80" stroke-width="0.6" stroke-opacity="0.7"/>`+"\n",
							vx(pr[0]), vy(pr[0]), vx(pr[1]), vy(pr[1]))
					}
				}
			}
			for _, nd := range c.Nodes {
				if nd.Corner {
					fmt.Fprintf(w, `<rect class="corner" x="%g" y="%g" width="6" height="6" fill="#000" transform="rotate(45 %g %g)"/>`+"\n",
						vx(nd.P)-3, vy(nd.P)-3, vx(nd.P), vy(nd.P))
				} else {
					fmt.Fprintf(w, `<rect class="node" x="%g" y="%g" width="5" height="5" fill="#fff" stroke="%s" stroke-width="1.2"/>`+"\n",
						vx(nd.P)-2.5, vy(nd.P)-2.5, colNode)
				}
			}
		}
	}
	lp := o.labelPx
	if lp == 0 {
		lp = 16
	}
	if o.runs {
		for _, r := range a.Runs {
			if r.Straight || len(r.Samples) == 0 {
				continue
			}
			mid := r.Samples[len(r.Samples)/2]
			in := Point{X: -mid.T.Y, Y: mid.T.X}
			q := add(mid.P, mul(in, 10))
			fmt.Fprintf(w, `<text class="runlabel" x="%.1f" y="%.1f" font-size="%g" fill="#555" text-anchor="middle" dominant-baseline="middle">R%d</text>`+"\n",
				vx(q), vy(q), lp*0.7, r.Number)
		}
	}
	mark := func(f Finding, col, label string) {
		fmt.Fprintf(w, `<circle class="finding" cx="%.1f" cy="%.1f" r="9" fill="none" stroke="%s" stroke-width="2"/>`+"\n", f.X, -f.Y, col)
		q := a.labelPos(f)
		fmt.Fprintf(w, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="0.8"/>`+"\n", f.X, -f.Y, vx(q), vy(q), col)
		fmt.Fprintf(w, `<text class="findlabel" x="%.1f" y="%.1f" font-size="%g" font-weight="bold" fill="%s" stroke="#fff" stroke-width="3" paint-order="stroke" text-anchor="middle" dominant-baseline="middle">%s</text>`+"\n",
			vx(q), vy(q), lp, col, label)
	}
	if o.notes {
		for i, f := range a.Notes {
			mark(f, colNote, fmt.Sprintf("t%d", i+1))
		}
	}
	if o.findings {
		for _, f := range a.Findings {
			mark(f, colFinding, fmt.Sprint(f.Number))
		}
	}
}

// runOf returns the index of the run which holds the arc length s of the
// contour ci, and s in the arc length of the run.
func (a *Analysis) runOf(ci int, s float64) (int, float64) {
	l := a.Contours[ci].Length
	for i, r := range a.Runs {
		if r.Contour != ci || len(r.Samples) == 0 {
			continue
		}
		s0, s1 := r.Samples[0].S, r.Samples[len(r.Samples)-1].S
		for _, off := range []float64{0, l} {
			if t := s + off; t >= s0-1e-6 && t <= s1+1e-6 {
				return i, t
			}
		}
	}
	return -1, 0
}

// bounds returns the bounding box of the outline in font units.
func (a *Analysis) bounds() (x0, y0, x1, y1 float64) {
	x0, y0, x1, y1 = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, c := range a.Contours {
		for _, s := range c.Samples {
			x0, x1 = min(x0, s.P.X), max(x1, s.P.X)
			y0, y1 = min(y0, s.P.Y), max(y1, s.P.Y)
		}
	}
	if math.IsInf(x0, 1) {
		return 0, 0, 500, 500
	}
	return
}

// Picture writes the picture of the analysis as SVG: the glyph with its
// nodes, its curvature comb (teeth at right angles to the edge, on the side
// away from the ink, blue where the ink is convex and orange where it is
// concave, of a length proportional to the curvature) and its findings
// numbered in red; beside it, for every run which is not a straight segment,
// the curvature against the arc length with the nodes (grey) and the
// findings (red), and under it the thickness of the ink; below, the list of
// findings.  Pinches and swellings of the thickness which are not findings
// are marked in purple as t1, t2 and so on.
func (a *Analysis) Picture(w io.Writer, caption string) error {
	bw := bufio.NewWriter(w)
	x0, y0, x1, y1 := a.bounds()
	gw := x1 - x0 + 2*glyphPad
	gh := y1 - y0 + 2*glyphPad
	var plots []plotPiece
	for i, r := range a.Runs {
		if r.Straight || len(r.Samples) < 2 {
			continue
		}
		s0, l := r.Samples[0].S, r.Length()
		n := int(math.Ceil(l / plotMaxLen))
		for k := 0; k < n; k++ {
			plots = append(plots, plotPiece{i, s0 + l*float64(k)/float64(n), s0 + l*float64(k+1)/float64(n), k + 1, n})
		}
	}
	cols := (len(plots) + plotsInCol - 1) / plotsInCol
	rows := min(len(plots), plotsInCol)
	plotH := plotGap + plotKH + 8 + plotTH + 18
	width := gw + float64(cols)*(plotW+60) + 20
	top := 40.0
	listTop := max(top+gh, top+float64(rows)*plotH) + 20
	nList := len(a.Findings) + len(a.Notes) + 3
	height := listTop + float64(nList)*listLine + 20
	width = max(width, 900)

	fmt.Fprintf(bw, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" font-family="sans-serif" font-size="14">`+"\n",
		width, height, width, height)
	fmt.Fprintf(bw, `<rect width="%.0f" height="%.0f" fill="#fff"/>`+"\n", width, height)
	fmt.Fprintf(bw, `<text x="10" y="26" font-size="20" fill="#000">%s</text>`+"\n", escape(caption))

	// The glyph.
	fmt.Fprintf(bw, `<g transform="translate(%.1f %.1f)">`+"\n", glyphPad-x0, top+glyphPad+y1)
	a.writeLayer(bw, layerOptions{comb: true, nodes: true, runs: true, findings: true, notes: true})
	fmt.Fprintln(bw, `</g>`)

	// The plots.
	for k, pc := range plots {
		col, row := k/plotsInCol, k%plotsInCol
		px := gw + 40 + float64(col)*(plotW+60)
		py := top + float64(row)*plotH
		a.writePlot(bw, pc, px, py)
	}

	// The list of findings.
	y := listTop
	fmt.Fprintf(bw, `<text x="10" y="%.0f" font-size="15" font-weight="bold">%d findings (red) and %d notes on the thickness (purple)</text>`+"\n", y, len(a.Findings), len(a.Notes))
	y += listLine + 2
	line := func(col, label string, f Finding) {
		ri, _ := a.runOf(f.Contour, f.S)
		fmt.Fprintf(bw, `<text x="10" y="%.0f" font-size="13" fill="%s">%s  %s at (%.0f, %.0f), R%d, edge %s: %s</text>`+"\n",
			y, col, label, f.Kind, f.X, f.Y, ri+1, Compass(f.Dir), escape(f.Detail))
		y += listLine
	}
	for _, f := range a.Findings {
		line(colFinding, fmt.Sprint(f.Number), f)
	}
	for i, f := range a.Notes {
		line(colNote, fmt.Sprintf("t%d", i+1), f)
	}
	fmt.Fprintln(bw, `</svg>`)
	return bw.Flush()
}

// plotPiece is a part of a run which one plot shows.
type plotPiece struct {
	run         int
	s0, s1      float64
	part, parts int
}

// writePlot writes the plot of curvature and thickness of a piece of a run
// at (px, py).
func (a *Analysis) writePlot(w io.Writer, pc plotPiece, px, py float64) {
	ri := pc.run
	r := a.Runs[ri]
	s0, s1 := pc.s0, pc.s1
	var ss []Sample
	for _, s := range r.Samples {
		if s.S >= s0-1e-9 && s.S <= s1+1e-9 {
			ss = append(ss, s)
		}
	}
	if len(ss) < 2 {
		return
	}
	length := max(s1-s0, 1)
	xs := func(s float64) float64 { return px + (s-s0)/length*plotW }

	// The range of curvature: the 98th percentile of |K| on either side,
	// so that a spike does not flatten the rest; values beyond are drawn at
	// the edge in red.
	var pos, neg []float64
	for _, s := range ss {
		if s.K > 0 {
			pos = append(pos, s.K)
		} else if s.K < 0 {
			neg = append(neg, -s.K)
		}
	}
	q := func(v []float64) float64 {
		if len(v) == 0 {
			return 0
		}
		sort.Float64s(v)
		return v[int(0.98*float64(len(v)-1))]
	}
	hi, lo := q(pos)*1.15, -q(neg)*1.15
	if hi-lo < 0.002 {
		hi = max(hi, 0.001)
		lo = min(lo, -0.001)
	}
	if hi < 0.0002 {
		hi = (hi - lo) * 0.1
	}
	if lo > -0.0002 {
		lo = -(hi - lo) * 0.1
	}
	ky := func(k float64) (float64, bool) {
		cut := false
		if k > hi {
			k, cut = hi, true
		}
		if k < lo {
			k, cut = lo, true
		}
		return py + plotGap + (hi-k)/(hi-lo)*plotKH, cut
	}
	first := ss[0]
	part := ""
	if pc.parts > 1 {
		part = fmt.Sprintf(" part %d of %d", pc.part, pc.parts)
	}
	fmt.Fprintf(w, `<text x="%.0f" y="%.0f" font-size="14" fill="#000">R%d%s  from (%.0f, %.0f)  length %.0f  curvature %.4f to %.4f</text>`+"\n",
		px, py+plotGap-10, r.Number, part, first.P.X, first.P.Y, length, lo/1.15, hi/1.15)
	fmt.Fprintf(w, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#fafafa" stroke="#999" stroke-width="0.8"/>`+"\n", px, py+plotGap, plotW, plotKH)
	zy, _ := ky(0)
	fmt.Fprintf(w, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#888" stroke-width="0.8"/>`+"\n", px, zy, px+plotW, zy)
	fmt.Fprintf(w, `<text x="%.0f" y="%.1f" font-size="11" fill="#666" text-anchor="end">0</text>`+"\n", px-4, zy+4)
	// Nodes.
	c := a.Contours[r.Contour]
	for _, ni := range r.Nodes {
		nd := c.Nodes[ni]
		for _, off := range []float64{0, c.Length} {
			if s := nd.S + off; s > s0 && s < s1 {
				x := xs(s)
				fmt.Fprintf(w, `<line x1="%.1f" y1="%.0f" x2="%.1f" y2="%.0f" stroke="#aaa" stroke-width="0.8" stroke-dasharray="3 2"/>`+"\n",
					x, py+plotGap, x, py+plotGap+plotKH+8+plotTH)
			}
		}
	}
	// The curvature, in pieces of one segment each, so that a jump at a node
	// is drawn as a vertical step.
	var pts []string
	flush := func() {
		if len(pts) > 1 {
			fmt.Fprintf(w, `<polyline points="%s" fill="none" stroke="%s" stroke-width="1.6"/>`+"\n", strings.Join(pts, " "), colConvex)
		}
		pts = pts[:0]
	}
	prevSeg := -1
	var prevPt string
	for _, s := range ss {
		if s.Seg != prevSeg && prevSeg >= 0 {
			flush()
			// The step at the node.
			y, _ := ky(s.K)
			fmt.Fprintf(w, `<polyline points="%s %.1f,%.1f" fill="none" stroke="%s" stroke-width="1.6"/>`+"\n", prevPt, xs(s.S), y, colConvex)
		}
		prevSeg = s.Seg
		y, cut := ky(s.K)
		prevPt = fmt.Sprintf("%.1f,%.1f", xs(s.S), y)
		pts = append(pts, prevPt)
		if cut {
			fmt.Fprintf(w, `<circle cx="%.1f" cy="%.1f" r="1.6" fill="%s"/>`+"\n", xs(s.S), y, colFinding)
		}
	}
	flush()
	// The thickness.
	ty0 := py + plotGap + plotKH + 8
	fmt.Fprintf(w, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#f6fbf6" stroke="#999" stroke-width="0.8"/>`+"\n", px, ty0, plotW, plotTH)
	var ts []Thick
	for _, t := range c.Thick {
		for _, off := range []float64{0, c.Length} {
			if s := t.S + off; s >= s0 && s <= s1 {
				t.S = s
				ts = append(ts, t)
			}
		}
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].S < ts[j].S })
	tmax := 1.0
	for _, t := range ts {
		if t.Valid {
			tmax = max(tmax, t.T)
		}
	}
	tmax *= 1.1
	pts = pts[:0]
	for _, t := range ts {
		if !t.Valid {
			if len(pts) > 1 {
				fmt.Fprintf(w, `<polyline points="%s" fill="none" stroke="%s" stroke-width="1.4"/>`+"\n", strings.Join(pts, " "), colThick)
			}
			pts = pts[:0]
			continue
		}
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", xs(t.S), ty0+plotTH-t.T/tmax*plotTH))
	}
	if len(pts) > 1 {
		fmt.Fprintf(w, `<polyline points="%s" fill="none" stroke="%s" stroke-width="1.4"/>`+"\n", strings.Join(pts, " "), colThick)
	}
	fmt.Fprintf(w, `<text x="%.0f" y="%.0f" font-size="11" fill="%s" text-anchor="end">%.0f</text>`+"\n", px-4, ty0+10, colThick, tmax)
	fmt.Fprintf(w, `<text x="%.0f" y="%.0f" font-size="11" fill="#666">thickness</text>`+"\n", px+plotW+4, ty0+12)
	fmt.Fprintf(w, `<text x="%.0f" y="%.0f" font-size="11" fill="#666">curvature</text>`+"\n", px+plotW+4, py+plotGap+12)
	// The findings and notes.
	marks := func(fs []Finding, col string, label func(i int, f Finding) string) {
		for i, f := range fs {
			fi, s := a.runOf(f.Contour, f.S)
			if fi != ri {
				continue
			}
			// The extension of a cyclic run holds its start twice.
			for _, t := range []float64{s, s + c.Length} {
				if t < s0 || t > s1 || (t != s && !r.Cyclic) {
					continue
				}
				x := xs(t)
				fmt.Fprintf(w, `<line x1="%.1f" y1="%.0f" x2="%.1f" y2="%.0f" stroke="%s" stroke-width="1.2"/>`+"\n", x, py+plotGap, x, py+plotGap+plotKH+8+plotTH, col)
				fmt.Fprintf(w, `<text x="%.1f" y="%.0f" font-size="13" font-weight="bold" fill="%s" text-anchor="middle">%s</text>`+"\n", x, py+plotGap+plotKH+8+plotTH+14, col, label(i, f))
			}
		}
	}
	marks(a.Notes, colNote, func(i int, f Finding) string { return fmt.Sprintf("t%d", i+1) })
	marks(a.Findings, colFinding, func(i int, f Finding) string { return fmt.Sprint(f.Number) })
}

// pathData returns the contours as SVG path data, with y negated.
func pathData(p outline.Path) string {
	var sb strings.Builder
	for _, c := range p {
		fmt.Fprintf(&sb, "M%g %g", c.Start.X, -c.Start.Y)
		for _, s := range c.Segs {
			if s.IsCurve {
				fmt.Fprintf(&sb, " C%g %g %g %g %g %g", s.P1.X, -s.P1.Y, s.P2.X, -s.P2.Y, s.P3.X, -s.P3.Y)
			} else {
				fmt.Fprintf(&sb, " L%g %g", s.P3.X, -s.P3.Y)
			}
		}
		sb.WriteString(" Z")
	}
	return sb.String()
}

func escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
