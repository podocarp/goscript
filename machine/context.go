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

type Node struct {
	Kind    token.Token
	Value   any
	Context *context
}

func (n *Node) GetFunc() (*ast.FuncLit, error) {
	if n.Kind != token.FUNC {
		return nil, errors.Errorf("value is not a function")
	}
	return n.Value.(*ast.FuncLit), nil
}

type context struct {
	Parent *context

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
	child := newContext(c.debugFlag)
	child.Parent = c
	return child
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
			fmt.Println("update context", name, value.Value, "context: ", c.String())
		}

		return nil
	}

	if c.Parent != nil {
		c.Parent.Update(name, value)
		return nil
	}

	return errors.Errorf("cannot find name %s to update to %v", name, value.Value)
}

func (c *context) Set(name string, value *Node) error {
	if _, ok := c.storage[name]; ok {
		return errors.Errorf("reassigning %s", name)
	}
	c.storage[name] = value

	if c.debugFlag {
		fmt.Println("set context", name, value.Value, "context: ", c.String())
	}
	return nil
}

func (c *context) String() string {
	strs := make([]string, 0)
	for k, v := range c.storage {
		switch v.Kind {
		case token.FUNC:
			strs = append(strs, fmt.Sprintf(
				"%s: λ",
				k,
			))
		default:
			strs = append(strs, fmt.Sprintf(
				"%s: %v",
				k,
				v.Value,
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
