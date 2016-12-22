// +build gofuzz

package plua

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFuzz(t *testing.T) {
	processFile := func(path string) (err error) {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			if r := recover(); r != nil {
				if rerr, ok := r.(error); ok {
					err = rerr
				} else {
					err = fmt.Errorf("%v", r)
				}
			}
		}()

		Fuzz(data)

		return nil
	}

	dir := filepath.Join("_fuzz", "corpus")

	d, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	names, err := d.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		err := processFile(filepath.Join(dir, name))
		if err != nil {
			t.Error(err)
		}
	}
}
