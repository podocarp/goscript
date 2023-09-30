package machine

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"
)

type context struct {
	Parent *context
	Child  *context

	literals map[string]*ast.BasicLit
}

func NewContext() *context {
	return &context{
		literals: make(map[string]*ast.BasicLit),
	}
}

// NewChildContext creates a new child context for this context and returns the
// child context.
func (c *context) NewChildContext() *context {
	c.Child = NewContext()
	c.Child.Parent = c
	return c.Child
}

func (c *context) Reset() {
	clear(c.literals)
}

func (c *context) FindLit(name string) *ast.BasicLit {
	if lit, ok := c.literals[name]; ok {
		return lit
	}

	if c.Parent != nil {
		return c.Parent.FindLit(name)
	}

	return nil
}

func (c *context) SetLit(name string, value *ast.BasicLit) {
	c.literals[name] = value
}

func (c *context) String() string {
	strs := make([]string, 0)
	for k, v := range c.literals {
		strs = append(strs, fmt.Sprintf(
			"%s: %v",
			k,
			v.Value,
		))
	}
	sort.Slice(strs, func(i, j int) bool {
		return i < j
	})
	return strings.Join(strs, " | ")
}
