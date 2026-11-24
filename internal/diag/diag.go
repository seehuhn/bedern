// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package diag draws glyph outlines for reviewing a design: the outline with its
// points, handles and the guide lines of the design.
package diag

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// Sizes of the drawing, in font units.
const (
	leftMargin  = 110.0 // space left of the origin, for point labels and the left sidebearing
	rightMargin = 170.0 // space right of the advance line, for the guide labels
	minWidth    = 700.0 // narrow glyphs are centred in a drawing of this width
	vMargin     = 70.0  // space above the highest and below the lowest feature
	footer      = 100.0 // space below the guides for the measurements and the caption
	fontSize    = 22.0  // height of the text
	guideSize   = 17.0  // height of the guide labels
	pointSize   = 12.0  // height of the coordinate labels
	markSize    = 7.0   // side of an on-curve square and diameter of a control circle
	labelGap    = 10.0  // distance of a coordinate label from its point
)

// Options are the optional parts of a drawing.
type Options struct {
	// Overlay is a glyph which is drawn as an unfilled red outline on top
	// of the main glyph.  It is ignored if it has no outline.
	Overlay *glyphs.Glyph
	// OverlayLabel says where the overlay comes from, for the caption.
	OverlayLabel string
}

// rnd rounds to one decimal, to hide the error of measuring curves.
func rnd(x float64) float64 { return math.Round(x*10) / 10 }

// SVG writes a drawing of the glyph g to w, with the variant name in the
// caption.  One font unit is one unit of the drawing, and the y axis points
// up in the font and down in the drawing, so that every y coordinate is
// negated; the text is not mirrored.
//
// The drawing shows the outline of g itself, that is the rounded outline
// which is written to the font, and not the unrounded construction.  What
// the reader sees is what the font contains, and the integer coordinates
// of the points are those of the font.
//
// The drawing shows the guide lines of the design, including the
// overshoot lines, the coordinates of the on-curve points, the sidebearings
// and the x coordinates of the extreme points.  The vertical range is
// cropped to the glyph, the baseline, the x-height and the overshoot lines,
// with a margin; the other guide lines are drawn only if they lie inside.
//
// A glyph without an outline still gets the guide lines and the caption.
func SVG(w io.Writer, g glyphs.Glyph, p design.Params, variant string, opt Options) error {
	bw := bufio.NewWriter(w)

	overlay := opt.Overlay
	if overlay != nil && len(overlay.Path) == 0 {
		overlay = nil
	}

	// The vertical range, in font units.
	hi := p.XHeight + p.OvershootLower
	lo := -p.OvershootLower
	x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	if len(g.Path) > 0 {
		x0, y0, x1, y1 = g.Path.BBox()
		x0, y0, x1, y1 = rnd(x0), rnd(y0), rnd(x1), rnd(y1)
		hi, lo = math.Max(hi, y1), math.Min(lo, y0)
	}
	if overlay != nil {
		_, oy0, _, oy1 := overlay.Path.BBox()
		hi, lo = math.Max(hi, oy1), math.Min(lo, oy0)
	}
	topFont, botFont := hi+vMargin, lo-vMargin
	top, bottom := -topFont, -botFont // in drawing coordinates

	width := leftMargin + g.Advance + rightMargin
	left := -leftMargin
	if width < minWidth {
		left -= (minWidth - width) / 2
		width = minWidth
	}
	right := left + width
	height := bottom - top + footer

	fmt.Fprintf(bw, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%g %g %g %g" width="%g" height="%g" font-family="sans-serif" font-size="%g">`+"\n",
		left, top, width, height, width, height, fontSize)
	fmt.Fprintf(bw, `<rect x="%g" y="%g" width="%g" height="%g" fill="#fff"/>`+"\n", left, top, width, height)

	// Guide lines and their labels, to the right of the advance line.  The
	// ascender and the cap height are 11 units apart, and so are the
	// x-height and its overshoot, so some labels go below their lines.
	hGuides := []struct {
		id, label string
		y         float64 // in font units
		light     bool    // overshoot line
		below     bool    // put the label below the line
	}{
		{"descender", "descender", -p.Descender, false, false},
		{"overshoot-bottom", "overshoot", -p.OvershootLower, true, true},
		{"baseline", "baseline", 0, false, false},
		{"xheight", "x-height", p.XHeight, false, true},
		{"overshoot-top", "overshoot", p.XHeight + p.OvershootLower, true, false},
		{"capheight", "cap height", p.CapHeight, false, true},
		{"ascender", "ascender", p.Ascender, false, false},
	}
	var avoid []box // regions of text which point labels must not cover
	for _, h := range hGuides {
		if h.y > topFont || h.y < botFont {
			continue
		}
		y := -h.y
		if h.light {
			fmt.Fprintf(bw, `<line class="guide light" id="%s" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#9c9" stroke-width="0.6" stroke-dasharray="2 3"/>`+"\n",
				h.id, left, y, right, y)
		} else {
			fmt.Fprintf(bw, `<line class="guide" id="%s" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#4a8" stroke-width="1" stroke-dasharray="6 4"/>`+"\n",
				h.id, left, y, right, y)
		}
		ty := y - 4
		if h.below {
			ty = y + guideSize
		}
		size, fill := guideSize, "#285"
		if h.light {
			size, fill = guideSize-3, "#6a6"
		}
		text := fmt.Sprintf("%s %g", h.label, math.Abs(h.y))
		avoid = append(avoid, box{g.Advance + 8, ty - size, g.Advance + 8 + float64(len(text))*size*0.6, ty + size/4})
		fmt.Fprintf(bw, `<text class="label guidelabel" x="%g" y="%g" font-size="%g" fill="%s">%s %g</text>`+"\n",
			g.Advance+8, ty, size, fill, h.label, math.Abs(h.y))
	}
	for _, v := range []struct {
		id string
		x  float64
	}{{"origin", 0}, {"advance", g.Advance}} {
		fmt.Fprintf(bw, `<line class="guide" id="%s" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#68c" stroke-width="1" stroke-dasharray="6 4"/>`+"\n",
			v.id, v.x, top, v.x, bottom)
	}
	// The labels of the vertical lines sit at the top, the 0 to the left
	// of its line and the advance width to the right of its line, so that
	// no text crosses a line.
	fmt.Fprintf(bw, `<text class="label" x="-6" y="%g" text-anchor="end" fill="#47a">0</text>`+"\n", top+fontSize+4)
	fmt.Fprintf(bw, `<text class="label" x="%g" y="%g" fill="#47a">advance %g</text>`+"\n", g.Advance+6, top+fontSize+4, g.Advance)

	// The filled outline.
	if len(g.Path) > 0 {
		fmt.Fprintf(bw, `<path class="outline" d="%s" fill="#ddd" fill-rule="nonzero" stroke="#222" stroke-width="2"/>`+"\n", pathData(g.Path))
	}

	// Handles first, so that the markers lie on top of them.
	for _, c := range g.Path {
		cur := c.Start
		for _, s := range c.Segs {
			if s.IsCurve {
				for _, pair := range [][2]outline.Point{{cur, s.P1}, {s.P3, s.P2}} {
					fmt.Fprintf(bw, `<line class="handle" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#d80" stroke-width="1"/>`+"\n",
						pair[0].X, -pair[0].Y, pair[1].X, -pair[1].Y)
				}
			}
			cur = s.P3
		}
	}
	// The overlay lies above the handles but below the markers.
	if overlay != nil {
		fmt.Fprintf(bw, `<path class="overlay" d="%s" fill="none" stroke="#e00" stroke-width="1.6"/>`+"\n", pathData(overlay.Path))
	}
	outer := outerSign(g.Path)
	var pts []labelled
	for _, c := range g.Path {
		cx, cy := centre(c)
		counter := c.SignedArea()*outer < 0
		seen := map[outline.Point]bool{}
		onCurve := func(q outline.Point) {
			if seen[q] {
				return
			}
			seen[q] = true
			fmt.Fprintf(bw, `<rect class="oncurve" x="%g" y="%g" width="%g" height="%g" fill="#117" stroke="#fff" stroke-width="0.8"/>`+"\n",
				q.X-markSize/2, -q.Y-markSize/2, markSize, markSize)
			pts = append(pts, labelled{q, cx, cy, counter})
		}
		onCurve(c.Start)
		for _, s := range c.Segs {
			onCurve(s.P3)
		}
		for _, s := range c.Segs {
			if s.IsCurve {
				for _, q := range []outline.Point{s.P1, s.P2} {
					fmt.Fprintf(bw, `<circle class="control" cx="%g" cy="%g" r="%g" fill="#fff" stroke="#d80" stroke-width="1.5"/>`+"\n",
						q.X, -q.Y, markSize/2)
				}
			}
		}
	}

	labelPoints(bw, pts, g.Path, avoid)

	// The sidebearings and the extreme x coordinates, below the guides.
	row1, row2 := bottom+24, bottom+50
	if len(g.Path) > 0 {
		for _, x := range []float64{x0, x1} {
			fmt.Fprintf(bw, `<line class="extent" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#888" stroke-width="0.8" stroke-dasharray="1 3"/>`+"\n",
				x, -y1, x, row2-fontSize)
		}
		dim := func(xa, xb float64) {
			fmt.Fprintf(bw, `<line class="dimension" x1="%g" y1="%g" x2="%g" y2="%g" stroke="#333" stroke-width="1.5"/>`+"\n", xa, row1, xb, row1)
		}
		dim(0, x0)
		dim(x1, g.Advance)
		fmt.Fprintf(bw, `<text class="metric" id="lsb" x="-6" y="%g" text-anchor="end" fill="#000">lsb %g</text>`+"\n", row1+6, rnd(x0))
		fmt.Fprintf(bw, `<text class="metric" id="rsb" x="%g" y="%g" fill="#000">rsb %g</text>`+"\n", g.Advance+6, row1+6, rnd(g.Advance-x1))
		fmt.Fprintf(bw, `<text class="metric" id="xmin" x="%g" y="%g" text-anchor="end" fill="#333">x %g</text>`+"\n", x0+4, row2, x0)
		fmt.Fprintf(bw, `<text class="metric" id="xmax" x="%g" y="%g" fill="#333">x %g</text>`+"\n", x1-4, row2, x1)
	}

	// The caption, at the bottom.
	caption := fmt.Sprintf("%s   advance %g   variant %s", g.Name, g.Advance, variant)
	if overlay != nil {
		caption += fmt.Sprintf("   overlay (red): %s %s", opt.OverlayLabel, overlay.Name)
	}
	fmt.Fprintf(bw, `<text class="caption" x="%g" y="%g" font-size="%g" fill="#000">%s</text>`+"\n",
		left+6, top+height-14, fontSize+2, escape(caption))
	fmt.Fprintln(bw, `</svg>`)
	return bw.Flush()
}

// outerSign returns the sign of the signed area of the largest contour,
// which is the outer contour of the glyph; contours with the opposite sign
// are counters.
func outerSign(path outline.Path) float64 {
	var best, sign float64
	for _, c := range path {
		a := c.SignedArea()
		if math.Abs(a) > best {
			best, sign = math.Abs(a), a
		}
	}
	return sign
}

// centre returns the centre of the bounding box of a contour.
func centre(c outline.Contour) (x, y float64) {
	x0, y0, x1, y1 := outline.Path{c}.BBox()
	return (x0 + x1) / 2, (y0 + y1) / 2
}

// A labelled is an on-curve point with the centre of its contour, and
// whether the contour is a counter.
type labelled struct {
	q       outline.Point
	cx, cy  float64
	counter bool
}

// A box is a rectangle in drawing coordinates.
type box struct{ x0, y0, x1, y1 float64 }

func (a box) overlaps(b box) bool {
	return a.x0 < b.x1 && b.x0 < a.x1 && a.y0 < b.y1 && b.y0 < a.y1
}

// labelPoints writes the coordinates of the on-curve points next to them.
// A label is first moved away from the centre of its contour, so that it
// lies outside the glyph; for a counter it is moved towards the centre,
// which again is the empty side.  If the label would overlap a marker or
// an earlier label or a guide label (avoid), the other positions around the point are tried, then
// positions further away, and the first free one is taken.
func labelPoints(w io.Writer, pts []labelled, path outline.Path, avoid []box) {
	const tol = 4.0 // points this close to the centre line are not moved sideways
	taken := slices.Clone(avoid)
	half := markSize/2 + 3
	for _, pt := range pts {
		taken = append(taken, box{pt.q.X - half, -pt.q.Y - half, pt.q.X + half, -pt.q.Y + half})
	}
	for _, c := range path {
		for _, s := range c.Segs {
			if s.IsCurve {
				for _, q := range []outline.Point{s.P1, s.P2} {
					taken = append(taken, box{q.X - half, -q.Y - half, q.X + half, -q.Y + half})
				}
			}
		}
	}

	for _, pt := range pts {
		q := pt.q
		sx, sy := 0, 0
		if q.X-pt.cx > tol {
			sx = 1
		} else if pt.cx-q.X > tol {
			sx = -1
		}
		if q.Y-pt.cy > tol {
			sy = 1
		} else if pt.cy-q.Y > tol {
			sy = -1
		}
		if pt.counter {
			sx, sy = -sx, -sy
		}
		if sx == 0 && sy == 0 {
			sy = 1
		}
		text := fmt.Sprintf("%g,%g", q.X, q.Y)
		textW := float64(len(text)) * pointSize * 0.58

		// place returns the text position and box for a direction.
		place := func(dx, dy int, push float64) (x, y float64, anchor string, b box) {
			anchor, x = "middle", q.X
			switch dx {
			case 1:
				anchor, x = "start", q.X+labelGap
			case -1:
				anchor, x = "end", q.X-labelGap
			}
			y = -q.Y + pointSize/3 // centred on the point
			switch dy {
			case 1:
				y = -q.Y - labelGap - push
			case -1:
				y = -q.Y + labelGap + pointSize + push
			}
			x0 := x - textW/2
			switch anchor {
			case "start":
				x0 = x
			case "end":
				x0 = x - textW
			}
			return x, y, anchor, box{x0, y - pointSize*0.85, x0 + textW, y + pointSize*0.2}
		}

		// The preferred direction, then the others by distance from it.
		dirs := [][2]int{{sx, sy}}
		for _, d := range [][2]int{{sx, -sy}, {-sx, sy}, {-sx, -sy}, {1, 1}, {-1, 1}, {1, -1}, {-1, -1}, {0, 1}, {0, -1}} {
			if d != [2]int{0, 0} {
				dirs = append(dirs, d)
			}
		}
		var bx, by float64
		var banchor string
		var bb box
		found := false
	search:
		for _, push := range []float64{0, 14, 28} {
			for _, d := range dirs {
				x, y, anchor, b := place(d[0], d[1], push)
				free := true
				for _, t := range taken {
					if b.overlaps(t) {
						free = false
						break
					}
				}
				if free {
					bx, by, banchor, bb, found = x, y, anchor, b, true
					break search
				}
			}
		}
		if !found { // nowhere free: use the preferred place
			bx, by, banchor, bb = place(sx, sy, 0)
		}
		taken = append(taken, bb)
		fmt.Fprintf(w, `<text class="ptlabel" x="%g" y="%g" text-anchor="%s" font-size="%g" fill="#117">%s</text>`+"\n",
			bx, by, banchor, pointSize, text)
	}
}

// pathData returns the contours as SVG path data, with y negated.
func pathData(path outline.Path) string {
	var sb strings.Builder
	for _, c := range path {
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
