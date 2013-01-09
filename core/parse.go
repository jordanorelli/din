package din

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

// tree is the representation of a fully parsed route file, ready to be
// executed.
type tree struct {
	root *listNode
	lex  *lexer
}

func newTree(src io.RuneReader) {

}

/*

// next returns the next token
func (t *tree) next() token {
}

// backup backs the input stream up one token
func (t *tree) backup() {
}

// peek returns but does not consume the topmost token
func (t *tree) peek() token {
}

// startParse initializes the parser using the given lexer
func (t *tree) startParse(lex *lexer) {

}

// stopParse stops the parsing process and throws away the lexer
func (t *tree) stopParse() {

}

func (t *tree) parse() {

}

*/

func ParseRouteFile(filename string) (*Router, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		// TODO: add surpressed error to error output object
		return nil, InternalServerError("din: unable to read route config")
	}
	return parse(bytes.NewBuffer(b))
}

func parse(src io.RuneReader) (*Router, error) {
	tokens, err := lexAll(src)
	if err != nil {
		// TODO: add surpressed error to error output object
		return nil, InternalServerError("din: unable to lex route config")
	}
	for _, t := range tokens {
		fmt.Println(t)
	}
	return nil, nil
}
