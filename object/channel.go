package object

type iChannel interface {
	Send(val Value)
	Recv() (val Value, ok bool)
	Close()
}

func (ch *Channel) Send(val Value) {
	ch.Impl.Send(val)
}

func (ch *Channel) Recv() (val Value, ok bool) {
	return ch.Impl.Recv()
}

func (ch *Channel) Close() {
	ch.Impl.Close()
}
