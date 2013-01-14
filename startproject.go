package main

import (
	"fmt"
	"github.com/jordanorelli/din/core"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// runCmd runs a shell command, piping its standard streams into stdout and
// stderr, making them visible to the calling shell.  If it cannot connect to
// os.Stdout or os.Stderr, there is no notification of failure, so this isn't a
// very good implementation.
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if out, err := cmd.StdoutPipe(); err != nil {
		return err
	} else {
		go io.Copy(os.Stdout, out)
	}
	if errOut, err := cmd.StderrPipe(); err != nil {
		return err
	} else {
		go io.Copy(os.Stderr, errOut)
	}
	return cmd.Run()
}

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
			fmt.Println("compiling project", args[0])
			if err := runCmd("go", "build"); err != nil {
				cmd.Bail(err)
			}
			fmt.Println("starting Din webserver...")
			if err := runCmd("./"+args[0], "runserver"); err != nil {
				cmd.Bail(err)
			}
		},
	})
}
