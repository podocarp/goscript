package benchmarks

import (
	"math/rand"
	"testing"

	"github.com/podocarp/goscript/machine"
)

var testDataA [][][]float64
var testDataB [][][]float64

func init() {
	testDataA = make([][][]float64, 10)
	for i := 0; i < 10; i++ {
		testDataA[i] = randArr()
	}
	testDataB = make([][][]float64, 10)
	for i := 0; i < 10; i++ {
		testDataB[i] = randArr()
	}
}

func randArr() [][]float64 {
	res := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		t := rand.Int63n(10e9)
		v := rand.Float64() * 1000
		res[i] = []float64{float64(t), v}
	}
	return res
}

func BenchmarkArrayLoop(b *testing.B) {
	m := machine.NewMachine()
	stmt := `func (A, B [][]float64) [][]float64 {
		res := make([][]float64)
		for i := range A {
			ta := A[i][0]
			va := A[i][1]
			tb := B[i][0]
			vb := B[i][1]

			res = append(res, []float64{ ta * tb, va * vb })
		}
		return res
	}`
	fun, _ := m.ParseAndEval(stmt)

	for i := 0; i < b.N; i++ {
		arg1, _ := machine.ValueToNode(testDataA[i%10])
		arg2, _ := machine.ValueToNode(testDataB[i%10])
		m.CallFunction(fun, []*machine.Node{arg1, arg2})
	}
}

func BenchmarkArrayLoopNative(b *testing.B) {
	stmt := func(A, B [][]float64) [][]float64 {
		res := make([][]float64, 0)
		for i := range A {
			ta := A[i][0]
			va := A[i][1]
			tb := B[i][0]
			vb := B[i][1]

			res = append(res, []float64{ta * tb, va * vb})
		}
		return res
	}

	for i := 0; i < b.N; i++ {
		arg1 := testDataA[i%10]
		arg2 := (testDataB[i%10])
		stmt(arg1, arg2)
	}
}
