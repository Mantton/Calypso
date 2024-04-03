package types

import (
	"fmt"
	"strings"
)

type Struct struct {
	Fields map[string]*Var
}

func (t *Struct) clyT()        {}
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
		Fields: make(map[string]*Var),
	}

	for i, e := range f {
		e.StructIndex = i
		s.Fields[e.name] = e
	}
	return s
}

func IsStruct(t Type) bool {
	_, ok := t.(*Struct)

	return ok
}

func (s *Struct) FindField(n string) *Var {
	v, ok := s.Fields[n]

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
