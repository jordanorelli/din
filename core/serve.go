package din

import (
	"flag"
	"fmt"
	"os"
)

func ParseAndRun() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print("unrecognized command.  available commands:\n\n")
			cmdRegistry.run([]string{"list-commands"})
			os.Exit(2)
		}
	}()

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Print("welcome to the Din framework!  available commands:\n\n")
		cmdRegistry.run([]string{"list-commands"})
		os.Exit(0)
	}
	cmdRegistry.run(args)
}
