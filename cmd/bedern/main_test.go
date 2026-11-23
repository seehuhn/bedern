// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
)

// The tests share the builds of the glyphs.
func init() { glyphSet = glyphstest.Set }

// runTime is the time at which the tests render the fonts, and runDate its
// date.
const (
	runTime = 1700000000
	runDate = "2023-11-14"
)

// version is the version of the fonts.
const version = "0.001"

// A runResult holds what run wrote into a directory which it created: the
// names of the files and the contents of each.
type runResult struct {
	names []string
	files map[string][]byte
}

var (
	runOnce   sync.Once
	runOut    runResult
	runFailed error
)

// ran returns what run, called once for all tests at runTime in a temporary
// directory of the first test which asks, has written.
func ran(t *testing.T) runResult {
	t.Helper()
	runOnce.Do(func() {
		dir := filepath.Join(t.TempDir(), "fonts")
		if runFailed = run(dir, time.Unix(runTime, 0).UTC(), true); runFailed != nil {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			runFailed = err
			return
		}
		runOut.files = map[string][]byte{}
		for _, e := range entries {
			runOut.names = append(runOut.names, e.Name())
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				runFailed = err
				return
			}
			runOut.files[e.Name()] = data
		}
	})
	if runFailed != nil {
		t.Fatal(runFailed)
	}
	return runOut
}

// TestReproducible renders the regular font at the time at which run wrote its
// file, which gives the same bytes.  Without -short it does the same for the
// bold font and renders the regular font at another time, which gives other
// bytes.
func TestReproducible(t *testing.T) {
	t.Parallel()
	type job struct {
		f    fontSpec
		sec  int64
		same bool
	}
	jobs := []job{{fonts[0], runTime, true}}
	if !testing.Short() {
		jobs = append(jobs, job{fonts[1], runTime, true}, job{fonts[0], runTime + 3*86400, false})
	}
	out := make([][]byte, len(jobs))
	errs := make([]error, len(jobs))
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			errs[i] = renderFont(&buf, j.f, time.Unix(j.sec, 0).UTC(), true)
			out[i] = buf.Bytes()
		}()
	}
	files := ran(t).files
	wg.Wait()
	for i, j := range jobs {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if same := bytes.Equal(out[i], files[j.f.file]); same != j.same {
			t.Errorf("%s at %d: identical bytes %t, want %t", j.f.file, j.sec, same, j.same)
		}
	}
}

// TestRun checks that run writes the font files and the licence, and nothing
// else, into a directory which it creates.
func TestRun(t *testing.T) {
	t.Parallel()
	r := ran(t)
	want := []string{bedern.LicenseFile}
	for _, f := range fonts {
		want = append(want, f.file)
	}
	slices.Sort(want)
	if !slices.Equal(r.names, want) {
		t.Errorf("files %v, want %v", r.names, want)
	}
	if string(r.files[bedern.LicenseFile]) != bedern.OFL {
		t.Error("OFL.txt differs from the embedded licence")
	}
	for _, f := range fonts {
		if !bytes.HasPrefix(r.files[f.file], []byte("OTTO")) {
			t.Errorf("%s does not start with the tag of a CFF font", f.file)
		}
	}
}

// familyFields are the fields of a font file which name it and its style.
type familyFields struct {
	names     map[string]*name.Table // the name tables, by platform and language
	weight    uint16                 // usWeightClass of the OS/2 table
	selection uint16                 // fsSelection of the OS/2 table
	macStyle  uint16                 // macStyle of the head table
}

// readFamilyFields reads the fields of familyFields from a font file.
func readFamilyFields(t *testing.T, file string, data []byte) familyFields {
	t.Helper()
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	table := func(tag string, n int) []byte {
		b, err := h.ReadTableBytes(r, tag)
		if err != nil {
			t.Fatalf("%s: %s: %v", file, tag, err)
		}
		if len(b) < n {
			t.Fatalf("%s: the table %s has %d bytes", file, tag, len(b))
		}
		return b
	}
	info, err := name.Decode(table("name", 6))
	if err != nil {
		t.Fatalf("%s: name: %v", file, err)
	}
	f := familyFields{names: map[string]*name.Table{}}
	for lang, tab := range info.Mac {
		f.names["Mac "+lang] = tab
	}
	for lang, tab := range info.Windows {
		f.names["Windows "+lang] = tab
	}
	os2 := table("OS/2", 64)
	f.weight = binary.BigEndian.Uint16(os2[4:])
	f.selection = binary.BigEndian.Uint16(os2[62:])
	f.macStyle = binary.BigEndian.Uint16(table("head", 46)[44:])
	return f
}

// TestFamily checks that the font files form one family: the same family
// name, version, copyright and licence, the subfamily, full name, PostScript
// name and unique identifier of each style, and the weight class, the bits
// of fsSelection (bold 0x20, regular 0x40, italic 0x01) and of macStyle (bold
// 0x01, italic 0x02).
func TestFamily(t *testing.T) {
	t.Parallel()
	files := ran(t).files
	want := []struct {
		file, sub, ps    string
		weight, sel, mac uint16
	}{
		{"Bedern-Regular.otf", "Regular", "Bedern-Regular", 400, 0x40, 0},
		{"Bedern-Bold.otf", "Bold", "Bedern-Bold", 700, 0x20, 0x01},
	}
	versions := map[string]bool{}
	ids := map[string]string{}
	for _, w := range want {
		f := readFamilyFields(t, w.file, files[w.file])
		if f.weight != w.weight {
			t.Errorf("%s: weight class %d, want %d", w.file, f.weight, w.weight)
		}
		if got := f.selection & 0x61; got != w.sel {
			t.Errorf("%s: fsSelection bits %#x, want %#x", w.file, got, w.sel)
		}
		if got := f.macStyle & 0x03; got != w.mac {
			t.Errorf("%s: macStyle bits %#x, want %#x", w.file, got, w.mac)
		}
		if len(f.names) < 2 {
			t.Errorf("%s: %d name tables, want one for Mac and one for Windows", w.file, len(f.names))
		}
		full := "Bedern " + w.sub
		for key, tab := range f.names {
			check := func(field, got, want string) {
				if got != want {
					t.Errorf("%s, %s: %s %q, want %q", w.file, key, field, got, want)
				}
			}
			check("family", tab.Family, "Bedern")
			check("subfamily", tab.Subfamily, w.sub)
			check("full name", tab.FullName, full)
			check("PostScript name", tab.PostScriptName, w.ps)
			check("typographic family", tab.TypographicFamily, "")
			check("typographic subfamily", tab.TypographicSubfamily, "")
			check("copyright", tab.Copyright, bedern.Copyright)
			check("licence", tab.License, bedern.License)
			check("licence URL", tab.LicenseURL, bedern.LicenseURL)
			check("version", tab.Version, "Version "+version)
			check("unique identifier", tab.Identifier, full+"; "+version+"; "+runDate)
			if other, ok := ids[tab.Identifier]; ok && other != w.file {
				t.Errorf("%s and %s share the unique identifier %q", w.file, other, tab.Identifier)
			}
			ids[tab.Identifier] = w.file
			versions[tab.Version] = true
		}
	}
	if len(versions) != 1 {
		t.Errorf("the files carry the versions %v, want one", versions)
	}
}

// TestUndated checks that the font files without a date of the build differ
// from the dated ones only in the unique identifier, which is the full name
// and the version.
func TestUndated(t *testing.T) {
	t.Parallel()
	files := ran(t).files
	for _, f := range fonts {
		dated := files[f.file]
		undated, err := build.Undated(dated)
		if err != nil {
			t.Fatal(err)
		}
		a, b := readTables(t, f.file, dated), readTables(t, f.file, undated)
		if len(a) != len(b) {
			t.Errorf("%s: %d tables, want %d", f.file, len(b), len(a))
		}
		for tag, data := range a {
			if tag != "name" && tag != "head" && !bytes.Equal(data, b[tag]) {
				t.Errorf("%s: the table %s differs", f.file, tag)
			}
		}
		// The head table differs in checkSumAdjustment only.
		ha, hb := a["head"], b["head"]
		if len(ha) != len(hb) || !bytes.Equal(ha[:8], hb[:8]) || !bytes.Equal(ha[12:], hb[12:]) {
			t.Errorf("%s: the table head differs beyond its checksum", f.file)
		}
		full := readFamilyFields(t, f.file, undated)
		want := readFamilyFields(t, f.file, dated)
		for key, tab := range full.names {
			w := *want.names[key]
			w.Identifier = f.style().Family + " " + f.style().Subfamily + "; " + version
			if !reflect.DeepEqual(*tab, w) {
				t.Errorf("%s, %s: name table %+v, want %+v", f.file, key, *tab, w)
			}
		}
	}
}

// readTables returns the tables of a font file by tag.
func readTables(t *testing.T, file string, data []byte) map[string][]byte {
	t.Helper()
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	out := map[string][]byte{}
	for tag := range h.Toc {
		b, err := h.ReadTableBytes(r, tag)
		if err != nil {
			t.Fatalf("%s: %s: %v", file, tag, err)
		}
		out[tag] = b
	}
	return out
}
