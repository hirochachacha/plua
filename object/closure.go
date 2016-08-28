package object

type iClosure interface {
	Prototype() *Proto
	GetUpvalue(i int) Value
	GetUpvalueName(i int) string
	GetUpvalueId(i int) LightUserdata
	SetUpvalue(i int, val Value)
	NUpvalues() int
	UpvalueJoin(i int, other *Closure, j int)
}

func (cl *Closure) Prototype() *Proto {
	return cl.Impl.Prototype()
}

func (cl *Closure) GetUpvalue(i int) Value {
	return cl.Impl.GetUpvalue(i)
}

func (cl *Closure) GetUpvalueName(i int) string {
	return cl.Impl.GetUpvalueName(i)
}

func (cl *Closure) GetUpvalueId(i int) LightUserdata {
	return cl.Impl.GetUpvalueId(i)
}

func (cl *Closure) SetUpvalue(i int, val Value) {
	cl.Impl.SetUpvalue(i, val)
}

func (cl *Closure) NUpvalues() int {
	return cl.Impl.NUpvalues()
}

func (cl *Closure) UpvalueJoin(i int, other *Closure, j int) {
	cl.Impl.UpvalueJoin(i, other, j)
}
