package schema

import (
	"bufio"
	"math/rand"
)

type Bool struct{}

const (
	trueJSON  = "true"
	falseJSON = "false"
)

func (b *Bool) Generate(w *bufio.Writer) error {
	v := falseJSON
	if rand.Float64() < 0.5 {
		v = trueJSON
	}
	_, err := w.WriteString(v)
	return err
}
