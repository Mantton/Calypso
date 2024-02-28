package types

var GlobalScope *Scope

var GlobalTypes = map[BasicType]*Basic{
	Unresolved: {Unresolved, "unresolved"},
	Bool:       {Bool, "bool"},

	// aliased integer types
	Int:  {Int, "int"},
	UInt: {UInt, "uint"},

	// signed integers

	Int8:  {Int8, "i8"},
	Int16: {Int16, "i16"},
	Int32: {Int32, "i32"},
	Int64: {Int64, "i64"},

	// unsigned integers
	UInt8:  {UInt8, "u8"},
	UInt16: {UInt16, "u16"},
	UInt32: {UInt32, "u32"},
	UInt64: {UInt64, "u64"},

	// floating points
	Float:  {Float, "float"},
	Double: {Double, "Double"},

	// strings
	String: {String, "string"},

	// aliases
	Char: {Char, "char"},
	Byte: {Byte, "byte"},

	// misc
	Null: {Null, "null"},
	Void: {Void, "void"},
	Any:  {Any, "any"},

	// group literals
	IntegerLiteral: {IntegerLiteral, "literal int"},
	FloatLiteral:   {FloatLiteral, "literal float"},
}

func init() {
	GlobalScope = NewScope(nil)
	// Define Global Types
	for _, t := range GlobalTypes {
		ok := GlobalScope.Define(NewTypeDef(t.name, t))

		if !ok {
			panic("GLOBAL TYPE ALREADY DEFINED")
		}
	}
}

func LookUp(t BasicType) *Basic {
	v, ok := GlobalTypes[t]

	if !ok {
		return GlobalTypes[Unresolved]
	}

	return v
}
