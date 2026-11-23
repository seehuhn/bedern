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
)

// TestReproducible renders the font twice at the same time, which gives the
// same bytes, and, without -short, once at another time, which gives other
// bytes.
func TestReproducible(t *testing.T) {
	t.Parallel()
	secs := []int64{1700000000, 1700000000}
	if !testing.Short() {
		secs = append(secs, 1700000000+3*86400)
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
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(out[0], out[1]) {
		t.Error("same time gives different bytes")
	}
	if len(out) > 2 && bytes.Equal(out[0], out[2]) {
		t.Error("different time gives identical bytes")
	}
}

// TestRun checks that run writes the font file and the licence, and nothing
// else, into a directory which it creates.
func TestRun(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "fonts")
	if err := run(dir, time.Unix(1700000000, 0).UTC()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if want := []string{fontFile, bedern.LicenseFile}; !slices.Equal(names, want) {
		t.Errorf("files %v, want %v", names, want)
	}
	licence, err := os.ReadFile(filepath.Join(dir, bedern.LicenseFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(licence) != bedern.OFL {
		t.Error("OFL.txt differs from the embedded licence")
	}
	data, err := os.ReadFile(filepath.Join(dir, fontFile))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("OTTO")) {
		t.Errorf("%s does not start with the tag of a CFF font", fontFile)
	}
}
