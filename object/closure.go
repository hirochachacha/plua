package object

type Closure interface {
	Value

	Prototype() *Proto
	GetUpvalue(i int) Value
	GetUpvalueName(i int) string
	GetUpvalueId(i int) LightUserdata
	SetUpvalue(i int, val Value)
	NUpvalues() int
	UpvalueJoin(i int, other Closure, j int)
}
