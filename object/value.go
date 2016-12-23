package object

type Value interface {
	Type() Type
	String() string
}

const (
	True  = Boolean(true)
	False = Boolean(false)
)
