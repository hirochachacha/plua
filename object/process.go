package object

type Process interface {
	Load(p *Proto)
	Resume(args ...Value) (rets []Value, err error)

	MainThread() *Thread

	// ↑ thread specific APIs

	NewTableSize(asize, msize int) *Table
	NewTableArray(a []Value) *Table
	NewThread() *Thread
	NewGoThread() *Thread
	NewUserdata(x interface{}) *Userdata
	NewClosure(p *Proto) *Closure
	NewChannel(capacity int) *Channel

	Registry() *Table
	Globals() *Table
	Loaded() *Table
	Preload() *Table

	GetMetatable(val Value) *Table
	SetMetatable(val Value, mt *Table)

	Requiref(openf GoFunction, modname string) bool

	NewMetatableNameSize(tname string, alen, mlen int) *Table
	GetMetatableName(tname string) *Table
	GetMetaField(val Value, field string) Value

	ValueOf(x interface{}) Value

	Fork() Process
}
