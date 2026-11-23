// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"seehuhn.de/go/bedern/internal/font/design"
)

// The directories of the files in a tree of the TeX Directory Structure.
const (
	DirTFM      = "fonts/tfm/public/" + Family
	DirPFB      = "fonts/type1/public/" + Family
	DirOpenType = "fonts/opentype/public/" + Family
	DirEnc      = "fonts/enc/dvips/" + Family
	DirMap      = "fonts/map/dvips/" + Family
	DirLaTeX    = "tex/latex/" + Family
	DirGeneric  = "tex/generic/" + Family
	DirPlain    = "tex/plain/" + Family
	DirDoc      = "doc/fonts/" + Family
)

// A Report says what the metric file of one font in one encoding holds.
type Report struct {
	// TFM is the name of the metric file, without the ending.
	TFM string

	// Missing lists the positions of the encoding for which the font has no
	// glyph.
	Missing []int

	// Chars, Ligs and Kerns are the numbers of characters, of ligatures and
	// of kerns.
	Chars, Ligs, Kerns int

	// Reductions says how the tables of the heights, the depths and the
	// italic corrections are reduced to the sizes of the format.
	Reductions []Reduction
}

// Files returns the metric files, the encoding files, the map file, the font
// definition files, the package for LaTeX, the macro file for plain TeX and
// the file of the mappings to Unicode for the fonts, as a map from the path of
// each file in a tree of the TeX Directory Structure to its content, with a
// report for each metric file.  The parameters and the kerns of the fonts are
// given in the order of the fonts.  Each upright font gets a slanted instance
// (Slanted) with a metric file of its own, which uses the outline file of the
// upright font.
func Files(fonts []*Font, params []design.Params, kerns []KernFunc, now time.Time) (map[string][]byte, []Report, error) {
	var slanted []*Font
	for i, f := range fonts {
		if f.Slant == 0 {
			slanted = append(slanted, Slanted([]*Font{f})...)
			params = append(slices.Clone(params), params[i])
			kerns = append(slices.Clone(kerns), kerns[i])
		}
	}
	fonts = append(slices.Clone(fonts), slanted...)
	files := make(map[string][]byte)
	var reports []Report
	for _, e := range Encodings {
		enc := EncFile(e, fonts[0].Glyphs)
		for _, f := range fonts[1:] {
			if !bytes.Equal(enc, EncFile(e, f.Glyphs)) {
				return nil, nil, fmt.Errorf("the fonts differ in their glyphs of the encoding %s", e.Name)
			}
		}
		files[DirEnc+"/"+e.EncFile()] = enc
		files[DirLaTeX+"/"+FDFile(e)] = FD(e, fonts, now)
		for i, f := range fonts {
			m := NewMetrics(e, params[i], f.Glyphs, kerns[i])
			m.Bold, m.Italic = f.Style.Bold, f.Style.Italic
			if f.Slant != 0 {
				m.Slant, m.Italic = f.Slant, true
				m.SlantItalics(f.Glyphs)
			}
			r := Report{TFM: f.TFM(e), Reductions: m.Reduce(params[i])}
			_, r.Missing = e.Resolve(f.Glyphs)
			for _, ch := range m.Chars {
				if ch != nil {
					r.Chars++
					r.Ligs += len(ch.Ligs)
					r.Kerns += len(ch.Kerns)
				}
			}
			data, err := m.TFM()
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %w", r.TFM, err)
			}
			files[DirTFM+"/"+r.TFM+".tfm"] = data
			reports = append(reports, r)
		}
	}
	files[DirMap+"/"+MapFile] = Map(fonts)
	files[DirLaTeX+"/"+Family+".sty"] = Package(fonts, now)
	files[DirGeneric+"/"+UnicodeFile] = GlyphToUnicode(fonts)
	files[DirPlain+"/"+PlainFile] = Plain(fonts)
	return files, reports, nil
}

// Reduce reduces the heights and the depths of the characters to 15 values
// other than zero each and the italic corrections to 63, the sizes of the
// tables of a metric file, and returns what it did.  The method is that of
// PLtoTF (see reduce): the largest change is the least which the sizes allow,
// zero stays zero, and no other value becomes zero.  The lines of the design
// (the x-height, the height of the figures, of the capitals and of the
// ascenders and the depth of the descenders) stay exact where that
// bound allows.
func (m *Metrics) Reduce(p design.Params) []Reduction {
	return []Reduction{
		m.reduceTable("height", 15,
			[]float64{p.XHeight, p.Figures.Height, p.CapHeight, p.Ascender},
			func(ch *Char) *float64 { return &ch.Height }),
		m.reduceTable("depth", 15, []float64{p.Descender},
			func(ch *Char) *float64 { return &ch.Depth }),
		m.reduceTable("italic", 63, nil,
			func(ch *Char) *float64 { return &ch.Italic }),
	}
}
