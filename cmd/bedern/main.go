// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Command bedern generates the Bedern fonts.
//
// Usage:
//
//	bedern [-o DIR] [-tex TREE]
//	bedern -private-use
//
// The command writes the font files Bedern-Regular.otf, Bedern-Bold.otf,
// Bedern-Italic.otf and BedernMath-Regular.otf, the file
// bedern-math-ranges.tex and the licence OFL.txt into the directory DIR, which
// is the current directory by default.
// BedernMath-Regular.otf is the mathematics font for unicode-math, and
// bedern-math-ranges.tex defines the list of its code points for the option
// range of \setmathfont and the command \setBedernMath, which loads the font
// with that list.  If the environment variable SOURCE_DATE_EPOCH holds a time in
// seconds since 1970, the fonts carry that time instead of the time of the
// run, and the unique identifier of each font ends in the date of that time;
// otherwise the unique identifier is the full name and the version.
//
// With -tex the command also writes the files with which pdfTeX uses the
// fonts, below the directory TREE in the layout of the TeX Directory
// Structure: the fonts in the PFB format and as OpenType fonts, the metric
// files of the encodings T1 and TS1, the encoding files, the map file, the
// font definition files, which declare the shapes n, sl and it, and the
// package bedern for LaTeX, the macro file
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
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"seehuhn.de/go/bedern/internal/build"
	bedern "seehuhn.de/go/bedern/internal/font"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
	"seehuhn.de/go/bedern/internal/font/mathfont"
	"seehuhn.de/go/bedern/internal/kern"
	"seehuhn.de/go/bedern/internal/tex"
	"seehuhn.de/go/sfnt"
)

// A fontSpec is one font of the family: the name of its file, its parameters,
// its style and its series in LaTeX.  The glyphs of the font with italic set
// are those of the italic (assembly), with their own kerning; the parameters
// are then those of that kerning.
type fontSpec struct {
	file   string
	params func() design.Params
	style  func() design.Style
	series string
	italic bool
}

// fonts lists the fonts which the command writes.
var fonts = []fontSpec{
	{"Bedern-Regular.otf", design.Regular, design.RegularStyle, "m", false},
	{"Bedern-Bold.otf", design.Bold, design.BoldStyle, "b", false},
	{assembly.File, assembly.Kerning, design.ItalicStyle, "m", true},
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
	var (
		math    map[string][]byte
		mathErr error
		wg      sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		math, mathErr = renderMath(now, dated)
	}()
	out, err := renderAll(now, dated, tree != "")
	wg.Wait()
	if err != nil {
		return err
	}
	if mathErr != nil {
		return mathErr
	}
	return write(dir, tree, out, math, now)
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

// write writes the font files, the files of the mathematics font and the
// licence into dir and, if tree is not empty, the files for TeX below tree.
func write(dir, tree string, out []*rendered, math map[string][]byte, now time.Time) error {
	files := map[string][]byte{bedern.LicenseFile: []byte(bedern.OFL)}
	maps.Copy(files, math)
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
		t := &tex.Font{Style: f.style(), Series: f.series, Glyphs: out[i].tex}
		if f.italic {
			t.Angle, t.Correction = assembly.Default().Slant(), assembly.TeXItalicCorrection
		}
		tf = append(tf, t)
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

// mathGlyphs returns the glyphs of Bedern Math.  Tests replace it by a
// function which shares the builds.
var mathGlyphs = mathfont.Glyphs

// renderMath builds Bedern Math and returns the font file and the file of the
// ranges of its code points, by name.  If dated is false, the unique
// identifier carries no date.
func renderMath(now time.Time, dated bool) (map[string][]byte, error) {
	p := design.Regular()
	gs, err := mathGlyphs(p)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mathfont.File, err)
	}
	font, err := mathfont.Font(p, gs, now)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mathfont.File, err)
	}
	var buf bytes.Buffer
	if _, err := build.Write(&buf, font); err != nil {
		return nil, fmt.Errorf("%s: %w", mathfont.File, err)
	}
	otf := buf.Bytes()
	if !dated {
		if otf, err = build.Undated(otf); err != nil {
			return nil, fmt.Errorf("%s: %w", mathfont.File, err)
		}
	}
	var ranges bytes.Buffer
	ranges.WriteString(tex.Header("%", mathfont.RangesFile, "the code points of Bedern Math for unicode-math"))
	if err := mathfont.WriteRanges(&ranges, gs); err != nil {
		return nil, fmt.Errorf("%s: %w", mathfont.RangesFile, err)
	}
	return map[string][]byte{mathfont.File: otf, mathfont.RangesFile: ranges.Bytes()}, nil
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

	// tex holds the glyphs as the files for TeX need them (tex.ForTeX).
	tex []glyphs.Glyph
}

// render builds the font f, with the PFB file if withPFB is set.  If dated
// is false, the unique identifier carries no date.
func render(f fontSpec, now time.Time, dated, withPFB bool) (*rendered, error) {
	p := f.params()
	r := &rendered{params: p}
	var (
		font *sfnt.Font
		err  error
	)
	if f.italic {
		ip := assembly.Default()
		if r.glyphs, err = assembly.Table(ip); err != nil {
			return nil, err
		}
		r.kern = assembly.KernTable(r.glyphs)
		if font, err = assembly.FontOf(ip, r.glyphs, r.kern, now); err != nil {
			return nil, err
		}
	} else {
		r.glyphs = glyphSet(p)
		r.kern = kern.Compute(p, r.glyphs)
		if font, err = build.FontWith(p, f.style(), r.glyphs, now, build.Options{Kern: r.kern}); err != nil {
			return nil, err
		}
	}
	r.tex = tex.ForTeX(p, r.glyphs)
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
		if r.pfb, err = build.PFB(p, f.style(), r.tex, now); err != nil {
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
