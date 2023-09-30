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
	resLit := res.Node.(*ast.BasicLit)
	assert.Equal(t, "1", resLit.Value)

	stmt = `func(a, b) {
		for i := 0; i < a; i++ {
			b += i
		}
		return b
	}( 10, 0 )
	`
	res, err = m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	resLit = res.Node.(*ast.BasicLit)
	assert.Equal(t, "45", resLit.Value)
}

func TestReturnFunctionLit(t *testing.T) {
	m := machine.NewMachine()

	stmt := "func(a) { return a }"
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err)
	resLit := res.Node.(*ast.FuncLit)
	callRes, err := m.CallFunction(resLit, []ast.Expr{
		machine.NewFloatLiteral(1),
	})
	assert.Nil(t, err)
	callResLit := callRes.Node.(*ast.BasicLit)
	assert.Equal(t, "1", callResLit.Value)
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
	resLit := res.Node.(*ast.BasicLit)
	assert.Equal(t, "100", resLit.Value)
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
	resLit := res.Node.(*ast.BasicLit)
	assert.Equal(t, "20", resLit.Value)
}
