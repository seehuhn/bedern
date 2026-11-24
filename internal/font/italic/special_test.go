// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestSpecialLetters asks of the letters made of the strokes of the basic
// ones, upright: ſ has the tail of f and no crossbar, a single run of ink
// just below the x-height; ĸ stands no higher than the entry of n; þ rises
// to the crest of an ascender and stands on the descender line; ŋ has the
// tail of j under its right stem; the i and the j of ĳ keep at least the
// floor of the kerning (20) apart at the x-height; and the dot of ŀ is clear
// of the stem by at least 40.
func TestSpecialLetters(t *testing.T) {
	p := Default()
	checkTail(t, p, "longs", p.FSideLeft)
	up := unsheared(p, glyph(t, "longs").Path)
	if rs := runs(up, p.XHeight-p.Hair/2); len(rs) != 1 || rs[0][1]-rs[0][0] > p.Stem+1 {
		t.Errorf("ſ has %v at the height of the crossbar of f, want the stem alone", rs)
	}
	if _, _, _, y1 := bbox(unsheared(p, glyph(t, "kgreenlandic").Path)); y1 > p.EntryTop+1 {
		t.Errorf("ĸ rises to %.0f, want at most %v", y1, p.EntryTop)
	}
	_, y0, _, y1 := bbox(unsheared(p, glyph(t, "thorn").Path))
	if top := p.XHeight + p.AscenderEntryRise + p.Overshoot - 1; y1 < top-1 || math.Abs(y0+p.Descender) > 1 {
		t.Errorf("þ runs from %.0f to %.0f, want %v to %v", y0, y1, -p.Descender, top)
	}
	checkTail(t, p, "eng", p.SideStemLeft+p.Stem+p.NCounter)
	rs := runs(unsheared(p, glyph(t, "ij").Path), p.XHeight-60)
	if len(rs) < 2 {
		t.Errorf("ĳ has %d runs of ink below the x-height, want 2", len(rs))
	} else if g := rs[1][0] - rs[0][1]; g < 20 {
		t.Errorf("the i and the j of ĳ are %.0f apart, want at least 20", g)
	}
	ld := unsheared(p, glyph(t, "ldot").Path)
	y := p.LDotY * p.XHeight
	rs = runs(ld, y)
	if len(rs) != 2 || rs[1][0]-rs[0][1] < 40 {
		t.Errorf("the dot of ŀ: runs %v at %.0f, want the stem and the dot 40 apart", rs, y)
	}
}

// TestEthSchwaMicro asks of ð that it have the bowl of o in the advance of o,
// two runs of ink at half the x-height as o has, an arm which reaches the
// ascender within the advance and a bar which thickens the arm at its middle;
// of ə that it be e turned through half a turn about the middle of its
// advance and its height, within a unit and a half; and of µ that it be the μ
// of the maths italic.
func TestEthSchwaMicro(t *testing.T) {
	p := Default()
	eth, o := glyph(t, "eth"), glyph(t, "o")
	if eth.Advance != o.Advance {
		t.Errorf("ð has the advance %v, want that of o, %v", eth.Advance, o.Advance)
	}
	y := p.XHeight / 2
	if r, ro := runs(eth.Path, y), runs(o.Path, y); len(r) != 2 || math.Abs(r[0][0]-ro[0][0]) > 2 || math.Abs(r[1][1]-ro[1][1]) > 2 {
		t.Errorf("ð has the runs %v at half the x-height, o has %v", r, ro)
	}
	x0, _, x1, y1 := bbox(eth.Path)
	if asc := p.XHeight + p.AscenderEntryRise; y1 < asc-6 || y1 > asc+p.Overshoot+1 {
		t.Errorf("the arm of ð reaches %.0f, want %v to %v", y1, asc-6, asc+p.Overshoot)
	}
	if x0 < -1 || x1 > eth.Advance+1 {
		t.Errorf("ð has ink from %.0f to %.0f, in the advance %v", x0, x1, eth.Advance)
	}
	// The bar reaches left of the arm: at the height of its left end the ink
	// starts at least 0.8 stems left of where it starts at the height of its
	// middle plus 60, above the bar; and the arm is alone above the bowl.
	top := p.XHeight + p.Overshoot
	mid := top + ethMidY*(p.XHeight+p.AscenderEntryRise+ethTopOver-top) + ethBarOff
	bar := mid - ethBarLeft*p.OBowlWidth*math.Tan(ethBarAngle*math.Pi/180) + 2
	if at, above := runs(eth.Path, bar), runs(eth.Path, mid+60); len(at) == 0 || len(above) != 1 || above[0][0]-at[0][0] < 0.8*p.Stem {
		t.Errorf("ð has %v at the height %.0f of the left end of the bar and %v above it, want ink at least 0.8 stems further left", at, bar, above)
	}
	if rs := runs(eth.Path, top+30); len(rs) != 1 {
		t.Errorf("ð has %d runs above the bowl, want the arm alone", len(rs))
	}

	e, s := glyph(t, "e"), glyph(t, "uni0259")
	ex0, ey0, ex1, ey1 := bbox(e.Path)
	sx0, sy0, sx1, sy1 := bbox(s.Path)
	if s.Advance != e.Advance || math.Abs(sx0-(e.Advance-ex1)) > 1.5 || math.Abs(sx1-(e.Advance-ex0)) > 1.5 ||
		math.Abs(sy0-ey0) > 1.5 || math.Abs(sy1-ey1) > 1.5 {
		t.Errorf("ə is not e turned: boxes %.0f,%.0f,%.0f,%.0f and %.0f,%.0f,%.0f,%.0f in the advance %v",
			sx0, sy0, sx1, sy1, ex0, ey0, ex1, ey1, e.Advance)
	}

	m := glyph(t, "micro")
	mu, err := Build(p, "mu")
	if err != nil {
		t.Fatal(err)
	}
	if m.Advance != mu.Advance || fingerprint(m.Path) != fingerprint(mu.Path) {
		t.Error("µ is not the μ of the maths italic")
	}
	if m.Rune != 'µ' || mu.Rune != 'μ' {
		t.Errorf("µ has the code point %U and μ %U", m.Rune, mu.Rune)
	}
}
