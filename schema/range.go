package schema

import (
	"errors"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type IntRange struct {
	Min, Max int
}

func (r IntRange) Rand() int {
	if r.Min == r.Max {
		return r.Min
	}
	return r.Min + rand.Intn(r.Max-r.Min+1)
}

func (r *IntRange) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if err := value.Decode(&r.Max); err != nil {
			return err
		}
		r.Min = r.Max
		return nil
	case yaml.SequenceNode:
		var aux [2]int
		if err := value.Decode(&aux); err != nil {
			return err
		}
		r.Min, r.Max = aux[0], aux[1]
	case yaml.MappingNode:
		var aux struct {
			Min *int `yaml:"min"`
			Max *int `yaml:"max"`
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
	}
	if r.Max < r.Min {
		return errors.New("min should be less or equal to max")
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
	case yaml.ScalarNode:
		if err := value.Decode(&r.Max); err != nil {
			return err
		}
		r.Min = r.Max
		return nil
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
	}
	if r.Max < r.Min {
		return errors.New("min should be less or equal to max")
	}
	return nil
}
