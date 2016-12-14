package object

type Value interface {
	Type() Type
	String() string
}

var (
	True  = Boolean(true)
	False = Boolean(false)
)
