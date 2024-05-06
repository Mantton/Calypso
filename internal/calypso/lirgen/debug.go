package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) debugPrint() {
	fmt.Println()
	fmt.Println("Module Completed:", b.Mod.TModule.SymbolName())
	fmt.Println("\nFunctions")

	for k, fn := range b.Mod.Functions {

		fmt.Println(k, " | ", fn.Signature())

		for _, blk := range fn.Blocks {
			fmt.Printf("\tBlock %p\n", blk)

			for _, i := range blk.Instructions {
				fmt.Printf("\t\t%T\n", i)
			}
		}
	}
	fmt.Println()

	fmt.Println("\nComposites")
	for _, c := range b.Mod.Composites {
		fmt.Println(c)
	}

	fmt.Println()

	fmt.Println("\nEnums")
	for _, c := range b.Mod.Enums {
		fmt.Println(types.SymbolName(c.Type))
	}
	fmt.Println()

}
