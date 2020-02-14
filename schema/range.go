package schema

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Length struct {
	Min, Max uint64
}

func (l Length) Rand() uint64 {
	if l.Min == l.Max {
		return l.Min
	}
	return l.Min + rand.Uint64()%(l.Max-l.Min+1)
}

func (l *Length) Set(s string) error {
	var (
		min, max uint64
		err      error
	)
	switch ss := strings.Split(s, ","); len(ss) {
	case 1:
		min, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as int: %w", s, err)
		}
		max = min
	case 2:
		min, err = strconv.ParseUint(ss[0], 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as int: %w", s, err)
		}
		max, err = strconv.ParseUint(ss[1], 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as int: %w", s, err)
		}
	default:
		return fmt.Errorf("length should be int[,int]")
	}
	*l = Length{
		Min: min,
		Max: max,
	}
	return l.validate()
}

func (l *Length) Type() string {
	return "uint64[,uint64]"
}

func (l *Length) String() string {
	if l.Min == l.Max {
		return strconv.FormatUint(l.Max, 10)
	}
	return fmt.Sprintf("%d,%d", l.Min, l.Max)
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
		var aux [2]uint64
		if err := value.Decode(&aux); err != nil {
			return err
		}
		l.Min, l.Max = aux[0], aux[1]
	case yaml.MappingNode:
		var aux struct {
			Min *uint64 `yaml:"min"`
			Max *uint64 `yaml:"max"`
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
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("length should be scalar, sequence or mapping, got: %s", value.Tag),
		}
	}
	if err := l.validate(); err != nil {
		return &yamlError{
			line: value.Line,
			err:  err,
		}
	}
	return nil
}

func (l *Length) validate() error {
	if l.Min > l.Max {
		return errors.New("min should be less than or equal to max")
	}
	if l.Min < 0 || l.Max < 0 {
		return errors.New("length should be equal to or greater than zero")
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
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("range should be either sequence or mapping, got: %s", value.Tag),
		}
	}
	if r.Min >= r.Max {
		return &yamlError{
			line: value.Line,
			err:  errors.New("min should be less than max"),
		}
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
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("range should be either sequence or mapping, got: %s", value.Tag),
		}
	}
	if r.Min >= r.Max {
		return &yamlError{
			line: value.Line,
			err:  errors.New("min should be less than max"),
		}
	}
	return nil
}
