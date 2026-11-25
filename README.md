# Bedern

Bedern is a serif text typeface for use with TeX, inspired by Computer Modern.  The family consists of Bedern Regular, Bedern Italic and Bedern Bold, with the mathematics font Bedern Math.  The fonts are work in progress, and their shapes and metrics are still subject to change.  The Go programs in this repository generate the font files.

## Generating the Fonts

The command `bedern` writes the font files into a directory.  It is run with

    go run seehuhn.de/go/bedern/cmd/bedern@latest

which writes Bedern Regular as `Bedern-Regular.otf`, Bedern Bold as `Bedern-Bold.otf`, Bedern Italic as `Bedern-Italic.otf`, Bedern Math as `BedernMath-Regular.otf`, the file `bedern-math-ranges.tex` and the licence `OFL.txt` into the current directory.  Bedern Math is an OpenType mathematics font for LuaLaTeX and XeLaTeX with the package `unicode-math`; the file `bedern-math-ranges.tex` defines the list `\BedernMathRanges` of its code points for the option `range` of `\setmathfont` and the command `\setBedernMath`, which loads the font with that list.  The option `-o DIR` names another directory.  The option `-tex TREE` also writes the files with which pdfTeX uses the fonts, below the directory `TREE` in the layout of the TeX Directory Structure: the fonts in the PFB format, the metric files of the encodings T1 and TS1, the encoding files, the map file `bedern.map`, the font definition files, which declare the shapes `n`, `sl` and `it` of the family, the package `bedern` for LaTeX and the macro file `bedern.tex` for plain TeX.  With the package `bedern`, pdfLaTeX sets `\textit` and `\emph` in Bedern Italic and `\textsl` in the slanted Bedern Regular, and plain pdfTeX sets `\it` in Bedern Italic.  If the environment variable `SOURCE_DATE_EPOCH` holds a time in seconds since 1970, the fonts carry that time instead of the time of the run, and two runs give identical files.  With the option `-private-use` the command writes no files and prints the table of the glyphs in the Private Use Area: the code point, the glyph name, the TeX command and a description of each glyph.

## Using the Fonts with LuaLaTeX

This section shows a document which sets text in Bedern and mathematics in Bedern Math with LuaLaTeX.  The font files written by `bedern` must be where LuaLaTeX finds them, for example in the directory of the document.

    \documentclass{article}
    \usepackage{amsmath}
    \usepackage{fontspec}
    \usepackage{unicode-math}
    \setmainfont{Bedern-Regular.otf}[ItalicFont=Bedern-Italic.otf, BoldFont=Bedern-Bold.otf]
    \setmathfont{OtherMath.otf}
    \input{bedern-math-ranges}
    \setBedernMath
    \begin{document}
    \[ \symbfit{v} = (v_1,\dots,v_n), \qquad
       \int_0^1 \frac{\sin x}{x}\,dx \le \sqrt{\sum_{i=1}^n x_i^2} \]
    \end{document}

The first `\setmathfont` loads another OpenType maths font, written `OtherMath.otf` here, which supplies every code point that Bedern Math does not.  The command `\setBedernMath` loads Bedern Math after it with `range=\BedernMathRanges`, so that the code points of Bedern Math are taken from it and all others from the first font.  The call `\setmathfont{BedernMath-Regular.otf}[range=\BedernMathRanges]` is equivalent, but `\setBedernMath` also sets the minus sign, the asterisk, the colon and the prime from Bedern Math.

Bold vectors are written `\symbfit{v}` for bold italic and `\symbfup{A}` for bold upright letters; they are set in the first maths font.  The commands `\boldsymbol` and `\bm` do not give bold letters with `unicode-math`.

## Licences

The code, the fonts and the files for TeX are released under three licences.  The code is released under the BSD 3-Clause licence, see `LICENSE`.  The font sources in the directory `internal/font` and the generated fonts in every format (the OpenType files, the PFB files and the metric files) are released under the SIL Open Font License, Version 1.1, with the Reserved Font Name Bedern, see `internal/font/OFL.txt`.  The package `bedern.sty`, the font definition files, the macro files `bedern.tex` and `bedern-glyphtounicode.tex`, the file `bedern-math-ranges.tex`, the encoding files and the map file, which the command writes, are released under the LaTeX Project Public License, version 1.3c or later, with the status maintained; each file carries the notice of the licence, whose text is at https://www.latex-project.org/lppl.txt.
