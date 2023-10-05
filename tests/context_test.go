package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestForLoopContext(t *testing.T) {
	m := machine.NewMachine()

	// the for block and for statment contexts should be separate
	stmt := `func(a) {
		for i:= 0; i < b; i++ {
			b := 10
		}

		return 1000
	}(10)`
	res, err := m.ParseAndEval(stmt)
	require.NotNil(t, err, "should have error")
	require.Nil(t, res, "should not have result", res)

	stmt = `func(a) {
		for i:= 0; i < 10; i = i+b {
			b := 10
		}

		return 1000
	}(10)`
	res, err = m.ParseAndEval(stmt)
	require.NotNil(t, err, "should have error")
	require.Nil(t, res, "should not have result", res)
}

func TestIfStmtContext(t *testing.T) {
	m := machine.NewMachine()

	// the if block and if statment contexts should be separate
	stmt := `func(a) {
		if a < b {
			b := 10
			return 1000
		}

		return 1000
	}(10)`
	res, err := m.ParseAndEval(stmt)
	require.NotNil(t, err, "should have error")
	require.Nil(t, res, "should not have result", res)
}

func TestAddToGlobalContext(t *testing.T) {
	m := machine.NewMachine()
	err := m.AddToGlobalContext("b", []int{10})
	require.Nil(t, err, err)

	stmt := `func(a) {
		return a + b[0]
	}(10)`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 20, res.Value)
}
