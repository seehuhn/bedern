// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"slices"
	"testing"

	"seehuhn.de/go/postscript/type1"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestPFB writes the regular font as a PFB file and reads it back: the file
// holds every glyph under its name with its advance and the box of its
// outline, the alignment zones of the OpenType font and the bounding box of
// the glyphs, and two renderings at the same time are identical.
func TestPFB(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	gs := glyphstest.Set(p)
	data, err := PFB(p, design.RegularStyle(), gs, testNow)
	if err != nil {
		t.Fatal(err)
	}
	again, err := PFB(p, design.RegularStyle(), gs, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, again) {
		t.Error("two renderings of the PFB file differ")
	}
	font, err := type1.Read(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if font.FontName != "Bedern-Regular" {
		t.Errorf("font name %q", font.FontName)
	}
	if len(font.Glyphs) != len(gs) {
		t.Errorf("%d glyphs in the PFB file, %d in the font", len(font.Glyphs), len(gs))
	}
	box := [4]float64{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
	for _, g := range gs {
		tg, ok := font.Glyphs[g.Name]
		if !ok {
			t.Errorf("glyph %s is missing", g.Name)
			continue
		}
		if tg.WidthX != g.Advance {
			t.Errorf("glyph %s: advance %g, want %g", g.Name, tg.WidthX, g.Advance)
		}
		if len(g.Path) == 0 {
			if !tg.IsBlank() {
				t.Errorf("glyph %s has an outline", g.Name)
			}
			continue
		}
		x0, y0, x1, y1 := g.Path.BBox()
		b := font.GlyphBBoxPDF(g.Name)
		got := [4]float64{b.LLx, b.LLy, b.URx, b.URy}
		want := [4]float64{x0, y0, x1, y1}
		for i := range got {
			if math.Abs(got[i]-want[i]) > 1e-6 {
				t.Errorf("glyph %s: box %v, want %v", g.Name, got, want)
				break
			}
		}
		box[0], box[1] = math.Min(box[0], x0), math.Min(box[1], y0)
		box[2], box[3] = math.Max(box[2], x1), math.Max(box[3], y1)
	}
	want := PrivateDict(p)
	if !slices.Equal(font.Private.BlueValues, want.BlueValues) || !slices.Equal(font.Private.OtherBlues, want.OtherBlues) {
		t.Errorf("alignment zones %v %v, want %v %v",
			font.Private.BlueValues, font.Private.OtherBlues, want.BlueValues, want.OtherBlues)
	}
	m := regexp.MustCompile(`/FontBBox \[(-?\d+) (-?\d+) (-?\d+) (-?\d+)\] def`).FindSubmatch(data)
	if m == nil {
		t.Fatal("the PFB file has no FontBBox entry")
	}
	wantBox := fmt.Sprintf("%.0f %.0f %.0f %.0f",
		math.Floor(box[0]), math.Floor(box[1]), math.Ceil(box[2]), math.Ceil(box[3]))
	if got := string(bytes.Join(m[1:], []byte(" "))); got != wantBox {
		t.Errorf("FontBBox [%s], want [%s]", got, wantBox)
	}
}
