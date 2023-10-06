package machine

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/podocarp/goscript/types"
	"golang.org/x/exp/constraints"
)

type Numeric interface {
	constraints.Integer | constraints.Float
}

type Number float64

func ToNumber[T Numeric](val T) Number {
	return Number(float64(val))
}

func (n Number) isIntegral() bool {
	return n == Number(int(n))
}

func (n Number) ToFloat() float64 {
	return float64(n)
}

func (n Number) ToLiteral() *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.FLOAT,
		Value: strconv.FormatFloat(n.ToFloat(), 'g', -1, 64),
	}
}

func (n Number) ToNode() *Node {
	return &Node{
		Type:  types.FloatType,
		Value: n,
	}
}

func LiteralToNumber(node *ast.BasicLit) Number {
	val, _ := strconv.ParseFloat(node.Value, 64)
	return Number(val)
}
