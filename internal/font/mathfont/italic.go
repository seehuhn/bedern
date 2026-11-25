// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"io"
	"math"
	"runtime"
	"strings"
	"sync"

	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/sfnt/opentype/mathtable"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
	"seehuhn.de/go/bedern/internal/font/italic/upper"
)

// The maths italic: the 52 letters of mathematical italic at
// U+1D434 to U+1D467 with h at U+210E, the dotless i and j at U+1D6A4 and
// U+1D6A5, and ℏ.  The letters are those of the italic, built by
// bedern/internal/font/italic (the lowercase) and bedern/internal/font/italic/upper (the capitals)
// with the side spaces of the fit of the italic (bedern/internal/font/italic/assembly),
// and respaced for formulas.  Each has two script forms, the letter drawn
// for the first and the second script level.  The italics correction, the
// top accent attachment and the math kerns are computed from the outline of
// each glyph by the rules below.
const (
	// mathSide is what every lowercase letter gains on each side over the
	// fit of the text italic.  The right side is at least zero: x gains 34
	// there and c 18.
	mathSide = 15
	// maxOverhangLeft and maxOverhangRight are how far the ink of a letter
	// may reach before its origin and beyond its advance.  The tails of f and
	// j which reach further come within 20 units of an opening bracket before
	// them, and the hook of f of a closing one after it.
	maxOverhangLeft  = 40
	maxOverhangRight = 50
	// icPad is added to the overhang of the ink beyond the advance above
	// half the height of a letter for its italics correction, as in the text
	// italic.
	icPad = 10
	// accentLower and accentUpper move the top accent attachment of the
	// lowercase and of the capitals right of the middle of the ink of their
	// top band, as fractions of the x-height.
	accentLower = 0.05
	accentUpper = 0.02
	// accentBowl is the place of the accent of b, h and k between the middle
	// of the top of the ascender (0) and the middle of the ink between 0.80
	// and 0.95 of the x-height (1).
	accentBowl = 0.5
	// kernGap is the white which a math kern leaves between a letter and its
	// script: the space after a script (SpaceAfterScript).
	kernGap = 50
	// kernMin is the least kern which is written.
	kernMin = 20
)

// capLeft is what the capitals whose top reaches far right over the left
// edge of their advance lose on the left against the text italic: T, U, V, W
// and Y move left in their advance by 73.5, 38, 71, 76.5 and 72.
var capLeft = map[string]float64{"T": -73, "U": -38, "V": -71, "W": -76, "Y": -72}

// The script levels.  The stems are 1.13 and 1.24 times as heavy as at text
// size, the thinnest strokes at least 36 and 50, the counter of n 1.10 and
// 1.21 times as wide, for which the centre lines widen by 1.11 and 1.22, and
// the advances 1.10 and 1.22 times that of the text glyph.  A letter whose
// pen cannot take the stem factor in a turn of its centre line takes the
// largest factor in steps of 0.02 down to the lowest with which it builds.
//
// The upright glyphs (the figures and the Greek capitals) are drawn by the
// constructions of the regular with a script parameter set (script.go): the
// factors thinParam of the thin strokes of the regular's parameters, and out
// the share of the weight added to a stroke which widens a glyph (the rest
// goes into its counters; the bold's is 0.6), the advance being that of the
// text form times advance.
var scriptLevels = [2]struct {
	stem, minStem, thin, minThin, width, advance, thinParam, out float64
	figStem, capStem, capOut, greekAdvance, oneWide              float64
}{
	{stem: 1.13, minStem: 1.08, thin: 36, minThin: 36, width: 1.11, advance: 1.10, thinParam: 1.15, out: 1.8,
		figStem: 1.10, capStem: 1.09, capOut: 3.7, greekAdvance: 1.12, oneWide: 1.08},
	{stem: 1.24, minStem: 1.15, thin: 50, minThin: 40, width: 1.22, advance: 1.22, thinParam: 1.40, out: 2.0,
		figStem: 1.16, capStem: 1.17, capOut: 3.6, greekAdvance: 1.27, oneWide: 1.13},
}

// An italicLetter is a letter of the maths italic.
type italicLetter struct {
	name string // the glyph name
	r    rune
	from string // the letter of the italic
	cap  bool
}

// italicLetters returns the letters of the maths italic in the order of the
// font: the capitals, the lowercase and the dotless i and j.
func italicLetters() []italicLetter {
	var out []italicLetter
	for i := range 26 {
		r := rune(0x1D434 + i)
		out = append(out, italicLetter{name: glyphName(r), r: r, from: string(rune('A' + i)), cap: true})
	}
	for i := range 26 {
		r := rune(0x1D44E + i)
		if i == 'h'-'a' {
			r = 0x210E
		}
		out = append(out, italicLetter{name: glyphName(r), r: r, from: string(rune('a' + i))})
	}
	out = append(out, italicLetter{name: glyphName(0x1D6A4), r: 0x1D6A4, from: "dotlessi"},
		italicLetter{name: glyphName(0x1D6A5), r: 0x1D6A5, from: "uni0237"})
	return out
}

// glyphName returns the name of the glyph of a code point.
func glyphName(r rune) string {
	if r > 0xFFFF {
		return fmt.Sprintf("u%05X", r)
	}
	return fmt.Sprintf("uni%04X", r)
}

// An italicForm is one form of a letter of the maths italic as it is built,
// with what it needs to build it.
type italicForm struct {
	Glyph
	Level    int      // 0 at text size, 1 and 2 for the script levels
	From     string   // the letter of the italic
	Stem     float64  // the factor of the stems, 1 at text size
	Warnings []string // the warnings of the stroke tool
	Findings []string // the findings of the italic profile of the smoothness check
}

// italicForms returns every form of the maths italic, the text forms first,
// then those of the first and the second script level, each in the order
// of italicLetters, and ℏ last.
var italicForms = sync.OnceValues(func() ([]italicForm, error) {
	ls := italicLetters()
	out := make([]italicForm, 3*len(ls))
	errs := make([]error, len(out))
	gate := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for level := range 3 {
		for i, l := range ls {
			wg.Add(1)
			go func() {
				defer wg.Done()
				gate <- struct{}{}
				defer func() { <-gate }()
				out[level*len(ls)+i], errs[level*len(ls)+i] = buildForm(l, level)
			}()
		}
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	// The script forms take their advance from the text form.
	for level := 1; level <= 2; level++ {
		for i := range ls {
			f := &out[level*len(ls)+i]
			fitScript(f, out[i], scriptLevels[level-1].advance)
		}
	}
	for i := range out {
		finish(&out[i])
	}
	for i, l := range ls {
		out[i].Script = [2]string{l.name + ".s1", l.name + ".s2"}
	}
	hbar, err := planck(out[26+'h'-'a'])
	if err != nil {
		return nil, err
	}
	return append(out, hbar), nil
})

// buildForm builds one form of a letter, with the side spaces of the maths
// italic at text size and the side spaces of the fit of the italic for the
// script forms, which fitScript then widens.
func buildForm(l italicLetter, level int) (italicForm, error) {
	f := italicForm{Level: level, From: l.from, Stem: 1}
	f.Name, f.Rune, f.Class = l.name, l.r, Ord
	if level > 0 {
		f.Name, f.Rune = fmt.Sprintf("%s.s%d", l.name, level), 0
	}
	src := l.from
	switch src {
	case "dotlessi":
		src = "i"
	case "uni0237":
		src = "j"
	}
	var g glyphs.Glyph
	switch {
	case l.cap:
		cp := assembly.Default().Caps
		if level > 0 {
			s := scriptLevels[level-1]
			cp.Weight *= s.stem
			cp.Width *= s.width
			cp.CapStem = math.Round(cp.CapStem * s.stem)
			cp.ThinFloor = s.thin
			f.Stem = s.stem
		}
		u, err := upper.Build(cp, src)
		if err != nil {
			return f, err
		}
		g = glyphs.Glyph{Name: l.from, Rune: rune(l.from[0]), Advance: u.Advance, Path: u.Path}
	default:
		ig, stem, err := buildLower(src, level, false)
		if err != nil {
			return f, err
		}
		f.Stem = stem
		f.Warnings = ig.Warnings
		for _, fd := range ig.Findings {
			f.Findings = append(f.Findings, fmt.Sprintf("%s at %.0f,%.0f: %s", fd.Kind, fd.X, fd.Y, fd.Detail))
		}
		g = glyphs.Glyph{Name: l.from, Rune: ig.Rune, Advance: ig.Advance, Path: ig.Path}
		if src != l.from {
			g.Path = assembly.WithoutDot(g.Path, design.Regular().XHeight)
			g.Rune = 'ı'
		}
	}
	s, ok := assembly.FitSides(l.from)
	if !ok {
		return f, fmt.Errorf("mathfont: the italic has no side spaces for %s", l.from)
	}
	if !l.cap && level == 0 {
		s.Left += mathSide
		s.Right = math.Max(s.Right+mathSide, 0)
	}
	if level == 0 {
		s.Left += capLeft[l.from]
	}
	g = assembly.Fitted(g, s)
	f.Path, f.Advance = g.Path, g.Advance
	if level == 0 {
		f.Path, f.Advance = clearOfBrackets(withinOverhang(f.Path, f.Advance))
	}
	return f, nil
}

// narrowStart are the letters, by name and level, whose search for a script
// form starts with the centre lines already narrowed by this much: the
// searches from the full width take seconds before they find these.
var narrowStart = map[string]float64{}

// extraWidth are the letters, by name, whose centre lines widen by this much
// more than the rule of the level, at both levels: ν, which the text form
// draws as wide as v; a script ν is as wide as v at the first level.
var extraWidth = map[string]float64{"nu": 0.25}

// ScriptForm returns the script of the first try of buildLower for the
// letter at the level (1 or 2): the width, the stem and the thin of the
// level, for ν with its extra width.
func ScriptForm(name string, level int) italic.Script {
	s := scriptLevels[level-1]
	return italic.Script{Width: s.width + narrowStart[fmt.Sprintf("%s/%d", name, level)] + extraWidth[name], Stem: s.stem, Thin: s.thin, Level: level}
}

// buildLower builds a lowercase letter of the italic at the level, and
// returns it with the factor of its stems.  A letter whose pen does not fit
// at the width of the level (the hooks of ζ, ξ, ς and μ at the second level)
// is drawn with the centre lines widened by less, down to 0.10 less; a hidden
// end which the text form has too does not refuse a script form.  With
// quick the smoothness check of the outline is left out.
func buildLower(name string, level int, quick bool) (*italic.Glyph, float64, error) {
	p := italic.Default()
	if level == 0 {
		g, err := italic.Build(p, name)
		return g, 1, err
	}
	s := scriptLevels[level-1]
	var last error
	textBroken := sync.OnceValue(func() bool {
		g, err := italic.Build(p, name)
		return err == nil && brokenBuild(g.Warnings)
	})
	for dw := narrowStart[fmt.Sprintf("%s/%d", name, level)]; dw >= -0.10-1e-9; dw -= 0.02 {
		for thin := s.thin; thin >= s.minThin-1e-9; thin -= 2 {
			for stem := s.stem; stem >= s.minStem-1e-9; stem -= 0.02 {
				g, err := italic.BuildScript(p, name, italic.Script{Width: s.width + dw + extraWidth[name], Stem: stem, Thin: thin, NoFindings: quick, Level: level})
				if err == nil && (!brokenBuild(g.Warnings) || textBroken()) {
					return g, stem, nil
				}
				last = err
				if err == nil {
					last = fmt.Errorf("%s", strings.Join(g.Warnings, "; "))
				}
			}
		}
	}
	return nil, 0, fmt.Errorf("mathfont: no script form of %s at level %d: %v", name, level, last)
}

// brokenBuild says whether the warnings of the stroke tool name a part of
// the outline which does not belong to the shape: an end which lies
// inside another stroke and shows.
func brokenBuild(warnings []string) bool {
	for _, w := range warnings {
		if strings.Contains(w, "hidden end shows") || strings.Contains(w, "is open") {
			return true
		}
	}
	return false
}

// withinOverhang returns the outline moved and the advance grown until the
// ink reaches no more than maxOverhangLeft left of the origin and
// maxOverhangRight beyond the advance.
func withinOverhang(path outline.Path, adv float64) (outline.Path, float64) {
	x0, _, x1, _ := path.BBox()
	if d := -x0 - maxOverhangLeft; d > 0 {
		d = math.Ceil(d)
		path = shifted(path, d, 0)
		adv += d
		x1 += d
	}
	if d := x1 - adv - maxOverhangRight; d > 0 {
		adv += math.Ceil(d)
	}
	return path, adv
}

// bracketGap is the least white between a letter and a parenthesis or a
// bracket of the regular set next to it without space, the floor of the
// kerning in text.
const bracketGap = 20

// brackets returns the outlines and advances of the opening and the closing
// parenthesis and bracket of the maths font, which are those of the regular
// moved onto the axis (centredOnAxis).
var brackets = sync.OnceValue(func() (out [2][]glyphs.Glyph) {
	p := design.Regular()
	for i, ns := range [2][]string{{"parenleft", "bracketleft"}, {"parenright", "bracketright"}} {
		for _, n := range ns {
			path, adv, _ := glyphs.Outline(n, p)
			out[i] = append(out[i], glyphs.Glyph{Name: n, Advance: adv, Path: centredOnAxis(path, p.Punct.MathAxis).Round()})
		}
	}
	return out
})

// clearOfBrackets returns the outline moved and the advance grown until an
// opening parenthesis or bracket left of the letter and a closing one after
// it keep bracketGap of white to its ink.
func clearOfBrackets(path outline.Path, adv float64) (outline.Path, float64) {
	bs := brackets()
	for range 4 {
		d := math.Inf(1)
		for _, b := range bs[0] {
			d = math.Min(d, inkDistance(b.Path, shifted(path, b.Advance, 0)))
		}
		if d >= bracketGap {
			break
		}
		dx := math.Ceil(bracketGap - d)
		path, adv = shifted(path, dx, 0), adv+dx
	}
	for range 4 {
		d := math.Inf(1)
		for _, b := range bs[1] {
			d = math.Min(d, inkDistance(path, shifted(b.Path, adv, 0)))
		}
		if d >= bracketGap {
			break
		}
		adv += math.Ceil(bracketGap - d)
	}
	return path, adv
}

// inkDistance returns the least distance between the outlines of a and b,
// or zero where they overlap.
func inkDistance(a, b outline.Path) float64 {
	ax0, ay0, ax1, ay1 := a.BBox()
	bx0, by0, bx1, by1 := b.BBox()
	var pa, pb []outline.Point
	for _, c := range a {
		pa = append(pa, c.Flatten()...)
	}
	for _, c := range b {
		pb = append(pb, c.Flatten()...)
	}
	if ax0 < bx1 && bx0 < ax1 && ay0 < by1 && by0 < ay1 {
		for _, q := range pb {
			if inside(a, q) {
				return 0
			}
		}
		for _, q := range pa {
			if inside(b, q) {
				return 0
			}
		}
	}
	d := math.Inf(1)
	for _, p := range pa {
		if p.X < bx0-d || p.X > bx1+d || p.Y < by0-d || p.Y > by1+d {
			continue
		}
		for _, q := range pb {
			d = math.Min(d, math.Hypot(p.X-q.X, p.Y-q.Y))
		}
	}
	return d
}

// inside says whether the point lies in the ink, by the crossings of the
// horizontal through it.
func inside(p outline.Path, q outline.Point) bool {
	n := 0
	for _, x := range p.CrossingsX(q.Y) {
		if x < q.X {
			n++
		}
	}
	return n%2 == 1
}

// fitScript gives a script form the advance of its text form times the
// factor, with the white which the advance adds beyond the ink of the
// script form divided equally between its two sides, within the overhangs.
func fitScript(f *italicForm, text italicForm, factor float64) {
	src := glyphs.Glyph{Name: f.From, Rune: rune(f.From[0]), Advance: f.Advance, Path: f.Path}
	if f.From == "dotlessi" || f.From == "uni0237" {
		src.Name = f.From
	}
	tg := glyphs.Glyph{Name: src.Name, Rune: src.Rune, Advance: text.Advance, Path: text.Path}
	ts, ok1 := assembly.SidesOf(tg)
	ss, ok2 := assembly.SidesOf(src)
	if !ok1 || !ok2 {
		return
	}
	inkScript := f.Advance - ss.Left - ss.Right
	e := factor*text.Advance - inkScript - ts.Left - ts.Right
	g := assembly.Fitted(src, assembly.Sides{Left: ts.Left + e/2, Right: ts.Right + e/2})
	f.Path, f.Advance = clearOfBrackets(withinOverhang(g.Path, g.Advance))
}

// finish computes the italics correction, the top accent attachment and the
// math kerns of a form from its outline.
func finish(f *italicForm) {
	g := glyphs.Glyph{Name: f.Name, Rune: f.Rune, Advance: f.Advance, Path: f.Path}
	f.Italic = assembly.ItalicCorrection(g, icPad)
	f.Attach, f.Attached = accentAttachment(f.From, f.Path), true
	f.Kern = mathKerns(f.From, f.Path, f.Advance)
}

// accentAttachment returns the top accent attachment of a letter of the
// italic from its outline: the middle of its ink in the top 15 per cent of
// its own height, moved right by accentLower or accentUpper of the x-height;
// for b, h and k the point accentBowl of the way from there to the middle
// of the ink between 0.80 and 0.95 of the x-height.
func accentAttachment(from string, path outline.Path) float64 {
	xh := design.Regular().XHeight
	_, _, _, top := path.BBox()
	lo, hi := inkSpan(path, 0.85*top, top)
	a := (lo + hi) / 2
	switch {
	case from >= "A" && from <= "Z":
		return math.Round(a + accentUpper*xh)
	case from == "b" || from == "h" || from == "k":
		blo, bhi := inkSpan(path, 0.80*xh, 0.95*xh)
		a += accentBowl * ((blo+bhi)/2 - a)
		return math.Round(a + accentLower*xh)
	}
	return math.Round(a + accentLower*xh)
}

// kernedLetters are the letters of the italic which may get math kerns at
// the bottom right: the capitals F, P, T, V, W and Y, which have an open
// corner at the lower right, A and L, and f and j, whose tails leave the
// corner open.  A corner gets a kern only where the outline leaves white for
// it (kernMin).  The top right gets none.
var kernedLetters = map[string]bool{
	"A": true, "F": true, "L": true, "P": true, "T": true, "V": true, "W": true, "Y": true,
	"f": true, "j": true,
}

// bottomKernHeights are the heights at which the kerns at the bottom right
// change: those of the top of a subscript, in steps of 100 (the top of a
// lowercase subscript lies at 94, of a capital or a figure at 258).
var bottomKernHeights = []float64{100, 200, 300}

// kernMax is the largest kern, less than the white which the outlines of
// F, P, T, V, W, Y and f leave.
const kernMax = 120

// mathKerns returns the math kerns of a letter at the bottom right: for the
// interval between two heights of bottomKernHeights, the white between the
// advance and the ink from the lowest ink up to the top of the interval, the
// top of a subscript in the interval, less kernGap and at most kernMax.  The
// kern is negative, and zero where it is less than kernMin.
func mathKerns(from string, path outline.Path, adv float64) *mathtable.KernInfo {
	if !kernedLetters[from] {
		return nil
	}
	_, y0, _, y1 := path.BBox()
	white := func(hi float64) float64 {
		w := math.Inf(1)
		for y := y0; y <= hi; y += 4 {
			if _, x := inkSpan(path, y, y); !math.IsInf(x, 0) {
				w = math.Min(w, adv-x)
			}
		}
		return w
	}
	k := &mathtable.Kern{}
	any := false
	for i := 0; i <= len(bottomKernHeights); i++ {
		hi := y1
		if i < len(bottomKernHeights) {
			hi = bottomKernHeights[i]
		}
		v := 0.0
		if w := white(math.Max(hi, y0)) - kernGap; !math.IsInf(w, 1) && w >= kernMin {
			v = -math.Round(math.Min(w, kernMax))
		}
		any = any || v != 0
		k.Kerns = append(k.Kerns, mathtable.Value{V: funit.Int16(v)})
	}
	if !any {
		return nil
	}
	for _, h := range bottomKernHeights {
		k.Heights = append(k.Heights, mathtable.Value{V: funit.Int16(h)})
	}
	return &mathtable.KernInfo{BottomRight: k}
}

// planck returns ℏ: the maths italic h with a bar of the hairline of the
// italic at 0.75 of the ascender, horizontal, reaching as far beyond the
// edges of the stem as the crossbar of t.
func planck(h italicForm) (italicForm, error) {
	p := italic.Default()
	y := 0.75 * p.Ascender
	lo, hi := math.Inf(1), math.Inf(-1)
	xs := h.Path.CrossingsX(y)
	if len(xs) < 2 {
		return italicForm{}, fmt.Errorf("mathfont: no stem of h at %.0f", y)
	}
	lo, hi = xs[0], xs[1]
	bar := outline.Path{box(lo-p.CrossbarLeft, y-p.Hair/2, hi+p.CrossbarRight, y+p.Hair/2)}
	f := h
	f.Name, f.Rune, f.Script = "uni210F", 0x210F, [2]string{}
	f.Path = glyphs.Union(h.Path, bar)
	f.Path, f.Advance = withinOverhang(f.Path, h.Advance)
	finish(&f)
	return f, nil
}

// inkSpan returns the leftmost and the rightmost ink of the outline between
// the heights y0 and y1, and infinities where there is none.
func inkSpan(path outline.Path, y0, y1 float64) (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, ct := range path {
		pts := ct.Flatten()
		n := len(pts)
		for i := range n {
			a, b := pts[i], pts[(i+1)%n]
			if a.Y >= y0 && a.Y <= y1 {
				lo, hi = math.Min(lo, a.X), math.Max(hi, a.X)
			}
			for _, y := range []float64{y0, y1} {
				if (a.Y < y) != (b.Y < y) {
					x := a.X + (y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
					lo, hi = math.Min(lo, x), math.Max(hi, x)
				}
			}
		}
	}
	return lo, hi
}

// italicGlyphs returns the glyphs of the maths italic in the order of the
// font.
func italicGlyphs() ([]Glyph, error) {
	fs, err := italicForms()
	if err != nil {
		return nil, err
	}
	out := make([]Glyph, len(fs))
	for i, f := range fs {
		out[i] = f.Glyph
	}
	return out, nil
}

// WriteItalicTable writes for every form of the maths italic its name, the
// letter of the italic, its level, advance, side spaces (as the fit of the
// italic measures them), overhangs, italics correction, top accent
// attachment, the factor of its stems, the number of warnings of the stroke
// tool and of findings of the smoothness check, and its kerns.
func WriteItalicTable(w io.Writer) error {
	fs, err := italicForms()
	if err != nil {
		return err
	}
	for _, f := range fs {
		g := glyphs.Glyph{Name: f.From, Rune: rune(f.From[0]), Advance: f.Advance, Path: f.Path}
		if f.From == "dotlessi" || f.From == "uni0237" {
			g.Name = f.From
		}
		s, _ := assembly.SidesOf(g)
		x0, _, x1, _ := f.Path.BBox()
		kern := ""
		if f.Kern != nil {
			for _, c := range []struct {
				tag string
				k   *mathtable.Kern
			}{{"BR", f.Kern.BottomRight}, {"TR", f.Kern.TopRight}} {
				if c.k == nil {
					continue
				}
				kern += " " + c.tag
				for _, v := range c.k.Kerns {
					kern += fmt.Sprintf(" %d", v.V)
				}
			}
		}
		if _, err := fmt.Fprintf(w, "%-12s %-8s %d %5.0f %5.0f %5.0f  over %4.0f %4.0f  ic %4.0f  acc %4.0f  stem %.2f  warn %d fair %d%s\n",
			f.Name, f.From, f.Level, f.Advance, s.Left, s.Right, -x0, x1-f.Advance, f.Italic, f.Attach, f.Stem,
			len(f.Warnings), len(f.Findings), kern); err != nil {
			return err
		}
	}
	return nil
}
