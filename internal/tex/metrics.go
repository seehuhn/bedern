// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"math"
	"slices"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// A Lig is one ligature of a character: the character followed by Next
// becomes Result.
type Lig struct {
	Next, Result int
}

// A Kern is one kern of a character: the space Amount, in units of the em
// divided by 1000, is added between the character and Next.
type Kern struct {
	Next   int
	Amount float64
}

// A Char holds the metrics of the character of one position, in units of the
// em divided by 1000.
type Char struct {
	// Glyph is the index of the glyph in the glyph list.
	Glyph int

	Width, Height, Depth, Italic float64

	// Ligs and Kerns are in increasing order of Next.
	Ligs  []Lig
	Kerns []Kern
}

// Metrics holds what TeX needs to know about one font in one encoding.
type Metrics struct {
	Encoding *Encoding

	// Chars holds the characters; a position without a glyph is nil.
	Chars [256]*Char

	// The parameters of the font, in units of the em divided by 1000,
	// except Slant, which is a ratio.
	Slant, Space, Stretch, Shrink, XHeight, Quad, ExtraSpace float64

	// CapHeight, Ascender, Descender, CapStem and BaselineSkip are the
	// parameters 8, 9, 11, 15 and 16 of the font, which the seven of
	// TeX82 do not hold (see TFM).
	CapHeight, Ascender, Descender, CapStem, BaselineSkip float64

	// Bold and Italic say that the face byte of the metric file names a
	// bold weight and an italic slope.
	Bold, Italic bool
}

// inputLigs lists the ligatures of T1 which stand for characters that have no
// key: the quotation marks, the dashes, the guillemets and the inverted marks.
// Each entry gives the two characters and the result, as positions of T1.
var inputLigs = [][3]int{
	{0x60, 0x60, 0x10}, // two left quotes
	{0x27, 0x27, 0x11}, // two right quotes
	{0x2C, 0x2C, 0x12}, // two commas
	{0x3C, 0x3C, 0x13}, // two less-than signs
	{0x3E, 0x3E, 0x14}, // two greater-than signs
	{0x2D, 0x2D, 0x15}, // two hyphens
	{0x15, 0x2D, 0x16}, // en dash and hyphen
	{0x2D, 0x7F, 0x7F}, // hyphen and hyphen character
	{0x21, 0x60, 0xBD}, // exclamation mark and left quote
	{0x3F, 0x60, 0xBE}, // question mark and left quote
}

// A KernFunc returns the kern of the glyphs with the indices a and b, in
// units.
type KernFunc func(a, b int) int

// NewMetrics returns the metrics of the glyphs gs in the encoding e.
//
// The width of a character is the advance of its glyph.  The height and the
// depth are the extent of the outline above and below the baseline, with the
// overshoot of a round top or bottom left out (see nominal), and the italic
// correction is the extent of the outline beyond the advance.  The compound
// word marks have the heights of the lines for which they stand.
//
// In T1 the ligatures are those of the glyphs which have parts, each as the
// pair of the ligature of all parts but the last, or of the first part, and
// the last part, and the ligatures of inputLigs.  The kerns are those which
// kern gives for the pairs of characters; a pair with a ligature has no kern.
//
// The interword space is the advance of the space, its stretch is half of it
// and its shrink and the extra space are a third of it.
func NewMetrics(e *Encoding, p design.Params, gs []glyphs.Glyph, kern KernFunc) *Metrics {
	m := &Metrics{Encoding: e}
	index, _ := e.Resolve(gs)
	for c, i := range index {
		if i < 0 {
			continue
		}
		g := gs[i]
		ch := &Char{Glyph: i, Width: g.Advance}
		if len(g.Path) > 0 {
			_, y0, x1, y1 := g.Path.BBox()
			ch.Height, ch.Depth = nominal(p, y1, y0)
			ch.Italic = math.Max(0, x1-g.Advance)
		}
		switch g.Name {
		case "uni200C":
			ch.Height = p.XHeight
		case "cwmcapital":
			ch.Height = p.CapHeight
		case "cwmascender":
			ch.Height = p.Ascender
		}
		m.Chars[c] = ch
	}

	if e == T1 {
		slotOf := make(map[int]int) // glyph index to the first position
		for c := 255; c >= 0; c-- {
			if index[c] >= 0 {
				slotOf[index[c]] = c
			}
		}
		byName := make(map[string]int, len(gs))
		byParts := make(map[string]int)
		for i, g := range gs {
			byName[g.Name] = i
			if len(g.Parts) > 0 {
				byParts[partsKey(g.Parts)] = i
			}
		}
		add := func(a, b, res int) {
			if m.Chars[a] == nil || m.Chars[b] == nil || m.Chars[res] == nil {
				return
			}
			m.Chars[a].Ligs = append(m.Chars[a].Ligs, Lig{Next: b, Result: res})
		}
		for i, g := range gs {
			n := len(g.Parts)
			res, ok := slotOf[i]
			if n < 2 || !ok {
				continue
			}
			first, ok := byName[g.Parts[0]]
			if n > 2 {
				first, ok = byParts[partsKey(g.Parts[:n-1])]
			}
			last, okLast := byName[g.Parts[n-1]]
			a, okA := slotOf[first]
			b, okB := slotOf[last]
			if ok && okLast && okA && okB {
				add(a, b, res)
			}
		}
		for _, l := range inputLigs {
			add(l[0], l[1], l[2])
		}
	}

	for _, ca := range m.Chars {
		if ca != nil {
			slices.SortFunc(ca.Ligs, func(x, y Lig) int { return x.Next - y.Next })
		}
	}
	forms := m.ligatureForms()
	for _, ca := range m.Chars {
		if ca == nil {
			continue
		}
		for b, cb := range m.Chars {
			if cb == nil || slices.ContainsFunc(ca.Ligs, func(l Lig) bool { return l.Next == b }) {
				continue
			}
			k := float64(kern(ca.Glyph, cb.Glyph))
			if fs := forms[b]; len(fs) > 0 {
				var ks []float64
				for _, f := range fs {
					ks = append(ks, float64(kern(ca.Glyph, m.Chars[f].Glyph)))
				}
				k = firstKern(b, k, ks)
			}
			if k != 0 {
				ca.Kerns = append(ca.Kerns, Kern{Next: b, Amount: k})
			}
		}
	}

	for _, g := range gs {
		if g.Name == "space" {
			m.Space = g.Advance
		}
	}
	m.Stretch = m.Space / 2
	m.Shrink = m.Space / 3
	m.ExtraSpace = m.Space / 3
	m.XHeight = p.XHeight
	m.Quad = p.UnitsPerEm
	m.CapHeight = p.CapHeight
	m.Ascender = p.Ascender
	m.Descender = p.Descender
	m.CapStem = p.Caps.Stem
	m.BaselineSkip = p.TypoAscender - p.TypoDescender + p.LineGap
	return m
}

// partsKey returns a key for the letters of a ligature.
func partsKey(parts []string) string {
	key := ""
	for _, s := range parts {
		key += s + "\x00"
	}
	return key
}

// nominal returns the height and the depth of a character whose outline
// reaches from bottom to top.  A top which lies above the x-height, the height
// of the figures or the height of the capitals by no more than the overshoot
// of that line has the height of the line, unless it lies on the line of the
// ascenders; a bottom which lies below the baseline or the line of the
// descenders by no more than an overshoot has the depth of the line.
func nominal(p design.Params, top, bottom float64) (height, depth float64) {
	height = math.Max(0, top)
	if top != p.Ascender {
		for _, l := range [][2]float64{
			{p.XHeight, p.OvershootLower},
			{p.Figures.Height, p.Figures.OverTop},
			{p.CapHeight, p.OvershootUpper},
		} {
			if top > l[0] && top <= l[0]+l[1] {
				height = l[0]
				break
			}
		}
	}
	over := math.Max(p.OvershootLower, p.OvershootUpper)
	depth = math.Max(0, -bottom)
	switch {
	case depth <= over:
		depth = 0
	case depth > p.Descender && depth <= p.Descender+p.OvershootLower:
		depth = p.Descender
	}
	return height, depth
}

// kernTolerance is the difference of two kerns, in units, which is too small
// to see: where the kerns of a character and of the ligatures which it begins
// differ by no more, the character keeps its own kern.
const kernTolerance = 5

// ligatureForms returns, for each character which begins a ligature, the
// characters which it can become: the results of its ligatures and of theirs
// in turn.  A ligature whose result is the character which follows (the hyphen
// character of T1) forms no new shape and does not count.
func (m *Metrics) ligatureForms() map[int][]int {
	forms := make(map[int][]int)
	var walk func(c int, seen map[int]bool, out *[]int)
	walk = func(c int, seen map[int]bool, out *[]int) {
		for _, l := range m.Chars[c].Ligs {
			if l.Result == l.Next || seen[l.Result] {
				continue
			}
			seen[l.Result] = true
			*out = append(*out, l.Result)
			walk(l.Result, seen, out)
		}
	}
	for c, ch := range m.Chars {
		if ch == nil || len(ch.Ligs) == 0 {
			continue
		}
		var out []int
		walk(c, map[int]bool{c: true}, &out)
		if len(out) > 0 {
			slices.Sort(out)
			forms[c] = out
		}
	}
	return forms
}

// firstKern returns the kern which a letter takes before the character c,
// given the kern k of the pair and the kerns ks of the letter with the
// ligatures which c begins.
//
// TeX looks at a pair of characters before it knows whether the second one
// begins a ligature, and it keeps the kern which it inserted when the
// ligature forms, so that one kern stands for the character and for all the
// ligatures which it begins.  Which kern that is depends on the character:
//
//   - the less-than and the greater-than sign, which a text does not set
//     alone (the guillemets stand for them), take the kern of the first
//     ligature;
//   - the exclamation and the question mark, whose ligatures with a left
//     quotation mark do not follow a letter in a text, keep their own kern;
//   - every other character takes the kern of least size among its own and
//     those of its ligatures (the hyphen of the hyphen, the en dash and the
//     em dash), so that no form stands closer to the letter than its own kern
//     allows.
//
// Kerns which differ by no more than kernTolerance are not corrected.
func firstKern(c int, k float64, ks []float64) float64 {
	spread := 0.0
	for _, x := range ks {
		spread = math.Max(spread, math.Abs(x-k))
	}
	if spread <= kernTolerance {
		return k
	}
	switch c {
	case '!', '?':
		return k
	case '<', '>':
		k = ks[0]
		ks = ks[1:]
	}
	for _, x := range ks {
		if math.Abs(x) < math.Abs(k) {
			k = x
		}
	}
	return k
}
