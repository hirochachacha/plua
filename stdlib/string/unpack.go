package string

import (
	"fmt"
	"math"
	"strings"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

// unpack(fmt, s, [, pos])
func unpack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	pos, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}

	if pos < 0 {
		pos = len(s) + 1 + pos
	}

	if pos <= 0 || len(s) < pos-1 {
		return nil, ap.ArgError(2, "initial position out of string")
	}

	u := &unpacker{
		optParser: optParser{
			ap:       ap,
			fmt:      fmt,
			maxAlign: 1,
			endian:   littleEndian,
			off:      pos - 1,
		},
		s: s,
	}

	return u.unpack()
}

type unpacker struct {
	optParser
	s string
}

func (u *unpacker) unpack() (rets []object.Value, err *object.RuntimeError) {
	for {
		opt, err := u.nextkOpt()
		if err != nil {
			return nil, err
		}

		if opt == nil {
			break
		}

		err = u.skipPadding(opt.padding)
		if err != nil {
			return nil, err
		}

		u.off += opt.padding

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
			err = u.skipPadding(opt.size)
			if err != nil {
				return nil, err
			}
		case kPaddingAlign, kNop:
		default:
			panic("unreachable")
		}

		u.off += opt.size
	}

	rets = append(rets, object.Integer(u.off+1))

	return rets, nil
}

func (u *unpacker) unpackUint64(opt *kOption) (uint64, *object.RuntimeError) {
	if len(u.s)-u.off < opt.size {
		return 0, u.ap.ArgError(0, "data string is too short")
	}

	s := u.s[u.off : u.off+opt.size]

	var u64 uint64

	switch u.endian {
	case littleEndian:
		if len(s) > 8 {
			for i := 0; i < 8; i++ {
				u64 |= uint64(s[i]) << uint(8*i)
			}
			for i := 8; i < len(s); i++ {
				if s[i] != 0 {
					return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
				}
			}
		} else {
			for i := 0; i < len(s); i++ {
				u64 |= uint64(s[i]) << uint(8*i)
			}
		}
	case bigEndian:
		if len(s) > 8 {
			for i := 0; i < len(s)-8; i++ {
				if s[i] != 0 {
					return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
				}
			}
			for i := len(s) - 8; i < len(s); i++ {
				u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
			}
		} else {
			for i := 0; i < len(s); i++ {
				u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
			}
		}
	default:
		panic("unreachable")
	}

	return u64, nil
}

func (u *unpacker) unpackInt64(opt *kOption) (int64, *object.RuntimeError) {
	if len(u.s)-u.off < opt.size {
		return 0, u.ap.ArgError(0, "data string is too short")
	}

	s := u.s[u.off : u.off+opt.size]

	var u64 uint64

	switch u.endian {
	case littleEndian:
		if s[len(s)-1]&0x80 != 0 {
			if len(s) > 8 {
				for i := 0; i < 8; i++ {
					u64 |= uint64(s[i]) << uint(8*i)
				}
				for i := 8; i < len(s); i++ {
					if s[i] != 0xff {
						return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
					}
				}
				u64 = (u64 - (1<<64 - 1)) - 1
			} else {
				for i := 0; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*i)
				}
				u64 = u64 - 1<<uint(len(s)*8)
			}
		} else {
			if len(s) > 8 {
				for i := 0; i < 8; i++ {
					u64 |= uint64(s[i]) << uint(8*i)
				}
				for i := 8; i < len(s); i++ {
					if s[i] != 0 {
						return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
					}
				}
			} else {
				for i := 0; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*i)
				}
			}
		}
	case bigEndian:
		if s[0]&0x80 != 0 {
			if len(s) > 8 {
				for i := 0; i < len(s)-8; i++ {
					if s[i] != 0xff {
						return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
					}
				}
				for i := len(s) - 8; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
				}
				u64 = (u64 - (1<<64 - 1)) - 1
			} else {
				for i := 0; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
				}
				u64 = u64 - 1<<uint(len(s)*8)
			}
		} else {
			if len(s) > 8 {
				for i := 0; i < len(s)-8; i++ {
					if s[i] != 0 {
						return 0, object.NewRuntimeError(fmt.Sprintf("%d-byte integer does not fit into Lua Integer", len(s)))
					}
				}
				for i := len(s) - 8; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
				}
			} else {
				for i := 0; i < len(s); i++ {
					u64 |= uint64(s[i]) << uint(8*(len(s)-1-i))
				}
			}
		}
	default:
		panic("unreachable")
	}

	return int64(u64), nil
}

func (u *unpacker) unpackUint(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackUint64(opt)
	if err != nil {
		return nil, err
	}

	return object.Integer(u64), nil
}

func (u *unpacker) unpackInt(opt *kOption) (object.Value, *object.RuntimeError) {
	i64, err := u.unpackInt64(opt)
	if err != nil {
		return nil, err
	}

	return object.Integer(i64), nil
}

func (u *unpacker) unpackFloat(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackUint64(opt)
	if err != nil {
		return nil, err
	}

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
	if len(u.s)-u.off < opt.size {
		return nil, u.ap.ArgError(0, "data string is too short")
	}

	val := object.String(u.s[u.off : u.off+opt.size])

	return val, nil
}

func (u *unpacker) unpackString(opt *kOption) (object.Value, *object.RuntimeError) {
	u64, err := u.unpackUint64(opt)
	if err != nil {
		return nil, err
	}

	if u64 > uint64(limits.MaxInt) {
		return nil, u.ap.ArgError(0, "integer overflow")
	}

	slen := int(u64)

	if len(u.s)-u.off-opt.size < slen {
		return nil, u.ap.ArgError(0, "data string is too short")
	}

	val := object.String(u.s[u.off+opt.size : u.off+opt.size+slen])

	u.off += slen

	return val, nil
}

func (u *unpacker) unpackZeroString() (object.Value, *object.RuntimeError) {
	slen := strings.IndexByte(u.s[u.off:], 0x00)
	if slen == -1 {
		return nil, u.ap.ArgError(0, "data string does not contains zero")
	}

	val := object.String(u.s[u.off : u.off+slen])

	u.off += slen + 1

	return val, nil
}

func (u *unpacker) skipPadding(padding int) *object.RuntimeError {
	if len(u.s)-padding < u.off {
		return u.ap.ArgError(0, "data string is too short")
	}

	return nil
}
