package schema

import (
	"bufio"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Integer struct {
	Range IntRange
}

func (i *Integer) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Range IntRange `yaml:"range"`
	}{
		Range: IntRange{
			Min: 0,
			Max: 100,
		},
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*i = Integer{
		Range: aux.Range,
	}
	return nil
}

func (i *Integer) Generate(w *bufio.Writer) error {
	_, err := w.WriteString(strconv.FormatInt(i.Range.Rand(), 10))
	return err
}
