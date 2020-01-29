package schema

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

type Schema struct {
	Files []string `yaml:"files"`
	Root  Node     `yaml:"root"`
}

func (s *Schema) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Files []string        `yaml:"files"`
		Root  map[string]Node `yaml:"root"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*s = Schema{
		Files: aux.Files,
		Root: Node{
			Type:   Object,
			Fields: aux.Root,
		},
	}
	return nil
}

func (s *Schema) Generate(w io.Writer, arrayLen int) error {
	encoder := json.NewEncoder(w)
	if arrayLen >= 0 {
		array := make([]interface{}, 0, arrayLen)
		for i := 0; i < arrayLen; i++ {
			array = append(array, s.Root.Generate())
		}
		return encoder.Encode(array)
	}
	return json.NewEncoder(w).Encode(s.Root.Generate())
}
