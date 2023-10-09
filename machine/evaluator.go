package machine

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
	"golang.org/x/exp/constraints"
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
	case *ast.AssignStmt:
		node, err = m.evalAssign(n)
	case *ast.BasicLit, *ast.FuncLit:
		node = m.evalLit(n)
	case *ast.BinaryExpr:
		node, err = m.evalBinary(n)
	case *ast.BlockStmt:
		node, err = m.evalBlock(n)
	case *ast.CallExpr:
		node, err = m.evalFunctionCall(n.Fun, n.Args)
	case *ast.CompositeLit:
		node, err = m.evalComposite(n)
	case *ast.DeclStmt:
		node, err = m.evalDecl(n)
	case *ast.ExprStmt:
		node, err = m.Evaluate(n.X)
	case *ast.ForStmt:
		node, err = m.evalFor(n)
	case *ast.IfStmt:
		node, err = m.evalIf(n)
	case *ast.Ident:
		node, err = m.evalIdent(n)
	case *ast.IndexExpr:
		node, err = m.evalIndex(n)
	case *ast.ParenExpr:
		node, err = m.Evaluate(n.X)
	case *ast.RangeStmt:
		node, err = m.evalRange(n)
	case *ast.UnaryExpr:
		node, err = m.evalUnary(n)
	case *ast.IncDecStmt:
		node, err = m.evalIncDec(n)
	case *ast.ReturnStmt:
		node, err = m.evalReturn(n)
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

func (m *Machine) evalIdent(lit *ast.Ident) (*Node, error) {
	switch lit.Name {
	case "true":
		return NewBoolNode(true), nil
	case "false":
		return NewBoolNode(false), nil
	}

	node := m.Context.Get(lit.Name)
	if node == nil {
		return nil, errors.Errorf("cannot find identifier %s", lit.Name)
	}
	return node, nil
}

func (m *Machine) evalLit(lit ast.Node) *Node {
	switch n := lit.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.FLOAT:
			val, _ := strconv.ParseFloat(n.Value, 64)
			return NewFloatNode(val)
		case token.INT:
			val, _ := strconv.ParseInt(n.Value, 10, 64)
			return NewIntNode(val)
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

func downgradeAssignArithmeticToken(tok token.Token) token.Token {
	switch tok {
	case token.ADD_ASSIGN:
		return token.ADD
	case token.SUB_ASSIGN:
		return token.SUB
	case token.MUL_ASSIGN:
		return token.MUL
	case token.QUO_ASSIGN:
		return token.QUO
	case token.REM_ASSIGN:
		return token.REM
	default:
		return -1
	}
}

func (m *Machine) evalAssign(stmt *ast.AssignStmt) (*Node, error) {
	switch stmt.Tok {
	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN,
		token.QUO_ASSIGN, token.REM_ASSIGN:
		if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
			return nil, errors.Errorf("syntax error at %v", stmt.Tok)
		}

		node, err := m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: downgradeAssignArithmeticToken(stmt.Tok),
			Y:  stmt.Rhs[0],
		})
		if err != nil {
			return nil, err
		}

		name := stmt.Lhs[0].(*ast.Ident).Name
		err = m.Context.Update(name, node)
		if err != nil {
			return nil, err
		}

	case token.ASSIGN, token.DEFINE:
		rhs := make([]*Node, 0, 1)
		for _, expr := range stmt.Rhs {
			node, err := m.Evaluate(expr)
			if err != nil {
				return nil, err
			}

			if node.Elems != nil {
				rhs = append(rhs, node.Elems...)
			} else {
				rhs = append(rhs, node)
			}
		}

		if len(stmt.Lhs) != len(rhs) {
			return nil, errors.Errorf(
				"assignment mismatch: %d variables on lhs but %d values on rhs",
				len(stmt.Lhs),
				len(rhs),
			)
		}

		for i := range stmt.Lhs {
			name := stmt.Lhs[i].(*ast.Ident).Name
			var err error
			if stmt.Tok == token.ASSIGN {
				err = m.Context.Update(name, rhs[i])
			} else {
				err = m.Context.Set(name, rhs[i])
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
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
	index, err := indexNode.ToInt()
	if err != nil {
		return nil, err
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
	case "float32", "float64":
		return types.FloatType, nil
	case "int", "int8", "int16", "int32", "int64":
		return types.IntType, nil
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
				// try to promote types
				if elemType.Kind() == types.Float {
					elemNode.Value, err = elemNode.ToFloat()
					elemNode.Type = types.FloatType
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.Errorf(
						"array type mismatch, element %v is not a %v",
						elemNode,
						elemType,
					)
				}
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
	m.Context = ifContext

	if n.Init != nil {
		_, err := m.Evaluate(n.Init)
		if err != nil {
			return nil, err
		}
	}

	cond, err := m.Evaluate(n.Cond)
	if err != nil {
		return nil, err
	}

	// context for the stuff in the if block
	blockContext := ifContext.NewChildContext("if block")
	m.Context = blockContext
	var res *Node
	if cond.Type.Kind() != types.Bool {
		return nil, errors.New("if condition evaluated to a non-boolean")
	}

	if cond.Value.(bool) {
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

func (m *Machine) evalRange(expr *ast.RangeStmt) (*Node, error) {
	// save context before for block
	oldContext := m.Context
	// context for the contents in the (...).
	forContext := oldContext.NewChildContext("for stmt")
	// context for the for block
	blockContext := forContext.NewChildContext("for block")

	m.Context = forContext
	rangeTarget, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	var res *Node
	switch rangeTarget.Type.Kind() {
	case types.Array:
		arr := rangeTarget.Value.([]*Node)
		for i, elem := range arr {
			m.Context = forContext
			if expr.Key != nil {
				val, _ := ValueToNode(i)
				name := expr.Key.(*ast.Ident).Name
				m.Context.Set(name, val)
			}
			if expr.Value != nil {
				name := expr.Value.(*ast.Ident).Name
				m.Context.Set(name, elem)
			}

			m.Context = blockContext
			res, err = m.Evaluate(expr.Body)
			if err != nil {
				return nil, errors.WrapPrefix(err, "cannot eval range body", 10)
			}

			if res != nil && res.IsReturnValue {
				break
			}
		}
	default:
		return nil, errors.Errorf("range not implemented for type %v", rangeTarget.Type)
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
		if cond.Type.Kind() != types.Bool {
			return nil, errors.New("for condition evaluated to a non-boolean")
		}
		if !(cond.Value.(bool)) {
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

func flattenArgList(fieldList []*ast.Field) ([]*ast.Field, error) {
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
	fieldList, err := flattenArgList(params.List)

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
	funNode, err := m.Evaluate(fun)
	if err != nil {
		return nil, err
	}

	if funNode.Type.Kind() == types.Builtin {
		return m.CallBuiltin(funNode, args)
	}

	nodeArgs := make([]*Node, len(args))
	for i, arg := range args {
		n, err := m.Evaluate(arg)
		if err != nil {
			return nil, errors.WrapPrefix(err, "error evaluating function arguments", 10)
		}
		nodeArgs[i] = n
	}

	return m.applyFunction(funNode, nodeArgs)
}

func (m *Machine) evalUnary(expr *ast.UnaryExpr) (*Node, error) {
	node, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case token.SUB:
		switch node.Type.Kind() {
		case types.Float:
			return NewFloatNode(-node.Value.(float64)), nil
		case types.Int:
			return NewIntNode(-node.Value.(int64)), nil
		case types.Uint:
			return NewUintNode(-node.Value.(uint64)), nil
		default:
			return nil, errors.Errorf("unsupported operand types %v", node.Type)
		}
	case token.NOT:
		if node.Type.Kind() != types.Bool {
			return nil, errors.Errorf("unsupported operand types %v", node.Type)
		}
		return NewBoolNode(!node.Value.(bool)), nil
	default:
		return nil, errors.New("operation not supported")
	}
}

func (m *Machine) evalIncDec(expr *ast.IncDecStmt) (*Node, error) {
	// translate into arithmetic
	var tok token.Token
	if expr.Tok == token.INC {
		tok = token.ADD_ASSIGN
	} else if expr.Tok == token.DEC {
		tok = token.SUB_ASSIGN
	} else {
		return nil, errors.New("impossible inc dec stmt!")
	}
	return m.Evaluate(&ast.AssignStmt{
		Lhs: []ast.Expr{expr.X},
		Rhs: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.INT,
				Value: "1",
			},
		},
		Tok: tok,
	})
}

func (m *Machine) evalBool(expr *ast.BinaryExpr) (*Node, error) {
	nodeX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	if nodeX.Type.Kind() != types.Bool {
		return nil, errors.Errorf(
			"left operand type %v is not boolean",
			nodeX.Type,
		)
	}

	// short circuit
	if expr.Op == token.LOR && nodeX.Value.(bool) {
		return NewBoolNode(true), nil
	} else if expr.Op == token.LAND && !nodeX.Value.(bool) {
		return NewBoolNode(false), nil
	}

	nodeY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}
	if nodeY.Type.Kind() != types.Bool {
		return nil, errors.Errorf(
			"right operand type %v is not boolean",
			nodeX.Type,
		)
	}

	if expr.Op == token.LOR {
		return NewBoolNode(nodeY.Value.(bool)), nil
	} else if expr.Op == token.LAND {
		return NewBoolNode(nodeY.Value.(bool)), nil
	} else {
		return nil, errors.Errorf("impossible boolean op %v", expr.Op)
	}
}

type Numeric interface {
	constraints.Integer | constraints.Float
}

func binop[T Numeric](op token.Token, operand1, operand2 T) T {
	switch op {
	case token.ADD: // +
		return operand1 + operand2
	case token.SUB: // -
		return operand1 - operand2
	case token.MUL: // *
		return operand1 * operand2
	case token.QUO: // /
		return operand1 / operand2
	case token.REM: // %
		return T(int64(operand1) % int64(operand2))
	default:
		return 0
	}
}

func bincomp[T Numeric](op token.Token, operand1, operand2 T) bool {
	switch op {
	case token.GTR: // >
		return operand1 > operand2
	case token.GEQ: // >
		return operand1 > operand2
	case token.LSS: // <
		return operand1 < operand2
	case token.LEQ: // <=
		return operand1 <= operand2
	case token.EQL: // ==
		return operand1 == operand2
	case token.NEQ: // !=
		return operand1 != operand2
	default:
		return false
	}
}

func (m *Machine) evalBinary(expr *ast.BinaryExpr) (*Node, error) {
	if expr.Op == token.LAND || expr.Op == token.LOR {
		return m.evalBool(expr)
	}

	nodeX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	nodeY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}

	if !nodeX.Type.Kind().IsNumeric() || !nodeY.Type.Kind().IsNumeric() {
		return nil, errors.Errorf(
			"unsupported operand type %v %v",
			nodeX.Type,
			nodeY.Type,
		)
	}

	switch expr.Op {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		if nodeX.Type.Kind() == types.Float || nodeY.Type.Kind() == types.Float {
			operand1, err := nodeX.ToFloat()
			if err != nil {
				return nil, err
			}
			operand2, err := nodeY.ToFloat()
			if err != nil {
				return nil, err
			}
			return NewFloatNode(binop(expr.Op, operand1, operand2)), nil
		} else if nodeX.Type.Kind() == types.Int && nodeY.Type.Kind() == types.Int {
			operand1 := nodeX.Value.(int64)
			operand2 := nodeY.Value.(int64)
			return NewIntNode(binop(expr.Op, operand1, operand2)), nil
		} else {
			return nil, errors.Errorf("unsupported types %v %v", nodeX.Type, nodeY.Type)
		}
	case token.GTR, token.GEQ, token.LSS, token.LEQ, token.EQL, token.NEQ:
		if nodeX.Type.Kind() == types.Float || nodeY.Type.Kind() == types.Float {
			operand1, err := nodeX.ToFloat()
			if err != nil {
				return nil, err
			}
			operand2, err := nodeY.ToFloat()
			if err != nil {
				return nil, err
			}
			return NewBoolNode(bincomp(expr.Op, operand1, operand2)), nil
		} else if nodeX.Type.Kind() == types.Int && nodeY.Type.Kind() == types.Int {
			operand1 := nodeX.Value.(int64)
			operand2 := nodeY.Value.(int64)
			return NewBoolNode(bincomp(expr.Op, operand1, operand2)), nil
		} else {
			return nil, errors.Errorf("unsupported types %v %v", nodeX.Type, nodeY.Type)
		}
	default:
		return nil, errors.New("Operation not supported")
	}
}

func (m *Machine) evalReturn(expr *ast.ReturnStmt) (*Node, error) {
	switch len(expr.Results) {
	case 0:
		return &Node{}, nil
	case 1:
		node, err := m.Evaluate(expr.Results[0])
		node.IsReturnValue = true
		return node, err
	default:
		node := &Node{}
		node.IsReturnValue = true
		node.Elems = make([]*Node, len(expr.Results))
		for i := range expr.Results {
			resultNode, err := m.Evaluate(expr.Results[i])
			if err != nil {
				return nil, err
			}
			node.Elems[i] = resultNode
		}
		return node, nil
	}
}
