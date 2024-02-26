package ssagen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/types"
)

type builder struct {
	Mod *ssa.Module
	Fn  *ssa.Function
}

func build(e *ssa.Executable) {
	b := &builder{
		Mod: ssa.NewModule(e.IncludedFile, "main"),
	}

	for _, decl := range e.IncludedFile.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclaration:
			b.resolveFunction(d.Func)
		default:
			panic(fmt.Sprintf("unknown decl %T\n", decl))
		}
	}

	e.Modules["main"] = b.Mod
}

func (b builder) resolveFunction(n *ast.FunctionExpression) {

	fn := ssa.NewFunction(n.Signature)
	b.Mod.Functions[n.Identifier.Value] = fn
	b.Fn = fn
	fn.CurrentBlock = fn.NewBlock()
	b.resolveBlockStatement(n.Body, fn, n.Signature.Type().(*types.FunctionSignature).Scope)
}

func (b builder) resolveStmt(n ast.Statement, fn *ssa.Function, s *types.Scope) {

	switch n := n.(type) {
	case *ast.VariableStatement:
		b.resolveVariableStmt(n, fn, s)
	case *ast.ReturnStatement:
		b.resolveReturnStmt(n, fn)
	case *ast.BlockStatement:
		panic("CANNOT BE CALLED DIRECTLY")
	case *ast.ExpressionStatement:
		i, ok := b.resolveExpr(n.Expr, fn).(ssa.Instruction)

		if ok {
			fn.Emit(i)
		}
	case *ast.IfStatement:
		// 1 - Resolve Condition
		cond := b.resolveExpr(n.Condition, fn) // eval condition
		prevBlock := fn.CurrentBlock           // the current block

		// Resolve Action Block
		actionBlock := fn.NewBlock() // the action block i.e then
		var altBlock *ssa.Block
		fn.CurrentBlock = actionBlock
		b.resolveBlockStatement(n.Action, fn, s)

		// Resolve Alternative Block
		if n.Alternative != nil {
			altBlock = fn.NewBlock() // the alternate block i.e else

			fn.CurrentBlock = altBlock
			b.resolveBlockStatement(n.Alternative, fn, s)
		}

		joinBlock := fn.NewBlock() // the join block

		actionBlock.Emit(&ssa.Jump{
			Block: joinBlock,
		})

		if altBlock != nil {
			altBlock.Emit(&ssa.Jump{
				Block: joinBlock,
			})
			prevBlock.Emit(&ssa.Branch{
				Condition:   cond,
				Action:      actionBlock,
				Alternative: altBlock,
			})
		} else {
			prevBlock.Emit(&ssa.Branch{
				Condition:   cond,
				Action:      actionBlock,
				Alternative: joinBlock,
			})
		}

		fn.CurrentBlock = joinBlock

	default:
		panic(fmt.Sprintf("unknown statement %T\n", n))

	}

}

func (b *builder) resolveVariableStmt(n *ast.VariableStatement, fn *ssa.Function, s *types.Scope) {

	val := b.resolveExpr(n.Value, fn)

	// globals are constants known at compile-time
	if n.IsGlobal {
		v, ok := val.(*ssa.Constant)
		if !ok {
			panic("GLOBAL VALUE MUST BE KNOWN AT COMPILE TIME")
		}
		emitGlobalVar(b.Mod, v, n.Identifier.Value)
		return
	}

	// non global constants
	if n.IsConstant {
		v, ok := val.(*ssa.Constant)

		// constant is a compile time constant
		if ok {
			emitConstantVar(fn, v, n.Identifier.Value)
			return
		}
	}

	// -- at this point, the variable is either a runtime constant or a mutable variable

	// Variable, Allocate Memory & Store Address
	v, ok := s.Resolve(n.Identifier.Value)
	if !ok {
		panic("Var cannot be found in scope")
	}

	addr := emitLocalVar(fn, v.(*types.Var))

	// Store Data @ Addr

	emitStore(fn, addr, val)
}

func (b *builder) resolveReturnStmt(n *ast.ReturnStatement, fn *ssa.Function) {

	val := b.resolveExpr(n.Value, fn)

	i := &ssa.Return{
		Result: val,
	}

	fn.Emit(i)
}

func (b *builder) resolveBlockStatement(n *ast.BlockStatement, fn *ssa.Function, sc *types.Scope) {

	for _, s := range n.Statements {
		b.resolveStmt(s, fn, sc)
	}
}

func (b *builder) resolveExpr(n ast.Expression, fn *ssa.Function) ssa.Value {
	switch n := n.(type) {
	case *ast.IntegerLiteral:
		return &ssa.Constant{
			Value: n.Value,
			Typ:   types.LookUp(types.Int),
		}
	case *ast.BooleanLiteral:
		return &ssa.Constant{
			Value: n.Value,
			Typ:   types.LookUp(types.Bool),
		}
	case *ast.FloatLiteral:
		return &ssa.Constant{
			Value: n.Value,
			Typ:   types.LookUp(types.Float),
		}
	case *ast.StringLiteral:
		return &ssa.Constant{
			Value: n.Value,
			Typ:   types.LookUp(types.String),
		}
	case *ast.CallExpression:
		panic("fix this")
		return &ssa.Call{
			Target:    "Foo",
			Arguments: nil,
		}
	case *ast.IdentifierExpression:
		val := fn.Variables[n.Value]

		switch val := val.(type) {
		case *ssa.Allocate:
			i := &ssa.Load{
				Address: val,
			}

			i.SetType(val.Type())
			fn.Emit(i)
			return i
		case *ssa.Constant:
			return val
		default:
			panic("identifier found invalid type")
		}

	case *ast.AssignmentExpression:
		a := b.resolveExpr(n.Target, fn)
		v := b.resolveExpr(n.Value, fn)
		emitStore(fn, a, v)
		return nil
	case *ast.BinaryExpression:
		lhs, rhs := b.resolveExpr(n.Left, fn), b.resolveExpr(n.Right, fn)
		i := &ssa.Binary{
			Left:  lhs,
			Op:    n.Op,
			Right: rhs,
		}
		i.SetType(lhs.Type())
		fn.Emit(i)
		return i

	default:
		panic(fmt.Sprintf("unknown expr %T\n", n))

	}
}
