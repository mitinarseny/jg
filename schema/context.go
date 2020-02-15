package schema

import (
	"fmt"
)

type Context struct {
	sortKeys bool
	files    map[string]*file
}

func NewContext() *Context {
	return &Context{
		files: make(map[string]*file),
	}
}

func (c *Context) SetSortKeys(b bool) {
	c.sortKeys = b
}

func (c *Context) SortKeys() bool {
	return c.sortKeys
}

func (c *Context) AddFile(name string, file string) error {
	f, err := ScanFile(file)
	if err != nil {
		return err
	}
	c.files[name] = f
	return nil
}

func (c *Context) Rand(name string) ([]byte, error) {
	f, ok := c.files[name]
	if !ok {
		return nil, fmt.Errorf("unknown file %q", name)
	}
	return f.Rand()
}

func (c *Context) Close() error {
	var errs Errors
	for _, f := range c.files {
		if err := f.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[1]
	default:
		return errs
	}
}
