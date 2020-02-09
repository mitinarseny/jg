package schema

import (
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Float struct {
	Range FloatRange `yaml:"float"`
}

func (f *Float) UnmarshalYAML(value *yaml.Node) error {
	type raw Float
	aux := raw(Float{Range: FloatRange{
		Min: 0,
		Max: 1,
	}})
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*f = Float(aux)
	return nil
}

func (f *Float) GenerateJSON(_ *Context, w io.Writer) error {
	_, err := w.Write([]byte(strconv.FormatFloat(f.Range.Rand(), 'f', -1, 64)))
	return err
}
