package llir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
	"tinygo.org/x/go-llvm"
)

func (c *compiler) createConstant(n *lir.Constant) llvm.Value {
	switch t := n.Yields().Parent().(type) {
	case *types.Basic:
		switch t.Literal {
		case types.Int, types.Int64, types.IntegerLiteral, types.UInt, types.UInt64:
			v := n.Value.(int64)
			return llvm.ConstInt(c.context.Int64Type(), uint64(v), true)
		case types.Char, types.Int32, types.UInt32:
			v := n.Value.(int64)
			return llvm.ConstInt(c.context.Int32Type(), uint64(v), true)
		case types.Int16, types.UInt16:
			v := n.Value.(int64)
			return llvm.ConstInt(c.context.Int16Type(), uint64(v), true)
		case types.Int8, types.UInt8:
			v := n.Value.(int64)
			return llvm.ConstInt(c.context.Int8Type(), uint64(v), true)
		case types.Bool:
			v := n.Value.(bool)
			o := 0
			if v {
				o = 1
			}
			return llvm.ConstInt(c.context.Int1Type(), uint64(o), true)
		case types.NilLiteral:
			panic("unreachable")
		default:
			panic("basic type constant type has not been defined yet")
		}
	case *types.Pointer:
		return llvm.ConstPointerNull(c.getType(t.PointerTo))
	default:
		panic(" type constant type has not been defined yet")
	}
}

func (c *compiler) getType(t types.Type) llvm.Type {
	switch t := t.Parent().(type) {
	case *types.Basic:
		switch t.Literal {
		case types.Void:
			return c.context.VoidType()
		case types.Int, types.UInt, types.IntegerLiteral:
			return c.context.Int64Type()
		case types.Int64, types.UInt64:
			return c.context.Int64Type()
		case types.Int32, types.UInt32, types.Char:
			return c.context.Int32Type()
		case types.Int16, types.UInt16:
			return c.context.Int16Type()
		case types.Int8, types.UInt8:
			return c.context.Int8Type()
		case types.Float:
			return c.context.FloatType()
		case types.Double:
			return c.context.DoubleType()
		case types.Bool:
			return c.context.Int1Type()
		case types.NilLiteral:
			panic("INVALID")
		default:
			panic("unhandled basic type")
		}

	case *types.Pointer:
		pt := c.getType(t.PointerTo)
		return llvm.PointerType(pt, 0)
	}

	panic(fmt.Sprintf("Unsupported Type: %T", t))
}

func (c *compiler) getFunction(fn *types.Function) (llvm.Value, llvm.Type) {
	llvmFn := c.module.NamedFunction(fn.Name())

	if !llvmFn.IsNil() {
		return llvmFn, llvmFn.GlobalValueType()
	}
	sg := fn.Type().(*types.FunctionSignature)
	retType := c.getType(sg.Result.Type())

	var params []llvm.Type

	for _, param := range sg.Parameters {
		t := c.getType(param.Type())
		params = append(params, t)
	}

	fnType := llvm.FunctionType(retType, params, false)
	llvmFn = llvm.AddFunction(c.module, fn.Name(), fnType)
	return llvmFn, fnType

}