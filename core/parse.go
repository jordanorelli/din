package din

import (
// "encoding/json"
// "bytes"
// "fmt"
// "io"
// "io/ioutil"
// "os"
)

// type parseError string
// 
// func (p parseError) Error() string {
// 	return fmt.Sprintf("parse error: %s", string(p))
// }
// 
// func unexpected(badthing, context string) error {
// 	return parseError(fmt.Sprintf("unexpected %s in %s", badthing, context))
// }
// 
// // tree is the representation of a fully parsed route file, ready to be
// // executed.
// type tree struct {
// 	root      *listNode
// 	lex       *lexer
// 	lookahead [2]token
// 	peekCount int
// 	types     map[string]constructorType
// }
// 
// func newTree(src io.RuneReader) *tree {
// 	return &tree{
// 		lex: newLexer(src),
// 	}
// }
// 
// // I'm basically just copying Rob Pike at this point.
// func (t *tree) backup() {
// 	t.peekCount++
// }
// 
// // this one is extra weird.
// func (t *tree) backup2(tok token) {
// 	t.lookahead[1] = tok
// 	t.peekCount = 2
// }
// 
// // peek returns but does not consume the topmost token
// func (t *tree) peek() token {
// 	if t.peekCount > 0 {
// 		return t.lookahead[t.peekCount-1]
// 	}
// 	t.peekCount = 1
// 	t.lookahead[0], _ = t.lex.nextToken()
// 	return t.lookahead[0]
// }
// 
// // next consumes the next token.  If there's stuff in the lookahead buffer,
// // it's taken off.  If not, we just read from the lexer diretly.
// func (t *tree) next() token {
// 	if t.peekCount > 0 {
// 		t.peekCount--
// 	} else {
// 		t.lookahead[0], _ = t.lex.nextToken()
// 	}
// 	return t.lookahead[t.peekCount]
// }
// 
// func (t *tree) parseRoute() {
// 
// }
// 
// func (t *tree) assignment() (*assignmentNode, error) {
// 	leftToken := t.next()
// 	switch leftToken.T {
// 	case routeToken:
// 		break
// 	default:
// 		return nil, unexpected(leftToken.T.String(), "assignment")
// 	}
// 	if tok := t.next(); tok.T != assignmentToken {
// 		return nil, unexpected(tok.T.String(), "assignment")
// 	}
// 	switch tok := t.next(); tok.T {
// 	case symbolToken:
// 		switch next := t.peek(); next.T {
// 		case hashStartToken:
// 
// 		}
// 	}
// }
// 
// // startParse initializes the parser using the given lexer
// func (t *tree) startParse(lex *lexer) {
// 
// }
// 
// // stopParse stops the parsing process and throws away the lexer
// func (t *tree) stopParse() {
// 
// }
// 
// func (t *tree) parse() {
// 	t.root = newList()
//     for t.peek().T != endToken {
//         if t.peek().T == routeToken {
//         }
//     }
// }

// func parse(src io.RuneReader) (*Router, error) {
// 	tokens, err := lexAll(src)
// 	if err != nil {
// 		// TODO: add surpressed error to error output object
// 		return nil, InternalServerError("din: unable to lex route config")
// 	}
// 	for _, t := range tokens {
// 		fmt.Println(t)
// 	}
// 	return nil, nil
// }
