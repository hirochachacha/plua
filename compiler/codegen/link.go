package codegen

type kind int

const (
	linkLocal kind = iota
	linkUpval
)

type link struct {
	kind kind
	v    int
}
