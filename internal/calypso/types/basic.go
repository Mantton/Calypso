package types

type Basic struct {
	Literal int
}

func (s *Basic) ssaType() {}
