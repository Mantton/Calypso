package types

import (
	"fmt"
	"strings"
)

type Struct struct {
	Fields []*Var
	Map    map[string]*Var
}

func (t *Struct) Parent() Type { return t }

func (t *Struct) String() string {
	a := "struct { %s }"

	b := []string{}

	for _, c := range t.Fields {

		d := "%s : %s;"
		e := fmt.Sprintf(d, c.name, c.typ)
		b = append(b, e)
	}

	return fmt.Sprintf(a, strings.Join(b, " "))

}

func NewStruct(f []*Var) *Struct {
	s := &Struct{
		Fields: nil,
		Map:    make(map[string]*Var),
	}

	for i, e := range f {
		e.StructIndex = i
		s.Fields = append(s.Fields, e)
		s.Map[e.name] = e
	}
	return s
}

func IsStruct(t Type) bool {
	_, ok := t.(*Struct)

	return ok
}

func (s *Struct) FindField(n string) *Var {
	v, ok := s.Map[n]

	if ok {
		return v
	}

	return nil
}

func GetFieldIndex(n string, t Type) int {
	switch t := t.Parent().(type) {
	case *Struct:
		v := t.FindField(n)
		if v != nil {
			return v.StructIndex
		}
	}

	return -1
}
