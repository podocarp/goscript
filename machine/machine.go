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
	res, err := parser.ParseExpr(stmt)
	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot parse", 10)
	}

	if m.debugFlag {
		fs := token.NewFileSet()
		ast.Print(fs, res)
	}

	return res, err
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

func (m *Machine) CallFunction(fun any, args []ast.Expr) (*Node, error) {
	if funNode, ok := fun.(*Node); ok {
		if funLit, ok := funNode.Value.(*ast.FuncLit); ok {
			return m.Evaluate(&ast.CallExpr{
				Fun:  funLit,
				Args: args,
			})
		} else {
			return nil, errors.New("the supplied argument is not a function.")
		}
	} else {
		return nil, errors.New("param \"fun\" must be the result of an Evaluate call.")
	}
}
