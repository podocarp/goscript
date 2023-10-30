package machine

import (
	"go/ast"
	"reflect"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
)

type builtin struct {
	fun func(m *Machine, args any) (*Node, error)

	// should the arguments be `Evaluate`d before feeding into `fun`. If
	// true, the array `fun` receives will be of type []*Node. If false,
	// the array `fun` receives will be of type []ast.Expr.
	evalArgs bool
}

func (m *Machine) AddBuiltinsToContext() {
	var builtins = map[string]*builtin{
		"append": {
			fun:      Append,
			evalArgs: true,
		},
		"len": {
			fun:      Len,
			evalArgs: true,
		},
		"make": {
			fun:      Make,
			evalArgs: false,
		},
	}

	for name, val := range builtins {
		m.Context.Set(name, &Node{
			Type:  types.BuiltinType,
			Value: val,
		})
	}
}

func (m *Machine) CallBuiltin(fun *Node, args []ast.Expr) (*Node, error) {
	builtin := fun.Value.(*builtin)
	if builtin.evalArgs {
		nodeArgs := make([]*Node, len(args))
		for i, arg := range args {
			n, err := m.Evaluate(arg)
			if err != nil {
				return nil, errors.WrapPrefix(err, "error evaluating function arguments", 10)
			}
			nodeArgs[i] = n
		}
		return builtin.fun(m, nodeArgs)
	} else {
		return builtin.fun(m, args)
	}
}

// append(s []T, vs ...T) []T
func Append(_ *Machine, a any) (*Node, error) {
	args := a.([]*Node)
	arr := args[0]
	vals := args[1:]
	newValue := arr.Value.([]*Node)
	newValue = append(newValue, vals...)
	arr.Value = newValue
	return arr, nil
}

func Len(_ *Machine, a any) (*Node, error) {
	args := a.([]*Node)
	arg := args[0]
	var res int
	switch arg.Type.Kind() {
	case types.String:
		res = len(arg.Value.(string))
	case types.Array:
		res = len(arg.Value.([]*Node))
	default:
		return nil, errors.Errorf("unsupported type %v for len", arg.Type)
	}

	return NewIntNode(int64(res)), nil
}

// make(t Type, size ...IntegerType) Type
func Make(m *Machine, a any) (*Node, error) {
	args := a.([]ast.Expr)

	typeInfo := args[0]
	switch n := typeInfo.(type) {
	case *ast.ArrayType:
		switch len(args) {
		case 1:
			return m.evalArray(n.Elt, []ast.Expr{})
		case 2:
			sizeNode := args[1]

			size, err := m.Evaluate(sizeNode)
			if err != nil {
				return nil, err
			}
			sizeInt, err := size.ToInt()
			if err != nil {
				return nil, err
			}

			return m.evalArray(n.Elt, []ast.Expr{}, sizeInt)
		case 3:
			defaultSize := args[1]
			capacity := args[2]

			size, err := m.Evaluate(defaultSize)
			if err != nil {
				return nil, err
			}
			sizeInt, err := size.ToInt()
			if err != nil {
				return nil, err
			}

			capacityNode, err := m.Evaluate(capacity)
			if err != nil {
				return nil, err
			}
			capInt, err := capacityNode.ToInt()
			if err != nil {
				return nil, err
			}

			return m.evalArray(n.Elt, []ast.Expr{}, sizeInt, capInt)
		default:
			return nil, errors.Errorf("wrong number of arguments %d to make", len(args))
		}
	default:
		return nil, errors.Errorf("unsupported type %v for make", reflect.TypeOf(typeInfo))
	}

}
