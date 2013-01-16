package main

import (
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
			if err := router.ListenAndServe(":8000"); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(3)
			}
		},
	})
}
