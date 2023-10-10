package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestLoopsBasic(t *testing.T) {
	m := machine.NewMachine()

	// test loops
	stmt := `
	func (A float64, B float64) {
		for i := 0; i < B; i++ {
			A += i
		}
		return A
	} ( 1 , 10)
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 46, res.Value)

	stmt = `func(a, b) {
		for i := 0; i < a; i++ {
			b += i
		}
		return b
	}( 10, 1 )
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 46.0, res.Value)

	stmt = `func(a, b) {
		for i := 0; i < a; {
			b = i
			i++
		}
		return b
	}( 10, 1 )
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 9, res.Value)
}

func TestLoopsContinue(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)

	stmt := `
	func (A, B int) {
		for i := 0; i < B; i++ {
			if i > 4 {
				if i % 2 == 0 {
					continue
				}
			}
			A += i
		}
		return A
	} ( 1 , 10)
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 32, res.Value)
}

func TestLoopsRangeArray(t *testing.T) {
	m := machine.NewMachine()

	// test range over array
	stmt := `
	func () {
		a := 0
		vals := []int{2,4,6,8,10}
		for i := range vals {
			a += i
		}
		return a
	} ()
	`
	res, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 10, res.Value)

	// test range over array with elem
	stmt = `
	func () {
		a := 0
		vals := []int{2,4,6,8,10}
		for i, b := range vals {
			a += i + b
		}
		return a
	} ()
	`
	res, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 40, res.Value)
}
