package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestBasicArithmetic(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)
	stmt := "3 + 4.2 * (5 - 2)"
	val, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.InDelta(t, 15.6, val.Value, 1e-6)

	stmt = `
	func (A, B) {
		C := 10
		return A + B + C
	} ( 1 , 2)
	`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 13, val.Value)
}

func TestMultiAssign(t *testing.T) {
	m := machine.NewMachine()

	// test that we can assign and return multi values in one line
	stmt := `func() {
		a, b := 1, 2
		return a, b
	} ()`
	val, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, val.Elems[0].Value)
	require.EqualValues(t, 2, val.Elems[1].Value)

	// test that functions can return multi values
	stmt = `func() {
		ho := func() {
			return 2, 3
		}
		a, b := ho()
		return a, b, 4
	} ()`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 2, val.Elems[0].Value)
	require.EqualValues(t, 3, val.Elems[1].Value)
	require.EqualValues(t, 4, val.Elems[2].Value)
}
