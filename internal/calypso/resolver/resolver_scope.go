package resolver

type Scope struct {
	data map[string]State
}

type State byte

const (
	DECLARED State = iota
	DEFINED
)

func NewScope() *Scope {
	return &Scope{
		data: make(map[string]State),
	}
}

func (s *Scope) Declare(ident string) {
	s.data[ident] = DECLARED
}

func (s *Scope) Define(ident string) {
	s.data[ident] = DEFINED
}

func (s *Scope) Has(ident string) bool {
	sc, ok := s.Get(ident)

	if ok && sc == DECLARED {
		panic("declared but not defined")
	}

	return ok
}

func (s *Scope) Get(ident string) (State, bool) {
	v, ok := s.data[ident]
	return v, ok
}