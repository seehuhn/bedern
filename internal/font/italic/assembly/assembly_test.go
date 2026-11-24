// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"seehuhn.de/go/sfnt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/kern"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/upper"
)

// The tests which build the font cost about as much as a font and run in the
// full tier only.

var table = sync.OnceValues(func() ([]glyphs.Glyph, error) { return Table(Default()) })

var kerned = sync.OnceValue(func() *kern.Table {
	gs, err := table()
	if err != nil {
		return nil
	}
	return KernTable(gs)
})

func fullTable(t *testing.T) []glyphs.Glyph {
	t.Helper()
	if testing.Short() {
		t.Skip("builds the whole italic")
	}
	gs, err := table()
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

// TestFitTableNamesTheGlyphs checks, without building, that the fit table
// names every letter of the lowercase and every glyph of the upper package
// and the dotless i and j, and nothing else.
func TestFitTableNamesTheGlyphs(t *testing.T) {
	var want []string
	for _, n := range italic.Names() {
		want = append(want, italic.GlyphName(n))
	}
	want = append(append(want, upper.Names()...), ".notdef", "dotlessi", "uni0237", "uni00AD", "ordfeminine", "ordmasculine")
	for _, b := range barred {
		want = append(want, b.name)
	}
	for _, s := range uprightSigns {
		if s.r != '\u00A0' && s.r < '\u2000' {
			want = append(want, s.name)
		}
	}
	for _, n := range want {
		if _, ok := fitTable[n]; !ok {
			t.Errorf("the fit table has no sides for %s", n)
		}
	}
	for n := range fitTable {
		if !slices.Contains(want, n) {
			t.Errorf("the fit table has sides for %s, which is not a glyph", n)
		}
	}
}

// TestFitTableCoversEveryGlyph checks that every glyph with ink which is not
// composed has its sides in the fit table and stands with them in the font.
func TestFitTableCoversEveryGlyph(t *testing.T) {
	for _, g := range fullTable(t) {
		if g.Base != "" || len(g.Parts) > 0 || len(g.Path) == 0 || len(g.Anchors) == 1 && g.Advance == 0 {
			continue
		}
		if g.Name == "S_S" || g.Name == "i.loclTRK" {
			continue // made of the fitted S and i
		}
		if kind(g) == "mark" && g.Name != "asciicircum" && g.Name != "grave" {
			continue
		}
		if unfitted(g.Name) {
			continue // marks of TS1, which keep the roman's place in their advance
		}
		want, ok := fitTable[g.Name]
		if !ok {
			t.Errorf("%s: not in the fit table", g.Name)
			continue
		}
		s, _ := sidesOf(g)
		if math.Abs(s.Left-want.Left) > 1 || math.Abs(s.Right-want.Right) > 1 {
			t.Errorf("%s: sides %.1f, %.1f, the table %v, %v", g.Name, s.Left, s.Right, want.Left, want.Right)
		}
	}
}

// TestFontIsMarkedItalic writes the font, reads it back and checks the marks
// of an italic: the italic bits, no regular bit, the italic angle of the
// slant, and the trial in the family name.
func TestFontIsMarkedItalic(t *testing.T) {
	gs := fullTable(t)
	p := Default()
	kt := kerned()
	if kt == nil {
		t.Fatal("no kerns")
	}
	f, err := FontOf(p, gs, kt, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	fname := filepath.Join(t.TempDir(), File)
	if err := os.WriteFile(fname, buf.Bytes(), 0o666); err != nil {
		t.Fatal(err)
	}
	g, err := sfnt.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	if rise, run := caretSlope(buf.Bytes()); rise != 4 || run != 1 {
		t.Errorf("caret slope %d/%d, want 4/1", rise, run)
	}
	if !g.IsItalic || g.IsRegular || g.IsBold {
		t.Errorf("italic %v, regular %v, bold %v", g.IsItalic, g.IsRegular, g.IsBold)
	}
	want := -math.Atan(0.25) * 180 / math.Pi
	if math.Abs(g.ItalicAngle-want) > 0.001 {
		t.Errorf("italic angle %.4f, want %.4f", g.ItalicAngle, want)
	}
	if !bytes.Contains([]byte(g.FamilyName), []byte("Trial")) {
		t.Errorf("family %q is not named as a trial", g.FamilyName)
	}
}

// caretSlope returns the caret slope of the hhea table of the font file.
func caretSlope(data []byte) (rise, run int16) {
	n := int(binary.BigEndian.Uint16(data[4:]))
	for i := range n {
		rec := data[12+16*i:]
		if string(rec[:4]) == "hhea" {
			o := int(binary.BigEndian.Uint32(rec[8:]))
			return int16(binary.BigEndian.Uint16(data[o+18:])), int16(binary.BigEndian.Uint16(data[o+20:]))
		}
	}
	return 0, 0
}

// inside says whether the point lies in the ink of the path, by the winding
// number of its flattened contours.
func inside(path outline.Path, q outline.Point) bool {
	w := 0
	for _, c := range path {
		pts := c.Flatten()
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			if (a.Y <= q.Y) != (b.Y <= q.Y) {
				x := a.X + (q.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
				if x > q.X {
					if b.Y > a.Y {
						w++
					} else {
						w--
					}
				}
			}
		}
	}
	return w != 0
}

// nearEdge says whether the point lies within d of an edge of the path.
func nearEdge(path outline.Path, q outline.Point, d float64) bool {
	for _, c := range path {
		pts := c.Flatten()
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			dx, dy := b.X-a.X, b.Y-a.Y
			l := dx*dx + dy*dy
			s := 0.0
			if l > 0 {
				s = math.Max(0, math.Min(1, ((q.X-a.X)*dx+(q.Y-a.Y)*dy)/l))
			}
			if math.Hypot(q.X-a.X-s*dx, q.Y-a.Y-s*dy) <= d {
				return true
			}
		}
	}
	return false
}

// TestComposedKeepTheBase checks that every composed letter contains its
// base: every point of the base's outline lies in the ink of the letter or
// within 1.5 of its edge, and the letter has the base's advance.
func TestComposedKeepTheBase(t *testing.T) {
	gs := fullTable(t)
	byName := map[string]glyphs.Glyph{}
	for _, g := range gs {
		byName[g.Name] = g
	}
	n := 0
	for _, g := range gs {
		if g.Base == "" {
			continue
		}
		n++
		b := byName[g.Base]
		if g.Advance != b.Advance {
			t.Errorf("%s: advance %v, base %s %v", g.Name, g.Advance, b.Name, b.Advance)
		}
		bad := 0
		for _, c := range b.Path {
			for _, q := range c.Flatten() {
				if !inside(g.Path, q) && !nearEdge(g.Path, q, 1.5) {
					bad++
				}
			}
		}
		if bad > 0 {
			t.Errorf("%s: %d points of the base %s lie outside the letter", g.Name, bad, b.Name)
		}
	}
	if n != len(glyphs.Compositions()) {
		t.Errorf("%d composed letters, the roman has %d", n, len(glyphs.Compositions()))
	}
}

// TestMadeOfOthers checks the glyphs which the italic makes of its own or takes
// from the roman: ḍ and Ḍ carry the dot of the italic (a circle of the
// diameter of the dot of i, which is not the roman's dot); the ordinal
// indicators stand over an underline as wide as the letter, with 30 of white
// on either side; the two letters S are two S in the advance of S and take
// the kerning of S; the form of i of the feature locl is the i; the ten
// spaces of fixed width have the advances of the roman, the punctuation space
// that of the italic full stop, and the compound word mark has none.
func TestMadeOfOthers(t *testing.T) {
	p := Default()
	gs := fullTable(t)
	by := map[string]glyphs.Glyph{}
	for _, g := range gs {
		by[g.Name] = g
	}
	for _, n := range []string{"uni1E0D", "uni1E0C"} {
		x0, y0, x1, y1 := by[n].Marks.BBox()
		if d := p.Lower.Dot; math.Abs(x1-x0-d) > 1 || math.Abs(y1-y0-d) > 1 {
			t.Errorf("%s: the dot below is %.0f by %.0f, want a circle of %v", n, x1-x0, y1-y0, d)
		}
		if y1 > 0 {
			t.Errorf("%s: the dot below reaches %.0f, above the baseline", n, y1)
		}
	}
	for _, n := range []string{"ordfeminine", "ordmasculine"} {
		g := by[n]
		x0, y0, x1, _ := g.Path.BBox()
		if math.Abs(x0-ordSide) > 1 || math.Abs(g.Advance-x1-ordSide) > 1 {
			t.Errorf("%s: sides %.0f and %.0f, want %v", n, x0, g.Advance-x1, ordSide)
		}
		if math.Abs(y0-ordBottom) > 1 {
			t.Errorf("%s: the underline starts at %.0f, want %v", n, y0, ordBottom)
		}
	}
	s, ss := by["S"], by["S_S"]
	if ss.Advance != 2*s.Advance || ss.Like != "S" || len(ss.Path) != 2*len(s.Path) || ss.Rune != 0 {
		t.Errorf("S_S: advance %v (S %v), like %q, %d contours (S %d)", ss.Advance, s.Advance, ss.Like, len(ss.Path), len(s.Path))
	}
	i, il := by["i"], by["i.loclTRK"]
	if il.Advance != i.Advance || len(il.Path) != len(i.Path) || il.Like != "i" || len(il.Anchors) != len(i.Anchors) || il.Rune != 0 {
		t.Errorf("i.loclTRK is not the i: advance %v and %v, like %q", il.Advance, i.Advance, il.Like)
	}
	for _, c := range []struct {
		name string
		adv  float64
	}{
		{"uni2000", 500}, {"uni2001", 1000}, {"uni2002", 500}, {"uni2003", 1000}, {"uni2004", 1000.0 / 3},
		{"uni2005", 250}, {"uni2006", 1000.0 / 6}, {"uni2007", by["zero"].Advance}, {"uni2008", by["period"].Advance},
		{"uni205F", 1000 * 4.0 / 18}, {"uni200C", 0},
	} {
		g, ok := by[c.name]
		if !ok || len(g.Path) != 0 || math.Abs(g.Advance-c.adv) > 0.5 {
			t.Errorf("%s: has %v, advance %v (ink %d contours), want %.1f", c.name, ok, g.Advance, len(g.Path), c.adv)
		}
	}
	for _, c := range []struct {
		name string
		like string
	}{{"uni2010", "hyphen"}, {"uni2011", "hyphen"}, {"figuredash", "endash"}, {"uni2015", "emdash"}, {"uni2034", ""}, {"uni2031", ""}, {"perthousandzero", ""}, {"uni2423", ""}} {
		g, ok := by[c.name]
		if !ok || len(g.Path) == 0 || g.Like != c.like {
			t.Errorf("%s: has %v, %d contours, like %q, want like %q", c.name, ok, len(g.Path), g.Like, c.like)
		}
	}
	// The hyphens are the hyphen, and the dashes and the primes lean as the
	// italic's other signs do.
	if h, h2 := by["hyphen"], by["uni2010"]; h.Advance != h2.Advance || len(h.Path) != len(h2.Path) {
		t.Errorf("the hyphen U+2010 is not the hyphen: advances %v and %v", h2.Advance, h.Advance)
	}
}

// TestMarksClearTheBase checks that a mark which is not joined to its base
// keeps the clearance floor from it.
func TestMarksClearTheBase(t *testing.T) {
	gs := fullTable(t)
	byName := map[string]glyphs.Glyph{}
	for _, g := range gs {
		byName[g.Name] = g
	}
	floor := design.ClearanceFloor(design.Regular())
	joined := map[string]bool{}
	for _, c := range glyphs.Compositions() {
		joined[c.Name] = c.Joined
	}
	for _, g := range gs {
		if g.Base == "" || joined[g.Name] {
			continue
		}
		d := math.Inf(1)
		for _, c := range g.Marks {
			for _, q := range c.Flatten() {
				for _, bc := range byName[g.Base].Path {
					for _, r := range bc.Flatten() {
						d = math.Min(d, math.Hypot(q.X-r.X, q.Y-r.Y))
					}
				}
			}
		}
		if d < floor {
			t.Errorf("%s: the mark comes %.1f near the base, floor %.1f", g.Name, d, floor)
		}
	}
}

// TestKernDirections checks that stems beside stems keep their side spaces on
// the slant, and that the pairs whose second letter can move into the white
// under an arm or a diagonal are drawn together.
func TestKernDirections(t *testing.T) {
	fullTable(t)
	tb := kerned()
	f := tb.Font
	idx := map[rune]int{}
	for _, g := range f.Glyphs {
		if g.Rune != 0 {
			idx[g.Rune] = g.Index
		}
	}
	kern := func(s string) int {
		r := []rune(s)
		return tb.Kern(idx[r[0]], idx[r[1]])
	}
	for _, s := range []string{"nn", "ll", "ln", "nl", "HH", "II", "HI", "IH", "nm", "un"} {
		if k := kern(s); k != 0 {
			t.Errorf("%s has the kern %d", s, k)
		}
	}
	for _, s := range []string{"To", "Ta", "AV", "LT", "LY", "Yo", "Va", "P."} {
		if k := kern(s); k >= 0 {
			t.Errorf("%s has the kern %d, but its second glyph can move into the white", s, k)
		}
	}
}

// TestKernKeepsTheFloor checks that no two letters or signs which are kerned,
// set with their kern, come closer than the floor of the kerning (the
// computation of the distance may be 1.5 too large; the test allows 2).
func TestKernKeepsTheFloor(t *testing.T) {
	fullTable(t)
	tb := kerned()
	if tb == nil || len(tb.Pairs) == 0 {
		t.Fatal("no kerns")
	}
	f := tb.Font
	floor := Kerning().Kern.Floor
	var kernable []*kern.Glyph
	for _, g := range f.Glyphs {
		if g.Kind != kern.Skip && g.Like < 0 && g.First < 0 {
			kernable = append(kernable, g)
		}
	}
	for _, a := range kernable {
		for _, b := range kernable {
			if a.Kind == kern.Punct && b.Kind == kern.Punct {
				continue
			}
			k := tb.Kern(a.Index, b.Index)
			if d := f.Closest(a, b, float64(k), 0, f.Rows()); d < floor-2 {
				t.Errorf("%s %s with kern %d: closest distance %.1f, floor %v", a.Name, b.Name, k, d, floor)
			}
		}
	}
}

// TestFitLimits checks the limits of the fit in the fitted font: the ball of
// r keeps 30 from the entry hooks of a following i, j, m, n, p, r, u, v, w
// and y below the top of the entries without kerning, and no two capitals
// need a positive kern, so that the fit and not the kerning keeps them
// apart.
func TestFitLimits(t *testing.T) {
	fullTable(t)
	tb := kerned()
	f := tb.Font
	byRune := map[rune]*kern.Glyph{}
	for _, g := range f.Glyphs {
		if g.Rune != 0 {
			byRune[g.Rune] = g
		}
	}
	r := byRune['r']
	for _, c := range "ijmnpruvwy" {
		// The nearest ink of a band of the kerning may lie up to half a
		// band above or below the band's middle, so that the profiles are
		// a little closer than the horizontal distance at one height.
		if d := f.Closest(r, byRune[c], 0, f.Row(0), f.Row(470)); d < 28 {
			t.Errorf("r%c: the ball comes %.1f near the entry, want at least 30", c, d)
		}
	}
	for a := 'A'; a <= 'Z'; a++ {
		for b := 'A'; b <= 'Z'; b++ {
			if k := tb.Kern(byRune[a].Index, byRune[b].Index); k > 0 {
				t.Errorf("%c%c has the positive kern %d", a, b, k)
			}
		}
	}
}
