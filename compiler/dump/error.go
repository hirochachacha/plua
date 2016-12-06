package dump

import "fmt"

type Error struct {
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("compiler/dump: %v", e.Err)
}
