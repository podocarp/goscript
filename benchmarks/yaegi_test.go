package benchmarks_test

import (
	"testing"

	"github.com/podocarp/goscript/machine"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func BenchmarkFib(b *testing.B) {
	m := machine.NewMachine(machine.MachineOptSetMaxDepth(1000))
	stmt := `func(n) {
		Fib := func (n) {
			if n < 2 {
				return n
			}
			return Fib(n-1) + Fib(n-2)
		}
		return Fib(n)
	}(30)
	`

	for i := 0; i < b.N; i++ {
		_, err := m.ParseAndEval(stmt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYaegi(b *testing.B) {
	y := interp.New(interp.Options{})
	y.Use(stdlib.Symbols)

	stmt := `func(n int) int {
		var Fib func(n int) int

		Fib = func (n int) int {
			if n < 2 {
				return n
			}
			return Fib(n-1) + Fib(n-2)
		}
		return Fib(n)
	}(30)
	`

	for i := 0; i < b.N; i++ {
		_, err := y.Eval(stmt)
		if err != nil {
			b.Fatal(err)
		}
	}
}
