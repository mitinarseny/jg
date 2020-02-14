package schema

import (
	"errors"
	"io"
	"strconv"

	"gopkg.in/yaml.v3"
)

type String struct {
	From string `yaml:"from"`
}

func (s *String) UnmarshalYAML(value *yaml.Node) error {
	type rawString String
	var tmp rawString
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	if tmp.From == "" {
		return &yamlError{
			line: value.Line,
			err:  errors.New("from: required"),
		}
	}
	*s = String(tmp)
	return nil
}

func (s *String) GenerateJSON(ctx *Context, w io.Writer) error {
	ss, err := ctx.Rand(s.From)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(strconv.Quote(ss))) // TODO: avoid copying []byte to string
	return err
}
