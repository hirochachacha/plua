package compiler_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hirochachacha/plua/compiler"
)

var compileErrorTestCases = []struct {
	fname string
	error string
}{
	{"unresolved_goto.lua", "unknown label 'L' for jump"},
	{"unclosed_function.lua", "expected 'end', found 'EOF'"},
	{"label_duplication.lua", "label 'L' already defined"},
}

func TestCompileError(t *testing.T) {
	c := compiler.NewCompiler()

	for i, test := range compileErrorTestCases {
		_, err := c.CompileFile(filepath.Join("testdata/errors", test.fname), compiler.Either)
		if err == nil {
			t.Fatal()
		}

		if !strings.Contains(err.Error(), test.error) {
			t.Errorf("%d: got: %q, want: %q", i+1, err.Error(), test.error)
		}
	}
}
