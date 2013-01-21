package din

import (
	"flag"
	"fmt"
	"os"
)

func ParseAndRun() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Print("welcome to the Din framework!  available commands:\n\n")
		cmdRegistry.run([]string{"list-commands"})
		os.Exit(2)
	}
	cmdRegistry.run(args)
}
