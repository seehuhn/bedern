// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// The tests share the builds of the glyphs.
func init() { glyphSet = glyphstest.Set }

// runTime is the time at which the tests render the font.
const runTime = 1700000000

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
		if runFailed = run(dir, time.Unix(runTime, 0).UTC()); runFailed != nil {
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

// TestReproducible renders the font at the time at which run wrote its file,
// which gives the same bytes, and, without -short, at another time, which
// gives other bytes.
func TestReproducible(t *testing.T) {
	t.Parallel()
	secs := []int64{runTime}
	if !testing.Short() {
		secs = append(secs, runTime+3*86400)
	}
	out := make([][]byte, len(secs))
	errs := make([]error, len(secs))
	var wg sync.WaitGroup
	for i, sec := range secs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			errs[i] = renderFont(&buf, time.Unix(sec, 0).UTC())
			out[i] = buf.Bytes()
		}()
	}
	written := ran(t).files[fontFile]
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(out[0], written) {
		t.Error("same time gives different bytes")
	}
	if len(out) > 1 && bytes.Equal(out[0], out[1]) {
		t.Error("different time gives identical bytes")
	}
}

// TestRun checks that run writes the font file and the licence, and nothing
// else, into a directory which it creates.
func TestRun(t *testing.T) {
	t.Parallel()
	r := ran(t)
	if want := []string{fontFile, bedern.LicenseFile}; !slices.Equal(r.names, want) {
		t.Errorf("files %v, want %v", r.names, want)
	}
	if string(r.files[bedern.LicenseFile]) != bedern.OFL {
		t.Error("OFL.txt differs from the embedded licence")
	}
	if !bytes.HasPrefix(r.files[fontFile], []byte("OTTO")) {
		t.Errorf("%s does not start with the tag of a CFF font", fontFile)
	}
}
