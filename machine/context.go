package machine

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
)

type Kind string

const (
	ELEM_FUN Kind = "FUNC"
	ELEM_LIT Kind = "LIT"
)

type Node struct {
	Node    ast.Node
	Kind    Kind
	Context *context
}

func LitToNode(lit *ast.BasicLit) *Node {
	return &Node{
		Node: lit,
		Kind: ELEM_LIT,
	}
}

type context struct {
	Parent *context
	Child  *context

	storage map[string]*Node

	debugFlag bool
}

func newContext(debug bool) *context {
	return &context{
		storage:   make(map[string]*Node),
		debugFlag: debug,
	}
}

// NewChildContext creates a new child context for this context and returns the
// child context.
func (c *context) NewChildContext() *context {
	c.Child = newContext(c.debugFlag)
	c.Child.Parent = c
	return c.Child
}

func (c *context) Reset() {
	clear(c.storage)
}

func (c *context) Get(name string) *Node {
	if lit, ok := c.storage[name]; ok {
		return lit
	}

	if c.Parent != nil {
		return c.Parent.Get(name)
	}

	return nil
}

func (c *context) Update(name string, value *Node) error {
	if _, ok := c.storage[name]; ok {
		c.storage[name] = value
		if c.debugFlag {
			fmt.Println("update context", name, value.Node, "context: ", c.String())
		}

		return nil
	}

	if c.Parent != nil {
		c.Parent.Update(name, value)
		return nil
	}

	return errors.Errorf("cannot find name %s to update to %v", name, value.Node)
}

func (c *context) Set(name string, value *Node) error {
	if _, ok := c.storage[name]; ok {
		return errors.Errorf("reassigning %s", name)
	}
	c.storage[name] = value

	if c.debugFlag {
		fmt.Println("set context", name, value.Node, "context: ", c.String())
	}
	return nil
}

func (c *context) String() string {
	strs := make([]string, 0)
	for k, v := range c.storage {
		switch n := v.Node.(type) {
		case *ast.BasicLit:
			strs = append(strs, fmt.Sprintf(
				"%s: %v",
				k,
				n.Value,
			))
		case *ast.FuncLit:
			strs = append(strs, fmt.Sprintf(
				"%s: λ",
				k,
			))
		}
	}

	sort.Slice(strs, func(i, j int) bool {
		return i < j
	})
	return strings.Join(strs, " | ")
}

func NewFloatLiteral(val float64) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.FLOAT,
		Value: strconv.FormatFloat(val, 'g', 4, 64),
	}
}
