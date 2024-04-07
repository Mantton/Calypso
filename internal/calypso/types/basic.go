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
	Void
	Any
	Placeholder

	// helper literals
	IntegerLiteral
	FloatLiteral
	NilLiteral
)

type Basic struct {
	Literal BasicType
	name    string
}

func (t *Basic) clyT()          {}
func (t *Basic) Name() string   { return t.name }
func (t *Basic) String() string { return t.name }
func (t *Basic) Parent() Type   { return t }

func IsNumeric(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case Int, Int8, Int16, Int32, Int64, FloatLiteral, IntegerLiteral:
			return true
		case UInt, UInt8, UInt16, UInt32, UInt64:
			return true
		case Char, Byte:
			return true
		case Float, Double:
			return true
		}
	case *DefinedType:
		return IsNumeric(t.Parent())
	case *Alias:
		return IsNumeric(t.RHS)
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
	case *DefinedType:
		return IsFloatingPoint(t.Parent())
	case *Alias:
		return IsFloatingPoint(t.RHS)
	}
	return false
}

func IsInteger(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case Int, Int8, Int16, Int32, Int64, IntegerLiteral:
			return true
		case UInt, UInt8, UInt16, UInt32, UInt64:
			return true
		default:
			return false
		}
	case *DefinedType:
		return IsInteger(t.Parent())
	case *Alias:
		return IsInteger(t.RHS)
	default:
		return false
	}
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
	case *DefinedType:
		return IsUnsigned(t.Parent())
	case *Alias:
		return IsUnsigned(t.RHS)
	}
	return false
}

// group literals are literals that can describe multiple types. e.g 100 can be i64, i32, i16
func IsGroupLiteral(t Type) bool {
	switch t := t.(type) {
	case *Basic:
		switch t.Literal {
		case IntegerLiteral, FloatLiteral:
			return true
		}
	case *DefinedType:
		return IsGroupLiteral(t.Parent())
	case *Alias:
		return IsGroupLiteral(t.RHS)
	}
	return false
}

func ResolveLiteral(t Type) Type {
	switch t := t.(type) {
	case *DefinedType:

		if t == LookUp(IntegerLiteral) {
			return LookUp(Int)
		} else if t == LookUp(FloatLiteral) {
			return LookUp(Double)
		}

		// No Type Param
		if len(t.TypeParameters) == 0 {
			return t
		}

		for _, p := range t.TypeParameters {
			// Unbounded
			if p.Bound == nil {
				continue
			}

			// Has Bounded Grouped Literal
			if IsGroupLiteral(p.Bound) {
				//
				return resolveDefined(t)
			} else {
				continue
			}

		}

	case *Pointer:
		ptr := ResolveLiteral(t.PointerTo)
		return NewPointer(ptr)
	}

	return t
}

func resolveDefined(t *DefinedType) Type {
	// Recreate mapping
	ctx := make(mappings)
	for _, p := range t.TypeParameters {
		ctx[p.Name()] = ResolveLiteral(p.Unwrapped())
	}

	if t.InstanceOf == nil {
		return Apply(ctx, t)
	} else {
		return Apply(ctx, t.InstanceOf)
	}
}

func IsBoolean(t Type) bool {
	switch t := t.(type) {
	case *DefinedType:
		return t == LookUp(Bool)
	case *Alias:
		return IsBoolean(t.RHS)
	default:
		return false
	}
}
func IsEquatable(t Type) bool {
	basic := IsBoolean(t) || IsNumeric(t) || IsPointer(t)
	if basic {
		return basic
	}

	// enums
	en, ok := t.Parent().(*Enum)

	if !ok {
		return basic
	}

	if !en.IsUnion() {
		return true
	}

	return false
}

func IsConstant(t Type) bool {
	_, ok := t.(*Basic)

	return ok
}

func IsPointer(t Type) bool {
	_, ok := t.(*Pointer)

	return ok
}
