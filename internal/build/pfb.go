// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/postscript/psenc"
	"seehuhn.de/go/postscript/type1"

	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// PFB returns the font as a PostScript font program in the PFB format, with
// every glyph of gs under its name, the alignment zones of PrivateDict and the
// bounding box of the glyphs.  The built-in encoding is the standard encoding
// of PostScript for the glyphs which the font has.  A zero time leaves the
// creation date out.
func PFB(p design.Params, style design.Style, gs []glyphs.Glyph, now time.Time) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := style.Validate(); err != nil {
		return nil, err
	}
	if len(gs) == 0 || gs[0].Name != ".notdef" {
		return nil, fmt.Errorf("the first glyph must be .notdef")
	}
	weight := "Regular"
	if style.Bold {
		weight = "Bold"
	}
	scale := 1 / p.UnitsPerEm
	font := &type1.Font{
		FontInfo: &type1.FontInfo{
			FontName:           style.PostScript,
			Version:            "0.001",
			Copyright:          bedern.Copyright,
			Notice:             bedern.License,
			FullName:           style.Family + " " + style.Subfamily,
			FamilyName:         style.Family,
			Weight:             weight,
			UnderlinePosition:  funit.Float64(-120),
			UnderlineThickness: funit.Float64(46),
			FontMatrix:         matrix.Matrix{scale, 0, 0, scale, 0, 0},
		},
		Outlines: &type1.Outlines{
			Glyphs:  make(map[string]*type1.Glyph, len(gs)),
			Private: PrivateDict(p),
		},
		CreationDate: now,
	}
	box := [4]float64{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
	for _, g := range gs {
		if _, dup := font.Glyphs[g.Name]; dup {
			return nil, fmt.Errorf("glyph %s occurs twice", g.Name)
		}
		tg := font.NewGlyph(g.Name, g.Advance)
		for _, c := range g.Path {
			tg.MoveTo(c.Start.X, c.Start.Y)
			for _, s := range c.Segs {
				if s.IsCurve {
					tg.CurveTo(s.P1.X, s.P1.Y, s.P2.X, s.P2.Y, s.P3.X, s.P3.Y)
				} else {
					tg.LineTo(s.P3.X, s.P3.Y)
				}
			}
			tg.ClosePath()
		}
		if len(g.Path) > 0 {
			x0, y0, x1, y1 := g.Path.BBox()
			box[0], box[1] = math.Min(box[0], x0), math.Min(box[1], y0)
			box[2], box[3] = math.Max(box[2], x1), math.Max(box[3], y1)
		}
	}
	font.Encoding = make([]string, 256)
	for i, name := range psenc.StandardEncoding {
		font.Encoding[i] = ".notdef"
		if _, ok := font.Glyphs[name]; ok {
			font.Encoding[i] = name
		}
	}
	var buf bytes.Buffer
	if err := font.Write(&buf, &type1.WriterOptions{Format: type1.FormatPFB}); err != nil {
		return nil, err
	}
	return setFontBBox(buf.Bytes(), box)
}

// setFontBBox returns the PFB file with the entry FontBBox of its first
// segment, which holds four zeros, replaced by the box, rounded outwards to
// whole units.
func setFontBBox(data []byte, box [4]float64) ([]byte, error) {
	if len(data) < 6 || data[0] != 128 || data[1] != 1 {
		return nil, fmt.Errorf("the PFB file does not start with a text segment")
	}
	n := int(data[2]) | int(data[3])<<8 | int(data[4])<<16 | int(data[5])<<24
	if n < 0 || 6+n > len(data) {
		return nil, fmt.Errorf("the first segment of the PFB file is cut short")
	}
	head := data[6 : 6+n]
	zero := []byte("/FontBBox [0 0 0 0] def")
	if bytes.Count(head, zero) != 1 {
		return nil, fmt.Errorf("the PFB file has no empty FontBBox entry")
	}
	entry := fmt.Sprintf("/FontBBox [%d %d %d %d] def",
		int(math.Floor(box[0])), int(math.Floor(box[1])), int(math.Ceil(box[2])), int(math.Ceil(box[3])))
	head = bytes.Replace(head, zero, []byte(entry), 1)
	m := len(head)
	out := make([]byte, 0, len(data)+m-n)
	out = append(out, 128, 1, byte(m), byte(m>>8), byte(m>>16), byte(m>>24))
	out = append(out, head...)
	out = append(out, data[6+n:]...)
	return out, nil
}
