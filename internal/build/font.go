// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package build assembles OpenType font files from glyph outlines.
package build

import (
	"fmt"
	"time"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/postscript/type1"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/cff"
	"seehuhn.de/go/sfnt/cmap"
	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/head"
	"seehuhn.de/go/sfnt/os2"

	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/kern"
)

// privateDict returns the CFF Private dict, which holds the alignment zones
// used for hinting.  BlueValues has the baseline, x-height and cap-height
// zones.  OtherBlues has the bottom zone of the descender line: the line
// itself and, below it, the overshoot of a round descender.  BlueShift and
// BlueFuzz are set explicitly.
func privateDict(p design.Params) *type1.PrivateDict {
	return &type1.PrivateDict{
		BlueValues: []float64{
			-p.OvershootLower, 0,
			p.XHeight, p.XHeight + p.OvershootLower,
			p.CapHeight, p.CapHeight + p.OvershootUpper,
		},
		OtherBlues: []float64{-p.Descender - p.OvershootLower, -p.Descender},
		BlueShift:  7,
		BlueFuzz:   1,
	}
}

// Options are the choices of Font which are not parameters of the design.
type Options struct {
	// NoKern leaves out the feature kern.
	NoKern bool
}

// Font assembles a Bedern font of the given style from the given glyphs, with
// the kerning computed from the outlines.  The first glyph must be .notdef.
// The time is used for the creation and modification dates.
func Font(p design.Params, style design.Style, gs []glyphs.Glyph, now time.Time) (*sfnt.Font, error) {
	return FontWith(p, style, gs, now, Options{})
}

// FontWith is Font with options.
func FontWith(p design.Params, style design.Style, gs []glyphs.Glyph, now time.Time, opts Options) (*sfnt.Font, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := style.Validate(); err != nil {
		return nil, err
	}
	if len(gs) == 0 || gs[0].Name != ".notdef" {
		return nil, fmt.Errorf("the first glyph must be .notdef")
	}

	outlines := &cff.Outlines{}
	cm := cmap.Format4{}
	for i, g := range gs {
		cg := cff.NewGlyph(g.Name, g.Advance)
		for _, c := range g.Path {
			cg.MoveTo(c.Start.X, c.Start.Y)
			for _, s := range c.Segs {
				if s.IsCurve {
					cg.CurveTo(s.P1.X, s.P1.Y, s.P2.X, s.P2.Y, s.P3.X, s.P3.Y)
				} else {
					cg.LineTo(s.P3.X, s.P3.Y)
				}
			}
		}
		outlines.Glyphs = append(outlines.Glyphs, cg)
		if g.Rune != 0 {
			if g.Rune > 0xFFFF {
				return nil, fmt.Errorf("glyph %s: code point %U is outside the BMP", g.Name, g.Rune)
			}
			cm[uint16(g.Rune)] = glyph.ID(i)
		}
	}
	outlines.Private = []*type1.PrivateDict{privateDict(p)}
	outlines.FDSelect = func(glyph.ID) int { return 0 }

	// The glyph outlines are drawn in font units, 1000 per em.
	scale := 1 / p.UnitsPerEm

	font := &sfnt.Font{
		// The PostScript name is stored in the CFF Name INDEX and as name ID 6.
		FontName:   style.PostScript,
		FamilyName: style.Family,
		Subfamily:  style.Subfamily,
		FullName:   style.Family + " " + style.Subfamily,
		Weight:     os2.Weight(style.Weight),
		Width:      os2.WidthNormal,
		IsRegular:  !style.Bold && !style.Italic,
		IsBold:     style.Bold,
		IsItalic:   style.Italic,
		IsSerif:    true,

		// The head table stores the revision as a 16.16 fixed-point number.
		// The value 0x42 is displayed as 0.001.
		Version:          head.Version(0x42),
		CreationTime:     now,
		ModificationTime: now,

		Copyright:  bedern.Copyright,
		License:    bedern.License,
		LicenseURL: bedern.LicenseURL,
		PermUse:    os2.PermInstall,

		UnitsPerEm: uint16(p.UnitsPerEm),
		FontMatrix: matrix.Matrix{scale, 0, 0, scale, 0, 0},

		UnderlinePosition:  -120,
		UnderlineThickness: 46,

		Ascent:  funit.Int16(p.TypoAscender),
		Descent: funit.Int16(p.TypoDescender),
		LineGap: funit.Int16(p.LineGap),

		// The hhea metrics equal the typographic metrics.
		HheaAscent:  funit.Int16(p.TypoAscender),
		HheaDescent: funit.Int16(p.TypoDescender),
		HheaLineGap: funit.Int16(p.LineGap),

		CapHeight: funit.Int16(p.CapHeight),
		XHeight:   funit.Int16(p.XHeight),

		Outlines: outlines,
	}
	if err := font.InstallCMap(cm); err != nil {
		return nil, err
	}
	// The code page range is Windows Latin 1 and the Unicode range is Basic
	// Latin, with the ranges and code pages of the accented letters where the
	// font has them.
	font.CodePageRange.Set(os2.CP1252)
	font.UnicodeRange.Set(os2.URBasicLatin)
	has := func(r rune) bool { _, ok := cm[uint16(r)]; return ok }
	if has('é') {
		font.UnicodeRange.Set(os2.URLatin1Sup)
	}
	if has('ő') && has('ş') && has('ž') {
		font.UnicodeRange.Set(os2.URLatinExtA)
		font.CodePageRange.Set(os2.CP1250)
		font.CodePageRange.Set(os2.CP1254)
		font.CodePageRange.Set(os2.CP1257)
	}
	if has('ẁ') {
		font.UnicodeRange.Set(os2.URLatinExtAdditional)
	}
	if has('\u0301') {
		font.UnicodeRange.Set(os2.URCombiningDiacriticalMarks)
	}
	if has('ˆ') {
		font.UnicodeRange.Set(os2.URSpacingModifierLetters)
	}
	font.Gdef = gdefInfo(gs)
	if !opts.NoKern {
		font.Gpos = gposInfo(kern.Compute(p, gs), gs)
	} else {
		font.Gpos = gposInfo(&kern.Table{}, gs)
	}
	return font, nil
}
