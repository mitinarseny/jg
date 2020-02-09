package schema

import (
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Integer struct {
	Range IntRange `yaml:"range"`
}

func (i *Integer) UnmarshalYAML(value *yaml.Node) error {
	type raw Integer
	aux := raw(Integer{Range:IntRange{
			Min: 0,
			Max: 100,
		}})
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*i = Integer(aux)
	return nil
}

func (i *Integer) GenerateJSON(_ *Context, w io.Writer) error {
	_, err := w.Write([]byte(strconv.FormatInt(i.Range.Rand(), 10)))
	return err
}
