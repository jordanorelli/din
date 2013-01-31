package main

import (
	"github.com/jordanorelli/din/core"
)

func HomeHandler(req *din.Request) (din.Response, error) {
	return din.NewTemplateResponse("index.html", nil, 200)
}

func main() {
	din.ParseAndRun()
}

func init() {
	din.RegisterHandler("HomeHandler", HomeHandler)
}
