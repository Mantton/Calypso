package interpreter

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
)

type Interpreter struct {
	executable *lir.Executable
	addresses  map[lir.Value]lir.Value
}

func New(e *lir.Executable) *Interpreter {
	return &Interpreter{
		executable: e,
		addresses:  make(map[lir.Value]lir.Value),
	}
}

func (i *Interpreter) Execute() error {

	main := i.executable.Modules["main"]

	if main == nil {
		return fmt.Errorf("\"main\" module not found")
	}

	mainFn := main.Functions["main"]

	if mainFn == nil {
		return fmt.Errorf("\"main\" function not found")
	}

	i.walkBlocks(mainFn)
	return nil
}

func (i *Interpreter) walkBlocks(fn *lir.Function) {

	for _, block := range fn.Blocks {
		for _, instruction := range block.Instructions {
			i.visitInstruction(instruction)

		}
	}

}

func (i *Interpreter) visitInstruction(n lir.Instruction) {
	// fmt.Printf("%T\n", n)

	switch n := n.(type) {
	case *lir.Load:
		i.visitLoadInstruction(n)
	case *lir.Allocate:
		i.visitAllocateInstruction(n)
	case *lir.Store:
		i.visitStoreInstruction(n)
	case *lir.Call:
		i.visitCallInstruction(n)
	case *lir.Jump:
	case *lir.Return:
	case *lir.Branch:
	case *lir.Binary:
	default:
		fmt.Printf("INSTRUCTION %T\n", n)

	}
}

func (i *Interpreter) visitLoadInstruction(n *lir.Load) {
}

func (i *Interpreter) visitStoreInstruction(n *lir.Store) {
	i.addresses[n.Address] = n.Value
}

func (i *Interpreter) visitCallInstruction(n *lir.Call) {
	i.walkBlocks(n.Target)
}

func (*Interpreter) visitAllocateInstruction(n *lir.Allocate) {
}
