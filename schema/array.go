package schema

import (
	"errors"

	"gopkg.in/yaml.v3"
)

type Array struct {
	Range    IntRange
	Elements Node
}

func (a *Array) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Range    IntRange `yaml:"range"`
		Elements *node    `yaml:"elements"`
	}{
		Range: IntRange{
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
		Range:    aux.Range,
		Elements: aux.Elements.Node,
	}
	return nil
}

func (a *Array) Generate() (interface{}, error) {
	elNum := a.Range.Rand()
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
