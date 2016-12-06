package undump

import "fmt"

type Error struct {
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("compiler/undump: %v", e.Err)
}
