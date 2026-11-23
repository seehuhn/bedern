// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"os"
	"testing"
)

// TestRounderTies: with the second stylistic set a shaping engine sets the
// rounder forms of the two ties, and without it the ties; the letters stay as
// they are; and every name table of the written file names the set.
func TestRounderTies(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	const text = "oO"
	for _, c := range []struct{ args, want string }{
		{"--features=ss02", "tieaccentlowercase.new o tieaccentcapital.new O"},
		{"--features=-ss02", "tieaccentlowercase o tieaccentcapital O"},
		{"--features=ss01", "tieaccentlowercase o tieaccentcapital O"},
		{"--language=de", "tieaccentlowercase o tieaccentcapital O"},
	} {
		if got := shapeNames(t, path, fname, text, c.args); got != c.want {
			t.Errorf("the ties with %s: %s, want %s", c.args, got, c.want)
		}
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	tables := nameTables(t, data)
	if len(tables) == 0 {
		t.Fatal("the written file has no name table")
	}
	for key, tab := range tables {
		if tab.Extra[ss02NameID] != StylisticSetNames[ss02NameID] || tab.Extra[ss02NameID] == "" {
			t.Errorf("the name table %s has %q for the second stylistic set, want %q", key, tab.Extra[ss02NameID], StylisticSetNames[ss02NameID])
		}
	}
}
