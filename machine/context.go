package machine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-errors/errors"
)

type context struct {
	Parent *context
	Name   string

	storage map[string]*Node
}

func NewContext(name string) *context {
	return &context{
		storage: make(map[string]*Node),
		Name:    name,
	}
}

// NewChildContext creates a new child context for this context and returns the
// child context.
func (c *context) NewChildContext(name string) *context {
	child := NewContext(name)
	child.Parent = c
	return child
}

func (c *context) Reset() {
	clear(c.storage)
}

func (c *context) Get(name string) *Node {
	if node, ok := c.storage[name]; ok {
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
		return nil
	}

	if c.Parent != nil {
		c.Parent.Update(name, value)
		return nil
	}

	return errors.Errorf("cannot find name %s to update to %v", name, value.Value)
}

func (c *context) Set(name string, value *Node) error {
	c.storage[name] = value

	return nil
}

func (c *context) String() string {
	strs := []string{}
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
	return fmt.Sprintf("Context \"%s\": %s", c.Name, strings.Join(strs, " | "))

}
