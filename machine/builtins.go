package machine

import (
	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
)

type builtin func(args []*Node) (*Node, error)

// A set of all valid builtins' names
var builtins = map[string]builtin{
	"append": Append,
	"len":    Len,
}

func (m *Machine) AddBuiltinsToContext() {
	for name := range builtins {
		m.Context.Set(name, &Node{
			Type:  types.BuiltinType,
			Value: name,
		})
	}
}

func (m *Machine) CallBuiltin(fun *Node, args []*Node) (*Node, error) {
	name := fun.Value.(string)
	if builtin, ok := builtins[name]; ok {
		return builtin(args)
	} else {
		return nil, errors.Errorf("non-existent builtin %s", name)
	}
}

// append(s []T, vs ...T) []T
func Append(args []*Node) (*Node, error) {
	arr := args[0]
	vals := args[1:]
	newValue := arr.Value.([]*Node)
	newValue = append(newValue, vals...)
	arr.Value = newValue
	return arr, nil
}

func Len(args []*Node) (*Node, error) {
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
