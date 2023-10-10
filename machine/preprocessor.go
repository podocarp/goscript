package machine

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
	"golang.org/x/tools/go/ast/astutil"
)

func (m *Machine) Preprocess(expr ast.Node) error {
	var err error
	astutil.Apply(
		expr,
		func(c *astutil.Cursor) bool {
			var newexpr ast.Node
			expr := c.Node()
			if m.debugFlag {
				fmt.Printf(
					"---Preprocessing %s\n", reflect.TypeOf(expr),
				)
			}
			switch n := expr.(type) {
			case *ast.BasicLit:
				newexpr, err = m.preprocessBasicLit(n)
				c.Replace(newexpr)
			case *ast.Ident:
				err = m.preprocessIdent(n)
			case *ast.FuncLit:
				err = m.preprocessFuncLit(n)
			}

			if m.debugFlag {
				fmt.Printf(
					"---Finished preproecssing %v, err: %v\n",
					reflect.TypeOf(expr),
					err,
				)
			}

			return err == nil
		},
		nil,
	)

	return err
}

// preprocessBasicLit converts literals into machine.Nodes beforehand
func (m *Machine) preprocessBasicLit(lit *ast.BasicLit) (ast.Node, error) {
	var node *Node
	switch lit.Kind {
	case token.FLOAT:
		val, _ := strconv.ParseFloat(lit.Value, 64)
		node = NewFloatNode(val)
	case token.INT:
		val, _ := strconv.ParseInt(lit.Value, 10, 64)
		node = NewIntNode(val)
	case token.CHAR, token.STRING:
		val, _ := strconv.Unquote(lit.Value)
		node = &Node{
			Type:  types.StringType,
			Value: val,
		}
	}

	return &ast.Ident{
		Name: "PREPROCESSED",
		Obj: &ast.Object{
			Data: node,
		},
	}, nil
}

func (m *Machine) preprocessIdent(lit *ast.Ident) error {
	switch lit.Name {
	case "true":
		lit.Obj = &ast.Object{
			Data: NewBoolNode(true),
		}
		return nil
	case "false":
		lit.Obj = &ast.Object{
			Data: NewBoolNode(false),
		}
		return nil
	}

	return nil
}

// flattenArgList makes things like (a,b,c float64) into
// (a float64, b float64, c float64)
func flattenArgList(fieldList []*ast.Field) ([]*ast.Field, error) {
	res := make([]*ast.Field, 0)
	for _, field := range fieldList {
		newField := *field
		if newField.Names == nil && newField.Type == nil {
			return nil, errors.New("field has undefined Names and Type")
		}

		if newField.Names == nil || len(newField.Names) == 0 {
			// in this case the Type element is used to store the
			// name of the field
			name := newField.Type.(*ast.Ident)
			newField.Names = []*ast.Ident{name}
			res = append(res, &newField)
			continue
		}

		if len(newField.Names) <= 1 {
			// it is already correct
			res = append(res, &newField)
			continue
		}

		for _, name := range newField.Names {
			newField := newField
			newField.Names = []*ast.Ident{name}
			res = append(res, &newField)
		}
	}

	return res, nil
}

// preprocessFuncLit preprocesses the argument list and makes it easier to
// traverse
func (m *Machine) preprocessFuncLit(lit *ast.FuncLit) error {
	params := lit.Type.Params
	fieldList, err := flattenArgList(params.List)
	if err != nil {
		return err
	}
	lit.Type.Params.List = fieldList
	return nil
}
