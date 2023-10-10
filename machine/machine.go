package machine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"

	"github.com/go-errors/errors"
)

type Machine struct {
	Context *context

	// whether to print out the ast and some other debugging stuff
	debugFlag bool
	maxDepth  int
}

type MachineOpt func(*Machine)

func MachineOptSetDebug(m *Machine) {
	m.debugFlag = true
}

func MachineOptSetMaxDepth(maxdepth int) MachineOpt {
	return func(m *Machine) {
		m.maxDepth = maxdepth
	}
}

func NewMachine(opts ...MachineOpt) *Machine {
	m := &Machine{
		Context:  NewContext("global"),
		maxDepth: math.MaxInt,
	}
	m.AddBuiltinsToContext()

	for _, o := range opts {
		o(m)
	}

	return m
}

func (m *Machine) ParseAndEval(stmt string) (*Node, error) {
	node, err := m.Parse(stmt)
	if err != nil {
		return nil, err
	}
	return m.Evaluate(node)
}

func (m *Machine) Parse(stmt string) (ast.Node, error) {
	parsed, err := parser.ParseExpr(stmt)
	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot parse", 10)
	}

	err = m.Preprocess(parsed)
	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot preprocess", 10)
	}

	if m.debugFlag {
		ast.Print(token.NewFileSet(), parsed)
	}

	return parsed, err
}

// AddToGlobalContext adds a variable to the global context. Similar to running
// the expression name := val before any script is executed.
func (m *Machine) AddToGlobalContext(name string, val any) error {
	node, err := ValueToNode(val)
	if err != nil {
		return err
	}
	return m.Context.Set(name, node)
}

func (m *Machine) CallFunction(fun *Node, args []*Node) (*Node, error) {
	if _, ok := fun.Value.(*ast.FuncLit); ok {
		return m.applyFunction(fun, args)
	} else {
		return nil, errors.New(
			"the supplied function should be a result from calling Evaluate.",
		)
	}
}
