package schema

import "math/rand"

type Bool struct{}

func (b *Bool) Generate() (interface{}, error) {
	return rand.Float64() < 0.5, nil
}
