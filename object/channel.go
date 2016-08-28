package object

type Channel interface {
	Value

	Send(val Value)
	Recv() (val Value, ok bool)
	Close()
}
