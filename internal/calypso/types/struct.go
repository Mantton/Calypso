package types

import (
	"fmt"
	"strings"
)

type Struct struct {
	Fields []*Var
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
	return &Struct{
		Fields: f,
	}
}

func IsStruct(t Type) bool {
	_, ok := t.(*Struct)

	return ok
}

func (s *Struct) FindField(n string) *Var {
	for _, v := range s.Fields {
		if v.name == n {
			return v
		}
	}

	return nil
}
