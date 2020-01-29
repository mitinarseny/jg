package schema

import (
	"encoding/json"
	"io"
)

type Schema struct {
	Files []string `yaml:"files"`
	Root  Node     `yaml:"root"`
}

func (s *Schema) Generate(w io.Writer) error {
	return json.NewEncoder(w).Encode(s.Root.Generate())
}
