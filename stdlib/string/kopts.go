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
	length := len(s)

	var typ kType
	var size int

	maxAlign := 1
	endian := nativeEndian
	total := 0

	inpadding := false

	i := 0
	for {
		if i == length {
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
			size = limits.IntSize
			typ = kInt
			i++

			if i != length {
				if isDigit(s[i]) {
					size = int(s[i] - '0')
					i++

					if i != length {
						for isDigit(s[i]) {
							size = size*10 + int(s[i]-'0')
							i++

							if i == length {
								if size > maxIntSize {
									return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1, %d]", size, maxIntSize))
								}

								break
							}
						}
					}
				}
			}
		case 'I':
			size = limits.IntSize
			typ = kUint
			i++

			if i != length {
				if isDigit(s[i]) {
					size = int(s[i] - '0')
					i++

					if i != length {
						for isDigit(s[i]) {
							size = size*10 + int(s[i]-'0')
							i++

							if i == length {
								if size > maxIntSize {
									return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1, %d]", size, maxIntSize))
								}

								break
							}
						}
					}
				}
			}
		case 's':
			size = 8
			typ = kString
			i++

			if i != length {
				if isDigit(s[i]) {
					size = int(s[i] - '0')
					i++

					if i != length {
						for isDigit(s[i]) {
							size = size*10 + int(s[i]-'0')
							i++

							if i == length {
								break
							}
						}
					}
				}
			}
		case 'c':
			size = -1
			typ = kChar
			i++

			if i != length {
				if isDigit(s[i]) {
					size = int(s[i] - '0')
					i++

					if i != length {
						for isDigit(s[i]) {
							size = size*10 + int(s[i]-'0')
							i++

							if i == length {
								if size > maxIntSize {
									return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1, %d]", size, maxIntSize))
								}

								break
							}
						}
					}
				}
			}

			if size == -1 {
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

			if i == length {
				return
			}

			if isDigit(s[i]) {
				maxAlign = int(s[i] - '0')
				i++

				if i == length {
					return
				}

				for isDigit(s[i]) {
					maxAlign = size*10 + int(s[i]-'0')
					i++

					if i == length {
						if size > maxIntSize {
							return nil, object.NewRuntimeError(fmt.Sprintf("integral size (%d) out of limits [1, %d]", size, maxIntSize))
						}

						return
					}
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
