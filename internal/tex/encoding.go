// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"fmt"
	"strings"

	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// A Slot names the glyph of one position of an encoding: the glyph with the
// name Name or, if the font has no such glyph or Name is empty, the glyph
// with the code point Rune.  A position which has a form of its own of a
// character (quotesingle.ts1) names the form and gives the code point of the
// character, whose glyph stands in for the form while the font does not have
// it.  The zero value is a position which the encoding does not use.
type Slot struct {
	Rune rune
	Name string
}

// used says whether the encoding uses the position.
func (s Slot) used() bool { return s.Rune != 0 || s.Name != "" }

// An Encoding is one of the 8-bit encodings of LaTeX.
type Encoding struct {
	// Name is the name of the encoding in LaTeX, for example "T1".
	Name string

	// Scheme is the coding scheme which the metric files state.
	Scheme string

	// Slots holds the glyph of each of the 256 positions.
	Slots [256]Slot
}

// Tag returns the name of the encoding in lower case, as the names of files
// use it.
func (e *Encoding) Tag() string { return strings.ToLower(e.Name) }

// VectorName returns the name of the encoding vector in the encoding file.
func (e *Encoding) VectorName() string { return "Bedern" + e.Name + "Encoding" }

// EncFile returns the name of the encoding file.
func (e *Encoding) EncFile() string { return "bedern-" + e.Tag() + ".enc" }

// Resolve returns, for each position of the encoding, the index in gs of its
// glyph, or -1 for a position which the encoding does not use or for which gs
// has no glyph.  Missing lists the positions which the encoding uses and for
// which gs has no glyph, in increasing order.
func (e *Encoding) Resolve(gs []glyphs.Glyph) (index [256]int, missing []int) {
	index, missing, _ = e.resolve(gs)
	return index, missing
}

// StandIns returns the positions of the encoding which name a form of their
// own and take the glyph of their code point, since gs does not have the
// form, in increasing order.
func (e *Encoding) StandIns(gs []glyphs.Glyph) []int {
	_, _, standIns := e.resolve(gs)
	return standIns
}

func (e *Encoding) resolve(gs []glyphs.Glyph) (index [256]int, missing, standIns []int) {
	byRune := make(map[rune]int, len(gs))
	byName := make(map[string]int, len(gs))
	for i, g := range gs {
		if i == 0 {
			continue
		}
		if g.Rune != 0 {
			byRune[g.Rune] = i
		}
		byName[g.Name] = i
	}
	for c, s := range e.Slots {
		index[c] = -1
		if !s.used() {
			continue
		}
		if i, ok := byName[s.Name]; ok {
			index[c] = i
		} else if i, ok := byRune[s.Rune]; ok {
			index[c] = i
			if s.Name != "" {
				standIns = append(standIns, c)
			}
		} else {
			missing = append(missing, c)
		}
	}
	return index, missing, standIns
}

// Describe returns a short description of the glyph of the position c: its
// code point, its name or both.
func (e *Encoding) Describe(c int) string {
	s := e.Slots[c]
	switch {
	case s.Rune != 0 && s.Name != "":
		return fmt.Sprintf("%U %s", s.Rune, s.Name)
	case s.Rune != 0:
		return fmt.Sprintf("%U", s.Rune)
	default:
		return s.Name
	}
}

// T1 is the encoding T1 of LaTeX.  The glyph of a position is the glyph at
// the code point of the character of that position.  Position 127 holds the
// hyphen a second time, and position 223 the two letters S.
var T1 = &Encoding{Name: "T1", Scheme: "EXTENDED TEX FONT ENCODING - LATIN", Slots: t1Slots()}

// TS1 is the encoding TS1 of LaTeX.  The glyph of a position is the glyph at
// the code point to which LaTeX maps the command of that position; the
// positions whose glyphs have no code point, and those which have a form of
// their own of a character, name their glyphs.
var TS1 = &Encoding{Name: "TS1", Scheme: "TEX TEXT COMPANION SYMBOLS 1---TS1", Slots: ts1Slots()}

// Encodings lists the encodings for which the fonts are written.
var Encodings = []*Encoding{T1, TS1}

func t1Slots() [256]Slot {
	var s [256]Slot
	for c, r := range []rune{
		0x0060, 0x00B4, 0x02C6, 0x02DC, 0x00A8, 0x02DD, 0x02DA, 0x02C7,
		0x02D8, 0x00AF, 0x02D9, 0x00B8, 0x02DB, 0x201A, 0x2039, 0x203A,
		0x201C, 0x201D, 0x201E, 0x00AB, 0x00BB, 0x2013, 0x2014, 0x200C,
		0xEB4D, 0x0131, 0x0237, 0xFB00, 0xFB01, 0xFB02, 0xFB03, 0xFB04,
		0x2423,
	} {
		s[c].Rune = r
	}
	s[0x18].Name = "perthousandzero"
	for c := 0x21; c <= 0x7E; c++ {
		s[c].Rune = rune(c)
	}
	s[0x27].Rune = 0x2019
	s[0x60].Rune = 0x2018
	s[0x7F].Rune = '-'
	for i, r := range []rune{
		0x0102, 0x0104, 0x0106, 0x010C, 0x010E, 0x011A, 0x0118, 0x011E,
		0x0139, 0x013D, 0x0141, 0x0143, 0x0147, 0x014A, 0x0150, 0x0154,
		0x0158, 0x015A, 0x0160, 0x015E, 0x0164, 0x0162, 0x0170, 0x016E,
		0x0178, 0x0179, 0x017D, 0x017B, 0x0132, 0x0130, 0x0111, 0x00A7,
		0x0103, 0x0105, 0x0107, 0x010D, 0x010F, 0x011B, 0x0119, 0x011F,
		0x013A, 0x013E, 0x0142, 0x0144, 0x0148, 0x014B, 0x0151, 0x0155,
		0x0159, 0x015B, 0x0161, 0x015F, 0x0165, 0x0163, 0x0171, 0x016F,
		0x00FF, 0x017A, 0x017E, 0x017C, 0x0133, 0x00A1, 0x00BF, 0x00A3,
	} {
		s[0x80+i].Rune = r
	}
	for c := 0xC0; c <= 0xFF; c++ {
		s[c].Rune = rune(c)
	}
	s[0xD7].Rune = 0x0152
	s[0xDF] = Slot{Name: "S_S"}
	s[0xF7].Rune = 0x0153
	s[0xFF].Rune = 0x00DF
	return s
}

func ts1Slots() [256]Slot {
	var s [256]Slot
	for c, name := range []string{
		"grave", "acute", "circumflex", "tilde", "dieresis", "hungarumlaut",
		"ring", "caron", "breve", "macron", "dotaccent",
	} {
		s[c].Name = name + ".case"
	}
	for c, name := range []string{
		"zero", "one", "two", "three", "four", "five", "six", "seven",
		"eight", "nine",
	} {
		s[0x30+c].Name = name + ".taboldstyle"
	}
	for c, name := range map[int]string{
		0x0D: "quotesinglbase.ts1",
		0x12: "quotedblbase.ts1",
		0x1C: "tieaccentlowercase.new",
		0x1D: "tieaccentcapital.new",
		0x8A: "dollar.oldstyle",
		0x8B: "cent.oldstyle",
		// the forms of TS1 of characters of T1
		0x27: "quotesingle.ts1",
		0x60: "grave.ts1",
		0x7F: "hyphendbl.alt",
		0x80: "breve.ts1",
		0x81: "caron.ts1",
		0x82: "hungarumlaut.ts1",
		0x83: "dblgrave.ts1",
		0x99: "paragraph.alt",
		0xA8: "dieresis.ts1",
		0xAF: "macron.ts1",
		0xB4: "acute.ts1",
	} {
		s[c].Name = name
	}
	for c, r := range map[int]rune{
		0x0B: 0x00B8, 0x0C: 0x02DB,
		0x15: 0xE400, 0x16: 0xF6DE, 0x17: 0xEB12, 0x18: 0x2190, 0x19: 0x2192,
		0x1A: 0xEB65, 0x1B: 0xEB63, 0x1F: 0xEB11, 0x20: 0x2422, 0x24: 0x0024,
		0x27: 0x0027, 0x2A: 0x2217, 0x2C: 0x002C, 0x2D: 0x2E40, 0x2E: 0x002E,
		0x2F: 0x2044, 0x3C: 0x2329, 0x3D: 0x2212, 0x3E: 0x232A, 0x4D: 0x2127,
		0x4F: 0x25EF, 0x51: 0xEC10, 0x57: 0x2126, 0x5B: 0x27E6, 0x5D: 0x27E7,
		0x5E: 0x2191, 0x5F: 0x2193, 0x60: 0x0060, 0x62: 0xEB09, 0x63: 0x26AE,
		0x64: 0xEB16, 0x6C: 0xEB40, 0x6D: 0x26AD, 0x6E: 0x266A, 0x71: 0xEC11,
		0x7E: 0x02F7, 0x7F: 0x2E40,
		0x80: 0x02D8, 0x81: 0x02C7, 0x82: 0x02DD, 0x83: 0x02F5, 0x84: 0x2020,
		0x85: 0x2021, 0x86: 0x2016, 0x87: 0x2030, 0x88: 0x2022, 0x89: 0x2103,
		0x8C: 0x0192, 0x8D: 0x20A1, 0x8E: 0x20A9, 0x8F: 0x20A6, 0x90: 0x20B2,
		0x91: 0x20B1, 0x92: 0x20A4, 0x93: 0x211E, 0x94: 0x203D, 0x95: 0x2E18,
		0x96: 0x20AB, 0x97: 0x2122, 0x98: 0x2031, 0x99: 0x00B6, 0x9A: 0x0E3F,
		0x9B: 0x2116, 0x9C: 0x2052, 0x9D: 0x212E, 0x9E: 0x25E6, 0x9F: 0x2120,
		0xA0: 0x2045, 0xA1: 0x2046, 0xAB: 0x1F12F, 0xAD: 0x2117, 0xB8: 0x203B,
		0xBB: 0x221A, 0xBF: 0x20AC, 0xD6: 0x00D7, 0xF6: 0x00F7,
	} {
		s[c].Rune = r
	}
	for _, c := range []int{
		0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAC, 0xAE, 0xAF,
		0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB9, 0xBA, 0xBC, 0xBD,
		0xBE,
	} {
		s[c].Rune = rune(c)
	}
	return s
}
