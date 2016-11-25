package pattern

import (
	"fmt"
	"unicode/utf8"
)

type opCode rune

const (
	opMatch opCode = iota + utf8.MaxRune + 1
	opJmp
	opSplit
	opBeginSave
	opEndSave
	opCapture

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
	opZero

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
	opNotZero

	opSet
	opFrontier
	opBalance
)

type instruction struct {
	op opCode
	x  int
	y  int
}

func (inst instruction) String() string {
	switch inst.op {
	case opMatch:
		return fmt.Sprintf("op: match")
	case opJmp:
		return fmt.Sprintf("op: jmp, x: %d", inst.x)
	case opSplit:
		return fmt.Sprintf("op: split, x: %d, y: %d", inst.x, inst.y)
	case opBeginSave:
		return fmt.Sprintf("op: begin save, x: %d", inst.x)
	case opEndSave:
		return fmt.Sprintf("op: end save, x: %d", inst.x)
	case opCapture:
		return fmt.Sprintf("op: capture")

	case opAny:
		return fmt.Sprintf("op: any")
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
	case opZero:
		return fmt.Sprintf("op: zero")

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
	case opNotZero:
		return fmt.Sprintf("op: not zero")

	case opSet:
		return fmt.Sprintf("op: set")
	case opFrontier:
		return fmt.Sprintf("op: frontier")
	case opBalance:
		return fmt.Sprintf("op: balance, x: %c, y: %c", rune(inst.x), rune(inst.y))

	default:
		if 0 <= inst.op && inst.op <= utf8.MaxRune {
			return fmt.Sprintf("op: %c", inst.op)
		}

		return fmt.Sprintf("op: unknown")
	}
}

func instChar(r rune) instruction {
	return instruction{op: opCode(r)}
}

func instAny() instruction {
	return instruction{op: opAny}
}

func instClassOrEscChar(r rune) (instruction, bool) {
	switch r {
	case eos:
		return instruction{}, false
	case 'a':
		return instruction{op: opLetter}, true
	case 'A':
		return instruction{op: opNotLetter}, true
	case 'c':
		return instruction{op: opControl}, true
	case 'C':
		return instruction{op: opNotControl}, true
	case 'd':
		return instruction{op: opDigit}, true
	case 'D':
		return instruction{op: opNotDigit}, true
	case 'g':
		return instruction{op: opGraphic}, true
	case 'G':
		return instruction{op: opNotGraphic}, true
	case 'l':
		return instruction{op: opLower}, true
	case 'L':
		return instruction{op: opNotLower}, true
	case 'p':
		return instruction{op: opPunct}, true
	case 'P':
		return instruction{op: opNotPunct}, true
	case 's':
		return instruction{op: opSpace}, true
	case 'S':
		return instruction{op: opNotSpace}, true
	case 'u':
		return instruction{op: opUpper}, true
	case 'U':
		return instruction{op: opNotUpper}, true
	case 'w':
		return instruction{op: opAlphaNum}, true
	case 'W':
		return instruction{op: opNotAlphaNum}, true
	case 'x':
		return instruction{op: opHexDigit}, true
	case 'X':
		return instruction{op: opNotHexDigit}, true
	case 'z':
		return instruction{op: opZero}, true
	case 'Z':
		return instruction{op: opNotZero}, true
	default:
		if !isAlphaNum(r) {
			return instChar(r), true
		}
		return instruction{}, false
	}
}

func instMatch() instruction {
	return instruction{
		op: opMatch,
	}
}

// x: next instruction, higher precedence than y
// y: next instruction
func instSplit(x, y int) instruction {
	return instruction{
		op: opSplit,
		x:  x,
		y:  y,
	}
}

// x: next instruction
func instJmp(x int) instruction {
	return instruction{
		op: opJmp,
		x:  x,
	}
}

// x: depth
func instBeginSave(x int) instruction {
	return instruction{
		op: opBeginSave,
		x:  x,
	}
}

// x: depth
// y: 1 if the paren is empty
func instEndSave(x, y int) instruction {
	return instruction{
		op: opEndSave,
		x:  x,
		y:  y,
	}
}

// x: capture number
func instCapture(x int) instruction {
	return instruction{
		op: opCapture,
		x:  x,
	}
}

func instSet(x int) instruction {
	return instruction{
		op: opSet,
		x:  x,
	}
}

func instFrontier(x int) instruction {
	return instruction{
		op: opFrontier,
		x:  x,
	}
}

func instBalance(x, y int) instruction {
	return instruction{
		op: opBalance,
		x:  x,
		y:  y,
	}
}
