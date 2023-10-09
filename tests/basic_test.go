package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

func TestBasicArithmetic(t *testing.T) {
	m := machine.NewMachine()
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
}

func TestConditionalsBasic(t *testing.T) {
	m := machine.NewMachine()

	stmt := `
	func (A, B) {
		if (A>B) {
			return A
		} else {
			return B
		}
		return 1000000000
	} (1 ,2 )
	`
	val, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 2, val.Value)

	stmt = `
	func (A, B) {
		if (A < B) {
			A = B + 2
		}
		return A
	} ( 1 , 2)
	`
	val, err = m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 4, val.Value)
}
