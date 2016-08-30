package object

import (
	"github.com/hirochachacha/plua/opcode"
)

type Proto struct {
	Code      []opcode.Instruction
	Constants []Value
	Protos    []*Proto
	Upvalues  []UpvalueDesc

	Source          string
	LineDefined     int
	LastLineDefined int
	NParams         int
	IsVararg        bool
	MaxStackSize    int
	LineInfo        []int
	LocVars         []LocVar
}

type UpvalueDesc struct {
	Name    string
	Instack bool
	Index   int
}

type LocVar struct {
	Name    string
	StartPC int
	EndPC   int
}
