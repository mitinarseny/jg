package schema

import (
	"errors"
	"io"
	"math/rand"
	"strconv"

	"gopkg.in/yaml.v3"
)

type StringRander interface {
	Rand(ctx *Context, r *rand.Rand) ([]byte, error)
}

type StringChoices []string

func (c StringChoices) Rand(_ *Context, r *rand.Rand) ([]byte, error) {
	return []byte(c[r.Intn(len(c))]), nil
}

type StringFile string

func (f StringFile) Filename() string {
	return string(f)
}

func (f StringFile) Rand(ctx *Context, r *rand.Rand) ([]byte, error) {
	return ctx.Rand(r, string(f))
}

type String struct {
	StringRander
}

func (s *String) UnmarshalYAML(value *yaml.Node) error {
	var tmp struct {
		From    string   `yaml:"from"`
		Choices []string `yaml:"choices"`
	}
	if err := value.Decode(&tmp); err != nil {
		return err
	}

	if !trueOnlyOne(tmp.From == "", len(tmp.Choices) == 0) {
		return &yamlError{
			line: value.Line,
			err:  errors.New("string should have either from or choices"),
		}
	}

	switch {
	case tmp.From != "":
		s.StringRander = StringFile(tmp.From)
	case len(tmp.Choices) != 0:
		s.StringRander = StringChoices(tmp.Choices)
	}
	return nil
}

func (s *String) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand) error {
	str, err := s.StringRander.Rand(ctx, r)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(strconv.Quote(string(str))))
	return err
}

func trueOnlyOne(bs ...bool) bool {
	var was bool
	for _, b := range bs {
		if b {
			if was {
				return false
			}
			was = true
		}
	}
	return was
}
