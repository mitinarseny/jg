package schema

import (
	"errors"

	"gopkg.in/yaml.v3"
)

type Array struct {
	Length   Length
	Elements Node
}

func (a *Array) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Length   Length `yaml:"length"`
		Elements *node  `yaml:"elements"`
	}{
		Length: Length{
			Min: 0,
			Max: 10,
		},
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Elements == nil {
		return errors.New("array must specify its elements")
	}
	*a = Array{
		Length:   aux.Length,
		Elements: aux.Elements.Node,
	}
	return nil
}

func (a *Array) Generate() (interface{}, error) {
	elNum := a.Length.Rand()
	res := make([]interface{}, 0, elNum)
	for i := 0; i < elNum; i++ {
		gen, err := a.Elements.Generate()
		if err != nil {
			return nil, err
		}
		res = append(res, gen)
	}
	return res, nil
}
