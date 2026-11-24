// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"fmt"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/cff"
	"seehuhn.de/go/sfnt/glyph"

	"seehuhn.de/go/bedern/internal/outline"
)

// Glyph is a glyph read from a font file.
type Glyph struct {
	Name    string
	Rune    rune // 0 if the glyph has no code point among the runes asked for
	Advance float64
	Path    outline.Path
}

// ReadFont reads the glyphs of a font file with CFF outlines, scaled to an
// em of 1000 units.  If runes is not empty, only the glyphs of these
// characters are returned, in this order, and characters which the font
// lacks are left out; otherwise every glyph is returned.  Nothing of the font
// is kept: the outlines are for measuring.
func ReadFont(file string, runes []rune) ([]Glyph, error) {
	f, err := sfnt.ReadFile(file)
	if err != nil {
		return nil, err
	}
	outlines, ok := f.Outlines.(*cff.Outlines)
	if !ok {
		return nil, fmt.Errorf("%s: not a font with CFF outlines", file)
	}
	scale := 1000 / float64(f.UnitsPerEm)
	conv := func(gid glyph.ID, r rune) Glyph {
		cg := outlines.Glyphs[gid]
		var p outline.Path
		for _, op := range cg.Cmds {
			a := op.Args
			switch op.Op {
			case cff.OpMoveTo:
				p.MoveTo(a[0]*scale, a[1]*scale)
			case cff.OpLineTo:
				if len(p) > 0 {
					p.LineTo(a[0]*scale, a[1]*scale)
				}
			case cff.OpCurveTo:
				if len(p) > 0 {
					p.CurveTo(a[0]*scale, a[1]*scale, a[2]*scale, a[3]*scale, a[4]*scale, a[5]*scale)
				}
			}
		}
		return Glyph{Name: f.GlyphName(gid), Rune: r, Advance: f.GlyphWidth(gid) * scale, Path: p}
	}
	var out []Glyph
	if len(runes) == 0 {
		for i := range outlines.Glyphs {
			out = append(out, conv(glyph.ID(i), 0))
		}
		return out, nil
	}
	cm, err := f.CMapTable.GetBest()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	for _, r := range runes {
		gid := cm.Lookup(r)
		if gid == 0 || int(gid) >= len(outlines.Glyphs) {
			continue
		}
		out = append(out, conv(gid, r))
	}
	return out, nil
}
