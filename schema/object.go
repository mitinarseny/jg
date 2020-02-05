package schema

import (
	"errors"

	"gopkg.in/yaml.v3"
)

type Object map[string]Node

func (o *Object) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Fields nodeMap `yaml:"fields"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Fields == nil {
		return errors.New("object must specify its fields")
	}
	*o = Object(aux.Fields)
	return nil
}

func (o Object) Generate() (interface{}, error) {
	res := make(map[string]interface{}, len(o))
	var err error
	for field, node := range o {
		res[field], err = node.Generate()
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
