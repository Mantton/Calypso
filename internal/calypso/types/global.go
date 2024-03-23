package types

import "log"

var GlobalScope *Scope

var globalTypes = map[BasicType]*Basic{
	Unresolved:  {Unresolved, "unresolved type"},
	Placeholder: {Placeholder, "unspecialized placeholder"},
	Bool:        {Bool, "bool"},

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
	Double: {Double, "double"},

	// strings
	String: {String, "string"},

	// aliases
	Char: {Char, "char"},
	Byte: {Byte, "byte"},

	// misc
	Void: {Void, "void"},

	// group literals
	IntegerLiteral: {IntegerLiteral, "literal int"},
	FloatLiteral:   {FloatLiteral, "literal float"},
	NilLiteral:     {NilLiteral, "literal nil"},
}

func init() {
	GlobalScope = NewScope(nil)
	// Define Global Types
	for _, t := range globalTypes {
		s := NewScope(GlobalScope)
		d := NewDefinedType(t.name, t, nil, s)
		err := GlobalScope.Define(d)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func LookUp(t BasicType) Type {
	v, ok := globalTypes[t]

	if !ok {
		return GlobalScope.MustResolve(globalTypes[Unresolved].name).Type()
	}

	return GlobalScope.MustResolve(v.name).Type()
}
