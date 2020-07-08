package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type File1 struct {
	ss *String
}

type Schema struct {
	Files map[string]*struct{} `yaml:"files"` // pointer because it is
	Root  Node                 `yaml:"root"`
}

func (s *Schema) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Files map[string]*struct{} `yaml:"files"` // pointer because it is
		Root  *node                `yaml:"root"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if aux.Root == nil {
		return &yamlError{
			line: value.Line,
			err:  errors.New("root is not defined"),
		}
	}
	*s = Schema{
		Files: aux.Files,
		Root:  aux.Root.Node,
	}
	return nil
}

func (s *Schema) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand) error {
	if err := s.Root.GenerateJSON(ctx, w, r); err != nil {
		return err
	}
	_, err := w.Write([]byte{'\n'})
	return err
}

func (s *Schema) StreamJSON(ctx *Context, w io.Writer, r *rand.Rand, count int64) error {
	switch count {
	case -1:
		for {
			if err := s.GenerateJSON(ctx, w, r); err != nil {
				return err
			}
		}
	default:
		for i := int64(0); i < count; i++ {
			if err := s.GenerateJSON(ctx, w, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Schema) Validate() error {
	walker, ok := s.Root.(Walker)
	if !ok {
		return nil
	}
	return walker.Walk(func(n Node) (bool, error) {
		str, ok := n.(*String)
		if !ok {
			return true, nil
		}
		if str.From == "" {
			return false, nil
		}
		if _, found := s.Files[str.From]; !found {
			return false, fmt.Errorf("undefined file: %q", str.From)
		}
		return false, nil
	})
}
