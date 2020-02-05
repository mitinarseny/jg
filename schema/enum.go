package schema

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"

	"gopkg.in/yaml.v3"
)

type Enum []interface{}

func (e *Enum) UnmarshalYAML(value *yaml.Node) error {
	var aux struct {
		Choices []interface{} `yaml:"choices"`
	}
	if err := value.Decode(&aux); err != nil {
		return err
	}
	if len(aux.Choices) < 2 {
		return errors.New("not enough choices, should be >= 2")
	}
	*e = aux.Choices
	return nil
}

func (e Enum) Generate(w *bufio.Writer) error {
	return json.NewEncoder(w).Encode(e[rand.Intn(len(e))])
}
