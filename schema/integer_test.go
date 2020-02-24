package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestIntRange_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantRes IntRange
		wantErr bool
	}{
		{
			name:    "default",
			data:    "",
			wantRes: defaultIntRange,
		},
		{
			name: "max only",
			data: "11",
			wantRes: IntRange{
				Min: defaultIntRange.Min,
				Max: 11,
			},
		},
		{
			name: "min max",
			data: "[1, 11]",
			wantRes: IntRange{
				Min: 1,
				Max: 11,
			},
		},
		{
			name:    "empty array",
			data:    "[]",
			wantErr: true,
		},
		{
			name:    "1 element in array",
			data:    "[1]",
			wantErr: true,
		},
		{
			name:    ">2 elements in array",
			data:    "[1, 2, 3]",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := defaultIntRange
			err := yaml.Unmarshal([]byte(tt.data), &r)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantRes, r)
		})
	}
}

func TestIntRange_validate(t *testing.T) {
	assert.NoError(t, defaultIntRange.validate())

	tests := []struct {
		name    string
		r       IntRange
		wantErr bool
	}{
		{
			name: "min < max",
			r: IntRange{
				Min: -1,
				Max: 1,
			},
			wantErr: false,
		},
		{
			name: "min = max",
			r: IntRange{
				Min: 0,
				Max: 0,
			},
			wantErr: true,
		},
		{
			name: "min > max",
			r: IntRange{
				Min: 1,
				Max: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
