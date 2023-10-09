package tests

import (
	"fmt"
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestBasicBoolean(t *testing.T) {
	m := machine.NewMachine()
	stmt := "true"
	val, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, true, val.Value)

	stmt = "true && true"
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, true, val.Value)

	for _, op1 := range []bool{true, false} {
		for _, op2 := range []bool{true, false} {
			stmt = fmt.Sprintf("%v || %v", op1, op2)
			val, err = m.ParseAndEval(stmt)
			require.Nil(t, err, err)
			require.EqualValues(t, op1 || op2, val.Value)

			stmt = fmt.Sprintf("%v && %v", op1, op2)
			val, err = m.ParseAndEval(stmt)
			require.Nil(t, err, err)
			require.EqualValues(t, op1 && op2, val.Value)
		}
	}
}

func TestBooleanShortCircuit(t *testing.T) {
	m := machine.NewMachine()
	var stmt string
	var err error
	var val *machine.Node

	// sanity check that pred() actually does what we think it does
	stmt = `
	func() {
		b := 0
		pred := func() {
			b++
			return true
		}

		if  pred() {
			return b
		}

		return b
	}()
	`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, val.Value)

	// and statement should short circuit
	stmt = `
	func() {
		b := 0
		pred := func() {
			b++
			return true
		}

		if false && pred() {
			return b
		}

		return b
	}()
	`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 0, val.Value)

	// or statement should short circuit
	stmt = `
	func() {
		b := 0
		pred := func() {
			b++
			return true
		}

		if true || pred() {
			return b
		}

		return b
	}()
	`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 0, val.Value)
}
