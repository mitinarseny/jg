package schema

import (
	"errors"
	"fmt"
	"io"
	"math/rand"

	"gopkg.in/yaml.v3"
)

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
	// TODO: pass context.Context to write last row properly
	for count != 0 {
		if err := s.GenerateJSON(ctx, w, r); err != nil {
			return err
		}
		if count > 0 {
			count--
		}
	}
	return nil
}

func (s *Schema) Validate() error {
	return Walk(s.Root, func(n Node) (bool, error) {
		if str, ok := n.(*String); ok {
			if sf, ok := str.StringRander.(StringFile); ok {
				fn := sf.Filename()
				if _, found := s.Files[fn]; !found {
					return false, fmt.Errorf("undefined file: %q", fn)
				}
			}
		}
		return true, nil
	})
}
