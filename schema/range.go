package schema

import (
	"errors"
	"fmt"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type Length struct {
	Min, Max int
}

func (l Length) Rand() int {
	if l.Min == l.Max {
		return l.Min
	}
	return l.Min + rand.Intn(l.Max-l.Min+1)
}

func (l *Length) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if err := value.Decode(&l.Max); err != nil {
			return err
		}
		l.Min = l.Max
		return nil
	case yaml.SequenceNode:
		var aux [2]int
		if err := value.Decode(&aux); err != nil {
			return err
		}
		l.Min, l.Max = aux[0], aux[1]
	case yaml.MappingNode:
		var aux struct {
			Min *int `yaml:"min"`
			Max *int `yaml:"max"`
		}
		if err := value.Decode(&aux); err != nil {
			return err
		}
		if aux.Min != nil {
			l.Min = *aux.Min
		}
		if aux.Max != nil {
			l.Max = *aux.Max
		}
	default:
		return fmt.Errorf("length should be scalar, sequence or mapping, got: %s", value.Tag)
	}
	if l.Min > l.Max {
		return errors.New("min should be less than or equal to max")
	}
	return nil
}

type IntRange struct {
	Min, Max int64
}

func (r IntRange) Rand() int64 {
	return r.Min + rand.Int63n(r.Max-r.Min+1)
}

func (r *IntRange) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		var aux [2]int64
		if err := value.Decode(&aux); err != nil {
			return err
		}
		r.Min, r.Max = aux[0], aux[1]
	case yaml.MappingNode:
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
	default:
		return fmt.Errorf("range should be either sequence or mapping, got: %s", value.Tag)
	}
	if r.Min >= r.Max {
		return errors.New("min should be less than max")
	}
	return nil
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

func (r *FloatRange) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		var aux [2]float64
		if err := value.Decode(&aux); err != nil {
			return err
		}
		r.Min, r.Max = aux[0], aux[1]
	case yaml.MappingNode:
		var aux struct {
			Min *float64 `yaml:"min"`
			Max *float64 `yaml:"max"`
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
	default:
		return fmt.Errorf("range should be either sequence or mapping, got: %s", value.Tag)
	}
	if r.Min >= r.Max {
		return errors.New("min should be less than max")
	}
	return nil
}
