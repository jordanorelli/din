package main

import (
    "os"
    "fmt"
    "github.com/jordanorelli/din/core"
)

func main() {
    res, err := din.ParseRouteFile("/projects/src/github.com/jordanorelli/din/sample")
    if err != nil {
        fmt.Println("ERROR:", err)
        os.Exit(2)
    }
    fmt.Println(res)
}
