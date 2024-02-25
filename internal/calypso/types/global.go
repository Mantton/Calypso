package types

var GlobalScope *Scope

var GlobalTypes = map[BasicType]*Basic{
	Unresolved: {Unresolved, "unresolved"},
	Bool:       {Bool, "bool"},
	Int:        {Int, "int"},
	Float:      {Float, "float"},
	String:     {String, "string"},
	Null:       {Null, "null"},
	Void:       {Void, "void"},
	Any:        {Any, "any"},
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
