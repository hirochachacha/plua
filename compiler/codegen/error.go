package codegen

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
		return fmt.Sprintf("compiler/codegen: %v at %s", e.Err, e.Pos)
	}
	return fmt.Sprintf("compiler/codegen: %v", e.Err)
}
