package string

import (
	"bytes"
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

// pack(fmt, v1, v2, ...)
func pack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	p := &packer{
		optParser: optParser{
			ap:       ap,
			fmt:      fmt,
			maxAlign: 1,
			endian:   littleEndian,
		},
	}

	s, err := p.pack()
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(s)}, nil
}

// packsize(fmt)
func packsize(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	p := &packer{
		optParser: optParser{
			ap:       ap,
			fmt:      fmt,
			maxAlign: 1,
			endian:   littleEndian,
		},
	}

	for {
		opt, err := p.nextkOpt()
		if err != nil {
			return nil, err
		}

		if opt == nil {
			break
		}

		if p.off > int(limits.MaxInt)-(opt.padding+opt.size) {
			return nil, p.ap.ArgError(0, "format result too large")
		}

		p.off += opt.padding + opt.size

		switch opt.typ {
		case kString, kZeroString:
			return nil, ap.ArgError(0, "variable-length format")
		}
	}

	return []object.Value{object.Integer(p.off)}, nil
}

type packer struct {
	optParser
	bytes.Buffer
}

func (p *packer) pack() (string, *object.RuntimeError) {
	n := 1

	for {
		opt, err := p.nextkOpt()
		if err != nil {
			return "", err
		}

		if opt == nil {
			break
		}

		if p.off > int(limits.MaxInt)-(opt.padding+opt.size) {
			return "", p.ap.ArgError(0, "format result too large")
		}

		p.off += opt.padding + opt.size

		for j := 0; j < opt.padding; j++ {
			p.packPadding()
		}

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
			continue
		case kPaddingAlign, kNop:
			continue
		default:
			panic("unreachable")
		}
		if err != nil {
			return "", err
		}

		n++
	}

	return p.String(), nil
}

func (p *packer) packUint64(u64 uint64, opt *kOption) {
	switch p.endian {
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

func (p *packer) packInt64(i64 int64, opt *kOption) {
	switch p.endian {
	case littleEndian:
		for j := 0; j < opt.size; j++ {
			p.WriteByte(^byte(^i64 >> uint(8*j)))
		}
	case bigEndian:
		for j := 0; j < opt.size; j++ {
			p.WriteByte(^byte(^i64 >> uint(8*(opt.size-1-j))))
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

	p.packUint64(u64, opt)

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

	p.packInt64(i64, opt)

	return nil
}

func (p *packer) packFloat(n int, opt *kOption) *object.RuntimeError {
	f, err := p.ap.ToGoFloat64(n)
	if err != nil {
		return err
	}

	switch opt.size {
	case 4:
		p.packUint64(uint64(math.Float32bits(float32(f))), opt)
	case 8:
		p.packUint64(math.Float64bits(f), opt)
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

	if opt.size < len(s) {
		return p.ap.ArgError(n, "string longer than given size")
	}

	p.WriteString(s)

	for i := len(s); i < opt.size; i++ {
		p.WriteByte(0)
	}

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

	p.packUint64(uint64(len(s)), opt)

	p.WriteString(s)

	p.off += len(s)

	return nil
}

func (p *packer) packZeroString(n int) *object.RuntimeError {
	s, err := p.ap.ToGoString(n)
	if err != nil {
		return err
	}

	if strings.ContainsRune(s, 0x00) {
		return p.ap.ArgError(n, "string contains zeros")
	}

	p.WriteString(s)
	p.WriteByte(0x00)

	p.off += len(s) + 1

	return nil
}

func (p *packer) packPadding() {
	p.WriteByte(0x00)
}
