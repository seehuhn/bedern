// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

// specialList lists the special Latin letters, the further punctuation and the
// signs of Latin-1 and of text, in glyph order.  They follow the glyphs of
// compose.go in the table.
var specialList = []entry{
	// Letters of European languages.
	{"germandbls", 'ß', sideStem, sideRound, eszett},
	{"uni1E9E", 'ẞ', sideCapStem, sideCapRound, capEszett},
	{"ae", 'æ', sideBowlA, sideOpen, ae},
	{"AE", 'Æ', sideCapA, sideCapArmE, capAE},
	{"oe", 'œ', sideRound, sideOpen, oe},
	{"OE", 'Œ', sideCapRound, sideCapArmE, capOE},
	{"oslash", 'ø', sideOpen, sideOpen, oslash},
	{"Oslash", 'Ø', sideOpen, sideOpen, capOslash},
	{"lslash", 'ł', sideOpen, sideOpen, lslash},
	{"Lslash", 'Ł', sideCapStem, sideCapArmL, capLslash},
	{"dcroat", 'đ', sideOpen, sideStemD, dcroat},
	{"Dcroat", 'Đ', sideCapBowlStem, sideCapBowlD, capDcroat},
	{"eth", 'ð', sideOpen, sideOpen, eth},
	{"Eth", 'Ð', sideCapBowlStem, sideCapBowlD, capDcroat},
	{"thorn", 'þ', sideStemBP, sideBowlBP, thorn},
	{"Thorn", 'Þ', sideCapBowlStem, sideCapBowl, capThorn},
	{"hbar", 'ħ', sideStem, sideStem, hbar},
	{"Hbar", 'Ħ', sideCapStem, sideCapStem, capHbar},
	{"tbar", 'ŧ', sideBarT, sideOpen, tbar},
	{"Tbar", 'Ŧ', sideCapT, sideCapT, capTbar},
	{"eng", 'ŋ', sideStem, sideStem, eng},
	{"Eng", 'Ŋ', sideCapN, sideCapN, capEng},
	{"kgreenlandic", 'ĸ', sideStem, sideK, kgreenlandic},
	{"longs", 'ſ', sideBarF, sideOverhang, longs},
	{"ij", 'ĳ', sideNarrow, sideNarrow, ij},
	{"IJ", 'Ĳ', sideCapStem, sideCapStem, capIJ},
	{"ldot", 'ŀ', sideNarrow, sideOverhang, ldot},
	{"Ldot", 'Ŀ', sideCapStem, sideCapArmL, capLdot},
	{"napostrophe", 'ŉ', sideOpen, sideStem, napostrophe},

	// Punctuation and typographic marks.
	{"exclamdown", '¡', sideOpen, sideOpen, exclamdown},
	{"questiondown", '¿', sideOpen, sideOpen, questiondown},
	{"guillemotleft", '«', sideOpen, sideOpen, guillemotleft},
	{"guillemotright", '»', sideOpen, sideOpen, guillemotright},
	{"guilsinglleft", '‹', sideOpen, sideOpen, guilsinglleft},
	{"guilsinglright", '›', sideOpen, sideOpen, guilsinglright},
	{"quotesinglbase", '‚', sideOpen, sideOpen, quotesinglbase},
	{"quotedblbase", '„', sideOpen, sideOpen, quotedblbase},
	{"ellipsis", '…', sideOpen, sideOpen, ellipsis},
	{"bullet", '•', sideOpen, sideOpen, bullet},
	{"periodcentered", '·', sideOpen, sideOpen, periodcentered},
	{"dagger", '†', sideOpen, sideOpen, dagger},
	{"daggerdbl", '‡', sideOpen, sideOpen, daggerdbl},
	{"paragraph", '¶', sideOpen, sideOpen, paragraph},
	{"minute", '′', sideOpen, sideOpen, minute},
	{"second", '″', sideOpen, sideOpen, second},
	{"perthousand", '‰', sideOpen, sideOpen, perthousand},
	{"uni00A0", '\u00A0', sideOpen, sideOpen, nbspace},
	{"uni2009", '\u2009', sideOpen, sideOpen, thinspace},
	{"uni200A", '\u200A', sideOpen, sideOpen, hairspace},
	{"uni202F", '\u202F', sideOpen, sideOpen, thinspace},
	{"uni00AD", '\u00AD', sideHyphen, sideHyphen, softhyphen},

	// Signs of Latin-1 and of text.
	{"multiply", '×', sideMath, sideMath, multiply},
	{"divide", '÷', sideMath, sideMath, divide},
	{"plusminus", '±', sideMath, sideMath, plusminus},
	{"logicalnot", '¬', sideMath, sideMath, logicalnot},
	{"mu", 'µ', sideStem, sideStem, mu},
	{"uni00B9", '¹', sideOpen, sideOpen, onesuperior},
	{"uni00B2", '²', sideOpen, sideOpen, twosuperior},
	{"uni00B3", '³', sideOpen, sideOpen, threesuperior},
	{"fraction", '⁄', sideOpen, sideOpen, fraction},
	{"onequarter", '¼', sideOpen, sideOpen, onequarter},
	{"onehalf", '½', sideOpen, sideOpen, onehalf},
	{"threequarters", '¾', sideOpen, sideOpen, threequarters},
	{"ordfeminine", 'ª', sideOpen, sideOpen, ordfeminine},
	{"ordmasculine", 'º', sideOpen, sideOpen, ordmasculine},
	{"brokenbar", '¦', sideBar, sideBar, brokenbar},
	{"copyright", '©', sideOpen, sideOpen, copyright},
	{"registered", '®', sideOpen, sideOpen, registered},
	{"cent", '¢', sideOpen, sideOpen, cent},
	{"yen", '¥', sideOpen, sideOpen, yen},
	{"currency", '¤', sideOpen, sideOpen, currency},
}

func init() {
	table = append(table, specialList...)
}
