// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Command bedern generates the Bedern fonts.
//
// Usage:
//
//	bedern [-o DIR]
//
// The command writes the font files Bedern-Regular.otf and Bedern-Bold.otf and
// the licence OFL.txt into the directory DIR, which is the current directory
// by default.  If the environment variable SOURCE_DATE_EPOCH holds a time in
// seconds since 1970, the fonts carry that time instead of the time of the
// run, and the unique identifier of each font ends in the date of that time;
// otherwise the unique identifier is the full name and the version.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// A fontSpec is one font of the family: the name of its file, its parameters
// and its style.
type fontSpec struct {
	file   string
	params func() design.Params
	style  func() design.Style
}

// fonts lists the fonts which the command writes.
var fonts = []fontSpec{
	{"Bedern-Regular.otf", design.Regular, design.RegularStyle},
	{"Bedern-Bold.otf", design.Bold, design.BoldStyle},
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "usage: bedern [-o DIR]\n")
	flag.PrintDefaults()
}

func main() {
	dir := flag.String("o", ".", "write the font files into `DIR`")
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("bedern: ")
	if flag.NArg() > 0 {
		usage()
		os.Exit(2)
	}
	epoch := os.Getenv("SOURCE_DATE_EPOCH")
	now, err := build.SourceDate(epoch, time.Now())
	if err != nil {
		log.Print(err)
		os.Exit(2)
	}
	if err := run(*dir, now, epoch != ""); err != nil {
		log.Fatal(err)
	}
}

// run writes the font files and the licence into dir.  The fonts are built
// side by side.  If dated is false, the unique identifiers carry no date.
func run(dir string, now time.Time, dated bool) error {
	bufs := make([]bytes.Buffer, len(fonts))
	errs := make([]error, len(fonts))
	var wg sync.WaitGroup
	for i, f := range fonts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = renderFont(&bufs[i], f, now, dated)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return fmt.Errorf("%s: %w", fonts[i].file, err)
		}
	}
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	for i, f := range fonts {
		if err := os.WriteFile(filepath.Join(dir, f.file), bufs[i].Bytes(), 0o666); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(dir, bedern.LicenseFile), []byte(bedern.OFL), 0o666)
}

// glyphSet returns the glyphs of a parameter set.  Tests replace it by a
// function which shares the builds.
var glyphSet = glyphs.Set

// renderFont builds the font f and writes its bytes to w.  If dated is
// false, the unique identifier carries no date.
func renderFont(w io.Writer, f fontSpec, now time.Time, dated bool) error {
	p := f.params()
	font, err := build.Font(p, f.style(), glyphSet(p), now)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err := font.Write(&buf); err != nil {
		return err
	}
	data := buf.Bytes()
	if !dated {
		if data, err = build.Undated(data); err != nil {
			return err
		}
	}
	_, err = w.Write(data)
	return err
}
