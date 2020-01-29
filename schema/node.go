package schema

import (
	"errors"
	"fmt"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type Node struct {
	Type     Type             `yaml:"type"`
	Value    interface{}      `yaml:"value"`
	Fields   map[string]Node `yaml:"fields"`
	Elements *Node            `yaml:"elements"`
	Range    Range            `yaml:"range"`
	Choices  []interface{}    `yaml:"choices"`
	From     *string          `yaml:"from"`
}

func (n *Node) Generate() interface{} {
	if n.Value != nil {
		return n.Value
	}
	switch n.Type {
	case Object:
		res := make(map[string]interface{}, len(n.Fields))
		for field, node := range n.Fields {
			res[field] = node.Generate()
		}
		return res
	case Array:
		elNum := n.Range.Min + uint(rand.Intn(int(n.Range.Max-n.Range.Min+1)))
		res := make([]interface{}, 0, elNum)
		for i := uint(0); i < elNum; i++ {
			res = append(res, n.Elements.Generate())
		}
		return res
	case Enum:
		return n.Choices[rand.Intn(len(n.Choices))]
	}
	return nil
}

type Range struct {
	Min  uint `yaml:"min"`
	Max  uint `yaml:"min"`
	Step uint `yaml:"step"`
}

func (r *Range) UnmarshalYAML(value *yaml.Node) error {
	var exact uint
	if err := value.Decode(&exact); err == nil {
		*r = Range{
			Min:  exact,
			Max:  exact,
			Step: 1,
		}
		return nil
	}

	full := struct {
		Min  uint `yaml:"min"`
		Max  uint `yaml:"min"`
		Step uint `yaml:"step"`
	}{
		Min:  0,
		Max:  10,
		Step: 1,
	}
	if err := value.Decode(&full); err != nil {
		return err
	}
	*r = Range{
		Min:  full.Min,
		Max:  full.Max,
		Step: full.Step,
	}
	return nil
}

func (n *Node) UnmarshalYAML(value *yaml.Node) error {
	var aux interface{}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	switch aux.(type) {
	case bool:
		*n = Node{
			Type:  Bool,
			Value: aux,
		}
		return nil
	case string:
		*n = Node{
			Type:  String,
			Value: aux,
		}
		return nil
	case int:
		*n = Node{
			Type:  Integer,
			Value: aux,
		}
		return nil
	case []interface{}:
		*n = Node{
			Type:  Array,
			Value: aux,
		}
		return nil
	}

	var full struct {
		Type     *Type            `yaml:"type"`
		Value    interface{}      `yaml:"value"`
		Fields   map[string]Node `yaml:"fields"`
		Elements *Node            `yaml:"elements"`
		Range    Range            `yaml:"range"`
		Choices  []interface{}    `yaml:"choices"`
		From     *string          `yaml:"from"`
	}
	if err := value.Decode(&full); err != nil {
		return err
	}

	// type should be defined explicitly
	if full.Type == nil {
		return errors.New("empty type")
	}

	*n = Node{
		Type:     *full.Type,
		Value:    full.Value,
		Fields:   full.Fields,
		Elements: full.Elements,
		Range:    full.Range,
		Choices:  full.Choices,
		From:     full.From,
	}

	return n.validate()
}

func (n *Node) validate() error {
	if n.Value != nil {
		if err := n.checkValueType(); err != nil {
			return fmt.Errorf("wrong type: %w", err)
		}
		switch {
		case n.From != nil:
			return errors.New("combination of value and from")
		case n.Elements != nil:
			return errors.New("combination of value and elements")
		case n.Choices != nil:
			return errors.New("combination of value and choices")
		}
	}

	switch {
	case n.Type == Array && n.Elements == nil:
		return errors.New("array must specify its elements")
	case n.Type == Object && n.Fields == nil:
		return errors.New("object must specify its fields")
	}

	return nil
}

func (n *Node) checkValueType() error {
	var actual Type
	switch n.Value.(type) {
	case bool:
		actual = Bool
	case string:
		actual = String
	case int:
		actual = Integer
	case []interface{}:
		actual = Array
	default:
		actual = Object
	}
	if actual != n.Type {
		return fmt.Errorf("%v is not %q", n.Value, n.Type)
	}
	return nil
}

type Type string

func (t *Type) UnmarshalYAML(value *yaml.Node) error {
	var aux string
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if _, found := types[Type(aux)]; !found {
		return fmt.Errorf("unsupported type %q", aux)
	}
	*t = Type(aux)
	return nil
}

const (
	Bool    Type = "bool"
	String  Type = "string"
	Integer Type = "integer"
	Array   Type = "array"
	Object  Type = "object"
	Enum    Type = "enum"
)

var types = map[Type]struct{}{
	Bool:    {},
	String:  {},
	Integer: {},
	Array:   {},
	Object:  {},
	Enum:    {},
}
