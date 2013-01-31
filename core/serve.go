package din

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
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

func parseRoutesFile() (*Router, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, InternalServerError("unable to find routes file")
	}
	return ParseRouteFile(filepath.Join(cwd, "routes.json"))
}

func darwinOpenBrowser() {
	p := "http://localhost:8000"
	fmt.Println("attempting to open a browser to", p)
	exec.Command("open", p).Start()
}

// function openBrowser attempts to open the user's web browser to show them
// the home page, so that they know their project is running.  Right now it
// only supports OS X.  I don't really know how to do it otherwise, and for
// Linux I would want to detect a graphical environment, because it would be an
// annoying thing to try on a server.
func openBrowser() {
	switch runtime.GOOS {
	case "darwin":
		darwinOpenBrowser()
	}
}

func init() {
	var autoBrowse bool
	cmd := Command{
		UsageLine: "runserver",
		Short:     "runs the Din server",
		Long: `
This runs the Din webserver.  Blah blah blah this is the long description so
I should write some more stuff, But I don't really feel like it I'm pretty
wasted and I really like absinthe, specifically 'Vieux de Pontarlier' is really
great.
`,

		Run: func(cmd *Command, args []string) {
			router, err := parseRoutesFile()
			if err != nil {
				cmd.Bail(err)
			}
			path := locateConfig()
			if path == "" {
				cmd.Bail(errors.New("unable to locate din config file.  Please set environment variable $DIN_CONFIG to be the absolute path of your configuration file."))
			}
			fmt.Println("using config found at " + path)
			if err := ParseConfigFile(path); err != nil {
				cmd.Bail(err)
			}
			fmt.Println(Config)
			if autoBrowse {
				time.AfterFunc(time.Second, openBrowser)
			}
			if err := router.ListenAndServe(Config.Core.Addr); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(3)
			}
		},
	}
	cmd.Flag.BoolVar(&autoBrowse, "open-browser", false, "open a web browser after running the server (OS X only)")
	RegisterCommand(cmd)
}
