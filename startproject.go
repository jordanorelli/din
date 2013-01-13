package main

import (
	"fmt"
	"github.com/jordanorelli/din/core"
	"os"
	"path/filepath"
)

func init() {
	din.RegisterCommand(din.Command{
		UsageLine: "startproject [project_name]",
		Short:     "start a new din project",
		Long: `
the startproject subcommand creates a new din project, including config and route files
`,
		Run: func(cmd *din.Command, args []string) {
			if len(args) != 1 {
				fmt.Println(args)
				cmd.Usage()
			}
			srcRoot := filepath.Join(
				getPkgDir(cmdPath),
				projectTemplateDir,
				defaultProjectDir,
			)
			cwd, err := os.Getwd()
			if err != nil {
				cmd.Bail(err)
			}
			if err := copyTree(srcRoot, filepath.Join(cwd, args[0])); err != nil {
				cmd.Bail(err)
			}
		},
	})
}
