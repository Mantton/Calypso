package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
)

type CheckerMode byte

const (
	//  Standard Library, Certain Restrictions are lifted
	STD CheckerMode = iota

	// User Scripts, This is the standard language
	USER
)

// TODO: Replace this with unresolved property
var unresolved = newSymbolInfo("unresolved", TypeSymbol)

type Checker struct {
	Errors      lexer.ErrorList
	symbols     *SymbolTable
	depth       int
	mode        CheckerMode
	currentNode ast.Node
	currentSym  *SymbolInfo
}

func New(mode CheckerMode) *Checker {
	return &Checker{
		symbols: nil,
		depth:   0,
		mode:    mode,
	}
}

// * Scopes
func (c *Checker) enterScope() {
	c.depth += 1

	if c.depth > 1000 {
		panic("exceeded max scope depth") // TODO : Error
	}

	parent := c.symbols
	child := newSymbolTable(parent)
	c.symbols = child
}

func (c *Checker) leaveScope(isParent bool) {
	parent := c.symbols.Parent

	if parent != nil {
		c.symbols = parent
		c.depth -= 1

	} else if !isParent {
		panic("cannot exit scope")
	}
}

func (c *Checker) define(s *SymbolInfo) bool {
	result := c.symbols.Define(s)

	if result {
		switch s.Type {
		case VariableSymbol:
			fmt.Println("Defined Variable", s.Name, "As Type", s.TypeDesc.Name)
		default:
			fmt.Println("Defined", s.Name)

		}
	}
	return result
}

func (c *Checker) find(n string) (*SymbolInfo, bool) {
	return c.symbols.Resolve(n)
}
