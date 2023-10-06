package machine

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
)

// Evaluate evaluates a node and produces a literal
func (m *Machine) Evaluate(expr ast.Node) (*Node, error) {
	var err error
	var node *Node
	m.maxDepth--
	if m.maxDepth <= 0 {
		return nil, errors.New("evaluate depth exceeded")
	}

	if m.debugFlag {
		fmt.Printf(
			"---Evaluating %s\n", reflect.TypeOf(expr),
		)
	}

	switch n := expr.(type) {
	case *ast.BasicLit, *ast.FuncLit:
		node = m.evalLit(n)
	case *ast.CompositeLit:
		node, err = m.evalComposite(n)
	case *ast.Ident:
		node = m.Context.Get(n.Name)
		if node == nil {
			err = errors.Errorf("cannot find identifier %s", n.Name)
		}
	case *ast.IndexExpr:
		node, err = m.evalIndex(n)
	case *ast.DeclStmt:
		node, err = m.evalDecl(n)
	case *ast.AssignStmt:
		node, err = m.evalAssign(n)
	case *ast.ParenExpr:
		node, err = m.Evaluate(n.X)
	case *ast.ExprStmt:
		node, err = m.Evaluate(n.X)
	case *ast.BinaryExpr:
		node, err = m.evalBinary(n)
	case *ast.UnaryExpr:
		node, err = m.evalUnary(n)
	case *ast.IncDecStmt:
		var tok token.Token
		if n.Tok == token.INC {
			tok = token.ADD_ASSIGN
		} else if n.Tok == token.DEC {
			tok = token.SUB_ASSIGN
		} else {
			err = errors.New("impossible inc dec stmt!")
			break
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
		node.IsReturnValue = true
	default:
		err = errors.Errorf("unknown type %v", reflect.TypeOf(expr))
	}

	if m.debugFlag {
		fmt.Printf(
			"---Finished evaluating %v, result: %v, context: {%s}, err: %v\n",
			reflect.TypeOf(expr),
			node,
			m.Context.String(),
			err,
		)
	}

	m.maxDepth++
	return node, err
}

func (m *Machine) evalBlock(stmt *ast.BlockStmt) (*Node, error) {
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

func (m *Machine) evalAssign(stmt *ast.AssignStmt) (*Node, error) {
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
	case token.ASSIGN: // =
		node, err = m.Evaluate(stmt.Rhs[0])
		if err != nil {
			return nil, err
		}
		err = m.Context.Update(name, node)
	case token.DEFINE: // :=
		node, err = m.Evaluate(stmt.Rhs[0])
		if err != nil {
			return nil, err
		}
		err = m.Context.Set(name, node)
	default:
		err = errors.Errorf("unknown assign token %s", stmt.Tok.String())
	}

	return node, err
}

func (m *Machine) evalDecl(n *ast.DeclStmt) (*Node, error) {
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

func (m *Machine) evalIndex(lit *ast.IndexExpr) (*Node, error) {
	arrNode, err := m.Evaluate(lit.X)
	if err != nil {
		return nil, err
	}
	arr := arrNode.Value.([]*Node)
	indexNode, err := m.Evaluate(lit.Index)
	if err != nil {
		return nil, err
	}
	index := indexNode.Value.(Number)
	if !index.isIntegral() {
		err = errors.New("index is not an integer")
	}
	return arr[int(index)], nil
}

func (m *Machine) evalComposite(lit *ast.CompositeLit) (*Node, error) {
	switch n := lit.Type.(type) {
	case *ast.ArrayType:
		return m.evalArray(n.Elt, lit.Elts)
	default:
		return nil, errors.Errorf("unsupported composite type %v", n)
	}
}

// stringToType converts an ast.Ident.Name string into the corresponding type.
func stringToType(str string) (types.Type, error) {
	switch str {
	case "string":
		return types.StringType, nil
	case "float64":
		return types.FloatType, nil
	default:
		return nil, errors.Errorf("unknown type identifier %s", str)
	}
}

func (m *Machine) evalArray(elementType ast.Node, elems []ast.Expr) (*Node, error) {
	res := make([]*Node, 0, len(elems))
	var elemType types.Type
	var err error
	switch n := elementType.(type) {
	case *ast.Ident:
		elemType, err = stringToType(n.Name)
		if err != nil {
			return nil, err
		}
		// This is an array of elements that are not arrays.
		for _, elem := range elems {
			elemNode, err := m.Evaluate(elem)
			if err != nil {
				return nil, err
			}
			if !elemNode.Type.Equal(elemType) {
				return nil, errors.Errorf(
					"array type mismatch, element %v is not a %v",
					elemNode,
					elemType,
				)
			}
			res = append(res, elemNode)
		}
	case *ast.ArrayType:
		// Array of arrays. we have to eval each elem again
		for _, elem := range elems {
			elemLit := elem.(*ast.CompositeLit)
			elemNode, err := m.evalArray(n.Elt, elemLit.Elts)
			if err != nil {
				return nil, err
			}
			elemType = elemNode.Type
			if err != nil {
				return nil, err
			}
			res = append(res, elemNode)
		}
	}

	return &Node{
		Type:  types.ArrayOf(elemType),
		Value: res,
	}, nil
}

func (m *Machine) evalIf(n *ast.IfStmt) (*Node, error) {
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

func (m *Machine) evalFor(n *ast.ForStmt) (*Node, error) {
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

func FlattenFieldList(fieldList []*ast.Field) ([]*ast.Field, error) {
	res := make([]*ast.Field, 0)
	for _, field := range fieldList {
		if field.Names == nil && field.Type == nil {
			return nil, errors.New("field has undefined Names and Type")
		}

		if field.Names == nil {
			// in this case the Type element is used to store the
			// name of the field
			name := field.Type.(*ast.Ident)
			field.Names = []*ast.Ident{name}
			res = append(res, field)
			continue
		}

		if len(field.Names) <= 1 {
			// it is already correct
			res = append(res, field)
			continue
		}

		for _, name := range field.Names {
			newField := *field
			newField.Names = []*ast.Ident{name}
			res = append(res, &newField)
		}
	}

	return res, nil
}

func (m *Machine) applyFunction(fun *Node, args []*Node) (*Node, error) {
	m.Context = m.Context.NewChildContext("func block")
	var err error

	n := fun.Value.(*ast.FuncLit)
	// populate arguments
	params := n.Type.Params
	fieldList, err := FlattenFieldList(params.List)

	if err != nil {
		return nil, err
	}

	if len(fieldList) > len(args) {
		return nil, errors.New("not enough arguments to function")
	}

	for i, param := range fieldList {
		paramName := param.Names[0].Name
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval arg", 10)
		}
		m.Context.Set(paramName, args[i])
	}

	// evaluate body
	var res *Node
	res, err = m.Evaluate(n.Body)
	if err != nil {
		return nil, err
	}

	m.Context = m.Context.Parent
	return res, nil
}

func (m *Machine) evalFunctionCall(fun ast.Expr, args []ast.Expr) (*Node, error) {
	var err error
	nodeArgs := make([]*Node, len(args))
	for i, arg := range args {
		n, err := m.Evaluate(arg)
		if err != nil {
			return nil, errors.WrapPrefix(err, "error evaluating function arguments", 10)
		}
		nodeArgs[i] = n
	}

	funNode, err := m.Evaluate(fun)
	if err != nil {
		return nil, err
	}
	return m.applyFunction(funNode, nodeArgs)
}

func (m *Machine) evalUnary(expr *ast.UnaryExpr) (*Node, error) {
	node, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	if node.Type.Kind() != types.Float {
		return nil, errors.Errorf("unsupported operand types %v", node.Type)
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

func (m *Machine) evalBinary(expr *ast.BinaryExpr) (*Node, error) {
	nodeX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	nodeY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}

	if (nodeX.Type.Kind() != types.Float) || (nodeY.Type.Kind() != types.Float) {
		return nil, errors.Errorf(
			"unsupported operand types %v %v",
			nodeX.Type,
			nodeY.Type,
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

func (m *Machine) evalLit(lit ast.Node) *Node {
	switch n := lit.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.FLOAT, token.INT:
			return &Node{
				Type:  types.FloatType,
				Value: LiteralToNumber(n),
			}
		case token.CHAR, token.STRING:
			val, _ := strconv.Unquote(n.Value)
			return &Node{
				Type:  types.StringType,
				Value: val,
			}
		}
	case *ast.FuncLit:
		return &Node{
			Type:    types.FuncType,
			Value:   lit,
			Context: m.Context,
		}
	}

	return nil
}
