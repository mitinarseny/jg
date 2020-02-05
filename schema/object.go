package schema

import (
	"bufio"
	"errors"
	"strconv"

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

func (o Object) Generate(w *bufio.Writer) error {
	if err := w.WriteByte('{'); err != nil {
		return err
	}
	var i int
	for field, node := range o {
		if i > 0 {
			if err := w.WriteByte(','); err != nil {
				return err
			}
		}
		i++
		if _, err := w.WriteString(strconv.Quote(field) + ":"); err != nil {
			return err
		}
		if err := node.Generate(w); err != nil {
			return err
		}
	}
	return w.WriteByte('}')
}
