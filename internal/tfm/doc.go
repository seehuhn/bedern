// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package tfm reads and writes TeX font metric files and their property-list
// form.
//
// A TFM file is the binary file from which TeX82 takes the metrics of a font.
// It is a sequence of 32-bit words in big-endian order.  The first six words
// hold twelve lengths; they are followed by the header (check sum, design
// size, coding scheme, family, the seven-bit-safe flag, the face byte and any
// further words), one word for each character code between the smallest and
// the largest code of the font, the tables of widths, heights, depths and
// italic corrections, the ligature and kerning program, the table of kerns,
// the extensible recipes and the font parameters.  The word of a character
// holds indices into the four tables of dimensions, a tag and a remainder; the
// tag says whether the remainder is the start of a ligature and kerning
// program, the code of the next larger character or the index of an
// extensible recipe.  All dimensions are fix words, signed multiples of 2^-20
// of the design size.
//
// A PL file is the text form of the same data, a list of properties in
// parentheses.  The programs TFtoPL and PLtoTF convert between the two forms.
//
// The package has two levels.
//
// The faithful level is the type [File].  It holds the tables as they stand
// in a TFM file.  [Read] accepts the files which TeX82 accepts, and
// [File.Write] writes every such file back byte for byte.  [File.WritePL]
// prints a file as TFtoPL prints it, and [File.Check] lists the points which
// TFtoPL reports.
//
// The descriptive level is the type [Font].  It holds what a PL file says:
// the fields of the header, the parameters, the four dimensions and the tag of
// every character, the boundary character and the ligature and kerning
// program as a list of labels, steps, stops and skips.  [ReadPL] reads a PL
// file as PLtoTF reads it.  [Font.Build] turns a description into the tables
// by the method of PLtoTF, and returns a [Report] of the reductions of the
// tables of dimensions.  [File.Font] describes a file as TFtoPL describes it.
package tfm
