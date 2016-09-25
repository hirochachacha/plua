package opcode

// assertion flag for debugging
const assert = false

const (
	InstructionSize = 4 // sizeof Instruction (bytes)
	MaxOpcode       = EXTRAARG + 1

	// bit size
	SizeA  = 8
	SizeB  = 9
	SizeC  = 9
	SizeBx = SizeB + SizeC
	SizeAx = SizeA + SizeB + SizeC
	SizeOp = 6

	MaxA   = (1 << SizeA) - 1
	MaxB   = (1 << SizeB) - 1
	MaxC   = (1 << SizeC) - 1
	MaxBx  = (1 << SizeBx) - 1
	MaxSBx = MaxBx >> 1
	MaxAx  = (1 << SizeAx) - 1

	BitRK      = 256
	MaxRKindex = BitRK - 1
)

type Instruction uint32

func ABC(op OpCode, a, b, c int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if a > MaxA || a < 0 {
			panic("a overflow")
		}
		if b > MaxB || b < 0 {
			panic("b overflow")
		}
		if c > MaxC || c < 0 {
			panic("c overflow")
		}
	}
	return Instruction(int(op) | (a << 6) | (b << 23) | (c << 14))
}

func AB(op OpCode, a, b int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if a > MaxA || a < 0 {
			panic("a overflow")
		}
		if b > MaxB || b < 0 {
			panic("b overflow")
		}
	}
	return Instruction(int(op) | (a << 6) | (b << 23))
}

func AC(op OpCode, a, c int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if a > MaxA || a < 0 {
			panic("a overflow")
		}
		if c > MaxC || c < 0 {
			panic("c overflow")
		}
	}
	return Instruction(int(op) | (a << 6) | (c << 14))
}

func ABx(op OpCode, a, bx int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if bx > MaxBx || bx < 0 {
			panic("bx overflow")
		}
	}
	return Instruction(int(op) | (a << 6) | (bx << 14))
}

func AsBx(op OpCode, a, sbx int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if sbx > MaxSBx || sbx < -MaxSBx {
			panic("sbx overflow")
		}
	}
	return ABx(op, a, sbx+0x1ffff)
}

func Ax(op OpCode, ax int) Instruction {
	if assert {
		if op > MaxOpcode || op < 0 {
			panic("opcode overflow")
		}
		if ax > MaxAx || ax < 0 {
			panic("ax overflow")
		}
	}
	return Instruction(int(op) | (ax << 6))
}

func (i Instruction) OpCode() OpCode {
	return OpCode(int(i) & 0x3f)
}

func (i Instruction) OpName() string {
	return i.OpCode().Name()
}

func (i Instruction) A() int {
	return int(i) >> 6 & 0xff
}

func (i Instruction) B() int {
	return int(i) >> 23
}

func (i Instruction) C() int {
	return (int(i) >> 14) & 0x1ff
}

func (i Instruction) Bx() int {
	return int(i) >> 14
}

func (i Instruction) Ax() int {
	return int(i) >> 6
}

func (i Instruction) SBx() int {
	return int(i.Bx()) - 0x1ffff
}

func (i Instruction) OpMode() OpMode {
	return i.OpCode().OpMode()
}

func (i Instruction) BMode() OpArgMask {
	return i.OpCode().BMode()
}

func (i Instruction) CMode() OpArgMask {
	return i.OpCode().CMode()
}

func (i Instruction) TestAMode() bool {
	return i.OpCode().TestAMode()
}

func (i Instruction) TestTMode() bool {
	return i.OpCode().TestTMode()
}

type OpCode uint

func (op OpCode) Name() string {
	return opNames[op]
}

func (op OpCode) OpMode() OpMode {
	return OpMode(opModes[op] & 3)
}

func (op OpCode) BMode() OpArgMask {
	return OpArgMask((opModes[op] >> 4) & 3)
}

func (op OpCode) CMode() OpArgMask {
	return OpArgMask((opModes[op] >> 2) & 3)
}

func (op OpCode) TestAMode() bool {
	return (opModes[op] & (1 << 6)) != 0
}

func (op OpCode) TestTMode() bool {
	return (opModes[op] & (1 << 7)) != 0
}

const (
	MOVE     OpCode = iota /*	A B 	R(A) := R(B)					*/
	LOADK                  /*	A Bx	R(A) := Kst(Bx)					*/
	LOADKX                 /*	A 		R(A) := Kst(extra arg)				*/
	LOADBOOL               /*	A B C	R(A) := (Bool)B; if (C) pc++			*/
	LOADNIL                /*	A B		R(A), R(A+1), ..., R(A+B) := nil		*/
	GETUPVAL               /*	A B		R(A) := UpValue[B]				*/
	GETTABUP               /*	A B C	R(A) := UpValue[B][RK(C)]			*/
	GETTABLE               /*	A B C	R(A) := R(B)[RK(C)]				*/
	SETTABUP               /*	A B C	UpValue[A][RK(B)] := RK(C)			*/
	SETUPVAL               /*	A B		UpValue[B] := R(A)				*/
	SETTABLE               /*	A B C	R(A)[RK(B)] := RK(C)				*/
	NEWTABLE               /*	A B C	R(A) := {} (size = B,C)				*/
	SELF                   /*	A B C	R(A+1) := R(B); R(A) := R(B)[RK(C)]		*/
	ADD                    /*	A B C	R(A) := RK(B) + RK(C)				*/
	SUB                    /*	A B C	R(A) := RK(B) - RK(C)				*/
	MUL                    /*	A B C	R(A) := RK(B) * RK(C)				*/
	MOD                    /*	A B C	R(A) := RK(B) % RK(C)				*/
	POW                    /*	A B C	R(A) := RK(B) ^ RK(C)				*/
	DIV                    /*	A B C	R(A) := RK(B) / RK(C)				*/
	IDIV                   /*	A B C	R(A) := RK(B) // RK(C)				*/
	BAND                   /*	A B C	R(A) := RK(B) & RK(C)				*/
	BOR                    /*	A B C	R(A) := RK(B) | RK(C)				*/
	BXOR                   /*	A B C	R(A) := RK(B) ~ RK(C)				*/
	SHL                    /*	A B C	R(A) := RK(B) << RK(C)				*/
	SHR                    /*	A B C	R(A) := RK(B) >> RK(C)				*/
	UNM                    /*	A B		R(A) := -R(B)					*/
	BNOT                   /*	A B		R(A) := ~R(B)					*/
	NOT                    /*	A B		R(A) := not R(B)				*/
	LEN                    /*	A B		R(A) := length of R(B)				*/
	CONCAT                 /*	A B C	R(A) := R(B).. ... ..R(C)			*/
	JMP                    /*	A sBx	pc+=sBx; if (A) close all upvalues >= R(A) + 1	*/
	EQ                     /*	A B C	if ((RK(B) == RK(C)) ~= A) then pc++		*/
	LT                     /*	A B C	if ((RK(B) <  RK(C)) ~= A) then pc++		*/
	LE                     /*	A B C	if ((RK(B) <= RK(C)) ~= A) then pc++		*/
	TEST                   /*	A C		if not (R(A) <=> C) then pc++			*/
	TESTSET                /*	A B C	if (R(B) <=> C) then R(A) := R(B) else pc++	*/
	CALL                   /*	A B C	R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1)) */
	TAILCALL               /*	A B C	return R(A)(R(A+1), ... ,R(A+B-1))		*/
	RETURN                 /*	A B		return R(A), ... ,R(A+B-2)	(see note)	*/
	FORLOOP                /*	A sBx	R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }*/
	FORPREP                /*	A sBx	R(A)-=R(A+2); pc+=sBx				*/
	TFORCALL               /*	A C		R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));	*/
	TFORLOOP               /*	A sBx	if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }*/
	SETLIST                /*	A B C	R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B	*/
	CLOSURE                /*	A Bx	R(A) := closure(KPROTO[Bx])			*/
	VARARG                 /*	A B		R(A), R(A+1), ..., R(A+B-2) = vararg		*/
	EXTRAARG               /*	Ax		extra (larger) argument for previous opcode	*/
)

var opNames = [...]string{
	MOVE:     "MOVE",
	LOADK:    "LOADK",
	LOADKX:   "LOADKX",
	LOADBOOL: "LOADBOOL",
	LOADNIL:  "LOADNIL",
	GETUPVAL: "GETUPVAL",
	GETTABUP: "GETTABUP",
	GETTABLE: "GETTABLE",
	SETTABUP: "SETTABUP",
	SETUPVAL: "SETUPVAL",
	SETTABLE: "SETTABLE",
	NEWTABLE: "NEWTABLE",
	SELF:     "SELF",
	ADD:      "ADD",
	SUB:      "SUB",
	MUL:      "MUL",
	MOD:      "MOD",
	POW:      "POW",
	DIV:      "DIV",
	IDIV:     "IDIV",
	BAND:     "BAND",
	BOR:      "BOR",
	BXOR:     "BXOR",
	SHL:      "SHL",
	SHR:      "SHR",
	UNM:      "UNM",
	BNOT:     "BNOT",
	NOT:      "NOT",
	LEN:      "LEN",
	CONCAT:   "CONCAT",
	JMP:      "JMP",
	EQ:       "EQ",
	LT:       "LT",
	LE:       "LE",
	TEST:     "TEST",
	TESTSET:  "TESTSET",
	CALL:     "CALL",
	TAILCALL: "TAILCALL",
	RETURN:   "RETURN",
	FORLOOP:  "FORLOOP",
	FORPREP:  "FORPREP",
	TFORCALL: "TFORCALL",
	TFORLOOP: "TFORLOOP",
	SETLIST:  "SETLIST",
	CLOSURE:  "CLOSURE",
	VARARG:   "VARARG",
	EXTRAARG: "EXTRAARG",
}

type OpMode int

const (
	IABC OpMode = iota
	IABx
	IAsBx
	IAx
)

type OpArgMask int

const (
	OpArgN OpArgMask = iota /* the argument is not used */
	OpArgU                  /* the argument is used */
	OpArgR                  /* the argument is a register or a jump offset */
	OpArgK                  /* the argument is a register or a constant */
)

func opMode(t, a int, b, c OpArgMask, m OpMode) int {
	return (t << 7) | (a << 6) | (int(b) << 4) | (int(c) << 2) | int(m)
}

var opModes = [...]int{
	MOVE:     opMode(0, 1, OpArgR, OpArgN, IABC),
	LOADK:    opMode(0, 1, OpArgK, OpArgN, IABx),
	LOADKX:   opMode(0, 1, OpArgN, OpArgN, IABx),
	LOADBOOL: opMode(0, 1, OpArgU, OpArgU, IABC),
	LOADNIL:  opMode(0, 1, OpArgU, OpArgN, IABC),
	GETUPVAL: opMode(0, 1, OpArgU, OpArgN, IABC),
	GETTABUP: opMode(0, 1, OpArgU, OpArgK, IABC),
	GETTABLE: opMode(0, 1, OpArgR, OpArgK, IABC),
	SETTABUP: opMode(0, 0, OpArgK, OpArgK, IABC),
	SETUPVAL: opMode(0, 0, OpArgU, OpArgN, IABC),
	SETTABLE: opMode(0, 0, OpArgK, OpArgK, IABC),
	NEWTABLE: opMode(0, 1, OpArgU, OpArgU, IABC),
	SELF:     opMode(0, 1, OpArgR, OpArgK, IABC),
	ADD:      opMode(0, 1, OpArgK, OpArgK, IABC),
	SUB:      opMode(0, 1, OpArgK, OpArgK, IABC),
	MUL:      opMode(0, 1, OpArgK, OpArgK, IABC),
	MOD:      opMode(0, 1, OpArgK, OpArgK, IABC),
	POW:      opMode(0, 1, OpArgK, OpArgK, IABC),
	DIV:      opMode(0, 1, OpArgK, OpArgK, IABC),
	IDIV:     opMode(0, 1, OpArgK, OpArgK, IABC),
	BAND:     opMode(0, 1, OpArgK, OpArgK, IABC),
	BOR:      opMode(0, 1, OpArgK, OpArgK, IABC),
	BXOR:     opMode(0, 1, OpArgK, OpArgK, IABC),
	SHL:      opMode(0, 1, OpArgK, OpArgK, IABC),
	SHR:      opMode(0, 1, OpArgK, OpArgK, IABC),
	UNM:      opMode(0, 1, OpArgR, OpArgN, IABC),
	BNOT:     opMode(0, 1, OpArgR, OpArgN, IABC),
	NOT:      opMode(0, 1, OpArgR, OpArgN, IABC),
	LEN:      opMode(0, 1, OpArgR, OpArgN, IABC),
	CONCAT:   opMode(0, 1, OpArgR, OpArgR, IABC),
	JMP:      opMode(0, 1, OpArgR, OpArgN, IAsBx),
	EQ:       opMode(1, 0, OpArgK, OpArgK, IABC),
	LT:       opMode(1, 0, OpArgK, OpArgK, IABC),
	LE:       opMode(1, 0, OpArgK, OpArgK, IABC),
	TEST:     opMode(1, 0, OpArgN, OpArgU, IABC),
	TESTSET:  opMode(1, 1, OpArgR, OpArgU, IABC),
	CALL:     opMode(0, 1, OpArgU, OpArgU, IABC),
	TAILCALL: opMode(0, 1, OpArgU, OpArgU, IABC),
	RETURN:   opMode(0, 0, OpArgU, OpArgN, IABC),
	FORLOOP:  opMode(0, 1, OpArgR, OpArgN, IAsBx),
	FORPREP:  opMode(0, 1, OpArgR, OpArgN, IAsBx),
	TFORCALL: opMode(0, 0, OpArgN, OpArgU, IABC),
	TFORLOOP: opMode(0, 1, OpArgR, OpArgN, IAsBx),
	SETLIST:  opMode(0, 0, OpArgU, OpArgU, IABC),
	CLOSURE:  opMode(0, 1, OpArgU, OpArgN, IABx),
	VARARG:   opMode(0, 1, OpArgU, OpArgN, IABC),
	EXTRAARG: opMode(0, 0, OpArgU, OpArgU, IAx),
}
