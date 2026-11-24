// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

var update = flag.Bool("update", false, "rewrite the files of testdata/golden")

// goldenDir holds one file per glyph, named after the glyph, with the advance
// and the fingerprint of the rounded outline at the default parameters.
const goldenDir = "testdata/golden"

// fingerprint returns a short hash of the rounded outline, written as text,
// in the manner of the roman's fingerprints.
func fingerprint(path outline.Path) string {
	var sb strings.Builder
	pt := func(q outline.Point) { fmt.Fprintf(&sb, " %d %d", int(q.X), int(q.Y)) }
	for _, c := range path {
		sb.WriteString("M")
		pt(c.Start)
		for _, s := range c.Segs {
			if s.IsCurve {
				sb.WriteString(" C")
				pt(s.P1)
				pt(s.P2)
			} else {
				sb.WriteString(" L")
			}
			pt(s.P3)
		}
		sb.WriteString(" Z\n")
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(sb.String())))[:12]
}

// TestGolden catches unintended changes of the outlines: it compares the
// advance and the fingerprint of every letter, special letter and ligature
// at the default parameters with the file of the glyph in testdata/golden.
// An intended change is accepted with
// `go test . -run TestGolden -update`.
func TestGolden(t *testing.T) {
	want := map[string]bool{}
	for _, name := range Names() {
		want[name+".txt"] = true
		g := glyph(t, name)
		line := fmt.Sprintf("regular %s %g %s\n", name, g.Advance, fingerprint(g.Path))
		file := filepath.Join(goldenDir, name+".txt")
		if *update {
			old, _ := os.ReadFile(file)
			if string(old) != line {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(file, []byte(line), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("%s: rewritten", name)
			}
			continue
		}
		old, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("%s: no fingerprint (%v)", name, err)
			continue
		}
		if string(old) != line {
			t.Errorf("%s: the outline changed: %q, want %q", name, strings.TrimSpace(line), strings.TrimSpace(string(old)))
		}
	}
	files, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if !want[f.Name()] {
			t.Errorf("%s: a fingerprint of no glyph", f.Name())
		}
	}
}
