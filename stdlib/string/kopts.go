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
	endian  byteOrder
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

func parsekOpts(ap *fnutil.ArgParser, s string) (opts []*kOption, err *object.RuntimeError) {
	var typ kType
	var size int

	maxAlign := 1
	endian := nativeEndian
	total := 0

	inpadding := false

	i := 0
	for {
		if i == len(s) {
			if inpadding {
				return nil, ap.ArgError(0, "invalid next option for option 'X'")
			}

			break
		}

		b := s[i]

		switch b {
		case 'b':
			size = 1
			typ = kInt
			i++
		case 'B':
			size = 1
			typ = kUint
			i++
		case 'h':
			size = 2
			typ = kInt
			i++
		case 'H':
			size = 2
			typ = kUint
			i++
		case 'l':
			size = 8
			typ = kInt
			i++
		case 'L':
			size = 8
			typ = kUint
			i++
		case 'j':
			size = 8
			typ = kInt
			i++
		case 'J':
			size = 8
			typ = kUint
			i++
		case 'T':
			size = 8
			typ = kUint
			i++
		case 'f':
			size = 4
			typ = kFloat
			i++
		case 'd':
			size = 8
			typ = kFloat
			i++
		case 'n':
			size = 8
			typ = kFloat
			i++
		case 'i':
			i++

			size = limits.IntSize

			typ = kInt

			if i < len(s) && isDigit(s[i]) {
				size = int(s[i] - '0')

				i++

				for i < len(s) && isDigit(s[i]) {
					if (int(limits.MaxInt)-int(s[i]-'0'))/10 < size {
						return nil, object.NewRuntimeError("integer overflow")
					}

					size = size*10 + int(s[i]-'0')
					i++
				}

				if size == 0 || size > maxIntSize {
					return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
				}
			}
		case 'I':
			i++

			size = limits.IntSize

			typ = kUint

			if i < len(s) && isDigit(s[i]) {
				size = int(s[i] - '0')

				i++

				for i < len(s) && isDigit(s[i]) {
					if (int(limits.MaxInt)-int(s[i]-'0'))/10 < size {
						return nil, object.NewRuntimeError("integer overflow")
					}

					size = size*10 + int(s[i]-'0')
					i++
				}

				if size == 0 || size > maxIntSize {
					return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
				}
			}
		case 's':
			i++

			size = 8

			typ = kString

			if i < len(s) && isDigit(s[i]) {
				size = int(s[i] - '0')

				i++

				for i < len(s) && isDigit(s[i]) {
					if (int(limits.MaxInt)-int(s[i]-'0'))/10 < size {
						return nil, object.NewRuntimeError("integer overflow")
					}

					size = size*10 + int(s[i]-'0')
					i++
				}

				if size == 0 || size > maxIntSize {
					return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
				}
			}
		case 'c':
			i++

			size = 0

			typ = kChar

			if i < len(s) && isDigit(s[i]) {
				size = int(s[i] - '0')

				i++

				for i < len(s) && isDigit(s[i]) {
					if (int(limits.MaxInt)-int(s[i]-'0'))/10 < size {
						return nil, object.NewRuntimeError("integer overflow")
					}

					size = size*10 + int(s[i]-'0')
					i++
				}

				if size == 0 || size > maxIntSize {
					return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", size, maxIntSize))
				}
			}

			if size == 0 {
				return nil, object.NewRuntimeError("missing size for format option 'c'")
			}
		case 'z':
			size = 0
			typ = kZeroString
			i++
		case 'x':
			size = 1
			typ = kPadding
			i++
		case 'X':
			size = 0
			typ = kPaddingAlign
			i++

			opts = append(opts, &kOption{
				typ:    typ,
				size:   size,
				endian: endian,
			})

			inpadding = true

			continue
		case ' ':
			size = 0
			typ = kNop
			i++
		case '<':
			i++
			endian = littleEndian

			continue
		case '>':
			i++
			endian = bigEndian

			continue
		case '=':
			i++
			endian = nativeEndian

			continue
		case '!':
			i++

			maxAlign = nativeAlign

			if i < len(s) && isDigit(s[i]) {
				maxAlign = int(s[i] - '0')

				i++

				for i < len(s) && isDigit(s[i]) {
					if (int(limits.MaxInt)-int(s[i]-'0'))/10 < maxAlign {
						return nil, object.NewRuntimeError("integer overflow")
					}

					maxAlign = maxAlign*10 + int(s[i]-'0')
					i++
				}

				if maxAlign == 0 || maxAlign > maxIntSize {
					return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1,%d]", maxAlign, maxIntSize))
				}
			}

			continue
		default:
			return nil, object.NewRuntimeError("invalid format option '" + string(b) + "'")
		}

		align := size

		if !(align <= 1 || typ == kChar) && len(opts) > 0 {
			if align > maxAlign {
				align = maxAlign

				if align&(align-1) != 0 { // is not power of two?
					return nil, ap.ArgError(0, "format ask for alignment not power of 2")
				}

				padding := (align - total&(align-1)) & (align - 1) // (align - total % align) % align

				total += padding

				opts[len(opts)-1].padding = padding
			}
		}

		if inpadding {
			if typ == kChar || align == 0 {
				return nil, ap.ArgError(0, "invalid next option for option 'X'")
			}

			inpadding = false
		} else {
			total += size

			opts = append(opts, &kOption{
				typ:    typ,
				size:   size,
				endian: endian,
			})
		}
	}

	return
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}
