package machine_test

import (
	"testing"

	"github.com/podocarp/goscript/machine"
	"github.com/stretchr/testify/assert"
)

func TestArray(t *testing.T) {
	m := machine.NewMachine(machine.MachineOptSetDebug)

	stmt := `func(a, b) {
		c := [][]string{ {"1", "2" }, {" 3", "4"}}
	}( 10, 0 )
	`
	_, err := m.ParseAndEval(stmt)
	assert.Nil(t, err, err)
}
