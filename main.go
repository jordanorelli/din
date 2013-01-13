package main

import (
	"github.com/jordanorelli/din/core"
)

var (
	// alright, this is a bit circular.  cmdPath refers to the theoretical
	// import path if the din command was an importable package, but it isn't.
	// What this allows us to do is crawl the user's filesystem, looking for
	// the source directory in which din was installed.  The reason for this is
	// that the din project directory will also contain the files that are to
	// be duplicated on new project creation.  This allows us to bundle
	// everything up in one go-gettable package.
	cmdPath = "github.com/jordanorelli/din"

	// directory name of the directory containing project templates.
	projectTemplateDir = "project-templates"

	// directory name of the default project template.
	defaultProjectDir = "default"
)

func main() {
	din.ParseAndRun()
}
