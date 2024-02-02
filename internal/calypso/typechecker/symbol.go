package typechecker

import (
	"fmt"
	"strings"
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
	return &SymbolInfo{
		Name:        name,
		Type:        t,
		Properties:  make(map[string]*SymbolInfo),
		Constraints: make(map[string]*SymbolInfo),
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
	msg := fmt.Sprintf("\tAdded Property `%s` to `%s`", it.Name, s.Name)
	fmt.Println(msg)
	return true
}

func (s *SymbolInfo) addConstraint(c *SymbolInfo) error {

	if c.Type != StandardSymbol {
		return fmt.Errorf("`%s` is not a conformable standard", c.Name)
	}

	s.Constraints[c.Identifier()] = c
	msg := fmt.Sprintf("\tAdded Constraint `%s` to `%s`", c.Name, s.Name)
	fmt.Println(msg)
	return nil
}

func (s *SymbolInfo) addGenericArgument(a *SymbolInfo) error {
	if a.Type != TypeSymbol && a.Type != AliasSymbol && a.Type != GenericTypeSymbol {
		return fmt.Errorf(fmt.Sprintf("`%s` is not a type", a.Name))
	}

	s.GenericArguments = append(s.GenericArguments, a)
	msg := fmt.Sprintf("\tAdded Generic Argument `%s` to `%s`", a.Name, s.Name)
	fmt.Println(msg)
	return nil
}

func (s *SymbolInfo) addGenericParameter(p *SymbolInfo) error {
	if p.Type != GenericTypeSymbol {
		return fmt.Errorf(
			fmt.Sprintf("`%s` is not a generic type. Report this error", p.Name),
		)
	}

	s.GenericParams = append(s.GenericParams, p)
	msg := fmt.Sprintf("\tAdded Generic Parameter `%s` to `%s`", p.Name, s.Name)
	fmt.Println(msg)
	return nil
}

func (s *SymbolInfo) convertGenericParamsToArguments() {
	s.GenericArguments = s.GenericParams
	s.GenericParams = nil
}
