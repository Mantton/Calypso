package symbols

import (
	"fmt"
	"strings"
	"sync"
)

// Symbol Constants
type SymbolType byte
type Literal byte

const (
	VariableSymbol    SymbolType = iota // Variables
	FunctionSymbol                      // Functions
	StructSymbol                        // Structs
	StandardSymbol                      // Standards
	AliasSymbol                         // Aliases
	TypeSymbol                          // Types
	GenericTypeSymbol                   // Generic Arguments / Params
)

const (
	INTEGER Literal = iota
	FLOAT
	STRING
	BOOLEAN
	ARRAY
	MAP
	NULL
	VOID
	ANY
)

var (
	nextSymbolId int
	idLock       sync.Mutex // Ensures that ID assignment is thread-safe
)

// FunctionDescriptor describes the signature of a function, including parameters and return type.
type FunctionDescriptor struct {
	Parameters          []*SymbolInfo
	AnnotatedReturnType *SymbolInfo
	InferredReturnType  *SymbolInfo
	ValidatedReturnType *SymbolInfo
}

type SymbolInfo struct {
	Name        string     // name of the symbol
	ModuleName  string     // the module this symbol belongs to
	PackageName string     // the package this symbol belongs to
	Type        SymbolType // The type of symbol being represented
	TypeDesc    *SymbolInfo

	Fields    map[string]*SymbolInfo // For complex types like structs, this holds property types
	AliasOf   *SymbolInfo            // For aliases, points to the original type
	ChildOf   *SymbolInfo            // For defined types, points to the original/base type
	FuncDesc  *FunctionDescriptor    // For functions, describes the function's signature
	IsPrivate bool                   // if this symbol is a private property
	Mutable   bool

	GenericParams []*SymbolInfo          // Generic Params with this symbol
	Constraints   map[string]*SymbolInfo // For Types & Structs, points to the Standards being conformed to.
	ID            int

	Specializations SpecializationTable
	SpecializedOf   *SymbolInfo
}

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]*SymbolInfo
}

type SpecializationTable map[*SymbolInfo]*SymbolInfo

// NewSymbolTable creates a new symbol table with an optional parent scope.
func NewTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		Parent:  parent,
		Symbols: make(map[string]*SymbolInfo),
	}
}

// Define adds a new symbol to the symbol table.
func (t *SymbolTable) Define(symbol *SymbolInfo) bool {
	_, ok := t.Symbols[symbol.Name]

	// if already defined in scope, return false
	if ok {
		return false
	}

	t.Symbols[symbol.Name] = symbol
	return true
}

// Resolve searches for a symbol in the current table and parent scopes.
func (t *SymbolTable) Resolve(name string) (*SymbolInfo, bool) {
	symbol, exists := t.Symbols[name]
	if !exists && t.Parent != nil {
		return t.Parent.Resolve(name)
	}
	return symbol, exists
}

// * Symbol Info
func NewSymbol(name string, t SymbolType) *SymbolInfo {
	idLock.Lock()         // Lock the mutex before modifying the counter
	defer idLock.Unlock() // Unlock the mutex after modifying the counter
	nextSymbolId++        // Increment the global ID counter

	return &SymbolInfo{
		Name:            name,
		Type:            t,
		Fields:          make(map[string]*SymbolInfo),
		Constraints:     make(map[string]*SymbolInfo),
		Specializations: make(SpecializationTable),
		ID:              nextSymbolId,
	}
}

func (s *SymbolInfo) Identifier() string {
	ids := []string{s.PackageName, s.ModuleName, s.Name}
	return strings.Join(ids, "_")
}
func (s *SymbolInfo) AddProperty(it *SymbolInfo) bool {
	_, ok := s.Fields[it.Name]

	// is already defined
	if ok {
		return false
	}

	s.Fields[it.Name] = it

	fmt.Printf("Adding Property %s of type %s to %s\n", it, it.TypeDesc, s)
	return true
}

func (s *SymbolInfo) AddConstraint(c *SymbolInfo) error {

	if c.Type != StandardSymbol {
		return fmt.Errorf("`%s` is not a conformable standard", c.Name)
	}

	s.Constraints[c.Identifier()] = c
	return nil
}

func (s *SymbolInfo) AddGenericParameter(p *SymbolInfo) error {
	if p.Type != GenericTypeSymbol {
		return fmt.Errorf(
			fmt.Sprintf("`%s` is not a generic type. Report this error", p.Name),
		)
	}

	s.GenericParams = append(s.GenericParams, p)
	return nil
}

func (s *SymbolInfo) String() string {
	return fmt.Sprintf("%s(%d)", s.Name, s.ID)
}

func (s *SymbolInfo) Info() string {
	out := "\n==========================\n"
	out += fmt.Sprintf("Module: `%s`\n", s.ModuleName)
	out += fmt.Sprintf("Package: `%s`\n", s.PackageName)
	out += fmt.Sprintf("Name: `%s`\n", s.Name)
	out += fmt.Sprintf("ID: `%d`\n", s.ID)
	out += fmt.Sprintf("Type: `%s`\n", LookUpNameOfSymbolType(s.Type))

	if s.TypeDesc != nil {
		out += fmt.Sprintf("\t%s\n", s.TypeDesc.String())
	}

	if s.AliasOf != nil {
		out += fmt.Sprintf("Alias of: `%s`\n", s.AliasOf.String())
	}

	if s.SpecializedOf != nil {
		out += fmt.Sprintf("Specialization of: `%s`\n", s.SpecializedOf.String())
		out += fmt.Sprintf("%d Specs\n", len(s.Specializations))
	}

	for gen, spec := range s.Specializations {
		out += fmt.Sprintf("Specializes %s for %s\n", gen, spec)
	}

	for _, param := range s.GenericParams {
		out += fmt.Sprintf("Generic Param\n%s", param.String())
	}

	out += "==========================\n"

	return out
}

func LookUpNameOfSymbolType(t SymbolType) string {
	switch t {
	case VariableSymbol:
		return "Variable"
	case StructSymbol:
		return "Struct"
	case StandardSymbol:
		return "Standard"
	case AliasSymbol:
		return "Alias"
	case TypeSymbol:
		return "Type"
	case GenericTypeSymbol:
		return "Generic Type"
	default:
		return "UNDEFINED SYMBOL"
	}
}

func (t SpecializationTable) Get(s *SymbolInfo) (*SymbolInfo, bool) {

	if s.Type != GenericTypeSymbol {
		return s, true
	}

	v, ok := t[s]

	return v, ok
}

func (t SpecializationTable) Debug() {
	fmt.Println("\n[Specialization Table] DEBUG")
	for key, value := range t {
		fmt.Println(" >>>>>>", key, "Maps To", value, "Args")
	}
	fmt.Println()
}
