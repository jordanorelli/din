package main

import (
    "os"
)

func quit(code int, msg string) {
    os.Stderr.WriteString(msg)
    os.Exit(code)
}
