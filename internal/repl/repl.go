package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mantton/calypso/internal/evaluator"
)

func Run() {

	eval := evaluator.New()

	for {
		fmt.Print(">>> ")

		reader := bufio.NewReader(os.Stdin)

		input, err := reader.ReadString('\n')

		if err != nil {

			if err == io.EOF {
				return
			}

			fmt.Println(err)
			return
		}

		if len(input) < 1 {
			continue
		}

		err = eval.Run(input)

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
