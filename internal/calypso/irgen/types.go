package irgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
	"tinygo.org/x/go-llvm"
)

func (c *compiler) getFunction(fn *types.Function) (llvm.Value, llvm.Type) {
	llvmFn := c.module.NamedFunction(fn.Name())

	if !llvmFn.IsNil() {
		return llvmFn, llvmFn.GlobalValueType()
	}
	sg := fn.Type().(*types.FunctionSignature)
	retType := c.getType(sg.ReturnType)

	var params []llvm.Type

	for _, param := range sg.Parameters {
		t := c.getType(param.Type())
		params = append(params, t)
	}

	fnType := llvm.FunctionType(retType, params, false)
	llvmFn = llvm.AddFunction(c.module, fn.Name(), fnType)
	return llvmFn, fnType

}

func (c *compiler) getType(t types.Type) llvm.Type {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Literal {
		case types.Void:
			return c.context.VoidType()
		case types.Int:
			return c.context.Int64Type()
		case types.Float:
			return c.context.FloatType()
		case types.Bool:
			return c.context.Int1Type()
		}
	}

	panic(fmt.Sprintf("Unsupported Type: %T", t))
}
