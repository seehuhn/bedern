// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package diag

import (
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
	"seehuhn.de/go/bedern/internal/outline"
)

// element is one start element of the drawing with its attributes.
type element struct {
	name  string
	attrs map[string]string
	text  string
}

// parse reads the whole drawing and fails the test if it is not well formed.
func parse(t *testing.T, data []byte) []element {
	t.Helper()
	dec := xml.NewDecoder(bytes.NewReader(data))
	var els []element
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("not well-formed XML: %v", err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			e := element{name: tok.Name.Local, attrs: map[string]string{}}
			for _, a := range tok.Attr {
				e.attrs[a.Name.Local] = a.Value
			}
			els = append(els, e)
		case xml.CharData:
			if len(els) > 0 {
				els[len(els)-1].text += string(tok)
			}
		}
	}
	return els
}

func byClass(els []element, class string) []element {
	var out []element
	for _, e := range els {
		for _, c := range strings.Fields(e.attrs["class"]) {
			if c == class {
				out = append(out, e)
			}
		}
	}
	return out
}

func glyph(t *testing.T, name string, p design.Params) glyphs.Glyph {
	t.Helper()
	if g := glyphstest.Glyph(name, p); g.Name == name {
		return g
	}
	t.Fatalf("no glyph %q", name)
	return glyphs.Glyph{}
}

// wantCounts returns the number of distinct on-curve points and of control
// points of the outline of g.
func wantCounts(g glyphs.Glyph) (on, ctrl int) {
	for _, c := range g.Path {
		seen := map[[2]float64]bool{{c.Start.X, c.Start.Y}: true}
		for _, s := range c.Segs {
			if s.IsCurve {
				ctrl += 2
			}
			seen[[2]float64{s.P3.X, s.P3.Y}] = true
		}
		on += len(seen)
	}
	return
}

func num(t *testing.T, e element, attr string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(e.attrs[attr], 64)
	if err != nil {
		t.Fatalf("%s %s=%q: %v", e.attrs["id"], attr, e.attrs[attr], err)
	}
	return v
}

func TestSVG(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, name := range []string{"o", "l", "space", ".notdef"} {
		g := glyph(t, name, p)
		var buf bytes.Buffer
		if err := SVG(&buf, g, p, "regular", Options{}); err != nil {
			t.Fatal(err)
		}
		els := parse(t, buf.Bytes())

		wantPaths := 1
		if len(g.Path) == 0 {
			wantPaths = 0
		}
		if got := len(byClass(els, "outline")); got != wantPaths {
			t.Errorf("%s: %d outline paths, want %d", name, got, wantPaths)
		}
		for _, e := range byClass(els, "outline") {
			if e.name != "path" {
				t.Errorf("%s: outline is a %s element", name, e.name)
			}
		}
		on, ctrl := wantCounts(g)
		if got := len(byClass(els, "oncurve")); got != on {
			t.Errorf("%s: %d on-curve markers, want %d", name, got, on)
		}
		if got := len(byClass(els, "control")); got != ctrl {
			t.Errorf("%s: %d control markers, want %d", name, got, ctrl)
		}
		if got := len(byClass(els, "handle")); got != ctrl {
			t.Errorf("%s: %d handle lines, want %d", name, got, ctrl)
		}
		if name == "o" && (on != 8 || ctrl != 16) {
			t.Errorf("o: counted %d and %d points, want 8 and 16", on, ctrl)
		}

		guides := map[string]element{}
		for _, e := range byClass(els, "guide") {
			guides[e.attrs["id"]] = e
		}
		for id, y := range map[string]float64{
			"baseline": 0, "xheight": -p.XHeight, "capheight": -p.CapHeight,
			"ascender": -p.Ascender, "descender": p.Descender,
		} {
			e, ok := guides[id]
			if !ok {
				if id == "baseline" || id == "xheight" {
					t.Errorf("%s: no guide %q", name, id)
				}
				continue
			}
			if num(t, e, "y1") != y || num(t, e, "y2") != y {
				t.Errorf("%s: guide %s at y=%s, want %v (y axis flipped)", name, id, e.attrs["y1"], y)
			}
		}
		for id, x := range map[string]float64{"origin": 0, "advance": g.Advance} {
			e, ok := guides[id]
			if !ok {
				t.Errorf("%s: no guide %q", name, id)
				continue
			}
			if num(t, e, "x1") != x || num(t, e, "x2") != x {
				t.Errorf("%s: guide %s at x=%s, want %v", name, id, e.attrs["x1"], x)
			}
		}

		var text string
		for _, e := range byClass(els, "caption") {
			text += e.text + " "
		}
		for _, want := range []string{name, strconv.Itoa(int(g.Advance)), "regular"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s: caption %q lacks %q", name, text, want)
			}
		}
	}
}

// render draws g and returns the elements of the drawing.
func render(t *testing.T, g glyphs.Glyph, p design.Params, opt Options) ([]element, string) {
	t.Helper()
	var buf bytes.Buffer
	if err := SVG(&buf, g, p, "regular", opt); err != nil {
		t.Fatal(err)
	}
	return parse(t, buf.Bytes()), buf.String()
}

// numbers returns all numbers in path data.
func numbers(t *testing.T, d string) []float64 {
	t.Helper()
	var out []float64
	for _, f := range strings.FieldsFunc(d, func(r rune) bool { return r == ' ' || strings.ContainsRune("MLCZ", r) }) {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			t.Fatalf("path data %q: %v", d, err)
		}
		out = append(out, v)
	}
	return out
}

func TestFlip(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	g := glyph(t, "o", p)
	els, _ := render(t, g, p, Options{})

	top, bottom := math.Inf(1), math.Inf(-1)
	for _, e := range byClass(els, "oncurve") {
		y := num(t, e, "y") + num(t, e, "height")/2 // centre of the marker
		top, bottom = math.Min(top, y), math.Max(bottom, y)
	}
	if want := -(p.XHeight + p.OvershootLower); top != want {
		t.Errorf("topmost on-curve marker at y=%v, want %v", top, want)
	}
	if want := p.OvershootLower; bottom != want {
		t.Errorf("lowest on-curve marker at y=%v, want %v", bottom, want)
	}

	// The path data holds the negated y coordinates, in the order of the
	// outline.
	out := byClass(els, "outline")
	if len(out) != 1 {
		t.Fatalf("%d outlines", len(out))
	}
	nums := numbers(t, out[0].attrs["d"])
	var want []float64
	for _, c := range g.Path {
		want = append(want, -c.Start.Y)
		for _, s := range c.Segs {
			if s.IsCurve {
				want = append(want, -s.P1.Y, -s.P2.Y)
			}
			want = append(want, -s.P3.Y)
		}
	}
	var got []float64
	for i := 1; i < len(nums); i += 2 {
		got = append(got, nums[i])
	}
	if !slices.Equal(got, want) {
		t.Errorf("y coordinates in path data:\n got %v\nwant %v", got, want)
	}
}

func TestOvershootGuides(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	els, _ := render(t, glyph(t, "o", p), p, Options{})
	guides := map[string]element{}
	for _, e := range byClass(els, "guide") {
		guides[e.attrs["id"]] = e
	}
	for id, y := range map[string]float64{
		"overshoot-top":    -(p.XHeight + p.OvershootLower),
		"overshoot-bottom": p.OvershootLower,
	} {
		e, ok := guides[id]
		if !ok {
			t.Errorf("no guide %q", id)
			continue
		}
		if num(t, e, "y1") != y || num(t, e, "y2") != y {
			t.Errorf("guide %s at y=%s, want %v", id, e.attrs["y1"], y)
		}
		if num(t, e, "stroke-width") >= num(t, guides["baseline"], "stroke-width") {
			t.Errorf("guide %s is not lighter than the baseline", id)
		}
		if e.attrs["stroke"] == guides["baseline"].attrs["stroke"] {
			t.Errorf("guide %s has the colour of the baseline", id)
		}
	}
}

func TestMetrics(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, name := range []string{"o", "l", ".notdef"} {
		g := glyph(t, name, p)
		x0, _, x1, _ := g.Path.BBox()
		els, _ := render(t, g, p, Options{})
		got := map[string]string{}
		for _, e := range byClass(els, "metric") {
			f := strings.Fields(e.text)
			if len(f) != 2 {
				t.Errorf("%s: metric %q", name, e.text)
				continue
			}
			got[e.attrs["id"]] = f[1]
		}
		want := map[string]float64{"lsb": x0, "rsb": g.Advance - x1, "xmin": x0, "xmax": x1}
		for id, v := range want {
			if got[id] != strconv.FormatFloat(v, 'g', -1, 64) {
				t.Errorf("%s: %s = %q, want %v", name, id, got[id], v)
			}
		}
		if name == "o" && (got["lsb"] != "26" || got["rsb"] != "27") {
			t.Errorf("o: sidebearings %s and %s, want 26 and 27", got["lsb"], got["rsb"])
		}
	}
	// No outline, no numbers.
	space := glyph(t, "space", p)
	els, _ := render(t, space, p, Options{})
	if n := len(byClass(els, "metric")); n != 0 {
		t.Errorf("space: %d metrics", n)
	}
}

func TestPointLabels(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	g := glyph(t, "o", p)
	els, _ := render(t, g, p, Options{})
	marks := byClass(els, "oncurve")
	labels := byClass(els, "ptlabel")
	if len(labels) != len(marks) {
		t.Fatalf("%d labels for %d markers", len(labels), len(marks))
	}
	// Match the labels to the markers by their text.
	type pair struct{ mx, my, lx, ly float64 }
	byText := map[string]pair{}
	for _, c := range g.Path {
		pts := []outline.Point{c.Start}
		for _, s := range c.Segs {
			pts = append(pts, s.P3)
		}
		for _, q := range pts {
			text := strconv.FormatFloat(q.X, 'g', -1, 64) + "," + strconv.FormatFloat(q.Y, 'g', -1, 64)
			byText[text] = pair{q.X, -q.Y, 0, 0}
		}
	}
	for _, l := range labels {
		pr, ok := byText[strings.TrimSpace(l.text)]
		if !ok {
			t.Errorf("label %q matches no point", l.text)
			continue
		}
		pr.lx, pr.ly = num(t, l, "x"), num(t, l, "y")
		byText[strings.TrimSpace(l.text)] = pr
	}
	// The centre of o is at its advance/2 horizontally and half the
	// x-height vertically.  Labels of the outer contour point away from the
	// centre, labels of the counter point towards it.
	var outerTopP, outerBotP, outerLeftP, counterTopP, counterLeftP pair
	topY, botY, minX := math.Inf(1), math.Inf(-1), math.Inf(1)
	for _, pr := range byText {
		topY, botY, minX = math.Min(topY, pr.my), math.Max(botY, pr.my), math.Min(minX, pr.mx)
	}
	secondTop, secondLeft := math.Inf(1), math.Inf(1)
	for _, pr := range byText {
		if pr.my == topY {
			outerTopP = pr
		}
		if pr.my == botY {
			outerBotP = pr
		}
		if pr.mx == minX {
			outerLeftP = pr
		}
		if pr.my > topY && pr.my < secondTop && pr.mx > minX+50 {
			secondTop = pr.my
		}
		if pr.mx > minX && pr.mx < secondLeft {
			secondLeft = pr.mx
		}
	}
	for _, pr := range byText {
		if pr.my == secondTop {
			counterTopP = pr
		}
		if pr.mx == secondLeft {
			counterLeftP = pr
		}
	}
	if !(outerTopP.ly < outerTopP.my) {
		t.Errorf("label of the top point at y=%v, point at %v: want above", outerTopP.ly, outerTopP.my)
	}
	if !(outerBotP.ly > outerBotP.my) {
		t.Errorf("label of the bottom point at y=%v, point at %v: want below", outerBotP.ly, outerBotP.my)
	}
	if !(outerLeftP.lx < outerLeftP.mx) {
		t.Errorf("label of the leftmost point at x=%v, point at %v: want to the left", outerLeftP.lx, outerLeftP.mx)
	}
	if !(counterTopP.ly > counterTopP.my) {
		t.Errorf("label of the counter's top point at y=%v, point at %v: want below", counterTopP.ly, counterTopP.my)
	}
	if !(counterLeftP.lx > counterLeftP.mx) {
		t.Errorf("label of the counter's left point at x=%v, point at %v: want to the right", counterLeftP.lx, counterLeftP.mx)
	}
}

func TestOverlay(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	g := glyph(t, "o", p)
	other := glyph(t, "l", p)

	els, svg := render(t, g, p, Options{})
	if n := len(byClass(els, "overlay")); n != 0 {
		t.Errorf("%d overlays without overlay glyph", n)
	}
	if strings.Contains(svg, "overlay") {
		t.Error("the caption mentions an overlay although there is none")
	}

	els, _ = render(t, g, p, Options{Overlay: &other, OverlayLabel: "regular"})
	ov := byClass(els, "overlay")
	if len(ov) != 1 {
		t.Fatalf("%d overlays, want 1", len(ov))
	}
	if ov[0].name != "path" || ov[0].attrs["fill"] != "none" || ov[0].attrs["d"] != pathData(other.Path) {
		t.Errorf("overlay is %s with fill %q and d=%q", ov[0].name, ov[0].attrs["fill"], ov[0].attrs["d"])
	}
	if c := ov[0].attrs["stroke"]; !strings.HasPrefix(c, "#") || c[1] < 'c' {
		t.Errorf("overlay stroke %q is not red", c)
	}
	if n := len(byClass(els, "outline")); n != 1 {
		t.Errorf("%d main outlines", n)
	}
	var caption string
	for _, e := range byClass(els, "caption") {
		caption += e.text
	}
	for _, want := range []string{"overlay", "regular", "l"} {
		if !strings.Contains(caption, want) {
			t.Errorf("caption %q lacks %q", caption, want)
		}
	}
}

func TestCrop(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	cases := []struct {
		name      string
		want, not []string
	}{
		{"o", []string{"baseline", "xheight", "overshoot-top", "overshoot-bottom"}, []string{"descender", "capheight", "ascender"}},
		{"l", []string{"baseline", "xheight", "capheight", "ascender"}, []string{"descender"}},
		{".notdef", []string{"baseline", "xheight", "capheight"}, []string{"descender"}},
		{"space", []string{"baseline", "xheight"}, []string{"descender", "capheight", "ascender"}},
	}
	var heights []float64
	for _, c := range cases {
		els, _ := render(t, glyph(t, c.name, p), p, Options{})
		ids := map[string]bool{}
		for _, e := range byClass(els, "guide") {
			ids[e.attrs["id"]] = true
		}
		for _, id := range c.want {
			if !ids[id] {
				t.Errorf("%s: guide %s is missing", c.name, id)
			}
		}
		for _, id := range c.not {
			if ids[id] {
				t.Errorf("%s: guide %s is present, outside the cropped area", c.name, id)
			}
		}
		vb := strings.Fields(els[0].attrs["viewBox"])
		h, _ := strconv.ParseFloat(vb[3], 64)
		heights = append(heights, h)
		if full := p.Ascender + p.Descender; c.name == "o" && h >= full {
			t.Errorf("o: drawing height %v is not smaller than the full range %v", h, full)
		}
	}
	if heights[0] >= heights[1] {
		t.Errorf("o (%v) is not lower than l (%v)", heights[0], heights[1])
	}
}

func TestGuideLabels(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, name := range []string{"o", "l", "space"} {
		g := glyph(t, name, p)
		els, _ := render(t, g, p, Options{})
		labels := byClass(els, "guidelabel")
		if len(labels) < 2 {
			t.Errorf("%s: %d guide labels", name, len(labels))
		}
		for _, e := range labels {
			if x := num(t, e, "x"); x <= g.Advance || e.attrs["text-anchor"] == "end" {
				t.Errorf("%s: guide label %q at x=%v is not right of the advance line", name, e.text, x)
			}
		}
	}
}

func TestEscape(t *testing.T) {
	t.Parallel()
	if got, want := escape(`a"b<c>&d`), "a&quot;b&lt;c&gt;&amp;d"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
