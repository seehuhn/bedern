// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// punctDots returns the diameters of the round parts of the punctuation which
// stay circles: the dots and the heads of the comma and the quotes, and the
// ball of the question mark.
func punctDots(rp design.Params) []float64 {
	return []float64{rp.Punct.CommaHead, rp.BallDiameter}
}

// The signs.  Sloped from the roman, their rounds turned and their dots
// circles: the punctuation, the quotes, the brackets of all three kinds, the
// dashes, the number, percent, currency, section and at signs and the
// asterisk.  Upright as in the roman; they stand on the maths axis or
// lean by their own form in every style: the slash and the backslash, the
// bar, the underscore, the ASCII accents, the degree sign and the
// mathematical signs.
func init() {
	sloped := []struct {
		name string
		r    rune
	}{
		{"period", '.'}, {"comma", ','}, {"colon", ':'}, {"semicolon", ';'},
		{"exclam", '!'}, {"question", '?'},
		{"quotesingle", '\''}, {"quotedbl", '"'},
		{"quoteleft", '‘'}, {"quoteright", '’'}, {"quotedblleft", '“'}, {"quotedblright", '”'},
		{"hyphen", '-'}, {"endash", '–'}, {"emdash", '—'},
		{"parenleft", '('}, {"parenright", ')'}, {"bracketleft", '['}, {"bracketright", ']'},
		{"braceleft", '{'}, {"braceright", '}'},
		{"numbersign", '#'}, {"dollar", '$'}, {"sterling", '£'}, {"Euro", '€'},
		{"section", '§'},
	}
	for _, g := range sloped {
		sources[g.name] = source{r: g.r, dots: punctDots}
	}
	// The ball of the sterling sign hangs on a long arc, which needs the
	// widest zone (see circle) for the smoothness check to find it fair.
	st := sources["sterling"]
	st.zone = 3
	sources["sterling"] = st
	sources["percent"] = source{r: '%', rounds: allCounters("percent")}
	sources["at"] = source{r: '@', rounds: allCounters("at")}
	upright := []struct {
		name string
		r    rune
	}{
		{"slash", '/'}, {"backslash", '\\'}, {"bar", '|'}, {"underscore", '_'},
		{"asterisk", '*'}, {"asciitilde", '~'}, {"asciicircum", '^'}, {"grave", '`'}, {"degree", '°'},
		{"plus", '+'}, {"minus", '−'}, {"equal", '='}, {"less", '<'}, {"greater", '>'},
	}
	for _, g := range upright {
		sources[g.name] = source{r: g.r, upright: true}
	}
}

// signDots returns the diameters of the round parts of the further
// punctuation which stay circles: the dot of the full stop, which the
// inverted marks, the ellipsis and the middle dot use, the heads of the low
// quotes and the ball of the inverted question mark.
func signDots(rp design.Params) []float64 {
	return append([]float64{rp.DotDiameter}, punctDots(rp)...)
}

// bulletDots returns the diameter of the bullet, which is one circle.
func bulletDots(rp design.Params) []float64 {
	path, _, _ := glyphsOutline("bullet", rp)
	x0, _, x1, _ := path.BBox()
	return []float64{x1 - x0}
}

// supBall returns the diameter of the balls of the superior figures, which
// are the balls of the figures drawn as thick as Figures.SupWeight and then
// scaled.
func supBall(rp design.Params) float64 { return rp.Figures.Ball * rp.Figures.SupWeight }

// letterCounters returns the rounds of the counters of a letter in a ring (®):
// the counters but the largest, which is the ring.  The ring is a circle, whose
// turn is the identity, and the nodes of the letter lie inside it; turning it turns the letter.  The counter of R is a bowl on its stem
// and turns on the side away from the stem.
func letterCounters(name string) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline(name, rp)
		all := counters(path, -70, 70, 25)
		ring := 0
		for i, r := range all {
			if r.A > all[ring].A {
				ring = i
			}
		}
		var out []round
		for i, r := range all {
			if i != ring {
				out = append(out, r)
			}
		}
		return out
	}
}

// centRound returns the rounds function of the cent sign: the inner edge of
// its c, in the position of the c in the sign, turned between the directions
// from and to which the aperture leaves free.
func centRound(from, to float64) func(Params, design.Params) []round {
	return func(_ Params, rp design.Params) []round {
		c, _, _ := romanOutline("c", rp)
		cent, _, _ := romanOutline("cent", rp)
		cx0, _, cx1, _ := c.BBox()
		sx0, _, _, _ := cent.BBox()
		cx := (cx0 + cx1) / 2
		xs := sorted(c.CrossingsX(cy(c)))
		ys := sorted(c.CrossingsY(cx))
		if len(xs) < 2 || len(ys) < 4 {
			return nil
		}
		return []round{{C: pt(cx+sx0-cx0, (ys[1]+ys[2])/2), A: cx - xs[1], B: (ys[2] - ys[1]) / 2, From: from, To: to, Taper: 20}}
	}
}

// cy returns the height of the middle of the box of the path.
func cy(path outline.Path) float64 {
	_, y0, _, y1 := path.BBox()
	return (y0 + y1) / 2
}

// The further signs, made as the punctuation is: sloped from the roman with
// their rounds turned, their dots, balls and heads circles.
//
// Circles only (dots and heads): the low quotes, the ellipsis, the bullet and
// the middle dot, and the inverted exclamation and question marks, whose
// hook is not turned, as that of the question mark is not.  Nothing to turn:
// the guillemets, the primes, the fraction bar, the pilcrow (its bowl is
// solid) and the yen, which have straight strokes only, and the daggers, whose
// limbs are solid and whose round ends become ellipses with the shear.  The superior
// figures and the fractions take the balls of the figures 2 and 3 scaled with
// them; their figures 1, 2, 3 and 4 have no closed round counter (the counter
// of 4 is a triangle).  The per mille sign turns the counters of its three
// zeros as the percent sign does.  The rings of the currency sign and of the
// circled signs are circles, whose turn is the identity, and the C in the
// copyright sign is open: they have no rounds; the bowl of the R in the
// registered sign and the c of the cent sign turn.
func init() {
	for _, g := range []struct {
		name string
		r    rune
	}{
		{"guillemotleft", '«'}, {"guillemotright", '»'}, {"guilsinglleft", '‹'}, {"guilsinglright", '›'},
		{"dagger", '†'}, {"daggerdbl", '‡'}, {"paragraph", '¶'}, {"minute", '′'}, {"second", '″'},
		{"fraction", '⁄'}, {"yen", '¥'},
	} {
		sources[g.name] = source{r: g.r}
	}
	for _, g := range []struct {
		name string
		r    rune
	}{
		{"quotesinglbase", '‚'}, {"quotedblbase", '„'}, {"ellipsis", '…'}, {"periodcentered", '·'},
		{"exclamdown", '¡'}, {"questiondown", '¿'},
	} {
		sources[g.name] = source{r: g.r, dots: signDots}
	}
	sources["bullet"] = source{r: '•', dots: bulletDots}
	for _, g := range []struct {
		name string
		r    rune
	}{
		{"uni00B9", '¹'}, {"uni00B2", '²'}, {"uni00B3", '³'},
		{"onequarter", '¼'}, {"onehalf", '½'}, {"threequarters", '¾'},
	} {
		sources[g.name] = source{r: g.r, ball: supBall}
	}
	// The zone of the ball of the 3 (see circle), wide as that of the balls of
	// the long hooks, which the smoothness check needs for the superior 3.
	for _, n := range []string{"uni00B3", "threequarters"} {
		src := sources[n]
		src.zone = 3.5
		sources[n] = src
	}
	sources["perthousand"] = source{r: '‰', rounds: allCounters("perthousand")}
	sources["copyright"] = source{r: '©'}
	sources["registered"] = source{r: '®', rounds: letterCounters("registered")}
	sources["currency"] = source{r: '¤'}
	sources["cent"] = source{r: '¢', rounds: centRound(50, 310)}
}
