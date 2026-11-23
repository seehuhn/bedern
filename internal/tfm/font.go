// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"sort"
	"strings"
)

// Font describes a font as a PL file does.
//
// A dimension x of a character, a kern or a parameter stands for
// x/DesignUnits design sizes; with DesignUnits equal to Unity it is a fix word
// of the design size.
type Font struct {
	// Checksum is the check sum of the header if HasChecksum is set.
	// Without HasChecksum, [Font.Build] computes the check sum from the
	// widths.
	Checksum    uint32
	HasChecksum bool

	// DesignSize is the design size in points, at least 1.
	DesignSize Fix

	// DesignUnits is the number of units which make one design size.  Zero
	// stands for Unity.
	DesignUnits Fix

	// CodingScheme and Family are the strings of the header, of at most 39
	// and 19 printable ASCII characters without parentheses.
	CodingScheme string
	Family       string

	// Face is the face byte of the header.
	Face uint8

	// SevenBitSafe is the claim of a PL file that the font is seven-bit
	// safe.  [Font.Build] works the flag of the header out for itself.
	SevenBitSafe bool

	// Header holds the words of the header from word 18 on.
	Header []uint32

	// Params holds the font parameters; Params[0] is parameter 1, the slant,
	// which is a pure number.
	Params []Fix

	// BoundaryChar is the boundary character if HasBoundary is set.
	BoundaryChar uint8
	HasBoundary  bool

	// LigTable is the ligature and kerning program.
	LigTable []Instr

	// Chars holds the characters of the font; the entry of a character code
	// which the font does not have is nil.
	Chars [256]*Char

	// RecipeOrder lists the characters which have a recipe in the order in
	// which their recipes stand in the table of recipes.  Nil stands for the
	// order of the character codes.
	RecipeOrder []uint8

	// Unused holds values of the tables of widths, heights, depths and
	// italic corrections, in this order, which no character needs to have.
	// PLtoTF keeps in a table every value which a PL file gives to a
	// dimension of a character, also where another value replaces it.
	Unused [4][]Fix
}

// Char describes one character.
type Char struct {
	Width, Height, Depth, Italic Fix

	// Tag is NoTag, ListTag or ExtTag.  A character has a ligature and
	// kerning program through a label in Font.LigTable, not through its tag.
	Tag Tag

	// NextLarger is the next larger character of a character with ListTag.
	NextLarger uint8

	// Recipe is the extensible recipe of a character with ExtTag.
	Recipe Recipe
}

// InstrKind is the kind of an instruction of a ligature and kerning program.
type InstrKind uint8

// The kinds of an instruction.
const (
	KindLabel         InstrKind = iota // the program of Instr.Char starts here
	KindBoundaryLabel                  // the program of the left boundary starts here
	KindStep                           // the ligature or kern step Instr.Step
	KindStop                           // the program ends after the step before
	KindSkip                           // the program passes over Instr.Skip steps after the step before
)

// Instr is one instruction of the ligature and kerning program of a [Font].
// A program runs from its label through the steps which follow; a step
// without a stop or skip behind it is followed by the next step.
type Instr struct {
	Kind InstrKind
	Char uint8 // the character of a label
	Skip uint8 // the number of steps of a skip, at most 127
	Step Step
}

// Label returns the instruction which starts the program of character c.
func Label(c uint8) Instr {
	return Instr{Kind: KindLabel, Char: c}
}

// BoundaryLabel returns the instruction which starts the program of the left
// boundary.
func BoundaryLabel() Instr {
	return Instr{Kind: KindBoundaryLabel}
}

// Lig returns a ligature step: if next follows the current character, the
// ligature operation op inserts the character result.
func Lig(op Op, next, result uint8) Instr {
	return Instr{Kind: KindStep, Step: Step{Next: next, Op: op, Result: result}}
}

// Krn returns a kern step: if next follows the current character, a kern of
// the given amount is inserted between the two.
func Krn(next uint8, amount Fix) Instr {
	return Instr{Kind: KindStep, Step: Step{Next: next, Op: OpKern, Kern: amount}}
}

// Stop returns the instruction which ends a program.
func Stop() Instr {
	return Instr{Kind: KindStop}
}

// Skip returns the instruction which passes over n steps.
func Skip(n uint8) Instr {
	return Instr{Kind: KindSkip, Skip: n}
}

// NewFont returns a description with the values which PLtoTF gives to a
// property that a PL file does not mention: a design size of 10 points and
// the coding scheme and family UNSPECIFIED.
func NewFont() *Font {
	return &Font{
		DesignSize:   10 * Unity,
		DesignUnits:  Unity,
		CodingScheme: "UNSPECIFIED",
		Family:       "UNSPECIFIED",
	}
}

// Programs returns a ligature and kerning program in which every character
// has the given steps.  The key of the map is a character code or Boundary,
// for the program of the left boundary.  Characters with equal steps share
// one copy of them, every program ends with a stop and the programs stand in
// the order of the smallest code which uses them.  A character without steps
// gets no program.
func Programs(progs map[int][]Step) []Instr {
	keys := make([]int, 0, len(progs))
	for c, p := range progs {
		if len(p) > 0 && c >= 0 && c <= Boundary {
			keys = append(keys, c)
		}
	}
	sort.Ints(keys)
	type group struct {
		labels []int
		steps  []Step
	}
	var groups []*group
	byText := make(map[string]*group)
	var sb strings.Builder
	for _, c := range keys {
		sb.Reset()
		for _, s := range progs[c] {
			sb.Write([]byte{s.Next, byte(s.Op), s.Result,
				byte(uint32(s.Kern) >> 24), byte(uint32(s.Kern) >> 16), byte(uint32(s.Kern) >> 8), byte(s.Kern)})
		}
		g := byText[sb.String()]
		if g == nil {
			g = &group{steps: progs[c]}
			byText[sb.String()] = g
			groups = append(groups, g)
		}
		if c == Boundary {
			g.labels = append([]int{c}, g.labels...)
		} else {
			g.labels = append(g.labels, c)
		}
	}
	var table []Instr
	for _, g := range groups {
		for _, c := range g.labels {
			if c == Boundary {
				table = append(table, BoundaryLabel())
			} else {
				table = append(table, Label(uint8(c)))
			}
		}
		for _, s := range g.steps {
			table = append(table, Instr{Kind: KindStep, Step: s})
		}
		table = append(table, Stop())
	}
	return table
}

// Program returns the steps of the ligature and kerning program of c in the
// order in which TeX tries them, where c is a character code or Boundary.
func (f *Font) Program(c int) []Step {
	p, err := f.assemble()
	if err != nil {
		return nil
	}
	i := -1
	if c == Boundary {
		if p.bcharLabel < noLabel {
			i = p.bcharLabel
		}
	} else if c >= 0 && c < 256 && p.tag[c] == LigTag {
		i = p.rem[c]
	}
	if i < 0 {
		return nil
	}
	var prog []Step
	for i < len(p.lig) {
		s := p.lig[i]
		if s.Skip > 128 {
			break
		}
		st := Step{Next: s.Next, Op: Op(s.Op), Result: s.Remainder}
		if s.Op >= 128 {
			st = Step{Next: s.Next, Op: OpKern, Kern: p.kern[int(s.Op-128)<<8|int(s.Remainder)]}
		}
		prog = append(prog, st)
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
func (f *Font) Step(left, right int) (Step, bool) {
	if right == Boundary {
		if !f.HasBoundary {
			return Step{}, false
		}
		right = int(f.BoundaryChar)
	}
	for _, s := range f.Program(left) {
		if int(s.Next) == right {
			return s, true
		}
	}
	return Step{}, false
}

// plString returns a string of the header as TFtoPL prints it, and the
// findings of TFtoPL about it.
func (f *File) plString(pos, max int, name string) (string, []Finding) {
	var finds []Finding
	n := int(f.headerByte(pos))
	if n >= max {
		finds = append(finds, Finding{StringTooLong, "the " + name + " is longer than its field"})
		n = 1
	}
	b := make([]byte, n)
	for i := range b {
		c := f.headerByte(pos + 1 + i)
		switch {
		case c == '(' || c == ')':
			finds = append(finds, Finding{StringParenthesis, "the " + name + " holds a parenthesis"})
			c = '/'
		case c < ' ' || c > '~':
			finds = append(finds, Finding{StringNotASCII, "the " + name + " holds a byte which is no printable ASCII character"})
			c = '?'
		case c >= 'a' && c <= 'z':
			c -= 'a' - 'A'
		}
		b[i] = c
	}
	return string(b), finds
}

// The kinds of font for which TFtoPL names the parameters beyond the seventh.
const (
	vanilla = iota
	mathSy
	mathEx
)

// view holds what TFtoPL makes of a valid file.
type view struct {
	scheme, family string
	fontType       int
	finds          []Finding
	list           *listing
	loop           *loopChecker
	hasLoop        bool
	noLarger       [256]bool // the next larger character does not exist
}

// view analyses a valid file as TFtoPL does.
func (f *File) view() *view {
	v := &view{}
	if len(f.Trailer) > 0 {
		v.finds = append(v.finds, Finding{TrailingBytes, "bytes follow the stated length of the file"})
	}
	lh := len(f.Header)
	if lh >= 12 {
		var fs []Finding
		v.scheme, fs = f.plString(8, 40, "coding scheme")
		v.finds = append(v.finds, fs...)
		if strings.HasPrefix(v.scheme, "TEX MATH SY") {
			v.fontType = mathSy
		} else if strings.HasPrefix(v.scheme, "TEX MATH EX") {
			v.fontType = mathEx
		}
		if lh >= 17 {
			v.family, fs = f.plString(48, 20, "family")
			v.finds = append(v.finds, fs...)
		}
	}
	np := len(f.Param)
	if v.fontType == mathSy && np != 22 {
		v.finds = append(v.finds, Finding{ParamCount, "a font of mathematical symbols without 22 parameters"})
	} else if v.fontType == mathEx && np != 13 {
		v.finds = append(v.finds, Finding{ParamCount, "a font of mathematical extension characters without 13 parameters"})
	}
	v.list = f.list()
	if v.list.oddOps > 0 {
		v.finds = append(v.finds, Finding{NonstandardLig, "a ligature step has an operation which is none of the eight"})
	}
	if v.list.unused > 0 {
		v.finds = append(v.finds, Finding{UnusedSteps, "no program reaches some steps of the ligature and kerning program"})
	}
	if len(f.LigKern) > 0 {
		v.loop = f.loops(v.list)
		if v.loop.full() {
			v.finds = append(v.finds, Finding{TooManyPairs, "the ligature and kerning program has more pairs than TFtoPL holds"})
			v.hasLoop = true
			return v
		}
		if v.loop.run() {
			v.finds = append(v.finds, Finding{LigatureLoop, "a pair of characters starts an infinite loop of ligatures"})
			v.hasLoop = true
			return v
		}
	}
	for i, ci := range f.Chars {
		if ci.WidthIndex != 0 && ci.Tag == ListTag && !f.Exists(int(ci.Remainder)) {
			v.noLarger[f.First+i] = true
			v.finds = append(v.finds, Finding{LargerMissing, "the next larger character of a character does not exist"})
		}
	}
	return v
}

// Font returns the description of f which TFtoPL gives.  The description
// leaves out what a PL file does not say: the steps of the ligature and
// kerning program which no program reaches, the entries of the tables which
// no character uses, the order of the tables and the tag of a character code
// without a character unless it is the start of a program.  The strings of
// the header are in upper case, without leading spaces, with a slash for a
// parenthesis and a question mark for a byte which is no printable ASCII
// character; a header too short for a string gives UNSPECIFIED.  A ligature
// operation which is none of the eight is OpLig, and a next larger character
// which does not exist is left out.  The file must be one which [Read] or
// [Font.Build] returned or one which [File.Write] accepts.
func (f *File) Font() *Font {
	v := f.view()
	d := &Font{
		Checksum:     f.Checksum(),
		HasChecksum:  true,
		DesignSize:   f.DesignSize(),
		DesignUnits:  Unity,
		CodingScheme: "UNSPECIFIED",
		Family:       "UNSPECIFIED",
	}
	lh := len(f.Header)
	if lh >= 12 {
		d.CodingScheme = strings.TrimLeft(v.scheme, " ")
	}
	if lh >= 17 {
		d.Family = strings.TrimLeft(v.family, " ")
	}
	if lh >= 18 {
		d.Face = f.headerByte(71)
		d.SevenBitSafe = f.headerByte(68) > 127
		d.Header = append([]uint32(nil), f.Header[18:]...)
	}
	d.Params = append([]Fix(nil), f.Param...)
	if v.list.bchar >= 0 {
		d.HasBoundary = true
		d.BoundaryChar = uint8(v.list.bchar)
	}
	unused := false
	for _, it := range v.list.items {
		switch it.kind {
		case itemUnused:
			unused = true
		case itemUsed:
			unused = false
		}
		if unused {
			continue
		}
		switch it.kind {
		case itemLabel:
			if it.char == Boundary {
				d.LigTable = append(d.LigTable, BoundaryLabel())
			} else {
				d.LigTable = append(d.LigTable, Label(uint8(it.char)))
			}
		case itemStep:
			d.LigTable = append(d.LigTable, Instr{Kind: KindStep, Step: it.step})
		case itemStop:
			d.LigTable = append(d.LigTable, Stop())
		case itemSkip:
			d.LigTable = append(d.LigTable, Skip(uint8(it.n)))
		}
	}
	for i, ci := range f.Chars {
		if ci.WidthIndex == 0 {
			continue
		}
		c := f.First + i
		ch := &Char{
			Width:  f.Width[ci.WidthIndex],
			Height: f.Height[ci.HeightIndex],
			Depth:  f.Depth[ci.DepthIndex],
			Italic: f.Italic[ci.ItalicIndex],
		}
		switch ci.Tag {
		case ListTag:
			if !v.noLarger[c] {
				ch.Tag = ListTag
				ch.NextLarger = ci.Remainder
			}
		case ExtTag:
			ch.Tag = ExtTag
			ch.Recipe = f.Exten[ci.Remainder]
		}
		d.Chars[c] = ch
	}
	return d
}
