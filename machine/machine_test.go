package machine_test

import (
	"go/ast"
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
)

func TestMachineSimple(t *testing.T) {
	m := machine.NewMachine()

	stmt := "func(a) { return 1 }()"
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, res.Value)

	stmt = `func(a, b) {
		for i := 0; i < a; i++ {
			b += i
		}
		return b
	}( 10, 0 )
	`
	res, err = m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	assert.Equal(t, 45.0, res.Value)
}

// TestReturnFunctionLit test that we can return and execute a function literal
func TestReturnFunctionLit(t *testing.T) {
	m := machine.NewMachine()

	stmt := "func(a) { return a }"
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err)
	resLit := res.Value.(*ast.FuncLit)
	callRes, err := m.CallFunction(resLit, []ast.Expr{
		machine.NewFloatLiteral(1),
	})
	assert.Nil(t, err)
	assert.Equal(t, 1.0, callRes.Value)
}

// TestFunctionDefAndCall test that we can define functions and eval them
func TestFunctionDefAndCall(t *testing.T) {
	m := machine.NewMachine()

	stmt := `func(a) {
		b := func(c) {
			return c
		}

		return b(100)
	}()`
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err)
	assert.Equal(t, 100.0, res.Value)
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
	assert.Equal(t, 20.0, res.Value)
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
	assert.Equal(t, 6.0, res.Value)
}

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
