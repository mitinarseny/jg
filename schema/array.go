package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var defaultArrayLength = Length{
	Min: 0,
	Max: 10,
}

type Array struct {
	Length   Length
	Elements Node
}

func (a *Array) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Length   Length `yaml:"length"`
		Elements *node  `yaml:"elements"`
	}{
		Length: defaultArrayLength,
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Elements == nil {
		return &yamlError{
			line: value.Line,
			err:  errors.New("\"elements\" is required"),
		}
	}
	*a = Array{
		Length:   aux.Length,
		Elements: aux.Elements.Node,
	}
	return nil
}

func (a *Array) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand) error {
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}
	elNum := a.Length.Rand(r)
	for i := uint64(0); i < elNum; i++ {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := a.Elements.GenerateJSON(ctx, w, r); err != nil {
			return a.wrapIndexErr(i, err)
		}
	}
	_, err := w.Write([]byte{']'})
	return err
}

func (a *Array) Walk(fn WalkFn) (err error) {
	var errs Errors
	defer func() {
		switch len(errs) {
		case 0:
			err = nil
		case 1:
			err = errs[0]
		default:
			err = errs
		}
	}()
	proceed, err := fn(a.Elements)
	if err != nil {
		errs = append(errs, err)
	}
	if !proceed {
		return
	}
	walker, ok := a.Elements.(Walker)
	if !ok {
		return
	}
	if err := walker.Walk(fn); err != nil {
		errs = append(errs, a.wrapErr(err))
	}
	return
}

func (a *Array) wrapErr(err error) error {
	return WrapErr("[]", err)
}

func (a *Array) wrapIndexErr(ind uint64, err error) error {
	return WrapErr("["+strconv.FormatUint(ind, 10)+"]", err)
}

type Length struct {
	Min, Max uint64
}

func (l *Length) Rand(r *rand.Rand) uint64 {
	if l.Min == l.Max {
		return l.Min
	}
	return l.Min + r.Uint64()%(l.Max-l.Min+1)
}

func (l *Length) Set(s string) error {
	var err error
	switch ss := strings.Split(s, ","); len(ss) {
	case 1:
		l.Max, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as uint: %w", s, err)
		}
		l.Min = l.Max
		return nil
	case 2:
		l.Min, err = strconv.ParseUint(ss[0], 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as uint: %w", s, err)
		}
		l.Max, err = strconv.ParseUint(ss[1], 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse %q as uint: %w", s, err)
		}
		return l.validate()
	default:
		return fmt.Errorf("length should be [int,]int")
	}
}

func (l *Length) Type() string {
	return "[min,]max"
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
		return l.unmarshalYAMLScalar(value)
	case yaml.SequenceNode:
		if err := l.unmarshalYAMLSequence(value); err != nil {
			return err
		}
	default:
		return &yamlError{
			line: value.Line,
			err:  fmt.Errorf("length should be {uint64 | [uint64, uint64]} got: %s", value.Tag),
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
	return nil
}

func (l *Length) unmarshalYAMLScalar(value *yaml.Node) error {
	if err := value.Decode(&l.Max); err != nil {
		return err
	}
	l.Min = l.Max
	return nil
}

func (l *Length) unmarshalYAMLSequence(value *yaml.Node) error {
	var aux [2]uint64
	if err := value.Decode(&aux); err != nil {
		return err
	}
	l.Min, l.Max = aux[0], aux[1]
	return nil
}
