package pattern

import (
	"fmt"
	"unicode/utf8"
)

const (
	opMatch rune = iota + utf8.MaxRune + 1
	opJmp
	opSplit

	opAny

	opLetter
	opControl
	opDigit
	opGraphic
	opLower
	opPunct
	opSpace
	opUpper
	opAlphaNum
	opHexDigit

	opNotLetter
	opNotControl
	opNotDigit
	opNotGraphic
	opNotLower
	opNotPunct
	opNotSpace
	opNotUpper
	opNotAlphaNum
	opNotHexDigit

	opRange
	opNotRange

	opBalanceAny
	opBalanceUp
	opBalanceDown

	opCapture

	opFrontier

	// save
	opEnterSave
	opExitSave
)

type inst struct {
	op rune

	x int
	y int
}

func (ins inst) String() string {
	switch ins.op {
	case opMatch:
		return fmt.Sprintf("op: match")
	case opJmp:
		return fmt.Sprintf("op: jmp, x: %d", ins.x)
	case opSplit:
		return fmt.Sprintf("op: split, x: %d, y: %d", ins.x, ins.y)
	case opAny:
		return fmt.Sprintf("op: any")
	case opBalanceAny:
		return fmt.Sprintf("op: balance any, x: %d, y: %d", ins.x, ins.y)
	case opLetter:
		return fmt.Sprintf("op: letter")
	case opControl:
		return fmt.Sprintf("op: control")
	case opDigit:
		return fmt.Sprintf("op: digit")
	case opGraphic:
		return fmt.Sprintf("op: graphic")
	case opLower:
		return fmt.Sprintf("op: lower")
	case opPunct:
		return fmt.Sprintf("op: punctuation")
	case opSpace:
		return fmt.Sprintf("op: space")
	case opUpper:
		return fmt.Sprintf("op: upper")
	case opAlphaNum:
		return fmt.Sprintf("op: word")
	case opHexDigit:
		return fmt.Sprintf("op: hex digit")
	case opNotLetter:
		return fmt.Sprintf("op: not letter")
	case opNotControl:
		return fmt.Sprintf("op: not control")
	case opNotDigit:
		return fmt.Sprintf("op: not digit")
	case opNotGraphic:
		return fmt.Sprintf("op: not graphic")
	case opNotLower:
		return fmt.Sprintf("op: not lower")
	case opNotPunct:
		return fmt.Sprintf("op: not punctuation")
	case opNotSpace:
		return fmt.Sprintf("op: not space")
	case opNotUpper:
		return fmt.Sprintf("op: not upper")
	case opNotAlphaNum:
		return fmt.Sprintf("op: not word")
	case opNotHexDigit:
		return fmt.Sprintf("op: not hex digit")
	case opRange:
		return fmt.Sprintf("op: range, x: %d", ins.x)
	case opNotRange:
		return fmt.Sprintf("op: range, x: %d", ins.x)
	case opFrontier:
		return fmt.Sprintf("op: frontier, x: %d", ins.x)
	case opBalanceUp:
		return fmt.Sprintf("op: balance up, x: %d", ins.x)
	case opBalanceDown:
		return fmt.Sprintf("op: balance down, x: %d, y: %d", ins.x, ins.y)
	case opEnterSave:
		return fmt.Sprintf("op: enter save, x: %d, y: %d", ins.x, ins.y)
	case opExitSave:
		return fmt.Sprintf("op: exit save, x: %d, y: %d", ins.x, ins.y)
	case opCapture:
		return fmt.Sprintf("op: capture, x : %d", ins.x)
	default:
		if ins.op > utf8.MaxRune {
			return fmt.Sprintf("unknown op: %d", ins.op)
		}
		return fmt.Sprintf("op: char, r: %s", string(ins.op))
	}
}

type rangeTable struct {
	r32 [][2]rune
}

func (rt *rangeTable) is(r rune, preds predicates) bool {
	for _, t := range rt.r32 {
		if t[0] < utf8.MaxRune {
			if r < t[0] {
				continue
			}
			if r <= t[1] {
				return true
			}

			continue
		}

		if b, ok := preds.is(t[0], r); ok {
			return b
		}
	}

	return false
}

func (rt *rangeTable) String() string {
	return fmt.Sprint(rt.r32)
}
