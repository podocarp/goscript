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

	// For multi return/assign
	Elems []*Node

	IsReturnValue bool
}

func arrToString(arr []*Node) string {
	var arrContents strings.Builder
	arrContents.WriteString("[ ")
	for _, elem := range arr {
		if elem == nil {
			arrContents.WriteString("<NIL>")
		}

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
	if n.Value == nil {
		if n.Elems != nil {
			return fmt.Sprintf("t(%s)", arrToString(n.Elems))
		}
		return "nil"
	}
	switch n.Type.Kind() {
	case types.Array:
		val = arrToString(n.Value.([]*Node))
	case types.Float:
		val = fmt.Sprint(n.Value)
	case types.Int:
		val = fmt.Sprint(n.Value)
	case types.String:
		val = strconv.Quote(fmt.Sprint(n.Value))
	case types.Func:
		val = "λ"
	case types.Builtin:
		val = "builtin"
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

func (n *Node) ToInt() (int64, error) {
	switch n.Type.Kind() {
	case types.Float:
		return int64(n.Value.(float64)), nil
	case types.Int:
		return n.Value.(int64), nil
	case types.Uint:
		return int64(n.Value.(uint64)), nil
	default:
		return 0, errors.Errorf("cannot convert type %v to int", n.Type)
	}
}

func (n *Node) ToFloat() (float64, error) {
	switch n.Type.Kind() {
	case types.Float:
		return n.Value.(float64), nil
	case types.Int:
		return float64(n.Value.(int64)), nil
	case types.Uint:
		return float64(n.Value.(uint64)), nil
	default:
		return 0, errors.Errorf("cannot convert type %v to float", n.Type)
	}
}

func NewBoolNode(val bool) *Node {
	return &Node{
		Type:  types.BoolType,
		Value: val,
	}
}

func NewFloatNode(val float64) *Node {
	return &Node{
		Type:  types.FloatType,
		Value: val,
	}
}

func NewIntNode(val int64) *Node {
	return &Node{
		Type:  types.IntType,
		Value: val,
	}
}

func NewUintNode(val uint64) *Node {
	return &Node{
		Type:  types.UintType,
		Value: val,
	}
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
		return &Node{
			Type:  types.FloatType,
			Value: val.Float(),
		}, nil
	case reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Node{
			Type:  types.UintType,
			Value: val.Uint(),
		}, nil
	case reflect.Int, reflect.Int8,
		reflect.Int16, reflect.Int32, reflect.Int64:
		return &Node{
			Type:  types.IntType,
			Value: val.Int(),
		}, nil
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
