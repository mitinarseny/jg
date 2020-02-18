package schema

import (
	"errors"
	"io"
	"math/rand"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Object struct {
	Fields     map[string]Node
	sortedKeys []string
}

func (o *Object) sortKeys() {
	o.sortedKeys = make([]string, 0, len(o.Fields))
	for field := range o.Fields {
		o.sortedKeys = append(o.sortedKeys, field)
	}
	sort.Strings(o.sortedKeys)
}

func (o *Object) sorted() bool {
	return o.sortedKeys != nil
}

func (o *Object) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Fields nodeMap `yaml:"fields"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Fields == nil {
		return &yamlError{
			line: value.Line,
			err:  errors.New("\"fields\" is required"),
		}
	}
	*o = Object{
		Fields: aux.Fields,
	}
	return nil
}

func (o *Object) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand) error {
	if _, err := w.Write([]byte{'{'}); err != nil {
		return err
	}
	// TODO: sort when creating
	if ctx.SortKeys() {
		if !o.sorted() {
			o.sortKeys()
		}
		for i, key := range o.sortedKeys {
			if i > 0 {
				if _, err := w.Write([]byte{','}); err != nil {
					return err
				}
			}
			node := o.Fields[key]
			if err := o.writeField(ctx, w, r, key, node, ); err != nil {
				return o.wrapErr(key, err)
			}
		}
	} else {
		var wasFirst bool
		for field, node := range o.Fields {
			if wasFirst {
				if _, err := w.Write([]byte{','}); err != nil {
					return err
				}
			}
			wasFirst = true
			if err := o.writeField(ctx, w, r, field, node); err != nil {
				return o.wrapErr(field, err)
			}
		}
	}
	_, err := w.Write([]byte{'}'})
	return err
}

func (o *Object) writeField(ctx *Context, w io.Writer, r *rand.Rand, field string, node Node) error {
	if _, err := w.Write(append(strconv.AppendQuote(make([]byte, 0, 3*len(field)/2), field), ':')); err != nil {
		return err
	}
	return node.GenerateJSON(ctx, w, r)
}

func (o *Object) Walk(fn WalkFn) error {
	var errs Errors
	for k, n := range o.Fields {
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
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errs
	}
}

func (o *Object) wrapErr(field string, err error) error {
	return WrapErr("."+field, err)
}
