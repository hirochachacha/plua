package parser

import (
	"fmt"

	"github.com/hirochachacha/plua/position"
)

type Error struct {
	Pos position.Position
	Err error
}

func (e Error) Error() string {
	if e.Pos.SourceName != "" || e.Pos.IsValid() {
		return fmt.Sprintf("compiler/parser: %v at %s", e.Err, e.Pos)
	}
	return fmt.Sprintf("compiler/parser: %v", e.Err)
}
