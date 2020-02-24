package schema

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool_GenerateJSON(t *testing.T) {
	var (
		b Bool
		w bytes.Buffer
	)
	if assert.NoError(t, b.GenerateJSON(nil, &w, rand.New(fakeSource(0)))) {
		assert.Equal(t, string(trueJSON), w.String())
	}
	w.Reset()
	if assert.NoError(t, b.GenerateJSON(nil, &w, rand.New(fakeSource(1<<62)))) {
		assert.Equal(t, string(falseJSON), w.String())
	}
}

func BenchmarkBool_GenerateJSON(b *testing.B) {
	var n Bool
	for i := 0; i < b.N; i++ {
		_ = n.GenerateJSON(nil, ioutil.Discard, rand.New(rand.NewSource(1)))
	}
}
