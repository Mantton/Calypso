package typechecker

type ExpressionType interface {
	Ident() string
	String() string
}

// Base Literal Type
type BaseType struct {
	name string
}

func (t *BaseType) String() string {
	return t.name
}
func (t *BaseType) Ident() string {
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
	Params []ExpressionType
}

func (t *GenericType) Ident() string {
	return t.name
}

func (t *GenericType) String() string {

	out := t.name
	out += "<"

	for i, e := range t.Params {
		out += e.String()

		if i != len(t.Params)-1 {
			out += ", "
		}
	}

	out += ">"
	return out
}

func GenerateGenericType(name string, params ...ExpressionType) *GenericType {
	return &GenericType{
		name:   name,
		Params: params,
	}
}
