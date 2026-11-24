// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/outline"
)

// unsheared returns the outline in the upright description.
func unsheared(p Params, path outline.Path) outline.Path {
	f := func(q Point) Point { return p.unshear(q) }
	var out outline.Path
	for _, c := range path {
		d := outline.Contour{Start: f(c.Start)}
		for _, s := range c.Segs {
			d.Segs = append(d.Segs, outline.Segment{IsCurve: s.IsCurve, P1: f(s.P1), P2: f(s.P2), P3: f(s.P3)})
		}
		out = append(out, d)
	}
	return out
}

// An entry is the left edge of the stem which a letter's entry starts, and
// how far its top stands above the x-height.
type entry struct{ edge, up float64 }

// stemLetters are the letters with an entry and their entries, and those
// with an exit and the left edge of the stem it ends.
func stemLetters(p Params) (entries map[string]entry, exits map[string]float64) {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.NCounter
	m3 := e1 + 2*(p.Stem+p.MCounter)
	h1 := e1 - p.HLeftLess
	asc := p.AscenderEntryRise
	entries = map[string]entry{"n": {e1, 0}, "i": {e1, 0}, "u": {e1, 0}, "m": {e1, 0}, "h": {h1, asc}, "l": {e1, asc}}
	exits = map[string]float64{"n": e2, "i": e1, "u": e2, "m": m3, "h": h1 + p.Stem + p.NCounter, "l": e1, "t": p.TSideLeft}
	exits["a"], exits["d"] = p.bowlStemEdge(), p.bowlStemEdge()
	entries["d"] = entry{p.bowlStemEdge(), asc}
	entries["j"], entries["y"] = entry{p.JSideLeft, 0}, entry{e1, 0}
	entries["v"], entries["w"], entries["k"] = entry{e1, 0}, entry{e1, 0}, entry{h1, asc}
	exits["k"] = h1 + p.Stem + p.KLegCounter
	return entries, exits
}

// TestEntry asks of every entry: the opening between its start and the stem,
// measured horizontally at the height of the centre of the start in the
// upright description, is at least 40; the
// top of the hook stands at the top of o; and the concave inner edge under
// the hook turns with no radius below 10.
func TestEntry(t *testing.T) {
	p := Default()
	entries, _ := stemLetters(p)
	for name, en := range entries {
		edge := en.edge
		g := glyph(t, name)
		up := unsheared(p, g.Path)
		y := p.EntryLow + en.up + 0.5
		rs := runs(up, y)
		var gap float64 = -1
		for i := 0; i+1 < len(rs); i++ {
			if rs[i][1] < edge && rs[i+1][0] > edge-p.Stem/2 {
				gap = rs[i+1][0] - rs[i][1]
			}
		}
		t.Logf("%s: entry opening %.0f", name, gap)
		if gap < 40 {
			t.Errorf("%s: the opening under the entry is %.1f at %.0f, want at least 40 (runs %v)", name, gap, y, rs)
		}
		y1 := math.Inf(-1)
		for _, c := range g.Path {
			if _, _, _, t := bbox(outline.Path{c}); t < p.XHeight+en.up+50 {
				y1 = math.Max(y1, t)
			}
		}
		if top := p.XHeight + en.up + p.Overshoot; math.Abs(y1-top) > 1 {
			t.Errorf("%s: the entry rises to %.1f, want the top of o, %v", name, y1, top)
		}
		// The inner edge: the concave places of the outline left of the
		// stem's left edge and above the start.
		a := fair.Analyse(name, g.Path, fair.Italic)
		for _, c := range a.Contours {
			for _, s := range c.Samples {
				q := p.unshear(s.P)
				if q.Y < p.EntryLow+en.up-p.Thin || q.Y > p.XHeight+en.up+20 || q.X > edge+1 || s.K >= 0 {
					continue
				}
				if r := 1 / -s.K; r < 10 {
					t.Errorf("%s: the inner edge of the entry turns with a radius of %.1f at %.0f,%.0f", name, r, s.P.X, s.P.Y)
					break
				}
			}
		}
	}
}

// TestExit asks of every exit: its rightmost ink, upright, lies 85 to 110
// right of the right edge of its stem,
// its highest ink 90 to 120 above the baseline, its bottom overshoots the
// baseline as o does, and the white inside the hook, between the stem and
// the rising stroke, is at least 25 wide 20 below the highest ink.
func TestExit(t *testing.T) {
	p := Default()
	_, exits := stemLetters(p)
	for name, edge := range exits {
		g := glyph(t, name)
		up := unsheared(p, g.Path)
		right := edge + p.Stem
		var reach, top float64
		for y := 0.5; y < p.ExitTipHeight+20; y++ {
			for _, r := range runs(up, y) {
				if r[1] > right+5 && r[0] > edge {
					reach = math.Max(reach, r[1]-right)
					top = math.Max(top, y)
				}
			}
		}
		t.Logf("%s: exit reach %.0f, top %.0f", name, reach, top)
		if reach < 85 || reach > 110 {
			t.Errorf("%s: the exit reaches %.0f right of the stem, want 85 to 110", name, reach)
		}
		if top < 90 || top > 120 {
			t.Errorf("%s: the exit rises to %.0f, want 90 to 120", name, top)
		}
		_, y0, _, _ := bbox(g.Path)
		if math.Abs(y0+p.Overshoot) > 1 {
			t.Errorf("%s: the bottom of the exit is %.1f, want %v", name, y0, -p.Overshoot)
		}
		rs := runs(up, top-20)
		var white float64 = -1
		for i := 0; i+1 < len(rs); i++ {
			if rs[i][0] < right && rs[i][1] > right-p.Stem/2 {
				white = rs[i+1][0] - rs[i][1]
			}
		}
		if white < 25 {
			t.Errorf("%s: the white inside the exit is %.1f at %.0f, want at least 25 (runs %v)", name, white, top-20, rs)
		}
	}
}
