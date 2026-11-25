// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package mathfont builds Bedern Math, the mathematics font of the family:
// the glyphs of the regular which mathematics takes as they are (roman.go),
// the signs derived from them (derived.go), the constants of the MATH table
// computed from the parameters of the regular (constants.go), the per-glyph
// data of the table, the script math with the feature ssty, and the
// character maps of format 12 and 4.
package mathfont

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"golang.org/x/text/language"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/postscript/type1"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/cff"
	"seehuhn.de/go/sfnt/cmap"
	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/head"
	"seehuhn.de/go/sfnt/opentype/coverage"
	"seehuhn.de/go/sfnt/opentype/gtab"
	"seehuhn.de/go/sfnt/opentype/mathtable"
	"seehuhn.de/go/sfnt/os2"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// File is the name of the font file, and RangesFile the name of the file of
// the ranges of its code points for unicode-math.
const (
	File       = "BedernMath-Regular.otf"
	RangesFile = "bedern-math-ranges.tex"
)

// The names of the font.
const (
	Family     = "Bedern Math"
	PostScript = "BedernMath-Regular"
)

// Glyphs returns the glyphs of the maths font, in the order of the font:
// .notdef, the space, the roman glyphs, the derived signs and the sizes and
// parts of the delimiters (delims.go).  The accents
// are moved to start their ink at the origin, with an advance equal to
// their ink.
func Glyphs(p design.Params) ([]Glyph, error) {
	r := newRoman(p)
	names := append([]string{".notdef", "space"}, romanNames()...)
	out := make([]Glyph, len(names))
	var wg sync.WaitGroup
	for i, n := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, adv := r.get(n)
			g := Glyph{Name: n, Advance: math.Round(adv), Path: path, Class: Ord, Like: n}
			if accentNames[n] {
				g.Path, g.Advance = withSides(path.Round(), 0, 0)
				g.Class = Accent
			}
			out[i] = g
		}()
	}
	wg.Wait()
	runes := map[string]rune{}
	for _, e := range glyphTable() {
		runes[e.name] = e.r
	}
	for i := range out {
		out[i].Rune = runes[out[i].Name]
	}
	out = append(out, derived(r)...)
	out = append(out, greekCaps(r)...)
	it, err := italicGlyphs()
	if err != nil {
		return nil, err
	}
	out = append(out, it...)
	lower, err := greekLower()
	if err != nil {
		return nil, err
	}
	out = append(out, lower...)
	out, err = addScripts(out, true)
	if err != nil {
		return nil, err
	}
	sizes, variants, parts, err := delimGlyphs(p)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Variants = variants[out[i].Name]
		out[i].Parts = parts[out[i].Name]
	}
	placeDelimiters(out, p.Punct.MathAxis)
	out = append(out, sizes...)
	ops, err := operatorGlyphs(p)
	if err != nil {
		return nil, err
	}
	out = append(out, ops...)
	sg, err := bigSetGlyphs(p)
	if err != nil {
		return nil, err
	}
	out = append(out, sg...)
	out = append(out, dsGlyphs(p)...)
	out = append(out, partialItalic())
	hg, hc, err := horizontalGlyphs(p)
	if err != nil {
		return nil, err
	}
	out = append(out, hg...)
	for i := range out {
		if c, ok := hc[out[i].Name]; ok {
			out[i].Variants, out[i].Parts, out[i].Horizontal = c.Variants, c.Parts, true
		}
	}
	return out, nil
}

type nameRune struct {
	name string
	r    rune
}

// glyphTable returns the names and code points of the glyphs of the regular,
// read from its set once.
var glyphTable = sync.OnceValue(func() []nameRune {
	var out []nameRune
	for _, g := range glyphs.Set(design.Regular()) {
		out = append(out, nameRune{g.Name, g.Rune})
	}
	return out
})

// Font assembles the maths font from the glyphs.  The time is used for the
// creation and modification dates.
func Font(p design.Params, gs []Glyph, now time.Time) (*sfnt.Font, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if len(gs) == 0 || gs[0].Name != ".notdef" {
		return nil, fmt.Errorf("mathfont: the first glyph must be .notdef")
	}
	outlines := &cff.Outlines{}
	cm := cmap.Format12{}
	cm4 := cmap.Format4{}
	id := map[string]glyph.ID{}
	info := &mathtable.Info{
		Constants:           Constants(p),
		ItalicsCorrection:   map[glyph.ID]mathtable.Value{},
		TopAccent:           map[glyph.ID]mathtable.Value{},
		Kerns:               map[glyph.ID]*mathtable.KernInfo{},
		Vertical:            map[glyph.ID]*mathtable.Construction{},
		Horizontal:          map[glyph.ID]*mathtable.Construction{},
		MinConnectorOverlap: funit.Uint16(math.Round(p.Punct.MathThick / 2)),
	}
	for i, g := range gs {
		gid := glyph.ID(i)
		if _, dup := id[g.Name]; dup {
			return nil, fmt.Errorf("mathfont: two glyphs %s", g.Name)
		}
		id[g.Name] = gid
		path := g.Path.Round()
		cg := cff.NewGlyph(g.Name, g.Advance)
		for _, c := range path {
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
			if _, dup := cm[uint32(g.Rune)]; dup {
				return nil, fmt.Errorf("mathfont: two glyphs for %U", g.Rune)
			}
			cm[uint32(g.Rune)] = gid
			if g.Rune <= 0xFFFF {
				cm4[uint16(g.Rune)] = gid
			}
		}
		switch {
		case g.Attached:
			info.TopAccent[gid] = mathtable.Value{V: funit.Int16(math.Round(g.Attach))}
		case len(path) > 0:
			info.TopAccent[gid] = mathtable.Value{V: funit.Int16(math.Round(topAccent(path)))}
		}
		if g.Italic != 0 {
			info.ItalicsCorrection[gid] = mathtable.Value{V: funit.Int16(math.Round(g.Italic))}
		}
		if g.Kern != nil {
			info.Kerns[gid] = g.Kern
		}
	}
	info.ExtendedShapes = coverage.Set{}
	for _, g := range gs {
		if len(g.Variants) == 0 && len(g.Parts) == 0 {
			continue
		}
		c := &mathtable.Construction{}
		for _, n := range append([]string{g.Name}, g.Variants...) {
			v, ok := id[n]
			if !ok {
				return nil, fmt.Errorf("mathfont: no glyph %s", n)
			}
			x0, y0, x1, y1 := gs[v].Path.Round().BBox()
			extent := y1 - y0
			if g.Horizontal {
				extent = x1 - x0
			}
			c.Variants = append(c.Variants, mathtable.Variant{Glyph: v, Advance: funit.Uint16(math.Round(extent))})
			info.ExtendedShapes[v] = true
		}
		if len(g.Parts) > 0 {
			a := &mathtable.Assembly{}
			for _, pt := range g.Parts {
				v, ok := id[pt.Name]
				if !ok {
					return nil, fmt.Errorf("mathfont: no glyph %s", pt.Name)
				}
				a.Parts = append(a.Parts, mathtable.Part{
					Glyph:          v,
					StartConnector: funit.Uint16(math.Round(pt.StartConn)),
					EndConnector:   funit.Uint16(math.Round(pt.EndConn)),
					FullAdvance:    funit.Uint16(math.Round(pt.FullAdvance)),
					Extender:       pt.Extender,
				})
				info.ExtendedShapes[v] = true
			}
			c.Assembly = a
		}
		if g.Horizontal {
			info.Horizontal[id[g.Name]] = c
		} else {
			info.Vertical[id[g.Name]] = c
		}
	}
	outlines.Private = []*type1.PrivateDict{build.PrivateDict(p)}
	outlines.FDSelect = func(glyph.ID) int { return 0 }

	scale := 1 / p.UnitsPerEm
	font := &sfnt.Font{
		FontName:   PostScript,
		FamilyName: Family,
		Subfamily:  "Regular",
		FullName:   Family,
		Weight:     os2.WeightNormal,
		Width:      os2.WidthNormal,
		IsRegular:  true,
		IsSerif:    true,

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

		Ascent:      funit.Int16(p.TypoAscender),
		Descent:     funit.Int16(p.TypoDescender),
		LineGap:     funit.Int16(p.LineGap),
		HheaAscent:  funit.Int16(p.TypoAscender),
		HheaDescent: funit.Int16(p.TypoDescender),
		HheaLineGap: funit.Int16(p.LineGap),

		CapHeight: funit.Int16(p.CapHeight),
		XHeight:   funit.Int16(p.XHeight),

		Outlines: outlines,
		Math:     info,
	}
	// The character map is written by hand: Font.InstallCMap writes a
	// format 12 subtable without the format 4 subtable of the Basic
	// Multilingual Plane, and under (0, 3) and (3, 1) when its code points
	// lie in that plane.  The format 12 subtable stands under (0, 4) and
	// (3, 10), the format 4 subtable under (0, 3) and (3, 1).
	sub12, err := cm.Encode(0)
	if err != nil {
		return nil, err
	}
	sub4, err := cm4.Encode(0)
	if err != nil {
		return nil, err
	}
	font.CMapTable = cmap.Table{
		{PlatformID: 0, EncodingID: 4}:  sub12,
		{PlatformID: 3, EncodingID: 10}: sub12,
		{PlatformID: 0, EncodingID: 3}:  sub4,
		{PlatformID: 3, EncodingID: 1}:  sub4,
	}
	font.CodePageRange.Set(os2.CP1252)
	font.UnicodeRange.Set(os2.URBasicLatin)
	font.UnicodeRange.Set(os2.URLatin1Sup)
	font.UnicodeRange.Set(os2.URCombiningDiacriticalMarks)
	font.UnicodeRange.Set(os2.URGeneralPunctuation)
	// The library names the bits up to 33; 38 is Mathematical Operators and
	// 39 Miscellaneous Technical.
	font.UnicodeRange.Set(os2.UnicodeRangeBit(38))
	font.UnicodeRange.Set(os2.UnicodeRangeBit(39))
	font.Gsub = gsubInfo(id, gs)
	return font, nil
}

// topAccent returns the attachment of a top accent: the middle of the ink.
func topAccent(path outline.Path) float64 {
	x0, _, x1, _ := path.BBox()
	return (x0 + x1) / 2
}

// gsubInfo returns the GSUB table: the script math, with its default
// language system only, and the script DFLT, each with the feature ssty, an
// alternate substitution whose first alternate is the form of the first
// script level and the second that of the second.  The library writes the
// script math from the tag und-Zmth.
func gsubInfo(id map[string]glyph.ID, gs []Glyph) *gtab.Info {
	sub := &gtab.Gsub3_1{Cov: coverage.Table{}}
	alts := map[glyph.ID][]glyph.ID{}
	for _, g := range gs {
		if g.Script[0] != "" {
			alts[id[g.Name]] = []glyph.ID{id[g.Script[0]], id[g.Script[1]]}
		}
	}
	var cov []glyph.ID
	for g := range alts {
		cov = append(cov, g)
	}
	slices.Sort(cov)
	for i, g := range cov {
		sub.Cov[g] = i
		sub.Alternates = append(sub.Alternates, alts[g])
	}
	feat := &gtab.Features{Required: 0xFFFF, Optional: []gtab.FeatureIndex{0}}
	return &gtab.Info{
		ScriptList: gtab.ScriptListInfo{
			language.MustParse("und-Zmth"): feat,
			language.MustParse("und-Zzzz"): feat,
		},
		FeatureList: []*gtab.Feature{{Tag: "ssty", Lookups: []gtab.LookupIndex{0}}},
		LookupList: []*gtab.LookupTable{{
			Meta:      &gtab.LookupMetaInfo{LookupType: 3},
			Subtables: []gtab.Subtable{sub},
		}},
	}
}
