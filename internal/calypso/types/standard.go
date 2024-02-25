package types

type Standard struct {
	methods map[string]*Function
}

func NewStandard() *Standard {
	return &Standard{
		methods: make(map[string]*Function),
	}
}

func (s *Standard) clyT()          {}
func (s *Standard) String() string { return "" }

func (s *Standard) AddMethod(f *Function) bool {
	_, ok := s.methods[f.name]

	if ok {
		return false
	}

	s.methods[f.name] = f
	return true
}
