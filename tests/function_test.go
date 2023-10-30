package tests

import (
	"reflect"
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMachineFunctionsSimple(t *testing.T) {
	m := machine.NewMachine()

	stmt := "func() { return 1 }()"
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1.0, res.Value)
}

// TestReturnFunctionLit test that we can return and execute a function literal
func TestReturnFunctionLit(t *testing.T) {
	m := machine.NewMachine()

	stmt := "func(a) { return a }"
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	arg, err := machine.ValueToNode(1)
	assert.Nil(t, err, err)
	callRes, err := m.CallFunction(res, []*machine.Node{arg})
	assert.Nil(t, err)
	assert.EqualValues(t, 1.0, callRes.Value)
}

// TestFunctionDefAndCall test that we can define functions and eval them
func TestFunctionDefAndCall(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		b := func(c) {
			return c
		}

		return b(100)
	}()`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 100.0, res.Value)
}

// TestFunctionArgs tests declaring and using function args
func TestFunctionArgs(t *testing.T) {
	m := machine.NewMachine()

	// declare without type
	stmt := `func(a, b) {
		return b
	}(1, 2)`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 2, res.Value)

	// declare with one type
	stmt = `func(a, b float64) {
		return b
	}(1, 2)`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 2, res.Value)

	// declare with two types
	stmt = `func(a float64, b float64) {
		return a
	}(1, 2)`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)

	// extra arguments work
	stmt = `func(a float64, b float64) {
		return a
	}(1, 2, 3, 4, 5)`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)
}

// TestFunctionClosure tests that function closures work as expected
func TestFunctionClosure(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func() {
		b := 10
		fun := func(c) {
			b = b + c
		}

		fun(10)

		return b
	}()`
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	assert.EqualValues(t, 20.0, res.Value)
}

// TestFunctionReturn tests that return statements work as expected
func TestFunctionReturn(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func(a) {
		b := 0
		for i:= 0; i < a; i++ {
			b = i
			if b > 5 {
				return b
			}
		}

		return 1000
	}(10)`
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	assert.EqualValues(t, 6.0, res.Value)
}

// TestFunctionMultiReturn tests that multi return statements work as expected
func TestFunctionMultiReturn(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func(a) {
		b := 0
		for i:= 0; i < a; i++ {
			b += i
		}

		return b, 1000
	}(10)`
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	val := res.NodeToValue()
	assert.Equal(t, 2, val.Len())
	first := val.Index(0).Interface().(reflect.Value).Int()
	assert.EqualValues(t, 45, first)
	second := val.Index(1).Interface().(reflect.Value).Int()
	assert.EqualValues(t, 1000, second)
}

// TestRecursionBasic tests that recursion works
func TestRecursionBasic(t *testing.T) {
	// limit stack size if it is going to overflow
	m := machine.NewMachine(machine.MachineOptSetMaxDepth(100))
	stmt := `func() {
		Fib := func (n) {
			if n < 2 {
				return n
			}
			return Fib(n-1)
		}
		return Fib(2)
	}()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 1, res.Value)

	stmt = `func() {
		Fib := func (n) {
			if n < 2 {
				return n
			}
			return Fib(n-1) + Fib(n-2)
		}
		return Fib(6)
	}()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 8, res.Value)
}
