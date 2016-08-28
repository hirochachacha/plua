package undump

import (
	"encoding/binary"
	"io"

	"github.com/hirochachacha/blua"
	"github.com/hirochachacha/blua/errors"
	"github.com/hirochachacha/blua/internal/limits"
	"github.com/hirochachacha/blua/object"
	"github.com/hirochachacha/blua/opcode"
)

var (
	errUndumpOverflow         = errors.UndumpError.New("undump overflow")
	errShortHeader            = errors.UndumpError.New("header is too short")
	errSignatureMismatch      = errors.UndumpError.New("signature mismatch")
	errVersionMismatch        = errors.UndumpError.New("version mismatch")
	errFormatMismatch         = errors.UndumpError.New("format mismatch")
	errDataMismatch           = errors.UndumpError.New("data mismatch")
	errInvalidIntSize         = errors.UndumpError.New("int size is invalid")
	errInvalidIntegerSize     = errors.UndumpError.New("integer size is invalid")
	errInvalidNumberSize      = errors.UndumpError.New("number size is invalid")
	errInvalidInstructionSize = errors.UndumpError.New("instruction size is invalid")
	errEndiannessMismatch     = errors.UndumpError.New("endianness mismatch")
	errNumberFormatMismatch   = errors.UndumpError.New("number format mismatch")
	errMalformedByteCode      = errors.UndumpError.New("malformed byte code detected")
	errTruncatedChunk         = errors.UndumpError.New("truncated precompiled chunk")
)

const bufferSize = 20

type Mode uint // currently, no mode are defined

func Undump(r io.Reader, mode Mode) (*object.Proto, error) {
	u := &undumper{
		r:     r,
		order: binary.LittleEndian,
	}

	if br, ok := r.(io.ByteReader); ok {
		u.byte = func() (int, error) {
			c, err := br.ReadByte()
			return int(c), err
		}
	} else {
		u.byte = func() (int, error) {
			_, err := io.ReadFull(r, u.bs[:1])
			return int(u.bs[0]), err
		}
	}

	err := u.loadHeader()
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, errShortHeader
		}

		if _, ok := err.(*errors.Error); !ok {
			return nil, errors.UndumpError.Wrap(err)
		}

		return nil, err
	}

	n, err := u.loadByte()
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, errTruncatedChunk
		}

		if _, ok := err.(*errors.Error); !ok {
			return nil, errors.UndumpError.Wrap(err)
		}

		return nil, err
	}

	p, err := u.loadFunction("")
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, errTruncatedChunk
		}

		if _, ok := err.(*errors.Error); !ok {
			return nil, errors.UndumpError.Wrap(err)
		}

		return nil, err
	}

	if n != len(p.Upvalues) {
		return nil, errMalformedByteCode
	}

	return p, nil
}

type undumper struct {
	r     io.Reader
	order binary.ByteOrder

	bs [bufferSize]byte // buffer for short string

	byte    func() (int, error)
	int     func(*undumper) (int, error)
	sizeT   func(*undumper) (int, error)
	integer func(*undumper) (object.Integer, error)
	number  func(*undumper) (object.Number, error)
}

func (u *undumper) loadByte() (int, error) {
	return u.byte()
}

func (u *undumper) loadBool() (bool, error) {
	c, err := u.loadByte()

	return c != 0, err
}

func (u *undumper) loadInt() (int, error) {
	return u.int(u)
}

func (u *undumper) loadSizeT() (int, error) {
	return u.sizeT(u)
}

func (u *undumper) loadInst() (opcode.Instruction, error) {
	var i uint32
	err := binary.Read(u.r, u.order, &i)
	return opcode.Instruction(i), err
}

func (u *undumper) loadBoolean() (object.Boolean, error) {
	c, err := u.loadByte()

	return object.Boolean(c != 0), err
}

func (u *undumper) loadInteger() (object.Integer, error) {
	return u.integer(u)
}

func (u *undumper) loadNumber() (object.Number, error) {
	return u.number(u)
}

func (u *undumper) loadString() (string, error) {
	length, err := u.loadByte()
	if err != nil {
		return "", err
	}

	if length == 0xFF {
		length, err = u.loadSizeT()
	}

	if length == 0 {
		return "", nil
	}

	length--

	var bs []byte

	if length <= len(u.bs) {
		bs = u.bs[:length]
	} else {
		bs = make([]byte, length)
	}

	_, err = io.ReadFull(u.r, bs)

	return string(bs), err
}

func (u *undumper) loadCode() ([]opcode.Instruction, error) {
	n, err := u.loadInt()
	if err != nil {
		return nil, err
	}

	code := make([]opcode.Instruction, n)

	err = binary.Read(u.r, u.order, code)

	return code, err
}

func (u *undumper) loadConstants() ([]object.Value, error) {
	n, err := u.loadInt()
	if err != nil {
		return nil, err
	}

	constants := make([]object.Value, n)

	var v object.Value
	var t int
	for i := 0; i < n; i++ {
		t, err = u.loadByte()
		if err != nil {
			return nil, err
		}

		switch object.Type(t) {
		case object.TNIL:
			v = nil
		case object.TBOOLEAN:
			v, err = u.loadBoolean()
		case object.TNUMFLT:
			v, err = u.loadNumber()
		case object.TNUMINT:
			v, err = u.loadInteger()
		case object.TSHRSTR, object.TLNGSTR:
			s, _err := u.loadString()
			v = object.String(s)
			err = _err
		default:
			panic("unexpected")
		}

		if err != nil {
			return nil, err
		}

		constants[i] = v
	}

	return constants, nil
}

func (u *undumper) loadUpvalues() ([]object.UpvalueDesc, error) {
	n, err := u.loadInt()
	if err != nil {
		return nil, err
	}

	upvalues := make([]object.UpvalueDesc, n)

	var instack bool
	var idx int
	for i := 0; i < n; i++ {
		instack, err = u.loadBool()
		if err != nil {
			return nil, err
		}

		idx, err = u.loadByte()
		if err != nil {
			return nil, err
		}

		upvalues[i].Instack = instack
		upvalues[i].Index = idx
	}

	return upvalues, nil
}

func (u *undumper) loadProtos(source string) ([]*object.Proto, error) {
	n, err := u.loadInt()
	if err != nil {
		return nil, err
	}

	protos := make([]*object.Proto, n)

	for i := 0; i < n; i++ {
		protos[i], err = u.loadFunction(source)

		if err != nil {
			return nil, err
		}
	}

	return protos, nil
}

func (u *undumper) loadDebug() (lineInfo []int, locVars []object.LocVar, upvalueNames []string, err error) {
	var n int

	n, err = u.loadInt()
	if err != nil {
		return
	}

	lineInfo = make([]int, n)

	for i := 0; i < n; i++ {
		lineInfo[i], err = u.loadInt()

		if err != nil {
			return
		}
	}

	n, err = u.loadInt()
	if err != nil {
		return
	}

	locVars = make([]object.LocVar, n)

	var name string
	var startPC int
	var endPC int
	for i := 0; i < n; i++ {
		name, err = u.loadString()
		if err != nil {
			return
		}

		startPC, err = u.loadInt()
		if err != nil {
			return
		}

		endPC, err = u.loadInt()
		if err != nil {
			return
		}

		locVars[i].Name = name
		locVars[i].StartPC = startPC
		locVars[i].EndPC = endPC
	}

	n, err = u.loadInt()
	if err != nil {
		return
	}

	upvalueNames = make([]string, n)
	for i := 0; i < n; i++ {
		upvalueNames[i], err = u.loadString()

		if err != nil {
			return
		}
	}

	return
}

func (u *undumper) loadFunction(psource string) (*object.Proto, error) {
	source, err := u.loadString()
	if len(source) == 0 {
		source = psource
	}
	lineDefined, err := u.loadInt()
	if err != nil {
		return nil, err
	}
	lastLineDefined, err := u.loadInt()
	if err != nil {
		return nil, err
	}
	nParams, err := u.loadByte()
	if err != nil {
		return nil, err
	}
	isVararg, err := u.loadBool()
	if err != nil {
		return nil, err
	}
	maxStackSize, err := u.loadByte()
	if err != nil {
		return nil, err
	}
	code, err := u.loadCode()
	if err != nil {
		return nil, err
	}
	constants, err := u.loadConstants()
	if err != nil {
		return nil, err
	}
	upvalues, err := u.loadUpvalues()
	if err != nil {
		return nil, err
	}
	protos, err := u.loadProtos(source)
	if err != nil {
		return nil, err
	}
	lineInfo, locVars, upvalueNames, err := u.loadDebug()
	if err != nil {
		return nil, err
	}
	for i, name := range upvalueNames {
		upvalues[i].Name = name
	}

	p := &object.Proto{
		Code:            code,
		Constants:       constants,
		Protos:          protos,
		Upvalues:        upvalues,
		Source:          source,
		LineDefined:     lineDefined,
		LastLineDefined: lastLineDefined,
		NParams:         nParams,
		IsVararg:        isVararg,
		MaxStackSize:    maxStackSize,
		LineInfo:        lineInfo,
		LocVars:         locVars,
	}

	return p, nil
}

func (u *undumper) loadHeader() error {
	header := make([]byte, 12)

	_, err := io.ReadFull(u.r, header)

	if err != nil {
		if err == io.ErrShortBuffer {
			return errShortHeader
		}
		return err
	}

	if string(header[:4]) != blua.LUA_SIGNATURE {
		return errSignatureMismatch
	}

	if header[4] != blua.LUAC_VERSION {
		return errVersionMismatch
	}

	if header[5] != blua.LUAC_FORMAT {
		return errFormatMismatch
	}

	if string(header[6:]) != blua.LUAC_DATA {
		return errDataMismatch
	}

	intSize, err := u.loadByte()
	if err != nil {
		return err
	}
	sizeTSize, err := u.loadByte()
	if err != nil {
		return err
	}
	instSize, err := u.loadByte()
	if err != nil {
		return err
	}
	if instSize != opcode.InstructionSize {
		return errInvalidInstructionSize
	}
	integerSize, err := u.loadByte()
	if err != nil {
		return err
	}
	numberSize, err := u.loadByte()
	if err != nil {
		return err
	}

	u.int, err = makeInt(intSize)
	if err != nil {
		return err
	}
	u.sizeT, err = makeInt(sizeTSize)
	if err != nil {
		return err
	}
	u.integer, err = makeInteger(integerSize)
	if err != nil {
		return err
	}
	u.number, err = makeNumber(numberSize)
	if err != nil {
		return err
	}

	i, err := u.loadInteger()
	if err != nil {
		return err
	}

	if i != blua.LUAC_INT {

		// guess endian
		switch u.order {
		case binary.LittleEndian:
			if isReverseEndian(int64(i), integerSize) {
				u.order = binary.BigEndian
			} else {
				return errNumberFormatMismatch
			}
		case binary.BigEndian:
			if isReverseEndian(int64(i), integerSize) {
				u.order = binary.LittleEndian
			} else {
				return errNumberFormatMismatch
			}
		default:
			return errEndiannessMismatch
		}
	}

	f, err := u.loadNumber()
	if err != nil {
		return err
	}

	if f != blua.LUAC_NUM {
		return errNumberFormatMismatch
	}

	return nil
}

func makeInt(size int) (f func(*undumper) (int, error), err error) {
	switch size {
	case 1:
		f = func(u *undumper) (int, error) {
			var i int8
			err := binary.Read(u.r, u.order, &i)
			return int(i), err
		}
	case 2:
		f = func(u *undumper) (int, error) {
			var i int16
			err := binary.Read(u.r, u.order, &i)
			return int(i), err
		}
	case 4:
		f = func(u *undumper) (int, error) {
			var i int32
			err := binary.Read(u.r, u.order, &i)
			return int(i), err
		}
	case 8:
		f = func(u *undumper) (int, error) {
			var i int64
			err := binary.Read(u.r, u.order, &i)
			if err != nil {
				return 0, err
			}

			if limits.IntSize == 4 && (i > limits.MaxInt || i < limits.MinInt) {
				return 0, errUndumpOverflow
			}

			return int(i), nil
		}
	default:
		err = errInvalidIntSize
	}

	return
}

func makeInteger(size int) (f func(*undumper) (object.Integer, error), err error) {
	switch size {
	case 1:
		f = func(u *undumper) (object.Integer, error) {
			var i int8
			err := binary.Read(u.r, u.order, &i)
			return object.Integer(i), err
		}
	case 2:
		f = func(u *undumper) (object.Integer, error) {
			var i int16
			err := binary.Read(u.r, u.order, &i)
			return object.Integer(i), err
		}
	case 4:
		f = func(u *undumper) (object.Integer, error) {
			var i int32
			err := binary.Read(u.r, u.order, &i)
			return object.Integer(i), err
		}
	case 8:
		f = func(u *undumper) (object.Integer, error) {
			var i int64
			err := binary.Read(u.r, u.order, &i)
			if err != nil {
				return 0, err
			}

			return object.Integer(i), nil
		}
	default:
		err = errInvalidIntegerSize
	}

	return
}

func makeNumber(size int) (f func(*undumper) (object.Number, error), err error) {
	switch size {
	case 4:
		f = func(u *undumper) (object.Number, error) {
			var f float32
			err := binary.Read(u.r, u.order, &f)
			return object.Number(f), err
		}
	case 8:
		f = func(u *undumper) (object.Number, error) {
			var f float64
			err := binary.Read(u.r, u.order, &f)
			return object.Number(f), err
		}
	default:
		err = errInvalidNumberSize
	}

	return
}

func isReverseEndian(i int64, size int) bool {
	var r int64

	switch size {
	case 1:
		r = int64(int8(i))
	case 2:
		r = i & 0xFF
		r = (r << 8) | ((i >> 8) & 0xFF)
	case 4:
		r = i & 0xFF

		r = (r << 8) | ((i >> 8) & 0xFF)
		r = (r << 8) | ((i >> 16) & 0xFF)
		r = (r << 8) | ((i >> 24) & 0xFF)
	case 8:
		r = i & 0xFF

		r = (r << 8) | ((i >> 8) & 0xFF)
		r = (r << 8) | ((i >> 16) & 0xFF)
		r = (r << 8) | ((i >> 24) & 0xFF)
		r = (r << 8) | ((i >> 32) & 0xFF)
		r = (r << 8) | ((i >> 40) & 0xFF)
		r = (r << 8) | ((i >> 48) & 0xFF)
		r = (r << 8) | ((i >> 56) & 0xFF)
	default:
		panic("unreachable")
	}

	return i == r
}
