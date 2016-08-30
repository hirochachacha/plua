package runtime

import (
	"reflect"

	"github.com/hirochachacha/plua/object"
)

type channel chan object.Value

func newChannel(capacity int) object.Channel {
	return channel(make(chan object.Value, capacity))
}

func (ch channel) Type() object.Type {
	return object.TCHANNEL
}

func (ch channel) Send(val object.Value) {
	ch <- val
}

func (ch channel) Recv() (object.Value, bool) {
	val, ok := <-ch

	return val, ok
}

func (ch channel) Close() {
	close(ch)
}

func (th *thread) Select(cases []object.SelectCase) (int, object.Value, bool) {

	rcases := make([]reflect.SelectCase, len(cases))

	for i, cas := range cases {
		switch cas.Dir {
		case "send":
			rcases[i] = reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(cas.Chan),
				Send: reflect.ValueOf(cas.Send),
			}
		case "recv":
			rcases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(cas.Chan),
				Send: reflect.ValueOf(cas.Send),
			}
		case "default":
			rcases[i] = reflect.SelectCase{
				Dir: reflect.SelectDefault,
			}
		default:
			panic("unexpected")
		}
	}

	chosen, recv, recvOK := reflect.Select(rcases)

	return chosen, recv.Interface().(object.Value), recvOK
}
