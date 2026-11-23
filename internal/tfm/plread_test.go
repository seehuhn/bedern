// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// plSource is a PL text of the tests.
type plSource struct {
	name string
	text string
}

// formsPL uses every form of a number and every property name.
const formsPL = `(COMMENT a comment (with (nested) parentheses) is ignored)
(family bedern forms)
(FACE F BIE)
(CODINGSCHEME two words
and a second line)
(DESIGNSIZE D 12)
(DESIGNUNITS R 20.0)
(CHECKSUM H DEADBEEF)
(checksum o 1234567)
(HEADER D 20 H 1F)
(HEADER O 26 O 37777777777)
(SEVENBITSAFEFLAG TRUE)
(SEVENBITSAFEFLAG FALSE)
(BOUNDARYCHAR C b)
(FONTDIMEN
(SLANT R -.25)
(SPACE D 6)
(SHRINK R 2.)
(STRETCH R +3)
(XHEIGHT R 10.55)
(QUAD D 18)
(EXTRASPACE R - - 1.5)
(PARAMETER D 9 R 0.123456789)
(NUM3 R 7)
(BIGOPSPACING5 R 1.0)
(COMMENT a comment in a list)
)
(LIGTABLE
(LABEL C f)
(LIG C f O 200)
(SKIP D 1)
(LABEL O 200)
(LIG C i O 201)
(KRN O 51 R 1.5)
(/LIG C ? C f)
(STOP)
(LABEL BOUNDARYCHAR)
(LIG/ C a H 82)
(/LIG/ C c D 131)
(LIG/> C d H 82)
(/LIG> C e H 82)
(/LIG/> C g H 82)
(/LIG/>> C h H 82)
(KRN C b R -0.5)
)
(LIGTABLE
(LABEL C A)
(krn c f d 2)
(KRN C b R -2.0000001)
(STOP)
)
(CHARACTER C f (CHARWD D 6) (CHARHT R 13.5) (CHARIC R 1.5))
(CHARACTER C i (CHARWD R 3.25) (CHARHT R 13.5))
(CHARACTER O 200 (CHARWD D 11))
(CHARACTER O 201 (CHARWD D 14) (CHARDP R -0.000001))
(CHARACTER H 82 (CHARWD D 1))
(CHARACTER D 131 (CHARWD D 1))
(CHARACTER O 51 (CHARWD D 4))
(CHARACTER C ? (CHARWD D 5))
(CHARACTER C a
(CHARWD R 1)
)
(CHARACTER C a (CHARHT R 2))
(CHARACTER C A (CHARWD R 12.345678))
(CHARACTER C c (CHARWD D 5) (NEXTLARGER C d))
(CHARACTER C d (CHARWD D 5) (NEXTLARGER C e))
(CHARACTER C e (CHARWD D 5))
(CHARACTER C g (CHARWD D 5))
(CHARACTER C h (CHARWD D 5))
(CHARACTER C z (VARCHAR (TOP C a) (MID C c) (BOT C d) (REP C e)))
(character c y (varchar (ext c g)) (charwd r 3))
(CHARACTER D 33)
`

// repeatPL gives dimensions of a character more than once.
const repeatPL = `(CHARACTER C a (CHARWD R 1) (CHARWD R 2) (CHARHT R 3) (CHARHT R 0) (CHARDP R 0) (CHARDP R 1))
(CHARACTER C b)
(CHARACTER C b (CHARWD R 4))
(CHARACTER C a (CHARIC R 0.5))
(CHARACTER C a (CHARIC R 0.25) (CHARWD R 2))
`

// plSources returns the PL texts of the tests which PLtoTF converts.
var plSources = []plSource{
	{"forms", formsPL},
	{"plain", "(CHARACTER C a)\n"},
	{"repeat", repeatPL},
	{"units", unitsPL()},
	{"kernhigh", kernHighPL()},
	{"bcharword0", bcharWord0PL},
	{"skip127", skip127PL()},
	{"sevenbit", sevenBitPL},
}

// unitWidths are the widths, in thousandths of the design size, of the PL
// text units.
var unitWidths = []int{0, 1, 2, 3, 333, 499, 500, 501, 667, 999, 1000, 1001, 1234, 2047, -1, -333, -2047, 7, 77, 777}

// unitsPL returns a PL text with 1000 design units and the widths of
// unitWidths.
func unitsPL() string {
	var b strings.Builder
	b.WriteString("(DESIGNUNITS R 1000)\n")
	for i, w := range unitWidths {
		fmt.Fprintf(&b, "(CHARACTER D %d (CHARWD R %d))\n", 40+i, w)
	}
	return b.String()
}

// TestReadPL compares the files which ReadPL and Build make of the PL
// texts of the tests with the files which PLtoTF makes of them.
func TestReadPL(t *testing.T) {
	for _, src := range plSources {
		font, err := ReadPL(strings.NewReader(src.text))
		if err != nil {
			t.Errorf("%s: ReadPL failed: %v", src.name, err)
			continue
		}
		_, _, data := build(t, font)
		want := golden(t, "pl-"+src.name+".tfm")
		if !bytes.Equal(data, want) {
			t.Errorf("%s: other bytes than PLtoTF, first at %d", src.name, firstDiff(data, want))
		}
	}
	for _, tf := range testFonts() {
		font, err := ReadPL(bytes.NewReader(golden(t, tf.name+".pl")))
		if err != nil {
			t.Errorf("%s: ReadPL failed: %v", tf.name, err)
			continue
		}
		_, _, data := build(t, font)
		want := golden(t, tf.name+".tfm")
		if !bytes.Equal(data, want) {
			t.Errorf("%s: other bytes than PLtoTF, first at %d", tf.name, firstDiff(data, want))
		}
	}
}

// TestReadPLForms checks the values which ReadPL reads from the PL text
// forms.
func TestReadPLForms(t *testing.T) {
	f, err := ReadPL(strings.NewReader(formsPL))
	if err != nil {
		t.Fatal(err)
	}
	if f.Family != "BEDERN FORMS" || f.CodingScheme != "TWO WORDS AND A SECOND LINE" {
		t.Errorf("strings %q and %q", f.Family, f.CodingScheme)
	}
	if f.Face != 15 || f.DesignSize != 12*Unity || f.DesignUnits != 20*Unity || !f.HasChecksum || f.Checksum != 0o1234567 {
		t.Errorf("header fields %d %v %v %v %o", f.Face, f.DesignSize, f.DesignUnits, f.HasChecksum, f.Checksum)
	}
	if want := []uint32{0, 0, 0x1F, 0, 0xFFFFFFFF}; !reflect.DeepEqual(f.Header, want) {
		t.Errorf("header words %v, want %v", f.Header, want)
	}
	if f.SevenBitSafe || !f.HasBoundary || f.BoundaryChar != 'b' {
		t.Errorf("flag %v, boundary %v %d", f.SevenBitSafe, f.HasBoundary, f.BoundaryChar)
	}
	wantParams := []Fix{-Unity / 4, 6 * Unity, 3 * Unity, 2 * Unity, FixFromFloat(10.55), 18 * Unity, 3 * Unity / 2, 0,
		FixFromFloat(0.1234567), 7 * Unity, 0, 0, Unity}
	if !reflect.DeepEqual(f.Params, wantParams) {
		t.Errorf("parameters %v, want %v", f.Params, wantParams)
	}
	if want := []uint8{'z', 'y'}; !reflect.DeepEqual(f.RecipeOrder, want) {
		t.Errorf("order of the recipes %v, want %v", f.RecipeOrder, want)
	}
	if a := f.Chars['a']; a == nil || a.Width != Unity || a.Height != 2*Unity {
		t.Errorf("character a is %+v", a)
	}
	if y := f.Chars['y']; y == nil || y.Tag != ExtTag || y.Recipe != (Recipe{Rep: 'g'}) || y.Width != 3*Unity {
		t.Errorf("character y is %+v", y)
	}
	if c := f.Chars['c']; c == nil || c.Tag != ListTag || c.NextLarger != 'd' {
		t.Errorf("character c is %+v", c)
	}
	if c := f.Chars[33]; c == nil || *c != (Char{}) {
		t.Errorf("character 33 is %+v", c)
	}
	if s, ok := f.Step(Boundary, 'h'); !ok || s != (Step{Next: 'h', Op: OpSlashLigSlashOver2, Result: 0x82}) {
		t.Errorf("step of the left boundary and h is %+v, %v", s, ok)
	}
	if s, ok := f.Step('f', '?'); !ok || s != (Step{Next: '?', Op: OpSlashLig, Result: 'f'}) {
		t.Errorf("step of f and ? is %+v, %v", s, ok)
	}
	if s, ok := f.Step(0o200, ')'); !ok || s != (Step{Next: ')', Op: OpKern, Kern: 3 * Unity / 2}) {
		t.Errorf("step of 200 and ) is %+v, %v", s, ok)
	}
}

// TestReadPLRepeat checks the values which ReadPL keeps of dimensions
// which a PL text gives more than once.
func TestReadPLRepeat(t *testing.T) {
	f, err := ReadPL(strings.NewReader(repeatPL))
	if err != nil {
		t.Fatal(err)
	}
	want := [4][]Fix{{Unity, 0, 2 * Unity}, {3 * Unity}, nil, {Unity / 2}}
	if !reflect.DeepEqual(f.Unused, want) {
		t.Errorf("unused values %v, want %v", f.Unused, want)
	}
	if a := f.Chars['a']; a == nil || *a != (Char{Width: 2 * Unity, Depth: Unity, Italic: Unity / 4}) {
		t.Errorf("character a is %+v", a)
	}
	checkPL(t, f)
}

// TestUnits checks that Units gives the fix word which PLtoTF computes for
// a dimension in thousandths of the design size.
func TestUnits(t *testing.T) {
	file, err := Read(bytes.NewReader(golden(t, "pl-units.tfm")))
	if err != nil {
		t.Fatal(err)
	}
	for i, w := range unitWidths {
		if got, want := Units(w, 1000), file.CharWidth(40+i); got != want {
			t.Errorf("Units(%d, 1000) is %d, PLtoTF has %d", w, got, want)
		}
	}
}

// plError is a PL text which PLtoTF reports as faulty, and the line of the
// fault.
type plError struct {
	text string
	line int
}

var plErrors = []plError{
	{"(FOO D 1)", 1},
	{"()", 1},
	{"(DEFAULTRULETHICKNESSES R 1)", 1},
	{"(CHARWD R 1)", 1},
	{"(STOP)", 1},
	{"\n\n(FACE D 256)", 3},
	{"(FACE O 400)", 1},
	{"(FACE H 100)", 1},
	{"(FACE X 1)", 1},
	{"(FACE F MRX)", 1},
	{"(FACE F XRR)", 1},
	{"(FACE)", 1},
	{"(DESIGNSIZE R 0.5)", 1},
	{"(DESIGNSIZE R 2048)", 1},
	{"(DESIGNSIZE R 2047.99999999)", 1},
	{"(DESIGNSIZE C a)", 1},
	{"(DESIGNUNITS R 0)", 1},
	{"(DESIGNUNITS R -1)", 1},
	{"(CHECKSUM D 5)", 1},
	{"(CHECKSUM O 8)", 1},
	{"(CHECKSUM O 40000000000)", 1},
	{"(CHECKSUM H 100000000)", 1},
	{"(HEADER D 17 O 1)", 1},
	{"(FAMILY TWENTY CHARACTERS XX)", 1},
	{"(CODINGSCHEME 0123456789012345678901234567890123456789)", 1},
	{"(FAMILY A (B))", 1},
	{"(SEVENBITSAFEFLAG MAYBE)", 1},
	{"(FONTDIMEN\n(PARAMETER D 0 R 1))", 2},
	{"(FONTDIMEN (CHARWD R 1))", 1},
	{"(FONTDIMEN (SLANT R 1) junk)", 1},
	{"(LIGTABLE (STOP))", 1},
	{"(LIGTABLE (LABEL C a) (SKIP D 1))", 1},
	{"(LIGTABLE (KRN C a R 1) (STOP) (STOP))", 1},
	{"(LIGTABLE (KRN C a R 1) (SKIP D 128))", 1},
	{"(LIGTABLE\n(LABEL C a)\n(LABEL C a))", 3},
	{"(LIGTABLE (CHARWD R 1))", 1},
	{"(LIGTABLE (KRN C a R 1))\n(LIGTABLE (STOP))", 2},
	{"(LIGTABLE (LIG C a))", 1},
	{"(CHARACTER C a (NEXTLARGER C b) (VARCHAR))", 1},
	{"(CHARACTER C a (VARCHAR) (NEXTLARGER C b))", 1},
	{"(LIGTABLE (LABEL C a) (KRN C a R 1))\n(CHARACTER C a\n(NEXTLARGER C b))", 3},
	{"(CHARACTER C a (FACE D 1))", 1},
	{"(CHARACTER C a (VARCHAR (CHARWD R 1)))", 1},
	{"(CHARACTER C ( )", 1},
	{"(CHARACTER C a) junk", 1},
	{")", 1},
	{"(CHARACTER C a))", 1},
	{"(CHARACTER C a (CHARWD R 1) x)", 1},
	{"(CHARACTER C a (CHARWD R 1 x))", 1},
	{"(CHARACTER C a (CHARWD X 1))", 1},
	{"(CHARACTER C a (CHARWD R 1.2.3))", 1},
	{"(CHARACTER C a", 2},
	{"(COMMENT not closed", 2},
	{"(FAMILY A\tB)", 1},
	{"(FACE D 1)\n\x80", 2},
}

// TestReadPLErrors checks that ReadPL reports the faults which PLtoTF
// reports while it reads, with the line.
func TestReadPLErrors(t *testing.T) {
	for _, c := range plErrors {
		_, err := ReadPL(strings.NewReader(c.text))
		var syntax *SyntaxError
		if !errors.As(err, &syntax) {
			t.Errorf("%q: error %v, want a SyntaxError", c.text, err)
		} else if syntax.Line != c.line {
			t.Errorf("%q: error in line %d, want line %d", c.text, syntax.Line, c.line)
		}
	}
}

// buildErrors are PL texts which PLtoTF reads without a fault and whose data
// it then changes with a message.
var buildErrors = []string{
	"(CHARACTER C a (NEXTLARGER C b))",
	"(LIGTABLE (LABEL C a) (KRN C b R 1))(CHARACTER C a)",
	"(LIGTABLE (LABEL C a) (LIG C a C b))(CHARACTER C a)",
	"(CHARACTER C a (NEXTLARGER C b))(CHARACTER C b (NEXTLARGER C a))",
	"(LIGTABLE (LABEL C a) (/LIG C x C y) (/LIG C y C x))(CHARACTER C a)(CHARACTER C x)(CHARACTER C y)",
	"(CHARACTER C a (CHARHT R 17))",
	"(CHARACTER C a (CHARWD R -16))",
	"(DESIGNUNITS R 2)(CHARACTER C a (CHARDP R 32))",
	"(CHARACTER C a (VARCHAR (REP C b)))",
	"(CHARACTER C a (VARCHAR (TOP C b) (REP C a)))",
	"(LIGTABLE (LABEL C a) (KRN C a R 16))(CHARACTER C a)",
	"(FONTDIMEN (SPACE R 16))",
	"(LIGTABLE (LABEL C a) (KRN C a R 1) (STOP) (KRN C b R 1))(CHARACTER C a)",
	"(LIGTABLE (LABEL C a))(CHARACTER C a)",
}

// stricter are PL texts which PLtoTF converts without a message and which
// Build refuses.
var stricter = []string{
	"(LIGTABLE (LABEL C b) (KRN C a R 1) (STOP) (LABEL C a))(CHARACTER C a)(CHARACTER C b)(CHARACTER D 0)",
	"(LIGTABLE (LABEL C a) (KRN C a R 1) (SKIP D 0))(CHARACTER C a)(CHARACTER D 0)",
}

// TestBuildErrors checks that Build refuses the descriptions whose data
// PLtoTF changes.
func TestBuildErrors(t *testing.T) {
	for _, text := range append(append([]string{}, buildErrors...), stricter...) {
		font, err := ReadPL(strings.NewReader(text))
		if err != nil {
			t.Errorf("%q: ReadPL failed: %v", text, err)
			continue
		}
		var invalid *InvalidFontError
		if _, _, err := font.Build(); !errors.As(err, &invalid) {
			t.Errorf("%q: Build gives %v, want an InvalidFontError", text, err)
		}
	}
}

// checkPL checks what holds for every description which ReadPL returns: it
// writes as a PL text which reads back as the same description, and a file
// which it builds is valid.
func checkPL(t *testing.T, font *Font) {
	var buf bytes.Buffer
	if err := font.WritePL(&buf); err != nil {
		t.Fatalf("WritePL failed: %v", err)
	}
	again, err := ReadPL(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("second read failed: %v", err)
	}
	if !reflect.DeepEqual(again, font) {
		t.Fatalf("the second read gives another description")
	}
	file, _, err := font.Build()
	if err != nil {
		return
	}
	buf.Reset()
	if err := file.Write(&buf); err != nil {
		t.Fatalf("the file built is not valid: %v", err)
	}
	if _, err := Read(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("the file built does not read: %v", err)
	}
	file.Check()
	file.Font()
}

func FuzzReadPL(f *testing.F) {
	for _, tf := range testFonts() {
		file, _, _ := build(f, tf.font)
		f.Add(description(f, tf.font))
		f.Add(text(f, file))
	}
	for _, of := range oddFiles(f) {
		f.Add(text(f, of.file))
	}
	for _, src := range plSources {
		f.Add([]byte(src.text))
	}
	for _, c := range plErrors {
		f.Add([]byte(c.text))
	}
	for _, text := range buildErrors {
		f.Add([]byte(text))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		font, err := ReadPL(bytes.NewReader(data))
		if err != nil {
			return
		}
		checkPL(t, font)
	})
}
