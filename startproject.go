package main

import (
	"fmt"
	"github.com/jordanorelli/din/core"
	"io"
	"os"
	"os/exec"
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
			destRoot := filepath.Join(cwd, args[0])
			if err := copyTree(srcRoot, destRoot); err != nil {
				cmd.Bail(err)
			}
			if err := os.Chdir(destRoot); err != nil {
				cmd.Bail(err)
			}
			cwd, err = os.Getwd()
			if err != nil {
				cmd.Bail(err)
			}
			build := exec.Command("go", "build")
			if err := build.Run(); err != nil {
				cmd.Bail(err)
			}
			runserver := exec.Command("./"+args[0], "runserver")
			stdout, err := runserver.StdoutPipe()
			if err != nil {
				cmd.Bail(err)
			}
			stderr, err := runserver.StderrPipe()
			if err != nil {
				cmd.Bail(err)
			}
			go io.Copy(os.Stdout, stdout)
			go io.Copy(os.Stderr, stderr)
			if err := runserver.Run(); err != nil {
				cmd.Bail(err)
			}
		},
	})
}
