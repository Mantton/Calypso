package lirgen

import "fmt"

func (b *builder) debugPrint() {

	for k, fn := range b.Mod.Functions {

		fmt.Println("<FUNCTION> ", k, " | ", fn.Signature())

		for _, blk := range fn.Blocks {
			fmt.Printf("\tBlock %p\n", blk)

			for _, i := range blk.Instructions {
				fmt.Printf("\t\t%T\n", i)
			}
		}
	}
}
