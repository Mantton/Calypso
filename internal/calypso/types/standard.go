package types

import "fmt"

type Standard struct {
	Name      string
	Signature map[string]*Function
	Types     map[string]*Alias

	// TODO: Default Signature Methods
}

func NewStandard(name string) *Standard {
	return &Standard{
		Name:      name,
		Signature: make(map[string]*Function),
		Types:     make(map[string]*Alias),
	}
}

func (t *Standard) Parent() Type { return t }

func (s *Standard) String() string { return s.Name }

func (s *Standard) AddMethod(n string, f *Function) bool {
	_, ok := s.Signature[n]

	if ok {
		return false
	}

	s.Signature[n] = f
	return true
}

func (s *Standard) AddType(t *Alias) error {

	_, ok := s.Types[t.String()]

	if ok {
		return fmt.Errorf("type \"%s\" already exists in standard \"%s\"", t.name, s.Name)
	}

	s.Types[t.name] = t
	return nil
}

func AsStandard(t Type) *Standard {

	if t == nil {
		return nil
	}

	if a, ok := t.(*Standard); ok {
		return a
	}
	return nil

}
