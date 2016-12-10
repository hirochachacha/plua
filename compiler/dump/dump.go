package dump

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

var (
	errIntegerOverflow    = &Error{errors.New("integer overflow")}
	errFloatOverflow      = &Error{errors.New("float overflow")}
	errFloatUnderflow     = &Error{errors.New("float underflow")}
	errByteOverflow       = &Error{errors.New("byte overflow")}
	errInvalidIntSize     = &Error{errors.New("IntSize should be power of 2")}
	errInvalidSizeTSize   = &Error{errors.New("SizeTSize should be power of 2")}
	errInvalidIntegerSize = &Error{errors.New("IntegerSize should be power of 2")}
	errInvalidNumberSize  = &Error{errors.New("NumberSize should be 4 or 8")}
	errInvalidByteOrder   = &Error{errors.New("ByteOrder should not be nil")}
)

type Mode uint

const (
	StripDebugInfo Mode = 1 << iota
)

func DumpTo(w io.Writer, p *object.Proto, mode Mode) (err error) {
	return defaultConfig.DumpTo(w, p, mode)
}

var defaultConfig = &Config{
	IntSize:     8,
	SizeTSize:   8,
	IntegerSize: 8,
	NumberSize:  8,
	ByteOrder:   binary.LittleEndian,
}

type Config struct {
	IntSize     int              // size of int
	SizeTSize   int              // size of size_t
	IntegerSize int              // size of lua integer
	NumberSize  int              // size of lua number
	ByteOrder   binary.ByteOrder // byte order
}

func (cfg *Config) validate() error {
	if !isPowerOfTwo(cfg.IntSize) {
		return errInvalidIntSize
	}

	if !isPowerOfTwo(cfg.SizeTSize) {
		return errInvalidSizeTSize
	}

	if !isPowerOfTwo(cfg.IntegerSize) {
		return errInvalidIntegerSize
	}

	if !(cfg.NumberSize == 4 || cfg.NumberSize == 8) {
		return errInvalidNumberSize
	}

	if cfg.ByteOrder == nil {
		return errInvalidByteOrder
	}

	return nil
}

func isPowerOfTwo(x int) bool {
	return x&(x-1) == 0
	// for x%2 == 0 && x > 1 {
	// x /= 2
	// }
	// return x == 1
}

func (cfg *Config) DumpTo(w io.Writer, p *object.Proto, mode Mode) (err error) {
	err = cfg.validate()
	if err != nil {
		return err
	}

	d := &dumper{
		w:       w,
		mode:    mode,
		cfg:     cfg,
		int:     makeInt(cfg.IntSize),
		sizeT:   makeInt(cfg.SizeTSize),
		integer: makeInteger(cfg.IntegerSize),
		number:  makeNumber(cfg.NumberSize),
	}

	if bw, ok := w.(io.ByteWriter); ok {
		d.byte = func(b byte) error {
			return bw.WriteByte(b)
		}
	} else {
		d.byte = func(b byte) error {
			_, err := w.Write([]byte{b})

			return err
		}
	}

	if sw, ok := w.(stringWriter); ok {
		d.str = func(s string) error {
			_, err := sw.WriteString(s)
			return err
		}
	} else {
		d.str = func(s string) error {
			_, err := w.Write([]byte(s))
			return err
		}
	}

	err = d.dumpHeader()
	if err != nil {
		return err
	}

	err = d.dumpByte(len(p.Upvalues))
	if err != nil {
		return err
	}

	err = d.dumpFunction(p, "")
	if err != nil {
		return err
	}

	return
}

type stringWriter interface {
	WriteString(string) (int, error)
}

type dumper struct {
	w       io.Writer
	mode    Mode
	cfg     *Config
	byte    func(byte) error
	str     func(string) error
	int     func(*dumper, int) error
	sizeT   func(*dumper, int) error
	integer func(*dumper, object.Integer) error
	number  func(*dumper, object.Number) error
}

func (d *dumper) dumpByte(x int) error {
	if x < 0 || x > 255 {
		return errByteOverflow
	}
	return d.byte(byte(x))
}

func (d *dumper) dumpBool(b bool) error {
	if b {
		return d.dumpByte(1)
	}

	return d.dumpByte(0)
}

func (d *dumper) dumpInt(x int) error {
	return d.int(d, x)
}

func (d *dumper) dumpSizeT(x int) error {
	return d.sizeT(d, x)
}

func (d *dumper) dumpStr(s string) error {
	return d.str(s)
}

func (d *dumper) dumpNumber(x object.Number) error {
	return d.number(d, x)
}

func (d *dumper) dumpInteger(x object.Integer) error {
	return d.integer(d, x)
}

func (d *dumper) dumpString(s string) error {
	length := len(s)
	if length == 0 {
		return d.dumpByte(0)
	}

	length++ // trailing '\0'

	var err error
	if length < 0xFF {
		err = d.dumpByte(length)
	} else {
		err = d.dumpByte(0xFF)
		if err != nil {
			return err
		}
		err = d.dumpSizeT(length)
	}

	if err != nil {
		return err
	}

	return d.dumpStr(string(s))
}

func (d *dumper) dumpCode(p *object.Proto) error {
	err := d.dumpInt(len(p.Code))
	if err != nil {
		return err
	}

	return binary.Write(d.w, d.cfg.ByteOrder, p.Code)
}

func (d *dumper) dumpConstants(p *object.Proto) error {
	err := d.dumpInt(len(p.Constants))
	if err != nil {
		return err
	}

	for _, c := range p.Constants {
		switch c := c.(type) {
		case nil:
			// do nothing
			err = d.dumpByte(int(object.TNIL))
		case object.Boolean:
			err = d.dumpByte(int(object.TBOOLEAN))
			if err == nil {
				err = d.dumpBool(bool(c))
			}
		case object.Integer:
			err = d.dumpByte(int(object.TNUMINT))
			if err == nil {
				err = d.dumpInteger(c)
			}
		case object.Number:
			err = d.dumpByte(int(object.TNUMFLT))
			if err == nil {
				err = d.dumpNumber(c)
			}
		case object.String:
			err = d.dumpByte(int(object.TSTRING))
			if err == nil {
				err = d.dumpString(string(c))
			}
		default:
			panic("unreachable")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *dumper) dumpUpvalues(p *object.Proto) error {
	err := d.dumpInt(len(p.Upvalues))
	if err != nil {
		return err
	}

	for _, u := range p.Upvalues {
		d.dumpBool(u.Instack)
		d.dumpByte(u.Index)
	}

	return nil
}

func (d *dumper) dumpProtos(p *object.Proto) error {
	err := d.dumpInt(len(p.Protos))
	if err != nil {
		return err
	}

	for _, p := range p.Protos {
		err = d.dumpFunction(p, p.Source)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *dumper) dumpDebug(p *object.Proto) (err error) {
	if d.mode&StripDebugInfo != 0 {
		err = d.dumpInt(0) // len(p.LineInfo)
		if err != nil {
			return err
		}
		err = d.dumpInt(0) // len(p.LocVars)
		if err != nil {
			return err
		}
		err = d.dumpInt(0) // len(p.Upvalues)

		return
	}

	err = d.dumpInt(len(p.LineInfo))
	if err != nil {
		return err
	}
	for _, line := range p.LineInfo {
		err = d.dumpInt(line)
		if err != nil {
			return err
		}
	}

	err = d.dumpInt(len(p.LocVars))
	if err != nil {
		return err
	}
	for _, l := range p.LocVars {
		err = d.dumpString(l.Name)
		if err != nil {
			return err
		}
		err = d.dumpInt(l.StartPC)
		if err != nil {
			return err
		}
		err = d.dumpInt(l.EndPC)
		if err != nil {
			return err
		}
	}

	err = d.dumpInt(len(p.Upvalues))
	if err != nil {
		return err
	}
	for _, u := range p.Upvalues {
		err = d.dumpString(u.Name)
		if err != nil {
			return err
		}
	}

	return
}

func (d *dumper) dumpFunction(p *object.Proto, psource string) (err error) {
	if d.mode&StripDebugInfo != 0 || p.Source == psource {
		err = d.dumpString("")
	} else {
		err = d.dumpString(p.Source)
	}
	if err != nil {
		return err
	}

	err = d.dumpInt(p.LineDefined)
	if err != nil {
		return err
	}
	err = d.dumpInt(p.LastLineDefined)
	if err != nil {
		return err
	}
	err = d.dumpByte(p.NParams)
	if err != nil {
		return err
	}
	err = d.dumpBool(p.IsVararg)
	if err != nil {
		return err
	}
	err = d.dumpByte(p.MaxStackSize)
	if err != nil {
		return err
	}
	err = d.dumpCode(p)
	if err != nil {
		return err
	}
	err = d.dumpConstants(p)
	if err != nil {
		return err
	}
	err = d.dumpUpvalues(p)
	if err != nil {
		return err
	}
	err = d.dumpProtos(p)
	if err != nil {
		return err
	}
	err = d.dumpDebug(p)

	return
}

func (d *dumper) dumpHeader() (err error) {
	err = d.dumpStr(version.LUA_SIGNATURE)
	if err != nil {
		return err
	}
	err = d.dumpByte(version.LUAC_VERSION)
	if err != nil {
		return err
	}
	err = d.dumpByte(version.LUAC_FORMAT)
	if err != nil {
		return err
	}
	err = d.dumpStr(version.LUAC_DATA)
	if err != nil {
		return err
	}
	err = d.dumpByte(d.cfg.IntSize)
	if err != nil {
		return err
	}
	err = d.dumpByte(d.cfg.SizeTSize)
	if err != nil {
		return err
	}
	err = d.dumpByte(opcode.InstructionSize)
	if err != nil {
		return err
	}
	err = d.dumpByte(d.cfg.IntegerSize)
	if err != nil {
		return err
	}
	err = d.dumpByte(d.cfg.NumberSize)
	if err != nil {
		return err
	}
	err = d.dumpInteger(version.LUAC_INT)
	if err != nil {
		return err
	}
	err = d.dumpNumber(version.LUAC_NUM)

	return
}

func makeInt(size int) (f func(*dumper, int) error) {
	switch size {
	case 1:
		f = func(d *dumper, x int) error {
			if x < limits.MinInt8 || x > limits.MaxInt8 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int8(x))
		}
	case 2:
		f = func(d *dumper, x int) error {
			if x < limits.MinInt16 || x > limits.MaxInt16 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int16(x))
		}
	case 4:
		f = func(d *dumper, x int) error {
			if x < limits.MinInt32 || x > limits.MaxInt32 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int32(x))
		}
	case 8:
		f = func(d *dumper, x int) error {
			return binary.Write(d.w, d.cfg.ByteOrder, int64(x))
		}
	default:
		panic("unreachable")
	}

	return
}

func makeInteger(size int) (f func(*dumper, object.Integer) error) {
	switch size {
	case 1:
		f = func(d *dumper, x object.Integer) error {
			if x < limits.MinInt8 || x > limits.MaxInt8 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int8(x))
		}
	case 2:
		f = func(d *dumper, x object.Integer) error {
			if x < limits.MinInt16 || x > limits.MaxInt16 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int16(x))
		}
	case 4:
		f = func(d *dumper, x object.Integer) error {
			if x < limits.MinInt32 || x > limits.MaxInt32 {
				return errIntegerOverflow
			}
			return binary.Write(d.w, d.cfg.ByteOrder, int32(x))
		}
	case 8:
		f = func(d *dumper, x object.Integer) error {
			return binary.Write(d.w, d.cfg.ByteOrder, int64(x))
		}
	default:
		panic("unreachable")
	}

	return
}

func makeNumber(size int) (f func(*dumper, object.Number) error) {
	switch size {
	case 4:
		f = func(d *dumper, x object.Number) error {
			abs := x
			if x < 0 {
				abs = -x
			}

			if abs > limits.MaxFloat32 {
				return errFloatOverflow
			}

			if abs < limits.SmallestNonzeroFloat32 {
				return errFloatUnderflow
			}

			return binary.Write(d.w, d.cfg.ByteOrder, float32(x))
		}
	case 8:
		f = func(d *dumper, x object.Number) error {
			return binary.Write(d.w, d.cfg.ByteOrder, float64(x))
		}
	default:
		panic("unreachable")
	}

	return
}
