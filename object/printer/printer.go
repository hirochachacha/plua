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

func Print(x interface{}) error {
	return Fprint(os.Stdout, x)
}

func Fprint(w io.Writer, x interface{}) error {
	pr := &printer{w: w}
	pr.printMain(x)
	return pr.err
}

type printer struct {
	w   io.Writer
	err error
}

func (pr *printer) printMain(x interface{}) {
	switch x := x.(type) {
	case *object.Proto:
		pr.printFunc(x)
	case *object.DebugInfo:
		pr.printDebug(x)
	case object.Value:
		pr.printf("%s value = %v", object.ToType(x), x)
		pr.printValue(x)
		pr.println()
	case object.Process:
		// TODO
	}
}

func (pr *printer) Write(p []byte) (n int, err error) {
	if pr.err == nil {
		n, pr.err = pr.w.Write(p)
	}
	return n, pr.err
}

func (pr *printer) printf(s string, args ...interface{}) {
	fmt.Fprintf(pr, s, args...)
}

func (pr *printer) print(args ...interface{}) {
	fmt.Fprint(pr, args...)
}

func (pr *printer) println(args ...interface{}) {
	fmt.Fprintln(pr, args...)
}

func (pr *printer) printDebug(d *object.DebugInfo) {
	pr.printf("debug info %p = {\n", d)
	pr.printf("\tName            =\t%s\n", d.Name)
	pr.printf("\tNameWhat        =\t%s\n", d.NameWhat)
	pr.printf("\tWhat            =\t%s\n", d.What)
	pr.printf("\tSource          =\t%s\n", d.Source)
	pr.printf("\tCurrentLine     =\t%d\n", d.CurrentLine)
	pr.printf("\tLineDefined     =\t%d\n", d.LineDefined)
	pr.printf("\tLastLineDefined =\t%d\n", d.LastLineDefined)
	pr.printf("\tNUpvalues       =\t%d\n", d.NUpvalues)
	pr.printf("\tNParams         =\t%d\n", d.NParams)
	pr.printf("\tIsVararg        =\t%t\n", d.IsVararg)
	pr.printf("\tIsTailCall      =\t%t\n", d.IsTailCall)
	pr.printf("\tShortSource     =\t%s\n", d.ShortSource)
	pr.printf("\tFunc            =\t%v\n", d.Func)
	pr.printf("}\n")
}

func (pr *printer) printFunc(p *object.Proto) {
	pr.printHeader(p)
	pr.printCode(p)
	pr.printConstants(p)
	pr.printLocals(p)
	pr.printUpvalues(p)
	pr.printProtos(p)
}

func (pr *printer) printHeader(p *object.Proto) {
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

	pr.printf(
		"\n%s <%s:%d,%d> (%d instructions at %p)\n",
		typ, s, p.LineDefined, p.LastLineDefined, len(p.Code), p,
	)

	var vararg string
	if p.IsVararg {
		vararg = "+"
	}

	pr.printf(
		"%d%s params, %d slots, %d upvalues, ",
		p.NParams, vararg, p.MaxStackSize, len(p.Upvalues),
	)

	pr.printf(
		"%d locals, %d constants, %d functions\n",
		len(p.LocVars), len(p.Constants), len(p.Protos),
	)
}

func (pr *printer) printValue(val object.Value) {
	if val, ok := val.(object.String); ok {
		pr.print(strconv.Quote(string(val)))

		return
	}

	pr.print(val)
}

func (pr *printer) printCode(p *object.Proto) {
	for pc, code := range p.Code {
		a := code.A()
		b := code.B()
		c := code.C()
		bx := code.Bx()
		ax := code.Ax()
		sbx := code.SBx()

		pr.printf("\t%d\t", pc+1)

		if p.LineInfo != nil {
			pr.printf("[%d]\t", p.LineInfo[pc])
		} else {
			pr.printf("[-]\t")
		}

		pr.printf("%-9s\t", code.OpName())

		switch code.OpMode() {
		case opcode.IABC:
			pr.printf("%d", a)
			if code.BMode() != opcode.OpArgN {
				if b&opcode.BitRK != 0 {
					pr.printf(" %d", -1-(b & ^opcode.BitRK))
				} else {
					pr.printf(" %d", b)
				}
			}
			if code.CMode() != opcode.OpArgN {
				if c&opcode.BitRK != 0 {
					pr.printf(" %d", -1-(c & ^opcode.BitRK))
				} else {
					pr.printf(" %d", c)
				}
			}
		case opcode.IABx:
			pr.printf("%d", a)
			switch code.BMode() {
			case opcode.OpArgK:
				pr.printf(" %d", -1-bx)
			case opcode.OpArgU:
				pr.printf(" %d", bx)
			}
		case opcode.IAsBx:
			pr.printf("%d %d", a, sbx)
		case opcode.IAx:
			pr.printf("%d", -1-ax)
		default:
			panic("unreachable")
		}

		switch code.OpCode() {
		case opcode.LOADK:
			pr.print("\t; ")
			pr.printValue(p.Constants[bx])
		case opcode.GETUPVAL, opcode.SETUPVAL:
			pr.print("\t; ")
			pr.print(upvalName(p, b))
		case opcode.GETTABUP:
			pr.print("\t; ")
			pr.print(upvalName(p, b))
			if c&opcode.BitRK != 0 {
				pr.print(" ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.SETTABUP:
			pr.print("\t; ")
			pr.print(upvalName(p, a))
			if b&opcode.BitRK != 0 {
				pr.print(" ")
				pr.printValue(p.Constants[b & ^opcode.BitRK])
			}
			if c&opcode.BitRK != 0 {
				pr.print(" ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.GETTABLE, opcode.SELF:
			if c&opcode.BitRK != 0 {
				pr.print("\t; ")
				pr.printValue(p.Constants[c & ^opcode.BitRK])
			}
		case opcode.SETTABLE, opcode.ADD, opcode.SUB, opcode.MUL,
			opcode.POW, opcode.DIV, opcode.IDIV, opcode.BAND,
			opcode.BOR, opcode.BXOR, opcode.SHL, opcode.SHR,
			opcode.EQ, opcode.LT, opcode.LE:
			if b&opcode.BitRK != 0 || c&opcode.BitRK != 0 {
				pr.print("\t; ")
				if b&opcode.BitRK != 0 {
					pr.printValue(p.Constants[b & ^opcode.BitRK])
				} else {
					pr.print("-")
				}

				pr.print(" ")

				if c&opcode.BitRK != 0 {
					pr.printValue(p.Constants[c & ^opcode.BitRK])
				} else {
					pr.print("-")
				}
			}
		case opcode.JMP, opcode.FORLOOP, opcode.FORPREP, opcode.TFORLOOP:
			pr.printf("\t; to %d", sbx+pc+2)
		case opcode.CLOSURE:
			pr.printf("\t; %p", p.Protos[bx])
		case opcode.SETLIST:
			if c == 0 {
				pc++
				pr.printf("\t; %d", p.Code[pc])
			} else {
				pr.printf("\t; %d", c)
			}
		}

		pr.print("\n")
	}
}

func (pr *printer) printConstants(p *object.Proto) {
	pr.printf("constants (%d) for %p: \n", len(p.Constants), p)
	for i, c := range p.Constants {
		pr.printf("\t%d\t", i+1)
		pr.printValue(c)
		pr.println()
	}
}

func (pr *printer) printLocals(p *object.Proto) {
	pr.printf("locals (%d) for %p: \n", len(p.LocVars), p)
	for i, locvar := range p.LocVars {
		pr.printf("\t%d\t%s\t%d\t%d\n", i, locvar.Name, locvar.StartPC, locvar.EndPC)
	}
}

func (pr *printer) printUpvalues(p *object.Proto) {
	pr.printf("upvalues (%d) for %p: \n", len(p.Upvalues), p)
	for i, upval := range p.Upvalues {
		pr.printf("\t%d\t%s\t%t\t%d\n", i, upval.Name, upval.Instack, upval.Index)
	}
}

func (pr *printer) printProtos(p *object.Proto) {
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
