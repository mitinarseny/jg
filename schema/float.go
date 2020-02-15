package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"

	"gopkg.in/yaml.v3"
)

var defaultFloatRange = FloatRange{
	Min: 0,
	Max: 1,
}

type Float struct {
	Range   *FloatRange `yaml:"range"`
	Choices []float64   `yaml:"choices"`
}

func (f *Float) UnmarshalYAML(value *yaml.Node) error {
	type raw Float
	var aux raw
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*f = Float(aux)
	if f.Range != nil && len(f.Choices) > 0 {
		return &yamlError{
			line: value.Line,
			err:  errors.New("float should have either range or choices, not both"),
		}
	}
	if f.Range == nil && len(f.Choices) == 0 {
		f.Range = &defaultFloatRange
	}
	return nil
}

func (f *Float) GenerateJSON(_ *Context, w io.Writer) error {
	var num float64
	if f.Range != nil {
		num = f.Range.Rand()
	} else if l := len(f.Choices); l > 0 {
		num = f.Choices[rand.Intn(l)]
	}
	_, err := w.Write([]byte(strconv.FormatFloat(num, 'f', -1, 64)))
	return err
}

type FloatRange struct {
	Min, Max float64
}

func (r FloatRange) Rand() float64 {
	if r.Min == r.Max {
		return r.Min
	}
	return r.Min + (r.Max-r.Min)*rand.Float64()
}

func (r *FloatRange) UnmarshalYAML(value *yaml.Node) (err error) {
	*r = defaultFloatRange
	switch value.Kind {
	case yaml.ScalarNode:
		if err := r.unmarshalYAMLScalar(value); err != nil {
			return err
		}
	case yaml.SequenceNode:
		if err := r.unmarshalYAMLSequence(value); err != nil {
			return err
		}
	default:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("length should be {float64 | [float64, float64]}, got: %s", value.Tag),
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

func (r *FloatRange) validate() error {
	if r.Min >= r.Max {
		return errors.New("min should be less than max")
	}
	return nil
}

func (r *FloatRange) unmarshalYAMLScalar(value *yaml.Node) error {
	return value.Decode(&r.Max)
}

func (r *FloatRange) unmarshalYAMLSequence(value *yaml.Node) error {
	var aux [2]float64
	if err := value.Decode(&aux); err != nil {
		return err
	}
	r.Min, r.Max = aux[0], aux[1]
	return nil
}
