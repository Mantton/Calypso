package t

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/types"
)

type CheckerMode byte

const (
	//  Standard Library, Certain Restrictions are lifted
	STD CheckerMode = iota

	// User Scripts, This is the standard language
	USER
)

type Checker struct {
	Errors lexer.ErrorList
	depth  int
	mode   CheckerMode
	scope  *types.Scope
}

func New(mode CheckerMode) *Checker {
	return &Checker{
		depth: 0,
		mode:  mode,
	}
}

var unresolved types.Type

func init() {
	unresolved = types.LookUp(types.Unresolved)
}

// * Scopes
func (c *Checker) enterScope() {
	c.depth += 1

	if c.depth > 1000 {
		panic("exceeded max scope depth") // TODO : Error
	}

	parent := c.scope
	child := types.NewScope(parent)
	c.scope = child
}

func (c *Checker) leaveScope() {
	parent := c.scope.Parent

	if parent != nil {
		c.scope = parent
		c.depth -= 1

	} else {
		panic("BASE SCOPE")
	}
}

func (c *Checker) define(s types.Symbol) bool {
	result := c.scope.Define(s)
	return result
}

func (c *Checker) find(n string) (types.Symbol, bool) {
	return c.scope.Resolve(n)
}
