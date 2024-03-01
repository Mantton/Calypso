package types

type Standard struct {
	Name string
	Dna  map[string]*Function // core functions to implement
	// possible default methods
	// other methods that this standard shares
}

func NewStandard(name string) *Standard {
	return &Standard{
		Name: name,
		Dna:  make(map[string]*Function),
	}
}

func (s *Standard) clyT()        {}
func (t *Standard) Parent() Type { return t }

func (s *Standard) String() string { return s.Name }

func (s *Standard) AddMethod(n string, f *Function) bool {
	_, ok := s.Dna[n]

	if ok {
		return false
	}

	s.Dna[n] = f
	return true
}
