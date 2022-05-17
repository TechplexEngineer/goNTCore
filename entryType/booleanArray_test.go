package entryType

import (
	"fmt"
	"github.com/matryer/is"
	"testing"
)

func TestBooleanArray_MarshalJSON(t *testing.T) {
	is := is.New(t)
	ba := NewBooleanArray([]bool{true, false, true})

	json, err := ba.MarshalJSON()
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%s", json), "[true,false,true]")

}
