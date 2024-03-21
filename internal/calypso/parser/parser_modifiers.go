package parser

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
)

type modifiers struct {
	vis      ast.Visibility
	async    bool
	static   bool
	mutating bool
}

func (m *modifiers) IsValidNonFunctionModifierList() bool {
	if m.async || m.static || m.mutating {
		return false
	}
	return true
}

func (p *Parser) handleModifier(t token.Token) error {

	// No Modifiers
	if len(p.modifiers) == 0 {
		p.modifiers = append(p.modifiers, t)
		return nil
	}

	// Has at least ONE (1) modifier, ensure precedence order
	last_idx := len(p.modifiers) - 1
	last_mod := p.modifiers[last_idx]
	last_prec := token.ModifierPrecedent[last_mod]
	next_prec := token.ModifierPrecedent[t]

	if next_prec < last_prec {
		return p.error(fmt.Sprintf("invalid member modifier order, \"%s\" must precede \"%s\".",
			token.LookUp(token.Token(t)),
			token.LookUp(last_mod)))
	} else if next_prec == last_prec {
		return p.error(fmt.Sprintf("cannot use \"%s\" & \"%s\" modifiers on the same member",
			token.LookUp(token.Token(t)),
			token.LookUp(last_mod)))
	}

	p.modifiers = append(p.modifiers, t)
	return nil

}

func (p *Parser) consumeModifiers() *modifiers {

	out := &modifiers{}
	for _, mod := range p.modifiers {
		switch mod {
		case token.PUB:
			out.vis = ast.PUBLIC
		case token.ASYNC:
			out.async = true
		case token.STATIC:
			out.static = true
		case token.MUTATING:
			out.mutating = true
		}
	}

	p.modifiers = nil

	return out
}

func (p *Parser) resolveNonFuncMods() (ast.Visibility, error) {
	vis := ast.INTERNAL
	if len(p.modifiers) != 0 {
		mods := p.consumeModifiers()

		if !mods.IsValidNonFunctionModifierList() {
			return vis, p.error("invalid modifiers for member")
		}
		vis = mods.vis
	}

	return vis, nil
}
