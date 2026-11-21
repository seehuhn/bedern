# Bedern

Bedern is a serif text typeface for use with TeX, inspired by Computer Modern.  The Go programs in this repository generate the font files.

## Generating the Fonts

The command `bedern` writes the font files into a directory.  It is run with

    go run seehuhn.de/go/bedern/cmd/bedern@latest

which writes `Bedern-Regular.otf` and the licence `OFL.txt` into the current directory.  The option `-o DIR` names another directory.  If the environment variable `SOURCE_DATE_EPOCH` holds a time in seconds since 1970, the font carries that time instead of the time of the run, and two runs give identical files.

## Licences

The code and the fonts are released under two licences.  The code is released under the BSD 3-Clause licence, see `LICENSE`.  The font sources in the directory `internal/font` and the generated fonts are released under the SIL Open Font License, Version 1.1, with the Reserved Font Name Bedern, see `internal/font/OFL.txt`.
