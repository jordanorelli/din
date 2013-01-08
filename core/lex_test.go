package din

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// I was going to put the test cases all in json but json doesn't allow real
// multiline strings, because json is fucking retarded, and XML is too fucking
// retarded for me to want to suffer, and YAML isn't in the standard library.
var lexTests = []struct {
	in      string
	outjson string
}{
	{"#a comment line", `[]`},
	{"name", `[["name", "symbol"]]`},
	{"name name", `[
        ["name", "symbol"],
        ["name", "symbol"]
    ]`},
	{`name
    name`, `[
        ["name", "symbol"],
        ["name", "symbol"]
    ]`},
	{`x = "some string"`, `[
        ["x", "symbol"],
        ["=", "assignment"],
        ["some string", "string"]
    ]`},
	{`(`, `[["(", "argsStart"]]`},
	{`[`, `[["[", "listStart"]]`},
	{`,`, `[[",", "elemSeparator"]]`},
	{`{`, `[["{", "hashStart"]]`},
	{`{key: "value"}`, `[
        ["{", "hashStart"],
        ["key", "symbol"],
        [":", "hashSeparator"],
        ["value", "string"],
        ["}", "hashEnd"]
    ]`},
	{`{ key: "value"}`, `[
        ["{", "hashStart"],
        ["key", "symbol"],
        [":", "hashSeparator"],
        ["value", "string"],
        ["}", "hashEnd"]
    ]`},
	{`{ key : "value"}`, `[
        ["{", "hashStart"],
        ["key", "symbol"],
        [":", "hashSeparator"],
        ["value", "string"],
        ["}", "hashEnd"]
    ]`},
	{`{key:"value"}`, `[
        ["{", "hashStart"],
        ["key", "symbol"],
        [":", "hashSeparator"],
        ["value", "string"],
        ["}", "hashEnd"]
    ]`},
	{`/`, `[
        ["/", "route"]
    ]`},
	{`/users/{id}`, `[
        ["/users/{id}", "route"]
    ]`},
    {`/help = template("help.html")`, `[
        ["/help", "route"],
        ["=", "assignment"],
        ["template", "symbol"],
        ["(", "argsStart"],
        ["help.html", "string"],
        [")", "argsEnd"]
    ]`},
	{`/users/{id:int}`, `[
        ["/users/{id:int}", "route"]
    ]`},
	{`/users/{id:int} = `, `[
        ["/users/{id:int}", "route"],
        ["=", "assignment"]
    ]`},
	{`/users/{id:int} = route {
        doc: "user profile",
        get: UserProfileHandler,
    }`, `[
        ["/users/{id:int}", "route"],
        ["=", "assignment"],
        ["route", "symbol"],
        ["{", "hashStart"],
        ["doc", "symbol"],
        [":", "hashSeparator"],
        ["user profile", "string"],
        [",", "elemSeparator"],
        ["get", "symbol"],
        [":", "hashSeparator"],
        ["UserProfileHandler", "symbol"],
        [",", "elemSeparator"],
        ["}", "hashEnd"]
    ]`},
	{`/users/{id:int} = route{
        "doc": "user profile",
        "get": UserProfileHandler,
    }`, `[
        ["/users/{id:int}", "route"],
        ["=", "assignment"],
        ["route", "symbol"],
        ["{", "hashStart"],
        ["doc", "string"],
        [":", "hashSeparator"],
        ["user profile", "string"],
        [",", "elemSeparator"],
        ["get", "string"],
        [":", "hashSeparator"],
        ["UserProfileHandler", "symbol"],
        [",", "elemSeparator"],
        ["}", "hashEnd"]
    ]`},
    {`/static = dir("staticfiles")`, `[
        ["/static", "route"],
        ["=", "assignment"],
        ["dir", "symbol"],
        ["(", "argsStart"],
        ["staticfiles", "string"],
        [")", "argsEnd"]
    ]`},
    {`files[
        "robots.txt",
        "humans.txt",
        "favicon.ico",
    ]`, `[
        ["files", "symbol"],
        ["[", "listStart"],
        ["robots.txt", "string"],
        [",", "elemSeparator"],
        ["humans.txt", "string"],
        [",", "elemSeparator"],
        ["favicon.ico", "string"],
        [",", "elemSeparator"],
        ["]", "listEnd"]
    ]`},
}

func match(tokens []token, testCase [][2]string) error {
	if len(tokens) != len(testCase) {
		return fmt.Errorf("mismatched lengths: wanted %d, got %d", len(testCase), len(tokens))
	}
	for i, token := range tokens {
		if testCase[i][0] != token.Lexeme {
			return errors.New("mismatched lexeme")
		}
		if testCase[i][1] != token.T.String() {
			return errors.New("mismatched token type")
		}
	}
	return nil
}

func parseTest(testString string) ([][2]string, error) {
	out := [][2]string{}
	if err := json.Unmarshal([]byte(testString), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func TestLexing(t *testing.T) {
	for i, spec := range lexTests {
		testCase, err := parseTest(spec.outjson)
		if err != nil {
			t.Errorf("FAIL %d: invalid test case: %v", i, err)
			continue
		}
		lexed, err := lexAll(strings.NewReader(spec.in + " "))
		if err != nil {
			t.Errorf("FAIL %d: bad input: %v", i, err)
			continue
		}
		if err := match(lexed, testCase); err != nil {
			t.Errorf("FAIL %d: %v\n%v != %v", i, err, lexed, spec.outjson)
		} else {
			t.Logf("PASS %d: %v == %v", i, lexed, spec.outjson)
		}
	}
}
