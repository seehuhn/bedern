// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"io"
	"strconv"
)

// plWriter prints nested property lists in the layout of TFtoPL.  With a
// destination it passes the text on in pieces.
type plWriter struct {
	buf      []byte
	level    int
	fontType int
	dst      io.Writer
	err      error
}

// flush passes the text so far on to the destination.
func (w *plWriter) flush() {
	if w.dst == nil {
		return
	}
	if w.err == nil {
		_, w.err = w.dst.Write(w.buf)
	}
	w.buf = w.buf[:0]
}

func (w *plWriter) out(s string) {
	w.buf = append(w.buf, s...)
}

// outLn ends a line and indents the next one.
func (w *plWriter) outLn() {
	w.buf = append(w.buf, '\n')
	if len(w.buf) >= 1<<16 {
		w.flush()
	}
	for i := 0; i < w.level; i++ {
		w.buf = append(w.buf, "   "...)
	}
}

func (w *plWriter) left() {
	w.level++
	w.buf = append(w.buf, '(')
}

func (w *plWriter) right() {
	w.level--
	w.buf = append(w.buf, ')')
	w.outLn()
}

func (w *plWriter) octal(v uint32) {
	w.out(" O ")
	w.buf = strconv.AppendUint(w.buf, uint64(v), 8)
}

func (w *plWriter) char(c uint8) {
	letter := c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
	if w.fontType != vanilla || !letter {
		w.octal(uint32(c))
		return
	}
	w.out(" C ")
	w.buf = append(w.buf, c)
}

func (w *plWriter) fix(x Fix) {
	w.out(" R ")
	w.buf = appendFix(w.buf, x)
}

// face prints the property FACE.
func (w *plWriter) face(face uint8) {
	w.left()
	w.out("FACE")
	if face >= 18 {
		w.octal(uint32(face))
	} else {
		w.out(" F ")
		w.buf = append(w.buf, "MBL"[face/2%3], "RI"[face%2], "RCE"[face/2/3])
	}
	w.right()
}

// prop prints a property with a fix word as its value.
func (w *plWriter) prop(name string, x Fix) {
	w.left()
	w.out(name)
	w.fix(x)
	w.right()
}

func (w *plWriter) step(s Step) {
	w.left()
	w.out(s.Op.String())
	w.char(s.Next)
	if s.Op == OpKern {
		w.fix(s.Kern)
	} else {
		w.char(s.Result)
	}
	w.right()
}

var paramNames = [...]string{"", "SLANT", "SPACE", "STRETCH", "SHRINK", "XHEIGHT", "QUAD", "EXTRASPACE"}

var mathSyNames = [...]string{"NUM1", "NUM2", "NUM3", "DENOM1", "DENOM2", "SUP1", "SUP2", "SUP3", "SUB1", "SUB2",
	"SUPDROP", "SUBDROP", "DELIM1", "DELIM2", "AXISHEIGHT"}

var mathExNames = [...]string{"DEFAULTRULETHICKNESS", "BIGOPSPACING1", "BIGOPSPACING2", "BIGOPSPACING3",
	"BIGOPSPACING4", "BIGOPSPACING5"}

// paramName returns the name of parameter i in a font of the given type, or
// the empty string if the parameter has a number only.
func paramName(i, fontType int) string {
	switch {
	case i <= 7:
		return paramNames[i]
	case i <= 22 && fontType == mathSy:
		return mathSyNames[i-8]
	case i <= 13 && fontType == mathEx:
		return mathExNames[i-8]
	}
	return ""
}

// WritePL writes the PL form of the file to w, character for character as
// TFtoPL prints it with its default options: character codes of letters and
// digits in the form C, other codes in octal, in a font of mathematical
// symbols or extension characters all codes in octal.
//
// Where TFtoPL changes the data of a file which TeX82 accepts, the output has
// the changed data and ends with the comment of TFtoPL about it; see
// [FindingKind].  For a file with an infinite loop of ligatures the output
// ends after the ligature table, as that of TFtoPL does.
//
// WritePL returns an [*InvalidFontError], and writes nothing, if the tables
// break a rule which [Read] checks.
func (f *File) WritePL(w io.Writer) error {
	if err := f.validate(); err != nil {
		return err
	}
	p := &plWriter{dst: w}
	f.pl(p)
	p.flush()
	return p.err
}

// pl prints the PL form of a valid file.
func (f *File) pl(w *plWriter) {
	v := f.view()
	w.fontType = v.fontType
	lh := len(f.Header)

	// the header
	if lh >= 12 {
		if lh >= 17 {
			w.left()
			w.out("FAMILY ")
			w.out(v.family)
			w.right()
			if lh >= 18 {
				w.face(f.headerByte(71))
				for i := 18; i < lh; i++ {
					w.left()
					w.out("HEADER D ")
					w.buf = strconv.AppendInt(w.buf, int64(i), 10)
					w.octal(f.Header[i])
					w.right()
				}
			}
		}
		w.left()
		w.out("CODINGSCHEME ")
		w.out(v.scheme)
		w.right()
	}
	w.prop("DESIGNSIZE", f.DesignSize())
	w.out("(COMMENT DESIGNSIZE IS IN POINTS)")
	w.outLn()
	w.out("(COMMENT OTHER SIZES ARE MULTIPLES OF DESIGNSIZE)")
	w.outLn()
	w.left()
	w.out("CHECKSUM")
	w.octal(f.Checksum())
	w.right()
	if f.SevenBitSafe() {
		w.left()
		w.out("SEVENBITSAFEFLAG TRUE")
		w.right()
	}

	// the parameters
	if len(f.Param) > 0 {
		w.left()
		w.out("FONTDIMEN")
		w.outLn()
		for i, x := range f.Param {
			name := paramName(i+1, v.fontType)
			if name == "" {
				name = "PARAMETER D " + strconv.Itoa(i+1)
			}
			w.prop(name, x)
		}
		w.right()
	}

	// the ligature and kerning program
	if len(f.LigKern) > 0 {
		if v.list.bchar >= 0 {
			w.left()
			w.out("BOUNDARYCHAR")
			w.char(uint8(v.list.bchar))
			w.right()
		}
		w.left()
		w.out("LIGTABLE")
		w.outLn()
		for _, it := range v.list.items {
			switch it.kind {
			case itemUnused:
				w.left()
				w.out("COMMENT THIS PART OF THE PROGRAM IS NEVER USED!")
				w.outLn()
			case itemUsed:
				w.right()
			case itemLabel:
				w.left()
				w.out("LABEL")
				if it.char == Boundary {
					w.out(" BOUNDARYCHAR")
				} else {
					w.char(uint8(it.char))
				}
				w.right()
			case itemStep:
				w.step(it.step)
			case itemStop:
				w.out("(STOP)")
				w.outLn()
			case itemSkip:
				w.out("(SKIP D ")
				w.buf = strconv.AppendInt(w.buf, int64(it.n), 10)
				w.out(")")
				w.outLn()
			}
		}
		w.right()
		if v.hasLoop {
			if v.loop.cycle {
				w.out("(INFINITE LIGATURE LOOP MUST BE BROKEN!)")
			}
			return
		}
	}

	// the characters
	for i, ci := range f.Chars {
		if ci.WidthIndex == 0 {
			continue
		}
		c := f.First + i
		w.left()
		w.out("CHARACTER")
		w.char(uint8(c))
		w.outLn()
		w.prop("CHARWD", f.Width[ci.WidthIndex])
		if ci.HeightIndex > 0 {
			w.prop("CHARHT", f.Height[ci.HeightIndex])
		}
		if ci.DepthIndex > 0 {
			w.prop("CHARDP", f.Depth[ci.DepthIndex])
		}
		if ci.ItalicIndex > 0 {
			w.prop("CHARIC", f.Italic[ci.ItalicIndex])
		}
		switch ci.Tag {
		case LigTag:
			w.left()
			w.out("COMMENT")
			w.outLn()
			lig := v.list.lig
			for k := v.list.start(ci); k < len(lig); {
				s := lig[k]
				if s.Skip <= 128 {
					w.step(f.step(s))
				}
				if s.Skip >= 128 {
					break
				}
				k += int(s.Skip) + 1
			}
			w.right()
		case ListTag:
			if !v.noLarger[c] {
				w.left()
				w.out("NEXTLARGER")
				w.char(ci.Remainder)
				w.right()
			}
		case ExtTag:
			w.left()
			w.out("VARCHAR")
			w.outLn()
			r := f.Exten[ci.Remainder]
			for k, p := range [4]uint8{r.Top, r.Mid, r.Bot, r.Rep} {
				if k == 3 || p > 0 {
					w.left()
					w.out([4]string{"TOP", "MID", "BOT", "REP"}[k])
					w.char(p)
					w.right()
				}
			}
			w.right()
		}
		w.right()
	}
	for _, fd := range v.finds {
		switch fd.Kind {
		case StringTooLong, StringParenthesis, StringNotASCII, LargerMissing:
			w.out("(COMMENT THE TFM FILE WAS BAD, SO THE DATA HAS BEEN CHANGED!)")
			w.buf = append(w.buf, '\n')
			return
		}
	}
}

// plText reports whether a PL file can hold the string s as the value of a
// property which takes fewer than max characters.
func plText(s string, max int) bool {
	if !validString(s, max) || s != "" && s[0] == ' ' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] >= 'a' && s[i] <= 'z' {
			return false
		}
	}
	return true
}

// WritePL writes the description to w as a PL file, in the layout of TFtoPL.
// [ReadPL] reads the description back from the output.
//
// WritePL returns an [*InvalidFontError], and writes nothing, for a
// description which the PL language cannot hold: a coding scheme or family
// with a lower-case letter, a leading blank or a character which the header
// cannot hold, more than 255 parameters, more than 238 further header words,
// a fix word of -2048, an instruction or a tag which is not valid and unused
// values in a font without characters.
func (f *Font) WritePL(w io.Writer) error {
	switch {
	case !plText(f.CodingScheme, 40):
		return invalid("a PL file cannot hold the coding scheme")
	case !plText(f.Family, 20):
		return invalid("a PL file cannot hold the family")
	case len(f.Params) > 255:
		return invalid("a PL file cannot hold more than 255 parameters")
	case len(f.Header) > 238:
		return invalid("a PL file cannot hold more than 238 further header words")
	}
	fontType := vanilla
	if len(f.CodingScheme) >= 11 {
		switch f.CodingScheme[:11] {
		case "TEX MATH SY":
			fontType = mathSy
		case "TEX MATH EX":
			fontType = mathEx
		}
	}
	p := &plWriter{fontType: fontType}
	ok := true
	fix := func(name string, x Fix) {
		if x == minFix {
			ok = false
		}
		p.prop(name, x)
	}
	p.left()
	p.out("FAMILY ")
	p.out(f.Family)
	p.right()
	p.face(f.Face)
	for i, x := range f.Header {
		p.left()
		p.out("HEADER D ")
		p.buf = strconv.AppendInt(p.buf, int64(18+i), 10)
		p.octal(x)
		p.right()
	}
	p.left()
	p.out("CODINGSCHEME ")
	p.out(f.CodingScheme)
	p.right()
	fix("DESIGNSIZE", f.DesignSize)
	if f.DesignUnits != 0 {
		fix("DESIGNUNITS", f.DesignUnits)
	}
	if f.HasChecksum {
		p.left()
		p.out("CHECKSUM")
		p.octal(f.Checksum)
		p.right()
	}
	if f.SevenBitSafe {
		p.left()
		p.out("SEVENBITSAFEFLAG TRUE")
		p.right()
	}
	if len(f.Params) > 0 {
		p.left()
		p.out("FONTDIMEN")
		p.outLn()
		for i, x := range f.Params {
			name := paramName(i+1, fontType)
			if name == "" {
				name = "PARAMETER D " + strconv.Itoa(i+1)
			}
			fix(name, x)
		}
		p.right()
	}
	if f.HasBoundary {
		p.left()
		p.out("BOUNDARYCHAR")
		p.char(f.BoundaryChar)
		p.right()
	}
	if len(f.LigTable) > 0 {
		p.left()
		p.out("LIGTABLE")
		p.outLn()
		for _, in := range f.LigTable {
			switch in.Kind {
			case KindLabel:
				p.left()
				p.out("LABEL")
				p.char(in.Char)
				p.right()
			case KindBoundaryLabel:
				p.left()
				p.out("LABEL BOUNDARYCHAR")
				p.right()
			case KindStep:
				if in.Step.Op != OpKern && !in.Step.Op.IsLig() || in.Step.Kern == minFix {
					ok = false
				}
				p.step(in.Step)
			case KindStop:
				p.out("(STOP)")
				p.outLn()
			case KindSkip:
				p.out("(SKIP D ")
				p.buf = strconv.AppendInt(p.buf, int64(in.Skip), 10)
				p.out(")")
				p.outLn()
			default:
				ok = false
			}
		}
		p.right()
	}
	unused := f.Unused
	character := func(c int) {
		ch := f.Chars[c]
		p.left()
		p.out("CHARACTER")
		p.char(uint8(c))
		p.outLn()
		var replaced [4]bool
		for i, name := range [4]string{"CHARWD", "CHARHT", "CHARDP", "CHARIC"} {
			for _, x := range unused[i] {
				if i == 0 || x != 0 {
					fix(name, x)
					replaced[i] = true
				}
			}
			unused[i] = nil
		}
		fix("CHARWD", ch.Width)
		if ch.Height != 0 || replaced[1] {
			fix("CHARHT", ch.Height)
		}
		if ch.Depth != 0 || replaced[2] {
			fix("CHARDP", ch.Depth)
		}
		if ch.Italic != 0 || replaced[3] {
			fix("CHARIC", ch.Italic)
		}
		switch ch.Tag {
		case NoTag:
		case ListTag:
			p.left()
			p.out("NEXTLARGER")
			p.char(ch.NextLarger)
			p.right()
		case ExtTag:
			p.left()
			p.out("VARCHAR")
			p.outLn()
			r := ch.Recipe
			for k, piece := range [4]uint8{r.Top, r.Mid, r.Bot, r.Rep} {
				if k == 3 || piece > 0 {
					p.left()
					p.out([4]string{"TOP", "MID", "BOT", "REP"}[k])
					p.char(piece)
					p.right()
				}
			}
			p.right()
		default:
			ok = false
		}
		p.right()
	}
	first := [256]bool{}
	for _, c := range f.RecipeOrder {
		if f.Chars[c] == nil || first[c] {
			ok = false
			continue
		}
		first[c] = true
		character(int(c))
	}
	for c, ch := range f.Chars {
		if ch != nil && !first[c] {
			character(c)
		}
	}
	for _, u := range unused {
		if len(u) > 0 {
			ok = false
		}
	}
	if !ok {
		return invalid("a PL file cannot hold the description")
	}
	_, err := w.Write(p.buf)
	return err
}
