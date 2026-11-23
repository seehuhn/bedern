// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"io"
	"slices"
)

// A PrivateUse is a glyph which is encoded in the Private Use Area: its code
// point, its glyph name, the TeX command which selects it and what it is.
type PrivateUse struct {
	Rune        rune
	Name        string
	Command     string
	Description string
}

// privateUse lists the glyphs in the Private Use Area.  The glyph table takes
// the code points of these glyphs from this list.
var privateUse = []PrivateUse{
	{0xEB11, "cwmascender", `\textascendercompwordmark`, "the compound word mark as high as the ascenders, an empty glyph without width"},
	{0xEB12, "cwmcapital", `\textcapitalcompwordmark`, "the compound word mark as high as the capitals, an empty glyph without width"},
	{0xEB16, "died", `\textdied`, "the sign for the year of death in genealogy"},
	{0xEB32, "hyphendbl.alt", `\textdblhyphenchar`, "the double hyphen of TS1 as a hyphenation character"},
	{0xEB40, "leaf", `\textleaf`, "the leaf of TS1"},
	{0xEB4C, "paragraph.alt", `\textpilcrow`, "the pilcrow of TS1 with a single stem"},
	{0xEB4D, "perthousandzero", `\textperthousandzero`, "the small zero of T1 which is appended to the per cent sign, once to give ‰ and twice to give ‱"},
	{0xEB52, "quotedblbase.ts1", `\textquotestraightdblbase`, "the straight double quotation mark on the baseline"},
	{0xEB56, "quotesinglbase.ts1", `\textquotestraightbase`, "the straight single quotation mark on the baseline"},
	{0xEB63, "tieaccentcapital", `\capitaltie`, "the tie over two capitals"},
	{0xEB65, "tieaccentlowercase", `\t`, "the tie over two lowercase letters"},
	{0xEC10, "Orogate", `\textOrogate`, "the capital o with a rogate of TS1"},
	{0xEC11, "orogate", `\textorogate`, "the lowercase o with a rogate of TS1"},
	{0xF6DE, "threequartersemdash", `\textthreequartersemdash`, "the dash of three quarters of an em"},
	{0xF724, "dollar.oldstyle", `\textdollaroldstyle`, "the dollar sign of the old-style figures"},
	{0xF7A2, "cent.oldstyle", `\textcentoldstyle`, "the cent sign of the old-style figures"},
	{0xE400, "twelveudash", `\texttwelveudash`, "the dash of two thirds of an em"},
	{0xE401, "born", `\textborn`, "the sign for the year of birth in genealogy"},
}

// PrivateUseFirst and PrivateUseLast are the first and the last code point of
// the Private Use Area of the Basic Multilingual Plane.
const (
	PrivateUseFirst = ''
	PrivateUseLast  = ''
)

// PrivateUseTable returns the table of the glyphs in the Private Use Area, in
// the order of the table.
func PrivateUseTable() []PrivateUse {
	return slices.Clone(privateUse)
}

// privateUseRune returns the code point of the named glyph of the table.
func privateUseRune(name string) rune {
	for _, e := range privateUse {
		if e.Name == name {
			return e.Rune
		}
	}
	panic("glyphs: " + name + " is not in the table of the Private Use Area")
}

// WritePrivateUse writes the table of the glyphs in the Private Use Area as
// text: one line for each glyph with the code point, the glyph name, the TeX
// command and the description, separated by tabs.
func WritePrivateUse(w io.Writer) error {
	for _, e := range privateUse {
		if _, err := fmt.Fprintf(w, "U+%04X\t%s\t%s\t%s\n", e.Rune, e.Name, e.Command, e.Description); err != nil {
			return err
		}
	}
	return nil
}
