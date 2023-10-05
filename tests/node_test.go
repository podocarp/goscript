package tests

import (
	"reflect"
	"testing"

	"github.com/podocarp/goscript/kind"
	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestWrapInNode(t *testing.T) {
	tests := []any{
		100, "100", 100.0, uint(100),
	}

	for _, test := range tests {
		node, err := machine.ValueToNode(test)
		require.Nil(t, err)
		require.EqualValues(t, test, node.Value)
	}
}

func TestWrapArrayInNode(t *testing.T) {
	tests := []any{
		[]float64{1.0, 2.0, 3.0},
		[]float64{0},
		[][]float64{{1}, {2, 3, 4}},
		[][][]string{{{"1", "2"}}, {{"4"}}},
	}

	for _, test := range tests {
		node, err := machine.ValueToNode(test)
		require.Nil(t, err, err)
		nodeArr := NodeToValue(node).Interface()
		require.EqualValues(t, reflect.ValueOf(test).Interface(), nodeArr)
	}
}

func NodeToValue(n *machine.Node) reflect.Value {
	switch n.Kind {
	case kind.STRING:
		return reflect.ValueOf(n.Value)
	case kind.FLOAT:
		return reflect.ValueOf(n.Value.(machine.Number).ToFloat())
	case kind.FUNC:
		return reflect.ValueOf(n.Value)
	case kind.ARRAY:
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
