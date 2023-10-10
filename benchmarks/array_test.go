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

func BenchmarkProcessEntireCurve(b *testing.B) {
	m := machine.NewMachine()
	stmt := `func (A, B [][]float64) [][]float64 {
		res := make([][]float64)
		for i := range A {
			ta := A[i][0]
			va := A[i][1]
			tb := B[i][0]
			vb := B[i][1]

			res = append(res, []float64{ ta , va * vb + (ta - tb)})
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

func BenchmarkProcessElementWise(b *testing.B) {
	m := machine.NewMachine()
	stmt := `func (va, vb, ta, tb float64) (float64) {
		return va * vb + (ta - tb)
	}`
	fun, _ := m.ParseAndEval(stmt)

	for i := 0; i < b.N; i++ {
		res := make([][]float64, 0)
		arr1 := testDataA[i%10]
		arr2 := (testDataB[i%10])
		for j := range arr1 {
			arg3, _ := machine.ValueToNode(arr1[j][0])
			arg1, _ := machine.ValueToNode(arr1[j][1])
			arg4, _ := machine.ValueToNode(arr2[j][0])
			arg2, _ := machine.ValueToNode(arr2[j][1])
			val, _ := m.CallFunction(fun, []*machine.Node{arg1, arg2, arg3, arg4})
			r := val.NodeToValue().Float()
			res = append(res, []float64{arr1[j][0], r})
		}
	}
}

func BenchmarkProcessEntireCurveNative(b *testing.B) {
	stmt := func(A, B [][]float64) [][]float64 {
		res := make([][]float64, 0)
		for i := range A {
			ta := A[i][0]
			va := A[i][1]
			tb := B[i][0]
			vb := B[i][1]

			res = append(res, []float64{ta, va*vb + (ta - tb)})
		}
		return res
	}

	for i := 0; i < b.N; i++ {
		arg1 := testDataA[i%10]
		arg2 := (testDataB[i%10])
		stmt(arg1, arg2)
	}
}

func BenchmarkProcessElementWiseNative(b *testing.B) {
	fun := func(va, vb, ta, tb float64) float64 {
		return va*vb + (ta - tb)
	}

	for i := 0; i < b.N; i++ {
		res := make([][]float64, 0)
		arr1 := testDataA[i%10]
		arr2 := (testDataB[i%10])
		for j := range arr1 {
			val := fun(arr1[j][1], arr2[j][1], arr1[j][0], arr2[j][0])
			res = append(res, []float64{arr1[j][0], val})
		}
	}
}
