package schema

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type File1 struct {
	ss *String
}

type Schema struct {
	Files map[string]*struct{} `yaml:"files"` // pointer because it is
	Root  Object               `yaml:"root"`
}

func (s *Schema) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Files map[string]*struct{} `yaml:"files"`
		Root  nodeMap              `yaml:"root"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*s = Schema{
		Files: aux.Files,
		Root:  Object(aux.Root),
	}
	return nil
}

func (s *Schema) GenerateJSON(ctx *Context, w io.Writer, length *Length) (err error) {
	if length != nil {
		a := Array{
			Length:   *length,
			Elements: s.Root,
		}
		return a.GenerateJSON(ctx, w)
	}
	return s.Root.GenerateJSON(ctx, w)
}

func (s *Schema) Validate() error {
	return s.Root.Walk(func(n Node) (bool, error) {
		str, ok := n.(*String)
		if !ok {
			return true, nil
		}
		if _, found := s.Files[str.From]; !found {
			return false, fmt.Errorf("undefined file: %q", str.From)
		}
		return false, nil
	})
}
