// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"seehuhn.de/go/membudget"
	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/cmap"
	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/design"
)

// firstSteps are the code points which the steps of the plan done so far
// claim: the roman glyphs of the first tier, the derived signs of step 2,
// the maths italic of step 3, the signs of steps 5 to 9 and 13 to 15 and the
// symbols of steps 10 to 12.
var firstSteps = func() []rune {
	var rs []rune
	for r := '0'; r <= '9'; r++ {
		rs = append(rs, r)
	}
	for r := 'A'; r <= 'Z'; r++ {
		rs = append(rs, r, r+'a'-'A')
	}
	rs = append(rs, []rune("!()*+,./:;<=>?[\\]{|}")...)
	rs = append(rs, 0x2212, 0xB1, 0xD7, 0xF7, 0xAC, 0x2026, 0x2032, 0x2033,
		0x300, 0x301, 0x302, 0x303, 0x304, 0x306, 0x307, 0x308, 0x30A, 0x30C,
		0x131, 0x237, 0x2020, 0x2021, 0xA7, 0xB6, 0x2022)
	rs = append(rs, 0x2236, 0x2217, 0x2219, 0x22C5, 0x2213, 0x2034, 0x2057,
		0x22EF, 0x22EE, 0x22F1, 0x22F0, 0x2016, 0x2223, 0x2225, 0x2264, 0x2265,
		0x226A, 0x226B, 0x2260, 0x338, 0x2261, 0x2308, 0x2309, 0x230A, 0x230B,
		0x2203, 0x2216, 0x22A5, 0x22A4, 0x22BA, 0x27C2, 0x25A1)
	rs = append(rs, newRunes...)
	rs = append(rs, symbolRunes...)
	for r := rune(0x1D434); r <= 0x1D467; r++ {
		if r != 0x1D455 {
			rs = append(rs, r)
		}
	}
	rs = append(rs, 0x210E, 0x210F, 0x1D6A4, 0x1D6A5)
	rs = append(rs, horizontalRunes...)
	for r := rune(0x391); r <= 0x3A9; r++ {
		if r != 0x3A2 {
			rs = append(rs, r)
		}
	}
	return rs
}()

// regularGlyphs returns the glyphs of the font of the regular parameters,
// built once.
var regularGlyphs = sync.OnceValues(func() ([]Glyph, error) { return Glyphs(design.Regular()) })

// written builds the font, writes it and reads it back.
func written(t *testing.T) (*sfnt.Font, []Glyph) {
	t.Helper()
	p := design.Regular()
	gs, err := regularGlyphs()
	if err != nil {
		t.Fatal(err)
	}
	f, err := Font(p, gs, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	back, err := sfnt.Read(bytes.NewReader(buf.Bytes()), membudget.New(64<<20))
	if err != nil {
		t.Fatal(err)
	}
	return back, gs
}

// TestFont reads the written font back and checks what LuaTeX and XeTeX
// read: every claimed code point in the format 12 character map and, below
// U+10000, in a format 4 subtable under (3, 1); the constants as computed;
// a top accent attachment for every glyph with ink; and the script math with
// its default language system and the feature ssty, an alternate
// substitution with two alternates for each glyph it covers.
func TestFont(t *testing.T) {
	if testing.Short() {
		t.Skip("the font is built in the full tier")
	}
	f, gs := written(t)
	best, err := f.CMapTable.GetBest()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := best.(cmap.Format12); !ok {
		t.Errorf("the best character map is %T, want format 12", best)
	}
	sub, err := f.CMapTable.Get(cmap.Key{PlatformID: 3, EncodingID: 1})
	if err != nil {
		t.Fatalf("no character map under (3, 1): %v", err)
	}
	if _, ok := sub.(cmap.Format4); !ok {
		t.Errorf("the character map under (3, 1) is %T, want format 4", sub)
	}
	for _, r := range firstSteps {
		if best.Lookup(r) == 0 {
			t.Errorf("%U is not in the format 12 character map", r)
		}
		if r <= 0xFFFF && sub.Lookup(r) == 0 {
			t.Errorf("%U is not in the format 4 character map", r)
		}
	}
	if f.Math == nil {
		t.Fatal("no MATH table")
	}
	if f.Math.Constants != Constants(design.Regular()) {
		t.Error("the constants read back differ from those written")
	}
	for i, g := range gs {
		if len(g.Path) > 0 {
			if _, ok := f.Math.TopAccent[glyphID(i)]; !ok {
				t.Errorf("%s has no top accent attachment", g.Name)
			}
		}
	}
	if f.Gsub == nil {
		t.Fatal("no GSUB table")
	}
	found := false
	for tag, feats := range f.Gsub.ScriptList {
		if !strings.Contains(tag.String(), "Zmth") {
			continue
		}
		found = true
		if feats == nil || len(feats.Optional) == 0 {
			t.Errorf("the script math has no feature")
			continue
		}
		for _, fi := range feats.Optional {
			ft := f.Gsub.FeatureList[fi]
			if ft.Tag != "ssty" {
				continue
			}
			for _, li := range ft.Lookups {
				for _, st := range f.Gsub.LookupList[li].Subtables {
					a, ok := st.(*gtab.Gsub3_1)
					if !ok {
						t.Errorf("ssty has a subtable %T, want an alternate substitution", st)
						continue
					}
					for _, alts := range a.Alternates {
						if len(alts) != 2 {
							t.Errorf("ssty has %d alternates for a glyph, want 2", len(alts))
						}
					}
				}
			}
		}
	}
	if !found {
		t.Error("no script math")
	}
}

// TestRanges requires the file of ranges to name, as code points or as an
// alphabet, every code point of the font but the space and the remapped ASCII
// signs, and nothing else.
func TestRanges(t *testing.T) {
	gs := []Glyph{{Name: "space", Rune: ' '}, {Name: "colon", Rune: ':'}}
	for _, r := range firstSteps {
		gs = append(gs, Glyph{Name: string(r), Rune: r})
	}
	var buf bytes.Buffer
	if err := WriteRanges(&buf, gs); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "directlua") {
		t.Error("the file of ranges uses \\directlua, which XeLaTeX lacks")
	}
	m := regexp.MustCompile(`\\def\\BedernMathRanges\{([^}]*)\}`).FindStringSubmatch(buf.String())
	if m == nil {
		t.Fatalf("no \\BedernMathRanges in\n%s", buf.String())
	}
	named := map[rune]bool{}
	for _, part := range strings.Split(m[1], ",") {
		alphabet := false
		for _, a := range alphabets {
			if a.name == part {
				for _, r := range a.cps {
					named[r] = true
				}
				alphabet = true
			}
		}
		if alphabet {
			continue
		}
		ends := strings.Split(part, "-")
		var lo, hi int64
		for i, e := range ends {
			v, err := strconv.ParseInt(strings.TrimPrefix(e, `"`), 16, 32)
			if err != nil {
				t.Fatalf("bad range %q", part)
			}
			if i == 0 {
				lo = v
			}
			hi = v
		}
		for r := lo; r <= hi; r++ {
			named[rune(r)] = true
		}
	}
	for _, g := range gs {
		want := g.Rune > ' ' && !remapped[g.Rune]
		if named[g.Rune] != want {
			t.Errorf("%U: named %v, want %v", g.Rune, named[g.Rune], want)
		}
		delete(named, g.Rune)
	}
	for r := range named {
		t.Errorf("%U is named but not in the font", r)
	}
}

func glyphID(i int) glyph.ID { return glyph.ID(i) }
