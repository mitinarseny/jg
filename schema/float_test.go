package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFloatRange_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantRes FloatRange
		wantErr bool
	}{
		{
			name:    "default",
			data:    "",
			wantRes: defaultFloatRange,
		},
		{
			name: "max only",
			data: "3.14",
			wantRes: FloatRange{
				Min: defaultFloatRange.Min,
				Max: 3.14,
			},
		},
		{
			name: "min max",
			data: "[1, 3.14]",
			wantRes: FloatRange{
				Min: 1,
				Max: 3.14,
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
			r := defaultFloatRange
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

func TestFloatRange_validate(t *testing.T) {
	assert.NoError(t, defaultFloatRange.validate())

	tests := []struct {
		name    string
		r       FloatRange
		wantErr bool
	}{
		{
			name: "min < max",
			r: FloatRange{
				Min: -1,
				Max: 1,
			},
			wantErr: false,
		},
		{
			name: "min = max",
			r: FloatRange{
				Min: 0,
				Max: 0,
			},
			wantErr: true,
		},
		{
			name: "min > max",
			r: FloatRange{
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
