package schema

import (
	"errors"
	"fmt"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type Node interface {
	Generate() (interface{}, error)
}

type Bool struct{}

func (b *Bool) Generate() (interface{}, error) {
	return rand.Float64() < 0.5, nil
}

type Integer struct {
	Min, Max int
}

func (i *Integer) Generate() (interface{}, error) {
	return rand.Int63(), nil
}

type Float struct {
	Min, Max float64
}

func (f *Float) Generate() (interface{}, error) {
	return rand.Float64(), nil
}

type String struct {
}

func (s *String) Generate() (interface{}, error) {
	return "example", nil
}

type Array struct {
	Min, Max int
	Elements Node
}

func (a *Array) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Min      int  `yaml:"min"`
		Max      int  `yaml:"max"`
		Elements node `yaml:"elements"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*a = Array{
		Min:      aux.Min,
		Max:      aux.Max,
		Elements: aux.Elements.Node,
	}
	return nil
}

func (a *Array) Generate() (interface{}, error) {
	elNum := int64(a.Min) + rand.Int63n(int64(a.Max-a.Min+1))
	res := make([]interface{}, 0, elNum)
	for i := int64(0); i < elNum; i++ {
		gen, err := a.Elements.Generate()
		if err != nil {
			return nil, err
		}
		res = append(res, gen)
	}
	return res, nil
}

type Object map[string]Node

func (o *Object) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Fields nodeMap `yaml:"fields"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
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

type Enum []interface{}

func (e *Enum) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Choices []interface{} `yaml:"choices"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	// TODO: check len of aux: 0 and 1 are meaningless
	*e = aux.Choices
	return nil
}

func (e Enum) Generate() (interface{}, error) {
	return e[rand.Intn(len(e))], nil
}

type nodeMap map[string]Node

func (n *nodeMap) UnmarshalYAML(value *yaml.Node) error {
	var aux map[string]node
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*n = make(nodeMap, len(aux))
	for k, v := range aux {
		(*n)[k] = v.Node
	}
	return nil
}

type nodeSlice []Node

func (n *nodeSlice) UnmarshalYAML(value *yaml.Node) error {
	var aux []node
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*n = make(nodeSlice, 0, len(aux))
	for _, v := range aux {
		*n = append(*n, v)
	}
	return nil
}

// node is a helper struct for unmarshal
type node struct {
	Node
}

func (n *node) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var typ string
		if err := value.Decode(&typ); err != nil {
			return err
		}
		switch typ {
		case "bool":
			n.Node = &Bool{}
		case "integer":
			n.Node = &Integer{
				Min: 0,
				Max: 100,
			}
		case "float":
			n.Node = &Float{
				Min: 0,
				Max: 1,
			}
		case "array", "object", "enum":
			return fmt.Errorf("unable to unmarshal inline %q", typ)
		default:
			return fmt.Errorf("unsupported type: %q", typ)
		}
	case yaml.MappingNode:
		var aux struct {
			Type *string `yaml:"type"`
		}
		if err := value.Decode(&aux); err != nil {
			return err
		}
		if aux.Type == nil {
			return errors.New("type should be specified")
		}
		switch typ := *aux.Type; typ {
		case "bool":
			var tmp Bool
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "integer":
			var tmp Integer
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "float":
			var tmp Float
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "string":
			var tmp String
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "array":
			var tmp Array
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "object":
			var tmp Object
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		case "enum":
			var tmp Enum
			if err := value.Decode(&tmp); err != nil {
				return err
			}
			n.Node = &tmp
		default:
			return fmt.Errorf("unsupported type: %q", typ)
		}
	default:
		return errors.New("node should be either scalar or mapping")
	}
	return nil
}
