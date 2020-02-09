package schema

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type WalkFn func(Node) (bool, error)

type Node interface {
	GenerateJSON(*Context, io.Writer) error

}

type Walker interface {
	Walk(fn WalkFn) error
}

// node is a helper type for unmarshal Node
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
				Range: IntRange{
					Min: 0,
					Max: 100,
				},
			}
		case "float":
			n.Node = &Float{
				Range: FloatRange{
					Min: 0,
					Max: 1,
				},
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
			n.Node = &Bool{}
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
		return fmt.Errorf("node should be either scalar or mapping, got: %s", value.Tag)
	}
	return nil
}

// nodeSlice is a helper type for unmarshal map[string]Node
type nodeMap map[string]Node

func (n *nodeMap) UnmarshalYAML(value *yaml.Node) error {
	var aux map[string]*node
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*n = make(nodeMap, len(aux))
	for k, v := range aux {
		if v == nil {
			return errors.New("empty node")
		}
		(*n)[k] = v.Node
	}
	return nil
}
