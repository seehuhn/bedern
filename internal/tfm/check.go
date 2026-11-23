// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

// FindingKind is the kind of a [Finding].
type FindingKind uint8

// The kinds of a finding.  TFtoPL prints a message for every kind but
// UnusedSteps.  It changes the data which it prints for StringTooLong,
// StringParenthesis, StringNotASCII, LargerMissing and NonstandardLig, and it
// prints no characters for LigatureLoop and TooManyPairs.
const (
	// StringTooLong says that the length byte of the coding scheme or of
	// the family is larger than the field.
	StringTooLong FindingKind = iota

	// StringParenthesis says that the coding scheme or the family holds a
	// parenthesis.
	StringParenthesis

	// StringNotASCII says that the coding scheme or the family holds a byte
	// which is no printable ASCII character.
	StringNotASCII

	// LargerMissing says that the next larger character of a character
	// does not exist.
	LargerMissing

	// NonstandardLig says that the operation of a ligature step is none of
	// the eight ligature operations.
	NonstandardLig

	// ParamCount says that a font of mathematical symbols has not 22
	// parameters, or a font of mathematical extension characters not 13.
	ParamCount

	// TrailingBytes says that bytes follow the stated length of the file.
	TrailingBytes

	// LigatureLoop says that the ligature steps of a pair of characters
	// never come to an end.
	LigatureLoop

	// TooManyPairs says that the ligature and kerning program has more
	// pairs of characters than TFtoPL holds.
	TooManyPairs

	// UnusedSteps says that no program reaches some of the steps of the
	// ligature and kerning program.
	UnusedSteps
)

// Finding is a point which TeX82 accepts in a TFM file and which TFtoPL
// reports or marks.
type Finding struct {
	Kind FindingKind
	Text string
}

// Check returns the findings about a file.  The file must be one which [Read]
// or [Font.Build] returned or one which [File.Write] accepts.
func (f *File) Check() []Finding {
	return f.view().finds
}
