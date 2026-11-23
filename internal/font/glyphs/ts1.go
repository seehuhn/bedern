// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

// The glyphs which complete the encoding TS1 follow the glyphs of special.go
// in the table.
func init() {
	table = append(table, caseList...)
	table = append(table, textMarksList...)
	table = append(table, textSignsList...)
	table = append(table, figureList...)
	table = append(table, currencyList...)
	table = append(table, letterlikeList...)
	table = append(table, pictureList...)
	table = append(table, tieList...)
	table = append(table, signList...)
}
