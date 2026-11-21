// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Command bedern generates the Bedern fonts.
//
// Usage:
//
//	bedern [-o DIR]
//
// The command writes the font file Bedern-Regular.otf and the licence OFL.txt
// into the directory DIR, which is the current directory by default.  If the
// environment variable SOURCE_DATE_EPOCH holds a time in seconds since 1970,
// the font carries that time instead of the time of the run.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// fontFile is the name of the font file.
const fontFile = "Bedern-Regular.otf"

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
	now, err := build.SourceDate(os.Getenv("SOURCE_DATE_EPOCH"), time.Now())
	if err != nil {
		log.Print(err)
		os.Exit(2)
	}
	if err := run(*dir, now); err != nil {
		log.Fatal(err)
	}
}

// run writes the font file and the licence into dir.
func run(dir string, now time.Time) error {
	var buf bytes.Buffer
	if err := renderFont(&buf, now); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, fontFile), buf.Bytes(), 0o666); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, bedern.LicenseFile), []byte(bedern.OFL), 0o666)
}

// renderFont builds the regular font and writes its bytes to w.
func renderFont(w io.Writer, now time.Time) error {
	p := design.Regular()
	font, err := build.Font(p, design.RegularStyle(), glyphs.Set(p), now)
	if err != nil {
		return err
	}
	_, err = font.Write(w)
	return err
}
