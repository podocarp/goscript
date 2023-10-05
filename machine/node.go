package machine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/podocarp/goscript/kind"
)

type Node struct {
	Kind    kind.Kind
	Value   any
	Context *context

	IsReturnValue bool
}

func arrToString(arr []*Node) string {
	var arrContents strings.Builder
	arrContents.WriteString("[ ")
	for _, elem := range arr {
		if elem.Kind == kind.ARRAY {
			arrContents.WriteString(arrToString(elem.Value.([]*Node)))
		} else {
			arrContents.WriteString(strconv.Quote(fmt.Sprint(elem.Value)))
		}
		arrContents.WriteString(" ")
	}
	arrContents.WriteString("]")

	return arrContents.String()
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
			Kind:  kind.STRING,
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
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			node, err := valueToNodeHelper(elem)
			if err != nil {
				return nil, err
			}
			res[i] = node
		}
		return &Node{
			Kind:  kind.ARRAY,
			Value: res,
		}, nil
	default:
		return nil, errors.Errorf("unsupported type %s", val.Type())
	}
}

func (n *Node) String() string {
	var val string
	switch n.Kind {
	case kind.ARRAY:
		val = arrToString(n.Value.([]*Node))
	case kind.FLOAT:
		val = fmt.Sprint(n.Value)
	case kind.STRING:
		val = strconv.Quote(fmt.Sprint(n.Value))
	case kind.FUNC:
		val = "λ"
	default:
		return "unknown type"
	}

	if n.IsReturnValue {
		return val + "[r]"
	}
	return val
}
