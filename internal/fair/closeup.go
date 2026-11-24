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
)

// Sizes of the close-up sheet, in pixels.
const (
	sheetBig    = 760.0 // height of the large glyph
	cropUnits   = 170.0 // side of a crop, in font units
	cropPx      = 340.0 // side of a crop on the sheet
	cropsInRow  = 4
	terminalDeg = 150.0 // the edge turns by this much ...
	terminalLen = 130.0 // ... within this arc length at a terminal
	joinRatio   = 1.7   // the thickness changes by this factor ...
	joinLen     = 70.0  // ... within this arc length where a thin part meets a thick one
	mergeDist   = 50.0  // places closer than this share a crop
)

// Place is a part of a glyph which is shown large.
type Place struct {
	X, Y  float64
	What  []string
	Score float64 // for the order: findings first
}

// Places returns the corners, terminals, the places where a thin part meets
// a thick one, and the findings of the glyph, those close together merged.
func (a *Analysis) Places() []Place {
	var ps []Place
	add := func(x, y float64, what string, score float64) {
		for i := range ps {
			if math.Hypot(ps[i].X-x, ps[i].Y-y) < mergeDist {
				for _, w := range ps[i].What {
					if w == what {
						return
					}
				}
				ps[i].What = append(ps[i].What, what)
				ps[i].Score += score
				return
			}
		}
		ps = append(ps, Place{X: x, Y: y, What: []string{what}, Score: score})
	}
	for _, f := range a.Findings {
		add(f.X, f.Y, fmt.Sprintf("%d %s", f.Number, f.Kind), 10)
	}
	for _, c := range a.Contours {
		// Terminals: the tangent turns by terminalDeg within terminalLen.
		ss := c.Samples
		if len(ss) > 0 && c.Length > 3*terminalLen {
			theta := make([]float64, len(ss))
			for i, s := range ss {
				theta[i] = math.Atan2(s.T.Y, s.T.X)
				if i > 0 {
					d := theta[i] - theta[i-1]
					d -= 2 * math.Pi * math.Round(d/(2*math.Pi))
					theta[i] = theta[i-1] + d
				}
			}
			j := 0
			best, bestI := 0.0, -1
			for i := range ss {
				for j < len(ss)-1 && ss[j].S-ss[i].S < terminalLen {
					j++
				}
				turn := (theta[j] - theta[i]) * 180 / math.Pi
				if turn >= terminalDeg {
					if turn > best {
						best, bestI = turn, (i+j)/2
					}
				} else if bestI >= 0 {
					add(ss[bestI].P.X, ss[bestI].P.Y, "terminal", 2)
					best, bestI = 0, -1
				}
			}
			if bestI >= 0 {
				add(ss[bestI].P.X, ss[bestI].P.Y, "terminal", 2)
			}
		}
		for _, nd := range c.Nodes {
			if nd.Corner {
				add(nd.P.X, nd.P.Y, "corner", 1)
			}
		}
		// Thin meets thick: the thickness changes by joinRatio within joinLen.
		ts := c.Thick
		for i := range ts {
			if !ts[i].Valid {
				continue
			}
			for j := i + 1; j < len(ts) && ts[j].S-ts[i].S <= joinLen; j++ {
				if !ts[j].Valid {
					break
				}
				lo, hi := min(ts[i].T, ts[j].T), max(ts[i].T, ts[j].T)
				if hi >= joinRatio*lo {
					m := ts[(i+j)/2]
					add(m.P.X, m.P.Y, "thin meets thick", 1)
					break
				}
			}
		}
	}
	sort.SliceStable(ps, func(i, j int) bool {
		if (ps[i].Score >= 10) != (ps[j].Score >= 10) {
			return ps[i].Score >= 10
		}
		if ps[i].Y != ps[j].Y {
			return ps[i].Y > ps[j].Y
		}
		return ps[i].X < ps[j].X
	})
	return ps
}

// Closeup writes the close-up sheet of the glyph as SVG: the glyph large
// with its comb and its numbered findings, beside it the glyph cmp (the
// same glyph in another font, for example main's) when it is not nil, and
// below, magnified, every place of Places with the comb.  The sheet leaves
// space at the bottom of the height extra for the lines of text which the
// caller adds.
func (a *Analysis) Closeup(w io.Writer, caption string, cmp *Analysis, cmpLabel string) error {
	bw := bufio.NewWriter(w)
	x0, y0, x1, y1 := a.bounds()
	if cmp != nil {
		cx0, cy0, cx1, cy1 := cmp.bounds()
		x0, y0, x1, y1 = min(x0, cx0), min(y0, cy0), max(x1, cx1), max(y1, cy1)
	}
	pad := 140.0
	vw, vh := x1-x0+2*pad, y1-y0+2*pad
	scale := sheetBig / vh
	bigW := vw * scale
	places := a.Places()
	rows := (len(places) + cropsInRow - 1) / cropsInRow
	top := 50.0
	cropTop := top + sheetBig + 50
	listTop := cropTop + float64(rows)*(cropPx+40) + 20
	width := max(2*bigW+60, float64(cropsInRow)*(cropPx+20)+20)
	height := listTop + float64(len(a.Findings)+2)*listLine + 20

	fmt.Fprintf(bw, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" font-family="sans-serif" font-size="14">`+"\n",
		width, height, width, height)
	fmt.Fprintf(bw, `<rect width="%.0f" height="%.0f" fill="#fff"/>`+"\n", width, height)
	fmt.Fprintln(bw, `<defs><g id="glyph">`)
	a.writeLayer(bw, layerOptions{comb: true, nodes: true, findings: true, labelPx: 13})
	fmt.Fprintln(bw, `</g>`)
	if cmp != nil {
		fmt.Fprintln(bw, `<g id="cmp">`)
		cmp.writeLayer(bw, layerOptions{comb: true, findings: true, labelPx: 13})
		fmt.Fprintln(bw, `</g>`)
	}
	fmt.Fprintln(bw, `</defs>`)
	fmt.Fprintf(bw, `<text x="10" y="30" font-size="20">%s</text>`+"\n", escape(caption))

	vb := fmt.Sprintf("%.1f %.1f %.1f %.1f", x0-pad, -y1-pad, vw, vh)
	fmt.Fprintf(bw, `<svg x="10" y="%.0f" width="%.0f" height="%.0f" viewBox="%s"><use href="#glyph" xlink:href="#glyph"/></svg>`+"\n", top, bigW, sheetBig, vb)
	fmt.Fprintf(bw, `<text x="14" y="%.0f" font-size="15">this branch</text>`+"\n", top+16)
	if cmp != nil {
		fmt.Fprintf(bw, `<svg x="%.0f" y="%.0f" width="%.0f" height="%.0f" viewBox="%s"><use href="#cmp" xlink:href="#cmp"/></svg>`+"\n", bigW+40, top, bigW, sheetBig, vb)
		fmt.Fprintf(bw, `<text x="%.0f" y="%.0f" font-size="15">%s</text>`+"\n", bigW+44, top+16, escape(cmpLabel))
	}
	// The crops, and on the large glyph a box for each.
	for i, p := range places {
		col, row := i%cropsInRow, i/cropsInRow
		cx := 10 + float64(col)*(cropPx+20)
		cy := cropTop + float64(row)*(cropPx+40)
		h := cropUnits / 2
		what := strings.Join(p.What, ", ")
		if len(what) > 46 {
			what = what[:44] + "..."
		}
		fmt.Fprintf(bw, `<text x="%.0f" y="%.0f" font-size="13" font-weight="bold">P%d</text><text x="%.0f" y="%.0f" font-size="12">(%.0f, %.0f) %s</text>`+"\n",
			cx, cy+14, i+1, cx+30, cy+14, p.X, p.Y, escape(what))
		fmt.Fprintf(bw, `<svg x="%.0f" y="%.0f" width="%.0f" height="%.0f" viewBox="%.1f %.1f %.1f %.1f"><use href="#glyph" xlink:href="#glyph"/></svg>`+"\n",
			cx, cy+20, cropPx, cropPx, p.X-h, -p.Y-h, cropUnits, cropUnits)
		fmt.Fprintf(bw, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="none" stroke="#888"/>`+"\n", cx, cy+20, cropPx, cropPx)
		bx := 10 + (p.X-h-(x0-pad))*scale
		by := top + (-p.Y-h-(-y1-pad))*scale
		fmt.Fprintf(bw, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#0a0" stroke-width="1"/><text x="%.1f" y="%.1f" font-size="12" fill="#0a0">P%d</text>`+"\n",
			bx, by, cropUnits*scale, cropUnits*scale, bx+2, by+12, i+1)
	}
	y := listTop
	fmt.Fprintf(bw, `<text x="10" y="%.0f" font-size="15" font-weight="bold">%d findings</text>`+"\n", y, len(a.Findings))
	y += listLine + 2
	for _, f := range a.Findings {
		fmt.Fprintf(bw, `<text x="10" y="%.0f" font-size="13" fill="%s">%d  %s at (%.0f, %.0f), edge %s: %s</text>`+"\n",
			y, colFinding, f.Number, f.Kind, f.X, f.Y, Compass(f.Dir), escape(f.Detail))
		y += listLine
	}
	fmt.Fprintln(bw, `</svg>`)
	return bw.Flush()
}
