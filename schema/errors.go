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

func (e Errors) CheckLen() error {
	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	default:
		return e
	}
}