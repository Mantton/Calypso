package resolver

// Symbol Constants
type SymbolType byte

const (
	VariableSymbol SymbolType = iota // Variables
	FunctionSymbol                   // Functions
	StructSymbol                     // Structs
	StandardSymbol                   // Standards
	AliasSymbol                      // Aliases
	TypeSymbol                       // Types
)

// Visibility Contants
type SymbolVisibility byte

// Definition State
type SymbolState byte

const (
	SymbolDeclared SymbolState = iota
	SymbolDefined
)

type SymbolInfo struct {
	Name  string      // name of the symbol
	Type  SymbolType  // The type of symbol being represented
	State SymbolState // the Symbol State
}

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]*SymbolInfo
}

// * Table Methods
// Returns a pointer to a newly created symbol table
func newSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		Parent:  parent,
		Symbols: make(map[string]*SymbolInfo),
	}
}

// Define a new symbol to the symbol table
func (t *SymbolTable) Define(symbol *SymbolInfo) bool {
	s, ok := t.Symbols[symbol.Name]

	// Already Declared in Scope, Define
	if ok && s.State == SymbolDeclared {
		t.Symbols[symbol.Name] = symbol
		symbol.State = SymbolDefined
		return true
	} else {
		// Is not declared
		return false
	}
}

// Declares a new symbol to the symbol table
func (t *SymbolTable) Declare(symbol *SymbolInfo) bool {
	_, ok := t.Symbols[symbol.Name]

	if ok {
		// Already Defined in Scope
		return false
	}
	t.Symbols[symbol.Name] = symbol
	symbol.State = SymbolDeclared
	return true
}

// Resolve searches for a symbol in the current table and parent scopes & returns a pointer to the symbol
func (t *SymbolTable) Resolve(name string) (*SymbolInfo, bool) {
	s, ok := t.Symbols[name]
	if !ok && t.Parent != nil {
		return t.Parent.Resolve(name)
	}
	return s, ok
}

// * Info Methods
func newSymbolInfo(name string, itype SymbolType) *SymbolInfo {
	return &SymbolInfo{
		Name: name,
		Type: itype,
	}
}
