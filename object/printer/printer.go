package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

func Fprint(w io.Writer, x interface{}) {
	printer{w}.print(x)
}

func Print(x interface{}) {
	printer{os.Stdout}.print(x)
}

type printer struct {
	w io.Writer
}

func (pr printer) print(x interface{}) {
	switch x := x.(type) {
	case *object.Proto:
		pr.printFunc(x)
	case *object.DebugInfo:
		pr.printDebug(x)
	case object.Value:
		fmt.Fprintf(pr.w, "%s value = %s", object.ToType(x), object.Repr(x))
		pr.printValue(x)
		fmt.Fprintln(pr.w)
	case object.Process:
		// TODO
	}
}

func (pr printer) printDebug(d *object.DebugInfo) {
	fmt.Fprintf(pr.w, "debug info %p = {\n", d)
	fmt.Fprintf(pr.w, "\tName            =\t%s\n", d.Name)
	fmt.Fprintf(pr.w, "\tNameWhat        =\t%s\n", d.NameWhat)
	fmt.Fprintf(pr.w, "\tWhat            =\t%s\n", d.What)
	fmt.Fprintf(pr.w, "\tSource          =\t%s\n", d.Source)
	fmt.Fprintf(pr.w, "\tCurrentLine     =\t%d\n", d.CurrentLine)
	fmt.Fprintf(pr.w, "\tLineDefined     =\t%d\n", d.LineDefined)
	fmt.Fprintf(pr.w, "\tLastLineDefined =\t%d\n", d.LastLineDefined)
	fmt.Fprintf(pr.w, "\tNUpvalues       =\t%d\n", d.NUpvalues)
	fmt.Fprintf(pr.w, "\tNParams         =\t%d\n", d.NParams)
	fmt.Fprintf(pr.w, "\tIsVararg        =\t%t\n", d.IsVararg)
	fmt.Fprintf(pr.w, "\tIsTailCall      =\t%t\n", d.IsTailCall)
	fmt.Fprintf(pr.w, "\tShortSource     =\t%s\n", d.ShortSource)
	fmt.Fprintf(pr.w, "\tFunc            =\t%s\n", object.Repr(d.Func))
	fmt.Fprintf(pr.w, "}\n")
}

func (pr printer) printFunc(p *object.Proto) {
	pr.printHeader(p)
	pr.printCode(p)
	pr.printConstants(p)
	pr.printLocals(p)
	pr.printUpvalues(p)
	pr.printProtos(p)
}

func (pr printer) printHeader(p *object.Proto) {
	s := "=?"
	if len(p.Source) != 0 {
		s = string(p.Source)
	}

	if s[0] == '@' || s[0] == '=' {
		s = s[1:]
	} else if s[:len(version.LUA_SIGNATURE)] == version.LUA_SIGNATURE {
		s = "(bstring)"
	} else {
		s = "(string)"
	}

	var typ string
	if p.LineDefined == 0 {
		typ = "main"
	} else {
		typ = "function"
	}

	fmt.Fprintf(
		pr.w,
		"\n%s <%s:%d,%d> (%d instructions at %p)\n",
		typ, s, p.LineDefined, p.LastLineDefined, len(p.Code), p,
	)

	var vararg string
	if p.IsVararg {
		vararg = "+"
	}

	fmt.Fprintf(
		pr.w,
		"%d%s params, %d slots, %d upvalues, ",
		p.NParams, vararg, p.MaxStackSize, len(p.Upvalues),
	)

	fmt.Fprintf(
		pr.w,
		"%d locals, %d constants, %d functions\n",
		len(p.LocVars), len(p.Constants), len(p.Protos),
	)
}

func (pr printer) printValue(val object.Value) {
	if val, ok := val.(object.String); ok {
		fmt.Fprint(pr.w, strconv.Quote(string(val)))

		return
	}

	fmt.Fprint(pr.w, object.Repr(val))
}

func (pr printer) printCode(p *object.Proto) {
	for pc, code := range p.Code {
		a := code.A()
		b := code.B()
		c := code.C()
		bx := code.Bx()
		ax := code.Ax()
		sbx := code.SBx()

		fmt.Fprintf(pr.w, "\t%d\t", pc+1)

		if p.LineInfo != nil {
			fmt.Fprintf(pr.w, "[%d]\t", p.LineInfo[pc])
		} else {
			fmt.Fprintf(pr.w, "[-]\t")
		}

		fmt.Fprintf(pr.w, "%-9s\t", code.OpName())

		switch code.OpMode() {
		case opcode.IABC:
			fmt.Fprintf(pr.w, "%d", a)
			if code.BMode() != opcode.OpArgN {
				if b&opcode.BitRK != 0 {
					fmt.Fprintf(pr.w, " %d", -1-(b & ^opcode.BitRK))
				} else {
					fmt.Fprintf(pr.w, " %d", b)
				}
			}
			if code.CMode() != opcode.OpArgN {
				if c&opcode.BitRK != 0 {
					fmt.Fprintf(pr.w, " %d", -1-(c & ^opcode.BitRK))
				} else {
					fmt.Fprintf(pr.w, " %d", c)
				}
			}
		case opcode.IABx:
			fmt.Fprintf(pr.w, "%d", a)
			switch code.BMode() {
			case opcode.OpArgK:
				fmt.Fprintf(pr.w, " %d", -1-bx)
			case opcode.OpArgU:
				fmt.Fprintf(pr.w, " %d", bx)
			}
		case opcode.IAsBx:
			fmt.Fprintf(pr.w, "%d %d", a, sbx)
		case opcode.IAx:
			fmt.Fprintf(pr.w, "%d", -1-ax)
		default:
			panic("unreachable")
		}

		switch code.OpCode() {
		case opcode.LOADK:
			fmt.Fprint(pr.w, "\t; ")
			pr.printValue(p.Constants[bx])
		case opcode.GETUPVAL, opcode.SETUPVAL:
			fmt.Fprint(pr.w, "\t; ")
			fmt.Fprint(pr.w, upvalName(p, b))
		case opcode.GETTABUP:
			fmt.Fprint(pr.w, "\t; ")
			fmt.Fprint(pr.w, upvalName(p, b))
			if c&opcode.BitRK != 0 {
				fmt.Fprint(pr.w, " ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.SETTABUP:
			fmt.Fprint(pr.w, "\t; ")
			fmt.Fprint(pr.w, upvalName(p, a))
			if b&opcode.BitRK != 0 {
				fmt.Fprint(pr.w, " ")
				pr.printValue(p.Constants[b & ^opcode.BitRK])
			}
			if c&opcode.BitRK != 0 {
				fmt.Fprint(pr.w, " ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.GETTABLE, opcode.SELF:
			if c&opcode.BitRK != 0 {
				fmt.Fprint(pr.w, "\t; ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.SETTABLE, opcode.ADD, opcode.SUB, opcode.MUL,
			opcode.POW, opcode.DIV, opcode.IDIV, opcode.BAND,
			opcode.BOR, opcode.BXOR, opcode.SHL, opcode.SHR,
			opcode.EQ, opcode.LT, opcode.LE:
			if b&opcode.BitRK != 0 || c&opcode.BitRK != 0 {
				fmt.Fprint(pr.w, "\t; ")
				if b&opcode.BitRK != 0 {
					pr.printValue(p.Constants[b & ^opcode.BitRK])
				} else {
					fmt.Fprint(pr.w, "-")
				}

				fmt.Fprint(pr.w, " ")

				if c&opcode.BitRK != 0 {
					pr.printValue(p.Constants[c & ^opcode.BitRK])
				} else {
					fmt.Fprint(pr.w, "-")
				}
			}
		case opcode.JMP, opcode.FORLOOP, opcode.FORPREP, opcode.TFORLOOP:
			fmt.Fprintf(pr.w, "\t; to %d", sbx+pc+2)
		case opcode.CLOSURE:
			fmt.Fprintf(pr.w, "\t; %p", p.Protos[bx])
		case opcode.SETLIST:
			if c == 0 {
				pc++
				fmt.Fprintf(pr.w, "\t; %d", p.Code[pc])
			} else {
				fmt.Fprintf(pr.w, "\t; %d", c)
			}
		}

		fmt.Fprint(pr.w, "\n")
	}
}

func (pr printer) printConstants(p *object.Proto) {
	fmt.Fprintf(pr.w, "constants (%d) for %p: \n", len(p.Constants), p)
	for i, c := range p.Constants {
		fmt.Fprintf(pr.w, "\t%d\t", i+1)
		pr.printValue(c)
		fmt.Fprintln(pr.w)
	}
}

func (pr printer) printLocals(p *object.Proto) {
	fmt.Fprintf(pr.w, "locals (%d) for %p: \n", len(p.LocVars), p)
	for i, locvar := range p.LocVars {
		fmt.Fprintf(pr.w, "\t%d\t%s\t%d\t%d\n", i, locvar.Name, locvar.StartPC, locvar.EndPC)
	}
}

func (pr printer) printUpvalues(p *object.Proto) {
	fmt.Fprintf(pr.w, "upvalues (%d) for %p: \n", len(p.Upvalues), p)
	for i, upval := range p.Upvalues {
		fmt.Fprintf(pr.w, "\t%d\t%s\t%t\t%d\n", i, upval.Name, upval.Instack, upval.Index)
	}
}

func (pr printer) printProtos(p *object.Proto) {
	for _, f := range p.Protos {
		pr.printFunc(f)
	}
}

func upvalName(p *object.Proto, r int) (name string) {
	name = string(p.Upvalues[r].Name)
	if len(name) == 0 {
		name = "-"
	}
	return
}
