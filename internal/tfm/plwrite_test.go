// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"reflect"
	"testing"
)

// oddFile is a file which TeX82 accepts and which PLtoTF does not write.
type oddFile struct {
	name  string
	file  *File
	kinds []FindingKind // the findings of File.Check
	gives bool          // TFtoPL ends without success
}

// setString writes a string into the header: the length byte at byte pos
// and the characters behind it.
func setString(f *File, pos int, length int, s string) {
	set := func(i int, b byte) {
		shift := 24 - 8*(i%4)
		f.Header[i/4] = f.Header[i/4]&^(0xFF<<shift) | uint32(b)<<shift
	}
	set(pos, byte(length))
	for i := 0; i < len(s); i++ {
		set(pos+1+i, s[i])
	}
}

// oddFiles returns files which TeX82 accepts and which are not as PLtoTF
// writes them.
func oddFiles(t testing.TB) []oddFile {
	var list []oddFile
	add := func(name string, font *Font, kinds []FindingKind, fn func(f *File)) {
		file, _, _ := build(t, font)
		fn(file)
		if err := file.validate(); err != nil {
			t.Fatalf("%s: not valid: %v", name, err)
		}
		gives := false
		for _, k := range kinds {
			if k == LigatureLoop || k == TooManyPairs {
				gives = true
			}
		}
		list = append(list, oddFile{name, file, kinds, gives})
	}
	char := func(f *File, c int) *CharInfo { return &f.Chars[c-f.First] }

	add("lowercase", textFont(), nil, func(f *File) {
		setString(f, 8, 8, "TeX text")
		setString(f, 48, 6, "bedern")
	})
	add("blank", textFont(), nil, func(f *File) {
		setString(f, 8, 9, "  LEADING")
	})
	add("padding", textFont(), nil, func(f *File) {
		setString(f, 8, 3, "ABCDEFGH")
		f.Header[17] |= 0x00123400
	})
	add("parenthesis", textFont(), []FindingKind{StringParenthesis, StringParenthesis}, func(f *File) {
		setString(f, 8, 6, "A(B)C ")
	})
	add("long", textFont(), []FindingKind{StringTooLong}, func(f *File) {
		setString(f, 48, 25, "TOO LONG A FAMILY")
	})
	add("ascii", textFont(), []FindingKind{StringNotASCII, StringNotASCII}, func(f *File) {
		setString(f, 8, 4, "A\x07\xe9B")
	})
	add("ligop", textFont(), []FindingKind{NonstandardLig}, func(f *File) {
		f.LigKern[3].Op = 4
		f.LigKern[4].Op = 9
	})
	add("loop", textFont(), []FindingKind{LigatureLoop}, func(f *File) {
		i := int(char(f, 'f').Remainder)
		f.LigKern[i] = LigKern{0, 'i', 2, 'l'}
		f.LigKern[i+2] = LigKern{0, 'l', 2, 'i'}
	})
	add("unused", textFont(), []FindingKind{UnusedSteps}, func(f *File) {
		f.LigKern[int(char(f, 'f').Remainder)].Skip = 128
		f.LigKern[int(char(f, 'A').Remainder)].Skip = 3
	})
	add("larger", textFont(), []FindingKind{LargerMissing}, func(f *File) {
		char(f, 'x').Tag, char(f, 'x').Remainder = ListTag, 0o22
	})
	add("header2", textFont(), nil, func(f *File) { f.Header = f.Header[:2] })
	add("header12", textFont(), nil, func(f *File) { f.Header = f.Header[:12] })
	add("header17", textFont(), nil, func(f *File) { f.Header = f.Header[:17] })
	add("params", symbolFont(), []FindingKind{ParamCount}, func(f *File) { f.Param = f.Param[:7] })
	add("params13", extensionFont(), []FindingKind{ParamCount}, func(f *File) { f.Param = append(f.Param, 0) })
	add("trailer", textFont(), []FindingKind{TrailingBytes}, func(f *File) { f.Trailer = []byte{1, 2, 3, 4, 5} })
	add("nostep", opsFont(), nil, func(f *File) {
		n := len(f.LigKern)
		f.LigKern = append(f.LigKern, LigKern{254, 0, 0, uint8(n + 1)}, LigKern{200, 7, 0, 0})
		char(f, 'p').Tag, char(f, 'p').Remainder = LigTag, uint8(n)
	})
	add("zero", textFont(), nil, func(f *File) {
		f.Height = append(f.Height, 0)
		char(f, 'a').HeightIndex = uint8(len(f.Height) - 1)
		f.Depth = append(f.Depth, f.Depth[1], -Unity)
		f.Width[1], f.Width[2] = f.Width[2], f.Width[1]
	})
	add("width16", textFont(), nil, func(f *File) { f.Width[1] = -16 * Unity })
	add("range", emptyFont(), nil, func(f *File) { f.First, f.Last = 200, 199 })
	add("missing", textFont(), nil, func(f *File) {
		*char(f, 0o22) = CharInfo{Tag: LigTag, Remainder: char(f, 'f').Remainder}
		*char(f, 0o23) = CharInfo{Tag: ListTag, Remainder: 'a'}
	})
	add("slant", textFont(), nil, func(f *File) { f.Param[0] = -1500 * Unity })
	add("recipes", extensionFont(), nil, func(f *File) {
		f.Exten[0], f.Exten[1] = f.Exten[1], f.Exten[0]
		f.Exten = append(f.Exten, f.Exten[0])
		char(f, 96).Remainder = char(f, 98).Remainder
	})
	add("kerns", textFont(), nil, func(f *File) {
		f.Kern = append(f.Kern, f.Kern[0], 12345)
		f.Kern[0], f.Kern[1] = f.Kern[1], f.Kern[0]
	})
	many := NewFont()
	var steps []Step
	for c := 0; c < 256; c++ {
		many.Chars[c] = box(500, 0, 0, 0)
		if c > 0 {
			steps = append(steps, Step{Next: uint8(c), Op: OpKern, Kern: Units(c, 1000)})
		}
	}
	many.LigTable = Programs(map[int][]Step{0: steps})
	add("pairs", many, []FindingKind{TooManyPairs}, func(f *File) {
		for c := 1; c <= 130; c++ {
			char(f, c).Tag, char(f, c).Remainder = LigTag, char(f, 0).Remainder
		}
	})
	return list
}

// TestWritePL compares the PL text of the fonts of the tests and of the odd
// files with the output of TFtoPL.
func TestWritePL(t *testing.T) {
	check := func(name string, file *File) {
		var buf bytes.Buffer
		if err := file.WritePL(&buf); err != nil {
			t.Errorf("%s: WritePL failed: %v", name, err)
			return
		}
		want := golden(t, name+".pl")
		if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("%s: PL text differs from the output of TFtoPL at byte %d", name, firstDiff(buf.Bytes(), want))
		}
	}
	for _, tf := range testFonts() {
		file, _, _ := build(t, tf.font)
		check(tf.name, file)
	}
	for _, of := range oddFiles(t) {
		check("odd-"+of.name, of.file)
	}
}

// TestCheck compares the findings about the odd files with the expected
// ones.
func TestCheck(t *testing.T) {
	for _, tf := range testFonts() {
		file, _, _ := build(t, tf.font)
		if finds := file.Check(); len(finds) != 0 {
			t.Errorf("%s: findings %v", tf.name, finds)
		}
	}
	for _, of := range oddFiles(t) {
		var kinds []FindingKind
		for _, fd := range of.file.Check() {
			kinds = append(kinds, fd.Kind)
		}
		if !reflect.DeepEqual(kinds, of.kinds) {
			t.Errorf("%s: findings %v, want %v", of.name, kinds, of.kinds)
		}
	}
}

// TestWriteInvalidPL checks that the PL text of invalid tables is refused.
func TestWriteInvalidPL(t *testing.T) {
	file, _, _ := build(t, textFont())
	file.Width[0] = 1
	if err := file.WritePL(&bytes.Buffer{}); err == nil {
		t.Error("no error for invalid tables")
	}
	bad := []func(f *Font){
		func(f *Font) { f.CodingScheme = "lower case" },
		func(f *Font) { f.Family = " BLANK" },
		func(f *Font) { f.Family = "PAREN(" },
		func(f *Font) { f.Params = make([]Fix, 256) },
		func(f *Font) { f.Header = make([]uint32, 239) },
		func(f *Font) { f.Chars['a'].Width = minFix },
		func(f *Font) { f.Chars['a'].Tag = LigTag },
		func(f *Font) { f.LigTable[0].Kind = 99 },
		func(f *Font) { f.LigTable[1].Step.Op = 4 },
		func(f *Font) { f.RecipeOrder = []uint8{200} },
	}
	for i, fn := range bad {
		f := textFont()
		fn(f)
		if err := f.WritePL(&bytes.Buffer{}); err == nil {
			t.Errorf("case %d: no error for a description which a PL file cannot hold", i)
		}
	}
}

// TestDescriptionPL checks that the PL text of a description reads back as
// the description.
func TestDescriptionPL(t *testing.T) {
	for _, tf := range testFonts() {
		got, err := ReadPL(bytes.NewReader(description(t, tf.font)))
		if err != nil {
			t.Errorf("%s: ReadPL failed: %v", tf.name, err)
			continue
		}
		if !reflect.DeepEqual(got, tf.font) {
			t.Errorf("%s: the PL text reads as another description", tf.name)
		}
	}
}
