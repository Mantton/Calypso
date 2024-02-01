package resolver

import "github.com/mantton/calypso/internal/calypso/ast"

// * Resolvers
func (r *Resolver) ResolveFile(file *ast.File) {

	r.enterScope() // Global Scope
	// PreDefine

	// Resolve
	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			r.resolveDeclaration(decl)
		}
	}

	r.leaveScope(true) // Leave Global Scope

	// TODO: Ensure Global Scope is Left here
}
