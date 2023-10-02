package machine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-errors/errors"
)

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
	if node, ok := c.storage[name]; ok {
		if c.debugFlag {
			fmt.Println("get context", name, "=", node.Value)
			fmt.Println("context: ", c.String())
		}
		return node
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
			fmt.Println("update context", name, "<-", value.Value)
			fmt.Println("context: ", c.String())
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
		fmt.Println("set context", name, "<-", value.Value)
		fmt.Println("context: ", c.String())
	}
	return nil
}

func (c *context) String() string {
	strs := make([]string, 0)
	for k, v := range c.storage {
		strs = append(strs, fmt.Sprintf(
			"%s: %s",
			k,
			v.String(),
		))
	}

	sort.Slice(strs, func(i, j int) bool {
		return i < j
	})
	return strings.Join(strs, " | ")
}
