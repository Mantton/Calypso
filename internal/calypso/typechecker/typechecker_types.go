package typechecker

type ExpressionType interface {
	Name() string
}

// Base Literal Type
type BaseType struct {
	name string
}

func (t *BaseType) Name() string {
	return t.name
}

func GenerateBaseType(name string) *BaseType {
	return &BaseType{
		name: name,
	}
}

type BaseLiteral byte

const (
	INT BaseLiteral = iota
	FLOAT
	STRING
	NULL
	VOID
	ANY
)

// Generic Type
type GenericType struct {
	name   string
	params []ExpressionType
}

func (t *GenericType) Name() string {

	out := t.name
	out += "<"

	for i, e := range t.params {
		out += e.Name()

		if i != len(t.params) {
			out += ","
		}
	}

	out += ">"
	return out
}

var builtin = map[BaseLiteral]ExpressionType{}
