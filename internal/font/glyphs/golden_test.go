// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

var update = flag.Bool("update", false, "rewrite the files of testdata/golden")

// goldenDir holds one file per glyph, named by goldenFileName, with one line for
// each parameter set.
const goldenDir = "testdata/golden"

// goldenFileName maps a glyph name to the stem of its file name: a capital
// letter is followed by an underscore, one leading dot is dropped, and every
// character other than a to z, the digits, "-" and a "." which is not at the
// start is written as % and the two hexadecimal digits of each byte.
func goldenFileName(glyph string) string {
	name := strings.TrimPrefix(glyph, ".")
	var sb strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case 'a' <= c && c <= 'z', '0' <= c && c <= '9', c == '-':
			sb.WriteByte(c)
		case 'A' <= c && c <= 'Z':
			sb.WriteByte(c)
			sb.WriteByte('_')
		case c == '.' && i > 0:
			sb.WriteByte(c)
		default:
			fmt.Fprintf(&sb, "%%%02x", c)
		}
	}
	if sb.Len() == 0 {
		return "%"
	}
	return sb.String()
}

// fingerprint returns a short hash of the rounded outline, written as text.
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

// goldenLines returns one line per parameter set and glyph: the set, the
// glyph, the advance and the fingerprint of the rounded outline.  The glyphs
// are built in parallel and shared with the other tests.
func goldenLines(t *testing.T) []string {
	sets := goldenParams()
	parts := make([][]string, len(sets))
	type job struct{ i, j int }
	var jobs []job
	for i, np := range sets {
		parts[i] = make([]string, len(table))
		for j, e := range table {
			if len(np.glyphs) > 0 && !contains(np.glyphs, e.name) {
				continue
			}
			jobs = append(jobs, job{i, j})
		}
	}
	forEachParallel(len(jobs), func(k int) {
		i, j := jobs[k].i, jobs[k].j
		g := cachedRounded(sets[i], table[j].name)
		parts[i][j] = fmt.Sprintf("%s %s %g %s", sets[i].name, g.Name, g.Advance, fingerprint(g.Path))
	})
	var all []string
	for _, l := range parts {
		for _, s := range l {
			if s != "" {
				all = append(all, s)
			}
		}
	}
	return all
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// goldenFiles groups the lines by glyph and returns the content of each file
// of goldenDir, by file name.
func goldenFiles(lines []string) map[string]string {
	files := map[string]string{}
	for _, l := range lines {
		f := strings.SplitN(l, " ", 3)
		name := goldenFileName(f[1]) + ".txt"
		files[name] += l + "\n"
	}
	return files
}

// TestGolden compares the advance and the rounded outline of every glyph in
// every variant and parameter set, and in the full tier on the bold
// parameters, with the files of testdata/golden.  The routine tier leaves out
// the lines of the bold.
// An intended change is accepted with `go test ./internal/glyphs -run
// TestGolden -update`, which rewrites the files whose content changes and
// removes the files of glyphs which are gone.  The test fails on a file which belongs to no glyph.
func TestGolden(t *testing.T) {
	t.Parallel()
	if *update && testing.Short() {
		t.Fatal("-update needs the full tier: run it without -short")
	}
	got := goldenLines(t)
	files := goldenFiles(got)
	sets := map[string]bool{}
	for _, np := range goldenParams() {
		sets[np.name] = true
	}
	entries, err := os.ReadDir(goldenDir)
	if err != nil && !(*update && os.IsNotExist(err)) {
		t.Fatalf("%v (create the files with -update)", err)
	}
	if *update {
		if err := os.MkdirAll(goldenDir, 0o777); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			path := filepath.Join(goldenDir, name)
			if old, err := os.ReadFile(path); err == nil && string(old) == content {
				continue
			}
			if err := os.WriteFile(path, []byte(content), 0o666); err != nil {
				t.Fatal(err)
			}
		}
		for _, e := range entries {
			if _, ok := files[e.Name()]; !ok {
				if err := os.Remove(filepath.Join(goldenDir, e.Name())); err != nil {
					t.Fatal(err)
				}
			}
		}
		return
	}
	var want []string
	for _, e := range entries {
		if _, ok := files[e.Name()]; !ok {
			t.Errorf("%s/%s belongs to no glyph; remove it with -update", goldenDir, e.Name())
			continue
		}
		data, err := os.ReadFile(filepath.Join(goldenDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
			if set, _, _ := strings.Cut(l, " "); sets[set] {
				want = append(want, l)
			}
		}
	}
	if rep := goldenReport(want, got); rep != "" {
		t.Errorf("outlines or advances differ from the files of %s; accept with -update if intended:\n%s", goldenDir, rep)
	}
}

// goldenReport lists the glyphs which changed, each with the number of
// parameter sets in which it did, and the sets which came or went.
func goldenReport(want, got []string) string {
	index := func(lines []string) map[string]string {
		m := make(map[string]string, len(lines))
		for _, l := range lines {
			f := strings.SplitN(l, " ", 3)
			m[f[0]+" "+f[1]] = f[2]
		}
		return m
	}
	w, g := index(want), index(got)
	changed := map[string][]string{} // glyph -> sets
	sets := map[string]bool{}
	for k, v := range g {
		set, glyph, _ := strings.Cut(k, " ")
		sets[set] = true
		if old, ok := w[k]; !ok || old != v {
			changed[glyph] = append(changed[glyph], set)
		}
	}
	for k := range w {
		if _, ok := g[k]; !ok {
			set, glyph, _ := strings.Cut(k, " ")
			changed[glyph] = append(changed[glyph], set+" (missing)")
		}
	}
	var names []string
	for glyph := range changed {
		names = append(names, glyph)
	}
	sort.Strings(names)
	var sb strings.Builder
	for _, glyph := range names {
		ss := changed[glyph]
		sort.Strings(ss)
		ex := ss
		if len(ex) > 3 {
			ex = append(ex[:3:3], "...")
		}
		fmt.Fprintf(&sb, "  %s: %d of %d sets (%s)\n", glyph, len(ss), len(sets), strings.Join(ex, ", "))
	}
	return sb.String()
}
