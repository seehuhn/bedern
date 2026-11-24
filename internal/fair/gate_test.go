// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/diag"
	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

var updateFair = flag.Bool("update-fair", false, "rewrite the baseline files of testdata/baseline: keep the judged entries which still occur, drop the others, add new findings as unjudged defects")

// fairDir holds the baseline of the fairness gate, one file for each glyph
// which has findings, named by diag.FileName.
const fairDir = "testdata/baseline"

// TestFair is the fairness gate.  It analyses the rounded outline of every
// glyph of the regular parameters with internal/fair and compares the
// findings with the baseline in testdata/baseline.  A glyph which is a scaled
// copy of a construction (glyphs.AnalysisScale) is analysed with thresholds
// scaled by the factor.  It fails on a finding which
// is not in the baseline, on a baseline entry which no longer occurs (so that
// the baseline only shrinks), and on an entry which is not judged (whose
// reason is "?").  `go test ./internal/fair -run TestFair -update-fair`
// rewrites the files: judged entries are kept, new findings are added as
// "defect: ?", which must then be judged by hand, as intended with a reason
// or as a defect with what is wrong.
func TestFair(t *testing.T) {
	t.Parallel()
	gs := glyphs.Set(design.Regular())
	results := make([][]fair.Finding, len(gs))
	gate := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for i, g := range gs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate <- struct{}{}
			defer func() { <-gate }()
			results[i] = fair.Analyse(g.Name, g.Path, fair.Default.Scaled(glyphs.AnalysisScale(design.Regular(), g.Name))).Findings
		}()
	}
	wg.Wait()

	files := map[string]bool{}
	var problems []string
	for i, g := range gs {
		name := diag.FileName(g.Name) + ".txt"
		path := filepath.Join(fairDir, name)
		var es []fair.Entry
		if data, err := os.ReadFile(path); err == nil {
			es, err = fair.ParseBaseline(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		fs := results[i]
		for _, b := range []string{g.Base, g.Like} {
			// A letter composed of a base and a mark has the findings of its
			// base, and the dotless i and j those of i and j, which the
			// baseline of those glyphs judges: only the others are compared
			// with its own baseline.
			if b != "" {
				fs = withoutBase(t, b, fs)
			}
		}
		if len(fs) > 0 || len(es) > 0 {
			files[name] = true
		}
		if *updateFair {
			writeFair(t, path, g.Name, fair.UpdateBaseline(fs, es))
			continue
		}
		newFs, gone, _ := fair.Compare(fs, es)
		for _, f := range newFs {
			problems = append(problems, fmt.Sprintf("%s: new finding %s at (%.0f, %.0f), edge %s: %s",
				g.Name, f.Kind, f.X, f.Y, fair.Compass(f.Dir), f.Detail))
		}
		for _, en := range gone {
			problems = append(problems, fmt.Sprintf("%s: the baseline entry %q no longer occurs; remove it from %s (or run with -update-fair)",
				g.Name, en.String(), path))
		}
		for _, en := range es {
			if en.Reason == fair.Unjudged {
				problems = append(problems, fmt.Sprintf("%s: the baseline entry %q is not judged; write intended or defect and a reason in %s",
					g.Name, en.String(), path))
			}
		}
	}
	entries, err := os.ReadDir(fairDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, en := range entries {
		if files[en.Name()] {
			continue
		}
		if *updateFair {
			if err := os.Remove(filepath.Join(fairDir, en.Name())); err != nil {
				t.Fatal(err)
			}
			continue
		}
		problems = append(problems, fmt.Sprintf("%s/%s belongs to no glyph with findings; remove it", fairDir, en.Name()))
	}
	sort.Strings(problems)
	if len(problems) > 0 {
		t.Errorf("the fairness of the outlines differs from the baseline in %s:\n%s",
			fairDir, strings.Join(problems, "\n"))
	}
}

// withoutBase returns the findings which the baseline entries of the glyph
// base do not match.
func withoutBase(t *testing.T, base string, fs []fair.Finding) []fair.Finding {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fairDir, diag.FileName(base)+".txt"))
	if os.IsNotExist(err) {
		return fs
	} else if err != nil {
		t.Fatal(err)
	}
	es, err := fair.ParseBaseline(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	newFs, _, _ := fair.Compare(fs, es)
	return newFs
}

// writeFair writes the baseline file of a glyph, or removes it when there
// is nothing to write.
func writeFair(t *testing.T, path, glyph string, es []fair.Entry) {
	t.Helper()
	if len(es) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		return
	}
	var buf bytes.Buffer
	if err := fair.FormatBaseline(&buf, glyph, es); err != nil {
		t.Fatal(err)
	}
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, buf.Bytes()) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o666); err != nil {
		t.Fatal(err)
	}
}
