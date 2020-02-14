package schema

import (
	"fmt"
)

type Context struct {
	sortKeys bool
	files    map[string]*File
}

func NewContext() *Context {
	return &Context{
		files: make(map[string]*File),
	}
}

func (c *Context) SetSortKeys(b bool) {
	c.sortKeys = b
}

func (c *Context) SortKeys() bool {
	return c.sortKeys
}

func (c *Context) AddFile(name string, file *File) error {
	c.files[name] = file
	return nil
}

func (c *Context) Rand(name string) (string, error) {
	f, ok := c.files[name]
	if !ok {
		return "", fmt.Errorf("unknown file %q", name)
	}
	if !f.Scanned() {
		if err := f.Scan(); err != nil {
			return "", fmt.Errorf("unable to scan: %w", err)
		}
	}
	return f.Rand()
}

func (c *Context) Close() error {
	var errs Errors
	for _, f := range c.files {
		if f.Scanned() {
			if err := f.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs.CheckLen()
}
