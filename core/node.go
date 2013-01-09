package din

import (
	"bytes"
	"fmt"
)

// a node is an element in the routing language's parse tree
type node interface {
	Type() nodeType
	String() string
}

type nodeType int

// Type() for a nodeType returns itself to simplify embedding the noteType type
// inside of the node type.
func (n nodeType) Type() nodeType {
	return n
}

const (
	nodeAssignment  nodeType = iota // assignment operation
	nodeConstructor                 // constructs some object
	nodeKeyValue                    // represents a key-value pair
	nodeList                        // contains a list of other nodes
	nodeRouteSpec                   // a routespec constant, which is a route pattern, e.g. /users/{id:int}
	nodeString                      // a string constant
)

/*
{
    nodeType: nodeAssignment
    left: {
        nodeType: nodeRouteSpec,
        value: "/user/{id:int}",
    }
    right: {
        nodeType: nodeConstructor,
        typeName: "route",
        kvPairs: [
            {
                key: "name",
                value: "user_profile",
            },
            {
                key: "doc",
                value: "this is the documentation",
            },
            {
                key: "get",
                value: {
                    ...
                }
            },
        ]
    }
}
*/

type listNode struct {
	nodeType
	nodes []node
}

func newList() *listNode {
	return &listNode{nodeType: nodeList}
}

func (l *listNode) append(n node) {
	l.nodes = append(l.nodes, n)
}

func (l *listNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

type assignmentNode struct {
	nodeType
	left  node // needs to be stringNode or routeSpecNode
	right node
}

func newAssignment(left, right node) *assignmentNode {
	return &assignmentNode{nodeType: nodeAssignment, left: left, right: right}
}

func (a *assignmentNode) String() string {
	return fmt.Sprintf("%s = %s", a.left, a.right)
}

// constructorNode represents an operation to create a data object.
// Constructors themselves need to be registered with the routes language
// at runtime.
type constructorNode struct {
	nodeType
	typeName string
	kvPairs  []keyValueNode
}

// newConstructor creates a new constructorNode, and should be the One True Way
// in which constructorNodes are created.
func newConstructor(typeName string) *constructorNode {
	return &constructorNode{
		nodeType: nodeConstructor,
		typeName: typeName,
	}
}

// adds a key-value pair to the constructor
func (c *constructorNode) addPair(pair keyValueNode) {
	c.kvPairs = append(c.kvPairs, pair)
}

func (c *constructorNode) String() string {
	return fmt.Sprintf("%s{}", c.typeName)
}

// a keyValueNode contains a key value pair.  keyValueNode itself will not stop
// you from assigning invalid keys to the node.  The values are themselves
// nodes; keyValueNode itself will not stop you from putting a node type in an
// invalid context.
type keyValueNode struct {
	nodeType
	key   string
	value node
}

// newKeyValue creates keyValueNode nodes, and should be the One True Way in
// which keyValueNodes are created.
func newKeyValue(key string, value node) *keyValueNode {
	return &keyValueNode{
		nodeType: nodeKeyValue,
		key:      key,
		value:    value,
	}
}

func (kv *keyValueNode) String() string {
	return fmt.Sprintf("{%s: %s}", kv.key, kv.value)
}

// a routeSpecNode holds a routeSpec literal.  E.g., the string /users/{id:int}
// would be a valid routespec.
type routeSpecNode struct {
	nodeType
	value string
}

// newRouteSpec is responsible for creating route spec nodes, and should be the
// One True Way in which routeSpecNode structs are created.
func newRouteSpec(spec string) *routeSpecNode {
	return &routeSpecNode{
		nodeType: nodeRouteSpec,
		value:    spec,
	}
}

func (r *routeSpecNode) String() string {
	return fmt.Sprintf("route{%s}", r.value)
}

// stringNode holds a string construct.
type stringNode struct {
	nodeType
	text string
}

func newString(text string) *stringNode {
	return &stringNode{nodeType: nodeString, text: text}
}

func (s *stringNode) String() string {
	return s.text
}
