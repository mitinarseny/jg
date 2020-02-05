package schema

import "bufio"

type String struct{}

func (s *String) Generate(w *bufio.Writer) error {
	_, err := w.WriteString("example")
	return err
}
