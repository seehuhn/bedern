// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// TestReadWrite checks that the bytes of every font of the tests read into
// tables which write back to the same bytes.
func TestReadWrite(t *testing.T) {
	for _, tf := range testFonts() {
		file, _, data := build(t, tf.font)
		got, err := Read(bytes.NewReader(data))
		if err != nil {
			t.Errorf("%s: read failed: %v", tf.name, err)
			continue
		}
		if !reflect.DeepEqual(got, file) {
			t.Errorf("%s: the tables read differ from the tables built", tf.name)
		}
		var buf bytes.Buffer
		if err := got.Write(&buf); err != nil {
			t.Errorf("%s: write failed: %v", tf.name, err)
		} else if !bytes.Equal(buf.Bytes(), data) {
			t.Errorf("%s: write gives other bytes, first at %d", tf.name, firstDiff(buf.Bytes(), data))
		}
	}
}

// TestReadOdd checks that the odd files, which are not as PLtoTF writes
// them, read and write back byte for byte.
func TestReadOdd(t *testing.T) {
	for _, of := range oddFiles(t) {
		data := of.file.bytes()
		got, err := Read(bytes.NewReader(data))
		if err != nil {
			t.Errorf("%s: read failed: %v", of.name, err)
			continue
		}
		var buf bytes.Buffer
		if err := got.Write(&buf); err != nil {
			t.Errorf("%s: write failed: %v", of.name, err)
		} else if !bytes.Equal(buf.Bytes(), data) {
			t.Errorf("%s: write gives other bytes, first at %d", of.name, firstDiff(buf.Bytes(), data))
		}
	}
}

// invalidFile holds tables which Read refuses.
type invalidFile struct {
	name     string
	file     *File
	stricter bool // the rule is one of the three which are stricter than TeX
}

// invalidFiles returns tables which break one rule each.
func invalidFiles(t testing.TB) []invalidFile {
	var list []invalidFile
	add := func(name string, font *Font, fn func(f *File)) {
		file, _, _ := build(t, font)
		fn(file)
		list = append(list, invalidFile{name: name, file: file})
	}
	char := func(f *File, c int) *CharInfo { return &f.Chars[c-f.First] }
	text, ext := textFont(), extensionFont()
	add("width index", text, func(f *File) { char(f, 'a').WidthIndex = uint8(len(f.Width)) })
	add("height index", text, func(f *File) { char(f, 'a').HeightIndex = uint8(len(f.Height)) })
	add("depth index", text, func(f *File) { char(f, 'a').DepthIndex = uint8(len(f.Depth)) })
	add("italic index", text, func(f *File) { char(f, 'a').ItalicIndex = uint8(len(f.Italic)) })
	add("index of a missing character", text, func(f *File) { char(f, 0o22).HeightIndex = 15 })
	add("start of a program", text, func(f *File) {
		*char(f, 'a') = CharInfo{WidthIndex: 1, Tag: LigTag, Remainder: uint8(len(f.LigKern))}
	})
	add("recipe index", ext, func(f *File) { char(f, 1).Tag, char(f, 1).Remainder = ExtTag, uint8(len(f.Exten)) })
	add("larger character out of range", text, func(f *File) { char(f, 'a').Tag, char(f, 'a').Remainder = ListTag, 200 })
	add("cycle of larger characters", text, func(f *File) {
		char(f, 'a').Tag, char(f, 'a').Remainder = ListTag, 'b'
		char(f, 'b').Tag, char(f, 'b').Remainder = ListTag, 'a'
	})
	add("larger character of itself", text, func(f *File) { char(f, 'a').Tag, char(f, 'a').Remainder = ListTag, 'a' })
	add("large width", text, func(f *File) { f.Width[1] = 16 * Unity })
	add("large negative depth", text, func(f *File) { f.Depth[1] = -16*Unity - 1 })
	add("large kern", text, func(f *File) { f.Kern[0] = 16 * Unity })
	add("large parameter", text, func(f *File) { f.Param[1] = 16 * Unity })
	add("character examined", text, func(f *File) { f.LigKern[3].Next = 0o22 })
	add("character inserted", text, func(f *File) { f.LigKern[1].Remainder = 0o22 })
	add("kern index", text, func(f *File) { f.LigKern[2].Remainder = uint8(len(f.Kern)) })
	add("skip", text, func(f *File) { f.LigKern[len(f.LigKern)-2].Skip = 1 })
	add("address", text, func(f *File) { f.LigKern[0].Remainder = uint8(len(f.LigKern)) })
	add("piece of a recipe", ext, func(f *File) { f.Exten[0].Top = 200 })
	add("repeated piece of a recipe", ext, func(f *File) { f.Exten[0].Rep = 200 })
	add("small design size", text, func(f *File) { f.Header[1] = uint32(Unity - 1) })
	add("negative design size", text, func(f *File) { f.Header[1] = 0x80000000 })
	n := len(list)
	add("width 0", text, func(f *File) { f.Width[0] = 1 })
	add("height 0", text, func(f *File) { f.Height[0] = 1 })
	add("depth 0", text, func(f *File) { f.Depth[0] = -1 })
	add("italic correction 0", text, func(f *File) { f.Italic[0] = 1 })
	add("too many recipes", ext, func(f *File) { f.Exten = append(f.Exten, make([]Recipe, 256)...) })
	for i := n; i < len(list); i++ {
		list[i].stricter = true
	}
	return list
}

// TestReadInvalidTables checks that Read and Write refuse tables which TeX
// refuses.
func TestReadInvalidTables(t *testing.T) {
	for _, c := range invalidFiles(t) {
		var invalid *InvalidFontError
		if _, err := Read(bytes.NewReader(c.file.bytes())); !errors.As(err, &invalid) {
			t.Errorf("%s: read gives %v, want an InvalidFontError", c.name, err)
		}
		if err := c.file.Write(&bytes.Buffer{}); !errors.As(err, &invalid) {
			t.Errorf("%s: write gives %v, want an InvalidFontError", c.name, err)
		}
	}
}

// input is a named input of Read.
type input struct {
	name string
	data []byte
}

// notTFMInputs returns inputs whose lengths do not describe them.
func notTFMInputs(t testing.TB) []input {
	_, _, data := build(t, textFont())
	set := func(i int, v int) []byte {
		d := bytes.Clone(data)
		d[2*i], d[2*i+1] = byte(v>>8), byte(v)
		return d
	}
	lengths := func(i int) int { return int(data[2*i])<<8 | int(data[2*i+1]) }
	return []input{
		{"empty input", nil},
		{"five words", data[:20]},
		{"cut at the end", data[:len(data)-4]},
		{"cut in a word", data[:len(data)-1]},
		{"more than the largest file", make([]byte, MaxFileSize+1)},
		{"other length", set(0, lengths(0)+1)},
		{"negative length", set(0, lengths(0)|0x8000)},
		{"short header", set(1, 1)},
		{"first code after last code", set(2, lengths(3)+2)},
		{"last code above 255", set(3, 256)},
		{"no widths", set(4, 0)},
		{"other number of heights", set(5, lengths(5)+1)},
		{"other metric format", append([]byte{0, 11, 0, 2}, data[:len(data)-4]...)},
	}
}

// TestReadNotTFM checks that Read refuses input whose lengths do not
// describe it.
func TestReadNotTFM(t *testing.T) {
	for _, c := range notTFMInputs(t) {
		var notTFM *NotTFMError
		if _, err := Read(bytes.NewReader(c.data)); !errors.As(err, &notTFM) {
			t.Errorf("%s: read gives %v, want a NotTFMError", c.name, err)
		}
	}
}

// TestQueries checks the functions which query a file and a description.
func TestQueries(t *testing.T) {
	font := textFont()
	file, _, _ := build(t, font)
	if got, want := file.CharWidth('b'), Units(410, 1000); got != want {
		t.Errorf("width of b is %v, want %v", got, want)
	}
	if got := file.CharWidth(200); got != 0 || file.Exists(200) {
		t.Errorf("character 200 exists with width %v", got)
	}
	if got, want := file.CharItalic(0o13), Units(77, 1000); got != want {
		t.Errorf("italic correction is %v, want %v", got, want)
	}
	if c, ok := file.BoundaryChar(); !ok || c != ' ' {
		t.Errorf("boundary character is %d, %v", c, ok)
	}
	if got, want := file.DesignSize(), 10*Unity; got != want {
		t.Errorf("design size is %v, want %v", got, want)
	}
	if s, ok := file.CodingScheme(); !ok || s != font.CodingScheme {
		t.Errorf("coding scheme is %q", s)
	}
	if s, ok := file.Family(); !ok || s != font.Family {
		t.Errorf("family is %q", s)
	}
	steps := []struct {
		left, right int
		want        Step
		ok          bool
	}{
		{'f', 'i', Step{Next: 'i', Op: OpLig, Result: 0o14}, true},
		{'f', 'f', Step{Next: 'f', Op: OpLig, Result: 0o13}, true},
		{0o13, 'l', Step{Next: 'l', Op: OpLig, Result: 0o17}, true},
		{'f', '\'', Step{Next: '\'', Op: OpKern, Kern: Units(78, 1000)}, true},
		{'k', 'a', Step{Next: 'a', Op: OpKern, Kern: Units(-56, 1000)}, true},
		{'v', 'a', Step{Next: 'a', Op: OpKern, Kern: Units(-28, 1000)}, true},
		{'k', 'o', Step{Next: 'o', Op: OpKern, Kern: Units(-28, 1000)}, true},
		{'A', 'y', Step{}, false},
		{'A', 'T', Step{Next: 'T', Op: OpKern, Kern: Units(-83, 1000)}, true},
		{'L', 'y', Step{Next: 'y', Op: OpKern, Kern: Units(-83, 1000)}, true},
		{'s', Boundary, Step{Next: ' ', Op: OpLigSlash, Result: 0o21}, true},
		{'w', Boundary, Step{Next: ' ', Op: OpKern, Kern: Units(-10, 1000)}, true},
		{Boundary, 's', Step{Next: 's', Op: OpSlashLigOver, Result: 0o20}, true},
		{Boundary, 'T', Step{Next: 'T', Op: OpKern, Kern: Units(-20, 1000)}, true},
		{Boundary, 'a', Step{}, false},
		{'a', 'b', Step{}, false},
		{200, 'b', Step{}, false},
	}
	for _, s := range steps {
		if got, ok := file.Step(s.left, s.right); got != s.want || ok != s.ok {
			t.Errorf("file: step of %d and %d is %+v, %v, want %+v, %v", s.left, s.right, got, ok, s.want, s.ok)
		}
		if got, ok := font.Step(s.left, s.right); got != s.want || ok != s.ok {
			t.Errorf("description: step of %d and %d is %+v, %v, want %+v, %v", s.left, s.right, got, ok, s.want, s.ok)
		}
		kern := Fix(0)
		if s.want.Op == OpKern {
			kern = s.want.Kern
		}
		if got := file.KernBetween(s.left, s.right); got != kern {
			t.Errorf("kern between %d and %d is %v, want %v", s.left, s.right, got, kern)
		}
	}
	for _, tf := range testFonts() {
		if tf.font.DesignUnits != Unity {
			continue
		}
		file, _, _ := build(t, tf.font)
		for c := 0; c <= Boundary; c++ {
			if a, b := file.Program(c), tf.font.Program(c); !reflect.DeepEqual(a, b) {
				t.Errorf("%s: the program of %d is %v in the file and %v in the description", tf.name, c, a, b)
			}
		}
	}
	ext, _, _ := build(t, extensionFont())
	if c, ok := ext.NextLarger(0); !ok || c != 16 {
		t.Errorf("next larger character of 0 is %d, %v", c, ok)
	}
	if r, ok := ext.Recipe(64); !ok || r != (Recipe{80, 84, 88, 92}) {
		t.Errorf("recipe of 64 is %v, %v", r, ok)
	}
	if _, ok := ext.Recipe(0); ok {
		t.Errorf("character 0 has a recipe")
	}
}

// checkFile checks what holds for every file which Read accepts: the
// tables write back to the bytes, the two readers of the PL text agree, and
// the description builds a file with the same description.
func checkFile(data []byte, file *File) error {
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		return fmt.Errorf("write gives other bytes, first at %d", firstDiff(buf.Bytes(), data))
	}
	again, err := Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("second read failed: %v", err)
	}
	if !sameFile(again, file) {
		return errors.New("the second read gives other tables")
	}
	finds := file.Check()
	desc := file.Font()
	buf.Reset()
	if err := file.WritePL(&buf); err != nil {
		return fmt.Errorf("WritePL failed: %v", err)
	}
	mayFail := reachesNoStep(file) || len(file.Header) > 256 || len(file.Param) > 255 ||
		len(file.Param) > 0 && file.Param[0] == minFix
	cut := false
	for _, fd := range finds {
		if fd.Kind == LigatureLoop || fd.Kind == TooManyPairs {
			// the PL text ends after the ligature table
			mayFail, cut = true, true
		}
	}
	if fromPL, err := ReadPL(bytes.NewReader(buf.Bytes())); err == nil {
		if !cut && !sameFont(fromPL, desc) {
			return errors.New("the PL text reads as another description than File.Font gives")
		}
	} else if !mayFail {
		return fmt.Errorf("the PL text does not read: %v", err)
	}
	built, rep, err := desc.Build()
	if err != nil {
		return nil
	}
	buf.Reset()
	if err := built.Write(&buf); err != nil {
		return fmt.Errorf("the file built from the description is not valid: %v", err)
	}
	desc2 := built.Font()
	desc.SevenBitSafe = rep.SevenBitSafe
	if len(desc.Params) > 0 && desc.Params[0] < -1024*Unity {
		// the slant of the file built is not compared
		desc2.Params[0] = desc.Params[0]
	}
	if !sameFont(desc2, desc) {
		return errors.New("the file built from the description has another description")
	}
	return nil
}

// reachesNoStep reports whether a program of the file runs into a word
// which is no step.  TFtoPL prints a stop without a step for it, which
// PLtoTF does not read after a label, a stop or a skip.
func reachesNoStep(f *File) bool {
	l := f.list()
	for i, s := range l.lig {
		if l.activity[i] == accessible && s.Skip > 128 {
			return true
		}
	}
	return false
}

// TestCheckFile applies the checks of the fuzzer to the fonts of the tests.
func TestCheckFile(t *testing.T) {
	for _, tf := range testFonts() {
		file, _, data := build(t, tf.font)
		if err := checkFile(data, file); err != nil {
			t.Errorf("%s: %v", tf.name, err)
		}
	}
	for _, of := range oddFiles(t) {
		if err := checkFile(of.file.bytes(), of.file); err != nil {
			t.Errorf("%s: %v", of.name, err)
		}
	}
}

func FuzzRead(f *testing.F) {
	for _, tf := range testFonts() {
		_, _, data := build(f, tf.font)
		f.Add(data)
	}
	for _, of := range oddFiles(f) {
		f.Add(of.file.bytes())
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		file, err := Read(bytes.NewReader(data))
		if err != nil {
			return
		}
		if err := checkFile(data, file); err != nil {
			t.Fatal(err)
		}
	})
}

// TestTypeset checks the main loop of TeX on a small font: a kern which is
// inserted before a ligature forms stays, a ligature continues with the next
// character, and a ligature which keeps the left character leaves it in place.
func TestTypeset(t *testing.T) {
	f := NewFont()
	f.CodingScheme = "BEDERN TEST TEXT"
	f.Params = []Fix{0, Units(333, 1000), Units(166, 1000), Units(111, 1000), Units(430, 1000), Unity, Units(111, 1000)}
	for _, c := range "ab-cd" {
		f.Chars[c] = box(500, 400, 0, 0)
	}
	k := Units(-50, 1000)
	f.LigTable = Programs(map[int][]Step{
		'a': {{Next: '-', Op: OpKern, Kern: k}},
		'-': {{Next: '-', Op: OpLig, Result: 'c'}},
		'c': {{Next: '-', Op: OpLig, Result: 'd'}, {Next: 'b', Op: OpSlashLig, Result: 'a'}},
	})
	var buf bytes.Buffer
	if _, err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	file, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	kern := Item{Char: -1, Kern: k}
	for _, c := range []struct {
		in   string
		want []Item
	}{
		{"a", []Item{{Char: 'a'}}},
		{"a-b", []Item{{Char: 'a'}, kern, {Char: '-'}, {Char: 'b'}}},
		{"a--", []Item{{Char: 'a'}, kern, {Char: 'c'}}},
		{"a---", []Item{{Char: 'a'}, kern, {Char: 'd'}}},
		{"--b", []Item{{Char: 'c'}, {Char: 'a'}}},
		{"cb", []Item{{Char: 'c'}, {Char: 'a'}}},
	} {
		var codes []int
		for _, r := range c.in {
			codes = append(codes, int(r))
		}
		got, err := file.Typeset(codes)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: %v, want %v", c.in, got, c.want)
		}
	}
}
