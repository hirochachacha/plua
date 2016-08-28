package runtime

import (
	"github.com/hirochachacha/blua/internal/tables"
	"github.com/hirochachacha/blua/object"
)

func newTableSize(asize, msize int) object.Table {
	return tables.NewTableSize(asize, msize)
}

func newTableArray(a []object.Value) object.Table {
	return tables.NewTableArray(a)
}

func newConcurrentTableSize(asize, msize int) object.Table {
	return tables.NewConcurrentTableSize(asize, msize)
}

func newLockedTableSize(asize, msize int) object.Table {
	return tables.NewLockedTableSize(asize, msize)
}
