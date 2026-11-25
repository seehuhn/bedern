// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic/assembly"
)

// The tests of the maths italic build every form of every letter, which
// takes several seconds of processor time, and belong to the full tier.

// forms returns the forms of the maths italic, or skips the test in the
// routine tier.
func forms(t *testing.T) []italicForm {
	t.Helper()
	if testing.Short() {
		t.Skip("the maths italic is built in the full tier")
	}
	fs, err := italicForms()
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

// textForms returns the forms at text size, by the letter of the italic,
// without ℏ.
func textForms(fs []italicForm) map[string]italicForm {
	out := map[string]italicForm{}
	for _, f := range fs {
		if f.Level == 0 && f.Rune != 0x210F {
			out[f.From] = f
		}
	}
	return out
}

// TestItalicCodePoints asks for a glyph at every code point of the maths
// italic, h at U+210E and not at the unassigned U+1D455, the dotless i and j,
// and ℏ, each with its two script forms but ℏ.
func TestItalicCodePoints(t *testing.T) {
	fs := forms(t)
	names := map[string]bool{}
	runes := map[rune]italicForm{}
	for _, f := range fs {
		names[f.Name] = true
		if f.Rune != 0 {
			runes[f.Rune] = f
		}
	}
	var want []rune
	for r := rune(0x1D434); r <= 0x1D467; r++ {
		if r != 0x1D455 {
			want = append(want, r)
		}
	}
	want = append(want, 0x210E, 0x1D6A4, 0x1D6A5, 0x210F)
	for _, r := range want {
		f, ok := runes[r]
		if !ok {
			t.Errorf("no glyph for %U", r)
			continue
		}
		if r == 0x210F {
			continue
		}
		for _, s := range f.Script {
			if !names[s] {
				t.Errorf("%U has no script form %q", r, s)
			}
		}
	}
	if _, ok := runes[0x1D455]; ok {
		t.Error("a glyph at the unassigned U+1D455")
	}
}

// TestItalicIsClean asks every form at text size for a clean build: no
// warning of the stroke tool and no finding of the italic profile of the
// smoothness check.
func TestItalicIsClean(t *testing.T) {
	for _, f := range forms(t) {
		if f.Level > 0 {
			continue
		}
		for _, w := range f.Warnings {
			t.Errorf("%s: %s", f.Name, w)
		}
		for _, d := range f.Findings {
			t.Errorf("%s: %s", f.Name, d)
		}
	}
}

// TestItalicOverhangs asks every form for ink at most maxOverhangLeft before
// its origin and maxOverhangRight beyond its advance.
func TestItalicOverhangs(t *testing.T) {
	for _, f := range forms(t) {
		x0, _, x1, _ := f.Path.BBox()
		if -x0 > maxOverhangLeft+0.5 || x1-f.Advance > maxOverhangRight+0.5 {
			t.Errorf("%s reaches %.0f before the origin and %.0f beyond the advance, want at most %d and %d", f.Name, -x0, x1-f.Advance, maxOverhangLeft, maxOverhangRight)
		}
	}
}

// TestItalicWidths asks that the lowercase at text size other than f and j,
// which grow by their overhangs, be 1.03 to 1.10 times as wide as the text
// italic on the mean, all the lowercase 0.90 to 1.14
// times, and that no lowercase letter lose white on either
// side against the text italic.
func TestItalicWidths(t *testing.T) {
	sum, n, all, na := 0.0, 0, 0.0, 0
	for from, f := range textForms(forms(t)) {
		if from >= "A" && from <= "Z" {
			continue
		}
		s, ok := assembly.FitSides(from)
		if !ok {
			t.Fatalf("no side spaces of %s", from)
		}
		g := sideGlyph(f)
		m, _ := assembly.SidesOf(g)
		if m.Left < s.Left+mathSide-0.5 || m.Right < s.Right+mathSide-0.5 {
			t.Errorf("%s has the sides %.0f and %.0f, the text italic %.0f and %.0f", from, m.Left, m.Right, s.Left, s.Right)
		}
		text := f.Advance - (m.Left - s.Left) - (m.Right - s.Right)
		all += f.Advance / text
		na++
		if from != "f" && from != "j" {
			sum += f.Advance / text
			n++
		}
	}
	t.Logf("the lowercase but f and j is %.3f times as wide as the text italic, all of it %.3f", sum/float64(n), all/float64(na))
	if mean := sum / float64(n); mean < 1.03 || mean > 1.10 {
		t.Errorf("the lowercase but f and j is %.3f times as wide as the text italic on the mean, want 1.03 to 1.10", mean)
	}
	if mean := all / float64(na); mean < 0.90 || mean > 1.14 {
		t.Errorf("the lowercase is %.3f times as wide as the text italic on the mean, want 0.90 to 1.14", mean)
	}
}

// sideGlyph returns the form as a glyph whose name gives the band in which
// the fit of the italic measures the sides.
func sideGlyph(f italicForm) glyphs.Glyph {
	g := glyphs.Glyph{Name: f.From, Rune: rune(f.From[0]), Advance: f.Advance, Path: f.Path}
	return g
}

// TestItalicNeighbours asks for at least 20 units of white between the inks
// of the pairs which TeX sets without space and without
// the italics correction: f before and after the parentheses and brackets,
// and ij, xj, fj and jj.
func TestItalicNeighbours(t *testing.T) {
	fs := textForms(forms(t))
	gs, err := Glyphs(design.Regular())
	if err != nil {
		t.Fatal(err)
	}
	signs := byName(gs)
	get := func(n string) (outline.Path, float64) {
		if f, ok := fs[n]; ok {
			return f.Path, f.Advance
		}
		g := signs[n]
		return g.Path, g.Advance
	}
	pairs := [][2]string{
		{"f", "parenleft"}, {"parenleft", "f"}, {"f", "parenright"}, {"f", "bracketright"},
		{"bracketleft", "f"}, {"f", "comma"}, {"i", "j"}, {"x", "j"}, {"f", "j"}, {"j", "j"},
		{"f", "f"}, {"T", "parenright"}, {"V", "parenright"}, {"f", "slash"},
		{"bracketleft", "j"}, {"parenleft", "j"}, {"bracketleft", "g"}, {"bracketleft", "y"},
		{"bracketleft", "p"}, {"bracketleft", "A"}, {"X", "bracketright"},
	}
	for _, pr := range pairs {
		a, adv := get(pr[0])
		b, _ := get(pr[1])
		if d := inkDistance(a, shifted(b, adv, 0)); d < 20 {
			t.Errorf("%s %s: %.0f units of white, want at least 20", pr[0], pr[1], d)
		}
	}
}

// TestItalicCorrectionsClearSuperscripts asks that a superscript set after
// the italics correction stand clear of the letter: the ink of every form
// above the bottom of a superscript (SuperscriptShiftUp) reaches at most 10
// units short of the advance with the italics correction, and the
// correction is no larger than needs be, with at most 10 units of white
// left there for a letter whose ink reaches beyond its advance.
func TestItalicCorrectionsClearSuperscripts(t *testing.T) {
	c := Constants(design.Regular())
	shift := float64(c.SuperscriptShiftUp.V)
	for _, f := range forms(t) {
		_, _, _, top := f.Path.BBox()
		if top <= shift {
			continue
		}
		_, x1 := inkSpan(f.Path, shift, top)
		edge := f.Advance + f.Italic
		if x1 > edge-icPad+0.5 {
			t.Errorf("%s: the ink above %.0f reaches %.0f, the superscript starts at %.0f", f.Name, shift, x1, edge)
		}
		if f.Italic < 0 {
			t.Errorf("%s: italics correction %.0f", f.Name, f.Italic)
		}
	}
}

// TestItalicAccentAttachments asks that the top accent attachment of every
// form lie over its ink near the top, between the leftmost and the
// rightmost ink of its top 15 per cent, widened by the shift of the rule;
// that it lie right of the middle of the advance on the slanted capitals
// and lowercase where the half advance leaves the accent behind; and
// that over b, h and k it stand between the top of the ascender and the
// middle of the bowl or the arch.
func TestItalicAccentAttachments(t *testing.T) {
	xh := design.Regular().XHeight
	for _, f := range forms(t) {
		_, _, _, top := f.Path.BBox()
		lo, hi := inkSpan(f.Path, 0.85*top, top)
		if f.Attach < lo || f.Attach > hi+accentLower*xh+0.5 {
			t.Errorf("%s: the attachment %.0f lies outside the ink of the top, %.0f to %.0f", f.Name, f.Attach, lo, hi)
		}
		switch f.From {
		case "b", "h", "k":
			blo, bhi := inkSpan(f.Path, 0.80*xh, 0.95*xh)
			mid := (lo + hi) / 2
			if f.Attach < mid || f.Attach > (blo+bhi)/2+accentLower*xh+0.5 {
				t.Errorf("%s: the attachment %.0f does not lie between the ascender at %.0f and the bowl at %.0f", f.Name, f.Attach, mid, (blo+bhi)/2)
			}
		}
	}
}

// TestItalicKernsLeaveTheGap places a subscript box with its top at the
// upper end of each interval of the kerns at the bottom right, its left edge
// at the advance plus the kern, and asks for at least kernGap of white to the
// ink of the letter below that top.
func TestItalicKernsLeaveTheGap(t *testing.T) {
	for _, f := range forms(t) {
		if f.Kern == nil || f.Kern.BottomRight == nil {
			continue
		}
		k := f.Kern.BottomRight
		_, y0, _, y1 := f.Path.BBox()
		for i, v := range k.Kerns {
			hi := y1
			if i < len(k.Heights) {
				hi = float64(k.Heights[i].V)
			}
			if v.V == 0 {
				continue
			}
			edge := f.Advance + float64(v.V)
			for y := y0; y <= hi; y += 2 {
				if _, x := inkSpan(f.Path, y, y); x > edge-kernGap+1 {
					t.Errorf("%s: below %.0f the kern %d brings a subscript within %.0f of the ink at %.0f", f.Name, hi, v.V, edge-x, y)
					break
				}
			}
		}
	}
}

// TestScriptFormsAreHeavierAndWider measures the stems of n, l, i and m at
// half the x-height and the top of o, the thinnest place of the round, on
// the script forms against the text forms: the stems 1.08 to 1.17 times as
// heavy at the first level and 1.15 to 1.35 at the second, the thin of o at
// least 36 at the first level and 40 at
// the second, and every form 1.02 to 1.17 and 1.05 to 1.47 times as wide as
// its text form.
func TestScriptFormsAreHeavierAndWider(t *testing.T) {
	fs := forms(t)
	xh := design.Regular().XHeight
	text := textForms(fs)
	stem := func(p outline.Path) float64 {
		xs := p.CrossingsX(xh / 2)
		return xs[1] - xs[0]
	}
	thin := func(p outline.Path) float64 {
		x0, _, x1, _ := p.BBox()
		ys := p.CrossingsY((x0 + x1) / 2)
		return ys[len(ys)-1] - ys[len(ys)-2]
	}
	ranges := [2][2]float64{{1.08, 1.17}, {1.15, 1.35}}
	widths := [2][2]float64{{1.02, 1.17}, {1.05, 1.47}}
	thins := [2]float64{36, 40}
	for _, f := range fs {
		if f.Level == 0 {
			continue
		}
		base := text[f.From]
		if r := f.Advance / base.Advance; r < widths[f.Level-1][0] || r > widths[f.Level-1][1] {
			t.Errorf("%s is %.3f times as wide as %s", f.Name, r, base.Name)
		}
		switch f.From {
		case "n", "l", "i", "m":
			r := stem(f.Path) / stem(base.Path)
			if r < ranges[f.Level-1][0] || r > ranges[f.Level-1][1] {
				t.Errorf("the stem of %s is %.3f times that of %s, want %v", f.Name, r, base.Name, ranges[f.Level-1])
			}
		case "o":
			if d := thin(f.Path); d < thins[f.Level-1]-0.5 {
				t.Errorf("the top of %s is %.1f thick, want at least %v", f.Name, d, thins[f.Level-1])
			}
		}
	}
}
