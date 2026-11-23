// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

//go:generate go test -run TestUpdate -update

// update makes TestUpdate write the files of testdata anew.
var update = flag.Bool("update", false, "write the files of testdata anew from the answers of the installed TFtoPL and PLtoTF")

// golden returns the content of a file of testdata.
func golden(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("no answer of the converters for %s (go generate writes the files): %v", name, err)
	}
	return data
}

// text returns the PL text of a file.
func text(t testing.TB, f *File) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := f.WritePL(&buf); err != nil {
		t.Fatalf("writing the PL text failed: %v", err)
	}
	return buf.Bytes()
}

// description returns the PL text of a description.
func description(t testing.TB, f *Font) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := f.WritePL(&buf); err != nil {
		t.Fatalf("writing the description failed: %v", err)
	}
	return buf.Bytes()
}

// TestUpdate writes the files of testdata: for every font of the tests the
// TFM file which PLtoTF makes of its description and the PL text which TFtoPL
// makes of that file, for every odd file the PL text of TFtoPL and for every
// PL text of the tests the TFM file of PLtoTF.  It runs with the flag -update
// only.
func TestUpdate(t *testing.T) {
	if !*update {
		t.Skip("runs with -update only")
	}
	cv := &converters{}
	var err error
	if cv.tftopl, err = exec.LookPath("tftopl"); err != nil {
		t.Fatal(err)
	}
	if cv.pltotf, err = exec.LookPath("pltotf"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write := func(name string, data []byte) {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join("testdata", name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	toPL := func(name string, data []byte, wantOK bool) []byte {
		path := filepath.Join(dir, "in.tfm")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		pl, messages, ok, err := cv.toPL(path)
		if err != nil || ok != wantOK {
			t.Fatalf("%s: TFtoPL: %v, success %v, %s", name, err, ok, messages)
		}
		return pl
	}
	for _, tf := range testFonts() {
		data, messages, ok, err := cv.toTFM(dir, description(t, tf.font))
		if err != nil || !ok {
			t.Fatalf("%s: PLtoTF: %v %s", tf.name, err, messages)
		}
		write(tf.name+".tfm", data)
		write(tf.name+".pl", toPL(tf.name, data, true))
	}
	for _, of := range oddFiles(t) {
		write("odd-"+of.name+".pl", toPL(of.name, of.file.bytes(), !of.gives))
	}
	for _, src := range plSources {
		data, messages, ok, err := cv.toTFM(dir, []byte(src.text))
		if err != nil || !ok {
			t.Fatalf("%s: PLtoTF: %v %s", src.name, err, messages)
		}
		write("pl-"+src.name+".tfm", data)
	}
}
