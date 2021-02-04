package schema

import (
	"fmt"
	"io"
	"math/rand"
)

type Context struct {
	sortKeys bool
	files    map[string]LineSource
}

func NewContext() *Context {
	return &Context{
		files: make(map[string]LineSource),
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

func (c *Context) Rand(r *rand.Rand, name string) ([]byte, error) {
	f, ok := c.files[name]
	if !ok {
		return nil, fmt.Errorf("unknown file %q", name)
	}
	return f.Rand(r)
}

func (c *Context) Close() error {
	var errs Errors
	for _, f := range c.files {
		if cl, ok := f.(io.Closer); ok {
			errs.Add(cl.Close())
		}
	}
	return errs.Err()
}
