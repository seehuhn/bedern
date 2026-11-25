# Bedern

Bedern is a serif text typeface for use with TeX, inspired by Computer Modern.  The family consists of Bedern Regular and Bedern Bold, with the mathematics font Bedern Math.  The fonts are work in progress, and their shapes and metrics are still subject to change.  The Go programs in this repository generate the font files.

## Generating the Fonts

The command `bedern` writes the font files into a directory.  It is run with

    go run seehuhn.de/go/bedern/cmd/bedern@latest

which writes Bedern Regular as `Bedern-Regular.otf`, Bedern Bold as `Bedern-Bold.otf`, Bedern Math as `BedernMath-Regular.otf`, the file `bedern-math-ranges.tex` and the licence `OFL.txt` into the current directory.  Bedern Math is an OpenType mathematics font for LuaLaTeX and XeLaTeX with the package `unicode-math`; the file `bedern-math-ranges.tex` defines the list `\BedernMathRanges` of its code points for the option `range` of `\setmathfont` and the command `\setBedernMath`, which loads the font with that list.  The option `-o DIR` names another directory.  The option `-tex TREE` also writes the files with which pdfTeX uses the fonts, below the directory `TREE` in the layout of the TeX Directory Structure: the fonts in the PFB format, the metric files of the encodings T1 and TS1, the encoding files, the map file `bedern.map`, the package `bedern` for LaTeX and the macro file `bedern.tex` for plain TeX.  If the environment variable `SOURCE_DATE_EPOCH` holds a time in seconds since 1970, the fonts carry that time instead of the time of the run, and two runs give identical files.  With the option `-private-use` the command writes no files and prints the table of the glyphs in the Private Use Area: the code point, the glyph name, the TeX command and a description of each glyph.

## Licences

The code, the fonts and the files for TeX are released under three licences.  The code is released under the BSD 3-Clause licence, see `LICENSE`.  The font sources in the directory `internal/font` and the generated fonts in every format (the OpenType files, the PFB files and the metric files) are released under the SIL Open Font License, Version 1.1, with the Reserved Font Name Bedern, see `internal/font/OFL.txt`.  The package `bedern.sty`, the font definition files, the macro files `bedern.tex` and `bedern-glyphtounicode.tex`, the encoding files and the map file, which `-tex` writes, are released under the LaTeX Project Public License, version 1.3c or later, with the status maintained; each file carries the notice of the licence, whose text is at https://www.latex-project.org/lppl.txt.
