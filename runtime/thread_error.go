package runtime

import "github.com/hirochachacha/plua/object"

func (th *thread) trackErrorOnce(err *object.RuntimeError) {
	if d := th.getInfo(0, "Slnt"); d != nil {
		err.Traceback = append(err.Traceback, th.stackTrace(d))
	}
}

func (th *thread) trackError(err *object.RuntimeError) {
	level := 0
	for {
		d := th.getInfo(level, "Slnt")
		if d == nil {
			break
		}

		err.Traceback = append(err.Traceback, th.stackTrace(d))

		level++
	}
}

func (th *thread) error(err *object.RuntimeError) {
	if th.status != object.THREAD_ERROR {
		th.trackError(err)
		th.status = object.THREAD_ERROR
		th.err = err
	}
}
