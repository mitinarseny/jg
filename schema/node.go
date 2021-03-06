package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type WalkFn func(Node) (bool, error)

type Node interface {
	GenerateJSON(*Context, io.Writer, *rand.Rand) error
}

type Walker interface {
	Walk(fn WalkFn) error
}

func Walk(n Node, fn WalkFn) error {
	proceed, err := fn(n)
	if !proceed {
		return err
	}
	if w, ok := n.(Walker); ok {
		return (&Errors{}).Add(err).Add(w.Walk(fn)).Err()
	}
	return err
}

type nodeType string

const (
	boolType    nodeType = "bool"
	integerType nodeType = "int"
	floatType   nodeType = "float"
	stringType  nodeType = "string"
	arrayType   nodeType = "array"
	objectType  nodeType = "object"
)

// node is a helper type for unmarshal Node
type node struct {
	Node
}

func (n *node) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return n.unmarshalYAMLScalar(value)
	case yaml.MappingNode:
		return n.unmarshalYAMLMapping(value)
	}
	return &yamlError{
		line: value.Line,
		err:  fmt.Errorf("node should be either scalar or mapping, got: %s", value.Tag),
	}
}

func (n *node) unmarshalYAMLScalar(value *yaml.Node) error {
	var typ nodeType
	if err := value.Decode(&typ); err != nil {
		return err
	}
	switch typ {
	case boolType:
		n.Node = &Bool{}
	case integerType:
		n.Node = &Integer{
			Range: &defaultIntRange,
		}
	case floatType:
		n.Node = &Float{
			Range: &defaultFloatRange,
		}
	case arrayType, objectType:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("unable to unmarshal inline %q", typ),
		}
	default:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("unsupported type: %q", typ),
		}
	}
	return nil
}

func (n *node) unmarshalYAMLMapping(value *yaml.Node) error {
	var aux struct {
		Type nodeType `yaml:"type"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Type == "" {
		return &yamlError{
			line: value.Line,
			err:  errors.New("type is required"),
		}
	}
	switch aux.Type {
	case boolType:
		n.Node = Bool{}
	case integerType:
		n.Node = &Integer{}
	case floatType:
		n.Node = &Float{}
	case stringType:
		n.Node = &String{}
	case arrayType:
		n.Node = &Array{}
	case objectType:
		n.Node = &Object{}
	default:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("unsupported type: %q", aux.Type),
		}
	}
	return value.Decode(n.Node)
}

// nodeMap is a helper type for unmarshal map[string]Node
type nodeMap map[string]Node

func (n *nodeMap) UnmarshalYAML(value *yaml.Node) error {
	var aux map[string]*node
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*n = make(nodeMap, len(aux))
	for k, v := range aux {
		if v == nil {
			return &yamlError{
				line: value.Line,
				err:  errors.New("empty node"),
			}
		}
		(*n)[k] = v.Node
	}
	return nil
}
