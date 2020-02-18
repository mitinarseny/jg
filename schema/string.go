package schema

import (
	"errors"
	"io"
	"math/rand"
	"strconv"

	"gopkg.in/yaml.v3"
)

type String struct {
	From    string   `yaml:"from"`
	Choices []string `yaml:"choices"`
}

func (s *String) UnmarshalYAML(value *yaml.Node) error {
	type rawString String
	var tmp rawString
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	*s = String(tmp)
	if (s.From == "") == (len(s.Choices) == 0) {
		return &yamlError{
			line: value.Line,
			err:  errors.New("string should have either from or choices"),
		}
	}
	*s = String(tmp)
	return nil
}

func (s *String) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand) error {
	var str string
	if s.From != "" {
		ss, err := ctx.Rand(r, s.From)
		if err != nil {
			return err
		}
		str = string(ss)
	} else if l := len(s.Choices); l > 0 {
		str = s.Choices[r.Intn(l)]
	}

	_, err := w.Write(strconv.AppendQuote(make([]byte, 0, 3*len(str)/2), str))
	return err
}
