package schema

import (
	"errors"
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Array struct {
	Length   Length
	Elements Node
}

func (a *Array) UnmarshalYAML(value *yaml.Node) error {
	aux := struct {
		Length   Length `yaml:"length"`
		Elements *node  `yaml:"elements"`
	}{
		Length: Length{
			Min: 0,
			Max: 10,
		},
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Elements == nil {
		return &yamlError{
			line: value.Line,
			err:  errors.New("elements: required"),
		}
	}
	*a = Array{
		Length:   aux.Length,
		Elements: aux.Elements.Node,
	}
	return nil
}

func (a *Array) GenerateJSON(ctx *Context, w io.Writer) error {
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}
	elNum := a.Length.Rand()
	for i := uint64(0); i < elNum; i++ {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := a.Elements.GenerateJSON(ctx, w); err != nil {
			return a.wrapIndexErr(i, err)
		}
	}
	_, err := w.Write([]byte{']'})
	return err
}

func (a *Array) Walk(fn WalkFn) error {
	var errs Errors
	proceed, err := fn(a.Elements)
	if err != nil {
		errs = append(errs, err)
	}
	if !proceed {
		return errs.CheckLen()
	}
	walker, ok := a.Elements.(Walker)
	if !ok {
		return errs.CheckLen()
	}
	if err := walker.Walk(fn); err != nil {
		errs = append(errs, a.wrapErr(err))
	}
	return errs.CheckLen()
}

func (a *Array) wrapErr(err error) error {
	return WrapErr("[]", err)
}

func (a *Array) wrapIndexErr(ind uint64, err error) error {
	return WrapErr("["+strconv.FormatUint(ind, 10)+"]", err)
}
