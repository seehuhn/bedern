// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"io"
	"strconv"
)

// SyntaxError is the error which [ReadPL] returns for input which PLtoTF
// reports as faulty.
type SyntaxError struct {
	Line   int // the number of the line, counted from 1
	Reason string
}

func (err *SyntaxError) Error() string {
	return "tfm: PL file, line " + strconv.Itoa(err.Line) + ": " + err.Reason
}

// The codes of the property names.
const (
	codeComment = iota
	codeChecksum
	codeDesignSize
	codeDesignUnits
	codeCodingScheme
	codeFamily
	codeFace
	codeSevenBitSafeFlag
	codeHeader
	codeFontDimen
	codeLigTable
	codeBoundaryChar
	codeCharacter
	codeParameter  = 20
	codeCharWd     = 51
	codeCharHt     = 52
	codeCharDp     = 53
	codeCharIc     = 54
	codeNextLarger = 55
	codeVarChar    = 56
	codeLabel      = 70
	codeStop       = 71
	codeSkip       = 72
	codeKrn        = 73
	codeLig        = 74
	codeUnknown    = -1
)

var propertyNames = map[string]int{
	"CHECKSUM":         codeChecksum,
	"DESIGNSIZE":       codeDesignSize,
	"DESIGNUNITS":      codeDesignUnits,
	"CODINGSCHEME":     codeCodingScheme,
	"FAMILY":           codeFamily,
	"FACE":             codeFace,
	"SEVENBITSAFEFLAG": codeSevenBitSafeFlag,
	"HEADER":           codeHeader,
	"FONTDIMEN":        codeFontDimen,
	"LIGTABLE":         codeLigTable,
	"BOUNDARYCHAR":     codeBoundaryChar,
	"CHARACTER":        codeCharacter,
	"PARAMETER":        codeParameter,
	"CHARWD":           codeCharWd,
	"CHARHT":           codeCharHt,
	"CHARDP":           codeCharDp,
	"CHARIC":           codeCharIc,
	"NEXTLARGER":       codeNextLarger,
	"VARCHAR":          codeVarChar,
	"TOP":              codeVarChar + 1,
	"MID":              codeVarChar + 2,
	"BOT":              codeVarChar + 3,
	"REP":              codeVarChar + 4,
	"EXT":              codeVarChar + 4,
	"COMMENT":          codeComment,
	"LABEL":            codeLabel,
	"STOP":             codeStop,
	"SKIP":             codeSkip,
	"KRN":              codeKrn,
	"LIG":              codeLig + int(OpLig),
	"/LIG":             codeLig + int(OpSlashLig),
	"/LIG>":            codeLig + int(OpSlashLigOver),
	"LIG/":             codeLig + int(OpLigSlash),
	"LIG/>":            codeLig + int(OpLigSlashOver),
	"/LIG/":            codeLig + int(OpSlashLigSlash),
	"/LIG/>":           codeLig + int(OpSlashLigSlashOver),
	"/LIG/>>":          codeLig + int(OpSlashLigSlashOver2),
}

func init() {
	for i := 1; i <= 7; i++ {
		propertyNames[paramNames[i]] = codeParameter + i
	}
	for i, name := range mathSyNames {
		propertyNames[name] = codeParameter + 8 + i
	}
	for i, name := range mathExNames {
		propertyNames[name] = codeParameter + 8 + i
	}
}

// plReader scans a PL file as PLtoTF does.
type plReader struct {
	data  []byte
	pos   int    // the start of the next line in data
	buf   []byte // the current line, followed by a blank
	loc   int    // the number of bytes of buf which are read
	line  int
	ended bool
	level int
	cur   byte
	err   *SyntaxError

	font      *Font
	tag       [256]Tag
	given     [4][256]bool // the dimension of the character has a value
	stepEnded bool
	recipes   []uint8
}

// fail records the first error and makes the rest of the input empty.
func (p *plReader) fail(reason string) {
	if p.err == nil {
		p.err = &SyntaxError{Line: max(p.line, 1), Reason: reason}
	}
	p.pos = len(p.data)
	p.buf = p.buf[:0]
	p.loc = 0
}

// fill reads the next line.  After the last line it supplies right
// parentheses without end.
func (p *plReader) fill() {
	if !p.ended {
		p.line++
	}
	p.loc = 0
	if p.pos >= len(p.data) {
		p.ended = true
		p.buf = append(p.buf[:0], ')')
		return
	}
	end := bytes.IndexByte(p.data[p.pos:], '\n')
	if end < 0 {
		end = len(p.data)
	} else {
		end += p.pos
	}
	text := p.data[p.pos:end]
	if i := bytes.IndexByte(text, '\r'); i >= 0 {
		text = text[:i]
	}
	p.pos = end + 1
	p.buf = append(append(p.buf[:0], text...), ' ')
	for p.loc < len(p.buf) && p.buf[p.loc] == ' ' {
		p.loc++
	}
}

// next reads a character into p.cur, in upper case.  A parenthesis is not
// passed.
func (p *plReader) next() {
	for p.loc == len(p.buf) {
		p.fill()
	}
	p.loc++
	c := p.buf[p.loc-1]
	switch {
	case c < ' ' || c > '~':
		p.fail("a character which is no printable ASCII character")
		c = ')'
	case c >= 'a' && c <= 'z':
		c -= 'a' - 'A'
	case c == '(' || c == ')':
		p.loc--
	}
	p.cur = c
}

// unread makes next read p.cur again.
func (p *plReader) unread() {
	if p.cur != '(' && p.cur != ')' && p.loc > 0 {
		p.loc--
	}
}

// keywordChar reads a character of a property name into p.cur, or sets p.cur
// to a blank if the next character is none.
func (p *plReader) keywordChar() {
	for p.loc == len(p.buf) && p.ended {
		p.fill()
	}
	p.cur = ' '
	if p.loc == len(p.buf) {
		return
	}
	c := p.buf[p.loc]
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	if c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c == '/' || c == '>' {
		p.loc++
		p.cur = c
	}
}

// skipItem passes the rest of the property list which is open.
func (p *plReader) skipItem() {
	l := p.level
	for p.level >= l {
		for p.loc == len(p.buf) {
			p.fill()
		}
		p.loc++
		switch p.buf[p.loc-1] {
		case ')':
			p.level--
		case '(':
			p.level++
		}
	}
	if p.ended {
		p.fail("the file ends inside a property list")
	}
	p.cur = ' '
}

func (p *plReader) skipBlanks() {
	for p.cur == ' ' {
		p.next()
	}
}

// skipToParen reads up to the next parenthesis.
func (p *plReader) skipToParen() {
	for {
		p.next()
		if p.cur == '(' || p.cur == ')' {
			return
		}
	}
}

// finish passes the right parenthesis which ends a property.
func (p *plReader) finish() {
	p.skipBlanks()
	if p.cur != ')' {
		p.fail("text follows the value of a property")
	}
	p.skipItem()
}

// finishInner puts back the right parenthesis which ends a property list.
func (p *plReader) finishInner() {
	if p.loc > 0 {
		p.loc--
	}
	p.level++
	p.cur = ')'
}

// name passes a left parenthesis and returns the code of the property name
// which follows.
func (p *plReader) name() int {
	p.loc++
	p.level++
	p.cur = ' '
	p.skipBlanks()
	p.unread()
	var name []byte
	p.keywordChar()
	for p.cur != ' ' {
		if len(name) <= 20 {
			name = append(name, p.cur)
		}
		p.keywordChar()
	}
	code, ok := propertyNames[string(name)]
	if !ok {
		p.fail("a property name which PLtoTF does not have")
		return codeUnknown
	}
	return code
}

// nextNonBlank reads the next character which is no blank.
func (p *plReader) nextNonBlank() {
	for {
		p.next()
		if p.cur != ' ' {
			return
		}
	}
}

func hexDigit(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	}
	return 0, false
}

// byteValue reads a value of one byte in the form C, D, O, H or F.
func (p *plReader) byteValue() uint8 {
	p.nextNonBlank()
	t := p.cur
	p.nextNonBlank()
	acc := 0
	switch t {
	case 'C':
		if p.cur < '!' || p.cur > '~' || p.cur == '(' || p.cur == ')' || p.err != nil {
			p.fail("a C value is a printable ASCII character and no parenthesis")
			break
		}
		acc = int(p.buf[p.loc-1])
	case 'D', 'O', 'H':
		radix := map[byte]int{'D': 10, 'O': 8, 'H': 16}[t]
		for {
			d, ok := hexDigit(p.cur)
			if !ok || d >= radix {
				break
			}
			acc = acc*radix + d
			if acc > 255 {
				p.fail("a value of more than 255 where one byte is expected")
				return 0
			}
			p.next()
		}
		p.unread()
	case 'F':
		switch p.cur {
		case 'B':
			acc = 2
		case 'L':
			acc = 4
		case 'M':
		default:
			acc = 18
		}
		p.next()
		switch p.cur {
		case 'I':
			acc++
		case 'R':
		default:
			acc = 18
		}
		p.next()
		switch p.cur {
		case 'C':
			acc += 6
		case 'E':
			acc += 12
		case 'R':
		default:
			acc = 18
		}
		if acc >= 18 {
			p.fail("a face code which is none of the eighteen")
			return 0
		}
	default:
		p.fail("a value of one byte has the form C, D, O, H or F")
		return 0
	}
	p.cur = ' '
	return uint8(acc)
}

// fourBytes reads a value of four bytes in the form O or H.
func (p *plReader) fourBytes() uint32 {
	p.nextNonBlank()
	var radix uint64
	switch p.cur {
	case 'H':
		radix = 16
	case 'O':
		radix = 8
	default:
		p.fail("a value of four bytes has the form O or H")
		return 0
	}
	p.nextNonBlank()
	var v uint64
	for {
		d, ok := hexDigit(p.cur)
		if !ok {
			break
		}
		if v*radix > 0xFFFFFFFF {
			p.fail("a value of more than four bytes")
			return 0
		}
		if uint64(d) >= radix {
			p.fail("a digit which the octal form does not have")
			return 0
		}
		v = v*radix + uint64(d)
		p.next()
	}
	return uint32(v)
}

// fixValue reads a real value in the form R or D and returns the nearest
// fix word.
func (p *plReader) fixValue() Fix {
	p.nextNonBlank()
	if p.cur != 'R' && p.cur != 'D' {
		p.fail("a real value has the form R or D")
		return 0
	}
	negative := false
	for {
		p.next()
		if p.cur == '-' {
			p.cur = ' '
			negative = !negative
		} else if p.cur == '+' {
			p.cur = ' '
		}
		if p.cur != ' ' {
			break
		}
	}
	acc := 0
	for p.cur >= '0' && p.cur <= '9' {
		acc = acc*10 + int(p.cur-'0')
		if acc >= 2048 {
			p.fail("a real value of 2048 or more")
			return 0
		}
		p.next()
	}
	intPart := acc
	acc = 0
	if p.cur == '.' {
		var digits [7]int
		j := 0
		p.next()
		for p.cur >= '0' && p.cur <= '9' {
			if j < 7 {
				digits[j] = 0x200000 * int(p.cur-'0')
				j++
			}
			p.next()
		}
		for j > 0 {
			j--
			acc = digits[j] + acc/10
		}
		acc = (acc + 10) / 20
	}
	if acc >= int(Unity) && intPart == 2047 {
		p.fail("a real value of 2048 or more")
		return 0
	}
	acc += intPart * int(Unity)
	if negative {
		acc = -acc
	}
	return Fix(acc)
}

// stringValue reads a string of fewer than n characters.
func (p *plReader) stringValue(n int) string {
	var b []byte
	p.skipBlanks()
	for p.cur != '(' && p.cur != ')' {
		if len(b) >= n-1 {
			p.fail("a string of more than " + strconv.Itoa(n-1) + " characters")
			return ""
		}
		b = append(b, p.cur)
		p.next()
	}
	return string(b)
}

// list reads the properties of a property list with the function item, up to
// the right parenthesis which ends the list.
func (p *plReader) list(level int, item func(code int)) {
	for p.level == level && p.err == nil {
		p.skipBlanks()
		switch p.cur {
		case '(':
			code := p.name()
			switch code {
			case codeUnknown:
			case codeComment:
				p.skipItem()
			default:
				item(code)
			}
		case ')':
			p.skipItem()
		default:
			p.fail("text which is not in parentheses")
		}
	}
	p.finishInner()
}

// setTag gives character c a tag, if it has none.
func (p *plReader) setTag(c uint8, tag Tag) bool {
	switch p.tag[c] {
	case LigTag:
		p.fail("the character has a label in the ligature table already")
	case ListTag:
		p.fail("the character has a next larger character already")
	case ExtTag:
		p.fail("the character has a recipe already")
	default:
		p.tag[c] = tag
		return true
	}
	return false
}

func (p *plReader) char(c uint8) *Char {
	if p.font.Chars[c] == nil {
		p.font.Chars[c] = &Char{}
	}
	return p.font.Chars[c]
}

// property reads one property of the outer level.
func (p *plReader) property() {
	f := p.font
	code := p.name()
	switch {
	case code == codeUnknown:
		return
	case code == codeComment:
		p.skipItem()
		return
	case code > codeCharacter:
		p.fail("a property which does not belong to the outer level")
		return
	}
	switch code {
	case codeChecksum:
		f.HasChecksum = true
		f.Checksum = p.fourBytes()
	case codeDesignSize:
		x := p.fixValue()
		if x < Unity && p.err == nil {
			p.fail("a design size of less than 1")
		}
		f.DesignSize = x
	case codeDesignUnits:
		x := p.fixValue()
		if x <= 0 && p.err == nil {
			p.fail("a number of design units which is not positive")
		}
		f.DesignUnits = x
	case codeCodingScheme:
		f.CodingScheme = p.stringValue(40)
	case codeFamily:
		f.Family = p.stringValue(20)
	case codeFace:
		f.Face = p.byteValue()
	case codeSevenBitSafeFlag:
		p.skipBlanks()
		switch p.cur {
		case 'T':
			f.SevenBitSafe = true
		case 'F':
			f.SevenBitSafe = false
		default:
			p.fail("a flag which is neither TRUE nor FALSE")
			return
		}
		p.skipToParen()
	case codeHeader:
		c := int(p.byteValue())
		if c < 18 {
			p.fail("a header index of less than 18")
			return
		}
		for len(f.Header) < c-17 {
			f.Header = append(f.Header, 0)
		}
		f.Header[c-18] = p.fourBytes()
	case codeFontDimen:
		p.list(1, p.parameter)
	case codeLigTable:
		p.stepEnded = false
		p.list(1, p.instruction)
	case codeBoundaryChar:
		f.HasBoundary = true
		f.BoundaryChar = p.byteValue()
	case codeCharacter:
		c := p.byteValue()
		if p.err != nil {
			return
		}
		p.char(c)
		p.list(1, func(code int) { p.charProperty(c, code) })
		p.given[0][c] = true
	}
	p.finish()
}

// parameter reads one property of a FONTDIMEN list.
func (p *plReader) parameter(code int) {
	if code < codeParameter || code >= codeCharWd {
		p.fail("a property which does not belong to a FONTDIMEN list")
		return
	}
	c := code - codeParameter
	if code == codeParameter {
		c = int(p.byteValue())
	}
	if c == 0 {
		p.fail("a parameter with the number 0")
		return
	}
	f := p.font
	for len(f.Params) < c {
		f.Params = append(f.Params, 0)
	}
	f.Params[c-1] = p.fixValue()
	p.finish()
}

// instruction reads one property of a LIGTABLE list.
func (p *plReader) instruction(code int) {
	if code < codeLabel {
		p.fail("a property which does not belong to a LIGTABLE list")
		return
	}
	f := p.font
	switch code {
	case codeLabel:
		p.skipBlanks()
		if p.cur == 'B' {
			f.LigTable = append(f.LigTable, BoundaryLabel())
			p.skipToParen()
		} else {
			p.unread()
			c := p.byteValue()
			if p.err != nil || !p.setTag(c, LigTag) {
				return
			}
			f.LigTable = append(f.LigTable, Label(c))
		}
		p.stepEnded = false
	case codeStop:
		if !p.stepEnded {
			p.fail("a STOP which does not follow a LIG or KRN")
			return
		}
		f.LigTable = append(f.LigTable, Stop())
		p.stepEnded = false
	case codeSkip:
		if !p.stepEnded {
			p.fail("a SKIP which does not follow a LIG or KRN")
			return
		}
		c := p.byteValue()
		if c >= 128 {
			p.fail("a SKIP of more than 127 steps")
			return
		}
		f.LigTable = append(f.LigTable, Skip(c))
		p.stepEnded = false
	case codeKrn:
		c := p.byteValue()
		f.LigTable = append(f.LigTable, Krn(c, p.fixValue()))
		p.stepEnded = true
	default:
		c := p.byteValue()
		f.LigTable = append(f.LigTable, Lig(Op(code-codeLig), c, p.byteValue()))
		p.stepEnded = true
	}
	p.finish()
}

// charProperty reads one property of the CHARACTER list of c.
func (p *plReader) charProperty(c uint8, code int) {
	if code < codeCharWd || code > codeVarChar {
		p.fail("a property which does not belong to a CHARACTER list")
		return
	}
	ch := p.char(c)
	switch code {
	case codeCharWd, codeCharHt, codeCharDp, codeCharIc:
		i := code - codeCharWd
		dim := [4]*Fix{&ch.Width, &ch.Height, &ch.Depth, &ch.Italic}[i]
		if p.given[i][c] && (i == 0 || *dim != 0) {
			p.font.Unused[i] = append(p.font.Unused[i], *dim)
		}
		p.given[i][c] = true
		*dim = p.fixValue()
	case codeNextLarger:
		if !p.setTag(c, ListTag) {
			return
		}
		ch.Tag = ListTag
		ch.NextLarger = p.byteValue()
	case codeVarChar:
		if !p.setTag(c, ExtTag) {
			return
		}
		ch.Tag = ExtTag
		p.recipes = append(p.recipes, c)
		p.list(2, func(code int) {
			if code < codeVarChar+1 || code > codeVarChar+4 {
				p.fail("a property which does not belong to a VARCHAR list")
				return
			}
			piece := [4]*uint8{&ch.Recipe.Top, &ch.Recipe.Mid, &ch.Recipe.Bot, &ch.Recipe.Rep}[code-codeVarChar-1]
			*piece = p.byteValue()
			p.finish()
		})
	}
	p.finish()
}

// ReadPL reads a PL file as PLtoTF reads it.
//
// The input is a list of properties in parentheses.  The outer level has
// CHECKSUM, DESIGNSIZE, DESIGNUNITS, CODINGSCHEME, FAMILY, FACE,
// SEVENBITSAFEFLAG, HEADER, FONTDIMEN, LIGTABLE, BOUNDARYCHAR and CHARACTER.
// A FONTDIMEN list has the names of the parameters and PARAMETER with a
// number; a LIGTABLE list has LABEL, the eight forms of LIG, KRN, STOP and
// SKIP; a CHARACTER list has CHARWD, CHARHT, CHARDP, CHARIC, NEXTLARGER and
// VARCHAR; a VARCHAR list has TOP, MID, BOT and REP.  COMMENT is allowed
// in every list.  A number has one of the forms C (a character), D (decimal),
// O (octal), H (hexadecimal), F (a face code) and R (a real number).  Letters
// count as upper case, except the character of a C value.
//
// A property which the input does not mention has the value of [NewFont].  A
// property which occurs more than once has its last value; the values before
// it of a dimension of a character are kept in Font.Unused.  ReadPL returns a
// [*SyntaxError] with the number of the line where PLtoTF reports an error
// while it reads; it does not report the indentation.  What PLtoTF checks
// after it has read the file is checked by [Font.Build].
func ReadPL(r io.Reader) (*Font, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	p := &plReader{data: data, font: NewFont(), cur: ' '}
	for p.err == nil {
		p.skipBlanks()
		if p.cur == '(' {
			p.property()
		} else if !p.ended {
			if p.cur == ')' {
				p.fail("a right parenthesis without a left one")
			} else {
				p.fail("text which is not in parentheses")
			}
		}
		if p.ended {
			break
		}
	}
	if p.err != nil {
		return nil, p.err
	}
	ascending := true
	for i := 1; i < len(p.recipes); i++ {
		if p.recipes[i-1] > p.recipes[i] {
			ascending = false
		}
	}
	if !ascending {
		p.font.RecipeOrder = p.recipes
	}
	return p.font, nil
}
