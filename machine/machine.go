package machine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"

	"github.com/go-errors/errors"
)

type machine struct {
	Context *context

	// whether we hit a return statement and should exit
	returnFlag bool
	// whether to print out the ast and some other debugging stuff
	debugFlag bool
}

type MachineOpts func(*machine)

func MachineOptSetDebug(m *machine) {
	m.debugFlag = true
	m.Context = newContext(true)
}

func NewMachine(opts ...MachineOpts) *machine {
	m := &machine{
		Context:    newContext(false),
		returnFlag: false,
	}

	for _, o := range opts {
		o(m)
	}

	return m
}

func (m *machine) Reset() {
	m.Context.Reset()
	m.returnFlag = false
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

func (m *machine) AstNodeToNode(lit ast.Node) *Node {
	var val any
	switch n := lit.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.FLOAT, token.INT:
			val, _ = strconv.ParseFloat(n.Value, 64)
		case token.CHAR, token.STRING:
			val = n.Value
		}
		return &Node{
			Kind:  n.Kind,
			Value: val,
		}
	case *ast.FuncLit:
		return &Node{
			Kind:    token.FUNC,
			Value:   val,
			Context: m.Context,
		}
	}

	return nil
}

// Evaluate evaluates a node and produces a literal
func (m *machine) Evaluate(node ast.Node) (*Node, error) {
	var err error
	var res ast.Node
	var lit *Node

	switch n := node.(type) {
	case *ast.BasicLit:
		lit = m.AstNodeToNode(n)
	case *ast.FuncLit:
		lit = &Node{
			Kind:    token.FUNC,
			Value:   n,
			Context: m.Context,
		}
	case *ast.Ident:
		lit = m.Context.Get(n.Name)
		if lit == nil {
			return nil, errors.Errorf("cannot find identifier %s", n.Name)
		}
	case *ast.DeclStmt:
		err = m.evalDecl(n)
	case *ast.AssignStmt:
		err = m.evalAssign(n)
	case *ast.ParenExpr:
		lit, err = m.Evaluate(n.X)
	case *ast.ExprStmt:
		lit, err = m.Evaluate(n.X)
	case *ast.BinaryExpr:
		res, err = m.evalBinary(n)
		lit = m.AstNodeToNode(res)
	case *ast.UnaryExpr:
		res, err = m.evalUnary(n)
		lit = m.AstNodeToNode(res)
	case *ast.IncDecStmt:
		var tok token.Token
		if n.Tok == token.INC {
			tok = token.ADD_ASSIGN
		} else if n.Tok == token.DEC {
			tok = token.SUB_ASSIGN
		}
		err = m.evalAssign(&ast.AssignStmt{
			Lhs: []ast.Expr{n.X},
			Rhs: []ast.Expr{NewFloatLiteral(1)},
			Tok: tok,
		})
	case *ast.CallExpr:
		lit, err = m.evalFunctionCall(n.Fun, n.Args)
	case *ast.ReturnStmt:
		m.returnFlag = true
		lit, err = m.Evaluate(n.Results[0])
	case *ast.BlockStmt:
		for _, stmt := range n.List {
			lit, err = m.Evaluate(stmt)
			if m.returnFlag {
				break
			}
		}
	case *ast.IfStmt:
		lit, err = m.evalIf(n)
	case *ast.ForStmt:
		lit, err = m.evalFor(n)
	default:
		return nil, errors.Errorf("unknown type %v", reflect.TypeOf(node))
	}

	return lit, err
}

func (m *machine) evalAssign(stmt *ast.AssignStmt) error {
	name := stmt.Lhs[0].(*ast.Ident).Name
	switch stmt.Tok {
	case token.ADD_ASSIGN: // +=
		lit, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.ADD,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, m.AstNodeToNode(lit))
		if err != nil {
			return err
		}
	case token.SUB_ASSIGN: // -=
		lit, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.SUB,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, m.AstNodeToNode(lit))
		if err != nil {
			return err
		}
	case token.MUL_ASSIGN: // *=
		lit, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.MUL,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, m.AstNodeToNode(lit))
		if err != nil {
			return err
		}
	case token.QUO_ASSIGN: // /=
		lit, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.QUO,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, m.AstNodeToNode(lit))
		if err != nil {
			return err
		}
	case token.REM_ASSIGN: // %=
		lit, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.REM,
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			break
		}
		err = m.Context.Update(name, m.AstNodeToNode(lit))
		if err != nil {
			return err
		}
	case token.ASSIGN: // =
		lit, err := m.Evaluate(stmt.Rhs[0])
		if err != nil {
			return err
		}
		err = m.Context.Update(name, lit)
		if err != nil {
			return err
		}
	case token.DEFINE: // :=
		lit, err := m.Evaluate(stmt.Rhs[0])
		err = m.Context.Set(name, lit)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown assign token %s", stmt.Tok.String())
	}
	return nil
}

func (m *machine) evalDecl(n *ast.DeclStmt) error {
	decl := n.Decl.(*ast.GenDecl)
	for _, spec := range decl.Specs {
		s := spec.(*ast.ValueSpec)
		for i, name := range s.Names {
			res, err := m.Evaluate(s.Values[i])
			if err != nil {
				return err
			}
			err = m.Context.Set(name.Name, res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *machine) evalIf(n *ast.IfStmt) (*Node, error) {
	oldContext := m.Context
	m.Context = m.Context.NewChildContext()

	cond, err := m.Evaluate(n.Cond)
	if err != nil {
		return nil, errors.Errorf("cannot eval if cond %w", err)
	}

	var res *Node
	if isTruthy(cond) {
		res, err = m.Evaluate(n.Body)
	} else if n.Else != nil {
		res, err = m.Evaluate(n.Else)
	}

	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval if body", 10)
	}

	m.Context = oldContext
	return res, nil
}

func (m *machine) evalFor(n *ast.ForStmt) (*Node, error) {
	oldContext := m.Context
	m.Context = m.Context.NewChildContext()

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

		res, err = m.Evaluate(n.Body)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for body", 10)
		}

		if m.returnFlag {
			break
		}

		_, err = m.Evaluate(n.Post)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for post", 10)
		}
	}

	m.Context = oldContext
	return res, nil
}

func (m *machine) applyFunction(fun *Node, args []ast.Expr) (*Node, error) {
	oldContext := m.Context
	m.Context = fun.Context.NewChildContext()

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
			return nil, errors.Errorf("cannot eval arg %v", arg)
		}
		m.Context.Set(name, argNode)
	}

	// evaluate body
	var res *Node
	var err error
	for _, stmt := range n.Body.List {
		res, err = m.Evaluate(stmt)
		if err != nil {
			return nil, err
		}
		if m.returnFlag {
			break
		}
	}

	m.returnFlag = false
	m.Context = oldContext
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
		res, err = m.applyFunction(&Node{
			Kind:    token.FUNC,
			Value:   fun,
			Context: m.Context,
		}, args)

	default:
		return nil, errors.Errorf("unimplemented function type %s", n)
	}

	return res, err
}

func (m *machine) evalUnary(expr *ast.UnaryExpr) (*ast.BasicLit, error) {
	node, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	if node.Kind != token.INT && node.Kind != token.FLOAT {
		return nil, errors.Errorf("unsupported operand types %s", node.Kind)
	}

	operand := node.Value.(float64)
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

	return NewFloatLiteral(r), nil
}

func (m *machine) evalBinary(expr *ast.BinaryExpr) (*ast.BasicLit, error) {
	nodeX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	nodeY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}

	if (nodeX.Kind != token.INT && nodeX.Kind != token.FLOAT) ||
		(nodeY.Kind != token.INT && nodeY.Kind != token.FLOAT) {
		return nil, errors.Errorf(
			"unsupported operand types %s %s",
			nodeX.Kind,
			nodeY.Kind,
		)
	}

	operand1 := nodeX.Value.(float64)
	operand2 := nodeY.Value.(float64)
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
	case token.LOR:
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

	return NewFloatLiteral(r), nil
}

func isTruthyFloat(val float64) bool {
	if val > 0 {
		return true
	}
	return false
}

func isTruthy(node *Node) bool {
	switch node.Kind {
	case token.FLOAT, token.INT:
		val, _ := node.Value.(float64)
		return isTruthyFloat(val)
	case token.STRING:
		if node.Value.(string) == "" {
			return false
		}
		return true
	case token.FUNC:
		return true
	}
	return false
}
