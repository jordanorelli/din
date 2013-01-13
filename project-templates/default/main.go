package main

import (
    "github.com/jordanorelli/din/core"
)

func HomeHandler(req *din.Request) (din.Response, error) {
    return din.PlaintextResponse("Hello"), nil
}

func main() {
    din.Run()
}

func init() {
    din.RegisterHandler("HomeHandler", HomeHandler)
}
