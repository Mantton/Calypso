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
