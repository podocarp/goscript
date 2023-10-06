package machine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/types"
)

type Node struct {
	Type    types.Type
	Value   any
	Context *context

	IsReturnValue bool
}

// ValueToNode takes a value and returns a machien Node representing that
// value.
func ValueToNode(val any) (*Node, error) {
	return valueToNodeHelper(reflect.ValueOf(val))
}

func valueToNodeHelper(val reflect.Value) (*Node, error) {
	switch val.Kind() {
	case reflect.String:
		return &Node{
			Type:  types.StringType,
			Value: val.String(),
		}, nil
	case reflect.Float32, reflect.Float64:
		return Number(val.Float()).ToNode(), nil
	case reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Number(val.Uint()).ToNode(), nil
	case reflect.Int, reflect.Int8,
		reflect.Int16, reflect.Int32, reflect.Int64:
		return Number(val.Int()).ToNode(), nil
	case reflect.Array, reflect.Slice:
		res := make([]*Node, val.Len())
		val.Type().Elem()
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			node, err := valueToNodeHelper(elem)
			if err != nil {
				return nil, err
			}
			res[i] = node
		}
		arrType, err := types.ReflectTypeToType(val.Type())
		if err != nil {
			return nil, err
		}
		return &Node{
			Type:  arrType,
			Value: res,
		}, nil
	default:
		return nil, errors.Errorf("unsupported type %s", val.Type())
	}
}

func arrToString(arr []*Node) string {
	var arrContents strings.Builder
	arrContents.WriteString("[ ")
	for _, elem := range arr {
		if elem.Type.Kind() == types.Array {
			arrContents.WriteString(arrToString(elem.Value.([]*Node)))
		} else {
			arrContents.WriteString(fmt.Sprint(elem.Value))
		}
		arrContents.WriteString(" ")
	}
	arrContents.WriteString("]")

	return arrContents.String()
}

func (n *Node) String() string {
	var val string
	if n == nil {
		return "nil"
	}
	switch n.Type.Kind() {
	case types.Array:
		val = arrToString(n.Value.([]*Node))
	case types.Float:
		val = fmt.Sprint(n.Value)
	case types.String:
		val = strconv.Quote(fmt.Sprint(n.Value))
	case types.Func:
		val = "λ"
	default:
		return "unknown type"
	}

	if n.IsReturnValue {
		val = val + "[r]"
	}
	return fmt.Sprintf(
		"Node(%s) %s",
		n.Type.String(),
		val,
	)
}
