package types

type FunctionInstance struct {
	Signature *FunctionSignature
	Arguments []Type
}

func NewFunctionInstance(sg *FunctionSignature, args []Type) *FunctionInstance {
	return &FunctionInstance{
		Signature: sg,
		Arguments: args,
	}
}
