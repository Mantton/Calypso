package types

import "fmt"

type Scope struct {
	Parent  *Scope
	Symbols map[string]Symbol
}

func NewScope(p *Scope) *Scope {
	return &Scope{
		Parent:  p,
		Symbols: make(map[string]Symbol),
	}
}

// Defines a new entity in the scope
func (s *Scope) Define(e Symbol) bool {
	// fmt.Printf("defining %s\n", e.Name())
	// defer fmt.Println(s)
	k := e.Name()
	_, ok := s.Symbols[k]

	// if already defined in scope, return false
	if ok {
		return false
	}

	s.Symbols[k] = e
	return true
}

// Resolve searches for a symbol in the current table and parent scopes.
func (s *Scope) Resolve(name string) (Symbol, bool) {
	fmt.Printf("Resolving %s\n", name)
	// defer fmt.Println(s)

	symbol, exists := s.Symbols[name]
	if !exists && s.Parent != nil {
		return s.Parent.Resolve(name)
	}
	return symbol, exists
}
