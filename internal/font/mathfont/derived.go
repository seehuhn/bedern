// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/sfnt/opentype/mathtable"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs which the maths font derives from the signs of the regular.  Each is
// made from the unrounded outlines of the regular by moving, mirroring,
// turning or repeating them, or by running a construction of the regular
// with other lengths and the same stroke weights; no outline is scaled.
//
// Side spaces follow the class of the sign.  A binary operator or a relation
// whose ink is no wider than the plus is centred in the advance of the signs
// of the regular (design.Params.MathAdvance); a wider one has MathSide on
// each side.  The dots, the bars, the brackets, the primes and the
// existential quantifier keep the side spaces of their parent.
const (
	// leqHeight is the ink height of ≤ over that of <.
	leqHeight = 1.21
	// llPitch is the distance of the two < of ≪ over the ink width of <; ≪ is
	// 1.56 times as wide as <.
	llPitch = 0.56
	// equivGap is the white between the bars of ≡ over that of =.
	equivGap = 0.82
	// vertGap is the white between the bars of ‖.
	vertGap = 100
	// astSize is the diameter of ∗ over that of the asterisk.
	astSize = 0.90
	// setminusHeight is the ink height of ∖, centred on the axis.
	setminusHeight = 740
	// existsWidth is the ink width of ∃ over the cap height.
	existsWidth = 0.62
	// existsMiddle is the length of the middle arm of ∃ over its ink width.
	existsMiddle = 0.96
	// existsStem is the stem of ∃ over the rule.
	existsStem = 1.25
	// perpHeight is the height of ⊥ and ⊤ over their width.
	perpHeight = 1.07
	// transposeWidth is the width of ⊺ over that of ⊤.
	transposeWidth = 0.65
	// hookLength is the reach of the arms of ⌈ ⌉ ⌊ ⌋ over that of the
	// bracket, which reaches 112.
	hookLength = 1.4
)

// A Class is the class of a sign in mathematics, which decides its side
// spaces.
type Class int

// The classes.
const (
	Ord Class = iota
	Bin
	Rel
	Punct
	Fence
	Accent
)

// A Glyph is a glyph of the maths font.
type Glyph struct {
	Name    string
	Rune    rune // the code point, or 0
	Advance float64
	Path    outline.Path // unrounded
	Class   Class
	// Like names the glyph of the regular whose bounding-box rule applies,
	// to the mirror image of the glyph if Mirror is set.
	Like   string
	Mirror bool
	// Variants names the larger sizes of a delimiter in increasing order,
	// and Parts the parts of its assembly from the bottom up; with
	// Horizontal set they are the wider sizes and the parts from the left of
	// a horizontal construction.
	Variants   []string
	Parts      []Part
	Horizontal bool
	// Italic is the italics correction; Attach is the top accent attachment
	// where Attached is set, and the middle of the ink otherwise; Kern holds
	// the math kerns, and Script the names of the forms of the first and the
	// second script level.
	Italic   float64
	Attach   float64
	Attached bool
	Kern     *mathtable.KernInfo
	Script   [2]string
}

// delimiterNames are the glyphs which unicode-math uses as delimiters.  A
// document takes a delimiter from Bedern Math only once it has its sizes
// (Ranges).
var (
	delimiterNames = map[string]bool{
		"parenleft": true, "parenright": true, "bracketleft": true, "bracketright": true,
		"braceleft": true, "braceright": true, "bar": true, "uni2016": true,
		"uni2308": true, "uni2309": true, "uni230A": true, "uni230B": true,
		"uni27E8": true, "uni27E9": true,
	}
)

// roman holds the unrounded outlines of the regular for one set of
// parameters, each drawn once.
type roman struct {
	p  design.Params
	mu sync.Mutex
	m  map[string]*romanGlyph
	// nominal is the height of the delimiter whose proportions the brace
	// takes, where it differs from the height it is drawn at (braceProportions
	// in braces.go); zero means the height it is drawn at.
	nominal float64
}

type romanGlyph struct {
	once sync.Once
	path outline.Path
	adv  float64
	ok   bool
}

func newRoman(p design.Params) *roman { return &roman{p: p, m: map[string]*romanGlyph{}} }

// newRomanNominal returns the glyphs of the regular for p, with the braces
// drawn with the proportions of the delimiter nominal high.
func newRomanNominal(p design.Params, nominal float64) *roman {
	r := newRoman(p)
	r.nominal = nominal
	return r
}

// get returns the unrounded outline and the advance of the glyph of the
// regular.  The path is shared and must not be changed.
func (r *roman) get(name string) (outline.Path, float64) {
	r.mu.Lock()
	g := r.m[name]
	if g == nil {
		g = new(romanGlyph)
		r.m[name] = g
	}
	r.mu.Unlock()
	g.once.Do(func() {
		switch name {
		case "braceleft", "braceright":
			g.path, g.adv, g.ok = mathBrace(r.p, r.nominal, name == "braceright")
		default:
			g.path, g.adv, g.ok = glyphs.Outline(name, r.p)
		}
	})
	if !g.ok {
		panic(fmt.Sprintf("mathfont: the regular has no glyph %s", name))
	}
	return g.path, g.adv
}

// A derivation names a derived sign and how it is made.
type derivation struct {
	name  string
	r     rune
	class Class
	like  string
	draw  func(r *roman) (outline.Path, float64)
}

// mirrorLike are the derived signs to whose mirror image the bounding-box rule
// of their parent applies.
var mirrorLike = map[string]bool{"uni2203": true}

// derivations are the signs derived from the roman, in the order of the font.
var derivations = []derivation{
	{"uni2236", 0x2236, Rel, "colon", ratio},
	{"uni2217", 0x2217, Bin, "asterisk", centredAsterisk},
	{"uni2219", 0x2219, Bin, "bullet", copyOf("bullet")},
	{"uni22C5", 0x22C5, Bin, "periodcentered", copyOf("periodcentered")},
	{"uni2213", 0x2213, Bin, "plusminus", minusplus},
	{"uni2034", 0x2034, Ord, "second", primes(3)},
	{"uni2057", 0x2057, Ord, "second", primes(4)},
	{"uni22EF", 0x22EF, Ord, "ellipsis", cdots},
	{"uni22EE", 0x22EE, Ord, "period", vdots},
	{"uni22F1", 0x22F1, Ord, "ellipsis", ddots(false)},
	{"uni22F0", 0x22F0, Ord, "ellipsis", ddots(true)},
	{"uni2016", 0x2016, Fence, "bar", bars(vertGap)},
	{"uni2223", 0x2223, Rel, "bar", copyOf("bar")},
	{"uni2225", 0x2225, Rel, "bar", parallel},
	{"uni2264", 0x2264, Rel, "less", leq(false)},
	{"uni2265", 0x2265, Rel, "greater", leq(true)},
	{"uni226A", 0x226A, Rel, "less", ll(false)},
	{"uni226B", 0x226B, Rel, "greater", ll(true)},
	{"uni2260", 0x2260, Rel, "equal", neq},
	{"uni0338", 0x0338, Accent, "acutecomb", notOverlay},
	{"uni2261", 0x2261, Rel, "equal", equiv},
	{"uni2308", 0x2308, Fence, "bracketleft", ceilFloor(false, false)},
	{"uni2309", 0x2309, Fence, "bracketright", ceilFloor(true, false)},
	{"uni230A", 0x230A, Fence, "bracketleft", ceilFloor(false, true)},
	{"uni230B", 0x230B, Fence, "bracketright", ceilFloor(true, true)},
	{"uni2203", 0x2203, Ord, "E", exists},
	{"uni2216", 0x2216, Bin, "backslash", setminus},
	{"uni22A5", 0x22A5, Ord, "plus", perp(false, 1)},
	{"uni22A4", 0x22A4, Ord, "plus", perp(true, 1)},
	{"uni22BA", 0x22BA, Bin, "plus", perp(true, transposeWidth)},
	{"uni27C2", 0x27C2, Rel, "plus", axisSign(perp(false, 1))},
	{"uni25A1", 0x25A1, Ord, "plus", square},
	{"uni27E8", 0x27E8, Fence, "parenleft", angleBracket(false)},
	{"uni27E9", 0x27E9, Fence, "parenright", angleBracket(true)},
}

// copyOf returns the glyph of the regular unchanged.
func copyOf(name string) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) { return r.get(name) }
}

// dotAt returns the full stop moved to put its centre at (x, y).
func dotAt(r *roman, x, y float64) outline.Path {
	path, _ := r.get("period")
	x0, y0, x1, y1 := bbox(path)
	return shifted(path, math.Round(x-(x0+x1)/2), math.Round(y-(y0+y1)/2))
}

// periodSides returns the side spaces of the full stop.
func periodSides(r *roman) (left, right float64) {
	path, adv := r.get("period")
	x0, _, x1, _ := bbox(path)
	return x0, adv - x1
}

// ratio draws ∶: the dots of the colon, centred on the axis.
func ratio(r *roman) (outline.Path, float64) {
	path, adv := r.get("colon")
	_, y0, _, y1 := bbox(path)
	return shifted(path, 0, math.Round(r.p.Punct.MathAxis-(y0+y1)/2)), adv
}

// centredAsterisk draws ∗: the construction of the asterisk with astSize of
// its diameter and arms of the same weight, its ink centred on the axis, with
// the side spaces of the asterisk.
func centredAsterisk(r *roman) (outline.Path, float64) {
	q := r.p
	q.Punct.AsteriskSize = math.Round(astSize * r.p.Punct.AsteriskSize)
	path, _, _ := glyphs.Outline("asterisk", q)
	_, y0, _, y1 := bbox(path)
	path = shifted(path, 0, math.Round(r.p.Punct.MathAxis-(y0+y1)/2))
	return withSides(path, r.p.Punct.Asterisk.Left, r.p.Punct.Asterisk.Right)
}

// minusplus draws ∓: ± turned about the middle of its ink.
func minusplus(r *roman) (outline.Path, float64) {
	path, adv := r.get("plusminus")
	x0, y0, x1, y1 := bbox(path)
	return turned(path, (x0+x1)/2, (y0+y1)/2), adv
}

// primes returns the construction of n primes: the prime repeated with the
// pitch of the double prime, with the side spaces of the prime.
func primes(n int) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		one, adv := r.get("minute")
		x0, _, x1, _ := bbox(one)
		var path outline.Path
		for i := range n {
			path = join(path, shifted(one, float64(i)*r.p.Punct.StraightPitch, 0))
		}
		return withSides(path, x0, adv-x1)
	}
}

// cdots draws ⋯: the ellipsis with its dots on the axis.
func cdots(r *roman) (outline.Path, float64) {
	path, adv := r.get("ellipsis")
	return shifted(path, 0, math.Round(r.p.Punct.MathAxis-r.p.PeriodCentre())), adv
}

// vdots draws ⋮: three dots of the full stop above each other with the pitch
// of the ellipsis, centred on the axis, with the side spaces of the full stop.
func vdots(r *roman) (outline.Path, float64) {
	pitch := r.p.DotsAdvance()
	var path outline.Path
	for i := -1; i <= 1; i++ {
		path = join(path, dotAt(r, 0, r.p.Punct.MathAxis+float64(i)*pitch))
	}
	l, rt := periodSides(r)
	return withSides(path, l, rt)
}

// ddots draws ⋱, or ⋰ if up is true: three dots of the full stop with the
// pitch of the ellipsis across and up, centred on the axis, with the side
// spaces of the full stop.
func ddots(up bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		pitch := r.p.DotsAdvance()
		var path outline.Path
		for i := -1; i <= 1; i++ {
			dy := -float64(i) * pitch
			if up {
				dy = -dy
			}
			path = join(path, dotAt(r, float64(i)*pitch, r.p.Punct.MathAxis+dy))
		}
		l, rt := periodSides(r)
		return withSides(path, l, rt)
	}
}

// bars returns two vertical bars with the white gap between them, with the
// side spaces of the bar.
func bars(gap float64) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		one, adv := r.get("bar")
		x0, _, x1, _ := bbox(one)
		path := join(one, shifted(one, x1-x0+gap, 0))
		return withSides(path, x0, adv-x1)
	}
}

// parallel draws ∥: two bars with the white of the equals between them.
func parallel(r *roman) (outline.Path, float64) {
	return bars(r.p.Punct.EqualsGap)(r)
}

// lessWithHeight returns the construction of < with the angle of its strokes
// chosen to make its ink h high, at the width of the regular.
func lessWithHeight(p design.Params, h float64) outline.Path {
	lo, hi := 1.0, p.Punct.MathAngle
	var path outline.Path
	for range 60 {
		q := p
		q.Punct.MathAngle = (lo + hi) / 2
		path, _, _ = glyphs.Outline("less", q)
		_, y0, _, y1 := bbox(path)
		if y1-y0 > h {
			hi = q.Punct.MathAngle
		} else {
			lo = q.Punct.MathAngle
		}
	}
	return path
}

// pmGap returns the white of ± between its bar and its plus.
func pmGap(r *roman) float64 {
	path, _ := r.get("plusminus")
	lowest := math.Inf(1)
	for _, c := range path[1:] {
		_, y0, _, _ := bbox(outline.Path{c})
		lowest = math.Min(lowest, y0)
	}
	_, _, _, barTop := bbox(path[:1])
	return lowest - barTop
}

// leq draws ≤, or ≥ if mirror is true: a bar of the rule under the
// construction of <, as wide as it and with the white of ± between them, the
// whole leqHeight times as high as < and centred on the axis.  The < keeps
// its width and stroke and is made flatter.
func leq(mirror bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		p := r.p
		less, _ := r.get("less")
		_, ly0, _, ly1 := bbox(less)
		total := math.Round(leqHeight * (ly1 - ly0))
		t, gap := p.Punct.MathThick, pmGap(r)
		flat := lessWithHeight(p, total-t-gap)
		x0, y0, x1, _ := bbox(flat)
		bottom := p.Punct.MathAxis - total/2
		flat = shifted(flat, 0, bottom+t+gap-y0)
		path := join(outline.Path{box(x0, bottom, x1, bottom+t)}, flat)
		if mirror {
			path = mirroredX(path, (x0+x1)/2)
		}
		return centredIn(path, p.MathAdvance()), p.MathAdvance()
	}
}

// ll draws ≪, or ≫ if mirror is true: two of <, llPitch of its width apart,
// with MathSide on each side.
func ll(mirror bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		less, _ := r.get("less")
		x0, _, x1, _ := bbox(less)
		path := join(less, shifted(less, math.Round(llPitch*(x1-x0)), 0))
		if mirror {
			path = mirroredX(path, 0)
		}
		return withSides(path, r.p.Punct.MathSide, r.p.Punct.MathSide)
	}
}

// neq draws ≠: the equals crossed by the solidus of the negations at its
// middle (negationOver).
func neq(r *roman) (outline.Path, float64) {
	eq, adv := r.get("equal")
	return glyphs.Union(eq, negationOver(r, eq)), adv
}

// notOverlay draws U+0338: the solidus of the negations of a sign as high as
// ⊂ and ∈, alone, as an accent whose advance is its ink.
func notOverlay(r *roman) (outline.Path, float64) {
	return withSides(solidus(r, 0, r.p.Punct.MathAxis, setHeight+2*negOver).Round(), 0, 0)
}

// equiv draws ≡: three bars of the minus, equivGap of the white of the
// equals apart, centred on the axis.
func equiv(r *roman) (outline.Path, float64) {
	p := r.p
	h, t := p.Punct.MathWidth/2, p.Punct.MathThick
	pitch := math.Round(equivGap*p.Punct.EqualsGap) + t
	var path outline.Path
	for i := -1; i <= 1; i++ {
		y := p.Punct.MathAxis + float64(i)*pitch
		path = append(path, box(-h, y-t/2, h, y+t/2))
	}
	return centredIn(path, p.MathAdvance()), p.MathAdvance()
}

// ceilFloor draws ⌈, or ⌉ if right is true, and ⌊ and ⌋ if floor is true:
// the bracket of the regular without its lower arm (its upper arm for a
// floor), with the side spaces of the bracket.
func ceilFloor(right, floor bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		p := r.p
		q := p.Punct
		top, bottom := p.BracketTop(), p.BracketBottom()
		x := q.BracketStem + math.Round(hookLength*q.BracketReach)
		path := outline.Path{polygon(
			pt(0, bottom), pt(q.BracketStem, bottom), pt(q.BracketStem, top-q.BracketArm),
			pt(x, top-q.BracketArm), pt(x, top), pt(0, top))}
		if floor {
			path = mirroredY(path, (top+bottom)/2)
		}
		sides := q.BracketLeft
		if right {
			path = mirroredX(path, 0)
			sides = design.PunctSides{Left: sides.Right, Right: sides.Left}
		}
		return withSides(path, sides.Left, sides.Right)
	}
}

// exists draws ∃: three arms of the rule to the left of a stem existsStem
// times the rule, existsWidth of the cap height wide and as high, the middle
// arm at half the height and existsMiddle of the width long, with the side
// spaces of E mirrored.
func exists(r *roman) (outline.Path, float64) {
	p := r.p
	t := p.Punct.MathThick
	s := math.Round(existsStem * t)
	w := math.Round(existsWidth * p.CapHeight)
	h := p.CapHeight
	m := h / 2
	l := w - math.Round(existsMiddle*w)
	path := outline.Path{polygon(
		pt(0, 0), pt(w, 0), pt(w, h), pt(0, h), pt(0, h-t), pt(w-s, h-t),
		pt(w-s, m+t/2), pt(l, m+t/2), pt(l, m-t/2), pt(w-s, m-t/2), pt(w-s, t), pt(0, t))}
	e, adv := r.get("E")
	x0, _, x1, _ := bbox(e)
	return withSides(path, adv-x1, x0)
}

// setminus draws ∖: the construction of the backslash, setminusHeight high
// and centred on the axis, with the side spaces of the backslash.
func setminus(r *roman) (outline.Path, float64) {
	p := r.p
	a := p.Punct.SlashAngle * math.Pi / 180
	w := p.Punct.SlashThick / math.Cos(a)
	run := setminusHeight * math.Tan(a)
	y0, y1 := p.Punct.MathAxis-setminusHeight/2, p.Punct.MathAxis+setminusHeight/2
	slash := outline.Path{polygon(pt(0, y0), pt(w, y0), pt(run+w, y1), pt(run, y1))}
	return withSides(mirroredX(slash, 0), p.Punct.Backslash.Left, p.Punct.Backslash.Right)
}

// perp draws ⊥, or ⊤ if down is true: the bars of the plus, the horizontal
// one widthShare of the width of the plus long, the vertical one perpHeight
// times the width of the plus high, standing on the baseline, centred in the
// advance of the signs.
func perp(down bool, widthShare float64) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		p := r.p
		t := p.Punct.MathThick
		hw := math.Round(widthShare*p.Punct.MathWidth) / 2
		h := math.Round(perpHeight * p.Punct.MathWidth)
		path := outline.Path{polygon(
			pt(-hw, 0), pt(hw, 0), pt(hw, t), pt(t/2, t), pt(t/2, h), pt(-t/2, h), pt(-t/2, t), pt(-hw, t))}
		if down {
			path = mirroredY(path, h/2)
		}
		return centredIn(path, p.MathAdvance()), p.MathAdvance()
	}
}

// axisSign moves a sign up to centre its ink on the axis (⟂ is a relation, ⊥
// stands on the baseline).
func axisSign(draw func(r *roman) (outline.Path, float64)) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		path, adv := draw(r)
		return onAxisPath(r, path), adv
	}
}

// square draws □: four bars of the rule round a square as wide as the plus,
// standing on the baseline, centred in the advance of the signs.
func square(r *roman) (outline.Path, float64) {
	p := r.p
	w, t := p.Punct.MathWidth, p.Punct.MathThick
	inner := box(t, t, w-t, w-t)
	path := outline.Path{box(0, 0, w, w), reversed(inner)}
	return centredIn(path, p.MathAdvance()), p.MathAdvance()
}

// Derived returns the derived signs of the parameters, in the order of the
// font.
func Derived(p design.Params) []Glyph {
	return derived(newRoman(p))
}

func derived(r *roman) []Glyph {
	out := make([]Glyph, len(derivations))
	var wg sync.WaitGroup
	for i, d := range derivations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, adv := d.draw(r)
			out[i] = Glyph{Name: d.name, Rune: d.r, Advance: math.Round(adv), Path: path, Class: d.class, Like: d.like, Mirror: mirrorLike[d.name]}
		}()
	}
	wg.Wait()
	return out
}
