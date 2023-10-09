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
		nodeArr := node.NodeToValue().Interface()
		require.EqualValues(t, reflect.ValueOf(test).Interface(), nodeArr)
	}
}

func TestNodeToValue(t *testing.T) {
	// Test float
	node := machine.Node{
		Type:  types.FloatType,
		Value: 1.0,
	}
	res := node.NodeToValue()
	require.EqualValues(t, 1, res.Float())

	// Test string
	node = machine.Node{
		Type:  types.StringType,
		Value: "1",
	}
	res = node.NodeToValue()
	require.EqualValues(t, "1", res.String())

	// Test array
	elem1, err := machine.ValueToNode(1.0)
	require.Nil(t, err, err)
	elem2, err := machine.ValueToNode(10.0)
	require.Nil(t, err, err)
	node = machine.Node{
		Type: types.ArrayOf(types.FloatType),
		Value: []*machine.Node{
			elem1,
			elem2,
		},
	}
	res = node.NodeToValue()
	require.EqualValues(t, 2, res.Len())
	require.EqualValues(t, 1, res.Index(0).Float())
	require.EqualValues(t, 10, res.Index(1).Float())

	// Test array of array
	node1 := machine.Node{
		Type: types.ArrayOf(types.FloatType),
		Value: []*machine.Node{
			elem2,
		},
	}
	node2 := machine.Node{
		Type: types.ArrayOf(types.FloatType),
		Value: []*machine.Node{
			elem1,
		},
	}
	node = machine.Node{
		Type: types.ArrayOf(types.ArrayOf(types.FloatType)),
		Value: []*machine.Node{
			&node1,
			&node2,
		},
	}
	res = node.NodeToValue()
	require.EqualValues(t, 2, res.Len())
	require.EqualValues(t, 1, res.Index(0).Len())
	require.EqualValues(t, 1, res.Index(1).Len())
	require.EqualValues(t, 10, res.Index(0).Index(0).Float())
	require.EqualValues(t, 1, res.Index(1).Index(0).Float())
}

func TestValueToNodeInverse(t *testing.T) {
	tests := []any{
		int64(100), "100", 100.0, uint64(100),
	}

	for _, test := range tests {
		node, err := machine.ValueToNode(test)
		require.Nil(t, err)
		val := node.NodeToValue()
		expected := reflect.ValueOf(test)
		require.True(t, expected.Equal(val))
	}
}
