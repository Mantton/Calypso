package typechecker

import (
	"fmt"
	"strings"
	"sync"
)

// Symbol Constants
type SymbolType byte

const (
	VariableSymbol    SymbolType = iota // Variables
	FunctionSymbol                      // Functions
	StructSymbol                        // Structs
	StandardSymbol                      // Standards
	AliasSymbol                         // Aliases
	TypeSymbol                          // Types
	GenericTypeSymbol                   //Generic Arguments / Params
)

var (
	nextSymbolId int
	idLock       sync.Mutex // Ensures that ID assignment is thread-safe
)

// FunctionDescriptor describes the signature of a function, including parameters and return type.
type FunctionDescriptor struct {
	Parameters []*SymbolInfo
	ReturnType *SymbolInfo
}

type SymbolInfo struct {
	Name        string     // name of the symbol
	ModuleName  string     // the module this symbol belongs to
	PackageName string     // the package this symbol belongs to
	Type        SymbolType // The type of symbol being represented
	TypeDesc    *SymbolInfo

	Properties map[string]*SymbolInfo // For complex types like structs, this holds property types
	AliasOf    *SymbolInfo            // For aliases, points to the original type
	ChildOf    *SymbolInfo            // For defined types, points to the original/base type
	ConcreteOf *SymbolInfo
	FuncDesc   *FunctionDescriptor // For functions, describes the function's signature
	IsPrivate  bool                // if this symbol is a private property

	GenericParams    []*SymbolInfo          // Generic Params with this symbol
	GenericArguments []*SymbolInfo          // generic arguments with this symbol
	Constraints      map[string]*SymbolInfo // For Types & Structs, points to the Standards being conformed to.
	ID               int
	// TODO: Keep Track of Specializations
}

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]*SymbolInfo
}

// NewSymbolTable creates a new symbol table with an optional parent scope.
func newSymbolTable(parent *SymbolTable) *SymbolTable {
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
func newSymbolInfo(name string, t SymbolType) *SymbolInfo {
	idLock.Lock()         // Lock the mutex before modifying the counter
	defer idLock.Unlock() // Unlock the mutex after modifying the counter
	nextSymbolId++        // Increment the global ID counter

	return &SymbolInfo{
		Name:        name,
		Type:        t,
		Properties:  make(map[string]*SymbolInfo),
		Constraints: make(map[string]*SymbolInfo),
		ID:          nextSymbolId,
	}
}

func (s *SymbolInfo) Identifier() string {
	ids := []string{s.PackageName, s.ModuleName, s.Name}
	return strings.Join(ids, "_")
}
func (s *SymbolInfo) addProperty(it *SymbolInfo) bool {
	_, ok := s.Properties[it.Name]

	// is already defined
	if ok {
		return false
	}

	s.Properties[it.Name] = it
	return true
}

func (s *SymbolInfo) addConstraint(c *SymbolInfo) error {

	if c.Type != StandardSymbol {
		return fmt.Errorf("`%s` is not a conformable standard", c.Name)
	}

	s.Constraints[c.Identifier()] = c
	return nil
}

func (s *SymbolInfo) addGenericArgument(a *SymbolInfo) error {
	if a.Type != TypeSymbol && a.Type != AliasSymbol && a.Type != GenericTypeSymbol {
		return fmt.Errorf(fmt.Sprintf("`%s` is not a type", a.Name))
	}

	s.GenericArguments = append(s.GenericArguments, a)
	return nil
}

func (s *SymbolInfo) addGenericParameter(p *SymbolInfo) error {
	if p.Type != GenericTypeSymbol {
		return fmt.Errorf(
			fmt.Sprintf("`%s` is not a generic type. Report this error", p.Name),
		)
	}

	s.GenericParams = append(s.GenericParams, p)
	return nil
}

func (s *SymbolInfo) convertGenericParamsToArguments() {
	s.GenericArguments = s.GenericParams
	s.GenericParams = nil
}
