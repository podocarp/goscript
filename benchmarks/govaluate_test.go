package benchmarks_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/podocarp/goscript/machine"
)

var exprs []string
var randints []int

func init() {
	exprs = make([]string, 1000)
	randints = make([]int, 1000)
	for i := 0; i < 1000; i++ {
		exprs[i] = makeRandExpr()
		randints[i] = rand.Intn(1000) + 1
	}
}

func makeRandExpr() string {
	// just random large numbers that won't overflow
	op1 := rand.Intn(1<<30) + 1
	op2 := rand.Intn(1<<30) + 1
	op := []string{"+", "-", "*", "/"}[rand.Intn(4)]

	return fmt.Sprintf("%d %s %d", op1, op, op2)
}

func BenchmarkArithmetic(b *testing.B) {
	m := machine.NewMachine()

	for i := 0; i < b.N; i++ {
		stmt := exprs[i%1000]
		m.ParseAndEval(stmt)
	}

}

func BenchmarkGovaluateArithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stmt := exprs[i%1000]
		expression, _ := govaluate.NewEvaluableExpression(stmt)
		expression.Evaluate(nil)
	}
}
