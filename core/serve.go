package din

import (
	"flag"
)

func ParseAndRun() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}
	cmdRegistry.run(args)
}
