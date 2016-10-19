package object

const MaxTagType = TM_CALL + 1

type TagType uint

func (t TagType) String() string {
	return tagNames[t]
}

const (
	TM_INDEX TagType = iota
	TM_NEWINDEX
	TM_GC
	TM_MODE
	TM_LEN
	TM_EQ
	TM_ADD
	TM_SUB
	TM_MUL
	TM_MOD
	TM_POW
	TM_DIV
	TM_IDIV
	TM_BAND
	TM_BOR
	TM_BXOR
	TM_SHL
	TM_SHR
	TM_UNM
	TM_BNOT
	TM_LT
	TM_LE
	TM_CONCAT
	TM_CALL
)

var tagNames = [...]string{
	TM_INDEX:    "__index",
	TM_NEWINDEX: "__newindex",
	TM_GC:       "__gc",
	TM_MODE:     "__mode",
	TM_LEN:      "__len",
	TM_EQ:       "__eq",
	TM_ADD:      "__add",
	TM_SUB:      "__sub",
	TM_MUL:      "__mul",
	TM_MOD:      "__mod",
	TM_POW:      "__pow",
	TM_DIV:      "__div",
	TM_IDIV:     "__idiv",
	TM_BAND:     "__band",
	TM_BOR:      "__bor",
	TM_BXOR:     "__bxor",
	TM_SHL:      "__shl",
	TM_SHR:      "__shr",
	TM_UNM:      "__unm",
	TM_BNOT:     "__bnot",
	TM_LT:       "__lt",
	TM_LE:       "__le",
	TM_CONCAT:   "__concat",
	TM_CALL:     "__call",
}
