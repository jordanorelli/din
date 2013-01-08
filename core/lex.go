package din

import (
	"errors"
	"fmt"
	"io"
)

const (
	ROUTE_START_CHAR                = '/'
	COMMENT_CHAR                    = '#'
	ARGS_START_CHAR                 = '('
	ARGS_END_CHAR                   = ')'
	ELEM_SEPARATOR_CHAR             = ','
	LIST_START_CHAR                 = '['
	LIST_END_CHAR                   = ']'
	HASH_START_CHAR                 = '{'
	HASH_END_CHAR                   = '}'
	HASH_SEPARATOR_CHAR             = ':'
	STRING_DELIMITER_CHAR           = '"'
	MULTILINE_STRING_DELIMITER_CHAR = '`'
	ESCAPE_CHAR                     = '\\'
	ASSIGNMENT_CHAR                 = '='
	CONCATENATION_CHAR              = '+'
)

type lexError struct {
	msg   string
	inner error
}

func newLexError(msg string, inner error) error {
	return &lexError{msg, inner}
}

func (l lexError) Error() string {
	if l.inner == nil {
		return l.msg
	}
	return fmt.Sprintf("%v: %v", l.msg, l.inner)
}

type tokenType int

const (
	invalidToken tokenType = iota
	routeToken
	symbolToken        //
	stringToken        //
	argsStartToken     //
	argsEndToken       //
	elemSeparatorToken // ,
	listStartToken     // [
	listEndToken       // ]
	hashStartToken     // {
	hashEndToken       // }
	hashSeparatorToken // :
	assignmentToken    // =
	concatenationToken // +
)

var tokenNames = map[tokenType]string{
	invalidToken:       "invalid",
	routeToken:         "route",
	symbolToken:        "symbol",
	stringToken:        "string",
	argsStartToken:     "argsStart",
	argsEndToken:       "argsEnd",
	elemSeparatorToken: "elemSeparator",
	listStartToken:     "listStart",
	listEndToken:       "listEnd",
	hashStartToken:     "hashStart",
	hashEndToken:       "hashEnd",
	hashSeparatorToken: "hashSeparator",
	assignmentToken:    "assignment",
	concatenationToken: "concatenation",
}

func (t tokenType) String() string {
	s, ok := tokenNames[t]
	if !ok {
		panic("unknown token type")
	}
	return s
}

type token struct {
	Lexeme string
	T      tokenType
}

type stateFn func(*lexer) (stateFn, error)

type lexer struct {
	io.RuneReader
	buf       []rune
	cur       rune
	out       chan token
	lineCount int
	charCount int
}

func (l *lexer) emit(t tokenType) {
	l.out <- token{Lexeme: string(l.buf), T: t}
	l.buf = nil
}

func (l *lexer) next() error {
	r, _, err := l.ReadRune()
	if err != nil {
		return err
	}
	l.cur = r
	if isLineEnding(r) {
		l.lineCount++
		l.charCount = 0
	} else {
		l.charCount++
	}
	return nil
}

func (l *lexer) unexpectedChar(stateName string) error {
	s := "unexpected %c in %v at %d, %d"
	return errors.New(fmt.Sprintf(s, l.cur, stateName, l.lineCount, l.charCount))
}

func lexRoot(l *lexer) (stateFn, error) {
	switch {
	case isAlpha(l.cur):
		l.keep()
		return lexSymbol, nil
	case isWhitespace(l.cur):
		return lexRoot, nil
	}

	switch l.cur {
	case ROUTE_START_CHAR:
		l.keep()
		return lexRoute, nil
	case COMMENT_CHAR:
		return lexComment, nil
	case ARGS_START_CHAR:
		l.keep()
		l.emit(argsStartToken)
		return lexRoot, nil
	case ARGS_END_CHAR:
		l.keep()
		l.emit(argsEndToken)
		return lexRoot, nil
	case ELEM_SEPARATOR_CHAR:
		l.keep()
		l.emit(elemSeparatorToken)
		return lexRoot, nil
	case LIST_START_CHAR:
		l.keep()
		l.emit(listStartToken)
		return lexRoot, nil
	case LIST_END_CHAR:
		l.keep()
		l.emit(listEndToken)
		return lexRoot, nil
	case HASH_START_CHAR:
		l.keep()
		l.emit(hashStartToken)
		return lexRoot, nil
	case HASH_END_CHAR:
		l.keep()
		l.emit(hashEndToken)
		return lexRoot, nil
	case HASH_SEPARATOR_CHAR:
		l.keep()
		l.emit(hashSeparatorToken)
		return lexRoot, nil
	case STRING_DELIMITER_CHAR:
		return lexString, nil
	case MULTILINE_STRING_DELIMITER_CHAR:
		return lexMultilineString, nil
	case ASSIGNMENT_CHAR:
		l.keep()
		l.emit(assignmentToken)
		return lexRoot, nil
	case CONCATENATION_CHAR:
		l.keep()
		l.emit(concatenationToken)
		return lexRoot, nil
	}
	return nil, l.unexpectedChar("lexRoot")
}

func lexRoute(l *lexer) (stateFn, error) {
	switch {
	case isWhitespace(l.cur):
		l.emit(routeToken)
		return lexRoot, nil
	case l.cur == ASSIGNMENT_CHAR:
		l.emit(routeToken)
		l.keep()
		l.emit(assignmentToken)
		return lexRoot, nil
	}
	l.keep()
	return lexRoute, nil
}

func lexString(l *lexer) (stateFn, error) {
	switch l.cur {
	case STRING_DELIMITER_CHAR:
		l.emit(stringToken)
		return lexRoot, nil
	case ESCAPE_CHAR:
		return lexStringEscape, nil
	}
	l.keep()
	return lexString, nil
}

func lexMultilineString(l *lexer) (stateFn, error) {
	switch l.cur {
	case MULTILINE_STRING_DELIMITER_CHAR:
		l.emit(stringToken)
		return lexRoot, nil
	case ESCAPE_CHAR:
		return lexStringEscape, nil
	}
	l.keep()
	return lexMultilineString, nil
}

func lexStringEscape(l *lexer) (stateFn, error) {
	l.keep()
	return lexString, nil
}

func lexComment(l *lexer) (stateFn, error) {
	if isLineEnding(l.cur) {
		return lexRoot, nil
	}
	return lexComment, nil
}

func lexSymbol(l *lexer) (stateFn, error) {
	switch {
	case isAlpha(l.cur):
		l.keep()
		return lexSymbol, nil
	case isWhitespace(l.cur):
		l.emit(symbolToken)
		return lexRoot, nil
	}

	switch l.cur {
	case '_', '-':
		l.keep()
		return lexSymbol, nil
	case CONCATENATION_CHAR:
		l.emit(symbolToken)
		l.keep()
		l.emit(concatenationToken)
		return lexRoot, nil
	case COMMENT_CHAR:
		l.emit(symbolToken)
		return lexComment, nil
	case HASH_SEPARATOR_CHAR:
		l.emit(symbolToken)
		l.keep()
		l.emit(hashSeparatorToken)
		return lexRoot, nil
	case ELEM_SEPARATOR_CHAR:
		l.emit(symbolToken)
		l.keep()
		l.emit(elemSeparatorToken)
		return lexRoot, nil
	case ARGS_START_CHAR:
        l.emit(symbolToken)
        l.keep()
        l.emit(argsStartToken)
        return lexRoot, nil
	case LIST_START_CHAR:
        l.emit(symbolToken)
        l.keep()
        l.emit(listStartToken)
        return lexRoot, nil
	case HASH_START_CHAR:
        l.emit(symbolToken)
        l.keep()
        l.emit(hashStartToken)
		return lexRoot, nil
	}

	return nil, l.unexpectedChar("lexSymbol")
}

func (l *lexer) keep() {
	if l.buf == nil {
		l.buf = make([]rune, 0, 32)
	}
	l.buf = append(l.buf, l.cur)
}

func lex(input io.RuneReader, c chan token, e *error) {
	defer close(c)
	l := &lexer{input, nil, ' ', c, 0, 0}

	var err error
	f := stateFn(lexRoot)
	for {
		f, err = f(l)
		if err != nil {
			break
		}
		err = l.next()
		if err != nil {
			break
		}
	}

	switch t := err.(type) {
	case lexError:
		if t.inner != io.EOF {
			*e = t
		}
	default:
		if err != io.EOF {
			*e = newLexError(err.Error(), nil)
		}
	}
}

// lexes from an io.RuneReader until EOF, returning all tokens as a token
// slice.
func lexAll(input io.RuneReader) ([]token, error) {
	c := make(chan token)
	out := make([]token, 0, 32)
	var err error
	go lex(input, c, &err)
	for t := range c {
		out = append(out, t)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}
