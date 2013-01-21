package main

import (
    "fmt"
    "errors"
	"github.com/jordanorelli/din/core"
	"os"
	"path/filepath"
)

func parseRoutesFile() (*din.Router, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, din.InternalServerError("unable to find routes file")
	}
	return din.ParseRouteFile(filepath.Join(cwd, "routes.json"))
}

func locateConfig() string {
    path := os.Getenv("DIN_CONFIG")
    if path != "" {
        return path
    }
    cwd, err := os.Getwd()
    if err != nil {
        return ""
    }
    return filepath.Join(cwd, "config.json")
}

func init() {
	din.RegisterCommand(din.Command{
		UsageLine: "runserver",
		Short:     "runs the Din server",
		Long: `
This runs the Din webserver.  Blah blah blah this is the long description so
I should write some more stuff, But I don't really feel like it I'm pretty
wasted and I really like absinthe, specifically 'Vieux de Pontarlier' is really
great.
`,
		Run: func(cmd *din.Command, args []string) {
			router, err := parseRoutesFile()
			if err != nil {
				cmd.Bail(err)
			}
            path := locateConfig()
            if path == "" {
                cmd.Bail(errors.New("unable to locate din config file.  Please set environment variable $DIN_CONFIG to be the absolute path of your configuration file."))
            }
            fmt.Println("using config found at " + path)
            if err := din.ParseConfigFile(path); err != nil {
                cmd.Bail(err)
            }
            fmt.Println(din.Config)
			if err := router.ListenAndServe(din.Config.Core.Addr); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(3)
			}
		},
	})
}
