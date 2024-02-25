package types

type BasicType byte

const (
	Unresolved BasicType = iota
	Bool
	Int
	Float
	String

	Null
	Void
	Any
)

type Basic struct {
	Literal BasicType
	name    string
}

func (t *Basic) clyT()          {}
func (t *Basic) Name() string   { return t.name }
func (t *Basic) String() string { return t.name }

func IsNumeric(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case Int, Float:
			return true
		}
	}
	return false
}

func IsEquatable(t Type) bool {
	return t == LookUp(Bool) || IsNumeric(t)
}
