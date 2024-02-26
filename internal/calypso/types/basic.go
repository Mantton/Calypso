package types

type BasicType byte

const (
	Unresolved BasicType = iota
	Bool
	Int // Either 32 or 64
	Int8
	Int16
	Int32
	Int64
	// unsigned integers
	UInt // Either 32 or 64
	UInt8
	UInt16
	UInt32
	UInt64

	// floating point
	Float
	Double

	// string
	String

	// helpful aliases
	Char // alias for uint32
	Byte // alias for uint8

	// misc
	Null
	Void
	Any

	// helper literals
	IntegerLiteral
	FloatingPointLiteral
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
		case Int, Int8, Int16, Int32, Int64:
			return true
		case UInt, UInt8, UInt16, UInt32, UInt64:
			return true
		case Char, Byte:
			return true
		case Float, Double:
			return true
		}
	}
	return false
}

func IsFloatingPoint(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case Float, Double:
			return true
		}
	}
	return false
}

func IsUnsigned(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case UInt8, UInt16, UInt32, UInt64:
			return true
		case Char, Byte:
			return true
		}

	}
	return false
}

// group literals are literals that can describe multiple types. e.g 100 can be i64, i32, i16
func IsGroupLiteral(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case IntegerLiteral, FloatingPointLiteral:
			return true
		}

	}
	return false
}

func IsEquatable(t Type) bool {
	return t == LookUp(Bool) || IsNumeric(t)
}
