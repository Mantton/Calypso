package types

type Standard struct {
	methods map[string]*FunctionSignature
}

func NewStandard() *Standard {
	return &Standard{
		methods: make(map[string]*FunctionSignature),
	}
}

func (s *Standard) clyT()        {}
func (t *Standard) Parent() Type { return t }

func (s *Standard) String() string { return "" }

func (s *Standard) AddMethod(n string, f *FunctionSignature) bool {
	_, ok := s.methods[n]

	if ok {
		return false
	}

	s.methods[n] = f
	return true
}
