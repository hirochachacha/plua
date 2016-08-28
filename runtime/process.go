package runtime

import (
	"github.com/hirochachacha/blua/object"
)

type process struct {
	object.Thread
}

func NewProcess() object.Process {
	return &process{newMainThread()}
}

func (p *process) Fork() object.Process {
	th := p.Thread.(*thread)

	th = th.newThreadWith(threadMain, th.env, 0)

	return &process{th}
}

func (p *process) Exec(proto *object.Proto, args ...object.Value) (rets []object.Value, err error) {
	p.Load(proto)

	return p.Resume(args...)
}
