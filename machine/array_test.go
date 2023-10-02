package machine_test

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
)

// TestArrayDefine test that we can define and index arrays properly
func TestArrayDefine(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)

	stmt := `func() {
		c := []string{"as\nd"}
		return c[0]
	}()
	`
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	assert.Equal(t, "as\nd", res.Value)

	stmt = `func() {
		c := [][]string{ {"1", "2" }, {" 3", "4"}}
		return c[0][1]
	}()
	`
	res, err = m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
	assert.Equal(t, "2", res.Value, res)
}
