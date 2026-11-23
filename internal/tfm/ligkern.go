// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tfm

import "sort"

// Op is the operation of a step of a ligature and kerning program.
//
// A ligature operation inserts a character between the current character and
// the character to its right.  Its name in a PL file has a slash on the side
// of each of the two characters which it keeps, and one > sign for each
// character which TeX passes over before it seeks the next step.
type Op uint8

// The operations of a step.
const (
	OpLig                Op = 0   // LIG: replaces both characters
	OpLigSlash           Op = 1   // LIG/: keeps the right character
	OpSlashLig           Op = 2   // /LIG: keeps the left character
	OpSlashLigSlash      Op = 3   // /LIG/: keeps both characters
	OpLigSlashOver       Op = 5   // LIG/>: keeps the right character, passes over one character
	OpSlashLigOver       Op = 6   // /LIG>: keeps the left character, passes over one character
	OpSlashLigSlashOver  Op = 7   // /LIG/>: keeps both characters, passes over one character
	OpSlashLigSlashOver2 Op = 11  // /LIG/>>: keeps both characters, passes over two characters
	OpKern               Op = 128 // KRN: inserts a kern
)

// standardOp returns the operation of an operation byte below 128; a byte
// which is none of the eight ligature operations stands for OpLig.
func standardOp(b uint8) Op {
	if b == 4 || (b > 7 && b != 11) {
		return OpLig
	}
	return Op(b)
}

// IsLig reports whether op is one of the eight ligature operations.
func (op Op) IsLig() bool {
	return op <= 3 || (op >= 5 && op <= 7) || op == 11
}

// String returns the name of the operation in a PL file.
func (op Op) String() string {
	if op == OpKern {
		return "KRN"
	}
	if !op.IsLig() {
		return "?"
	}
	s := "LIG"
	if op&2 != 0 {
		s = "/" + s
	}
	if op&1 != 0 {
		s += "/"
	}
	for n := op >> 2; n > 0; n-- {
		s += ">"
	}
	return s
}

// Step is one step of a ligature and kerning program: if the character Next
// follows the current character, the operation is carried out and the program
// ends.
type Step struct {
	Next   uint8 // the character to the right
	Op     Op
	Result uint8 // the character which a ligature operation inserts
	Kern   Fix   // the amount of a kern
}

// The states of a word of the ligature and kerning program.
const (
	unreachable = iota // no program runs through the word
	passThrough        // the word holds an address or the boundary character
	accessible         // a program runs through the word
)

// The kinds of a listItem.
const (
	itemLabel  = iota // a label; char is the character or Boundary
	itemStep          // a ligature or kern step
	itemStop          // the end of a program
	itemSkip          // n steps to pass over
	itemUnused        // the start of a run of steps which no program reaches
	itemUsed          // the end of such a run
)

// listItem is one item of the listing of a ligature and kerning program.
type listItem struct {
	kind int
	char int
	n    int
	step Step
}

// label is the start of a program.
type label struct {
	char int
	at   int
}

// listing holds the ligature and kerning program of a file in the form in
// which TFtoPL lists it.
type listing struct {
	lig      []LigKern // the program, with every operation byte standard
	activity []uint8
	labels   []label // sorted by address
	items    []listItem
	bchar    int // the boundary character, or -1
	bstart   int // the start of the program of the left boundary, or -1
	oddOps   int // the number of steps whose operation byte is not standard
	unused   int // the number of steps which no program reaches
}

// start returns the address of the first step of the program of c.
func (l *listing) start(ci CharInfo) int {
	r := int(ci.Remainder)
	if l.lig[r].Skip > 128 {
		r = int(l.lig[r].Op)<<8 | int(l.lig[r].Remainder)
	}
	return r
}

// list analyses the ligature and kerning program of a valid file.
func (f *File) list() *listing {
	nl := len(f.LigKern)
	l := &listing{bchar: -1, bstart: -1}
	l.lig = append([]LigKern(nil), f.LigKern...)
	l.activity = make([]uint8, nl)
	if nl > 0 {
		if l.lig[0].Skip == 255 {
			l.bchar = int(l.lig[0].Next)
			l.activity[0] = passThrough
		}
		if last := l.lig[nl-1]; last.Skip == 255 {
			r := int(last.Op)<<8 | int(last.Remainder)
			l.labels = append(l.labels, label{Boundary, r})
			l.bstart = r
			l.activity[r] = accessible
			l.activity[nl-1] = passThrough
		}
	}
	for i, ci := range f.Chars {
		if ci.Tag != LigTag {
			continue
		}
		r := int(ci.Remainder)
		if l.lig[r].Skip > 128 {
			if l.activity[r] == unreachable {
				l.activity[r] = passThrough
			}
			r = int(l.lig[r].Op)<<8 | int(l.lig[r].Remainder)
		}
		l.labels = append(l.labels, label{f.First + i, r})
		l.activity[r] = accessible
	}
	sort.SliceStable(l.labels, func(i, j int) bool { return l.labels[i].at < l.labels[j].at })
	for i := range l.lig {
		if l.activity[i] == accessible && l.lig[i].Skip < 128 {
			l.activity[i+int(l.lig[i].Skip)+1] = accessible
		}
	}

	next := 0
	inUnused := false
	for i := range l.lig {
		if l.activity[i] == passThrough {
			continue
		}
		if l.activity[i] == unreachable {
			l.unused++
			if !inUnused {
				l.items = append(l.items, listItem{kind: itemUnused})
				inUnused = true
			}
		} else if inUnused {
			l.items = append(l.items, listItem{kind: itemUsed})
			inUnused = false
		}
		for next < len(l.labels) && l.labels[next].at == i {
			l.items = append(l.items, listItem{kind: itemLabel, char: l.labels[next].char})
			next++
		}
		s := &l.lig[i]
		if s.Skip <= 128 {
			if s.Op < 128 && standardOp(s.Op) != Op(s.Op) {
				l.oddOps++
				s.Op = 0
			}
			l.items = append(l.items, listItem{kind: itemStep, step: f.step(*s)})
		}
		if s.Skip > 0 && !inUnused {
			if s.Skip >= 128 {
				l.items = append(l.items, listItem{kind: itemStop})
			} else {
				count := 0
				for j := i + 1; j <= i+int(s.Skip); j++ {
					if l.activity[j] == accessible {
						count++
					}
				}
				l.items = append(l.items, listItem{kind: itemSkip, n: count})
			}
		}
	}
	if inUnused {
		l.items = append(l.items, listItem{kind: itemUsed})
	}
	return l
}

// The classes of a pair of characters in a loopChecker.
const (
	classSimple  = iota // the result is z
	classLeft           // the result is that of the pair (z, y)
	classRight          // the result is that of the pair (x, z)
	classBoth           // the result is that of the pair (result of (x, z), y)
	classPending        // the result is being worked out
)

// maxPairs is the number of pairs beyond which TFtoPL and PLtoTF give up.
const maxPairs = 32579

// numKeys is the number of keys of pairs: 256·x+y+1 for x and y up to 257.
const numKeys = 256*257 + 257 + 2

type ligPair struct {
	set   bool
	class uint8
	z     uint16
}

// loopChecker finds the pairs of characters whose ligature steps never come
// to an end.  A pair (x, y) has the key 256·x+y+1.
type loopChecker struct {
	pairs  []ligPair
	order  []int32
	cycle  bool
	xCycle int
	yCycle int
}

// add enters a word of the program of character c.  It returns false if a
// word before it in the program has the same right character.
func (lc *loopChecker) add(c int, s LigKern) bool {
	if lc.pairs == nil {
		lc.pairs = make([]ligPair, numKeys)
	}
	if len(lc.order) >= maxPairs {
		return false
	}
	y := int(s.Next)
	p := ligPair{set: true, class: classSimple, z: uint16(s.Remainder)}
	if s.Op >= 128 {
		p.z = uint16(y)
	} else {
		switch s.Op {
		case 5, 11:
			p.z = uint16(y)
		case 1, 7:
			p.class = classLeft
		case 2:
			p.class = classRight
		case 3:
			p.class = classBoth
		}
	}
	key := 256*c + y + 1
	if lc.pairs[key].set {
		return false
	}
	lc.pairs[key] = p
	lc.order = append(lc.order, int32(key))
	return true
}

// full reports whether the checker holds as many pairs as it accepts.
func (lc *loopChecker) full() bool {
	return len(lc.order) >= maxPairs
}

// run works out the result of every pair, and reports whether a pair leads
// into an infinite loop.
func (lc *loopChecker) run() bool {
	for _, key := range lc.order {
		if p := &lc.pairs[key]; p.class > classSimple {
			lc.result(p, int(key-1)/256, int(key-1)%256)
		}
	}
	return lc.cycle
}

func (lc *loopChecker) eval(x, y int) int {
	p := &lc.pairs[256*x+y+1]
	if !p.set {
		return y
	}
	return lc.result(p, x, y)
}

func (lc *loopChecker) result(p *ligPair, x, y int) int {
	switch p.class {
	case classLeft:
		p.class = classPending
		p.z = uint16(lc.eval(int(p.z), y))
		p.class = classSimple
	case classRight:
		p.class = classPending
		p.z = uint16(lc.eval(x, int(p.z)))
		p.class = classSimple
	case classBoth:
		p.class = classPending
		p.z = uint16(lc.eval(lc.eval(x, int(p.z)), y))
		p.class = classSimple
	case classPending:
		lc.cycle = true
		lc.xCycle, lc.yCycle = x, y
		p.z = 257
		p.class = classSimple
	}
	return int(p.z)
}

// loops enters the programs of the file into a loopChecker, as TFtoPL does.
func (f *File) loops(l *listing) *loopChecker {
	lc := &loopChecker{}
	nl := len(l.lig)
	enter := func(c, i int) {
		for i < nl {
			lc.add(c, l.lig[i])
			if l.lig[i].Skip >= 128 {
				break
			}
			i += int(l.lig[i].Skip) + 1
		}
	}
	for i, ci := range f.Chars {
		if ci.Tag == LigTag {
			enter(f.First+i, l.start(ci))
		}
	}
	if l.bstart >= 0 {
		enter(Boundary, l.bstart)
	}
	return lc
}
