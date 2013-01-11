package main

import (
	"github.com/jordanorelli/din/core"
)

func init() {
	din.RegisterCommand(din.Command{
		UsageLine: "help",
		Short:     "shows help information for the Din framework",
		Long: `
welcome to the Din framework!
`,
	})
}
