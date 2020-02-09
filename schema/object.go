package schema

import (
	"errors"
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Object map[string]Node

func (o *Object) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Fields nodeMap `yaml:"fields"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Fields == nil {
		return errors.New("object must specify its fields")
	}
	*o = Object(aux.Fields)
	return nil
}

func (o Object) GenerateJSON(ctx *Context, w io.Writer) error {
	if ctx.SortKeys() {
		// TODO: sort keys
	}
	if _, err := w.Write([]byte{'{'}); err != nil {
		return err
	}
	var wasFirst bool
	for field, node := range o {
		if wasFirst {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		wasFirst = true
		if _, err := w.Write([]byte(strconv.Quote(field) + ":")); err != nil {
			return err
		}
		if err := node.GenerateJSON(ctx, w); err != nil {
			return o.wrapErr(field, err)
		}
	}
	_, err := w.Write([]byte{'}'})
	return err
}

func (o Object) Walk(fn WalkFn) error {
	var errs Errors
	for k, n := range o {
		proceed, err := fn(n)
		if err != nil {
			errs = append(errs, o.wrapErr(k, err))
		}
		if !proceed {
			continue
		}
		walker, ok := n.(Walker)
		if !ok {
			continue
		}
		if err := walker.Walk(fn); err != nil {
			errs = append(errs, o.wrapErr(k, err))
		}
	}
	return errs.CheckLen()
}

func (o Object) wrapErr(field string, err error) error {
	return WrapErr("."+field, err)
}
