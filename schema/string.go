package schema

import (
	"fmt"
	"io"
	"strconv"
)

type String struct {
	From string `yaml:"from"`
}

func (s *String) GenerateJSON(ctx *Context, w io.Writer) error {
	f := ctx.File(s.From)
	if f == nil {
		return fmt.Errorf("unknown file %q", s.From)
	}
	if !f.Scanned {
		if err := f.Scan(); err != nil {
			return fmt.Errorf("unable to scan %q: %w", f.Path, err)
		}
	}
	ss, err := f.Rand()
	if err != nil {
		return fmt.Errorf("unable to generate from file: %w", err)
	}
	_, err = w.Write([]byte(strconv.Quote(ss))) // TODO: avoid copying []byte to string
	return err
}
