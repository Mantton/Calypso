package typechecker

type Scope struct {
	data map[string]ExpressionType
}

func NewScope() *Scope {
	return &Scope{
		data: make(map[string]ExpressionType),
	}
}

func (s *Scope) Define(ident string, t ExpressionType) {
	s.data[ident] = t
}

func (s *Scope) Has(ident string) bool {
	_, ok := s.Get(ident)

	return ok
}

func (s *Scope) Get(ident string) (ExpressionType, bool) {
	v, ok := s.data[ident]
	return v, ok
}
