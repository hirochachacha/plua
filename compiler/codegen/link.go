package codegen

type kind int

const (
	linkLocal kind = iota
	linkUpval
)

type link struct {
	kind  kind
	index int

	// kind == linkLocal => v == index of stack (stack pointer)
	// kind == linkUpval => v == index of g.UpvalueDescs
}
