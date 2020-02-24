package schema

import (
	"io"
	"math/rand"
)

type testNode []byte

func (n testNode) GenerateJSON(_ *Context, w io.Writer, _ *rand.Rand) error {
	_, err := w.Write(n)
	return err
}
