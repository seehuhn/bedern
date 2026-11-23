// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"fmt"
	"io"
)

// MaxFileSize is the largest number of bytes which [Read] accepts.
const MaxFileSize = 4 * 65535

// Boundary stands for the boundary of a word where a character code is
// expected: as the left character of a pair it selects the program of the
// left boundary, as the right character it stands for the boundary character
// of the font.
const Boundary = 256

// NotTFMError is the error which [Read] returns for input whose first six
// words do not describe a TFM file of TeX82.
type NotTFMError struct {
	Reason string
}

func (err *NotTFMError) Error() string {
	return "tfm: not a TFM file of TeX82: " + err.Reason
}

// InvalidFontError is the error for tables which break a rule of the format,
// and for a description which the format cannot hold.
type InvalidFontError struct {
	Reason string
}

func (err *InvalidFontError) Error() string {
	return "tfm: " + err.Reason
}

func notTFM(format string, args ...any) error {
	return &NotTFMError{Reason: fmt.Sprintf(format, args...)}
}

func invalid(format string, args ...any) error {
	return &InvalidFontError{Reason: fmt.Sprintf(format, args...)}
}

// Tag is the tag of a character: the meaning of the remainder in its word.
type Tag uint8

// The tags of a character.
const (
	NoTag   Tag = iota // the remainder is unused
	LigTag             // the remainder is the start of a ligature and kerning program
	ListTag            // the remainder is the code of the next larger character
	ExtTag             // the remainder is the index of an extensible recipe
)

// CharInfo is the word of one character code in a TFM file.  A character
// exists if and only if its WidthIndex is not zero.
type CharInfo struct {
	WidthIndex  uint8 // index into File.Width
	HeightIndex uint8 // index into File.Height, at most 15
	DepthIndex  uint8 // index into File.Depth, at most 15
	ItalicIndex uint8 // index into File.Italic, at most 63
	Tag         Tag
	Remainder   uint8
}

// LigKern is one word of the ligature and kerning program in a TFM file.
//
// A Skip of 128 ends a program after the step, a smaller Skip is the number of
// words to pass over after the step.  A word with a Skip above 128 is no step:
// it ends a program, and 256·Op+Remainder is an address.  In the first word of
// a character's program this address is the start of the program proper; in
// word 0 a Skip of 255 marks Next as the boundary character; in the last word
// a Skip of 255 marks the address as the start of the program of the left
// boundary.
//
// In a step, Next is the character to the right.  An Op below 128 is a
// ligature operation which inserts the character Remainder; an Op of 128 or
// more is a kern with index 256·(Op-128)+Remainder into File.Kern.
type LigKern struct {
	Skip      uint8
	Next      uint8
	Op        uint8
	Remainder uint8
}

// Recipe is an extensible recipe: the codes of the pieces of an extensible
// character.  A Top, Mid or Bot of zero says that the piece is absent.
type Recipe struct {
	Top, Mid, Bot, Rep uint8
}

// File holds the tables of a TFM file as they stand in the file.
type File struct {
	// Header holds the words of the header.  Word 0 is the check sum and
	// word 1 the design size.  Words 2 to 11 hold the coding scheme, words 12
	// to 16 the family and word 17 the seven-bit-safe flag in its first byte
	// and the face in its last byte.
	Header []uint32

	// First and Last are the smallest and the largest character code.  A
	// file without characters has First = Last+1.
	First, Last int

	// Chars holds the words of the character codes First to Last.
	Chars []CharInfo

	// Width, Height, Depth and Italic are the tables of dimensions.  Entry 0
	// of each table is zero.
	Width, Height, Depth, Italic []Fix

	LigKern []LigKern
	Kern    []Fix
	Exten   []Recipe

	// Param holds the font parameters; Param[0] is parameter 1, the slant.
	Param []Fix

	// Trailer holds the bytes which follow the stated length of the file.
	Trailer []byte
}

// Read reads a TFM file.  It makes the checks which TeX82 makes when it loads
// a font, and TeX accepts every file which Read accepts.  The checks are the
// relations between the lengths, the range of the character codes, the range
// of every index, the range of every dimension, the zero in entry 0 of the
// tables of dimensions, the steps of the ligature and kerning program, the
// recipes and the absence of cycles in the lists of larger characters.  Three
// checks are stricter than those of TeX: there are at most 256 recipes, entry
// 0 of the tables of dimensions is zero in every bit and the input has at most
// MaxFileSize bytes.  Read takes at most MaxFileSize+1 bytes from r.
//
// The error is a [*NotTFMError] if the lengths in the first six words do not
// describe the input, and an [*InvalidFontError] if the tables break a rule.
func Read(r io.Reader) (*File, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return nil, err
	}
	return parse(data)
}

// parse decodes the bytes of a TFM file.
func parse(data []byte) (*File, error) {
	if len(data) > MaxFileSize {
		return nil, notTFM("more than %d bytes", MaxFileSize)
	}
	if len(data) < 24 {
		return nil, notTFM("shorter than six words")
	}
	var n [12]int
	for i := range n {
		if data[2*i] > 127 {
			return nil, notTFM("length %d of the first six words is negative", i)
		}
		n[i] = int(data[2*i])<<8 | int(data[2*i+1])
	}
	lf, lh, bc, ec := n[0], n[1], n[2], n[3]
	nw, nh, nd, ni, nl, nk, ne, np := n[4], n[5], n[6], n[7], n[8], n[9], n[10], n[11]
	switch {
	case bc > ec+1 || ec > 255:
		return nil, notTFM("the character codes %d to %d are no valid range", bc, ec)
	case lf != 6+lh+(ec-bc+1)+nw+nh+nd+ni+nl+nk+ne+np:
		return nil, notTFM("the lengths of the tables do not add up to the stated length")
	case 4*lf > len(data):
		return nil, notTFM("the input is shorter than the stated length")
	case lh < 2:
		return nil, notTFM("the header has fewer than two words")
	case nw == 0 || nh == 0 || nd == 0 || ni == 0:
		return nil, notTFM("a table of dimensions is empty")
	}

	pos := 24
	word := func() uint32 {
		w := uint32(data[pos])<<24 | uint32(data[pos+1])<<16 | uint32(data[pos+2])<<8 | uint32(data[pos+3])
		pos += 4
		return w
	}
	fixes := func(n int) []Fix {
		if n == 0 {
			return nil
		}
		t := make([]Fix, n)
		for i := range t {
			t[i] = Fix(word())
		}
		return t
	}

	f := &File{First: bc, Last: ec}
	f.Header = make([]uint32, lh)
	for i := range f.Header {
		f.Header[i] = word()
	}
	if ec >= bc {
		f.Chars = make([]CharInfo, ec-bc+1)
		for i := range f.Chars {
			w := word()
			f.Chars[i] = CharInfo{
				WidthIndex:  uint8(w >> 24),
				HeightIndex: uint8(w >> 20 & 15),
				DepthIndex:  uint8(w >> 16 & 15),
				ItalicIndex: uint8(w >> 10 & 63),
				Tag:         Tag(w >> 8 & 3),
				Remainder:   uint8(w),
			}
		}
	}
	f.Width = fixes(nw)
	f.Height = fixes(nh)
	f.Depth = fixes(nd)
	f.Italic = fixes(ni)
	if nl > 0 {
		f.LigKern = make([]LigKern, nl)
		for i := range f.LigKern {
			w := word()
			f.LigKern[i] = LigKern{uint8(w >> 24), uint8(w >> 16), uint8(w >> 8), uint8(w)}
		}
	}
	f.Kern = fixes(nk)
	if ne > 0 {
		f.Exten = make([]Recipe, ne)
		for i := range f.Exten {
			w := word()
			f.Exten[i] = Recipe{uint8(w >> 24), uint8(w >> 16), uint8(w >> 8), uint8(w)}
		}
	}
	f.Param = fixes(np)
	if pos < len(data) {
		f.Trailer = append([]byte(nil), data[pos:]...)
	}
	if err := f.validate(); err != nil {
		return nil, err
	}
	return f, nil
}

// validate makes the checks of [Read] on the tables of f.
func (f *File) validate() error {
	bc, ec := f.First, f.Last
	lh, nw, nh, nd, ni := len(f.Header), len(f.Width), len(f.Height), len(f.Depth), len(f.Italic)
	nl, nk, ne, np := len(f.LigKern), len(f.Kern), len(f.Exten), len(f.Param)
	switch {
	case bc < 0 || bc > ec+1 || ec > 255:
		return invalid("the character codes %d to %d are no valid range", bc, ec)
	case len(f.Chars) != ec-bc+1:
		return invalid("%d words for the character codes %d to %d", len(f.Chars), bc, ec)
	case lh < 2:
		return invalid("the header has fewer than two words")
	case nw == 0 || nh == 0 || nd == 0 || ni == 0:
		return invalid("a table of dimensions is empty")
	case ne > 256:
		return invalid("%d extensible recipes", ne)
	}
	for _, n := range []int{lh, nw, nh, nd, ni, nl, nk, np} {
		if n > 32767 {
			return invalid("a table has %d words", n)
		}
	}
	lf := 6 + lh + (ec - bc + 1) + nw + nh + nd + ni + nl + nk + ne + np
	if lf > 32767 {
		return invalid("the file has %d words", lf)
	}
	if 4*lf+len(f.Trailer) > MaxFileSize {
		return invalid("the file has more than %d bytes", MaxFileSize)
	}
	if Fix(f.Header[1]) < Unity {
		return invalid("the design size is less than 1")
	}

	for i, ci := range f.Chars {
		c := bc + i
		switch {
		case int(ci.WidthIndex) >= nw:
			return invalid("width index of character %d", c)
		case ci.HeightIndex > 15 || int(ci.HeightIndex) >= nh:
			return invalid("height index of character %d", c)
		case ci.DepthIndex > 15 || int(ci.DepthIndex) >= nd:
			return invalid("depth index of character %d", c)
		case ci.ItalicIndex > 63 || int(ci.ItalicIndex) >= ni:
			return invalid("italic correction index of character %d", c)
		}
		d := int(ci.Remainder)
		switch ci.Tag {
		case NoTag:
		case LigTag:
			if d >= nl {
				return invalid("start of the program of character %d", c)
			}
		case ExtTag:
			if d >= ne {
				return invalid("recipe index of character %d", c)
			}
		case ListTag:
			if d < bc || d > ec {
				return invalid("character %d has the larger character %d", c, d)
			}
			for d < c {
				di := f.Chars[d-bc]
				if di.Tag != ListTag {
					break
				}
				d = int(di.Remainder)
			}
			if d == c {
				return invalid("the list of larger characters of character %d is a cycle", c)
			}
		default:
			return invalid("tag of character %d", c)
		}
	}

	tables := []struct {
		name string
		t    []Fix
	}{{"width", f.Width}, {"height", f.Height}, {"depth", f.Depth}, {"italic correction", f.Italic}}
	for _, tb := range tables {
		if tb.t[0] != 0 {
			return invalid("%s 0 is not zero", tb.name)
		}
		for i, x := range tb.t {
			if !small(x) {
				return invalid("%s %d is not below 16", tb.name, i)
			}
		}
	}

	bchar := -1
	for k, s := range f.LigKern {
		if s.Skip > 128 {
			if int(s.Op)<<8|int(s.Remainder) >= nl {
				return invalid("address in word %d of the ligature and kerning program", k)
			}
			if s.Skip == 255 && k == 0 {
				bchar = int(s.Next)
			}
			continue
		}
		if int(s.Next) != bchar && !f.Exists(int(s.Next)) {
			return invalid("step %d of the ligature and kerning program examines character %d", k, s.Next)
		}
		if s.Op < 128 {
			if !f.Exists(int(s.Remainder)) {
				return invalid("step %d of the ligature and kerning program inserts character %d", k, s.Remainder)
			}
		} else if int(s.Op-128)<<8|int(s.Remainder) >= nk {
			return invalid("kern index in step %d of the ligature and kerning program", k)
		}
		if s.Skip < 128 && k+int(s.Skip)+1 >= nl {
			return invalid("step %d of the ligature and kerning program skips beyond the end", k)
		}
	}
	for i, x := range f.Kern {
		if !small(x) {
			return invalid("kern %d is not below 16", i)
		}
	}

	for i, r := range f.Exten {
		for _, p := range [3]uint8{r.Top, r.Mid, r.Bot} {
			if p != 0 && !f.Exists(int(p)) {
				return invalid("recipe %d uses character %d", i, p)
			}
		}
		if !f.Exists(int(r.Rep)) {
			return invalid("recipe %d repeats character %d", i, r.Rep)
		}
	}

	for i, x := range f.Param {
		if i > 0 && !small(x) {
			return invalid("parameter %d is not below 16", i+1)
		}
	}
	return nil
}

// small reports whether the first byte of x is 0 or 255.
func small(x Fix) bool {
	return x >= -1<<24 && x < 1<<24
}

// encode returns the bytes of the TFM file, if the tables are valid.
func (f *File) encode() ([]byte, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}
	return f.bytes(), nil
}

// bytes returns the bytes of the tables.
func (f *File) bytes() []byte {
	bc, ec := f.First, f.Last
	lf := 6 + len(f.Header) + (ec - bc + 1) + len(f.Width) + len(f.Height) + len(f.Depth) + len(f.Italic) +
		len(f.LigKern) + len(f.Kern) + len(f.Exten) + len(f.Param)
	out := make([]byte, 0, 4*lf+len(f.Trailer))
	for _, n := range []int{lf, len(f.Header), bc, ec, len(f.Width), len(f.Height), len(f.Depth), len(f.Italic),
		len(f.LigKern), len(f.Kern), len(f.Exten), len(f.Param)} {
		out = append(out, byte(n>>8), byte(n))
	}
	word := func(w uint32) {
		out = append(out, byte(w>>24), byte(w>>16), byte(w>>8), byte(w))
	}
	for _, w := range f.Header {
		word(w)
	}
	for _, ci := range f.Chars {
		out = append(out, ci.WidthIndex, ci.HeightIndex<<4|ci.DepthIndex, ci.ItalicIndex<<2|uint8(ci.Tag), ci.Remainder)
	}
	for _, t := range [][]Fix{f.Width, f.Height, f.Depth, f.Italic} {
		for _, x := range t {
			word(uint32(x))
		}
	}
	for _, s := range f.LigKern {
		out = append(out, s.Skip, s.Next, s.Op, s.Remainder)
	}
	for _, x := range f.Kern {
		word(uint32(x))
	}
	for _, r := range f.Exten {
		out = append(out, r.Top, r.Mid, r.Bot, r.Rep)
	}
	for _, x := range f.Param {
		word(uint32(x))
	}
	out = append(out, f.Trailer...)
	return out
}

// Write writes the TFM file to w.  A file which [Read] returned is written
// back byte for byte.  Write returns an [*InvalidFontError], and writes
// nothing, if the tables break a rule which [Read] checks.
func (f *File) Write(w io.Writer) error {
	data, err := f.encode()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Checksum returns the check sum of the header.
func (f *File) Checksum() uint32 {
	if len(f.Header) < 1 {
		return 0
	}
	return f.Header[0]
}

// DesignSize returns the design size in points.
func (f *File) DesignSize() Fix {
	if len(f.Header) < 2 {
		return 0
	}
	return Fix(f.Header[1])
}

// headerByte returns byte i of the header, or zero beyond its end.
func (f *File) headerByte(i int) byte {
	if i/4 >= len(f.Header) {
		return 0
	}
	return byte(f.Header[i/4] >> (24 - 8*(i%4)))
}

// headerString returns the string of at most max-1 bytes which starts at
// byte pos of the header, without change, and whether the header holds it.
func (f *File) headerString(pos, max int) (string, bool) {
	if len(f.Header) < (pos+max)/4 {
		return "", false
	}
	n := int(f.headerByte(pos))
	if n >= max {
		n = max - 1
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = f.headerByte(pos + 1 + i)
	}
	return string(b), true
}

// CodingScheme returns the coding scheme as it stands in the header, and
// whether the header is long enough to hold one.
func (f *File) CodingScheme() (string, bool) {
	return f.headerString(8, 40)
}

// Family returns the family as it stands in the header, and whether the
// header is long enough to hold one.
func (f *File) Family() (string, bool) {
	return f.headerString(48, 20)
}

// Face returns the face byte, and whether the header is long enough to hold
// one.
func (f *File) Face() (byte, bool) {
	return f.headerByte(71), len(f.Header) >= 18
}

// SevenBitSafe reports whether the header has the seven-bit-safe flag set.
func (f *File) SevenBitSafe() bool {
	return len(f.Header) >= 18 && f.headerByte(68) > 127
}

// Exists reports whether the font has the character c.
func (f *File) Exists(c int) bool {
	return c >= f.First && c <= f.Last && c-f.First < len(f.Chars) && f.Chars[c-f.First].WidthIndex != 0
}

// info returns the word of character c, or the zero word outside the range
// of the character codes.
func (f *File) info(c int) CharInfo {
	if c < f.First || c > f.Last || c-f.First >= len(f.Chars) {
		return CharInfo{}
	}
	return f.Chars[c-f.First]
}

func entry(t []Fix, i uint8) Fix {
	if int(i) >= len(t) {
		return 0
	}
	return t[i]
}

// CharWidth returns the width of character c, or zero if c does not exist.
func (f *File) CharWidth(c int) Fix {
	return entry(f.Width, f.info(c).WidthIndex)
}

// CharHeight returns the height of character c, or zero if c does not exist.
func (f *File) CharHeight(c int) Fix {
	return entry(f.Height, f.info(c).HeightIndex)
}

// CharDepth returns the depth of character c, or zero if c does not exist.
func (f *File) CharDepth(c int) Fix {
	return entry(f.Depth, f.info(c).DepthIndex)
}

// CharItalic returns the italic correction of character c, or zero if c does
// not exist.
func (f *File) CharItalic(c int) Fix {
	return entry(f.Italic, f.info(c).ItalicIndex)
}

// NextLarger returns the next larger character of c, if c has one.
func (f *File) NextLarger(c int) (int, bool) {
	ci := f.info(c)
	if ci.Tag != ListTag {
		return 0, false
	}
	return int(ci.Remainder), true
}

// Recipe returns the extensible recipe of c, if c has one.
func (f *File) Recipe(c int) (Recipe, bool) {
	ci := f.info(c)
	if ci.Tag != ExtTag || int(ci.Remainder) >= len(f.Exten) {
		return Recipe{}, false
	}
	return f.Exten[ci.Remainder], true
}

// BoundaryChar returns the boundary character of the font, if it has one.
func (f *File) BoundaryChar() (int, bool) {
	if len(f.LigKern) == 0 || f.LigKern[0].Skip != 255 {
		return 0, false
	}
	return int(f.LigKern[0].Next), true
}

// programStart returns the index of the first step of the program of c,
// where c is a character code or Boundary.
func (f *File) programStart(c int) (int, bool) {
	nl := len(f.LigKern)
	if c == Boundary {
		if nl == 0 || f.LigKern[nl-1].Skip != 255 {
			return 0, false
		}
		i := int(f.LigKern[nl-1].Op)<<8 | int(f.LigKern[nl-1].Remainder)
		return i, i < nl
	}
	ci := f.info(c)
	if ci.Tag != LigTag {
		return 0, false
	}
	i := int(ci.Remainder)
	if i < nl && f.LigKern[i].Skip > 128 {
		i = int(f.LigKern[i].Op)<<8 | int(f.LigKern[i].Remainder)
	}
	return i, i < nl
}

// step converts a word of the program into a step.
func (f *File) step(s LigKern) Step {
	if s.Op >= 128 {
		st := Step{Next: s.Next, Op: OpKern}
		if i := int(s.Op-128)<<8 | int(s.Remainder); i < len(f.Kern) {
			st.Kern = f.Kern[i]
		}
		return st
	}
	return Step{Next: s.Next, Op: standardOp(s.Op), Result: s.Remainder}
}

// Program returns the steps of the ligature and kerning program of c in the
// order in which TeX tries them, where c is a character code or Boundary.  A
// ligature operation which is not one of the eight operations of the format
// is returned as OpLig.
func (f *File) Program(c int) []Step {
	i, ok := f.programStart(c)
	if !ok {
		return nil
	}
	var prog []Step
	for i < len(f.LigKern) {
		s := f.LigKern[i]
		if s.Skip > 128 {
			break
		}
		prog = append(prog, f.step(s))
		if s.Skip == 128 {
			break
		}
		i += int(s.Skip) + 1
	}
	return prog
}

// Step runs the ligature and kerning program of the character left against
// the character right, and returns the step which TeX applies to the pair.
// Left is a character code or Boundary, for the program of the left boundary;
// right is a character code or Boundary, for the boundary character.
func (f *File) Step(left, right int) (Step, bool) {
	if right == Boundary {
		var ok bool
		if right, ok = f.BoundaryChar(); !ok {
			return Step{}, false
		}
	}
	i, ok := f.programStart(left)
	if !ok {
		return Step{}, false
	}
	for i < len(f.LigKern) {
		s := f.LigKern[i]
		if s.Skip > 128 {
			break
		}
		if int(s.Next) == right {
			return f.step(s), true
		}
		if s.Skip == 128 {
			break
		}
		i += int(s.Skip) + 1
	}
	return Step{}, false
}

// KernBetween returns the kern which TeX inserts between the characters left
// and right, or zero if the pair has no kern step.  The arguments are as for
// [File.Step].
func (f *File) KernBetween(left, right int) Fix {
	if s, ok := f.Step(left, right); ok && s.Op == OpKern {
		return s.Kern
	}
	return 0
}

// An Item is one item of the horizontal list which [File.Typeset] returns: a
// character or a kern.
type Item struct {
	Char int // the character code, or -1 for a kern
	Kern Fix // the amount of a kern
}

// Typeset returns the horizontal list which TeX makes of the characters
// codes in the font, as the main loop of TeX does: it looks at each pair of
// characters in the order of their program, inserts a kern or forms a
// ligature, and does not look back at a kern once it has inserted it.  The
// characters must exist and the font must have no boundary character, and a
// ligature must use one of the operations LIG, LIG/, /LIG and /LIG/.
func (f *File) Typeset(codes []int) ([]Item, error) {
	var out []Item
	if len(codes) == 0 {
		return out, nil
	}
	l, next := codes[0], 1
	for next < len(codes) {
		r := codes[next]
		next++
	pair:
		for steps := 0; ; steps++ {
			if steps > 100 {
				return nil, fmt.Errorf("tfm: the ligatures of %d and %d do not end", l, r)
			}
			s, ok := f.Step(l, r)
			switch {
			case !ok:
				out = append(out, Item{Char: l})
				l = r
				break pair
			case s.Op == OpKern:
				out = append(out, Item{Char: l}, Item{Char: -1, Kern: s.Kern})
				l = r
				break pair
			}
			switch s.Op {
			case OpLig: // =:
				l = int(s.Result)
				if next == len(codes) {
					return append(out, Item{Char: l}), nil
				}
				r = codes[next]
				next++
			case OpLigSlash: // =:|
				l = int(s.Result)
			case OpSlashLig: // |=:
				r = int(s.Result)
			case OpSlashLigSlash: // |=:|
				out = append(out, Item{Char: l})
				l = int(s.Result)
			default:
				return nil, fmt.Errorf("tfm: the operation %s is not supported", s.Op)
			}
		}
	}
	return append(out, Item{Char: l}), nil
}
