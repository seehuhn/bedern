// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"
	"path/filepath"
	"testing"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/cmap"
)

// TestCharacterMaps checks the character maps of every font: the subtable
// for the Basic Multilingual Plane maps A and does not map the copyleft
// sign, and the subtable for all of Unicode maps both, the copyleft sign
// U+1F12F to the glyph u1F12F.
func TestCharacterMaps(t *testing.T) {
	t.Parallel()
	const copyleft = '\U0001F12F'
	r := ran(t)
	for _, f := range fonts {
		fname := filepath.Join(t.TempDir(), f.file)
		if err := os.WriteFile(fname, r.files[f.file], 0o666); err != nil {
			t.Fatal(err)
		}
		font, err := sfnt.ReadFile(fname)
		if err != nil {
			t.Fatalf("%s: %v", f.file, err)
		}
		bmp, err := font.CMapTable.Get(cmap.Key{PlatformID: 3, EncodingID: 1})
		if err != nil {
			t.Fatalf("%s: no subtable for the Basic Multilingual Plane: %v", f.file, err)
		}
		if bmp.Lookup('A') == 0 || bmp.Lookup(copyleft) != 0 {
			t.Errorf("%s: the subtable for the Basic Multilingual Plane maps A to %d and %U to %d", f.file, bmp.Lookup('A'), copyleft, bmp.Lookup(copyleft))
		}
		for _, key := range []cmap.Key{{PlatformID: 3, EncodingID: 10}, {PlatformID: 0, EncodingID: 4}} {
			full, err := font.CMapTable.Get(key)
			if err != nil {
				t.Errorf("%s: no subtable %d, %d: %v", f.file, key.PlatformID, key.EncodingID, err)
				continue
			}
			if full.Lookup('A') != bmp.Lookup('A') {
				t.Errorf("%s: the subtable %d, %d maps A to %d, want %d", f.file, key.PlatformID, key.EncodingID, full.Lookup('A'), bmp.Lookup('A'))
			}
			if name := font.GlyphName(full.Lookup(copyleft)); name != "u1F12F" {
				t.Errorf("%s: the subtable %d, %d maps %U to the glyph %q, want u1F12F", f.file, key.PlatformID, key.EncodingID, copyleft, name)
			}
		}
	}
}
