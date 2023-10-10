package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/podocarp/goscript/types"
	"github.com/stretchr/testify/require"
)

// TestArrayDefine test that we can define and index arrays properly
func TestArrayDefine(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		c := []string{"as\nd"}
		return c[0]
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.Equal(t, "as\nd", res.Value)

	stmt = `func() {
		c := [][]string{ {"1", "2" }, {" 3", "4"}}
		return c[0][1]
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.Equal(t, "2", res.Value, res)
}

func TestArrayType(t *testing.T) {
	m := machine.NewMachine()
	var stmt string
	var err error
	var res *machine.Node

	// empty array is ok
	stmt = `func() {
	 	c := []float64{ }
	 	return c
	 }()
	 `
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, []*machine.Node{}, res.Value)

	// normal array is ok
	stmt = `func() {
	 	c := []float64{1}
	 	return c[0]
	 }()
	 `
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)

	// type mismatch is not ok
	stmt = `func() {
	 	c := []float64{ "1" }
	 	return c[0]
	 }()
	`
	res, err = m.ParseAndEval(stmt)
	require.NotNil(t, err, err)
	stmt = `func() {
	 	c := []float64{ 1, 2, "3" }
	 	return c[0]
	 }()
	`
	res, err = m.ParseAndEval(stmt)
	require.NotNil(t, err, err)

	// nested array type is ok
	stmt = `func() {
		c := [][]float64{ { 1 }, {2, 3} }
		return c[0][0]
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)

	// nested array type mismatch not ok
	stmt = `func() {
		c := [][]float64{ { "1" }, {2, 3} }
		return c
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.NotNil(t, err, err)
}

// TestArrayMake tests that make() works as expected
func TestArrayMake(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		c := make([]float64)
		return c
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, []*machine.Node{}, res.Value)

	stmt = `func() {
		c := make([][]float64)
		return c
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, []*machine.Node{}, res.Value)
	expectedType := types.ArrayOf(types.ArrayOf(types.FloatType))
	require.True(t, res.Type.Equal(expectedType), res.Type)
}

// TestArrayAppend tests that append() works as expected
func TestArrayAppend(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		c := []float64{ }
		c = append(c, 1)
		return c[0]
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)

}

func TestArrayAppendMultiD(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func (A, B [][]float64) [][]float64 {
		res := make([][]float64)
		for i := range A {
			ta := A[i][0]
			va := A[i][1]
			tb := B[i][0]
			vb := B[i][1]

			res = append(res, []float64{ ta + tb, va - vb })
		}
		return res
	}`

	fun, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)

	arg1, err := machine.ValueToNode([][]float64{{1, 2}, {3, 4}})
	require.Nil(t, err, err)
	arg2, err := machine.ValueToNode([][]float64{{100, 200}, {30, 40}})
	require.Nil(t, err, err)
	res, err := m.CallFunction(fun, []*machine.Node{arg1, arg2})
	require.Nil(t, err, err)

	val := res.NodeToValue()
	require.EqualValues(t, 2, val.Len())

	arr1 := val.Index(0)
	require.EqualValues(t, 2, arr1.Len())
	require.EqualValues(t, 101, arr1.Index(0).Float())
	require.EqualValues(t, -198, arr1.Index(1).Float())

	arr2 := val.Index(1)
	require.EqualValues(t, 2, arr2.Len())
	require.EqualValues(t, 33, arr2.Index(0).Float())
	require.EqualValues(t, -36, arr2.Index(1).Float())
}

// TestArrayLen tests that len() works as expected
func TestArrayLen(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		c := []float64{1, 2,3 }
		return len(c)
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 3, res.Value)

	stmt = `func() {
		c := []float64{}
		return len(c)
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 0, res.Value)

	stmt = `func() {
		c := [][]float64{ {}, {1, 2, 3, 4}, {2} }
		return len(c)
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 3, res.Value)
}
