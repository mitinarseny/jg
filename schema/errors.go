package schema

import (
	"fmt"
	"strings"
)

type PathError struct {
	path string
	err  error
}

func (n *PathError) Error() string {
	return fmt.Sprintf("%s: %s", n.path, n.err)
}

func WrapErr(name string, err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*PathError); ok {
		return &PathError{
			path: name + e.path,
			err:  e.err,
		}
	}
	return &PathError{
		path: name,
		err:  err,
	}
}

type Errors []error

func (e *Errors) Add(err error) *Errors {
	if err == nil {
		return e
	}
	if e == nil {
		*e = make(Errors, 0, 1)
	}
	*e = append(*e, err)
	return e
}

func (e Errors) Error() string {
	var b strings.Builder
	for i, err := range e {
		if i != 0 {
			b.WriteString("; ")
		}
		fmt.Fprint(&b, err)
	}
	return b.String()
}

func (e Errors) Err() error {
	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	default:
		return e
	}
}

type yamlError struct {
	line int
	err  error
}

func (e *yamlError) Error() string {
	return fmt.Sprintf("line %d: %s", e.line, e.err)
}
