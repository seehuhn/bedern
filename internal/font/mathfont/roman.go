// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

// The glyphs of the first tier which the maths font takes from the regular as
// they are: the figures, the
// upright Latin letters, the ASCII signs, the signs of Latin-1 and of
// mathematics, the ellipsis, the primes, the combining accents, the dotless i
// and j and the text signs which papers use in formulas.  Each takes the code
// point of the regular, which is the one unicode-math uses for it.

// romanSigns are the glyphs of the regular other than the figures and the
// letters, by their names in the regular.
var romanSigns = []string{
	"exclam", "parenleft", "parenright", "asterisk", "plus", "comma", "period",
	"slash", "colon", "semicolon", "less", "equal", "greater", "question",
	"bracketleft", "backslash", "bracketright", "braceleft", "bar", "braceright",
	"minus", "plusminus", "multiply", "divide", "logicalnot",
	"ellipsis", "minute", "second",
	"gravecomb", "acutecomb", "uni0302", "tildecomb", "uni0304", "uni0306",
	"uni0307", "uni0308", "uni030A", "uni030C",
	"dotlessi", "uni0237",
	"dagger", "daggerdbl", "section", "paragraph", "bullet",
}

// accentNames are the combining accents among romanSigns.
var accentNames = map[string]bool{
	"gravecomb": true, "acutecomb": true, "uni0302": true, "tildecomb": true,
	"uni0304": true, "uni0306": true, "uni0307": true, "uni0308": true,
	"uni030A": true, "uni030C": true,
}

// romanNames returns the names of the roman glyphs of the first tier, in the
// order of the font: the figures, the capitals, the lowercase and the signs.
func romanNames() []string {
	var names []string
	for _, n := range []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"} {
		names = append(names, n)
	}
	for c := 'A'; c <= 'Z'; c++ {
		names = append(names, string(c))
	}
	for c := 'a'; c <= 'z'; c++ {
		names = append(names, string(c))
	}
	return append(names, romanSigns...)
}
