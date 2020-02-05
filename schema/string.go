package schema

type String struct{}

func (s *String) Generate() (interface{}, error) {
	return "example", nil
}
