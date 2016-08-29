package parser

import (
	"fmt"

	"github.com/hirochachacha/blua/position"
)

type Error struct {
	Pos position.Position
	Err error
}

func (e Error) Error() string {
	if e.Pos.Filename != "" || e.Pos.IsValid() {
		return fmt.Sprintf("compiler/parser: %s: %v", e.Pos, e.Err)
	}
	return fmt.Sprintf("compiler/parser: %v", e.Err)
}
