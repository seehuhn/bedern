// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Command bedern generates the Bedern fonts.
//
// Usage:
//
//	bedern [-o DIR] [-tex TREE]
//	bedern -private-use
//
// The command writes the font files Bedern-Regular.otf and Bedern-Bold.otf and
// the licence OFL.txt into the directory DIR, which is the current directory
// by default.  If the environment variable SOURCE_DATE_EPOCH holds a time in
// seconds since 1970, the fonts carry that time instead of the time of the
// run, and the unique identifier of each font ends in the date of that time;
// otherwise the unique identifier is the full name and the version.
//
// With -tex the command also writes the files with which pdfTeX uses the
// fonts, below the directory TREE in the layout of the TeX Directory
// Structure: the fonts in the PFB format and as OpenType fonts, the metric
// files of the encodings T1 and TS1, the encoding files, the map file, the
// font definition files and the package bedern for LaTeX, the macro file
// bedern.tex for plain TeX, the mappings of the glyphs to Unicode and the
// licence.
//
// With -private-use the command writes no files and prints the table of the
// glyphs in the Private Use Area: one line for each glyph with the code
// point, the glyph name, the TeX command which selects the glyph and a
// description, separated by tabs.
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
	"seehuhn.de/go/bedern/internal/kern"
	"seehuhn.de/go/bedern/internal/tex"
)

// A fontSpec is one font of the family: the name of its file, its parameters,
// its style and its series in LaTeX.
type fontSpec struct {
	file   string
	params func() design.Params
	style  func() design.Style
	series string
}

// fonts lists the fonts which the command writes.
var fonts = []fontSpec{
	{"Bedern-Regular.otf", design.Regular, design.RegularStyle, "m"},
	{"Bedern-Bold.otf", design.Bold, design.BoldStyle, "b"},
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "usage: bedern [-o DIR] [-tex TREE]\n       bedern -private-use\n")
	flag.PrintDefaults()
}

func main() {
	dir := flag.String("o", ".", "write the font files into `DIR`")
	tree := flag.String("tex", "", "write the files for TeX below `TREE`")
	table := flag.Bool("private-use", false, "print the table of the glyphs in the Private Use Area and write no files")
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("bedern: ")
	if flag.NArg() > 0 {
		usage()
		os.Exit(2)
	}
	if *table {
		if err := glyphs.WritePrivateUse(os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}
	epoch := os.Getenv("SOURCE_DATE_EPOCH")
	now, err := build.SourceDate(epoch, time.Now())
	if err != nil {
		log.Print(err)
		os.Exit(2)
	}
	if err := run(*dir, *tree, now, epoch != ""); err != nil {
		log.Fatal(err)
	}
}

// run writes the font files and the licence into dir and, if tree is not
// empty, the files for TeX below tree.  If dated is false, the unique
// identifiers carry no date.
func run(dir, tree string, now time.Time, dated bool) error {
	out, err := renderAll(now, dated, tree != "")
	if err != nil {
		return err
	}
	return write(dir, tree, out, now)
}

// renderAll builds the fonts side by side, with the PFB files if withPFB is
// set.
func renderAll(now time.Time, dated, withPFB bool) ([]*rendered, error) {
	out := make([]*rendered, len(fonts))
	errs := make([]error, len(fonts))
	var wg sync.WaitGroup
	for i, f := range fonts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i], errs[i] = render(f, now, dated, withPFB)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fonts[i].file, err)
		}
	}
	return out, nil
}

// write writes the font files and the licence into dir and, if tree is not
// empty, the files for TeX below tree.
func write(dir, tree string, out []*rendered, now time.Time) error {
	files := map[string][]byte{bedern.LicenseFile: []byte(bedern.OFL)}
	for i, f := range fonts {
		files[f.file] = out[i].otf
	}
	if err := writeFiles(dir, files); err != nil {
		return err
	}
	if tree == "" {
		return nil
	}
	files, _, err := texFiles(out, now)
	if err != nil {
		return err
	}
	return writeFiles(tree, files)
}

// writeFiles writes the files, whose names are paths with slashes, below dir.
func writeFiles(dir string, files map[string][]byte) error {
	for name, data := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0o666); err != nil {
			return err
		}
	}
	return nil
}

// texFiles returns the files for TeX of the rendered fonts, as a map from
// the path of each file in the tree to its content, and the reports on the
// metric files.
func texFiles(out []*rendered, now time.Time) (map[string][]byte, []tex.Report, error) {
	var (
		tf     []*tex.Font
		params []design.Params
		kerns  []tex.KernFunc
	)
	for i, f := range fonts {
		tf = append(tf, &tex.Font{Style: f.style(), Series: f.series, Glyphs: out[i].glyphs})
		params = append(params, out[i].params)
		kerns = append(kerns, out[i].kern.Kern)
	}
	files, reports, err := tex.Files(tf, params, kerns, now)
	if err != nil {
		return nil, nil, err
	}
	for i, f := range fonts {
		files[tex.DirOpenType+"/"+f.file] = out[i].otf
		files[tex.DirPFB+"/"+tf[i].PFB()] = out[i].pfb
	}
	files[tex.DirDoc+"/"+bedern.LicenseFile] = []byte(bedern.OFL)
	return files, reports, nil
}

// glyphSet returns the glyphs of a parameter set.  Tests replace it by a
// function which shares the builds.
var glyphSet = glyphs.Set

// A rendered font holds what is built for one font: the OpenType file, the
// PFB file if it is asked for, and the parameters, the glyphs and the kerns
// from which they are made.
type rendered struct {
	otf, pfb []byte
	params   design.Params
	glyphs   []glyphs.Glyph
	kern     *kern.Table
}

// render builds the font f, with the PFB file if withPFB is set.  If dated
// is false, the unique identifier carries no date.
func render(f fontSpec, now time.Time, dated, withPFB bool) (*rendered, error) {
	p := f.params()
	r := &rendered{params: p, glyphs: glyphSet(p)}
	r.kern = kern.Compute(p, r.glyphs)
	font, err := build.FontWith(p, f.style(), r.glyphs, now, build.Options{Kern: r.kern})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if _, err := build.Write(&buf, font); err != nil {
		return nil, err
	}
	r.otf = buf.Bytes()
	if !dated {
		if r.otf, err = build.Undated(r.otf); err != nil {
			return nil, err
		}
	}
	if withPFB {
		if r.pfb, err = build.PFB(p, f.style(), r.glyphs, now); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// renderFont builds the font f and writes its bytes to w.  If dated is
// false, the unique identifier carries no date.
func renderFont(w io.Writer, f fontSpec, now time.Time, dated bool) error {
	r, err := render(f, now, dated, false)
	if err != nil {
		return err
	}
	_, err = w.Write(r.otf)
	return err
}
