package format

import (
	"testing"

	"github.com/hirochachacha/plua/object"
)

func TestMain(t *testing.T) {
	s, err := Format("%d -e \"%s\" & echo $!", object.String("4"), object.String("bar"))
	if err != nil {
		panic(err)
	}

	println("%d -e \"%s\" & echo $!")

	println(s)

	s, err = Format("%.14f", object.Number(4))
	if err != nil {
		panic(err)
	}
	println(s)
}
