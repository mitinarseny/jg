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

func (s *Schema) GenerateJSON(ctx *Context, w io.Writer, r *rand.Rand, length *Length) (err error) {
	if length != nil {
		a := Array{
			Length:   *length,
			Elements: s.Root,
		}
		return a.GenerateJSON(ctx, w, r)
	}
	return s.Root.GenerateJSON(ctx, w, r)
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
