package tests

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
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
	assert.NotNil(t, err, "should have error")
	assert.Nil(t, res, "should not have result", res)

	stmt = `func(a) {
		for i:= 0; i < 10; i = i+b {
			b := 10
		}

		return 1000
	}(10)`
	res, err = m.ParseAndEval(stmt)
	assert.NotNil(t, err, "should have error")
	assert.Nil(t, res, "should not have result", res)
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
	assert.NotNil(t, err, "should have error")
	assert.Nil(t, res, "should not have result", res)
}
