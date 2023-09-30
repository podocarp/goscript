
import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
)

type machine struct {
	context map[string]*ast.BasicLit

	// whether we hit a return statement and should exit
	returnFlag bool
	// whether to print out the ast and some other debugging stuff
	debugFlag bool
}

func NewMachine(debug bool) *machine {
	return &machine{
		context:   make(map[string]*ast.BasicLit),
		debugFlag: debug,
	}
}

func (m *machine) Reset() {
	clear(m.context)
	m.returnFlag = false
}

func (m *machine) ParseAndEval(stmt string) (*ast.BasicLit, error) {
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

func (m *machine) contextToString() string {
	strs := make([]string, 0)
	for k, v := range m.context {
		strs = append(strs, fmt.Sprintf(
			"%s: %v\t",
			k,
			v.Value,
		))
	}
	gslice.Sort(strs)
	return strings.Join(strs, " | ")
}

func (m *machine) set(name, value string) {
	m.context[name] = &ast.BasicLit{
		Kind:  token.FLOAT,
		Value: value,
	}
}

// Evaluate evaluates a node and produces a literal
func (m *machine) Evaluate(node ast.Node) (*ast.BasicLit, error) {
	var err error

	var lit *ast.BasicLit
	switch n := node.(type) {
	case *ast.BasicLit:
		lit = n
	case *ast.Ident:
		if val, ok := m.context[n.Name]; ok {
			lit = val
		} else {
			return nil, errors.Errorf("cannot find ident %s in context", n.Name)
		}
	case *ast.DeclStmt:
		m.evalDecl(n)
	case *ast.AssignStmt:
		lit, err = m.evalAssign(n)
	case *ast.ParenExpr:
		lit, err = m.Evaluate(n.X)
	case *ast.BinaryExpr:
		lit, err = m.evalBinary(n)
	case *ast.UnaryExpr:
		lit, err = m.evalUnary(n)
	case *ast.IncDecStmt:
		var tok token.Token
		if n.Tok == token.INC {
			tok = token.ADD
		} else if n.Tok == token.DEC {
			tok = token.SUB
		}
		lit, err = m.evalAssign(&ast.AssignStmt{
			Lhs: []ast.Expr{n.X},
			Rhs: []ast.Expr{&ast.BinaryExpr{
				X:  n.X,
				Op: tok,
				Y: &ast.BasicLit{
					Kind:  token.FLOAT,
					Value: "1",
				},
			}},
			Tok: token.ASSIGN,
		})
	case *ast.CallExpr:
		lit, err = m.evalFunctionCall(n.Fun, n.Args)
	case *ast.ReturnStmt:
		m.returnFlag = true
		lit, err = m.Evaluate(n.Results[0])
	case *ast.BlockStmt:
		for _, stmt := range n.List {
			lit, err = m.Evaluate(stmt)
		}
	case *ast.IfStmt:
		lit, err = m.evalIf(n)
	case *ast.ForStmt:
		lit, err = m.evalFor(n)
	}
	return lit, err
}

func (m *machine) evalAssign(stmt *ast.AssignStmt) (*ast.BasicLit, error) {
	var lit *ast.BasicLit
	var err error
	switch stmt.Tok {
	case token.ADD_ASSIGN: // +=
		lit, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.ADD,
			Y:  stmt.Rhs[0],
		})
	case token.SUB_ASSIGN: // -=
		lit, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.SUB,
			Y:  stmt.Rhs[0],
		})
	case token.MUL_ASSIGN: // *=
		lit, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.MUL,
			Y:  stmt.Rhs[0],
		})
	case token.QUO_ASSIGN: // /=
		lit, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.QUO,
			Y:  stmt.Rhs[0],
		})
	case token.REM_ASSIGN: // %=
		lit, err = m.evalBinary(&ast.BinaryExpr{
			X:  stmt.Lhs[0],
			Op: token.REM,
			Y:  stmt.Rhs[0],
		})
	case token.ASSIGN, token.DEFINE:
		lit, err = m.Evaluate(stmt.Rhs[0])
	default:
		err = errors.Errorf("unknown assign token %s", stmt.Tok.String())
	}

	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval assign", 10)
	}

	name := stmt.Lhs[0].(*ast.Ident).Name
	m.set(name, lit.Value)

	return nil, nil
}

func (m *machine) evalDecl(n *ast.DeclStmt) (*ast.BasicLit, error) {
	decl := n.Decl.(*ast.GenDecl)
	for _, spec := range decl.Specs {
		s := spec.(*ast.ValueSpec)
		for i, name := range s.Names {
			res, err := m.Evaluate(s.Values[i])
			if err != nil {
				return nil, errors.Wrap(err, 10)
			}
			m.set(name.Name, res.Value)
		}
	}

	return nil, nil
}

func (m *machine) evalIf(n *ast.IfStmt) (*ast.BasicLit, error) {
	var res *ast.BasicLit

	cond, err := m.Evaluate(n.Cond)
	if err != nil {
		return nil, errors.Errorf("cannot eval if cond %w", err)
	}

	condVal, _ := strconv.ParseFloat(cond.Value, 64)
	if isTruthy(condVal) {
		res, err = m.Evaluate(n.Body)
	} else {
		res, err = m.Evaluate(n.Else)
	}

	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval if body", 10)
	}

	return res, nil
}

func (m *machine) evalFor(n *ast.ForStmt) (*ast.BasicLit, error) {
	var res *ast.BasicLit
	_, err := m.Evaluate(n.Init)
	if err != nil {
		return nil, errors.WrapPrefix(err, "cannot eval for init block %w", 10)
	}

	for {
		cond, err := m.Evaluate(n.Cond)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for cond block %w", 10)
		}
		condVal, _ := strconv.ParseFloat(cond.Value, 64)
		if !isTruthy(condVal) {
			break
		}

		res, err = m.Evaluate(n.Body)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for body %w", 10)
		}

		_, err = m.Evaluate(n.Post)
		if err != nil {
			return nil, errors.WrapPrefix(err, "cannot eval for post %w", 10)
		}
	}

	return res, nil
}

func (m *machine) evalFunctionCall(fun ast.Expr, args []ast.Expr) (*ast.BasicLit, error) {
	var err error

	lit := new(ast.BasicLit)
	switch n := fun.(type) {
	case *ast.FuncLit:
		// populate arguments
		params := n.Type.Params
		for i, arg := range args {
			name := params.List[i].Names[0].Name
			if val, ok := arg.(*ast.BasicLit); ok {
				m.set(name, val.Value)
			} else {
				return nil, errors.Errorf("arg not literal %v", arg)
			}
		}

		// evaluate body
		for _, stmt := range n.Body.List {
			lit, err = m.Evaluate(stmt)
			if m.returnFlag {
				m.returnFlag = false
				break
			}
		}

	default:
		return nil, errors.Errorf("unimplemented function type %s", n)
	}

	return lit, err
}

func (m *machine) evalUnary(expr *ast.UnaryExpr) (*ast.BasicLit, error) {
	lit, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}

	if lit.Kind != token.INT && lit.Kind != token.FLOAT {
		return nil, errors.Errorf("unsupported operand types %s", lit.Kind)
	}

	operand, _ := strconv.ParseFloat(lit.Value, 64)
	res := new(ast.BasicLit)
	var r float64
	switch expr.Op {
	case token.SUB:
		r = -operand
	case token.NOT:
		if isTruthy(operand) {
			r = 0
		} else {
			r = 1
		}
	default:
		return nil, errors.New("Operation not supported")
	}

	res.Value = strconv.FormatFloat(float64(r), 'g', 4, 64)
	res.ValuePos = 0
	res.Kind = token.FLOAT
	return res, nil
}

func (m *machine) evalBinary(expr *ast.BinaryExpr) (*ast.BasicLit, error) {
	litX, err := m.Evaluate(expr.X)
	if err != nil {
		return nil, err
	}
	litY, err := m.Evaluate(expr.Y)
	if err != nil {
		return nil, err
	}

	if (litX.Kind != token.INT && litX.Kind != token.FLOAT) ||
		(litY.Kind != token.INT && litY.Kind != token.FLOAT) {
		return nil, errors.Errorf("unsupported operand types %s %s", litX.Kind, litY.Kind)
	}

	operand1, _ := strconv.ParseFloat(litX.Value, 64)
	operand2, _ := strconv.ParseFloat(litY.Value, 64)
	res := new(ast.BasicLit)
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
		if !isTruthy(operand1) {
			r = 0
		} else if !isTruthy(operand2) {
			r = 0
		} else {
			r = 1
		}
	case token.LOR:
		if isTruthy(operand1) {
			r = 1
		} else if isTruthy(operand2) {
			r = 1
		} else {
			r = 0
		}
	default:
		return nil, errors.New("Operation not supported")
	}

	res.Value = strconv.FormatFloat(float64(r), 'g', 4, 64)
	res.ValuePos = 0
	res.Kind = token.FLOAT
	return res, nil
}

func isTruthy(val float64) bool {
	if val > 0 {
		return true
	}
	return false
}
