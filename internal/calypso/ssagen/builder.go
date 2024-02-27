package ssagen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/typechecker"
	"github.com/mantton/calypso/internal/calypso/types"
)

type builder struct {
	Mod *ssa.Module
	Tbl *typechecker.SymbolTable
}

func build(e *ssa.Executable, t *typechecker.SymbolTable) {
	b := &builder{
		Mod: ssa.NewModule(e.IncludedFile, "main"),
		Tbl: t,
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

	nd, ok := b.Tbl.GetNode(n)

	if !ok {
		panic("node not in table")
	}

	sym, ok := nd.Symbol.(*types.Function)

	if !ok {
		panic("expecting function symbol")
	}

	fn := ssa.NewFunction(sym)
	b.Mod.Functions[n.Identifier.Value] = fn
	sg := sym.Type().(*types.FunctionSignature)

	for _, p := range sg.Parameters {
		fn.AddParameter(p)
	}
	fn.CurrentBlock = fn.NewBlock()
	b.resolveBlockStatement(n.Body, fn, sg.Scope)
}

func (b builder) resolveStmt(n ast.Statement, fn *ssa.Function, s *types.Scope) {
	fmt.Printf("[SSAGEN] Building %T\n", n)
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

	case *ast.IntegerLiteral,
		*ast.BooleanLiteral,
		*ast.FloatLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral:

		tn, ok := b.Tbl.GetNode(n)
		fmt.Printf("[Resolver] %T\n", n)
		fmt.Println(n)
		if !ok {
			panic("cannot resolve constant literal, this path should be unreachable")
		}
		return ssa.NewConst(tn.Value, tn.Type)

	case *ast.CallExpression:
		val := b.resolveExpr(n.Target, fn)

		f, ok := val.(*ssa.Function)

		if !ok {
			panic("cannot be invoked")
		}

		var args []ssa.Value

		for _, p := range n.Arguments {
			v := b.resolveExpr(p, fn)
			args = append(args, v)
		}
		i := &ssa.Call{
			Target:    f,
			Arguments: args,
		}

		i.SetType(f.Symbol.Type().(*types.FunctionSignature).ReturnType)

		fn.Emit(i)
		return i
	case *ast.IdentifierExpression:
		val, ok := fn.Variables[n.Value]

		if ok {
			switch val := val.(type) {
			case *ssa.Allocate:
				i := &ssa.Load{
					Address: val,
				}

				i.SetType(val.Type())
				fn.Emit(i)
				return i
			case *ssa.Constant, *ssa.Parameter:
				return val
			default:
				panic(fmt.Sprintf("identifier found invalid type: %T", val))
			}
		}

		val, ok = b.Mod.GlobalConstants[n.Value]

		if ok {
			return val
		}

		fn, ok = b.Mod.Functions[n.Value]

		if ok {
			return fn
		}

		panic("unable to locate identifier")

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
