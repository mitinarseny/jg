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

type yamlError struct {
	line int
	err error
}

func (e *yamlError) Error() string {
	return fmt.Sprintf("line %d: %s", e.line, e.err)
}