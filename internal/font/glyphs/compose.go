// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"slices"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letters composed of a base letter and a mark, the dotless i and j, the
// spacing accents and the combining marks.
//
// A composed letter is its base letter, unchanged, with one mark (marks.go)
// whose anchor is set on an anchor of the base, moved by whole units, so that
// the rounded letter is the rounded base with the rounded mark.  It has the
// advance and the side kinds of its base.  The cedilla and the ogonek are
// joined to the base where they meet it: the letter is the union of the two,
// and the points where their edges cross are rounded to whole units.
//
// The anchors of a base are these.  The top anchor stands at the height of the
// x-height for the lowercase, of the cap height for the capitals and of the
// ascender height for the letters with ascenders; the mark has its lowest ink
// Acc.Gap above the x-height, or Acc.GapCap above the cap height and the
// ascender height, where the capital form of the mark stands.  Its abscissa
// is the middle of the ink of the base between the baseline and the x-height
// or the cap height, except where topBand gives another band: the middle of
// the stem below the head serif, moved by design.DotOffset as the dot of i, for
// the letters whose top is a stem (i, j and l, and h), the upper part of a and
// the bowl of g.  The bottom anchor stands at the middle of the ink of the
// base, of its stem (t, l, r and the capitals T, L) or of its foot (D), at the lowest ink of
// the base there for the cedilla and on the baseline for the comma below; the
// cedilla of t stands at the lowest point of t instead.  The ogonek anchor
// stands at the lowest ink of the base where ogonekX puts it: under the stem
// of a, u, i and I, near the right end of the foot of E and A, and right of
// the lowest point of e and U.
// The side anchor of d, l, t and L stands Acc.ApoGap to the right of the right
// edge of the stem, at the height of its top.

// A composite is a letter made of a base letter and a mark.
type composite struct {
	base string
	mark markKind
}

// composedList lists the composed letters in glyph order.
var composedList = []struct {
	name string
	r    rune
	composite
}{
	// Latin-1 Supplement.
	{"Agrave", 'À', composite{"A", markGrave}},
	{"Aacute", 'Á', composite{"A", markAcute}},
	{"Acircumflex", 'Â', composite{"A", markCircumflex}},
	{"Atilde", 'Ã', composite{"A", markTilde}},
	{"Adieresis", 'Ä', composite{"A", markDieresis}},
	{"Aring", 'Å', composite{"A", markRing}},
	{"Ccedilla", 'Ç', composite{"C", markCedilla}},
	{"Egrave", 'È', composite{"E", markGrave}},
	{"Eacute", 'É', composite{"E", markAcute}},
	{"Ecircumflex", 'Ê', composite{"E", markCircumflex}},
	{"Edieresis", 'Ë', composite{"E", markDieresis}},
	{"Igrave", 'Ì', composite{"I", markGrave}},
	{"Iacute", 'Í', composite{"I", markAcute}},
	{"Icircumflex", 'Î', composite{"I", markCircumflex}},
	{"Idieresis", 'Ï', composite{"I", markDieresis}},
	{"Ntilde", 'Ñ', composite{"N", markTilde}},
	{"Ograve", 'Ò', composite{"O", markGrave}},
	{"Oacute", 'Ó', composite{"O", markAcute}},
	{"Ocircumflex", 'Ô', composite{"O", markCircumflex}},
	{"Otilde", 'Õ', composite{"O", markTilde}},
	{"Odieresis", 'Ö', composite{"O", markDieresis}},
	{"Ugrave", 'Ù', composite{"U", markGrave}},
	{"Uacute", 'Ú', composite{"U", markAcute}},
	{"Ucircumflex", 'Û', composite{"U", markCircumflex}},
	{"Udieresis", 'Ü', composite{"U", markDieresis}},
	{"Yacute", 'Ý', composite{"Y", markAcute}},
	{"agrave", 'à', composite{"a", markGrave}},
	{"aacute", 'á', composite{"a", markAcute}},
	{"acircumflex", 'â', composite{"a", markCircumflex}},
	{"atilde", 'ã', composite{"a", markTilde}},
	{"adieresis", 'ä', composite{"a", markDieresis}},
	{"aring", 'å', composite{"a", markRing}},
	{"ccedilla", 'ç', composite{"c", markCedilla}},
	{"egrave", 'è', composite{"e", markGrave}},
	{"eacute", 'é', composite{"e", markAcute}},
	{"ecircumflex", 'ê', composite{"e", markCircumflex}},
	{"edieresis", 'ë', composite{"e", markDieresis}},
	{"igrave", 'ì', composite{"dotlessi", markGrave}},
	{"iacute", 'í', composite{"dotlessi", markAcute}},
	{"icircumflex", 'î', composite{"dotlessi", markCircumflex}},
	{"idieresis", 'ï', composite{"dotlessi", markDieresis}},
	{"ntilde", 'ñ', composite{"n", markTilde}},
	{"ograve", 'ò', composite{"o", markGrave}},
	{"oacute", 'ó', composite{"o", markAcute}},
	{"ocircumflex", 'ô', composite{"o", markCircumflex}},
	{"otilde", 'õ', composite{"o", markTilde}},
	{"odieresis", 'ö', composite{"o", markDieresis}},
	{"ugrave", 'ù', composite{"u", markGrave}},
	{"uacute", 'ú', composite{"u", markAcute}},
	{"ucircumflex", 'û', composite{"u", markCircumflex}},
	{"udieresis", 'ü', composite{"u", markDieresis}},
	{"yacute", 'ý', composite{"y", markAcute}},
	{"ydieresis", 'ÿ', composite{"y", markDieresis}},

	// Latin Extended-A.
	{"Amacron", 'Ā', composite{"A", markMacron}},
	{"amacron", 'ā', composite{"a", markMacron}},
	{"Abreve", 'Ă', composite{"A", markBreve}},
	{"abreve", 'ă', composite{"a", markBreve}},
	{"Aogonek", 'Ą', composite{"A", markOgonek}},
	{"aogonek", 'ą', composite{"a", markOgonek}},
	{"Cacute", 'Ć', composite{"C", markAcute}},
	{"cacute", 'ć', composite{"c", markAcute}},
	{"Ccircumflex", 'Ĉ', composite{"C", markCircumflex}},
	{"ccircumflex", 'ĉ', composite{"c", markCircumflex}},
	{"Cdotaccent", 'Ċ', composite{"C", markDot}},
	{"cdotaccent", 'ċ', composite{"c", markDot}},
	{"Ccaron", 'Č', composite{"C", markCaron}},
	{"ccaron", 'č', composite{"c", markCaron}},
	{"Dcaron", 'Ď', composite{"D", markCaron}},
	{"dcaron", 'ď', composite{"d", markApostrophe}},
	{"Emacron", 'Ē', composite{"E", markMacron}},
	{"emacron", 'ē', composite{"e", markMacron}},
	{"Ebreve", 'Ĕ', composite{"E", markBreve}},
	{"ebreve", 'ĕ', composite{"e", markBreve}},
	{"Edotaccent", 'Ė', composite{"E", markDot}},
	{"edotaccent", 'ė', composite{"e", markDot}},
	{"Eogonek", 'Ę', composite{"E", markOgonek}},
	{"eogonek", 'ę', composite{"e", markOgonek}},
	{"Ecaron", 'Ě', composite{"E", markCaron}},
	{"ecaron", 'ě', composite{"e", markCaron}},
	{"Gcircumflex", 'Ĝ', composite{"G", markCircumflex}},
	{"gcircumflex", 'ĝ', composite{"g", markCircumflex}},
	{"Gbreve", 'Ğ', composite{"G", markBreve}},
	{"gbreve", 'ğ', composite{"g", markBreve}},
	{"Gdotaccent", 'Ġ', composite{"G", markDot}},
	{"gdotaccent", 'ġ', composite{"g", markDot}},
	{"Gcommaaccent", 'Ģ', composite{"G", markCommaBelow}},
	{"gcommaaccent", 'ģ', composite{"g", markCommaAbove}},
	{"Hcircumflex", 'Ĥ', composite{"H", markCircumflex}},
	{"hcircumflex", 'ĥ', composite{"h", markCircumflex}},
	{"Itilde", 'Ĩ', composite{"I", markTilde}},
	{"itilde", 'ĩ', composite{"dotlessi", markTilde}},
	{"Imacron", 'Ī', composite{"I", markMacron}},
	{"imacron", 'ī', composite{"dotlessi", markMacron}},
	{"Ibreve", 'Ĭ', composite{"I", markBreve}},
	{"ibreve", 'ĭ', composite{"dotlessi", markBreve}},
	{"Iogonek", 'Į', composite{"I", markOgonek}},
	{"iogonek", 'į', composite{"i", markOgonek}},
	{"Idotaccent", 'İ', composite{"I", markDot}},
	{"Jcircumflex", 'Ĵ', composite{"J", markCircumflex}},
	{"jcircumflex", 'ĵ', composite{"uni0237", markCircumflex}},
	{"Kcommaaccent", 'Ķ', composite{"K", markCommaBelow}},
	{"kcommaaccent", 'ķ', composite{"k", markCommaBelow}},
	{"Lacute", 'Ĺ', composite{"L", markAcute}},
	{"lacute", 'ĺ', composite{"l", markAcute}},
	{"Lcommaaccent", 'Ļ', composite{"L", markCommaBelow}},
	{"lcommaaccent", 'ļ', composite{"l", markCommaBelow}},
	{"Lcaron", 'Ľ', composite{"L", markApostrophe}},
	{"lcaron", 'ľ', composite{"l", markApostrophe}},
	{"Nacute", 'Ń', composite{"N", markAcute}},
	{"nacute", 'ń', composite{"n", markAcute}},
	{"Ncommaaccent", 'Ņ', composite{"N", markCommaBelow}},
	{"ncommaaccent", 'ņ', composite{"n", markCommaBelow}},
	{"Ncaron", 'Ň', composite{"N", markCaron}},
	{"ncaron", 'ň', composite{"n", markCaron}},
	{"Omacron", 'Ō', composite{"O", markMacron}},
	{"omacron", 'ō', composite{"o", markMacron}},
	{"Obreve", 'Ŏ', composite{"O", markBreve}},
	{"obreve", 'ŏ', composite{"o", markBreve}},
	{"Ohungarumlaut", 'Ő', composite{"O", markDblAcute}},
	{"ohungarumlaut", 'ő', composite{"o", markDblAcute}},
	{"Racute", 'Ŕ', composite{"R", markAcute}},
	{"racute", 'ŕ', composite{"r", markAcute}},
	{"Rcommaaccent", 'Ŗ', composite{"R", markCommaBelow}},
	{"rcommaaccent", 'ŗ', composite{"r", markCommaBelow}},
	{"Rcaron", 'Ř', composite{"R", markCaron}},
	{"rcaron", 'ř', composite{"r", markCaron}},
	{"Sacute", 'Ś', composite{"S", markAcute}},
	{"sacute", 'ś', composite{"s", markAcute}},
	{"Scircumflex", 'Ŝ', composite{"S", markCircumflex}},
	{"scircumflex", 'ŝ', composite{"s", markCircumflex}},
	{"Scedilla", 'Ş', composite{"S", markCedilla}},
	{"scedilla", 'ş', composite{"s", markCedilla}},
	{"Scaron", 'Š', composite{"S", markCaron}},
	{"scaron", 'š', composite{"s", markCaron}},
	{"uni0162", 'Ţ', composite{"T", markCedilla}},
	{"uni0163", 'ţ', composite{"t", markCedilla}},
	{"Tcaron", 'Ť', composite{"T", markCaron}},
	{"tcaron", 'ť', composite{"t", markApostrophe}},
	{"Utilde", 'Ũ', composite{"U", markTilde}},
	{"utilde", 'ũ', composite{"u", markTilde}},
	{"Umacron", 'Ū', composite{"U", markMacron}},
	{"umacron", 'ū', composite{"u", markMacron}},
	{"Ubreve", 'Ŭ', composite{"U", markBreve}},
	{"ubreve", 'ŭ', composite{"u", markBreve}},
	{"Uring", 'Ů', composite{"U", markRing}},
	{"uring", 'ů', composite{"u", markRing}},
	{"Uhungarumlaut", 'Ű', composite{"U", markDblAcute}},
	{"uhungarumlaut", 'ű', composite{"u", markDblAcute}},
	{"Uogonek", 'Ų', composite{"U", markOgonek}},
	{"uogonek", 'ų', composite{"u", markOgonek}},
	{"Wcircumflex", 'Ŵ', composite{"W", markCircumflex}},
	{"wcircumflex", 'ŵ', composite{"w", markCircumflex}},
	{"Ycircumflex", 'Ŷ', composite{"Y", markCircumflex}},
	{"ycircumflex", 'ŷ', composite{"y", markCircumflex}},
	{"Ydieresis", 'Ÿ', composite{"Y", markDieresis}},
	{"Zacute", 'Ź', composite{"Z", markAcute}},
	{"zacute", 'ź', composite{"z", markAcute}},
	{"Zdotaccent", 'Ż', composite{"Z", markDot}},
	{"zdotaccent", 'ż', composite{"z", markDot}},
	{"Zcaron", 'Ž', composite{"Z", markCaron}},
	{"zcaron", 'ž', composite{"z", markCaron}},

	// Latin Extended-B: the comma below of Romanian.
	{"Scommaaccent", 'Ș', composite{"S", markCommaBelow}},
	{"scommaaccent", 'ș', composite{"s", markCommaBelow}},
	{"uni021A", 'Ț', composite{"T", markCommaBelow}},
	{"uni021B", 'ț', composite{"t", markCommaBelow}},

	// Latin Extended Additional: the dot below.
	{"uni1E0C", 'Ḍ', composite{"D", markDotBelow}},
	{"uni1E0D", 'ḍ', composite{"d", markDotBelow}},
}

// spacingList lists the spacing accents, which are the marks standing alone,
// and combiningList the combining marks.
var spacingList = []struct {
	name string
	r    rune
	mark markKind
}{
	{"acute", '´', markAcute},
	{"dieresis", '¨', markDieresis},
	{"macron", '¯', markMacron},
	{"cedilla", '¸', markCedilla},
	{"circumflex", 'ˆ', markCircumflex},
	{"caron", 'ˇ', markCaron},
	{"breve", '˘', markBreve},
	{"dotaccent", '˙', markDot},
	{"ring", '˚', markRing},
	{"ogonek", '˛', markOgonek},
	{"tilde", '˜', markTilde},
	{"hungarumlaut", '˝', markDblAcute},
}

var combiningList = []struct {
	name string
	r    rune
	mark markKind
}{
	{"gravecomb", '̀', markGrave},
	{"acutecomb", '́', markAcute},
	{"uni0302", '̂', markCircumflex},
	{"tildecomb", '̃', markTilde},
	{"uni0304", '̄', markMacron},
	{"uni0306", '̆', markBreve},
	{"uni0307", '̇', markDot},
	{"uni0308", '̈', markDieresis},
	{"uni030A", '̊', markRing},
	{"uni030B", '̋', markDblAcute},
	{"uni030C", '̌', markCaron},
	{"dotbelowcomb", '̣', markDotBelow},
	{"uni0326", '̦', markCommaBelow},
	{"uni0327", '̧', markCedilla},
	{"uni0328", '̨', markOgonek},
}

// composites holds the composed letters by name.
var composites = map[string]*composite{}

// like names, for the glyphs which take the kerning of another glyph, that
// glyph: the base of a composed letter (or the base of the base), i and j
// for the dotless letters, i for its form of the feature locl, S for the two
// letters S, and the hyphen and the dashes for the glyphs which are drawn as
// they are.
func like(name string) string {
	switch name {
	case "dotlessi", "i.loclTRK":
		return "i"
	case "uni0237":
		return "j"
	case "S_S":
		return "S"
	case "uni2010", "uni2011":
		return "hyphen"
	case "figuredash":
		return "endash"
	case "uni2015", "twelveudash", "threequartersemdash":
		return "emdash"
	}
	if c := composites[name]; c != nil {
		if l := like(c.base); l != "" {
			return l
		}
		return c.base
	}
	return ""
}

func init() {
	table = append(table,
		entry{"dotlessi", 'ı', sideNarrow, sideNarrow, dotlessi},
		entry{"uni0237", 'ȷ', sideOverhang, sideNarrow, dotlessj},
	)
	for _, s := range spacingList {
		k := s.mark
		table = append(table, entry{s.name, s.r, sideOpen, sideOpen, func(p design.Params) (outline.Path, float64) {
			return spacingAccent(p, k)
		}})
	}
	for _, s := range combiningList {
		k := s.mark
		table = append(table, entry{s.name, s.r, sideMark, sideMark, func(p design.Params) (outline.Path, float64) {
			return combiningMark(p, k), 0
		}})
	}
	for _, c := range composedList {
		b, ok := lookup(c.base)
		if !ok {
			panic("glyphs: no base " + c.base + " for " + c.name)
		}
		comp := c.composite
		composites[c.name] = &comp
		table = append(table, entry{c.name, c.r, b.left, b.right, func(p design.Params) (outline.Path, float64) {
			bp, adv, _ := Outline(comp.base, p)
			dx, dy, f := comp.place(p, bp)
			return comp.assemble(bp, markOutline(p, comp.mark, f), dx, dy), adv
		}})
	}
}

// dotlessi draws the dotless i: the i without its dot.
func dotlessi(p design.Params) (outline.Path, float64) {
	path, adv := i(p)
	return path[:len(path)-1], adv
}

// dotlessj draws the dotless j: the j without its dot.
func dotlessj(p design.Params) (outline.Path, float64) {
	path, adv := j(p)
	return path[:len(path)-1], adv
}

// markHeight returns the height of the anchor of a mark above a lowercase
// letter without ascender, at which the spacing and the combining marks stand,
// and the height of the anchor of a mark below a letter.
func markHeight(p design.Params, k markKind) float64 {
	switch {
	case k.below():
		return 0
	case k == markApostrophe:
		return p.Ascender
	}
	return p.XHeight + p.Acc.Gap
}

// spacingAccent draws the mark standing alone: the lowercase form above the
// x-height, or below the baseline, with Acc.MarkSide on each side of its ink.
func spacingAccent(p design.Params, k markKind) (outline.Path, float64) {
	m := markOutline(p, k, formLower)
	x0, _, x1, _ := m.BBox()
	dx := math.Round(p.Acc.MarkSide - x0)
	return shiftInk(m, dx, markHeight(p, k)), math.Ceil(x1 + dx + p.Acc.MarkSide - 1e-6)
}

// combiningMark draws the combining mark: the lowercase form, which has no
// advance and stands over the middle of the advance of a preceding n.
func combiningMark(p design.Params, k markKind) outline.Path {
	return shiftInk(markOutline(p, k, formLower), combiningX(p), markHeight(p, k))
}

// combiningX returns the abscissa of the anchor of a combining mark: half the
// advance of n to the left of its own position.
func combiningX(p design.Params) float64 {
	return -math.Round(design.NAdvance(p) / 2)
}

// The kinds of the base letters, for the anchors.
var (
	ascenderBases = []string{"b", "d", "h", "k", "l"}
	stemTopBases  = []string{"i", "dotlessi", "j", "uni0237", "l", "h"}
)

// isCapital says whether the glyph is a capital letter.
func isCapital(name string) bool { return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' }

// place returns the offset by which the mark is moved from its frame into the
// letter, in whole units, and the form of the mark.
func (c *composite) place(p design.Params, base outline.Path) (dx, dy float64, f markForm) {
	a := p.Acc
	switch c.mark {
	case markApostrophe:
		x, top := sideAnchor(p, c.base, base)
		if c.base == "t" {
			// The mark stands clear above the bar of t.
			h := a.ApoScale * (commaTop(p) + p.Punct.CommaDepth)
			top = math.Max(top, p.XHeight+a.Gap/2+h)
		}
		// The mark keeps clear of a head serif on the right of the stem.
		_, serif := inkSpan(base, top-70, top)
		x = math.Max(x+a.ApoGap, serif+design.ClearanceFloor(p)+4)
		return math.Round(x), math.Round(top), formCap
	case markCedilla, markOgonek:
		x := bottomX(p, c.base, base)
		switch {
		case c.mark == markOgonek:
			x = ogonekX(p, c.base, base)
		case c.base == "t":
			// The neck of the cedilla leaves the lowest point of the curve of
			// t into its tail.
			x = lowestX(base)
		}
		return math.Round(x), math.Round(lowestInk(base, x)), formLower
	case markCommaBelow, markDotBelow:
		return math.Round(bottomX(p, c.base, base)), 0, formLower
	}
	x := topX(p, c.base, base)
	top, gap := p.XHeight, a.Gap
	switch {
	case isCapital(c.base):
		top, gap = p.CapHeight, a.GapCap
	case slices.Contains(ascenderBases, c.base):
		top, gap = p.Ascender, a.GapCap
	}
	return math.Round(x), math.Round(top + gap), c.form()
}

// form returns the form of the mark of the letter.
func (c *composite) form() markForm {
	switch {
	case c.mark == markApostrophe:
		return formCap
	case c.mark.below():
		return formLower
	case isCapital(c.base), slices.Contains(ascenderBases, c.base):
		return formCap
	case c.base == "dotlessi" && c.mark == markDieresis:
		return formNarrow
	}
	return formLower
}

// joined says whether the mark is joined to the base: the cedilla and the
// ogonek.
func (c *composite) joined() bool { return c.mark.joinable() }

// assemble returns the unrounded letter: the base with the mark moved by (dx,
// dy).  A joined mark is united with the base.
func (c *composite) assemble(base, mark outline.Path, dx, dy float64) outline.Path {
	m := shiftInk(mark, dx, dy)
	if !c.joined() {
		out := make(outline.Path, 0, len(base)+len(m))
		out = append(out, base...)
		return append(out, m...)
	}
	parts := append(pathChains(base), pathChains(m)...)
	return chainsPath(union(parts...)...)
}

// rounded returns the rounded letter.  A letter whose mark is not joined is
// the rounded base with the rounded mark, moved by (dx, dy), which are whole
// units; a letter with a joined mark is the union of the unrounded base and
// mark (raw and rawMark), rounded.
func (c *composite) rounded(base, mark, raw, rawMark outline.Path, dx, dy float64) outline.Path {
	if c.joined() {
		return c.assemble(raw, rawMark, dx, dy).Round()
	}
	return c.assemble(base, mark, dx, dy)
}

// inkSpan returns the least and the greatest abscissa of the outline between
// the heights y0 and y1, from the flattened contours and their crossings of
// the two heights.
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

// lowestX returns the abscissa of the lowest point of the outline.
func lowestX(path outline.Path) float64 {
	x, low := 0.0, math.Inf(1)
	for _, ct := range path {
		for _, q := range ct.Flatten() {
			if q.Y < low {
				x, low = q.X, q.Y
			}
		}
	}
	return x
}

// lowestInk returns the lowest height at which the vertical through x meets
// the outline.
func lowestInk(path outline.Path, x float64) float64 {
	low := math.Inf(1)
	for _, ct := range path {
		pts := ct.Flatten()
		n := len(pts)
		for i := range n {
			a, b := pts[i], pts[(i+1)%n]
			if (a.X < x) != (b.X < x) {
				y := a.Y + (x-a.X)/(b.X-a.X)*(b.Y-a.Y)
				low = math.Min(low, y)
			}
		}
	}
	return low
}

// topX returns the abscissa of the top anchor of the base.
func topX(p design.Params, name string, base outline.Path) float64 {
	h := p.XHeight
	if isCapital(name) {
		h = p.CapHeight
	}
	switch {
	case slices.Contains(stemTopBases, name):
		lo, hi := stemBelowTop(p, name, base)
		return (lo+hi)/2 + design.DotOffset(p)
	case name == "J" || name == "L" || name == "I":
		lo, hi := inkSpan(base, h-150, h-100)
		return (lo + hi) / 2
	case name == "a":
		// The mean of the middle of the bowl and the stem and the middle of
		// the ink with the tail.
		lo, hi := inkSpan(base, 0.25*h, 0.55*h)
		lo2, hi2 := inkSpan(base, 0, h)
		return (lo + hi + lo2 + hi2) / 4
	case name == "g":
		lo, hi := inkSpan(base, 0.45*h, 0.8*h)
		return (lo + hi) / 2
	}
	lo, hi := inkSpan(base, 0, h)
	return (lo + hi) / 2
}

// stemBelowTop returns the edges of the stem of a letter whose top is a stem
// with a head serif, at a height below the serif.
func stemBelowTop(p design.Params, name string, base outline.Path) (lo, hi float64) {
	top := p.XHeight
	if slices.Contains(ascenderBases, name) {
		top = p.Ascender
	}
	return inkSpan(base, top-110, top-80)
}

// bottomX returns the abscissa of the bottom anchor of the base.
func bottomX(p design.Params, name string, base outline.Path) float64 {
	switch name {
	case "t", "l", "r", "T", "L":
		h := p.XHeight
		if isCapital(name) {
			h = p.CapHeight
		}
		lo, hi := inkSpan(base, 0.5*h, 0.6*h)
		if name == "t" || name == "l" || name == "r" {
			lo, hi = stemAt(base, 0.5*h)
		}
		return (lo + hi) / 2
	}
	h := p.XHeight
	if isCapital(name) {
		h = p.CapHeight
	}
	if name == "D" {
		// The foot of D: its ink in the lowest fifth of its height.
		lo, hi := inkSpan(base, 0, 0.2*h)
		return (lo + hi) / 2
	}
	lo, hi := inkSpan(base, 0, h)
	return (lo + hi) / 2
}

// stemAt returns the edges of the leftmost stroke of the outline at the
// height y.
func stemAt(path outline.Path, y float64) (lo, hi float64) {
	var xs []float64
	for _, ct := range path {
		pts := ct.Flatten()
		n := len(pts)
		for i := range n {
			a, b := pts[i], pts[(i+1)%n]
			if (a.Y < y) != (b.Y < y) {
				xs = append(xs, a.X+(y-a.Y)/(b.Y-a.Y)*(b.X-a.X))
			}
		}
	}
	slices.Sort(xs)
	if len(xs) < 2 {
		return math.NaN(), math.NaN()
	}
	return xs[0], xs[1]
}

// ogonekX returns the abscissa at which the ogonek leaves the base: a quarter
// of the stem right of the axis of the right stem of a and u, the stem of i
// and I, OgoReach left of the right end of the foot of E and A, and
// right of the lowest point of e (by 0.6 of OgoReach) and of U (by three
// quarters of it).
func ogonekX(p design.Params, name string, base outline.Path) float64 {
	h := p.XHeight
	if isCapital(name) {
		h = p.CapHeight
	}
	switch name {
	case "a", "u":
		// A quarter of the stem right of the axis of the right stem.
		_, hi := inkSpan(base, 0.3*h, 0.5*h)
		return hi - p.Stem/4
	case "i", "I":
		lo, hi := inkSpan(base, 0.3*h, 0.5*h)
		return (lo + hi) / 2
	case "E", "A":
		_, hi := inkSpan(base, 0, 0.02*h)
		return hi - p.Acc.OgoReach
	case "U":
		return lowestX(base) + 0.75*p.Acc.OgoReach
	}
	return lowestX(base) + 0.6*p.Acc.OgoReach
}

// sideAnchor returns the right edge of the stem of d, l, t or L and the height
// of its top.
func sideAnchor(p design.Params, name string, base outline.Path) (x, top float64) {
	_, _, _, top = base.BBox()
	switch name {
	case "L":
		top = p.CapHeight
		_, x = stemAt(base, top-120)
	case "t":
		_, x = stemAt(base, top-40)
	default:
		top = p.Ascender
		_, x = inkSpan(base, top-110, top-80)
	}
	return x, top
}

// A markKey names the form of a mark.
type markKey struct {
	kind markKind
	form markForm
}

// finishSet completes the glyphs of Set: it composes the letters from the
// rounded bases and marks, with raws holding the unrounded outlines of the
// other glyphs, and sets the fields Base, Like, Marks and Anchors.
func finishSet(p design.Params, gs []Glyph, raws []outline.Path) {
	index := make(map[string]int, len(table))
	for i, e := range table {
		index[e.name] = i
	}
	marks := map[markKey][2]outline.Path{} // unrounded and rounded
	for i, e := range table {
		gs[i].Like = like(e.name)
		c := composites[e.name]
		if c == nil {
			gs[i].Anchors = baseAnchors(p, e.name, raws[i])
			continue
		}
		b := index[c.base]
		dx, dy, f := c.place(p, raws[b])
		key := markKey{c.mark, f}
		m, ok := marks[key]
		if !ok {
			raw := markOutline(p, c.mark, f)
			m = [2]outline.Path{raw, raw.Round()}
			marks[key] = m
		}
		gs[i] = Glyph{
			Name: e.name, Rune: e.r, Advance: gs[b].Advance,
			Path: c.rounded(gs[b].Path, m[1], raws[b], m[0], dx, dy),
			Base: c.base, Like: gs[i].Like, Marks: shiftInk(m[1], dx, dy),
		}
	}
	for i, e := range table {
		if k, ok := combiningKind(e.name); ok {
			gs[i].Anchors = []Anchor{markAnchor(p, k)}
		}
		if _, ok := capitalCombiningKind(e.name); ok {
			gs[i].Anchors = []Anchor{capitalMarkAnchor(p)}
		}
		gs[i].Height = inklessHeight(p, e.name)
	}
}

// roundedFromScratch returns the rounded letter, built from the base and the
// mark.
func (c *composite) roundedFromScratch(p design.Params) outline.Path {
	bp, _, _ := Outline(c.base, p)
	dx, dy, f := c.place(p, bp)
	m := markOutline(p, c.mark, f)
	return c.rounded(bp.Round(), m.Round(), bp, m, dx, dy)
}

// combiningKind returns the mark of a combining mark.
func combiningKind(name string) (markKind, bool) {
	for _, s := range combiningList {
		if s.name == name {
			return s.mark, true
		}
	}
	return 0, false
}

// anchorClass returns the anchor of the base to which a mark is attached.
func anchorClass(k markKind) string {
	switch k {
	case markCedilla, markCommaBelow, markDotBelow:
		return "bottom"
	case markOgonek:
		return "ogonek"
	}
	return "top"
}

// markAnchor returns the anchor of a combining mark: the anchor of its frame
// in the glyph, at the x-height for the marks above, so that the base anchor
// at the top of a letter lies on the line on which the mark stands.
func markAnchor(p design.Params, k markKind) Anchor {
	y := 0.0
	if !k.below() {
		y = p.XHeight
	}
	return Anchor{Class: anchorClass(k), X: combiningX(p), Y: y}
}

// baseAnchors returns the anchors of a letter for the combining marks: the
// top, bottom and ogonek anchors of the letters a to z, A to Z, the dotless
// i and j and the form of i of the feature locl, from the unrounded outline.  A combining mark is set on
// the top anchor with its lowest ink Acc.Gap above it, so that the top anchor
// of a capital stands Acc.GapCap - Acc.Gap above the cap height.
func baseAnchors(p design.Params, name string, path outline.Path) []Anchor {
	if name == "i.loclTRK" {
		name = "i"
	}
	letter := len(name) == 1 && (isCapital(name) || name[0] >= 'a' && name[0] <= 'z')
	if !letter && name != "dotlessi" && name != "uni0237" {
		return nil
	}
	top := p.XHeight
	switch {
	case isCapital(name):
		top = p.CapHeight + p.Acc.GapCap - p.Acc.Gap
	case slices.Contains(ascenderBases, name):
		top = p.Ascender + p.Acc.GapCap - p.Acc.Gap
	}
	return []Anchor{
		{Class: "top", X: math.Round(topX(p, name, path)), Y: math.Round(top)},
		{Class: "bottom", X: math.Round(bottomX(p, name, path)), Y: 0},
		{Class: "ogonek", X: math.Round(ogonekX(p, name, path)), Y: 0},
	}
}
