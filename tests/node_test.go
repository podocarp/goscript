package tests

import (
	"reflect"
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/podocarp/goscript/types"
	"github.com/stretchr/testify/require"
)

func TestValueToNode(t *testing.T) {
	tests := []any{
		100, "100", 100.0, uint(100),
	}

	for _, test := range tests {
		node, err := machine.ValueToNode(test)
		require.Nil(t, err)
		require.EqualValues(t, test, node.Value)
	}
}

func TestArrayToNode(t *testing.T) {
	tests := []any{
		[]float64{1.0, 2.0, 3.0},
		[]float64{0},
		[][]float64{{1}, {2, 3, 4}},
	}

	for _, test := range tests {
		node, err := machine.ValueToNode(test)
		require.Nil(t, err, err)
		// types are approximately correct
		require.Equal(t, types.Array, node.Type.Kind())
		elem, _ := node.Type.Elem()
		require.Contains(t, elem.String(), "float")
		// values are correct
		nodeArr := NodeToValue(node).Interface()
		require.EqualValues(t, reflect.ValueOf(test).Interface(), nodeArr)
	}
}

func NodeToValue(n *machine.Node) reflect.Value {
	switch n.Type.Kind() {
	case types.String:
		return reflect.ValueOf(n.Value)
	case types.Float:
		return reflect.ValueOf(n.Value.(float64))
	case types.Func:
		return reflect.ValueOf(n.Value)
	case types.Array:
		arr := n.Value.([]*machine.Node)
		length := len(arr)
		if length == 0 {
			return reflect.Value{}
		}

		firstElemValue := NodeToValue(arr[0])
		sliceType := reflect.SliceOf(firstElemValue.Type())
		res := reflect.MakeSlice(sliceType, 0, length)
		for _, elem := range arr {
			res = reflect.Append(res, NodeToValue(elem))
		}
		return res
	default:
		return reflect.Value{}
	}
}
