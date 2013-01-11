package main

import (
	"fmt"
	"github.com/jordanorelli/din/core"
)

var cmdStartProject = din.Command{
	UsageLine: "startproject [project_name]",
	Short:     "start a new din project",
	Long: `
the startproject subcommand creates a new din project, including config and route files
`,
	Run: runStartProject,
}

func runStartProject(cmd *din.Command, args []string) {
	fmt.Println("start project!")
}

func init() {
	din.RegisterCommand(cmdStartProject)
}
