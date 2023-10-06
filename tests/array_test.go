package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
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

	// empty array is ok
	stmt := `func() {
		c := []float64{ }
		return c
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, []*machine.Node{}, res.Value)

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

// func TestArrayMake(t *testing.T) {
// 	m := machine.NewMachine(machine.MachineOptSetDebug)
//
// 	stmt := `func() {
// 		c := make([]float64, 1)
// 		return c
// 	}()
// 	`
// 	res, err := m.ParseAndEval(stmt)
// 	require.Nil(t, err, err)
// 	require.EqualValues(t, 1, res.Value)
//
// }

func TestArrayAppend(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)

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

func TestArrayLen(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)

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
