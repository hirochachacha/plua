package string

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

type byteOrder int

const (
	littleEndian byteOrder = iota
	bigEndian
)

var nativeEndian = littleEndian

type packer struct {
	bytes.Buffer

	ap   *fnutil.ArgParser
	opts []*kOption
}

func newPacker(ap *fnutil.ArgParser, opts []*kOption) *packer {
	return &packer{ap: ap, opts: opts}
}

func (p *packer) pack() (string, *object.RuntimeError) {
	n := 1

	for _, opt := range p.opts {
		var err *object.RuntimeError

		switch opt.typ {
		case kInt:
			err = p.packInt(n, opt)
		case kUint:
			err = p.packUint(n, opt)
		case kFloat:
			err = p.packFloat(n, opt)
		case kChar:
			err = p.packChar(n, opt)
		case kString:
			err = p.packString(n, opt)
		case kZeroString:
			err = p.packZeroString(n)
		case kPadding:
			p.packPadding()
		case kPaddingAlign, kNop:
			n--
		default:
			panic("unreachable")
		}
		if err != nil {
			return "", err
		}

		for j := 0; j < opt.padding; j++ {
			p.packPadding()
		}

		n++
	}

	return p.String(), nil
}

func (p *packer) packRaw(u64 uint64, opt *kOption) {
	switch opt.endian {
	case littleEndian:
		for j := 0; j < opt.size; j++ {
			p.WriteByte(byte(u64 >> uint(8*j)))
		}
	case bigEndian:
		for j := 0; j < opt.size; j++ {
			p.WriteByte(byte(u64 >> uint(8*(opt.size-1-j))))
		}
	default:
		panic("unreachable")
	}
}

func (p *packer) packUint(n int, opt *kOption) *object.RuntimeError {
	i64, err := p.ap.ToGoInt64(n)
	if err != nil {
		return err
	}

	u64 := uint64(i64)

	if opt.size < 8 {
		lim := uint64(1 << uint((opt.size * 8)))

		if u64 >= lim {
			return p.ap.ArgError(n, "unsigned overflow")
		}
	}

	p.packRaw(u64, opt)

	return nil
}

func (p *packer) packInt(n int, opt *kOption) *object.RuntimeError {
	i64, err := p.ap.ToGoInt64(n)
	if err != nil {
		return err
	}

	if opt.size < 8 {
		lim := int64(1 << uint((opt.size*8)-1))

		if !(-lim <= i64 && i64 < lim) {
			return p.ap.ArgError(n, "integer overflow")
		}
	}

	var u64 uint64
	if i64 < 0 {
		u64 = uint64((1 << uint(opt.size*8)) + i64)
	} else {
		u64 = uint64(i64)
	}

	p.packRaw(u64, opt)

	return nil
}

func (p *packer) packFloat(n int, opt *kOption) *object.RuntimeError {
	f, err := p.ap.ToGoFloat64(n)
	if err != nil {
		return err
	}

	switch opt.size {
	case 4:
		p.packRaw(uint64(math.Float32bits(float32(f))), opt)
	case 8:
		p.packRaw(math.Float64bits(f), opt)
	default:
		panic("unreachable")
	}

	return nil
}

func (p *packer) packChar(n int, opt *kOption) *object.RuntimeError {
	s, err := p.ap.ToGoString(n)
	if err != nil {
		return err
	}

	if opt.size != len(s) {
		return p.ap.ArgError(n, "wrong length")
	}

	p.WriteString(s)

	return nil
}

func (p *packer) packString(n int, opt *kOption) *object.RuntimeError {
	s, err := p.ap.ToGoString(n)
	if err != nil {
		return err
	}

	if opt.size < 8 {
		lim := uint64(1 << uint((opt.size*8)-1))

		if uint64(len(s)) >= lim {
			return p.ap.ArgError(n, "string length does not fit in given size")
		}
	}

	p.packRaw(uint64(len(s)), opt)

	p.WriteString(s)

	return nil
}

func (p *packer) packZeroString(n int) *object.RuntimeError {
	s, err := p.ap.ToGoString(n)
	if err != nil {
		return err
	}

	if strings.ContainsRune(s, 0x00) {
		return p.ap.ArgError(n, "string contains zero")
	}

	p.WriteString(s)
	p.WriteByte(0x00)

	return nil
}

func (p *packer) packPadding() {
	p.WriteByte(0x00)
}

type unpacker struct {
	ap   *fnutil.ArgParser
	fmt  string
	opts []*kOption

	off int
}

func newUnpacker(ap *fnutil.ArgParser, fmt string, opts []*kOption) *unpacker {
	return &unpacker{fmt: fmt, ap: ap, opts: opts}
}

func (u *unpacker) unpack() (rets []object.Value, err *object.RuntimeError) {
	for _, opt := range u.opts {
		switch opt.typ {
		case kInt:
			ret, err := u.unpackInt(opt)
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kUint:
			ret, err := u.unpackUint(opt)
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kFloat:
			ret, err := u.unpackFloat(opt)
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kChar:
			ret, err := u.unpackChar(opt)
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kString:
			ret, err := u.unpackString(opt)
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kZeroString:
			ret, err := u.unpackZeroString()
			if err != nil {
				return nil, err
			}
			rets = append(rets, ret)
		case kPadding:
			err := u.unpackPadding()
			if err != nil {
				return nil, err
			}
		case kPaddingAlign, kNop:
		default:
			panic("unreachable")
		}

		for j := 0; j < opt.padding; j++ {
			err := u.unpackPadding()
			if err != nil {
				return nil, err
			}
		}
	}

	rets = append(rets, object.Integer(u.off))

	return rets, nil
}

func (u *unpacker) unpackRaw(opt *kOption) (uint64, *object.RuntimeError) {
	if len(u.fmt)-u.off < opt.size {
		return 0, u.ap.ArgError(0, "data string is too short")
	}

	var u64 uint64

	switch opt.endian {
	case littleEndian:
		for i := 0; i < opt.size; i++ {
			u64 |= uint64(u.fmt[u.off+i]) << uint(8*i)
		}
	case bigEndian:
		for i := 0; i < opt.size; i++ {
			u64 |= uint64(u.fmt[u.off+i]) << uint(8*(opt.size-1-i))
		}
	default:
		panic("unreachable")
	}

	return u64, nil
}

func (u *unpacker) unpackUint(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackRaw(opt)
	if err != nil {
		return nil, err
	}

	u.off += opt.size

	return object.Integer(u64), nil
}

func (u *unpacker) unpackInt(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackRaw(opt)
	if err != nil {
		return nil, err
	}

	var i64 int64
	if (u64 & (1 << (uint(opt.size*8) - 1))) != 0 {
		i64 = -int64((1 << uint(opt.size*8)) - u64)
	} else {
		i64 = int64(u64)
	}

	u.off += opt.size

	return object.Integer(i64), nil
}

func (u *unpacker) unpackFloat(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackRaw(opt)
	if err != nil {
		return nil, err
	}

	u.off += opt.size

	switch opt.size {
	case 4:
		return object.Number(math.Float32frombits(uint32(u64))), nil
	case 8:
		return object.Number(math.Float64frombits(u64)), nil
	default:
		panic("unreachable")
	}
}

func (u *unpacker) unpackChar(opt *kOption) (object.Value, *object.RuntimeError) {
	if len(u.fmt)-u.off < opt.size {
		return nil, u.ap.ArgError(0, "data string is too short")
	}

	val := object.String(u.fmt[u.off : u.off+opt.size])

	u.off += opt.size

	return val, nil
}

func (u *unpacker) unpackString(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackRaw(opt)
	if err != nil {
		return nil, err
	}

	if u64 > uint64(limits.MaxInt) {
		return nil, u.ap.ArgError(0, "integer overflow")
	}

	slen := int(u64)

	if len(u.fmt)-u.off < slen {
		return nil, u.ap.ArgError(0, "data string is too short")
	}

	val := object.String(u.fmt[u.off : u.off+slen])

	u.off += slen

	return val, nil
}

func (u *unpacker) unpackZeroString() (object.Value, *object.RuntimeError) {
	slen := strings.IndexByte(u.fmt[u.off:], 0x00)
	if slen == -1 {
		return nil, u.ap.ArgError(0, "data string does not contains zero")
	}

	val := object.String(u.fmt[u.off : u.off+slen])

	u.off += slen + 1

	return val, nil
}

func (u *unpacker) unpackPadding() *object.RuntimeError {
	if len(u.fmt) <= u.off {
		return u.ap.ArgError(0, "data string is too short")
	}

	if u.fmt[u.off] != 0x00 {
		return u.ap.ArgError(0, fmt.Sprintf("expected 0x00, but 0x%x", u.fmt[u.off]))
	}

	u.off++

	return nil
}
