package types

import (
	"fmt"
)

type Scope struct {
	Parent  *Scope
	symbols map[string]Symbol
	label   string
}

func NewScope(p *Scope, label string) *Scope {
	return &Scope{
		Parent:  p,
		symbols: make(map[string]Symbol),
		label:   label,
	}
}

// Defines a new entity in the scope
func (s *Scope) Define(e Symbol) error {

	switch typ := e.(type) {
	case *Function:
		return s.defineFnSymbol(typ, typ.Name())
	default:
		return s.defineNonFnSymbol(e, e.Name())
	}
}

func (s *Scope) CustomDefine(e Symbol, name string) error {
	switch typ := e.(type) {
	case *Function:
		return s.defineFnSymbol(typ, name)
	default:
		return s.defineNonFnSymbol(e, name)
	}
}

func (s *Scope) defineNonFnSymbol(e Symbol, k string) error {
	_, ok := s.symbols[k]

	// if already defined in scope, return false
	if ok {
		return fmt.Errorf("invalid redeclaration of symbol \"%s\"", k)
	}

	s.symbols[k] = e
	return nil
}

func (s *Scope) defineFnSymbol(fn *Function, k string) error {

	sym, ok := s.symbols[k]

	// no definitions found, create new
	if !ok {
		s.symbols[k] = fn
		return nil
	}

	switch t := sym.(type) {
	case *Function:
		// Create Function Set based of current symbol function
		set := NewFunctionSet(t)

		// add new function to updated set
		err := set.Add(fn)

		// handle addition error
		if err != nil {
			return err
		}

		// if no error, update the symbol set, return nil
		s.symbols[k] = set
		return nil
	case *FunctionSet:
		return t.Add(fn)
	default:
		return fmt.Errorf("invalid redeclaration of function %s", fn.name)
	}

}

// Resolve searches for a symbol in the current table and parent scopes.
func (s *Scope) Resolve(name string, fallback *Scope) (Symbol, bool) {
	symbol, exists := s.symbols[name]

	// found in current scope
	if exists {
		return symbol, exists
	}

	// does not exist in scope & has parent check
	if !exists && s.Parent != nil {
		symbol, exists = s.Parent.Resolve(name, nil)
	}

	// Parent had symbol
	if exists {
		return symbol, exists
	}

	// does not exist, and fall back is not nil
	if fallback != nil {
		symbol = fallback.ResolveInCurrent(name)
		if symbol != nil {
			return symbol, true
		}
	}

	return symbol, exists
}

func (s *Scope) MustResolve(name string) Symbol {
	symbol, exists := s.symbols[name]
	if !exists && s.Parent != nil {
		return s.Parent.MustResolve(name)
	}
	return symbol
}

func (s *Scope) ResolveInCurrent(name string) Symbol {
	symbol, exists := s.symbols[name]
	if !exists {
		return nil
	}
	return symbol
}

func (s *Scope) ResolveVarInCurrent(name string) *Var {
	symbol, exists := s.symbols[name]
	if !exists {
		return nil
	}
	x, ok := symbol.(*Var)

	if ok {
		return x
	}

	return nil
}

func (s *Scope) String() string {
	var str string
	str += fmt.Sprintf("------ SCOPE <%s>  -----\n", s.label)

	for k, v := range s.symbols {
		str += fmt.Sprintf("%s : ", k)
		str += fmt.Sprintf("%s\n", v.Type())
	}
	return str
}

func (s *Scope) IsEmpty() bool {
	return len(s.symbols) == 0
}

func (s *Scope) DebugPrintChildrenScopes() {

	for _, symbol := range s.symbols {
		definition := AsDefined(symbol.Type())

		if definition == nil {
			continue
		}

		if definition.scope.IsEmpty() {
			continue
		}

		fmt.Println(definition.scope)
	}
}
