package pattern

import (
	"reflect"
	"testing"
)

var testFindIndexCases = []struct {
	input    string
	pat      string
	off      int
	captures []Capture
}{
	{"", "", 0, []Capture{{0, 0, false}}},
	{"a", "a", 0, []Capture{{0, 1, false}}},
	{"b", "a", 0, nil},
	{"testfootest", "foo", 0, []Capture{{4, 7, false}}},
	{"testfootest", "foo.", 0, []Capture{{4, 8, false}}},
	{"baaac", "a+", 0, []Capture{{1, 4, false}}},
	{"bc", "a+", 0, nil},
	{"baaac", "a*", 0, []Capture{{0, 0, false}}},
	{"aaac", "a*", 0, []Capture{{0, 3, false}}},
	{"aab", "a+b+", 0, []Capture{{0, 3, false}}},
	{"aa", "a-", 0, []Capture{{0, 0, false}}},
	{"aab", "a-b", 0, []Capture{{0, 3, false}}},
	{"xaabaacba", "((a+)(b+))%2c%3", 0, []Capture{
		{1, 8, false},
		{1, 4, false},
		{1, 3, false},
		{3, 4, false},
	}},
	{"abaac", "[ab]*", 0, []Capture{{0, 4, false}}},
	{"abaac", "([ab]*)c", 0, []Capture{{0, 5, false}, {0, 4, false}}},
	{"a", "^a", 0, []Capture{{0, 1, false}}},
	{"ba", "^a", 0, nil},
	{"a", "a$", 0, []Capture{{0, 1, false}}},
	{"ab", "a$", 0, nil},
	{"$ab", "%$a", 0, []Capture{{0, 2, false}}},
	{"abcabc", "$", 0, []Capture{{6, 6, false}}},
	{"aabb", "%f[a]", 0, []Capture{{0, 0, false}}},
	{"aabb", "%f[b]", 0, []Capture{{2, 2, false}}},
	{"aa(bb)cc", "%b()", 0, []Capture{{2, 6, false}}},
	{"aa'bb'cc", "%b''", 0, []Capture{{2, 6, false}}},
	{"ab ! test", "[^%sa-z]", 0, []Capture{{3, 4, false}}},
	{`[]`, `^%[%]$`, 0, []Capture{{0, 2, false}}},
	{`["]`, `^%["%]$`, 0, []Capture{{0, 3, false}}},
	{`[string "]`, `^%[string "%]$`, 0, []Capture{{0, 10, false}}},
	{`[string "foo bar baz"]`, `^%[string [^x]*"%]$`, 0, []Capture{{0, 22, false}}},
	{`xyz`, `[^x]+`, 0, []Capture{{1, 3, false}}},
	{`xyz`, `[yz]+`, 0, []Capture{{1, 3, false}}},
	{`]]]a`, `[^]]`, 0, []Capture{{3, 4, false}}},
	{`xyz`, `()y()`, 0, []Capture{{1, 2, false}, {1, 1, true}, {2, 2, true}}},
	{"ab", "b", 1, []Capture{{1, 2, false}}},
	{" a ", "%w", 1, []Capture{{1, 2, false}}},
	{"x=y", "x%py", 0, []Capture{{0, 3, false}}},
	{"aa5bb", "[%d]", 0, []Capture{{2, 3, false}}},
	{"aa5bb", "[^%D]", 0, []Capture{{2, 3, false}}},
	{"aa", "%f[a]", 0, []Capture{{0, 0, false}}},
	{"aa", "%f[^a]", 0, []Capture{{2, 2, false}}},
	{"aa", "%f[^\x00]", 0, []Capture{{0, 0, false}}},
	{"aa", "%f[\x00]", 0, []Capture{{2, 2, false}}},
}

func TestFindIndex(t *testing.T) {
	for i, test := range testFindIndexCases {
		got, err := FindIndex(test.input, test.pat, test.off)
		if err != nil {
			t.Fatal(err)
		}
		want := test.captures
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v, want %v", i+1, got, want)
		}
	}
}

var testFindAllIndexCases = []struct {
	input       string
	pat         string
	allCaptures [][]Capture
}{
	{"abc", "", [][]Capture{
		{{0, 0, false}},
		{{1, 1, false}},
		{{2, 2, false}},
		{{3, 3, false}},
	}},
	{"a bc", " *", [][]Capture{
		{{0, 0, false}},
		{{1, 2, false}},
		{{3, 3, false}},
		{{4, 4, false}},
	}},
	{"abc abc ababcabc", "abc", [][]Capture{
		{{0, 3, false}},
		{{4, 7, false}},
		{{10, 13, false}},
		{{13, 16, false}},
	}},
}

func TestFindAllIndex(t *testing.T) {
	for i, test := range testFindAllIndexCases {
		got, err := FindAllIndex(test.input, test.pat, -1)
		if err != nil {
			t.Fatal(err)
		}
		want := test.allCaptures
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v, want %v", i+1, got, want)
		}
	}
}

var testReplaceAllCases = []struct {
	input  string
	pat    string
	repl   string
	output string
}{
	{"abc", "", "-", "-a-b-c-"},
	{"a bc", " *", "-", "-a-b-c-"},
}

func TestReplaceAllFunc(t *testing.T) {
	for i, test := range testReplaceAllCases {
		got, _, err := ReplaceAllFunc(test.input, test.pat, func(caps []Capture) (string, error) {
			return test.repl, nil
		}, -1)
		if err != nil {
			t.Fatal(err)
		}
		want := test.output
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v, want %v", i+1, got, want)
		}
	}
}
