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

// TestGreekLowercase checks that the regular font maps the Greek lowercase
// U+03B1 to U+03C9 and the bold font does not.
func TestGreekLowercase(t *testing.T) {
	t.Parallel()
	r := ran(t)
	for _, f := range fonts {
		if f.italic {
			continue
		}
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
			t.Fatalf("%s: %v", f.file, err)
		}
		for c := rune(0x3B1); c <= 0x3C9; c++ {
			if got := bmp.Lookup(c) != 0; got != f.greek {
				t.Errorf("%s: %U is mapped: %t, want %t", f.file, c, got, f.greek)
			}
		}
	}
}
