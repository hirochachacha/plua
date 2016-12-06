package compiler

import "fmt"

type Error struct {
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("compiler: %v", e.Err)
}
