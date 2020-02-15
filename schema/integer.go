package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"

	"gopkg.in/yaml.v3"
)

var defaultIntRange = IntRange{
	Min: 0,
	Max: 100,
}

type Integer struct {
	Range   *IntRange `yaml:"range"`
	Choices []int64   `yaml:"choices"`
}

func (i *Integer) UnmarshalYAML(value *yaml.Node) error {
	type raw Integer
	var aux raw
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*i = Integer(aux)
	if i.Range != nil && len(i.Choices) > 0 {
		return &yamlError{
			line: value.Line,
			err:  errors.New("integer should have either range or choices, not both"),
		}
	}
	if i.Range == nil && len(i.Choices) == 0 {
		i.Range = &defaultIntRange
	}
	return nil
}

func (i *Integer) GenerateJSON(_ *Context, w io.Writer) error {
	var num int64
	if i.Range != nil {
		num = i.Range.Rand()
	} else if l := len(i.Choices); l > 0 {
		num = i.Choices[rand.Intn(l)]
	}
	_, err := w.Write([]byte(strconv.FormatInt(num, 10)))
	return err
}

type IntRange struct {
	Min, Max int64
}

func (r IntRange) Rand() int64 {
	return r.Min + rand.Int63n(r.Max-r.Min+1)
}

func (r *IntRange) UnmarshalYAML(value *yaml.Node) error {
	*r = defaultIntRange
	switch value.Kind {
	case yaml.ScalarNode:
		if err := r.unmarshalYAMLScalar(value); err != nil {
			return err
		}
	case yaml.SequenceNode:
		if err := r.unmarshalYAMLSequence(value); err != nil {
			return err
		}
	case yaml.MappingNode:
		if err := r.unmarshalYAMLMapping(value); err != nil {
			return err
		}
	default:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("length should be <int64>, [<int64>, <int64>]" +
				" or {min: <int64>, max: <int64>}, got: %s", value.Tag),
		}
	}
	if err := r.validate(); err != nil {
		return &yamlError{
			line: value.Line,
			err:  err,
		}
	}
	return nil
}

func (r *IntRange) validate() error {
	if r.Min >= r.Max {
		return errors.New("min should be less than max")
	}
	return nil
}

func (r *IntRange) unmarshalYAMLScalar(value *yaml.Node) error {
	return value.Decode(&r.Max)
}

func (r *IntRange) unmarshalYAMLSequence(value *yaml.Node) error {
	var aux [2]int64
	if err := value.Decode(&aux); err != nil {
		return err
	}
	r.Min, r.Max = aux[0], aux[1]
	return nil
}

func (r *IntRange) unmarshalYAMLMapping(value *yaml.Node) error {
	var aux struct {
		Min *int64 `yaml:"min"`
		Max *int64 `yaml:"max"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Min != nil {
		r.Min = *aux.Min
	}
	if aux.Max != nil {
		r.Max = *aux.Max
	}
	return nil
}
