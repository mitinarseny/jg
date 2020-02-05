package schema

import (
	"bufio"
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
	n := s.Root
	if arrayLen >= 0 {
		n = &Array{
			Length: Length{
				Min: arrayLen,
				Max: arrayLen,
			},
			Elements: s.Root,
		}
	}
	rand.Seed(time.Now().UnixNano())
	bw := bufio.NewWriterSize(w, 1<<20)
	if err := n.Generate(bw); err != nil {
		return err
	}
	return bw.Flush()
}
