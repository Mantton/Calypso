package types

import (
	"fmt"
)

type Scope struct {
	Parent  *Scope
	symbols map[string]Symbol
}

func NewScope(p *Scope) *Scope {
	return &Scope{
		Parent:  p,
		symbols: make(map[string]Symbol),
	}
}

// Defines a new entity in the scope
func (s *Scope) Define(e Symbol) error {

	switch typ := e.(type) {
	case *Function:
		return s.defineFnSymbol(typ)
	default:
		return s.defineNonFnSymbol(e)
	}
}

func (s *Scope) defineNonFnSymbol(e Symbol) error {
	k := e.Name()

	_, ok := s.symbols[k]

	// if already defined in scope, return false
	if ok {
		return fmt.Errorf("invalid redeclaration of \"%s\"", k)
	}

	s.symbols[k] = e
	return nil
}

func (s *Scope) defineFnSymbol(fn *Function) error {

	k := fn.Name()
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
		return fmt.Errorf("invalid redeclaration of %s", fn.name)
	}

}

// Resolve searches for a symbol in the current table and parent scopes.
func (s *Scope) Resolve(name string) (Symbol, bool) {
	symbol, exists := s.symbols[name]
	if !exists && s.Parent != nil {
		return s.Parent.Resolve(name)
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

func (s *Scope) String() string {
	var str string
	str += "------SCOPE-----\n"

	for k, v := range s.symbols {
		str += fmt.Sprintf("%s : ", k)
		str += fmt.Sprintf("%s\n", v.Type())
	}
	return str
}

func (s *Scope) IsEmpty() bool {
	return len(s.symbols) == 0
}
