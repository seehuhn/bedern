// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import (
	"io"
	"math"
	"sort"
)

// Report says what [Font.Build] did beyond copying the description.
type Report struct {
	// Reductions lists the tables of dimensions with more values than the
	// format holds, in the order width, height, depth, italic correction.
	Reductions []Reduction

	// SevenBitSafe is the seven-bit-safe flag of the header: no character
	// code below 128 leads to a code of 128 or more through a ligature, a
	// list of larger characters or a recipe.
	SevenBitSafe bool
}

// Reduction describes the reduction of one table of dimensions.
type Reduction struct {
	// Table is "width", "height", "depth" or "italic correction".
	Table string

	// From and To are the numbers of entries of the table before and after
	// the reduction, with entry 0.
	From, To int

	// Tolerance is the distance within which neighbouring values are
	// merged, in the units of the description.  It is negative where the
	// search of PLtoTF overflows.
	Tolerance Fix

	// MaxChange is the largest change of a value, in the units of the
	// description.
	MaxChange Fix
}

// noLabel is the address of a program which does not exist.
const noLabel = 0x7FFF

// program holds a ligature and kerning program as PLtoTF holds it before it
// lays the words out.
type program struct {
	lig        []LigKern
	kern       []Fix
	tag        [256]Tag
	rem        [256]int
	bcharLabel int
}

// assemble turns the tags of the characters and the instructions of
// f.LigTable into words.
func (f *Font) assemble() (*program, error) {
	p := &program{bcharLabel: noLabel}
	for c, ch := range f.Chars {
		if ch == nil {
			continue
		}
		switch ch.Tag {
		case NoTag:
		case ListTag:
			p.tag[c] = ListTag
			p.rem[c] = int(ch.NextLarger)
		case ExtTag:
			p.tag[c] = ExtTag
		default:
			return nil, invalid("character %d has a tag which is not NoTag, ListTag or ExtTag", c)
		}
	}
	kernIndex := make(map[Fix]int)
	minNL := 0
	ended := false
	for i, in := range f.LigTable {
		nl := len(p.lig)
		switch in.Kind {
		case KindLabel:
			c := in.Char
			switch p.tag[c] {
			case LigTag:
				return nil, invalid("instruction %d: character %d has a label already", i, c)
			case ListTag:
				return nil, invalid("instruction %d: character %d has a next larger character", i, c)
			case ExtTag:
				return nil, invalid("instruction %d: character %d has a recipe", i, c)
			}
			p.tag[c] = LigTag
			p.rem[c] = nl
			minNL = max(minNL, nl+1)
			ended = false
		case KindBoundaryLabel:
			p.bcharLabel = nl
			minNL = max(minNL, nl+1)
			ended = false
		case KindStop:
			if !ended {
				return nil, invalid("instruction %d: a stop does not follow a step", i)
			}
			p.lig[nl-1].Skip = 128
			ended = false
		case KindSkip:
			if !ended {
				return nil, invalid("instruction %d: a skip does not follow a step", i)
			}
			if in.Skip >= 128 {
				return nil, invalid("instruction %d: a skip of more than 127 steps", i)
			}
			p.lig[nl-1].Skip = in.Skip
			minNL = max(minNL, nl+int(in.Skip)+1)
			ended = false
		case KindStep:
			s := in.Step
			switch {
			case s.Op == OpKern:
				k, ok := kernIndex[s.Kern]
				if !ok {
					k = len(p.kern)
					if k > 32767 {
						return nil, invalid("more than 32767 different kerns")
					}
					kernIndex[s.Kern] = k
					p.kern = append(p.kern, s.Kern)
				}
				p.lig = append(p.lig, LigKern{0, s.Next, uint8(128 + k>>8), uint8(k)})
			case s.Op.IsLig():
				p.lig = append(p.lig, LigKern{0, s.Next, uint8(s.Op), s.Result})
			default:
				return nil, invalid("instruction %d: operation %d is no ligature operation and no kern", i, s.Op)
			}
			if len(p.lig) > 32767 {
				return nil, invalid("more than 32767 steps")
			}
			ended = true
		default:
			return nil, invalid("instruction %d has no valid kind", i)
		}
	}
	if minNL > len(p.lig) {
		return nil, invalid("the last label or skip of the program has no step to go to")
	}
	if len(p.lig) > 0 {
		if p.bcharLabel < noLabel {
			p.lig = append(p.lig, LigKern{Skip: 255})
		}
		if last := &p.lig[len(p.lig)-1]; last.Skip == 0 {
			last.Skip = 128
		}
	}
	return p, nil
}

// infinity is larger than every dimension.
const infinity = math.MaxInt32

// minCover returns the smallest number of intervals of length d which cover
// the sorted values, and the smallest larger d which gives another cover.
// The arithmetic is that of PLtoTF: 32 bits, with wrap-around.
func minCover(v []Fix, d int32) (m int, nextD int32) {
	nextD = infinity
	p := 0
	for p < len(v) {
		m++
		l := int32(v[p])
		for p+1 < len(v) && int32(v[p+1]) <= l+d {
			p++
		}
		p++
		next := int32(infinity)
		if p < len(v) {
			next = int32(v[p])
		}
		nextD = min(nextD, next-l)
	}
	return m, nextD
}

// shorten returns the d which PLtoTF finds for a cover of the sorted values
// by m intervals of length d, or zero if there are at most m values.
func shorten(v []Fix, m int) int32 {
	if len(v) <= m {
		return 0
	}
	_, d := minCover(v, 0)
	half := d
	for {
		half = d
		d += d
		if k, _ := minCover(v, d); k <= m {
			break
		}
	}
	d = half
	k, nextD := minCover(v, d)
	for k > m {
		d = nextD
		k, nextD = minCover(v, d)
	}
	return d
}

// dimTable is one table of dimensions in the making.
type dimTable struct {
	vals  []Fix // the different values, sorted
	index []int // the entry of the table which stands for each value
	held  []Fix // the value which PLtoTF holds in place of each value
	out   []Fix // the entries of the table, without entry 0
}

func (t *dimTable) add(x Fix) {
	t.vals = append(t.vals, x)
}

// find returns the position of x in t.vals.
func (t *dimTable) find(x Fix) int {
	return sort.Search(len(t.vals), func(i int) bool { return t.vals[i] >= x })
}

// reduce sorts the values, merges neighbouring values until at most m are
// left, and returns the tolerance of the merge and the largest change.
func (t *dimTable) reduce(m int) (delta, change Fix) {
	sort.Slice(t.vals, func(i, j int) bool { return t.vals[i] < t.vals[j] })
	n := 0
	for i, x := range t.vals {
		if i == 0 || x != t.vals[n-1] {
			t.vals[n] = x
			n++
		}
	}
	t.vals = t.vals[:n]
	d := shorten(t.vals, m)
	delta = Fix(d)
	excess := n - m
	t.index = make([]int, n)
	t.held = append([]Fix(nil), t.vals...)
	p := 0
	for p < n {
		first := p
		l := int32(t.vals[p])
		for p+1 < n && int32(t.vals[p+1]) <= l+d {
			p++
			excess--
			if excess == 0 {
				d = 0
			}
		}
		mid := Fix(l + (int32(t.vals[p])-l)/2)
		t.out = append(t.out, mid)
		t.held[p] = mid
		for i := first; i <= p; i++ {
			t.index[i] = len(t.out)
			c := int64(t.vals[i]) - int64(mid)
			if c < 0 {
				c = -c
			}
			if c > int64(change) {
				change = Fix(min(c, math.MaxInt32))
			}
		}
		p++
	}
	return delta, change
}

// roundHalfAway rounds r to the nearest integer, halves away from zero.
func roundHalfAway(r float64) int64 {
	switch {
	case r > 2147483647.0:
		return 2147483647
	case r < -2147483647.0:
		return -2147483647
	case r >= 0:
		return int64(r + 0.5)
	}
	return int64(r - 0.5)
}

// builder holds the state of Font.Build.
type builder struct {
	f      *Font
	units  Fix
	prog   *program
	lc     loopChecker
	unsafe bool
}

func (b *builder) exists(c uint8) bool {
	return b.f.Chars[c] != nil
}

// scaled returns the word of the dimension x.
func (b *builder) scaled(x Fix, what string) (Fix, error) {
	if math.Abs(float64(x)/float64(b.units)) >= 16.0 {
		return 0, invalid("the %s %s is not below 16 design sizes", what, x)
	}
	v := int64(x)
	if b.units != Unity {
		v = roundHalfAway(float64(x) / float64(b.units) * 1048576.0)
	}
	if v < 0 {
		v += 1 << 24
		if v <= 0 {
			v = 1
		}
		return Fix(int32(0xFF000000 | uint32(v))), nil
	}
	if v >= 1<<24 {
		v = 1<<24 - 1
	}
	return Fix(v), nil
}

// checkProgram enters the program of c, a character code or Boundary, into
// the loop checker, and checks the characters which the steps name.
func (b *builder) checkProgram(c, start int) error {
	p := b.prog
	bchar := -1
	if b.f.HasBoundary {
		bchar = int(b.f.BoundaryChar)
	}
	for i := start; i < len(p.lig); {
		s := p.lig[i]
		if b.lc.add(c, s) {
			if int(s.Next) != bchar && !b.exists(s.Next) {
				if s.Skip > 128 {
					return invalid("the program of character %d runs into the word of the left boundary, for which PLtoTF asks for character 0", c)
				}
				return invalid("a step of the program of character %d examines character %d, which the font does not have", c, s.Next)
			}
			if s.Op < 128 {
				if !b.exists(s.Remainder) {
					if s.Skip > 128 {
						return invalid("the program of character %d runs into the word of the left boundary, for which PLtoTF asks for character 0", c)
					}
					return invalid("a step of the program of character %d inserts character %d, which the font does not have", c, s.Remainder)
				}
				if s.Remainder >= 128 && (c < 128 || c == Boundary) && (s.Next < 128 || int(s.Next) == bchar) {
					b.unsafe = true
				}
			}
		}
		if s.Skip >= 128 {
			break
		}
		i += int(s.Skip) + 1
	}
	return nil
}

// safety notes that character c leads to character g.
func (b *builder) safety(c int, g uint8, what string) error {
	if g >= 128 && c < 128 {
		b.unsafe = true
	}
	if !b.exists(g) {
		return invalid("the %s of character %d is character %d, which the font does not have", what, c, g)
	}
	return nil
}

func validString(s string, max int) bool {
	if len(s) >= max {
		return false
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; c < ' ' || c > '~' || c == '(' || c == ')' {
			return false
		}
	}
	return true
}

// Build turns the description into the tables of a TFM file by the method of
// PLtoTF.  The file is, byte for byte, the one which PLtoTF writes for a PL
// file which says the same.
//
// The tables of dimensions are sorted.  A table with more values than the
// format holds (255 widths, 15 heights, 15 depths and 63 italic corrections
// beside entry 0) is reduced: neighbouring values are merged at the smallest
// tolerance which leaves few enough values, and every set of merged values is
// replaced by the mean of its smallest and its largest value.  The report
// lists the reductions.  The search for the tolerance uses the arithmetic of
// PLtoTF, 32 bits with wrap-around; it overflows for a table whose largest
// value is negative, and Build then merges the values which PLtoTF merges.
//
// Build returns an [*InvalidFontError] where PLtoTF reports an error or
// changes the data: for a character which a step, a list of larger characters
// or a recipe names and which the font does not have, a cycle of larger
// characters, an infinite loop of ligatures, a character with two tags, a
// stop or skip which does not follow a step, a dimension of 16 design sizes or
// more, a design size below 1 and a string which the header cannot hold.  It
// also returns one for a label or skip at the end of the program which has no
// step to go to and for a font which is too large for the format.
//
// Build is not as strict as PLtoTF about the number of different kerns: PLtoTF
// stops at 5000, and Build accepts as many as the format holds, 32767.
func (f *Font) Build() (*File, *Report, error) {
	b := &builder{f: f, units: f.DesignUnits}
	if b.units == 0 {
		b.units = Unity
	}
	switch {
	case b.units < 0:
		return nil, nil, invalid("the number of design units is negative")
	case f.DesignSize < Unity:
		return nil, nil, invalid("the design size is less than 1")
	case !validString(f.CodingScheme, 40):
		return nil, nil, invalid("the header cannot hold the coding scheme")
	case !validString(f.Family, 20):
		return nil, nil, invalid("the header cannot hold the family")
	case len(f.Header) > 32767-18 || len(f.Params) > 32767:
		return nil, nil, invalid("the font is too large for the format")
	case len(f.Params) > 0 && f.Params[0] == minFix:
		return nil, nil, invalid("the slant is not above -2048")
	}
	p, err := f.assemble()
	if err != nil {
		return nil, nil, err
	}
	b.prog = p

	// the recipes
	var exten []Recipe
	if f.RecipeOrder == nil {
		for c, ch := range f.Chars {
			if ch != nil && ch.Tag == ExtTag {
				p.rem[c] = len(exten)
				exten = append(exten, ch.Recipe)
			}
		}
	} else {
		seen := [256]bool{}
		for _, c := range f.RecipeOrder {
			ch := f.Chars[c]
			if ch == nil || ch.Tag != ExtTag || seen[c] {
				return nil, nil, invalid("the order of the recipes names character %d, which has no recipe or is named twice", c)
			}
			seen[c] = true
			p.rem[c] = len(exten)
			exten = append(exten, ch.Recipe)
		}
		for c, ch := range f.Chars {
			if ch != nil && ch.Tag == ExtTag && !seen[c] {
				return nil, nil, invalid("the order of the recipes leaves out character %d", c)
			}
		}
	}

	// the characters which the characters lead to
	for c, ch := range f.Chars {
		if ch == nil {
			continue
		}
		switch p.tag[c] {
		case LigTag:
			err = b.checkProgram(c, p.rem[c])
		case ListTag:
			err = b.safety(c, ch.NextLarger, "next larger character")
		case ExtTag:
			r := ch.Recipe
			if r.Top > 0 {
				err = b.safety(c, r.Top, "top piece")
			}
			if err == nil && r.Mid > 0 {
				err = b.safety(c, r.Mid, "middle piece")
			}
			if err == nil && r.Bot > 0 {
				err = b.safety(c, r.Bot, "bottom piece")
			}
			if err == nil {
				err = b.safety(c, r.Rep, "repeated piece")
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}
	if p.bcharLabel < noLabel {
		if err := b.checkProgram(Boundary, p.bcharLabel); err != nil {
			return nil, nil, err
		}
	}
	if b.lc.full() {
		return nil, nil, invalid("the ligature and kerning program has more pairs of characters than PLtoTF holds")
	}
	if b.lc.run() {
		return nil, nil, invalid("the pair of characters %d and %d starts an infinite loop of ligatures", b.lc.xCycle, b.lc.yCycle)
	}
	for i, s := range p.lig {
		if s.Skip > 128 {
			continue
		}
		if !b.exists(s.Next) && !(f.HasBoundary && s.Next == f.BoundaryChar) {
			return nil, nil, invalid("step %d examines character %d, which the font does not have", i, s.Next)
		}
		if s.Op < 128 && !b.exists(s.Remainder) {
			return nil, nil, invalid("step %d inserts character %d, which the font does not have", i, s.Remainder)
		}
	}
	for c := range f.Chars {
		if p.tag[c] != ListTag {
			continue
		}
		g := p.rem[c]
		for g < c && p.tag[g] == ListTag {
			g = p.rem[g]
		}
		if g == c {
			return nil, nil, invalid("the list of larger characters of character %d is a cycle", c)
		}
	}

	// the tables of dimensions
	var wd, ht, dp, ic dimTable
	for _, ch := range f.Chars {
		if ch == nil {
			continue
		}
		wd.add(ch.Width)
		if ch.Height != 0 {
			ht.add(ch.Height)
		}
		if ch.Depth != 0 {
			dp.add(ch.Depth)
		}
		if ch.Italic != 0 {
			ic.add(ch.Italic)
		}
	}
	for _, x := range f.Unused[0] {
		wd.add(x)
	}
	for i, t := range []*dimTable{&ht, &dp, &ic} {
		for _, x := range f.Unused[i+1] {
			if x != 0 {
				t.add(x)
			}
		}
	}
	rep := &Report{}
	for _, t := range []struct {
		t    *dimTable
		m    int
		name string
	}{{&wd, 255, "width"}, {&ht, 15, "height"}, {&dp, 15, "depth"}, {&ic, 63, "italic correction"}} {
		n := len(t.t.vals)
		delta, change := t.t.reduce(t.m)
		if n = len(t.t.vals); n > t.m {
			rep.Reductions = append(rep.Reductions, Reduction{
				Table: t.name, From: n + 1, To: len(t.t.out) + 1, Tolerance: delta, MaxChange: change,
			})
		}
	}

	// the range of the character codes
	bc, ec := 0, 255
	for bc < 255 && f.Chars[bc] == nil {
		bc++
	}
	for ec > 0 && f.Chars[ec] == nil {
		ec--
	}
	if bc > ec {
		bc = 1
	}

	// the addresses of the programs
	type label struct {
		rr int
		cc int
	}
	labels := []label{{-1, 0}}
	for c := bc; c <= ec; c++ {
		if p.tag[c] == LigTag {
			labels = append(labels, label{p.rem[c], c})
		}
	}
	sort.SliceStable(labels, func(i, j int) bool { return labels[i].rr < labels[j].rr })
	labelPtr := len(labels) - 1
	lkOffset := 0
	extraLoc := false
	if f.HasBoundary {
		extraLoc = true
		lkOffset = 1
	}
	sortPtr := labelPtr
	if labels[sortPtr].rr+lkOffset > 255 {
		lkOffset = 0
		extraLoc = false
		for {
			p.rem[labels[sortPtr].cc] = lkOffset
			for labels[sortPtr-1].rr == labels[sortPtr].rr {
				sortPtr--
				p.rem[labels[sortPtr].cc] = lkOffset
			}
			lkOffset++
			sortPtr--
			if lkOffset+labels[sortPtr].rr < 256 {
				break
			}
		}
	}
	if lkOffset > 0 {
		for ; sortPtr > 0; sortPtr-- {
			p.rem[labels[sortPtr].cc] += lkOffset
		}
	}
	if len(p.lig)+lkOffset > 32767 {
		return nil, nil, invalid("the font is too large for the format")
	}
	if p.bcharLabel < noLabel {
		last := &p.lig[len(p.lig)-1]
		last.Op = uint8((p.bcharLabel + lkOffset) >> 8)
		last.Remainder = uint8(p.bcharLabel + lkOffset)
	}

	out := &File{First: bc, Last: ec}

	// the header
	if !f.HasChecksum {
		c0, c1, c2, c3 := int64(bc), int64(ec), int64(bc), int64(ec)
		for c := bc; c <= ec; c++ {
			ch := f.Chars[c]
			if ch == nil {
				continue
			}
			w := int64(wd.held[wd.find(ch.Width)])
			if b.units != Unity {
				w = roundHalfAway(float64(w) / float64(b.units) * 1048576.0)
			}
			w += int64(c+4) << 22
			c0 = (c0 + c0 + w) % 255
			c1 = (c1 + c1 + w) % 253
			c2 = (c2 + c2 + w) % 251
			c3 = (c3 + c3 + w) % 247
		}
		out.Header = append(out.Header, uint32(uint8(c0))<<24|uint32(uint8(c1))<<16|uint32(uint8(c2))<<8|uint32(uint8(c3)))
	} else {
		out.Header = append(out.Header, f.Checksum)
	}
	out.Header = append(out.Header, uint32(f.DesignSize))
	var text [60]byte
	text[0] = byte(len(f.CodingScheme))
	copy(text[1:40], f.CodingScheme)
	text[40] = byte(len(f.Family))
	copy(text[41:60], f.Family)
	for i := 0; i < 60; i += 4 {
		out.Header = append(out.Header, uint32(text[i])<<24|uint32(text[i+1])<<16|uint32(text[i+2])<<8|uint32(text[i+3]))
	}
	word17 := uint32(f.Face)
	if !b.unsafe {
		word17 |= 128 << 24
	}
	rep.SevenBitSafe = !b.unsafe
	out.Header = append(out.Header, word17)
	out.Header = append(out.Header, f.Header...)

	// the words of the characters
	if ec >= bc {
		out.Chars = make([]CharInfo, ec-bc+1)
	}
	for c := bc; c <= ec; c++ {
		ci := &out.Chars[c-bc]
		ci.Tag = p.tag[c]
		ci.Remainder = uint8(p.rem[c])
		ch := f.Chars[c]
		if ch == nil {
			continue
		}
		ci.WidthIndex = uint8(wd.index[wd.find(ch.Width)])
		if ch.Height != 0 {
			ci.HeightIndex = uint8(ht.index[ht.find(ch.Height)])
		}
		if ch.Depth != 0 {
			ci.DepthIndex = uint8(dp.index[dp.find(ch.Depth)])
		}
		if ch.Italic != 0 {
			ci.ItalicIndex = uint8(ic.index[ic.find(ch.Italic)])
		}
	}

	// the dimensions
	for _, t := range []struct {
		t    *dimTable
		dst  *[]Fix
		name string
	}{{&wd, &out.Width, "width"}, {&ht, &out.Height, "height"}, {&dp, &out.Depth, "depth"}, {&ic, &out.Italic, "italic correction"}} {
		*t.dst = make([]Fix, 1, len(t.t.out)+1)
		for _, x := range t.t.out {
			w, err := b.scaled(x, t.name)
			if err != nil {
				return nil, nil, err
			}
			*t.dst = append(*t.dst, w)
		}
	}

	// the ligature and kerning program
	if extraLoc {
		out.LigKern = append(out.LigKern, LigKern{255, f.BoundaryChar, 0, 0})
	} else {
		for i := 0; i < lkOffset; i++ {
			t := labels[labelPtr].rr
			w := LigKern{254, 0, uint8((t + lkOffset) >> 8), uint8(t + lkOffset)}
			if f.HasBoundary {
				w.Skip, w.Next = 255, f.BoundaryChar
			}
			out.LigKern = append(out.LigKern, w)
			for labelPtr--; labels[labelPtr].rr >= t; labelPtr-- {
			}
		}
	}
	out.LigKern = append(out.LigKern, p.lig...)
	for _, x := range p.kern {
		w, err := b.scaled(x, "kern")
		if err != nil {
			return nil, nil, err
		}
		out.Kern = append(out.Kern, w)
	}
	out.Exten = exten

	// the parameters
	for i, x := range f.Params {
		if i == 0 {
			v := int64(x)
			top := v / (1 << 24)
			if v < 0 {
				v += 1 << 30
				top = v/(1<<24) + 256 - 64
			}
			out.Param = append(out.Param, Fix(int32(uint32(uint8(top))<<24|uint32(uint8(v/65536%256))<<16|uint32(uint8(v/256%256))<<8|uint32(uint8(v%256)))))
			continue
		}
		w, err := b.scaled(x, "parameter")
		if err != nil {
			return nil, nil, err
		}
		out.Param = append(out.Param, w)
	}

	if err := out.validate(); err != nil {
		return nil, nil, err
	}
	return out, rep, nil
}

// Write builds the TFM file of the description and writes it to w.  See
// [Font.Build] for the method and the errors.
func (f *Font) Write(w io.Writer) (*Report, error) {
	file, rep, err := f.Build()
	if err != nil {
		return nil, err
	}
	if err := file.Write(w); err != nil {
		return nil, err
	}
	return rep, nil
}
