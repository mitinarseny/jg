package schema

import (
	"bufio"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Float struct {
	Range FloatRange
}

func (f *Float) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Range FloatRange `yaml:"range"`
	}{
		Range: FloatRange{
			Min: 0,
			Max: 1,
		},
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*f = Float{
		Range: aux.Range,
	}
	return nil
}

func (f *Float) Generate(w *bufio.Writer) error {
	_, err := w.WriteString(strconv.FormatFloat(f.Range.Rand(), 'f', -1, 64))
	return err
}
