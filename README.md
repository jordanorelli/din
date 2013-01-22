# the Din web framework

Din is a web framework for the Go programming language.  Din aims to
consolidate a lot of repetitive application programming tasks with regular Go
code, trying its best to utilize the features of the Go programming language.

# install

To install Din, use [the go command](http://golang.org/cmd/go/) to get the
package.  The only requirements are git and Go.

`go get github.com/jordanorelli/din`

The go get command will do two things: clone this git repository, and build the
`din` command.  You will use the `din` command to start new Din projects.

# start a new project

To start a new project named `hello_din`, execute the following command:

`din startproject hello_din`

The `startproject` command will do the following:
  - create a hello\_din directory
  - compile your Din project
  - run the Din webserver

After running the `startproject` command, you should be able to navigate to
`localhost:8000` to see a "Hello, World" page.
