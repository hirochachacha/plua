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
	{"forward_jump_over_local_assign.lua", "forward jump over local 'name'"},
	{"unreachable_code.lua", "expected 'end', found 'NAME'"},
	{"unreachable_code2.lua", "expected 'EOF', found 'return'"},
}

func TestCompileError(t *testing.T) {
	c := compiler.NewCompiler()

	for i, test := range compileErrorTestCases {
		_, err := c.CompileFile(filepath.Join("testdata/errors", test.fname), compiler.Either)
		if err == nil {
			t.Fatalf("%d: got: nil, want: %q", i+1, test.error)
		}

		if !strings.Contains(err.Error(), test.error) {
			t.Errorf("%d: got: %q, want: %q", i+1, err.Error(), test.error)
		}
	}
}
