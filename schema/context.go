package schema

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

func (c *Context) File(name string) *File {
	return c.files[name]
}
