// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"slices"
	"testing"
)

// testFont is a description which the tests build.
type testFont struct {
	name string
	font *Font
}

// box returns a character with the given dimensions in thousandths of the
// design size.
func box(wd, ht, dp, ic int) *Char {
	return &Char{Width: Units(wd, 1000), Height: Units(ht, 1000), Depth: Units(dp, 1000), Italic: Units(ic, 1000)}
}

// textFont returns a text font with ligatures, kerns, a boundary character
// and the programs of both boundaries.
func textFont() *Font {
	f := NewFont()
	f.CodingScheme = "BEDERN TEST TEXT"
	f.Family = "BEDERN"
	f.Face = 0
	f.Params = []Fix{0, Units(333, 1000), Units(166, 1000), Units(111, 1000), Units(430, 1000), Unity, Units(111, 1000)}
	for c := 'a'; c <= 'z'; c++ {
		f.Chars[c] = box(400+10*int(c-'a'), 430+250*int(c%2), 200*int(c%3/2), 0)
	}
	for c := 'A'; c <= 'Z'; c++ {
		f.Chars[c] = box(600+7*int(c-'A'), 683, 0, 0)
	}
	f.Chars[' '] = box(333, 0, 0, 0)
	f.Chars['-'] = box(333, 250, 0, 0)
	f.Chars['.'] = box(278, 105, 0, 0)
	f.Chars[','] = box(278, 105, 194, 0)
	f.Chars['`'] = box(278, 694, 0, 0)
	f.Chars['\''] = box(278, 694, 0, 0)
	f.Chars[0o13] = box(583, 694, 0, 77)  // ff
	f.Chars[0o14] = box(556, 694, 0, 0)   // fi
	f.Chars[0o15] = box(556, 694, 0, 0)   // fl
	f.Chars[0o16] = box(833, 694, 0, 0)   // ffi
	f.Chars[0o17] = box(833, 694, 0, 0)   // ffl
	f.Chars[0o20] = box(278, 430, 0, 0)   // an initial form of s
	f.Chars[0o21] = box(300, 430, 0, 0)   // a final form of s
	f.Chars[0o173] = box(500, 430, 0, 28) // en dash
	f.Chars[0o174] = box(1000, 430, 0, 28)
	f.Chars[0o134] = box(500, 694, 0, 0)
	f.Chars[0o42] = box(500, 694, 0, 0)
	f.HasBoundary = true
	f.BoundaryChar = ' '
	k := func(n int) Fix { return Units(n, 1000) }
	f.LigTable = []Instr{
		BoundaryLabel(),
		Lig(OpSlashLigOver, 's', 0o20),
		Krn('T', k(-20)),
		Stop(),
		Label('f'),
		Lig(OpLig, 'i', 0o14),
		Lig(OpLig, 'f', 0o13),
		Lig(OpLig, 'l', 0o15),
		Krn('\'', k(78)),
		Krn('?', k(78)),
		Stop(),
		Label(0o13),
		Lig(OpLig, 'i', 0o16),
		Lig(OpLig, 'l', 0o17),
		Krn('\'', k(78)),
		Stop(),
		Label('`'),
		Lig(OpLig, '`', 0o134),
		Stop(),
		Label('\''),
		Lig(OpLig, '\'', 0o42),
		Krn('?', k(111)),
		Stop(),
		Label('-'),
		Lig(OpLig, '-', 0o173),
		Stop(),
		Label(0o173),
		Lig(OpLig, '-', 0o174),
		Stop(),
		Label('s'),
		Lig(OpLigSlash, ' ', 0o21),
		Krn('a', k(-5)),
		Stop(),
		Label('k'),
		Krn('a', k(-56)),
		Label('v'),
		Label('w'),
		Krn('e', k(-28)),
		Krn('a', k(-28)),
		Krn('o', k(-28)),
		Krn(' ', k(-10)),
		Stop(),
		Label('A'),
		Krn('t', k(-28)),
		Krn('C', k(-28)),
		Skip(1),
		Label('L'),
		Krn('y', k(-83)),
		Krn('T', k(-83)),
		Krn('V', k(-111)),
		Stop(),
	}
	f.Chars['?'] = box(472, 694, 0, 0)
	return f
}

// opsFont returns a font with all eight ligature operations and with skips.
func opsFont() *Font {
	f := NewFont()
	f.CodingScheme = "BEDERN TEST OPERATIONS"
	f.Family = "BEDERN OPS"
	f.Face = 13
	for c := 'a'; c <= 'p'; c++ {
		f.Chars[c] = box(500, 400, 0, 0)
	}
	f.Chars[200] = box(700, 400, -10, 0)
	f.Chars[201] = box(0, 0, 0, 0)
	f.Chars[202] = box(-50, 400, 0, -20)
	f.LigTable = []Instr{
		Label('a'),
		Lig(OpLig, 'b', 200),
		Skip(0),
		Lig(OpLigSlash, 'c', 201),
		Skip(2),
		Label('e'),
		Lig(OpSlashLig, 'd', 202),
		Lig(OpSlashLigSlash, 'e', 200),
		Label('b'),
		Lig(OpLigSlashOver, 'f', 201),
		Lig(OpSlashLigOver, 'g', 202),
		Stop(),
		Label('c'),
		Lig(OpSlashLigSlashOver, 'h', 200),
		Lig(OpSlashLigSlashOver2, 'i', 201),
		Krn('j', -Unity/8),
		Krn('k', Unity/8),
		Krn('l', -Unity/8),
		Lig(OpLig, 'j', 202),
		Stop(),
		Label('d'),
		Krn('a', Units(-123456, 1000000)),
	}
	return f
}

// symbolFont returns a font of mathematical symbols with 22 parameters.
func symbolFont() *Font {
	f := NewFont()
	f.CodingScheme = "TEX MATH SYMBOLS"
	f.Family = "BEDERN SY"
	f.DesignSize = 7 * Unity
	f.Params = make([]Fix, 22)
	for i := range f.Params {
		f.Params[i] = Units(100+37*i, 1000)
	}
	f.Params[0] = Units(250, 1000)
	for c := 0; c < 128; c++ {
		f.Chars[c] = box(500+c, 300+c%5*100, c%4*50, c%3*20)
	}
	f.LigTable = []Instr{
		Label(0o60),
		Krn(0o60, Units(50, 1000)),
		Krn('A', Units(-50, 1000)),
		Stop(),
	}
	return f
}

// extensionFont returns a font of mathematical extension characters with 13
// parameters, lists of larger characters and recipes.
func extensionFont() *Font {
	f := NewFont()
	f.CodingScheme = "TEX MATH EXTENSION"
	f.Family = "BEDERN EX"
	f.Face = 200
	f.Params = make([]Fix, 13)
	for i := range f.Params {
		f.Params[i] = Units(40+11*i, 1000)
	}
	for c := 0; c < 100; c++ {
		f.Chars[c] = box(450+3*c, 40, 1160+600*(c/20), 0)
	}
	for start := 0; start < 4; start++ {
		c := start
		for c+16 < 80 {
			f.Chars[c].Tag = ListTag
			f.Chars[c].NextLarger = uint8(c + 16)
			c += 16
		}
		f.Chars[c].Tag = ExtTag
		f.Chars[c].Recipe = Recipe{Top: uint8(80 + start), Mid: uint8(84 + start), Bot: uint8(88 + start), Rep: uint8(92 + start)}
	}
	f.Chars[96].Tag = ExtTag
	f.Chars[96].Recipe = Recipe{Rep: 97}
	f.Chars[98].Tag = ExtTag
	f.Chars[98].Recipe = Recipe{Top: 99, Rep: 0}
	return f
}

// bigFont returns a font whose ligature and kerning program has more than
// 255 steps and many different programs.
func bigFont(boundary bool) *Font {
	f := NewFont()
	f.CodingScheme = "BEDERN TEST BIG"
	f.Family = "BEDERN BIG"
	for c := 0; c < 256; c++ {
		f.Chars[c] = box(300+c/2, 500, 0, 0)
	}
	progs := make(map[int][]Step)
	for c := 0; c < 250; c++ {
		n := 2 + c%7
		for j := 0; j < n; j++ {
			progs[c] = append(progs[c], Step{Next: uint8((c*5 + j*17) % 256), Op: OpKern, Kern: Units(-(c%40)-j, 1000)})
		}
		if c%10 == 3 {
			progs[c] = progs[c-1]
		}
	}
	progs[250] = []Step{{Next: 1, Op: OpLig, Result: 2}}
	if boundary {
		f.HasBoundary = true
		f.BoundaryChar = 255
		progs[Boundary] = []Step{{Next: 'A', Op: OpKern, Kern: Units(15, 1000)}}
		progs[251] = []Step{{Next: 255, Op: OpKern, Kern: Units(-15, 1000)}}
	}
	f.LigTable = Programs(progs)
	return f
}

// reductionFont returns a font with more values in every table of
// dimensions than the format holds.
func reductionFont() *Font {
	f := NewFont()
	f.CodingScheme = "BEDERN TEST REDUCTION"
	f.Family = "BEDERN RED"
	for c := 0; c < 256; c++ {
		f.Chars[c] = box(200+7*c+c*c%7, 400+c*c%97*3, c*c%41*5-20, c%80*2-30)
	}
	return f
}

// headerFont returns a font with further header words, a stated check sum
// and other units.
func headerFont() *Font {
	f := NewFont()
	f.CodingScheme = ""
	f.Family = "B"
	f.Face = 17
	f.DesignSize = Units(12345, 1000)
	f.DesignUnits = 100 * Unity
	f.HasChecksum = true
	f.Checksum = 0xDEADBEEF
	f.Header = []uint32{1, 0, 0xFFFFFFFF, 0o1234567}
	f.Params = []Fix{-Unity / 4, 30 * Unity, 0, 0, 0, 0, 0, 0, 55 * Unity / 10}
	f.Chars[200] = &Char{Width: 50 * Unity, Height: 70 * Unity, Depth: -1 * Unity, Italic: 333 * Unity / 7}
	f.Chars[201] = &Char{Width: 0, Height: 1599 * Unity}
	f.Chars[202] = &Char{Width: -1599 * Unity}
	f.HasBoundary = true
	f.BoundaryChar = 7
	f.LigTable = []Instr{Label(200), Krn(7, 12*Unity), Krn(201, -12*Unity), Stop()}
	return f
}

// emptyFont returns a font without characters.
func emptyFont() *Font {
	f := NewFont()
	f.SevenBitSafe = true
	return f
}

// testFonts returns the descriptions which the tests build.
func testFonts() []testFont {
	return []testFont{
		{"text", textFont()},
		{"ops", opsFont()},
		{"symbol", symbolFont()},
		{"extension", extensionFont()},
		{"big", bigFont(false)},
		{"bigboundary", bigFont(true)},
		{"reduction", reductionFont()},
		{"header", headerFont()},
		{"empty", emptyFont()},
	}
}

// build returns the tables and the bytes of a description.
func build(t testing.TB, f *Font) (*File, *Report, []byte) {
	t.Helper()
	file, rep, err := f.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	return file, rep, buf.Bytes()
}

// sameFile reports whether two files have the same tables.
func sameFile(a, b *File) bool {
	return a.First == b.First && a.Last == b.Last &&
		slices.Equal(a.Header, b.Header) && slices.Equal(a.Chars, b.Chars) &&
		slices.Equal(a.Width, b.Width) && slices.Equal(a.Height, b.Height) &&
		slices.Equal(a.Depth, b.Depth) && slices.Equal(a.Italic, b.Italic) &&
		slices.Equal(a.LigKern, b.LigKern) && slices.Equal(a.Kern, b.Kern) &&
		slices.Equal(a.Exten, b.Exten) && slices.Equal(a.Param, b.Param) &&
		bytes.Equal(a.Trailer, b.Trailer)
}

// sameFont reports whether two descriptions say the same.
func sameFont(a, b *Font) bool {
	if a.Checksum != b.Checksum || a.HasChecksum != b.HasChecksum || a.DesignSize != b.DesignSize ||
		a.DesignUnits != b.DesignUnits || a.CodingScheme != b.CodingScheme || a.Family != b.Family ||
		a.Face != b.Face || a.SevenBitSafe != b.SevenBitSafe || a.BoundaryChar != b.BoundaryChar ||
		a.HasBoundary != b.HasBoundary || !slices.Equal(a.Header, b.Header) || !slices.Equal(a.Params, b.Params) ||
		!slices.Equal(a.LigTable, b.LigTable) || !slices.Equal(a.RecipeOrder, b.RecipeOrder) {
		return false
	}
	for i := range a.Unused {
		if !slices.Equal(a.Unused[i], b.Unused[i]) {
			return false
		}
	}
	for c, x := range a.Chars {
		y := b.Chars[c]
		if (x == nil) != (y == nil) || x != nil && *x != *y {
			return false
		}
	}
	return true
}
