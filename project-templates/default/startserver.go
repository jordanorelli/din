package main

import (
	"github.com/jordanorelli/din/core"
    "os"
	"net/http"
)

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
			router := din.NewRouter(nil, nil)
			router.AddRoute("^/$", "HomeHandler", HomeHandler)
			server := http.Server{
				Addr:    ":8000",
				Handler: router,
			}
            if err := server.ListenAndServe(); err != nil {
                os.Stderr.WriteString(err.Error())
                os.Exit(3)
            }
		},
	})
}
