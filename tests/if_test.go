package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/require"
)

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

func TestConditionalsAssign(t *testing.T) {
	m := machine.NewMachine()

	stmt := `
	func (A, B) {
		if n := A+B; n < 10 {
			return n
		} else {
			return B
		}
	} (1, 2)
	`
	val, err := m.ParseAndEval(stmt)
	require.Nil(t, err, err)
	require.EqualValues(t, 3, val.Value)
}
