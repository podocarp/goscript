package machine_test

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
)

func TestMachineSimple(t *testing.T) {
	m := machine.NewMachine(true)

	stmt := "func(a) { return 1 }()"
	res, err := m.ParseAndEval(stmt)
	assert.Nil(t, err)
	assert.Equal(t, "1", res.Value)

	stmt = `func(a, b) {
		b := 0
		for i := 0; i < a; i++ {
			b += i
		}
		return b
	}( 10 )
	`
	res, err = m.ParseAndEval(stmt)
	assert.Nil(t, err)
	assert.Equal(t, "45", res.Value)
}
