package main

import (
	"fmt"
)

var (
	importPath = "github.com/jordanorelli/din"
	assetsPath = "assets"
)

func main() {
	fmt.Println(getPkgDir(importPath))
}
