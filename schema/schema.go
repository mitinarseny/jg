package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"gopkg.in/yaml.v3"
)

type Schema struct {
	Files []string `yaml:"files"`
	Root  Node     `yaml:"root"`
}

func (s *Schema) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Files []string `yaml:"files"`
		Root  nodeMap  `yaml:"root"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*s = Schema{
		Files: aux.Files,
		Root:  node{Node: Object(aux.Root)},
	}
	return nil
}

func (s *Schema) Generate(w io.Writer, arrayLen int) error {
	encoder := json.NewEncoder(w)
	n := s.Root
	if arrayLen >= 0 {
		n = &Array{
			Range: IntRange{
				Min: arrayLen,
				Max: arrayLen,
			},
			Elements: s.Root,
		}
	}
	rand.Seed(time.Now().UnixNano())
	gen, err := n.Generate()
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	return encoder.Encode(gen)
}
