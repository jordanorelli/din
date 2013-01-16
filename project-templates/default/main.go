package main

import (
	"github.com/jordanorelli/din/core"
)

func HomeHandler(req *din.Request) (din.Response, error) {
	return din.PlaintextResponseString("Hello, World!", 200), nil
}

func main() {
	din.ParseAndRun()
}
