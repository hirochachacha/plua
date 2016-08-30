package tables

import (
	"sort"

	"github.com/hirochachacha/plua/object"
)

type tableSorter struct {
	a    []object.Value
	less func(x, y object.Value) bool
}

func (ts *tableSorter) Len() int {
	return len(ts.a)
}

func (ts *tableSorter) Swap(i, j int) {
	ts.a[i], ts.a[j] = ts.a[j], ts.a[i]
}

func (ts *tableSorter) Less(i, j int) bool {
	return ts.less(ts.a[i], ts.a[j])
}

func (ts *tableSorter) Sort() {
	sort.Sort(ts)
}
