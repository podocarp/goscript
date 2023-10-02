package machine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/kind"
)

type machine struct {
	Context *context

	// whether to print out the ast and some other debugging stuff
	debugFlag bool
	maxDepth  int
}

type MachineOpt func(*machine)

func MachineOptSetDebug(m *machine) {
	m.debugFlag = true
}

func MachineOptSetMaxDepth(maxdepth int) MachineOpt {
	return func(m *machine) {
		m.maxDepth = maxdepth
	}
}

func NewMachine(opts ...MachineOpt) *machine {
	m := &machine{
		Context:  NewContext("global"),
		maxDepth: math.MaxInt,
	}

	for _, o := range opts {
		o(m)
	}

	return m
}

func (m *machine) ParseAndEval(stmt string) (*Node, error) {
	node, err := m.Parse(stmt)
	if err != nil {
		return nil, err
	}
	return m.Evaluate(node)
}

func (m *machine) Parse(stmt string) (ast.Node, error) {
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

func (m *machine) CallFunction(fun *ast.FuncLit, args []ast.Expr) (*Node, error) {
	return m.Evaluate(&ast.CallExpr{
		Fun:  fun,
		Args: args,
	})
}

// Evaluate evaluates a node and produces a literal
func (m *machine) Evaluate(expr ast.Node) (*Node, error) {
	var err error
	var node *Node
	m.maxDepth--
	if m.maxDepth <= 0 {
		return nil, errors.New("evaluate depth exceeded")
	}

	if m.debugFlag {
		fmt.Printf(
			"---evaluating %s\n", reflect.TypeOf(expr),
		)
		ast.Print(token.NewFileSet(), expr)
	}

	switch n := expr.(type) {
	case *ast.BasicLit, *ast.FuncLit:
		node = m.evalLit(n)
	case *ast.CompositeLit:
		node, err = m.evalComposite(n)
		if err != nil {
			return nil, err
		}
	case *ast.Ident:
		node = m.Context.Get(n.Name)
		if node == nil {
			return nil, errors.Errorf("cannot find identifier %s", n.Name)
		}
	case *ast.IndexExpr:
		arrNode, err := m.Evaluate(n.X)
		if err != nil {
			return nil, err
		}
		arr := arrNode.Value.([]*Node)
		indexNode, err := m.Evaluate(n.Index)
		if err != nil {
			return nil, err
		}
		index := indexNode.Value.(Number)
		if !index.isIntegral() {
			return nil, errors.New("index is not an integer")
		}
		node = arr[int(index)]
	case *ast.DeclStmt:
		node, err = m.evalDecl(n)
		if err != nil {
			return nil, err
		}
	case *ast.AssignStmt:
		node, err = m.evalAssign(n)
		if err != nil {
			return nil, err
		}
	case *ast.ParenExpr:
		node, err = m.Evaluate(n.X)
		if err != nil {
			return nil, err
		}
	case *ast.ExprStmt:
		node, err = m.Evaluate(n.X)
		if err != nil {
			return nil, err
		}
	case *ast.BinaryExpr:
		node, err = m.evalBinary(n)
		if err != nil {
			return nil, err
		}
	case *ast.UnaryExpr:
		node, err = m.evalUnary(n)
		if err != nil {
			return nil, err
		}
	case *ast.IncDecStmt:
		var tok token.Token
		if n.Tok == token.INC {
			tok = token.ADD_ASSIGN
		} else if n.Tok == token.DEC {
			tok = token.SUB_ASSIGN
		}
		node, err = m.evalAssign(&ast.AssignStmt{
			Lhs: []ast.Expr{n.X},
			Rhs: []ast.Expr{Number(1).ToLiteral()},
			Tok: tok,
		})
	case *ast.CallExpr:
		node, err = m.evalFunctionCall(n.Fun, n.Args)
	case *ast.BlockStmt:
		node, err = m.evalBlock(n)
	case *ast.IfStmt:
		node, err = m.evalIf(n)
	case *ast.ForStmt:
		node, err = m.evalFor(n)
	case *ast.ReturnStmt:
		node, err = m.Evaluate(n.Results[0])
		if err != nil {
			return nil, err
		}
		node.IsReturnValue = true
	default:
		return nil, errors.Errorf("unknown type %v", reflect.TypeOf(expr))
	}

	if m.debugFlag {
		fmt.Printf(
			"%s, evaluation result: %v\n", m.Context.String(), node,
		)
	}

	m.maxDepth++
	return node, err
}

func (m *machine) evalBlock(stmt *ast.BlockStmt) (*Node, error) {
	var res *Node
	var err error
	for _, stmt := range stmt.List {
		res, err = m.Evaluate(stmt)
		if err != nil {
			return nil, err
		}
		if res != nil && res.IsReturnValue {
			break
		}
	}

	return res, nil
}

func (m *machine) evalAssign(stmt *ast.AssignStmt) (*Node, error) {
	name := stmt.Lhs[0].(*ast.Ident).Name
	var node *Node
	var err error
	switch stmt.Tok {
	case token.ADD_ASSIGN: // +=
		node, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.ADD,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.SUB_ASSIGN: // -=
		node, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.SUB,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.MUL_ASSIGN: // *=
		node, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.MUL,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.QUO_ASSIGN: // /=
		node, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.QUO,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.REM_ASSIGN: // %=
		node, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.REM,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.ASSIGN: // =
		node, err = m.Evaluate(stmt.Rhs[0])
		if err != nil {
			return nil, err
		}
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}
	case token.DEFINE: // :=
		node, err = m.Evaluate(stmt.Rhs[0])
		err = m.Context.Set(name, node)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unknown assign token %s", stmt.Tok.String())
	}

	return node, err
}

func (m *machine) evalDecl(n *ast.DeclStmt) (*Node, error) {
	decl := n.Decl.(*ast.GenDecl)
	var res *Node
	for _, spec := range decl.Specs {
		s := spec.(*ast.ValueSpec)
		for i, name := range s.Names {
			res, err := m.Evaluate(s.Values[i])
			if err != nil {
				return nil, err
			}
			err = m.Context.Set(name.Name, res)
			if err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

func (m *machine) evalComposite(lit *ast.CompositeLit) (*Node, error) {
	switch n := lit.Type.(type) {
	case *ast.ArrayType:
		return m.createArray(n.Elt, lit.Elts)
	}
	return nil, nil
}

func (m *machine) createArray(elt ast.Node, elems []ast.Expr) (*Node, error) {
	value := make([]*Node, 0)
	switch n := elt.(type) {
	case *ast.Ident:
		for _, elem := range elems {
			elemNode, err := m.Evaluate(elem)
			if err != nil {
				return nil, err
			}
			value = append(value, elemNode)
		}
	case *ast.ArrayType:
		// array of arrays. we have to eval each elem again
		for _, elem := range elems {
			elemLit := elem.(*ast.CompositeLit)
			elemNode, err := m.createArray(n.Elt, elemLit.Elts)
			if err != nil {
				return nil, err
			}
			value = append(value, elemNode)
		}
	}

	return &Node{
		Kind:  kind.ARRAY,
		Value: value,
	}, nil
}

func (m *machine) evalIf(n *ast.IfStmt) (*Node, error) {
	// save machine context
	oldContext := m.Context
	// context for the stuff in (...)
	ifContext := oldContext.NewChildContext("if stmt")
	// context for the stuff in the if block
	blockContext := ifContext.NewChildContext("if block")

	m.Context = ifContext
	cond, err := m.Evaluate(n.Cond)
	if err != nil {
		return nil, errors.Errorf("cannot eval if cond %w", err)
	}

	m.Context = blockContext
	var res *Node
	if isTruthy(cond) {
		res, err = m.Evaluate(n.Body)
	} else if n.Else != nil {
		res, err = m.Evaluate(n.Else)
	}

	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval if", 10)
	}

	m.Context = oldContext
	return res, nil
}

func (m *machine) evalFor(n *ast.ForStmt) (*Node, error) {
	// save context before for block
	oldContext := m.Context
	// context for the contents in the (...).
	forContext := oldContext.NewChildContext("for stmt")
	// context for the for block
	blockContext := forContext.NewChildContext("for block")

	m.Context = forContext
	_, err := m.Evaluate(n.Init)
	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval for init block", 10)
	}

	var res *Node
	for {
		cond, err := m.Evaluate(n.Cond)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for cond block", 10)
		}
		if !isTruthy(cond) {
			break
		}

		m.Context = blockContext
		res, err = m.Evaluate(n.Body)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for body", 10)
		}

		if res != nil && res.IsReturnValue {
			break
		}

		m.Context = forContext
		_, err = m.Evaluate(n.Post)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for post", 10)
		}
	}

	m.Context = oldContext
	return res, nil
}

func (m *machine) applyFunction(fun *Node, args []ast.Expr) (*Node, error) {
	m.Context = m.Context.NewChildContext("func block")

	n := fun.Value.(*ast.FuncLit)
	// populate arguments
	params := n.Type.Params
	for i, arg := range args {
		var name string
		if ident, ok := params.List[i].Type.(*ast.Ident); ok {
			name = ident.Name
		} else {
			return nil, errors.New("cannot parse function param names")
		}

		argNode, err := m.Evaluate(arg)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval arg", 10)
		}
		m.Context.Set(name, argNode)
	}

	// evaluate body
	var res *Node
	var err error
	res, err = m.Evaluate(n.Body)
	if err != nil {
		return nil, err
	}

	m.Context = m.Context.Parent
	return res, nil
}

func (m *machine) evalFunctionCall(fun ast.Expr, args []ast.Expr) (*Node, error) {
	var err error
	var res *Node

	switch n := fun.(type) {
	case *ast.Ident:
		fun := m.Context.Get(n.Name)
		res, err = m.applyFunction(fun, args)
	case *ast.FuncLit:
		res, err = m.applyFunction(m.evalLit(fun), args)

	default:
		return nil, errors.Errorf("unimplemented function type %s", n)
	}

	return res, err
}

func (m *machine) evalUnary(expr *ast.UnaryExpr) (*Node, error) {
	node, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	if node.Kind != kind.FLOAT {
		return nil, errors.Errorf("unsupported operand types %v", node.Kind)
	}

	operand := node.Value.(Number).ToFloat()
	var r float64
	switch expr.Op {
	case token.SUB:
		r = -operand
	case token.NOT:
		if isTruthyFloat(operand) {
			r = 0
		} else {
			r = 1
		}
	default:
		return nil, errors.New("Operation not supported")
	}

	return Number(r).ToNode(), nil
}

func (m *machine) evalBinary(expr *ast.BinaryExpr) (*Node, error) {
	nodeX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	nodeY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}

	if (nodeX.Kind != kind.FLOAT) || (nodeY.Kind != kind.FLOAT) {
		return nil, errors.Errorf(
			"unsupported operand types %v %v",
			nodeX.Kind,
			nodeY.Kind,
		)
	}

	operand1 := nodeX.Value.(Number).ToFloat()
	operand2 := nodeY.Value.(Number).ToFloat()
	var r float64
	switch expr.Op {
	case token.ADD: // +
		r = operand1 + operand2
	case token.SUB: // -
		r = operand1 - operand2
	case token.MUL: // *
		r = operand1 * operand2
	case token.QUO: // /
		r = operand1 / operand2
	case token.REM: // %
		r = float64(int64(operand1) % int64(operand2))
	case token.GTR: // >
		if operand1 > operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.GEQ: // >=
		if operand1 >= operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.LSS: // <
		if operand1 < operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.LEQ: // <=
		if operand1 <= operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.EQL: // ==
		if operand1 == operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.NEQ: // !=
		if operand1 != operand2 {
			r = 1
		} else {
			r = 0
		}
	case token.LAND: // &&
		if !isTruthyFloat(operand1) {
			r = 0
		} else if !isTruthyFloat(operand2) {
			r = 0
		} else {
			r = 1
		}
	case token.LOR: // ||
		if isTruthyFloat(operand1) {
			r = 1
		} else if isTruthyFloat(operand2) {
			r = 1
		} else {
			r = 0
		}
	default:
		return nil, errors.New("Operation not supported")
	}

	return Number(r).ToNode(), nil
}

func (m *machine) evalLit(lit ast.Node) *Node {
	switch n := lit.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.FLOAT, token.INT:
			return &Node{
				Kind:  kind.FLOAT,
				Value: LiteralToNumber(n),
			}
		case token.CHAR, token.STRING:
			val, _ := strconv.Unquote(n.Value)
			return &Node{
				Kind:  kind.STRING,
				Value: val,
			}
		}
	case *ast.FuncLit:
		return &Node{
			Kind:    kind.FUNC,
			Value:   lit,
			Context: m.Context,
		}
	}

	return nil
}
