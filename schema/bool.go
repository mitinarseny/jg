package schema

import (
	"io"
	"math/rand"
)

type Bool struct{}

var (
	trueJSON  = []byte("true")
	falseJSON = []byte("false")
)

func (b Bool) GenerateJSON(_ *Context, w io.Writer) error {
	v := falseJSON
	if rand.Float64() < 0.5 {
		v = trueJSON
	}
	_, err := w.Write(v)
	return err
}
