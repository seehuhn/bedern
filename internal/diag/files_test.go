// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package diag

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// The tests share the builds of the glyphs.
func init() { GlyphSet = glyphstest.Set }

func TestFileName(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		".notdef":   "notdef",
		"space":     "space",
		"l":         "l",
		"A":         "A_",
		"Aacute":    "A_acute",
		"a.sc":      "a.sc",
		"f_i":       "f%5fi",
		dots + "/x": "%2e%2fx",
		"..":        "%2e",
		"a/b":       "a%2fb",
		`a\b`:       "a%5cb",
		"..a":       "%2ea",
		"":          "%",
		"a b":       "a%20b",
	}
	for name, want := range cases {
		if got := FileName(name); got != want {
			t.Errorf("FileName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestFileNameSafe(t *testing.T) {
	t.Parallel()
	names := []string{".notdef", "a", "A", dots + "/" + dots + "/etc/passwd", `..\x`, "/abs", "a/" + dots + "/b", ".", "..", "", ".hidden", "x\x00y"}
	for _, name := range names {
		f := FileName(name)
		if f == "" || f == "." || f == ".." || strings.HasPrefix(f, ".") ||
			strings.ContainsAny(f, "/\\\x00") {
			t.Errorf("FileName(%q) = %q is unsafe", name, f)
		}
	}
}

func TestFileNameCase(t *testing.T) {
	t.Parallel()
	// Names which differ only in case must differ on a case-insensitive
	// file system.
	names := []string{"a", "A", "ab", "aB", "Ab", "AB", "a_", "A_", "Aa", "aA", "f_i", "F_I", "F_i"}
	seen := map[string]string{}
	for _, name := range names {
		f := strings.ToLower(FileName(name))
		if prev, ok := seen[f]; ok {
			t.Errorf("%q and %q both give %q", prev, name, f)
		}
		seen[f] = name
	}
}

func TestWriteAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	vs := design.Variants()
	if err := WriteAll(dir, vs); err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		for _, name := range []string{"notdef", "space", "l", "o"} {
			path := filepath.Join(dir, v.Name, name+".svg")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Error(err)
				continue
			}
			if !bytes.Contains(data, []byte("<svg")) {
				t.Errorf("%s is not an SVG file", path)
			}
		}
	}
}

func TestWriteAllOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A modified parameter set as a second variant, which is drawn over the
	// regular one.
	vs := append(design.Variants(), design.Variant{Name: "stem70", Note: "a lighter stem", Params: designtest.Perturbed()[0].Params})
	if err := WriteAll(dir, vs); err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		for _, name := range []string{"notdef", "space", "l", "o"} {
			data, err := os.ReadFile(filepath.Join(dir, v.Name, name+".svg"))
			if err != nil {
				t.Fatal(err)
			}
			has := bytes.Contains(data, []byte(`class="overlay"`))
			want := v.Name != "regular" && name != "space"
			if has != want {
				t.Errorf("%s/%s: overlay present is %v, want %v", v.Name, name, has, want)
			}
		}
	}
}

// dots is the name of the parent directory.
const dots = ".."
