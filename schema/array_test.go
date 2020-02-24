package schema

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArray_GenerateJSON(t *testing.T) {
	var w bytes.Buffer
	tests := []struct {
		name     string
		length   uint64
		elements string
		wantW    string
	}{
		{
			name:   "empty",
			length: 0,
			wantW:  "[]",
		},
		{
			name:     "only element",
			length:   1,
			elements: "only",
			wantW:    "[only]",
		},
		{
			name:     "10 elements",
			length:   10,
			elements: "n",
			wantW:    "[n,n,n,n,n,n,n,n,n,n]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.Reset()
			a := &Array{
				Length: Length{
					Min: tt.length,
					Max: tt.length,
				},
				Elements: testNode(tt.elements),
			}
			require.NoError(t, a.GenerateJSON(nil, &w, nil))
			require.Equal(t, tt.wantW, w.String())
		})
	}
}

func BenchmarkArray_GenerateJSON_Small(b *testing.B) {
	n := Array{
		Length: Length{
			Min: 100,
			Max: 100,
		},
		Elements: testNode{},
	}
	for i := 0; i < b.N; i++ {
		_ = n.GenerateJSON(nil, ioutil.Discard, rand.New(rand.NewSource(1)))
	}
}

func BenchmarkArray_GenerateJSON_Medium(b *testing.B) {
	n := Array{
		Length: Length{
			Min: 1e4,
			Max: 1e4,
		},
		Elements: testNode{},
	}
	for i := 0; i < b.N; i++ {
		_ = n.GenerateJSON(nil, ioutil.Discard, rand.New(rand.NewSource(1)))
	}
}

func BenchmarkArray_GenerateJSON_Large(b *testing.B) {
	n := Array{
		Length: Length{
			Min: 1e7,
			Max: 1e7,
		},
		Elements: nil,
	}
	for i := 0; i < b.N; i++ {
		_ = n.GenerateJSON(nil, ioutil.Discard, nil)
	}
}
