package string

import (
	"fmt"
	"unsafe"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

type kOption struct {
	typ     kType
	size    int
	padding int
}

type kType uint

const (
	kInt kType = iota
	kUint
	kFloat
	kChar       // fixed-length string
	kString     // length + string
	kZeroString // string + '\x00'
	kPadding
	kPaddingAlign
	kNop
)

const maxIntSize = 16
const nativeAlign = int(unsafe.Offsetof(dummy{}.i))

type dummy struct {
	b byte
	i int64
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

type optParser struct {
	ap       *fnutil.ArgParser
	fmt      string
	maxAlign int
	endian   byteOrder
	off      int
}

func (p *optParser) nextkOpt() (opt *kOption, err *object.RuntimeError) {
	var typ kType
	var size int

next:
	if len(p.fmt) == 0 {
		return nil, nil
	}

	switch op := p.fmt[0]; op {
	case 'b':
		typ = kInt
		size = 1
		p.fmt = p.fmt[1:]
	case 'B':
		typ = kUint
		size = 1
		p.fmt = p.fmt[1:]
	case 'h':
		typ = kInt
		size = 2
		p.fmt = p.fmt[1:]
	case 'H':
		typ = kUint
		size = 2
		p.fmt = p.fmt[1:]
	case 'l':
		typ = kInt
		size = 8
		p.fmt = p.fmt[1:]
	case 'L':
		typ = kUint
		size = 8
		p.fmt = p.fmt[1:]
	case 'j':
		typ = kInt
		size = 8
		p.fmt = p.fmt[1:]
	case 'J':
		typ = kUint
		size = 8
		p.fmt = p.fmt[1:]
	case 'T':
		typ = kUint
		size = 8
		p.fmt = p.fmt[1:]
	case 'f':
		typ = kFloat
		size = 4
		p.fmt = p.fmt[1:]
	case 'd':
		typ = kFloat
		size = 8
		p.fmt = p.fmt[1:]
	case 'n':
		typ = kFloat
		size = 8
		p.fmt = p.fmt[1:]
	case 'i':
		typ = kInt
		size = limits.IntSize

		i := 1

		if i < len(p.fmt) && isDigit(p.fmt[i]) {
			size = int(p.fmt[i] - '0')

			i++

			for i < len(p.fmt) && isDigit(p.fmt[i]) {
				if (int(limits.MaxInt)-int(p.fmt[i]-'0'))/10 < size {
					return nil, object.NewRuntimeError("size is too large")
				}

				size = size*10 + int(p.fmt[i]-'0')
				i++
			}

			if size == 0 || size > maxIntSize {
				return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
			}
		}

		p.fmt = p.fmt[i:]
	case 'I':
		typ = kUint
		size = limits.IntSize

		i := 1

		if i < len(p.fmt) && isDigit(p.fmt[i]) {
			size = int(p.fmt[i] - '0')

			i++

			for i < len(p.fmt) && isDigit(p.fmt[i]) {
				if (int(limits.MaxInt)-int(p.fmt[i]-'0'))/10 < size {
					return nil, object.NewRuntimeError("size is too large")
				}

				size = size*10 + int(p.fmt[i]-'0')
				i++
			}

			if size == 0 || size > maxIntSize {
				return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
			}
		}

		p.fmt = p.fmt[i:]
	case 's':
		typ = kString
		size = 8

		i := 1

		if i < len(p.fmt) && isDigit(p.fmt[i]) {
			size = int(p.fmt[i] - '0')

			i++

			for i < len(p.fmt) && isDigit(p.fmt[i]) {
				if (int(limits.MaxInt)-int(p.fmt[i]-'0'))/10 < size {
					return nil, object.NewRuntimeError("size is too large")
				}

				size = size*10 + int(p.fmt[i]-'0')
				i++
			}

			if size == 0 || size > maxIntSize {
				return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
			}
		}

		p.fmt = p.fmt[i:]
	case 'c':
		typ = kChar
		size = 0

		i := 1

		if i < len(p.fmt) && isDigit(p.fmt[i]) {
			size = int(p.fmt[i] - '0')

			i++

			for i < len(p.fmt) && isDigit(p.fmt[i]) {
				if (int(limits.MaxInt)-int(p.fmt[i]-'0'))/10 < size {
					return nil, object.NewRuntimeError("size is too large")
				}

				size = size*10 + int(p.fmt[i]-'0')
				i++
			}
		} else {
			return nil, object.NewRuntimeError("missing size for format option 'c'")
		}

		p.fmt = p.fmt[i:]
	case 'z':
		typ = kZeroString
		size = 0
		p.fmt = p.fmt[1:]
	case 'x':
		typ = kPadding
		size = 1
		p.fmt = p.fmt[1:]
	case 'X':
		typ = kPaddingAlign
		size = 0
		p.fmt = p.fmt[1:]

		opt, err := p.nextkOpt()
		if err != nil {
			return nil, err
		}

		if opt == nil || opt.typ == kPaddingAlign || opt.typ == kChar || opt.size == 0 {
			return nil, p.ap.ArgError(0, "invalid next option for option 'X'")
		}

		opt.typ = kPaddingAlign
		opt.size = 0

		return opt, nil
	case ' ':
		p.fmt = p.fmt[1:]
		goto next
	case '<':
		p.endian = littleEndian
		p.fmt = p.fmt[1:]
		goto next
	case '>':
		p.endian = bigEndian
		p.fmt = p.fmt[1:]
		goto next
	case '=':
		p.endian = littleEndian
		p.fmt = p.fmt[1:]
		goto next
	case '!':
		p.maxAlign = nativeAlign

		i := 1

		if i < len(p.fmt) && isDigit(p.fmt[i]) {
			p.maxAlign = int(p.fmt[i] - '0')

			i++

			for i < len(p.fmt) && isDigit(p.fmt[i]) {
				if (int(limits.MaxInt)-int(p.fmt[i]-'0'))/10 < p.maxAlign {
					return nil, object.NewRuntimeError("size is too large")
				}

				p.maxAlign = p.maxAlign*10 + int(p.fmt[i]-'0')
				i++
			}

			if p.maxAlign == 0 || p.maxAlign > maxIntSize {
				return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", p.maxAlign, maxIntSize))
			}
		}

		p.fmt = p.fmt[i:]

		goto next
	default:
		return nil, p.ap.ArgError(0, "invalid format option '"+string(op)+"'")
	}

	opt = &kOption{
		typ:  typ,
		size: size,
	}

	align := size

	if align > 1 && typ != kChar {
		if align > p.maxAlign {
			align = p.maxAlign
		}

		if align&(align-1) != 0 { // is not power of two?
			return nil, p.ap.ArgError(0, "format ask for alignment not power of 2")
		}

		opt.padding = (align - p.off&(align-1)) & (align - 1) // (align - total % align) % align
	}

	// if p.total > int(limits.MaxInt)-(opt.padding+opt.size) {
	// return nil, p.ap.ArgError(0, "format result too large")
	// }

	// p.total += opt.padding + opt.size

	return opt, nil
}
