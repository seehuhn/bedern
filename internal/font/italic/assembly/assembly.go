// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package assembly puts the italic together into one font: the lowercase of
// italic, the capitals, figures, signs and ampersand of italic/upper, the
// dotless i and j, the marks and the composed letters, with the side spaces
// of the fit table and the kerning of the roman's model.
//
// The font is marked as an italic through the italic bits of OS/2
// fsSelection and head macStyle, the italic angle of post and CFF, and the
// caret slope of hhea.
package assembly

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"sync"
	"time"
	"unicode"

	"seehuhn.de/go/sfnt"

	"seehuhn.de/go/bedern/internal/build"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/kern"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/upper"
)

// File is the name of the trial font file.
const File = "Bedern-Italic-Trial.otf"

// Style is the style of the trial font.
func Style() design.Style {
	return design.Style{
		Family:     "Bedern Trial",
		Subfamily:  "Italic",
		PostScript: "BedernTrial-Italic",
		Weight:     400,
		Italic:     true,
	}
}

// Params are the parameters of the assembly.
type Params struct {
	Lower italic.Params // the lowercase
	Upper upper.Params  // the figures and the signs
	Caps  upper.Params  // the capitals and the ampersand
	Fit   bool          // apply the fit table
	Space float64       // the advance of the word space and the no-break space
}

// CapWeight is the weight of the capitals over that of the roman capitals,
// at right angles to the strokes.  The stem of the capitals over the stem of
// the lowercase in the italic is 1.141 times that of the roman: 96/81 times
// 0.963, which is 84 horizontally over the 74 of the lowercase.
const CapWeight = 84 * 0.970143 / 96

// WordSpace is the advance of the word space: 340, which makes the white
// between two italic words 0.97 of the roman's.
const WordSpace = 340

// KernFloor is the floor of the kerning of the italic: the least distance
// between the outlines of two glyphs set with their kern.
const KernFloor = 30

// Kerning returns the parameters with which the italic is kerned: those of
// the roman with the floor of the italic, the approach limit of 0.6 of the
// closest approach of HH and the allowance of 10 for two capitals.
// KernTable adds the rules of the lowercase.
func Kerning() design.Params {
	rp := design.Regular()
	rp.Kern.Floor = KernFloor
	rp.Kern.Approach = 0.6
	rp.Kern.AllowanceUpper = 10
	return rp
}

// Default returns the parameters of the trial.
func Default() Params {
	c := upper.Default()
	c.Weight = CapWeight
	c.CapStem = 84
	return Params{Lower: italic.Default(), Upper: upper.Default(), Caps: c, Fit: true, Space: WordSpace}
}

// slant returns the tangent of the slant.
func (p Params) slant() float64 { return float64(p.Lower.SlantNum) / float64(p.Lower.SlantDen) }

// ItalicAngle returns the italic angle in degrees, counterclockwise from the
// vertical, as the post table has it: negative for a slant to the right.
func (p Params) ItalicAngle() float64 {
	return -math.Atan(p.slant()) * 180 / math.Pi
}

// isCap says whether the glyph of the upper package is drawn with the weight
// of the capitals.
func isCap(name string) bool {
	return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' || name == "ampersand"
}

// bases returns the glyphs which are not composed: the .notdef and the space
// of the roman, the lowercase, the dotless i and j and the glyphs of the
// upper package, unfitted and without anchors.
func bases(p Params) ([]glyphs.Glyph, error) {
	rp := design.Regular()
	var out []glyphs.Glyph
	for _, name := range []string{".notdef", "space"} {
		path, adv, _ := glyphs.Outline(name, rp)
		r := rune(0)
		if name == "space" {
			r = ' '
		}
		if name == "space" && p.Space > 0 {
			adv = p.Space
		}
		out = append(out, glyphs.Glyph{Name: name, Rune: r, Advance: adv, Path: path.Round()})
	}
	lower, err := italic.Glyphs(p.Lower)
	if err != nil {
		return nil, err
	}
	for _, g := range lower {
		out = append(out, glyphs.Glyph{Name: italic.GlyphName(g.Name), Rune: g.Rune, Advance: g.Advance, Path: g.Path})
	}
	for _, d := range []struct{ name, from string }{{"dotlessi", "i"}, {"uni0237", "j"}} {
		i := slices.IndexFunc(out, func(g glyphs.Glyph) bool { return g.Name == d.from })
		path, _ := withoutDot(out[i].Path, rp.XHeight)
		r := 'ı'
		if d.name == "uni0237" {
			r = 'ȷ'
		}
		out = append(out, glyphs.Glyph{Name: d.name, Rune: r, Advance: out[i].Advance, Path: path})
	}
	names := upper.Names()
	ups := make([]*upper.Glyph, len(names))
	errs := make([]error, len(names))
	gate := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for i, n := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate <- struct{}{}
			defer func() { <-gate }()
			up := p.Upper
			if isCap(n) {
				up = p.Caps
			}
			ups[i], errs[i] = upper.Build(up, n)
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	for _, g := range ups {
		out = append(out, glyphs.Glyph{Name: g.Name, Rune: g.Rune, Advance: g.Advance, Path: g.Path})
	}
	out = p.addSigns(out)
	out = addSpaces(out)
	out, err = p.addOrdinals(out)
	if err != nil {
		return nil, err
	}
	return p.addBarred(out)
}

// withoutDot returns the outline without the contours which lie wholly above
// the x-height and a little more, the dot of i and j, and the dot.
func withoutDot(path outline.Path, xh float64) (outline.Path, outline.Path) {
	var keep, dot outline.Path
	for _, c := range path {
		_, y0, _, _ := outline.Path{c}.BBox()
		if y0 > xh+40 {
			dot = append(dot, c)
		} else {
			keep = append(keep, c)
		}
	}
	return keep, dot
}

// Table returns the glyphs of the font: the bases with the side spaces of the
// fit table, then the spacing accents, the combining marks and the composed
// letters, with the anchors of the bases and the fields Base, Like and Marks
// of the composed letters.
func Table(p Params) ([]glyphs.Glyph, error) {
	gs, err := bases(p)
	if err != nil {
		return nil, err
	}
	if p.Fit {
		for i := range gs {
			if s, ok := fitTable[gs[i].Name]; ok {
				gs[i] = fitted(gs[i], s)
			}
		}
	}
	gs = derived(gs)
	gs, err = addLigatures(gs)
	if err != nil {
		return nil, err
	}
	gs, err = compose(p, gs)
	if err != nil {
		return nil, err
	}
	return finish(gs), nil
}

// finish gives the glyphs which are made of others, and are not composed, the
// anchors of the glyph they are made of and the kerning of the glyph they take
// it from.
func finish(gs []glyphs.Glyph) []glyphs.Glyph {
	by := map[string]int{}
	for i, g := range gs {
		by[g.Name] = i
	}
	for i, g := range gs {
		if l, ok := likeOf[g.Name]; ok {
			gs[i].Like = l
		}
	}
	gs[by["i.loclTRK"]].Anchors = gs[by["i"]].Anchors
	return gs
}

// Font builds the trial font, with the kerning of the model.
func Font(p Params, now time.Time) (f *sfnt.Font, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("assembly: %v", r)
		}
	}()
	gs, err := Table(p)
	if err != nil {
		return nil, err
	}
	return FontOf(p, gs, KernTable(gs), now)
}

// FontOf is Font for the glyphs gs and their kerns kt, as Table and
// KernTable compute them for p, for a caller which has these already.
func FontOf(p Params, gs []glyphs.Glyph, kt *kern.Table, now time.Time) (f *sfnt.Font, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("assembly: %v", r)
		}
	}()
	f, err = build.FontWith(Kerning(), Style(), gs, now, build.Options{Kern: kt})
	if err != nil {
		return nil, err
	}
	f.ItalicAngle = p.ItalicAngle()
	return f, nil
}

// kind returns the group of a glyph for the tables: "lower", "upper",
// "figure", "mark", "composed" or "sign".
func kind(g glyphs.Glyph) string {
	switch {
	case g.Base != "":
		return "composed"
	case len(g.Parts) > 0:
		return "ligature"
	case g.Rune == 0 || g.Rune == ' ':
		return "other"
	case unicode.Is(unicode.Mn, g.Rune) || unicode.Is(unicode.Lm, g.Rune) || unicode.Is(unicode.Sk, g.Rune):
		return "mark"
	case unicode.IsUpper(g.Rune):
		return "upper"
	case unicode.IsLetter(g.Rune):
		return "lower"
	case unicode.IsDigit(g.Rune):
		return "figure"
	}
	return "sign"
}
